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

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/database/sql"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/pulse/repository/dto"
	"github.com/rs/zerolog"
)

func ImportHeaders(db DbAdapter, cfg *config.AppConfig, log *zerolog.Logger) error {
	log.Info().Msg("Import headers from file to the database")

	hRepository := sql.NewHeadersDb(db.GetDBx(), log)
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

	importCount, err := db.ImportHeaders(tmpHeadersFile, log)
	if err != nil {
		return err
	}

	log.Info().Msgf("Inserted total of %d rows", importCount)

	if err := validateDbConsistency(importCount, hRepository); err != nil {
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

func parseRecord(record []string, rowIndex int32, previousBlockHash string) dto.DbBlockHeader {
	version := parseInt(record[1])
	bits := parseInt(record[4])
	nonce := parseInt(record[3])
	timestamp := parseInt64(record[6])
	chainWork := parseBigInt(record[5])
	cumulatedWork := parseBigInt(record[7])

	return dto.DbBlockHeader{
		Height:        rowIndex,
		Hash:          parseChainHash(record[0]).String(),
		Version:       int32(version),
		MerkleRoot:    parseChainHash(record[2]).String(),
		Timestamp:     time.Unix(timestamp, 0),
		Bits:          uint32(bits),
		Nonce:         uint32(nonce),
		State:         "LONGEST_CHAIN",
		Chainwork:     chainWork.String(),
		CumulatedWork: cumulatedWork.String(),
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

func validateDbConsistency(importCount int, repo *sql.HeadersDb) error {
	if dbHeadersCount, _ := repo.Count(context.Background()); dbHeadersCount != importCount {
		return errors.New("database is not consistent with csv file")
	}

	return nil
}
