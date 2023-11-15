package dbutil

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/bitcoin-sv/pulse/app/logger"
	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/database"
	"github.com/bitcoin-sv/pulse/database/sql"
	"github.com/bitcoin-sv/pulse/domains"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/pulse/repository"
	"github.com/jmoiron/sqlx"
)

type DbIndex struct {
	name string
	sql  string
}

type SQLitePragmaValues struct {
	Synchronous int
	JournalMode string
	CacheSize   int
}

const insertTransactionSize = 500

func ImportHeaders(cfg *config.Config) error {
	fmt.Println("Import headers from file to the database")

	startTime := time.Now()

	if !fileExistsAndIsReadable(compressedHeadersFilePath) {
		return fmt.Errorf("file %s does not exist or is not readable", compressedHeadersFilePath)
	}

	tmpHeadersFileName := "headers.csv"
	tmpHeadersFilePath := path.Join(tmpDir, tmpHeadersFileName)

	db, err := database.Connect(cfg.Db)
	if err != nil {
		return err
	}
	defer db.Close()

	fmt.Println("Running database migrations")

	if database.DoMigrations(db, cfg.Db); err != nil {
		return err
	}

	fmt.Println("Database migrations completed")

	headersRepo := initHeadersRepo(db, cfg)

	dbHeadersCount, _ := headersRepo.GetHeadersCount()

	if dbHeadersCount > 0 {
		fmt.Printf("Database already contains %d block headers\n", dbHeadersCount)
		return errors.New("the headers table in the database must be empty")
	}

	fmt.Printf("Decompressing file %s\n", compressedHeadersFilePath)

	if err := createDirectory(tmpDir); err != nil {
		return err
	}

	if err := gzipDecompress(compressedHeadersFilePath, tmpHeadersFilePath); err != nil {
		return err
	}

	fmt.Printf("Decompressed and wrote contents to %s\n", tmpHeadersFilePath)

	pragmas, err := getSQLitePragmaValues(db)
	if err != nil {
		return err
	}

	if err := modifySQLitePragmas(db); err != nil {
		return err
	}

	droppedIndexes, err := removeIndexes(db)
	if err != nil {
		return err
	}

	if err := importHeadersFromFile(db, headersRepo, tmpHeadersFilePath); err != nil {
		return err
	}

	if err = restoreIndexes(db, droppedIndexes); err != nil {
		return err
	}

	if err = restoreSQLitePragmas(db, *pragmas); err != nil {
		return err
	}

	if err := os.Remove(tmpHeadersFilePath); err != nil {
		return err
	}

	fmt.Printf("Deleted temporary file %s\n", tmpHeadersFilePath)

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	fmt.Printf("Start Time: %s\n", startTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("End Time: %s\n", endTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Elapsed Time: %s\n", elapsedTime)

	return nil
}

func initHeadersRepo(db *sqlx.DB, cfg *config.Config) *repository.HeaderRepository {
	lf := logger.DefaultLoggerFactory()
	headersDb := sql.NewHeadersDb(db, cfg.Db.Type, lf)
	headersRepo := repository.NewHeadersRepository(headersDb)
	return headersRepo
}

func importHeadersFromFile(db *sqlx.DB, repo *repository.HeaderRepository, inputFilePath string) error {
	fmt.Println("Inserting headers from file to the database")

	csvFile, err := os.Open(inputFilePath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	_, err = reader.Read() // Skipping the column headers line
	if err != nil {
		return err
	}

	previousBlockHash := chainhash.Hash{}
	rowIndex := 0

	var blocks []domains.BlockHeader

	for {
		for i := 0; i < insertTransactionSize; i++ {
			record, err := reader.Read()
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading record: %v\n", err)
				}
				break
			}

			block := parseRecord(record, int32(rowIndex), previousBlockHash)
			blocks = append(blocks, block)

			previousBlockHash = block.Hash
			rowIndex++
		}

		if len(blocks) == 0 {
			break
		}

		repo.AddMultipleHeadersToDatabase(blocks)

		fmt.Printf("Inserted %d rows so far\n", rowIndex)

		blocks = nil
	}

	fmt.Printf("Inserted total of %d rows\n", rowIndex)

	return nil
}

func parseRecord(record []string, rowIndex int32, previousBlockHash chainhash.Hash) domains.BlockHeader {
	version := parseInt(record[1])
	bits := parseInt(record[4])
	nonce := parseInt(record[3])
	timestamp := parseInt64(record[6])
	chainWork := parseBigInt(record[5])
	cumulatedWork := parseBigInt(record[7])

	return domains.BlockHeader{
		Height:        rowIndex,
		Hash:          *parseChainHash(record[0]),
		Version:       int32(version),
		MerkleRoot:    *parseChainHash(record[2]),
		Timestamp:     time.Unix(timestamp, 0),
		Bits:          uint32(bits),
		Nonce:         uint32(nonce),
		State:         domains.HeaderState(domains.LongestChain),
		Chainwork:     chainWork,
		CumulatedWork: cumulatedWork,
		PreviousBlock: previousBlockHash,
	}
}

func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func parseChainHash(s string) *chainhash.Hash {
	hash, _ := chainhash.NewHashFromStr(s)
	return hash
}

func parseBigInt(s string) *big.Int {
	bi := new(big.Int)
	bi.SetString(s, 10)
	return bi
}

// TODO hide SQLite-specific code behind some kind of abstraction
func getSQLitePragmaValues(db *sqlx.DB) (*SQLitePragmaValues, error) {
	var pragmaValues SQLitePragmaValues

	pragmaQueries := map[string]interface{}{
		"synchronous":  &pragmaValues.Synchronous,
		"journal_mode": &pragmaValues.JournalMode,
		"cache_size":   &pragmaValues.CacheSize,
	}

	for pragmaName, target := range pragmaQueries {
		query := fmt.Sprintf("PRAGMA %s", pragmaName)
		err := db.QueryRow(query).Scan(target)
		if err != nil {
			return nil, err
		}
	}

	return &pragmaValues, nil
}

// TODO hide SQLite-specific code behind some kind of abstraction
func modifySQLitePragmas(db *sqlx.DB) error {
	pragmas := []string{
		"PRAGMA synchronous = OFF;",
		"PRAGMA journal_mode = MEMORY;",
		"PRAGMA cache_size = 10000;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return err
		}
	}

	return nil
}

