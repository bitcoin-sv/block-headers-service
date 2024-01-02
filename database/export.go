package database

import (
	"database/sql"
	"encoding/csv"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"

	"github.com/bitcoin-sv/pulse/config"
)

const (
	selectHeadersSql = `
	SELECT
		hash,
		version,
		merkleroot,
		nonce,
		bits,
		chainwork,
		strftime('%s', timestamp) as timestamp,
		cumulatedWork
	FROM headers
	WHERE header_state = 'LONGEST_CHAIN'
	ORDER BY height asc
	`
)

func ExportHeaders(cfg *config.AppConfig, log *zerolog.Logger) error {
	log.Info().Msgf("Exporting headers from database to file %s", cfg.Db.PreparedDbFilePath)

	tmpHeadersFileName := "headers.csv"
	tmpHeadersFilePath := filepath.Clean(filepath.Join(os.TempDir(), tmpHeadersFileName))

	db, err := Connect(cfg.Db)
	if err != nil {
		return err
	}

	tmpCsvFile, err := os.Create(tmpHeadersFilePath)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(tmpCsvFile)

	// TODO: Consider querying the database for smaller data chunks to avoid potential performance issues
	rows, err := queryDatabaseTable(db, log)
	if err != nil {
		log.Error().Msgf("Error querying database table: %w", err)
		return err
	}
	defer func(log *zerolog.Logger) {
		if err := rows.Close(); err != nil {
			log.Error().Msgf("Error closing rows: %w", err)
		}
	}(log)

	if err := writeColumnNamesToCsvFile(rows, writer); err != nil {
		return err
	}

	if err := writeRowsToCsvFile(rows, writer); err != nil {
		return err
	}

	writer.Flush()

	log.Info().Msg("Data exported successfully")
	log.Info().Msg("Compressing exported file")

	compressedFile, err := os.Create(cfg.Db.PreparedDbFilePath)
	if err != nil {
		return err
	}

	if err := db.Close(); err != nil {
		return err
	}

	if err := gzipFastCompress(tmpCsvFile, compressedFile); err != nil {
		return err
	}

	if err := tmpCsvFile.Close(); err != nil {
		return err
	}
	if err := compressedFile.Close(); err != nil {
		return err
	}

	if err := os.Remove(tmpHeadersFilePath); err != nil {
		return err
	}

	log.Info().Msgf("File compressed successfully to %s", cfg.Db.PreparedDbFilePath)

	return nil
}

func queryDatabaseTable(db *sqlx.DB, log *zerolog.Logger) (*sqlx.Rows, error) {
	rows, err := db.Queryx(selectHeadersSql)
	if err != nil {
		log.Error().Msgf("Failed to query rows: %v", err)
		return nil, err
	}
	return rows, nil
}

func writeColumnNamesToCsvFile(rows *sqlx.Rows, writer *csv.Writer) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if err := writer.Write(columns); err != nil {
		return err
	}

	return nil
}

func writeRowsToCsvFile(rows *sqlx.Rows, writer *csv.Writer) error {
	for rows.Next() {
		var recordStrings []string

		colTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}

		pointers := make([]interface{}, len(colTypes))

		for i := range colTypes {
			pointers[i] = new(sql.RawBytes)
		}

		if err := rows.Scan(pointers...); err != nil {
			return err
		}

		for _, ptr := range pointers {
			recordStrings = append(recordStrings, string(*ptr.(*sql.RawBytes)))
		}

		if err := writer.Write(recordStrings); err != nil {
			return err
		}
	}

	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}
