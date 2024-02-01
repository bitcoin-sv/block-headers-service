package database

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/database/sql"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
	"github.com/golang-migrate/migrate/v4"
	postgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/rs/zerolog"

	// use blank import to use file source driver with the migrate package.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
)

type postgreSqlAdapter struct {
	db *sqlx.DB
}

const postgresDriverName = "postgres"
const postgresBatchSize = 500_000

func (a *postgreSqlAdapter) connect(cfg *config.DbConfig) error {
	dbCfg := cfg.Postgres
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.DbName, dbCfg.Sslmode)

	db, err := sqlx.Open(postgresDriverName, dsn)
	if err != nil {
		return err
	}

	a.db = db
	return nil
}

func (a *postgreSqlAdapter) doMigrations(cfg *config.DbConfig) error {
	driver, err := postgres.WithInstance(a.db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceUrl := fmt.Sprintf("file://%s", cfg.SchemaPath)

	m, err := migrate.NewWithDatabaseInstance(sourceUrl, postgresDriverName, driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (a *postgreSqlAdapter) getDBx() *sqlx.DB {
	if a.db == nil {
		panic("connection to the database has not been established")
	}
	return a.db
}

func (a *postgreSqlAdapter) importHeaders(inputFile *os.File, log *zerolog.Logger) (int, error) {
	// prepare db for bulk insterts
	restoreIndexes, err := a.dropTableIndexes(sql.HeadersTableName)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err = restoreIndexes(); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	if _, err := inputFile.Seek(0, 0); err != nil {
		return 0, err
	}

	reader := csv.NewReader(inputFile)
	_, err = reader.Read() // Skipping the column headers line
	if err != nil {
		return 0, err
	}

	// insert headers
	previousBlockHash := chainhash.Hash{}.String()
	rowIndex := 0
	guard := 0

	for {
		rowIndex, err = a.copyHeaders(reader, postgresBatchSize, previousBlockHash, rowIndex)
		if err != nil {
			return 0, err
		}

		if guard == rowIndex {
			break
		}

		guard = rowIndex
	}

	return rowIndex, nil
}

// dropTableIndexes removes indexes from a table. Returns the index restore function if successful
func (a *postgreSqlAdapter) dropTableIndexes(table string) (func() error, error) {
	q := fmt.Sprintf("SELECT indexname, indexdef FROM pg_indexes WHERE tablename ='%s' AND indexname != '%s_pkey' AND indexdef IS NOT NULL;", table, table)
	return dropIndexes(a.db, &q)
}

func (a *postgreSqlAdapter) copyHeaders(reader *csv.Reader, batchSize int, previousBlockHash string, rowIndex int) (lastRowIndex int, err error) {
	lastRowIndex = rowIndex
	copyQuery := pq.CopyIn(
		sql.HeadersTableName,
		/* columns */ "height", "hash", "version", "merkleroot", "timestamp", "bits", "nonce", "header_state", "chainwork", "cumulated_work", "previous_block",
	)

	dbTx, err := a.db.Begin()
	if err != nil {
		return
	}
	defer dbTx.Rollback()

	stmt, err := dbTx.Prepare(copyQuery)
	if err != nil {
		return
	}

	for i := 0; i < batchSize; i++ {
		record, readErr := reader.Read()
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}

			err = fmt.Errorf("error reading record: %v", readErr)
			_ = stmt.Close()
			return
		}

		if len(record) == 0 {
			break
		}

		b := parseRecord(record, int32(lastRowIndex), previousBlockHash)
		_, execErr := stmt.Exec(
			b.Height,
			b.Hash,
			b.Version,
			b.MerkleRoot,
			b.Timestamp,
			b.Bits,
			b.Nonce,
			b.State,
			b.Chainwork,
			b.CumulatedWork,
			b.PreviousBlock)

		if execErr != nil {
			err = fmt.Errorf("error preparing copy statement after %d row: %v", lastRowIndex, execErr)
			return
		}

		previousBlockHash = b.Hash
		lastRowIndex++
	}

	_, err = stmt.Exec()
	if err != nil {
		if closeErr := stmt.Close(); closeErr != nil {
			err = fmt.Errorf("execution err: %w. Smt close err: %w", err, closeErr)
		}
		return
	}

	err = stmt.Close()
	if err != nil {
		return
	}

	err = dbTx.Commit()
	return
}
