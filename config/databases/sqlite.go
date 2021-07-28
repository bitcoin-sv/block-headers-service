package databases

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/jmoiron/sqlx"

	// used to import the sqlite drivers.
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"github.com/libsv/bitcoin-hc/config"
)

func setupSqliteDB(c *config.Db) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", c.Dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup database")
	}
	if !c.MigrateDb {
		log.Println("migrate database set to false, skipping migration")
		return db, nil
	}
	log.Println("migrating database")
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("creating sqlite3 db driver failed %s", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", c.SchemaPath), "sqlite3",
		driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}

	}
	log.Println("migrating database completed")
	return db, nil
}
