package databases

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/libsv/bitcoin-hc/config"
)

func setupMySqlDB(c *config.Db) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", c.Dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup database")
	}
	if !c.MigrateDb {
		log.Println("migrate database set to false, skipping migration")
		return db, nil
	}
	log.Println("migrating database")
	driver, err := mysql.WithInstance(db.DB, &mysql.Config{})
	if err != nil {
		log.Fatalf("creating mysql db driver failed %s", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", c.SchemaPath), "mysql",
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
