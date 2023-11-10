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

func (a *SQLiteAdapter) Connect(cfg *config.Db) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", cfg.Dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (a *SQLiteAdapter) DoMigrations(db *sqlx.DB, cfg *config.Db) error {
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	sourceUrl := fmt.Sprintf("file://%s", cfg.SchemaPath)
	driverName := "sqlite3"

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
