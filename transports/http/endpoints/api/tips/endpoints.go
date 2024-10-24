package tips

import (
	"net/http"

	"github.com/bitcoin-sv/block-headers-service/bhserrors"
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/service"
	router "github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type handler struct {
	service service.Headers
	log     *zerolog.Logger
}

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.APIEndpoints {
	return &handler{service: s.Headers, log: s.Logger}
}

// RegisterAPIEndpoints registers routes that are part of service API.
func (h *handler) RegisterAPIEndpoints(router *gin.RouterGroup, _ *config.HTTPConfig) {
	tip := router.Group("/chain")
	{
		tip.GET("/tip", h.getTips)
		tip.GET("/tip/longest", h.getTipLongestChain)
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
		bhserrors.ErrorResponse(c, err, h.log)
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
