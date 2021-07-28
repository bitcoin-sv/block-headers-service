package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/theflyingcodr/centrifuge-go"

	"github.com/libsv/bitcoin-hc/config"
	"github.com/libsv/bitcoin-hc/config/databases"
	hcHttp "github.com/libsv/bitcoin-hc/data/http"
	"github.com/libsv/bitcoin-hc/data/sql"
	"github.com/libsv/bitcoin-hc/service"
	httpTransport "github.com/libsv/bitcoin-hc/transports/http"
	httpMiddleware "github.com/libsv/bitcoin-hc/transports/http/middleware"
	"github.com/libsv/bitcoin-hc/transports/socket"
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
	if err := cfg.Validate(); err != nil {
		log.Fatalf("%s", err)
	}
	lvl, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		log.Println(err)
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(lvl)
	}

	log.Printf("setting up %s db connection \n", cfg.Db.Type)
	db, err := databases.NewDbSetup().SetupDb(cfg.Db)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	log.Println("db connection setup")

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

	headerStore := sql.NewHeadersDb(db, cfg.Db.Type)
	headerService := service.NewHeadersService(headerStore, headerStore, hcHttp.NewWhatsOnChain(&http.Client{
		Timeout: time.Second * 30,
	}))
	httpTransport.NewHeader(headerService).Routes(g)
	// TODO - we'll need to read our header height from the and then set it.
	height, err := headerStore.Height(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	configWs := centrifuge.DefaultConfig()
	c := centrifuge.New(fmt.Sprintf("%s%d", cfg.Woc.URL, height), configWs)
	defer c.Close() // nolint:errcheck // this is why
	if err := c.Connect(); err != nil {
		log.Fatal(err)
	}
	headerSocket := socket.NewHeaders(c, cfg.Woc, headerService)
	defer headerSocket.Close()

	e.Logger.Fatal(e.Start(cfg.Server.Port))

}
