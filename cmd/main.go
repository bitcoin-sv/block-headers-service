// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"

	"github.com/libsv/bitcoin-hc/notification"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints"
	"github.com/libsv/bitcoin-hc/transports/websocket"

	"github.com/libsv/bitcoin-hc/app/logger"
	"github.com/libsv/bitcoin-hc/domains/logging"

	"github.com/ulikunitz/xz"

	"github.com/jmoiron/sqlx"
	"github.com/libsv/bitcoin-hc/data/sql"
	"github.com/libsv/bitcoin-hc/internal/wire"
	"github.com/libsv/bitcoin-hc/repository"
	"github.com/libsv/bitcoin-hc/service"
	httpserver "github.com/libsv/bitcoin-hc/transports/http/server"
	"github.com/libsv/bitcoin-hc/transports/p2p"
	peerpkg "github.com/libsv/bitcoin-hc/transports/p2p/peer"
	"github.com/libsv/bitcoin-hc/vconfig"
	"github.com/libsv/bitcoin-hc/vconfig/databases"
	"github.com/libsv/bitcoin-hc/vconfig/p2pconfig"
	"github.com/libsv/bitcoin-hc/vconfig/p2pconfig/limits"
	"github.com/libsv/bitcoin-hc/version"
	"github.com/spf13/viper"
)

const appname = "headers"

// nolint: godot
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	lf := logger.DefaultLoggerFactory()
	log := lf.NewLogger("main")

	flagCfg, err := p2pconfig.ParseFlags("")
	if err != nil {
		log.Criticalf("failed to parse config from flags: %v\n", err)
		os.Exit(1)
	}

	c := vconfig.NewViperConfig(appname, flagCfg).
		WithDb().
		WithAuthorization()

	// Unzip prepared db file if configured.
	if viper.GetBool(vconfig.EnvPreparedDb) {
		err := os.Remove(viper.GetString(vconfig.EnvDbFilePath))
		if err != nil {
			log.Error("Failed to remove old db file")
		}
		err = unzip(viper.GetString(vconfig.EnvPreparedDbFilePath), viper.GetString(vconfig.EnvDbFilePath))
		if err != nil {
			log.Error("Failed to unzip prepared db file - ", err)
		}
	}

	freshDbIfConfigured(log)

	db := runDatabase(c, log)
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

	logger.SetLevelFromString(lf, c.P2PConfig.LogLevel)
	logger.SetLevelFromString(log, c.P2PConfig.LogLevel)

	// Do required one-time initialization on wire
	wire.SetLimits(c.P2PConfig.ExcessiveBlockSize)

	// Show version at startup.
	log.Infof("Version %s", version.String())

	peers := make(map[*peerpkg.Peer]*peerpkg.PeerSyncState)
	headersStore := sql.NewHeadersDb(db, c.Db.Type, lf)
	repo := repository.NewRepositories(headersStore)
	hs := service.NewServices(service.Dept{
		Repositories:  repo,
		Peers:         peers,
		Params:        p2pconfig.ActiveNetParams.Params,
		AdminToken:    viper.GetString(vconfig.EnvHttpServerAuthToken),
		LoggerFactory: lf,
		P2PConfig:     c.P2PConfig,
	})
	p2pServer, err := p2p.NewServer(hs, peers, c.P2PConfig)
	if err != nil {
		log.Errorf("failed to init a new p2p server: %v\n", err)
		os.Exit(1)
	}

	server := httpserver.NewHttpServer(viper.GetInt(vconfig.EnvHttpServerPort), lf)
	server.ApplyConfiguration(endpoints.SetupPulseRoutes(hs))

	ws, err := websocket.NewServer(lf, hs, viper.GetBool(vconfig.EnvHttpServerUseAuth))
	if err != nil {
		log.Errorf("failed to init a new websocket server: %v\n", err)
		os.Exit(1)
	}
	server.ApplyConfiguration(ws.SetupEntrypoint)

	hs.Notifier.AddChannel(hs.Webhooks)
	hs.Notifier.AddChannel(notification.NewWebsocketChannel(lf, ws.Publisher()))

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

func freshDbIfConfigured(log logging.Logger) {
	if viper.GetBool(vconfig.EnvResetDbOnStartup) {
		err := os.Remove(viper.GetString(vconfig.EnvDbFilePath))
		if err != nil && fileExists(viper.GetString(vconfig.EnvDbFilePath)) {
			log.Errorf("%s", err.Error())
			os.Exit(1)
		}
	}
}

func runDatabase(vconfig *vconfig.Config, log logging.Logger) *sqlx.DB {
	db, err := databases.NewDbSetup().
		SetupDb(vconfig.Db)
	if err != nil {
		log.Errorf("cannot setup database, because of error %v", err)
		os.Exit(1)
	}
	return db
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func unzip(src, dest string) (err error) {
	fmt.Println("Unzipping file: " + src + " to " + dest)
	// Open the compressed file for reading
	f, err := os.Open(src) //nolint:gosec //variable is taken from config
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { err = f.Close() }() //nolint:all

	// Create a new file for writing the uncompressed data
	out, err := os.Create(dest) //nolint:gosec //variable is taken from config
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { err = out.Close() }() //nolint:all

	// Create an xz reader to uncompress the data
	r, err := xz.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	// Copy the uncompressed data to the output file
	_, err = io.Copy(out, r)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	fmt.Println("DB file extracted successfully")
	return nil
}
