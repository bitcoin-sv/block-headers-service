// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"

	"github.com/bitcoin-sv/pulse/cli"
	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/database"
	"github.com/bitcoin-sv/pulse/dbutil"
	"github.com/bitcoin-sv/pulse/notification"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints"
	"github.com/bitcoin-sv/pulse/transports/websocket"

	"github.com/bitcoin-sv/pulse/app/logger"

	"github.com/bitcoin-sv/pulse/config/p2pconfig"
	"github.com/bitcoin-sv/pulse/config/p2pconfig/limits"
	"github.com/bitcoin-sv/pulse/database/sql"
	"github.com/bitcoin-sv/pulse/internal/wire"
	"github.com/bitcoin-sv/pulse/repository"
	"github.com/bitcoin-sv/pulse/service"
	httpserver "github.com/bitcoin-sv/pulse/transports/http/server"
	"github.com/bitcoin-sv/pulse/transports/p2p"
	peerpkg "github.com/bitcoin-sv/pulse/transports/p2p/peer"
	"github.com/bitcoin-sv/pulse/version"
)

// nolint: godot
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	lf := logger.DefaultLoggerFactory()
	log := lf.NewLogger("main")

	cfg, cliFlags := config.Init(lf)

	cli.ParseCliFlags(cliFlags, cfg)

	// TODO: Should this still be here, or should we support DB import only via CLI commands?
	if cfg.Db.PreparedDb {
		if err := dbutil.ImportHeaders(cfg); err != nil {
			fmt.Printf("\nError: %v\n", err)
		}
	}

	db, err := database.Connect(cfg.Db)
	if err != nil {
		log.Errorf("cannot setup database because of error %v", err)
		os.Exit(1)
	}

	if cfg.Db.MigrateDb {
		log.Info("migrating database")
		if err := database.DoMigrations(db, cfg.Db); err != nil {
			log.Errorf("database migration failed because of error %v", err)
			os.Exit(1)
		}
		log.Info("migrating database completed")
	} else {
		log.Info("migrate database set to false, skipping migration")
	}

	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Block and transaction processing can cause bursty allocations.  This
	// limits the garbage collector from excessively overallocating during
	// bursts.  This value was arrived at with the help of profiling live
	// usage.
	debug.SetGCPercent(10)

	// Up some limits.
	if err := limits.SetLimits(); err != nil {
		log.Criticalf("failed to set limits: %v\n", err)
		os.Exit(1)
	}

	logger.SetLevelFromString(lf, cfg.P2P.LogLevel)
	logger.SetLevelFromString(log, cfg.P2P.LogLevel)

	// Do required one-time initialization on wire
	wire.SetLimits(cfg.P2P.ExcessiveBlockSize)

	// Show version at startup.
	log.Infof("Version %s", version.String())

	peers := make(map[*peerpkg.Peer]*peerpkg.PeerSyncState)
	headersStore := sql.NewHeadersDb(db, cfg.Db.Type, lf)
	repo := repository.NewRepositories(headersStore)
	hs := service.NewServices(service.Dept{
		Repositories:  repo,
		Peers:         peers,
		Params:        p2pconfig.ActiveNetParams.Params,
		AdminToken:    cfg.HTTP.AuthToken,
		LoggerFactory: lf,
		Config:        cfg,
	})
	p2pServer, err := p2p.NewServer(hs, peers, cfg.P2P, lf)
	if err != nil {
		log.Errorf("failed to init a new p2p server: %v\n", err)
		os.Exit(1)
	}

	server := httpserver.NewHttpServer(cfg.HTTP, lf)
	server.ApplyConfiguration(endpoints.SetupPulseRoutes(hs, cfg.HTTP))

	ws, err := websocket.NewServer(lf, hs, cfg.HTTP.UseAuth)
	if err != nil {
		log.Errorf("failed to init a new websocket server: %v\n", err)
		os.Exit(1)
	}
	server.ApplyConfiguration(ws.SetupEntrypoint)

	hs.Notifier.AddChannel(hs.Webhooks)
	hs.Notifier.AddChannel(notification.NewWebsocketChannel(lf, ws.Publisher(), cfg.Websocket))

	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("cannot start server because of an error: %v", err)
			os.Exit(1)
		}
	}()

	if err := ws.Start(); err != nil {
		log.Errorf("cannot start websocket server because of an error: %v", err)
		os.Exit(1)
	}

	go func() {
		if err := p2pServer.Start(); err != nil {
			log.Errorf("cannot start p2p server because of an error: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	if err := p2pServer.Shutdown(); err != nil {
		log.Errorf("failed to stop p2p server: %v", err)
	}

	if err := ws.Shutdown(); err != nil {
		log.Errorf("failed to stop websocket server: %v", err)
	}

	if err := server.Shutdown(); err != nil {
		log.Errorf("failed to stop http server: %v", err)
	}
}
