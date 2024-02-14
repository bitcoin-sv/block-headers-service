package swagger

import (
	"github.com/bitcoin-sv/block-headers-service/docs"
	"github.com/bitcoin-sv/block-headers-service/service"
	router "github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services, apiUrlPrefix string) router.RootEndpoints {
	return router.RootEndpointsFunc(func(router *gin.RouterGroup) {
		docs.SwaggerInfo.BasePath = apiUrlPrefix
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	})
}
