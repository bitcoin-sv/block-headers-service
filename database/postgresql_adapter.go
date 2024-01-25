package database

import (
	"fmt"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/golang-migrate/migrate/v4"
	postgres "github.com/golang-migrate/migrate/v4/database/postgres"

	// use blank import to use file source driver with the migrate package.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	// use blank import to register postgresql driver.
	_ "github.com/lib/pq"
)

type PostgreSqlAdapter struct{}

const postgresDriverName = "postgres"

func (a *PostgreSqlAdapter) Connect(cfg *config.DbConfig) (*sqlx.DB, error) {
	dbCfg := cfg.Postgres
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.DbName, dbCfg.Sslmode)

	db, err := sqlx.Open(postgresDriverName, dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (a *PostgreSqlAdapter) DoMigrations(db *sqlx.DB, cfg *config.DbConfig) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
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
