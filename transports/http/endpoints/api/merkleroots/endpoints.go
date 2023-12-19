package merkleroots

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/domains"
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
func (h *handler) RegisterApiEndpoints(router *gin.RouterGroup, cfg *config.HTTPConfig) {
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
//	@Success 200 {array} merkleroots.MerkleRootsConfirmationsResponse
//	@Router /chain/merkleroot/verify [post]
//	@Param request body []domains.MerkleRootConfirmationRequestItem true "JSON"
//	@Security Bearer
func (h *handler) verify(c *gin.Context) {
	var body []domains.MerkleRootConfirmationRequestItem
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, errors.New("at least one merkleroot is required").Error())
		return
	}

	mrcs, err := h.service.GetMerkleRootsConfirmations(body)

	if err == nil {
		c.JSON(http.StatusOK, mapToMerkleRootsConfirmationsResponses(mrcs))
	} else {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
}
