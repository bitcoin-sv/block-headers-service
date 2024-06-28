package network

import (
	"net/http"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/service"
	router "github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service service.Network
}

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.ApiEndpoints {
	return &handler{service: s.Network}
}

// RegisterApiEndpoints registers routes that are part of service API.
func (h *handler) RegisterApiEndpoints(router *gin.RouterGroup, cfg *config.HTTPConfig) {
	network := router.Group("/network")
	{
		network.GET("/peer", h.getPeers)
		network.GET("/peer/count", h.getPeersCount)
	}
}

// GetPeers godoc.
//
//	@Summary Gets all peers
//	@Tags network
//	@Accept */*
//	@Produce json
//	@Success 200
//	@Router /network/peer [get]
//	@Security Bearer
func (h *handler) getPeers(c *gin.Context) {
	peers := h.service.GetPeers()
	c.JSON(http.StatusOK, peers)
}

// GetPeersCount godoc.
//
//	@Summary Gets peers count
//	@Tags network
//	@Accept */*
//	@Produce json
//	@Success 200 {object} int
//	@Router /network/peer/count [get]
//	@Security Bearer
func (h *handler) getPeersCount(c *gin.Context) {
	count := h.service.GetPeersCount()
	c.JSON(http.StatusOK, count)
}
