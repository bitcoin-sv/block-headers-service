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

// DbAdapter defines the interface for a database adapter.
type DbAdapter interface {
	Connect(cfg *config.DbConfig) error
	DoMigrations(cfg *config.DbConfig) error
	ImportHeaders(inputFile *os.File, log *zerolog.Logger) (int, error)
	GetDBx() *sqlx.DB
}

type dbIndex struct {
	name string
	sql  string
}

func Init(cfg *config.AppConfig, log *zerolog.Logger) (*sqlx.DB, error) {
	dbLog := log.With().Str("subservice", "database").Logger()

	adapter, err := NewDbAdapter(cfg.Db)
	if err != nil {
		return nil, err
	}

	if err = adapter.Connect(cfg.Db); err != nil {
		return nil, err
	}

	if err := adapter.DoMigrations(cfg.Db); err != nil {
		return nil, err
	}

	if cfg.Db.PreparedDb {
		if err := ImportHeaders(adapter, cfg, &dbLog); err != nil {
			return nil, err
		}
	}

	return adapter.GetDBx(), nil
}

// NewDbAdapter provides the appropriate database adapter based on the config.
func NewDbAdapter(cfg *config.DbConfig) (DbAdapter, error) {
	switch cfg.Engine {
	case config.DBSqlite:
		return &SQLiteAdapter{}, nil
	case config.DBPostgreSql:
		return &PostgreSqlAdapter{}, nil
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

	for _, dbIndex := range dbIndexes {
		fmt.Printf("Drop Value: %v\n", dbIndex)

		_, err = db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s;", dbIndex.name))
		if err != nil {
			restoreIndexes(db, dbIndexes)
			return nil, err
		}
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
