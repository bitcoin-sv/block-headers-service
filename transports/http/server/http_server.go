package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/logging"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// GinEngineOpt represents functions to configure server engine.
type GinEngineOpt func(*gin.Engine)

// HTTPServer represents server http.
type HTTPServer struct {
	httpServer *http.Server
	handler    *gin.Engine
	log        *zerolog.Logger
}

// NewHTTPServer creates and returns HTTPServer instance.
func NewHTTPServer(cfg *config.HTTPConfig, log *zerolog.Logger) *HTTPServer {
	if log.GetLevel() > zerolog.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}

	ginLogger := log.With().Str("subservice", "gin").Logger()
	logging.SetGinWriters(&ginLogger)

	handler := gin.New()
	handler.Use(logging.GinMiddleware(&ginLogger), gin.Recovery())

	serverLogger := log.With().Str("subservice", "server").Logger()

	return &HTTPServer{
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
func (s *HTTPServer) ApplyConfiguration(opts ...GinEngineOpt) {
	for _, config := range opts {
		config(s.handler)
	}
}

// Start is used to start http server.
func (s *HTTPServer) Start() error {
	return s.httpServer.ListenAndServe()
}

// ShutdownWithContext is used to stop http server using provided context.
func (s *HTTPServer) ShutdownWithContext(ctx context.Context) error {
	s.log.Info().Msg("HTTP Server Shutdown")
	return s.httpServer.Shutdown(ctx)
}

// Shutdown is used to stop http server.
func (s *HTTPServer) Shutdown() error {
	return s.ShutdownWithContext(context.Background())
}
