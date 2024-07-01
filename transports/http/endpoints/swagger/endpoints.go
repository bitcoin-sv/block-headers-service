package swagger

import (
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/docs"
	"github.com/bitcoin-sv/block-headers-service/service"
	router "github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewHandler creates new endpoint handler.
func NewHandler(_ *service.Services, apiURLPrefix string) router.RootEndpoints {
	return router.RootEndpointsFunc(func(router *gin.RouterGroup) {
		docs.SwaggerInfo.BasePath = apiURLPrefix
		docs.SwaggerInfo.Version = config.Version()
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	})
}
