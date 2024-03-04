package database

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/database/sql"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

const (
	rowsInCSVDatabaseFile = 5
)

func importHeaders(db dbAdapter, cfg *config.AppConfig, log *zerolog.Logger) error {
	log.Info().Msg("Import headers from file to the database")

	hRepository := sql.NewHeadersDb(db.getDBx(), log)
	hCount, _ := hRepository.Count(context.Background())

	if hCount > 0 {
		log.Info().Msgf("skipping preloading database from file, database already contains %d block headers", hCount)
		return nil
	}

	tmpHeadersFile, tmpHeadersFilePath, err := getHeadersFile(cfg.Db.PreparedDbFilePath, log)
	if err != nil {
		return err
	}
	defer dropHeadersFile(tmpHeadersFile, tmpHeadersFilePath, log)

	log.Info().Msg("Inserting headers from file to the database")

	importCount, err := db.importHeaders(tmpHeadersFile, log)
	if err != nil {
		return err
	}

	log.Info().Msgf("Inserted total of %d rows", importCount)

	if err := validateDbConsistency(importCount, hRepository, db.getDBx()); err != nil {
		return err
	}

	return nil
}

func getHeadersFile(preparedDbFilePath string, log *zerolog.Logger) (*os.File, string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	if !fileExistsAndIsReadable(preparedDbFilePath) {
		return nil, "", fmt.Errorf("file %s does not exist or is not readable", preparedDbFilePath)
	}

	tmpHeadersFileName := fmt.Sprintf("%d-blockheaders.csv", time.Now().Unix())

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

func dropHeadersFile(tmpHeadersFile *os.File, tmpHeadersFilePath string, log *zerolog.Logger) {
	_ = tmpHeadersFile.Close()

	if fileExistsAndIsReadable(tmpHeadersFilePath) {
		if err := os.Remove(tmpHeadersFilePath); err == nil {
			log.Info().Msgf("Deleted temporary file %s", tmpHeadersFilePath)
		} else {
			log.Warn().Msgf("Unable to delete temporary file %s", tmpHeadersFilePath)
		}
	}
}

func PrepareRecord(record []string, previousBlockHash string, bh service.BlockHasher, cumulatedChainWork string, rowIndex int) (*dto.DbBlockHeader, error) {
	parsedRow, err := parseRecordToBlockHeadersSource(record, previousBlockHash)
	if err != nil {
		return nil, err
	}
	preparedRecord := calculateFields(bh, parsedRow, cumulatedChainWork, rowIndex)
	return preparedRecord, nil
}

func parseRecordToBlockHeadersSource(record []string, previousBlockHash string) (*domains.BlockHeaderSource, error) {
	if len(record) != rowsInCSVDatabaseFile {
		return nil, fmt.Errorf("invalid record length: expected %d elements, got %d", rowsInCSVDatabaseFile, len(record))
	}
	version, err := strconv.ParseInt(record[0], 10, 32)
	if err != nil {
		return nil, err
	}
	merkleroot, err := parseChainHash(record[1])
	if err != nil {
		return nil, err
	}
	nonce, err := strconv.ParseUint(record[2], 10, 32)
	if err != nil {
		return nil, err
	}
	bits, err := strconv.ParseUint(record[3], 10, 32)
	if err != nil {
		return nil, err
	}
	timestamp, err := strconv.ParseInt(record[4], 10, 64)
	if err != nil {
		return nil, err
	}
	prevBlockHash, err := parseChainHash(previousBlockHash)
	if err != nil {
		return nil, err
	}

	blockHeader := domains.BlockHeaderSource{
		Version:    int32(version),
		PrevBlock:  *prevBlockHash,
		MerkleRoot: *merkleroot,
		Timestamp:  time.Unix(timestamp, 0),
		Bits:       uint32(bits),
		Nonce:      uint32(nonce),
	}
	return &blockHeader, nil
}

func calculateFields(bh service.BlockHasher, dbBlock *domains.BlockHeaderSource, cumulatedChainWork string, rowIndex int) *dto.DbBlockHeader {
	blockhash := bh.BlockHash(dbBlock)
	chainWork := domains.CalculateWork(dbBlock.Bits).BigInt()
	cumulatedChainWorkBigInt := parseBigInt(cumulatedChainWork)
	cumulatedChainWorkBigInt.Add(cumulatedChainWorkBigInt, chainWork)

	dbBlockHeader := dto.DbBlockHeader{
		Height:        int32(rowIndex),
		Hash:          blockhash.String(),
		Version:       dbBlock.Version,
		MerkleRoot:    dbBlock.MerkleRoot.String(),
		Timestamp:     dbBlock.Timestamp,
		Bits:          dbBlock.Bits,
		Nonce:         dbBlock.Nonce,
		State:         "LONGEST_CHAIN",
		Chainwork:     chainWork.String(),
		CumulatedWork: cumulatedChainWorkBigInt.String(),
		PreviousBlock: dbBlock.PrevBlock.String(),
	}
	return &dbBlockHeader
}

func parseChainHash(s string) (*chainhash.Hash, error) {
	hash, err := chainhash.NewHashFromStr(s)
	return hash, err
}

func parseBigInt(s string) *big.Int {
	bi := new(big.Int)
	bi.SetString(s, 10)
	return bi
}

func validateDbConsistency(importCount int, repo *sql.HeadersDb, db *sqlx.DB) error {
	ctx := context.Background()

	if dbHeadersCount, _ := repo.Count(ctx); dbHeadersCount != importCount {
		return fmt.Errorf("database is not consistent with csv file, imported %d headers, number of headers in database %d", importCount, dbHeadersCount)
	}

	if maxHeight, _ := repo.Height(ctx); maxHeight != importCount-1 {
		return fmt.Errorf("database is not consistent with csv file, current maximum header height (%d) is different from imported headers number -1 (%d)", maxHeight, importCount)
	}

	if err := validateHeightUniqueness(db); err != nil {
		return fmt.Errorf("database is not consistent with csv file, %w", err)
	}

	if err := validateHashColumn(db); err != nil {
		return fmt.Errorf("database is not consistent with csv file, %w", err)
	}

	if err := validatePrevHashColumn(db); err != nil {
		return fmt.Errorf("database is not consistent with csv file, %w", err)
	}

	return nil
}

func validateHeightUniqueness(db *sqlx.DB) error {
	tmpIndex := "tmp_height_unique"
	_, err := db.Exec(fmt.Sprintf("CREATE UNIQUE INDEX %s ON headers (height)", tmpIndex))
	if err != nil {
		return errors.New("height values are not unique(they should be just after import)")
	} else {
		if _, err = db.Exec(fmt.Sprintf("DROP INDEX %s;", tmpIndex)); err != nil {
			return fmt.Errorf("height values are unique buy droping temporary index %s failed", tmpIndex)
		}
	}

	return nil
}

func validateHashColumn(db *sqlx.DB) error {
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE hash = '%s'", sql.HeadersTableName, chainhash.Hash{}.String())
	var count int

	if err := db.Get(&count, countQuery); err != nil {
		return fmt.Errorf("cannot validate hash column: %w", err)
	}

	if count != 0 {
		return fmt.Errorf("%d is invalid number of rows with hash eq %s", count, chainhash.Hash{}.String())
	}

	return nil
}

func validatePrevHashColumn(db *sqlx.DB) error {
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE previous_block = '%s'", sql.HeadersTableName, chainhash.Hash{}.String())
	var count int

	if err := db.Get(&count, countQuery); err != nil {
		return fmt.Errorf("cannot validate previous_block column: %w", err)
	}

	if count != 1 {
		return fmt.Errorf("%d is invalid number of rows with previous_block eq %s", count, chainhash.Hash{}.String())
	}

	return nil
}
