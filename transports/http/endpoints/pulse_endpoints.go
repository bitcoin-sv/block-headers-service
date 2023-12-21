package endpoints

import (
	"errors"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/service"
	"github.com/bitcoin-sv/pulse/transports/http/auth"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints/api/access"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints/api/headers"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints/api/merkleroots"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints/api/network"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints/api/tips"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints/api/webhook"
	router "github.com/bitcoin-sv/pulse/transports/http/endpoints/routes"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints/status"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints/swagger"
	httpserver "github.com/bitcoin-sv/pulse/transports/http/server"
	"github.com/gin-gonic/gin"
)

// SetupPulseRoutes main point where we're registering endpoints registrars (handlers that will register endpoints in gin engine)
//
//	and middlewares. It's returning function that can be used to setup engine of httpserver.HttpServer
func SetupPulseRoutes(s *service.Services, cfg *config.HTTPConfig) httpserver.GinEngineOpt {
	routes := []interface{}{
		status.NewHandler(s),
		swagger.NewHandler(s, "/api/v1"),
		access.NewHandler(s),
		headers.NewHandler(s),
		network.NewHandler(s),
		tips.NewHandler(s),
		webhook.NewHandler(s),
		merkleroots.NewHandler(s),
	}

	apiMiddlewares := toHandlers(auth.NewMiddleware(s, cfg))

	return func(engine *gin.Engine) {
		rootRouter := engine.Group("")
		prefix := "/api/v1"
		apiRouter := engine.Group(prefix, apiMiddlewares...)
		for _, r := range routes {
			switch r := r.(type) {
			case router.RootEndpoints:
				r.RegisterEndpoints(rootRouter)
			case router.ApiEndpoints:
				r.RegisterApiEndpoints(apiRouter, cfg)
			default:
				panic(errors.New("unexpected router endpoints registrar"))
			}
		}
	}
}

func toHandlers(middlewares ...router.ApiMiddleware) []gin.HandlerFunc {
	result := make([]gin.HandlerFunc, 0)
	for _, m := range middlewares {
		result = append(result, m.ApplyToApi)
	}
	return result
}
