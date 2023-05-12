package tips

import (
	"fmt"
	"net/http"

	"github.com/libsv/bitcoin-hc/service"
	router "github.com/libsv/bitcoin-hc/transports/http/endpoints/routes"

	"github.com/gin-gonic/gin"
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
		tip.GET("/tip", h.getTip)
		tip.GET("/tips", h.getTips)
		tip.GET("/tips/prune/:hash", h.pruneTip)
	}
}

// GetTips godoc.
//
//	@Summary Gets all tips
//	@Tags tip
//	@Accept */*
//	@Produce json
//	@Success 200 {array} []tips.tipStateResponse
//	@Router /chain/tips [get]
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
//	@Success 200 {object} tips.tipStateResponse
//	@Router /chain/tip [get]
//	@Security Bearer
func (h *handler) getTip(c *gin.Context) {
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
//	@Router /chain/tips/prune/{hash} [get]
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