// TODO hide SQLite-specific code behind some kind of abstraction
func restoreSQLitePragmas(db *sqlx.DB, values SQLitePragmaValues) error {
	pragmas := []string{
		fmt.Sprintf("PRAGMA synchronous = %d;", values.Synchronous),
		fmt.Sprintf("PRAGMA journal_mode = %s;", values.JournalMode),
		fmt.Sprintf("PRAGMA cache_size = %d;", values.CacheSize),
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return err
		}
	}

	return nil
}

func removeIndexes(db *sqlx.DB) ([]DbIndex, error) {
	var dbIndexes []DbIndex

	indexesQueryRows, err := db.Query("SELECT name, sql FROM sqlite_master WHERE type='index' AND tbl_name ='headers' AND sql IS NOT NULL;")
	if err != nil {
		return nil, err
	}

	for indexesQueryRows.Next() {
		var indexName, indexSQL string
		err := indexesQueryRows.Scan(&indexName, &indexSQL)
		if err != nil {
			return nil, err
		}

		dbIndex := DbIndex{
			name: indexName,
			sql:  indexSQL,
		}

		dbIndexes = append(dbIndexes, dbIndex)
	}

	defer indexesQueryRows.Close()

	for _, dbIndex := range dbIndexes {
		fmt.Printf("Value: %v\n", dbIndex)

		_, err = db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s;", dbIndex.name))
		if err != nil {
			return nil, err
		}
	}

	return dbIndexes, nil
}

func restoreIndexes(db *sqlx.DB, dbIndexes []DbIndex) error {
	for _, dbIndex := range dbIndexes {
		_, err := db.Exec(dbIndex.sql)
		if err != nil {
			return err
		}
	}
	return nil
}
