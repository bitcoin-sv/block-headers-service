package database

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/jmoiron/sqlx"
	// use blank import to register sqlite driver.
	_ "github.com/mattn/go-sqlite3"
	// use blank import to register postgresql driver.
	_ "github.com/lib/pq"

	"github.com/bitcoin-sv/pulse/config"
)

// DBAdapter defines the interface for a database adapter.
type DBAdapter interface {
	Connect(cfg *config.DbConfig) (*sqlx.DB, error)
	DoMigrations(db *sqlx.DB, cfg *config.DbConfig) error
}

func Init(cfg *config.AppConfig, log *zerolog.Logger) (*sqlx.DB, error) {
	dbLog := log.With().Str("subservice", "database").Logger()

	db, err := Connect(cfg.Db)
	if err != nil {
		return nil, err
	}

	if err := DoMigrations(db, cfg.Db); err != nil {
		return nil, err
	}

	if cfg.Db.PreparedDb {
		if err := ImportHeaders(db, cfg, &dbLog); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// Connect to the database using the specified adapter.
func Connect(cfg *config.DbConfig) (*sqlx.DB, error) {
	adapter, err := NewDBAdapter(cfg)
	if err != nil {
		return nil, err
	}
	return adapter.Connect(cfg)
}

func DoMigrations(db *sqlx.DB, cfg *config.DbConfig) error {
	adapter, err := NewDBAdapter(cfg)
	if err != nil {
		return err
	}

	return adapter.DoMigrations(db, cfg)
}

// NewDBAdapter provides the appropriate database adapter based on the config.
func NewDBAdapter(cfg *config.DbConfig) (DBAdapter, error) {
	switch cfg.Engine {
	case config.DBSqlite:
		return &SQLiteAdapter{}, nil
	case config.DBPostgreSql:
		return &PostgreSqlAdapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported database engine %s", cfg.Engine)
	}
}
