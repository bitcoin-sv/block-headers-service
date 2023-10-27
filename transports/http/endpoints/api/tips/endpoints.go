package tips

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitcoin-sv/pulse/service"
	router "github.com/bitcoin-sv/pulse/transports/http/endpoints/routes"
)

type handler struct {
	service service.Headers
}

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.ApiEndpoints {
	return &handler{service: s.Headers}
}

// RegisterApiEndpoints registers routes that are part of service API.
func (h *handler) RegisterApiEndpoints(router *gin.RouterGroup) {
	tip := router.Group("/chain")
	{
		tip.GET("/tip", h.getTips)
		tip.GET("/tip/longest", h.getTipLongestChain)
		tip.GET("/tip/prune/:hash", h.pruneTip)
	}
}

// GetTips godoc.
//
//	@Summary Gets all tips
//	@Tags tip
//	@Accept */*
//	@Produce json
//	@Success 200 {array} []TipStateResponse
//	@Router /chain/tip [get]
//	@Security Bearer
func (h *handler) getTips(c *gin.Context) {
	tips, err := h.service.GetTips()

	if err == nil {
		c.JSON(http.StatusOK, mapToTipStateResponse(tips))
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// getTip godoc.
//
//	@Summary Gets tip of longest chain
//	@Tags tip
//	@Accept */*
//	@Produce json
//	@Success 200 {object} TipStateResponse
//	@Router /chain/tip/longest [get]
//	@Security Bearer
func (h *handler) getTipLongestChain(c *gin.Context) {
	tip := h.service.GetTip()
	c.JSON(http.StatusOK, newTipStateResponse(tip))
}

// PruneTip godoc.
//
//	@Summary Prune tip
//	@Tags tip
//	@Accept */*
//	@Produce json
//	@Success 200 {object} string
//	@Router /chain/tip/prune/{hash} [get]
//	@Param hash path string true "Requested Header Hash"
//	@Security Bearer
func (h *handler) pruneTip(c *gin.Context) {
	param := c.Param("hash")
	fmt.Println(param)
	tip, err := h.service.GetPruneTip()

	if err == nil {
		c.JSON(http.StatusOK, tip)
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}
