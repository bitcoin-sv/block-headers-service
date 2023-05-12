package status

import (
	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/service"
	router "github.com/libsv/bitcoin-hc/transports/http/endpoints/routes"
)

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.RootEndpoints {
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
