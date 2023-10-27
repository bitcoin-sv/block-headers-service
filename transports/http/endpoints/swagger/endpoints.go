package swagger

import (
	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/docs"
	"github.com/bitcoin-sv/pulse/service"
	router "github.com/bitcoin-sv/pulse/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.RootEndpoints {
	return router.RootEndpointsFunc(func(router *gin.RouterGroup) {
		prefix := viper.GetString(config.EnvHttpServerUrlPrefix)
		docs.SwaggerInfo.BasePath = prefix
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	})
}
