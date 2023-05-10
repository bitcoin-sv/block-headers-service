package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/docs"
	p2pservice "github.com/libsv/bitcoin-hc/service"
	"github.com/libsv/bitcoin-hc/transports/http/auth"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const urlPrefix = "http.server.urlPrefix"

// Handler represents handler which creates all routes for http server
// and provide access to repositories.
type Handler struct {
	services       *p2pservice.Services
	authMiddleware auth.TokenMiddleware
}

// NewHandler creates and returns Handler instance.
func NewHandler(services *p2pservice.Services) *Handler {
	return &Handler{
		services:       services,
		authMiddleware: auth.NewAuthTokenMiddleware(services.Tokens),
	}
}

// Init is used to create router and init all routes for http server.
//
//	@title P2P Headers API
//	@version 1.0
//	@description P2P headers API
//	@host localhost:8080
//	@BasePath /
//	@schemes http.
func (h *Handler) Init() *gin.Engine {
	router := gin.Default()
	prefix := viper.GetString(urlPrefix)
	docs.SwaggerInfo.BasePath = prefix
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	v1 := router.Group(prefix, h.authMiddleware.Apply)
	h.initHeadersRoutes(v1)
	h.initNetworkRoutes(v1)
	h.initTipRoutes(v1)
	h.initAccessRoutes(v1)

	return router
}
