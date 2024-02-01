package database

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/jmoiron/sqlx"
	// use blank import to register sqlite driver.
	_ "github.com/mattn/go-sqlite3"
	// use blank import to register postgresql driver.
	_ "github.com/lib/pq"

	"github.com/bitcoin-sv/pulse/config"
)

type dbAdapter interface {
	connect(cfg *config.DbConfig) error
	doMigrations(cfg *config.DbConfig) error
	importHeaders(inputFile *os.File, log *zerolog.Logger) (int, error)
	getDBx() *sqlx.DB
}

type dbIndex struct {
	name string
	sql  string
}

func Init(cfg *config.AppConfig, log *zerolog.Logger) (*sqlx.DB, error) {
	dbLog := log.With().Str("subservice", "database").Logger()

	adapter, err := newDbAdapter(cfg.Db)
	if err != nil {
		return nil, err
	}

	if err = adapter.connect(cfg.Db); err != nil {
		return nil, err
	}

	if err := adapter.doMigrations(cfg.Db); err != nil {
		return nil, err
	}

	if cfg.Db.PreparedDb {
		if err := importHeaders(adapter, cfg, &dbLog); err != nil {
			return nil, err
		}
	}

	return adapter.getDBx(), nil
}

func newDbAdapter(cfg *config.DbConfig) (dbAdapter, error) {
	switch cfg.Engine {
	case config.DBSqlite:
		return &sqLiteAdapter{}, nil
	case config.DBPostgreSql:
		return &postgreSqlAdapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported database engine %s", cfg.Engine)
	}
}

func dropIndexes(db *sqlx.DB, indexQuery *string) (func() error, error) {
	qr, err := db.Query(*indexQuery)
	if err != nil {
		return nil, err
	}
	if qr.Err() != nil {
		return nil, qr.Err()
	}
	defer func() {
		_ = qr.Close()
	}()

	var dbIndexes []dbIndex
	for qr.Next() {
		var indexName, indexSQL string
		err := qr.Scan(&indexName, &indexSQL)
		if err != nil {
			return nil, err
		}

		dbIndex := dbIndex{
			name: indexName,
			sql:  indexSQL,
		}

		dbIndexes = append(dbIndexes, dbIndex)
	}

	dropedIndexes := make([]dbIndex, 0)
	for _, dbIndex := range dbIndexes {
		fmt.Printf("Drop Value: %v\n", dbIndex)

		_, err = db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s;", dbIndex.name))
		if err != nil {
			if restoreErr := restoreIndexes(db, dropedIndexes); restoreErr != nil {
				err = fmt.Errorf("%w. Restoring already droped indexes failed: %w", err, restoreErr)
			}

			return nil, err
		}

		dropedIndexes = append(dropedIndexes, dbIndex)
	}

	return func() error { return restoreIndexes(db, dbIndexes) }, nil
}

func restoreIndexes(db *sqlx.DB, dbIndexes []dbIndex) error {
	for _, dbIndex := range dbIndexes {
		fmt.Printf("Create Value: %v\n", dbIndex)

		_, err := db.Exec(dbIndex.sql)
		if err != nil {
			return err
		}
	}
	return nil
}
