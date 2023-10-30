package testpulse

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/config/p2pconfig"
	"github.com/bitcoin-sv/pulse/domains/logging"
	testlog "github.com/bitcoin-sv/pulse/internal/tests/log"
	"github.com/bitcoin-sv/pulse/internal/tests/testrepository"
	"github.com/bitcoin-sv/pulse/notification"
	"github.com/bitcoin-sv/pulse/repository"
	"github.com/bitcoin-sv/pulse/service"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints"
	httpserver "github.com/bitcoin-sv/pulse/transports/http/server"
	"github.com/bitcoin-sv/pulse/transports/websocket"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type pulseOpt interface{}

// ServicesOpt represents functions to configure test services.
type ServicesOpt func(*service.Services)

// ConfigOpt represents functions to configure test config.
type ConfigOpt func(*config.Config)

// RepoOpt represents functions to configure test repositories.
type RepoOpt func(*testrepository.TestRepositories)

// Cleanup represents function that is used to clean up TestPulse app.
type Cleanup func()

// TestPulse used to interact with pulse in e2e tests.
type TestPulse struct {
	t            *testing.T
	lf           logging.LoggerFactory
	config       *config.Config
	services     *service.Services
	repositories *repository.Repositories
	ws           websocket.Server
	engine       *gin.Engine
	port         int
	urlPrefix    string
}

// Api Provides test access to pulse API.
func (p *TestPulse) Api() *Api {
	return &Api{TestPulse: p}
}

// Websocket Provides test access to pulse websocket.
func (p *TestPulse) Websocket() *Websocket {
	return &Websocket{TestPulse: p}
}

// When Provides test access to pulse service operations.
func (p *TestPulse) When() *When {
	return &When{TestPulse: p}
}

// NewTestPulse Start pulse for testing reason.
func NewTestPulse(t *testing.T, ops ...pulseOpt) (*TestPulse, Cleanup) {
	//override arguments otherwise all flags provided to go test command will be parsed by LoadConfig
	os.Args = []string{""}

	viper.Reset()
	lf := testlog.NewTestLoggerFactory()
	cfg := config.Init(lf)

	for _, opt := range ops {
		switch opt := opt.(type) {
		case ConfigOpt:
			opt(cfg)
		}
	}

	repo := testrepository.NewCleanTestRepositories()

	for _, opt := range ops {
		switch opt := opt.(type) {
		case RepoOpt:
			opt(&repo)
		}
	}

	hs := service.NewServices(service.Dept{
		Repositories:  repo.ToDomainRepo(),
		Peers:         nil,
		Params:        p2pconfig.ActiveNetParams.Params,
		AdminToken:    cfg.HTTP.AuthToken,
		LoggerFactory: lf,
		Config:        cfg,
	})

	for _, opt := range ops {
		switch opt := opt.(type) {
		case ServicesOpt:
			opt(hs)
		}
	}

	port := cfg.HTTP.Port
	urlPrefix := cfg.HTTP.UrlRefix
	gin.SetMode(gin.TestMode)
	server := httpserver.NewHttpServer(cfg.HTTP, lf)
	server.ApplyConfiguration(endpoints.SetupPulseRoutes(hs, cfg.HTTP))
	engine := hijackEngine(server)

	ws, err := websocket.NewServer(lf, hs, cfg.HTTP.UseAuth)
	if err != nil {
		t.Fatalf("failed to init a new websocket server: %v\n", err)
	}
	server.ApplyConfiguration(ws.SetupEntrypoint)

	hs.Notifier.AddChannel(hs.Webhooks)
	hs.Notifier.AddChannel(notification.NewWebsocketChannel(lf, ws.Publisher(), cfg.Websocket))

	if err := ws.Start(); err != nil {
		panic(fmt.Sprintf("cannot start websocket server because of an error: %v", err))
	}

	go func() {
		err := server.Start()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("cannot start server because of an error: %v", err))
		}
	}()

	pulse := &TestPulse{
		t:            t,
		lf:           lf,
		config:       cfg,
		services:     hs,
		repositories: repo.ToDomainRepo(),
		ws:           ws,
		engine:       engine,
		port:         port,
		urlPrefix:    urlPrefix,
	}

	cleanup := func() {
		if err := ws.Shutdown(); err != nil {
			t.Fatalf("failed to stop websocket server: %v", err)
		}

		if err := server.Shutdown(); err != nil {
			t.Fatalf("failed to stop http server: %v", err)
		}
	}

	return pulse, cleanup
}

func hijackEngine(server *httpserver.HttpServer) *gin.Engine {
	var engine *gin.Engine
	server.ApplyConfiguration(func(e *gin.Engine) {
		engine = e
	})
	return engine
}
