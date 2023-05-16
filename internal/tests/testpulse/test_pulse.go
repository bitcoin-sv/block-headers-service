package testpulse

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/configs"
	"github.com/libsv/bitcoin-hc/internal/tests/testrepository"
	"github.com/libsv/bitcoin-hc/repository"
	"github.com/libsv/bitcoin-hc/service"
	handler "github.com/libsv/bitcoin-hc/transports/http/handlers"
	httpserver "github.com/libsv/bitcoin-hc/transports/http/server"
	"github.com/libsv/bitcoin-hc/vconfig"
	"github.com/spf13/viper"
)

type pulseOpt interface{}

// ServicesOpt represents functions to configure test services.
type ServicesOpt func(*service.Services)

// ConfigOpt represents functions to configure test config.
type ConfigOpt func(*vconfig.Config)

// RepoOpt represents functions to configure test repositories.
type RepoOpt func(*repository.Repositories)

// ServerOpt represents functions to configure test server.
type ServerOpt func(*gin.Engine)

// Cleanup represents function that is used to clean up TestPulse app.
type Cleanup func()

// TestPulse used to interact with pulse in e2e tests.
type TestPulse struct {
	t            *testing.T
	config       *vconfig.Config
	services     *service.Services
	repositories *repository.Repositories
	engine       *gin.Engine
	port         int
	urlPrefix    string
}

// Api Provides test access to pulse API.
func (p *TestPulse) Api() *Api {
	return &Api{TestPulse: p}
}

// NewTestPulse Start pulse for testing reason.
func NewTestPulse(t *testing.T, ops ...pulseOpt) (*TestPulse, Cleanup) {
	viper.Reset()
	conf := vconfig.NewViperConfig("test-pulse")

	//override arguments otherwise all flags provided to go test command will be parsed by LoadConfig
	os.Args = []string{""}

	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	err := configs.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v\n", err)
	}

	repo := testrepository.NewCleanTestRepositories()

	hs := service.NewServices(service.Dept{
		Repositories: &repo,
		Peers:        nil,
	})

	handlers := handler.NewHandler(hs)
	gin.SetMode(gin.TestMode)
	ginEngine := handlers.Init()

	for _, opt := range ops {
		switch opt := opt.(type) {
		case ConfigOpt:
			opt(conf)
		case RepoOpt:
			opt(&repo)
		case ServicesOpt:
			opt(hs)
		case ServerOpt:
			opt(ginEngine)
		}
	}

	port := viper.GetInt(vconfig.EnvHttpServerPort)
	urlPrefix := viper.GetString(vconfig.EnvHttpServerUrlPrefix)
	httpServer := httpserver.NewHttpServer(port, ginEngine)

	go func() {
		err := httpServer.Run()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("cannot start server because of an error: %v", err))
		}
	}()

	pulse := &TestPulse{
		t:            t,
		config:       conf,
		services:     hs,
		repositories: &repo,
		engine:       ginEngine,
		port:         port,
		urlPrefix:    urlPrefix,
	}

	cleanup := func() {
		if err := httpServer.Shutdown(context.Background()); err != nil {
			t.Fatalf("failed to stop http server: %v", err)
		}
	}

	return pulse, cleanup
}
