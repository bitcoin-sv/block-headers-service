package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/domains/logging"
)

// GinEngineOpt represents functions to configure server engine.
type GinEngineOpt func(*gin.Engine)

// HttpServer represents server http.
type HttpServer struct {
	httpServer *http.Server
	handler    *gin.Engine
	log        logging.Logger
}

// NewHttpServer creates and returns HttpServer instance.
func NewHttpServer(cfg *config.HTTP, lf logging.LoggerFactory) *HttpServer {
	handler := gin.Default()

	return &HttpServer{
		httpServer: &http.Server{
			Addr:         ":" + fmt.Sprint(cfg.Port),
			Handler:      handler,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		},
		handler: handler,
		log:     lf.NewLogger("http"),
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
	s.log.Infof("HTTP Server Shutdown")
	return s.httpServer.Shutdown(ctx)
}

// Shutdown is used to stop http server.
func (s *HttpServer) Shutdown() error {
	return s.ShutdownWithContext(context.Background())
}
