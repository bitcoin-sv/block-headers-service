package database

import (
	"fmt"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/jmoiron/sqlx"

	// use blank import to register sqlite driver.
	_ "github.com/mattn/go-sqlite3"
)

// DBAdapter defines the interface for a database adapter
type DBAdapter interface {
	Connect(cfg *config.Db) (*sqlx.DB, error)
	DoMigrations(db *sqlx.DB, cfg *config.Db) error
}

// NewDBAdapter provides the appropriate database adapter based on the config
func NewDBAdapter(cfg *config.Db) (DBAdapter, error) {
	switch cfg.Type {
	case "sqlite":
		return &SQLiteAdapter{}, nil
	// TODO: add adapters for other databases, e.g. PostgreSQL
	// case "postgresql":
	//     return &PostgresAdapter{}
	default:
		return nil, fmt.Errorf("unsupported database type %s", cfg.Type)
	}
}

// Connect to the database using the specified adapter
func Connect(cfg *config.Db) (*sqlx.DB, error) {
	adapter, err := NewDBAdapter(cfg)
	if err != nil {
		return nil, err
	}
	return adapter.Connect(cfg)
}

func DoMigrations(db *sqlx.DB, cfg *config.Db) error {
	adapter, err := NewDBAdapter(cfg)
	if err != nil {
		return err
	}

	return adapter.DoMigrations(db, cfg)
}
