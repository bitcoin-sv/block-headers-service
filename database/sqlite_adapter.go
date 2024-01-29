package database

import (
	"fmt"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/golang-migrate/migrate/v4"
	sqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"

	// use blank import to use file source driver with the migrate package.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	// use blank import to register sqlite driver.
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteAdapter struct{}

const sqliteDriverName = "sqlite3"

func (a *SQLiteAdapter) Connect(cfg *config.DbConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=true&pooling=true", cfg.Sqlite.FilePath)
	db, err := sqlx.Open(sqliteDriverName, dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (a *SQLiteAdapter) DoMigrations(db *sqlx.DB, cfg *config.DbConfig) error {
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
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
