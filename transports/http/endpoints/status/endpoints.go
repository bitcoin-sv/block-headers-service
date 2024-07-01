package status

import (
	"github.com/bitcoin-sv/block-headers-service/service"
	router "github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
)

// NewHandler creates new endpoint handler.
func NewHandler(_ *service.Services) router.RootEndpoints {
	return router.RootEndpointsFunc(func(router *gin.RouterGroup) {
		router.GET("status", getStatus)
	})
}

// getStatus godoc.
//
//	@Summary Check the status of the server
//	@Tags status
//	@Accept */*
//	@Produce json
//	@Success 200
//	@Router /../../status [get]
func getStatus(c *gin.Context) {
	c.Status(200)
}
