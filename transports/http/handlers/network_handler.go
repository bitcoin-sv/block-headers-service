package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPeers godoc
// @Summary Gets all peers
// @Tags network
// @Accept */*
// @Produce json
// @Success 200
// @Router /network/peers [get]
func (h *Handler) getPeers(c *gin.Context) {
	peers := h.services.Network.GetPeers()
	c.JSON(http.StatusOK, peers)
}

// GetPeersCount godoc
// @Summary Gets peers count
// @Tags network
// @Accept */*
// @Produce json
// @Success 200 {object} int
// @Router /network/peers/count [get]
func (h *Handler) getPeersCount(c *gin.Context) {
	count := h.services.Network.GetPeersCount()
	c.JSON(http.StatusOK, count)
}

func (h *Handler) initNetworkRoutes(router *gin.RouterGroup) {
	network := router.Group("/network")
	{
		network.GET("/peers", h.getPeers)
		network.GET("/peers/count", h.getPeersCount)
	}
}
