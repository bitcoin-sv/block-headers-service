package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/docs"
	p2pservice "github.com/libsv/bitcoin-hc/service"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	services *p2pservice.Services
}

func NewHandler(services *p2pservice.Services) *Handler {
	return &Handler{services: services}
}

// @title P2P Headers API
// @version 1.0
// @description P2P headers API
// @host localhost:8080
// @BasePath /
// @schemes http
func (h *Handler) Init() *gin.Engine {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	v1 := router.Group("/api/v1/")
	h.initHeadersRoutes(v1)
	h.initNetworkRoutes(v1)
	h.initTipRoutes(v1)

	return router
}
