package dbutil

import (
	"database/sql"
	"encoding/csv"
	"os"
	"path"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/database"
	"github.com/bitcoin-sv/pulse/domains/logging"
	"github.com/jmoiron/sqlx"
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

func ExportHeaders(cfg *config.Config, log logging.Logger) error {
	log.Infof("Exporting headers from database to file %s", compressedHeadersFilePath)

	tmpHeadersFileName := "headers.csv"
	tmpHeadersFilePath := path.Join(tmpDir, tmpHeadersFileName)

	db, err := database.Connect(cfg.Db)
	if err != nil {
		return err
	}
	defer db.Close()

	tmpCsvFile, err := os.Create(tmpHeadersFilePath)
	if err != nil {
		return err
	}
	defer tmpCsvFile.Close()

	writer := csv.NewWriter(tmpCsvFile)
	defer writer.Flush()

	rows := queryDatabaseTable(db, log)
	defer rows.Close()

	if err := writeColumnNamesToCsvFile(rows, writer); err != nil {
		return err
	}

	if err := writeRowsToCsvFile(rows, writer); err != nil {
		return err
	}

	log.Info("Data exported successfully")
	log.Info("Compressing exported file")

	compressedFile, err := os.Create(compressedHeadersFilePath)
	if err != nil {
		return err
	}
	defer compressedFile.Close()

	if err := gzipCompress(tmpCsvFile, compressedFile); err != nil {
		return err
	}

	if err := os.Remove(tmpHeadersFilePath); err != nil {
		return err
	}

	log.Infof("File compressed successfully to %s", compressedHeadersFilePath)

	return nil
}

func queryDatabaseTable(db *sqlx.DB, log logging.Logger) *sqlx.Rows {
	rows, err := db.Queryx(selectHeadersSql)
	if err != nil {
		log.Errorf("Failed to query rows: %v", err)
		os.Exit(1)
	}
	return rows
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

		writer.Write(recordStrings)
	}

	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}
