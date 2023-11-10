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

func ImportHeaders(cfg *config.Config) error {
	fmt.Println("Import headers from file to the database")

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

	if err := importHeadersFromFile(db, headersRepo, tmpHeadersFilePath); err != nil {
		return err
	}

	if err := os.Remove(tmpHeadersFilePath); err != nil {
		return err
	}

	fmt.Printf("Deleted temporary file %s\n", tmpHeadersFilePath)

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

	for {
		record, err := reader.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading record: %v\n", err)
			}
			break
		}

		block := parseRecord(record, int32(rowIndex), previousBlockHash)

		repo.AddHeaderToDatabase(block)

		if rowIndex%1000 == 999 {
			fmt.Printf("Inserted %d rows so far\n", rowIndex+1)
		}

		previousBlockHash = block.Hash
		rowIndex++
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
