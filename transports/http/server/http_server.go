package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/libsv/bitcoin-hc/configs"
)

type HttpServer struct {
	httpServer *http.Server
}

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

func (s *HttpServer) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	configs.Log.Infof("HTTP Server Shutdown")
	return s.httpServer.Shutdown(ctx)
}
