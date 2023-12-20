// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"

	"github.com/bitcoin-sv/pulse/cli"
	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/config/limits"
	"github.com/bitcoin-sv/pulse/database"
	"github.com/bitcoin-sv/pulse/domains/logging"
	"github.com/bitcoin-sv/pulse/notification"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints"
	"github.com/bitcoin-sv/pulse/transports/websocket"

	"github.com/bitcoin-sv/pulse/app/logger"

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
	lf, log := createLogger()

	defaultConfig := config.GetDefaultAppConfig()
	if err := cli.LoadFlags(defaultConfig); err != nil {
		log.Errorf("cannot load flags because of error: %v", err)
		os.Exit(1)
	}

	cfg, err := config.Load(lf, defaultConfig)
	if err != nil {
		log.Errorf("cannot load config because of error: %v", err)
		os.Exit(1)
	}

	db, err := database.Init(cfg, log)
	if err != nil {
		log.Errorf("cannot setup database because of error: %v", err)
		os.Exit(1)
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

	logger.SetLevelFromString(lf, cfg.P2PConfig.LogLevel)
	logger.SetLevelFromString(log, cfg.P2PConfig.LogLevel)

	// Do required one-time initialization on wire
	wire.SetLimits(cfg.P2PConfig.ExcessiveBlockSize)

	// Show version at startup.
	log.Infof("Version %s", version.String())

	peers := make(map[*peerpkg.Peer]*peerpkg.PeerSyncState)
	headersStore := sql.NewHeadersDb(db, cfg.DbConfig.Type, lf)
	repo := repository.NewRepositories(headersStore)
	hs := service.NewServices(service.Dept{
		Repositories:  repo,
		Peers:         peers,
		Params:        config.ActiveNetParams.Params,
		AdminToken:    cfg.HTTPConfig.AuthToken,
		LoggerFactory: lf,
		Config:        cfg,
	})
	p2pServer, err := p2p.NewServer(hs, peers, cfg.P2PConfig, lf)
	if err != nil {
		log.Errorf("failed to init a new p2p server: %v\n", err)
		os.Exit(1)
	}

	server := httpserver.NewHttpServer(cfg.HTTPConfig, lf)
	server.ApplyConfiguration(endpoints.SetupPulseRoutes(hs, cfg.HTTPConfig))

	ws, err := websocket.NewServer(lf, hs, cfg.HTTPConfig.UseAuth)
	if err != nil {
		log.Errorf("failed to init a new websocket server: %v\n", err)
		os.Exit(1)
	}
	server.ApplyConfiguration(ws.SetupEntrypoint)

	hs.Notifier.AddChannel(hs.Webhooks)
	hs.Notifier.AddChannel(notification.NewWebsocketChannel(lf, ws.Publisher(), cfg.WebsocketConfig))

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

func createLogger() (logging.LoggerFactory, logging.Logger) {
	lf := logger.DefaultLoggerFactory()
	log := lf.NewLogger("main")
	return lf, log
}
