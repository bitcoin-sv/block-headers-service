package endpoints

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/config"
	"github.com/libsv/bitcoin-hc/service"
	"github.com/libsv/bitcoin-hc/transports/http/auth"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints/api/access"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints/api/headers"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints/api/merkleroots"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints/api/network"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints/api/tips"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints/api/webhook"
	router "github.com/libsv/bitcoin-hc/transports/http/endpoints/routes"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints/status"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints/swagger"
	httpserver "github.com/libsv/bitcoin-hc/transports/http/server"
	"github.com/spf13/viper"
)

// SetupPulseRoutes main point where we're registering endpoints registrars (handlers that will register endpoints in gin engine)
//
//	and middlewares. It's returning function that can be used to setup engine of httpserver.HttpServer
func SetupPulseRoutes(s *service.Services) httpserver.GinEngineOpt {
	routes := []interface{}{
		status.NewHandler(s),
		swagger.NewHandler(s),
		access.NewHandler(s),
		headers.NewHandler(s),
		network.NewHandler(s),
		tips.NewHandler(s),
		webhook.NewHandler(s),
		merkleroots.NewHandler(s),
	}

	apiMiddlewares := toHandlers(auth.NewMiddleware(s))

	return func(engine *gin.Engine) {
		rootRouter := engine.Group("")
		prefix := viper.GetString(config.EnvHttpServerUrlPrefix)
		apiRouter := engine.Group(prefix, apiMiddlewares...)
		for _, r := range routes {
			switch r := r.(type) {
			case router.RootEndpoints:
				r.RegisterEndpoints(rootRouter)
			case router.ApiEndpoints:
				r.RegisterApiEndpoints(apiRouter)
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
