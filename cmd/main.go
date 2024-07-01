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

	"github.com/bitcoin-sv/block-headers-service/cli"
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/database"
	sqlrepository "github.com/bitcoin-sv/block-headers-service/database/repository"
	"github.com/bitcoin-sv/block-headers-service/database/sql"
	p2pexp "github.com/bitcoin-sv/block-headers-service/internal/transports/p2p"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/logging"
	"github.com/bitcoin-sv/block-headers-service/metrics"
	"github.com/bitcoin-sv/block-headers-service/notification"
	"github.com/bitcoin-sv/block-headers-service/repository"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints"
	httpserver "github.com/bitcoin-sv/block-headers-service/transports/http/server"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p"
	peerpkg "github.com/bitcoin-sv/block-headers-service/transports/p2p/peer"
	"github.com/bitcoin-sv/block-headers-service/transports/websocket"
)

// version version of the application that can be overridden with ldflags during build
// (e.g. go build -ldflags "-X main.version=1.2.3").
var version = "development"

type p2pServer interface {
	Start() error
	Shutdown() error
}

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	defaultLog := logging.GetDefaultLogger()

	if err := config.SetDefaults(version, defaultLog); err != nil {
		defaultLog.Error().Msgf("cannot set config default values: %v", err)
		os.Exit(1)
	}

	defaultCfg := config.GetDefaultAppConfig()

	if err := cli.LoadFlags(defaultCfg); err != nil {
		defaultLog.Error().Msgf("cannot load flags because of error: %v", err)
		os.Exit(1)
	}

	cfg, log, err := config.Load(defaultCfg)
	if err != nil {
		defaultLog.Error().Msgf("cannot load config because of error: %v", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		log.Error().Msgf("provided configuration is invalid: %v", err)
		os.Exit(1)
	}

	if cfg.Metrics.Enabled {
		log.Info().Msg("metrics enabled")
		metrics.EnableMetrics()
	}

	db, err := database.Init(cfg, log)
	if err != nil {
		log.Error().Msgf("cannot setup database because of error: %v", err)
		os.Exit(1)
	}

	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Block and transaction processing can cause bursty allocations.  This
	// limits the garbage collector from excessively overallocating during
	// bursts.  This value was arrived at with the help of profiling live
	// usage.
	debug.SetGCPercent(10)

	// Do required one-time initialization on wire
	wire.SetLimits(config.ExcessiveBlockSize)

	// Show version at startup.
	log.Info().Msgf("Version %s", config.Version())

	peers := make(map[*peerpkg.Peer]*peerpkg.SyncState)

	headersStore := sql.NewHeadersDb(db, log)
	repo := &repository.Repositories{
		Headers:  sqlrepository.NewHeadersRepository(headersStore),
		Tokens:   sqlrepository.NewTokensRepository(headersStore),
		Webhooks: sqlrepository.NewWebhooksRepository(headersStore),
	}

	hs := service.NewServices(service.Dept{
		Repositories: repo,
		Peers:        peers,
		AdminToken:   cfg.HTTP.AuthToken,
		Logger:       log,
		Config:       cfg,
	})

	server := httpserver.NewHTTPServer(cfg.HTTP, log)

	server.ApplyConfiguration(metrics.Register)

	server.ApplyConfiguration(endpoints.SetupRoutes(hs, cfg.HTTP))

	ws, err := websocket.NewServer(log, hs, cfg.HTTP.UseAuth)
	if err != nil {
		log.Error().Msgf("failed to init a new websocket server: %v\n", err)
		os.Exit(1)
	}
	server.ApplyConfiguration(ws.SetupEntrypoint)

	hs.Notifier.AddChannel(hs.Webhooks)
	hs.Notifier.AddChannel(notification.NewWebsocketChannel(log, ws.Publisher(), cfg.Websocket))

	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Msgf("cannot start server because of an error: %v", err)
			os.Exit(1)
		}
	}()

	if err := ws.Start(); err != nil {
		log.Error().Msgf("cannot start websocket server because of an error: %v", err)
		os.Exit(1)
	}

	var p2pServer p2pServer

	if cfg.P2P.Experimental {
		p2pServer = p2pexp.NewServer(cfg.P2P, hs.Headers, hs.Chains, log)
	} else {
		p2pServer, err = p2p.NewServer(hs, peers, cfg.P2P, log)
		if err != nil {
			log.Error().Msgf("failed to init a new p2p server: %v\n", err)
			os.Exit(1)
		}
	}

	go func() {
		if err := p2pServer.Start(); err != nil {
			log.Error().Msgf("cannot start p2p server because of an error: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	if err := p2pServer.Shutdown(); err != nil {
		log.Error().Msgf("failed to stop p2p server: %v", err)
	}

	if err := ws.Shutdown(); err != nil {
		log.Error().Msgf("failed to stop websocket server: %v", err)
	}

	if err := server.Shutdown(); err != nil {
		log.Error().Msgf("failed to stop http server: %v", err)
	}
}
