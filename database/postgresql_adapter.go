package database

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/database/sql"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
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

func (a *postgreSqlAdapter) importHeaders(inputFile *os.File, log *zerolog.Logger) (affectedRows int, err error) {
	// prepare db for bulk insterts
	restoreIndexes, err := a.dropTableIndexes(sql.HeadersTableName)
	if err != nil {
		return
	}
	defer func() {
		if rErr := restoreIndexes(); rErr != nil {
			err = wrapIfNeeded(err, rErr, "Resoring indexes failed")
		}
	}()

	if _, err = inputFile.Seek(0, 0); err != nil {
		return
	}

	reader := csv.NewReader(inputFile)
	_, err = reader.Read() // Skipping the column headers line
	if err != nil {
		return
	}

	// insert headers
	previousBlockHash := chainhash.Hash{}.String()
	var cumulatedChainWork string
	rowIndex := 0
	guard := 0

	for {
		rowIndex, previousBlockHash, cumulatedChainWork, err = a.copyHeaders(reader, postgresBatchSize, previousBlockHash, cumulatedChainWork, rowIndex)
		if err != nil {
			affectedRows = rowIndex
			return
		}

		if guard == rowIndex {
			break
		}

		guard = rowIndex
		affectedRows = rowIndex
	}

	return
}

// dropTableIndexes removes indexes from a table. Returns the index restore function if successful.
func (a *postgreSqlAdapter) dropTableIndexes(table string) (func() error, error) {
	q := fmt.Sprintf("SELECT indexname, indexdef FROM pg_indexes WHERE tablename ='%s' AND indexname != '%s_pkey' AND indexdef IS NOT NULL;", table, table)
	return dropIndexes(a.db, &q)
}

func (a *postgreSqlAdapter) copyHeaders(reader *csv.Reader, batchSize int, previousBlockHash string, cumulatedLastBlockChainWork string, rowIndex int) (lastRowIndex int, lastBlockHash string, cumulatedChainWork string, err error) {
	lastRowIndex = rowIndex
	lastBlockHash = previousBlockHash
	copyQuery := pq.CopyIn(
		sql.HeadersTableName,
		/* columns */ "height", "hash", "version", "merkleroot", "timestamp", "bits", "nonce", "header_state", "chainwork", "cumulated_work", "previous_block",
	)

	dbTx, err := a.db.Begin()
	if err != nil {
		return
	}
	defer dbTx.Rollback() // nolint

	stmt, err := dbTx.Prepare(copyQuery)
	if err != nil {
		return
	}

	cumulatedChainWork = cumulatedLastBlockChainWork
	for i := 0; i < batchSize; i++ {
		record, readErr := reader.Read()
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}

			err = fmt.Errorf("error reading record: %v", readErr)
			_ = stmt.Close() // nolint
			return
		}

		if len(record) == 0 {
			break
		}
		var b *dto.DbBlockHeader
		b, err = prepareRecord(record, lastBlockHash, cumulatedChainWork, lastRowIndex)
		if err != nil {
			return
		}

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

		cumulatedChainWork = b.CumulatedWork
		lastBlockHash = b.Hash
		lastRowIndex++
	}

	_, err = stmt.Exec()
	if err != nil {
		if closeErr := stmt.Close(); closeErr != nil {
			err = fmt.Errorf("execution err: %w. Smt close err: %w", err, closeErr)
		}
		return
	}

	err = stmt.Close() // nolint
	if err != nil {
		return
	}

	err = dbTx.Commit()
	return
}
