package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/gin-gonic/gin"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/logging"
)

// GinEngineOpt represents functions to configure server engine.
type GinEngineOpt func(*gin.Engine)

// HttpServer represents server http.
type HttpServer struct {
	httpServer *http.Server
	handler    *gin.Engine
	log        *zerolog.Logger
}

// NewHttpServer creates and returns HttpServer instance.
func NewHttpServer(cfg *config.HTTPConfig, log *zerolog.Logger) *HttpServer {
	if log.GetLevel() > zerolog.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}

	ginLogger := log.With().Str("subservice", "gin").Logger()
	logging.SetGinWriters(&ginLogger)

	handler := gin.New()
	handler.Use(logging.GinMiddleware(&ginLogger), gin.Recovery())

	serverLogger := log.With().Str("subservice", "server").Logger()

	return &HttpServer{
		httpServer: &http.Server{
			Addr:         ":" + fmt.Sprint(cfg.Port),
			Handler:      handler,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		},
		handler: handler,
		log:     &serverLogger,
	}
}

// ApplyConfiguration it's entrypoint to configure a gin engine used by a server.
func (s *HttpServer) ApplyConfiguration(opts ...GinEngineOpt) {
	for _, config := range opts {
		config(s.handler)
	}
}

// Start is used to start http server.
func (s *HttpServer) Start() error {
	return s.httpServer.ListenAndServe()
}

// ShutdownWithContext is used to stop http server using provided context.
func (s *HttpServer) ShutdownWithContext(ctx context.Context) error {
	s.log.Info().Msg("HTTP Server Shutdown")
	return s.httpServer.Shutdown(ctx)
}

// Shutdown is used to stop http server.
func (s *HttpServer) Shutdown() error {
	return s.ShutdownWithContext(context.Background())
}
