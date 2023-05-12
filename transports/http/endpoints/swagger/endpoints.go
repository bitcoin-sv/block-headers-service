package swagger

import (
	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/docs"
	"github.com/libsv/bitcoin-hc/service"
	router "github.com/libsv/bitcoin-hc/transports/http/endpoints/routes"
	"github.com/libsv/bitcoin-hc/vconfig"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.RootEndpoints {
	return router.RootEndpointsFunc(func(router *gin.RouterGroup) {
		prefix := viper.GetString(vconfig.EnvHttpServerUrlPrefix)
		docs.SwaggerInfo.BasePath = prefix
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	})
}
