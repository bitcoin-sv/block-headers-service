package profile

import (
	"net/http/pprof"

	"github.com/bitcoin-sv/pulse/service"
	router "github.com/bitcoin-sv/pulse/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
)

// RegisterPprofEndpoints registers routes that are part of pprof.
func NewHandler(s *service.Services, pprofEnabled bool) router.PprofEndpoints {
	return router.PprofEndpointsFunc(func(router *gin.RouterGroup) {
		if pprofEnabled {
			profile := router.Group("/pprof/debug/")
			{
				profile.GET("", gin.WrapF(pprof.Index))
				profile.GET("cmdline", gin.WrapF(pprof.Cmdline))
				profile.GET("profile", gin.WrapF(pprof.Profile))
				profile.GET("symbol", gin.WrapF(pprof.Symbol))
				profile.GET("trace", gin.WrapF(pprof.Trace))
				profile.GET("allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
				profile.GET("block", gin.WrapF(pprof.Handler("block").ServeHTTP))
				profile.GET("goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
				profile.GET("heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
				profile.GET("mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
				profile.GET("threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
			}
		}
	})
}
