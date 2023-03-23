package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/libsv/bitcoin-hc/configs"
)

// HttpServer represents server http.
type HttpServer struct {
	httpServer *http.Server
}

// NewHttpServer creates and returns HttpServer instance.
func NewHttpServer(port int, handler http.Handler) *HttpServer {
	return &HttpServer{
		httpServer: &http.Server{
			Addr:         ":" + fmt.Sprint(port),
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

// Run is used to start http server.
func (s *HttpServer) Run() error {
	return s.httpServer.ListenAndServe()
}


// Shutdown is used to stop http server.
func (s *HttpServer) Shutdown(ctx context.Context) error {
	configs.Log.Infof("HTTP Server Shutdown")
	return s.httpServer.Shutdown(ctx)
}
