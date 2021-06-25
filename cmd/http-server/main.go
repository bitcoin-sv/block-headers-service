package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/centrifugal/centrifuge-go"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"

	"github.com/libsv/headers-client/config"
	"github.com/libsv/headers-client/data/sqlite"
	"github.com/libsv/headers-client/service"
	httpTransport "github.com/libsv/headers-client/transports/http"
	httpMiddleware "github.com/libsv/headers-client/transports/http/middleware"
	"github.com/libsv/headers-client/transports/socket"
)

const appname = "go-headers"

func main() {
	log.Printf("starting %s\n", appname)
	cfg := config.NewViperConfig(appname).
		WithServer().
		WithDb().
		WithDeployment(appname).
		WithLog().
		WithWoc()
	log.Println("setting up db connection")
	db, err := sqlx.Open("sqlite3", cfg.Db.Dsn)
	if err != nil {
		log.Fatalf("failed to setup database: %s", err)
	}
	// nolint:errcheck // dont care about error.
	defer db.Close()
	log.Println("db connection setup")

	log.Println("migrating database")
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("creating sqlite3 db driver failed %s", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://data/sqlite/migrations", "sqlite3",
		driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil { // nolint:govet
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}

	}
	log.Println("migrating database completed")

	e := echo.New()
	e.HideBanner = true
	g := e.Group("/api/")
	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.HTTPErrorHandler = httpMiddleware.ErrorHandler

	headerStore := sqlite.NewHeadersDb(db)
	headerService := service.NewHeadersService(headerStore, headerStore)
	httpTransport.NewHeader(headerService).Routes(g)
	// TODO - we'll need to read our header height from the and then set it.
	height, err := headerStore.Height(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	c := centrifuge.New(fmt.Sprintf("%s%d", cfg.Woc.URL, height), centrifuge.DefaultConfig())
	_ = socket.NewHeaders(c, headerService)
	defer c.Close() // nolint:errcheck
	if err := c.Connect(); err != nil {
		log.Fatal(err)
	}

	e.Logger.Fatal(e.Start(cfg.Server.Port))
}
