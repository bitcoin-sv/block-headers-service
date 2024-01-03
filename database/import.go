package database

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/database/sql"
	"github.com/bitcoin-sv/pulse/domains"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/pulse/repository"
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

const insertBatchSize = 500

func ImportHeaders(db *sqlx.DB, cfg *config.AppConfig, log *zerolog.Logger) error {
	log.Info().Msg("Import headers from file to the database")

	headersRepo := initHeadersRepo(db, cfg, log)

	dbHeadersCount, _ := headersRepo.GetHeadersCount()

	if dbHeadersCount > 0 {
		log.Info().Msgf("skipping preloading database from file, database already contains %d block headers", dbHeadersCount)
		return nil
	}

	tmpHeadersFile, tmpHeadersFilePath, err := getHeadersFile(cfg.Db.PreparedDbFilePath, log)
	if err != nil {
		return err
	}
	defer func() {
		_ = tmpHeadersFile.Close()

		if fileExistsAndIsReadable(tmpHeadersFilePath) {
			if err := os.Remove(tmpHeadersFilePath); err == nil {
				log.Info().Msgf("Deleted temporary file %s", tmpHeadersFilePath)
			} else {
				log.Warn().Msgf("Unable to delete temporary file %s", tmpHeadersFilePath)
			}
		}
	}()

	pragmas, err := getSQLitePragmaValues(db)
	if err != nil {
		return err
	}

	if err := modifySQLitePragmas(db); err != nil {
		return err
	}
	defer func() {
		if err = restoreSQLitePragmas(db, *pragmas); err != nil {
			log.Error().Msg(err.Error())
			os.Exit(1)
		}
	}()

	droppedIndexes, err := removeIndexes(db)
	if err != nil {
		return err
	}
	defer func() {
		if err = restoreIndexes(db, droppedIndexes); err != nil {
			log.Error().Msg(err.Error())
			os.Exit(1)
		}
	}()

	importCount, err := importHeadersFromFile(headersRepo, tmpHeadersFile, log)
	if err != nil {
		return err
	}

	if err := validateDbConsistency(importCount, headersRepo); err != nil {
		return err
	}

	return nil
}

func initHeadersRepo(db *sqlx.DB, cfg *config.AppConfig, log *zerolog.Logger) *repository.HeaderRepository {
	headersDb := sql.NewHeadersDb(db, cfg.Db.Type, log)
	headersRepo := repository.NewHeadersRepository(headersDb)
	return headersRepo
}

func getHeadersFile(preparedDbFilePath string, log *zerolog.Logger) (*os.File, string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	if !fileExistsAndIsReadable(preparedDbFilePath) {
		return nil, "", fmt.Errorf("file %s does not exist or is not readable", preparedDbFilePath)
	}

	const tmpHeadersFileName = "pulse-blockheaders.csv"

	compressedHeadersFilePath := filepath.Clean(filepath.Join(currentDir, preparedDbFilePath))
	tmpHeadersFilePath := filepath.Clean(filepath.Join(os.TempDir(), tmpHeadersFileName))

	log.Info().Msgf("Decompressing file %s to %s", compressedHeadersFilePath, tmpHeadersFilePath)

	compressedHeadersFile, err := os.Open(compressedHeadersFilePath)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = compressedHeadersFile.Close()
	}()

	tmpHeadersFile, err := os.Create(tmpHeadersFilePath)
	if err != nil {
		return nil, "", err
	}

	if err := gzipDecompressWithBuffer(compressedHeadersFile, tmpHeadersFile); err != nil {
		return nil, "", err
	}

	log.Info().Msgf("Decompressed and wrote contents to %s", tmpHeadersFilePath)

	return tmpHeadersFile, tmpHeadersFilePath, nil
}

func importHeadersFromFile(repo repository.Headers, inputFile *os.File, log *zerolog.Logger) (int, error) {
	log.Info().Msg("Inserting headers from file to the database")

	// Read from the beginning of the file
	if _, err := inputFile.Seek(0, 0); err != nil {
		return 0, err
	}

	reader := csv.NewReader(inputFile)
	_, err := reader.Read() // Skipping the column headers line
	if err != nil {
		return 0, err
	}

	previousBlockHash := chainhash.Hash{}
	rowIndex := 0

	for {
		batch := make([]domains.BlockHeader, 0, insertBatchSize)

		for i := 0; i < insertBatchSize; i++ {
			record, err := reader.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return 0, fmt.Errorf("error reading record: %v", err)
			}

			block := parseRecord(record, int32(rowIndex), previousBlockHash)
			batch = append(batch, block)

			previousBlockHash = block.Hash
			rowIndex++
		}

		if len(batch) == 0 {
			break
		}

		if err := repo.AddMultipleHeadersToDatabase(batch); err != nil {
			return rowIndex, err
		}
	}

	log.Info().Msgf("Inserted total of %d rows", rowIndex)

	return rowIndex, nil
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
		State:         domains.LongestChain,
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

func validateDbConsistency(importCount int, repo *repository.HeaderRepository) error {
	if dbHeadersCount, _ := repo.GetHeadersCount(); dbHeadersCount != importCount {
		return fmt.Errorf("database is not consistent with csv file")
	}

	return nil
}

// TODO hide SQLite-specific code behind some kind of abstraction.
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

// TODO hide SQLite-specific code behind some kind of abstraction.
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

// TODO hide SQLite-specific code behind some kind of abstraction.
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
	if indexesQueryRows.Err() != nil {
		return nil, indexesQueryRows.Err()
	}
	defer func() {
		_ = indexesQueryRows.Close()
	}()

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

	for _, dbIndex := range dbIndexes {
		fmt.Printf("Drop Value: %v\n", dbIndex)

		_, err = db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s;", dbIndex.name))
		if err != nil {
			return nil, err
		}
	}

	return dbIndexes, nil
}

func restoreIndexes(db *sqlx.DB, dbIndexes []DbIndex) error {
	for _, dbIndex := range dbIndexes {
		fmt.Printf("Create Value: %v\n", dbIndex)

		_, err := db.Exec(dbIndex.sql)
		if err != nil {
			return err
		}
	}
	return nil
}
