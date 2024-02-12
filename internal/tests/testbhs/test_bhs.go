package testbhs

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/rs/zerolog"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testrepository"
	"github.com/bitcoin-sv/block-headers-service/notification"
	"github.com/bitcoin-sv/block-headers-service/repository"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints"
	httpserver "github.com/bitcoin-sv/block-headers-service/transports/http/server"
	"github.com/bitcoin-sv/block-headers-service/transports/websocket"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type bhsOpt interface{}

// ServicesOpt represents functions to configure test services.
type ServicesOpt func(*service.Services)

// ConfigOpt represents functions to configure test config.
type ConfigOpt func(*config.AppConfig)

// RepoOpt represents functions to configure test repositories.
type RepoOpt func(*testrepository.TestRepositories)

// Cleanup represents function that is used to clean up TestBHS app.
type Cleanup func()

// TestBHS used to interact with block headers service in e2e tests.
type TestBHS struct {
	t            *testing.T
	log          *zerolog.Logger
	config       *config.AppConfig
	services     *service.Services
	repositories *repository.Repositories
	ws           websocket.Server
	engine       *gin.Engine
	port         int
	urlPrefix    string
}

// Api Provides test access to block headers service API.
func (p *TestBHS) Api() *Api {
	return &Api{TestBHS: p}
}

// Websocket Provides test access to block headers service websocket.
func (p *TestBHS) Websocket() *Websocket {
	return &Websocket{TestBHS: p}
}

// When Provides test access to block headers service service operations.
func (p *TestBHS) When() *When {
	return &When{TestBHS: p}
}

// NewTestBHS Start block headers service for testing reason.
func NewTestBHS(t *testing.T, ops ...bhsOpt) (*TestBHS, Cleanup) {
	//override arguments otherwise all flags provided to go test command will be parsed by LoadConfig
	os.Args = []string{""}

	viper.Reset()
	testLog := zerolog.Nop()
	if err := config.SetDefaults(&testLog); err != nil {
		panic(fmt.Sprintf("cannot set config default values: %v", err))
	}
	defaultConfig := config.GetDefaultAppConfig()
	cfg, _, _ := config.Load(defaultConfig)

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
		Repositories: repo.ToDomainRepo(),
		Peers:        nil,
		Params:       config.ActiveNetParams.Params,
		AdminToken:   cfg.HTTP.AuthToken,
		Logger:       &testLog,
		Config:       cfg,
	})

	for _, opt := range ops {
		switch opt := opt.(type) {
		case ServicesOpt:
			opt(hs)
		}
	}

	port := cfg.HTTP.Port
	urlPrefix := "/api/v1"
	gin.SetMode(gin.TestMode)
	server := httpserver.NewHttpServer(cfg.HTTP, &testLog)
	server.ApplyConfiguration(endpoints.SetupBHSRoutes(hs, cfg.HTTP))
	engine := hijackEngine(server)

	ws, err := websocket.NewServer(&testLog, hs, cfg.HTTP.UseAuth)
	if err != nil {
		t.Fatalf("failed to init a new websocket server: %v\n", err)
	}
	server.ApplyConfiguration(ws.SetupEntrypoint)

	hs.Notifier.AddChannel(hs.Webhooks)
	hs.Notifier.AddChannel(notification.NewWebsocketChannel(&testLog, ws.Publisher(), cfg.Websocket))

	if err := ws.Start(); err != nil {
		panic(fmt.Sprintf("cannot start websocket server because of an error: %v", err))
	}

	go func() {
		err := server.Start()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("cannot start server because of an error: %v", err))
		}
	}()

	bhs := &TestBHS{
		t:            t,
		log:          &testLog,
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

	return bhs, cleanup
}

func hijackEngine(server *httpserver.HttpServer) *gin.Engine {
	var engine *gin.Engine
	server.ApplyConfiguration(func(e *gin.Engine) {
		engine = e
	})
	return engine
}
