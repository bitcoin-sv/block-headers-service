package database

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/database/sql"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/golang-migrate/migrate/v4"
	sqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/rs/zerolog"

	// use blank import to use file source driver with the migrate package.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	// use blank import to register sqlite driver.
	_ "github.com/mattn/go-sqlite3"
)

type sqLiteAdapter struct {
	db *sqlx.DB
}

type sqLitePragmaValues struct {
	Synchronous int
	JournalMode string
	CacheSize   int
}

const sqliteDriverName = "sqlite3"
const sqliteBatchSize = 500

func (a *sqLiteAdapter) connect(cfg *config.DbConfig) error {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=true&pooling=true", cfg.Sqlite.FilePath)
	db, err := sqlx.Open(sqliteDriverName, dsn)
	if err != nil {
		return err
	}

	a.db = db
	return nil
}

func (a *sqLiteAdapter) doMigrations(cfg *config.DbConfig) error {
	driver, err := sqlite3.WithInstance(a.db.DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	sourceUrl := fmt.Sprintf("file://%s", cfg.SchemaPath)
	driverName := sqliteDriverName

	m, err := migrate.NewWithDatabaseInstance(sourceUrl, driverName, driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (a *sqLiteAdapter) getDBx() *sqlx.DB {
	if a.db == nil {
		panic("connection to the database has not been established")
	}
	return a.db
}

func (a *sqLiteAdapter) importHeaders(inputFile *os.File, log *zerolog.Logger) (affectedRows int, err error) {
	// prepare db to bulk insterts
	restorePragmas, err := modifySqLitePragmas(a.db)
	if err != nil {
		return
	}
	defer func() {
		if rErr := restorePragmas(); rErr != nil {
			err = wrapIfNeeded(err, rErr, "Resoring previous pragmas failed")
		}
	}()

	restoreIndexes, err := a.dropTableIndexes(sql.HeadersTableName)
	if err != nil {
		return
	}
	defer func() {
		if rErr := restoreIndexes(); rErr != nil {
			err = wrapIfNeeded(err, rErr, "Resoring indexes failed")
		}
	}()

	// Read from the beginning of the file
	if _, err = inputFile.Seek(0, 0); err != nil {
		return
	}

	reader := csv.NewReader(inputFile)
	_, err = reader.Read() // Skipping the column headers line
	if err != nil {
		return
	}

	repo := sql.NewHeadersDb(a.db, log)

	previousBlockHash := chainhash.Hash{}.String()
	var cumulatedChainWork string
	rowIndex := 0
	guard := 0

	for {
		rowIndex, previousBlockHash, cumulatedChainWork, err = a.insertHeaders(reader, repo, sqliteBatchSize, previousBlockHash, cumulatedChainWork, rowIndex)
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

func modifySqLitePragmas(db *sqlx.DB) (func() error, error) {
	old_pragmas, err := getSqLitePragmaValues(db)
	if err != nil {
		return nil, err
	}

	pragmas := []string{
		"PRAGMA synchronous = OFF;",
		"PRAGMA journal_mode = MEMORY;",
		"PRAGMA cache_size = 10000;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			if rErr := restoreSqLitePragmas(db, *old_pragmas); rErr != nil {
				err = fmt.Errorf("%w. Resoring previous pragmas failed: %w", err, rErr)
			}
			return nil, err
		}
	}

	return func() error { return restoreSqLitePragmas(db, *old_pragmas) }, nil
}

func getSqLitePragmaValues(db *sqlx.DB) (*sqLitePragmaValues, error) {
	var pragmaValues sqLitePragmaValues

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

func restoreSqLitePragmas(db *sqlx.DB, values sqLitePragmaValues) error {
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

// dropTableIndexes removes indexes from a table. Returns the index restore function if successful.
func (a *sqLiteAdapter) dropTableIndexes(table string) (func() error, error) {
	q := fmt.Sprintf("SELECT name, sql FROM sqlite_master WHERE type='index' AND tbl_name ='%s' AND sql IS NOT NULL;", table)
	return dropIndexes(a.db, &q)
}

func (a *sqLiteAdapter) insertHeaders(reader *csv.Reader, repo *sql.HeadersDb, batchSize int, previousBlockHash string, cumulatedLastBlockChainWork string, rowIndex int) (lastRowIndex int, lastBlockHash string, cumulatedChainwork string, err error) {
	lastRowIndex = rowIndex
	lastBlockHash = previousBlockHash
	batch := make([]dto.DbBlockHeader, 0, batchSize)
	cumulatedChainwork = cumulatedLastBlockChainWork
	bh := service.DefaultBlockHasher()

	for i := 0; i < batchSize; i++ {
		record, readErr := reader.Read()
		if err != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			err = fmt.Errorf("error reading record: %v", readErr)
			return
		}

		if len(record) == 0 {
			break
		}
		var block dto.DbBlockHeader
		block, err = PrepareRecord(record, lastBlockHash, bh, cumulatedChainwork, lastRowIndex)
		if err != nil {
			fmt.Printf("Error while preparing record: %v", err.Error())
			os.Exit(1)
		}
		batch = append(batch, block)

		cumulatedChainwork = block.CumulatedWork
		lastBlockHash = block.Hash
		lastRowIndex++
	}

	if err = repo.CreateMultiple(context.Background(), batch); err != nil {
		return
	}

	return
}
