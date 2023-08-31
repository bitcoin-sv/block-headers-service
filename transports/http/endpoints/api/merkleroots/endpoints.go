package merkleroots

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/libsv/bitcoin-hc/service"
	router "github.com/libsv/bitcoin-hc/transports/http/endpoints/routes"
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
	merkle := router.Group("/chain/merkleroot")
	{
		merkle.POST("/verify", h.verify)
	}
}

// Verify godoc.
//
//	@Summary Verifies Merkle roots inclusion in the longest chain
//	@Tags merkleroots
//	@Accept */*
//	@Produce json
//	@Success 200 {array} merkleroots.merkleRootsConfirmationsResponse
//	@Router /chain/merkleroots/verify [post]
//	@Param merkleroots body []string true "JSON"
//	@Security Bearer
func (h *handler) verify(c *gin.Context) {
	var body []string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, errors.New("At least one merkleroot is required"))
		return
	}

	mrcs, err := h.service.GetMerkleRootsConfirmations(body)

	if err == nil {
		c.JSON(http.StatusOK, mapToMerkleRootsConfirmationsResponses(mrcs))
	} else {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
}
