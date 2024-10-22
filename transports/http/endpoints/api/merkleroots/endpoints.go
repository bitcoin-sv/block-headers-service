package merkleroots

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bitcoin-sv/block-headers-service/bhserrors"
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/service"
	router "github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const (
	// defaultBatchSize is the size of returned merkleroots per request
	defaultBatchSize = "2000"
)

type handler struct {
	service service.Merkleroots
	log     *zerolog.Logger
}

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.APIEndpoints {
	return &handler{service: s.Merkleroots, log: s.Logger}
}

// RegisterAPIEndpoints registers routes that are part of service API.
func (h *handler) RegisterAPIEndpoints(router *gin.RouterGroup, _ *config.HTTPConfig) {
	merkle := router.Group("/chain/merkleroot")
	{
		merkle.POST("/verify", h.verify)
		merkle.GET("", h.merkleroots)
	}
}

// Merkleroot godoc.
//
// @Summary Returns merkleroots for the specified range.
// @Tags merkleroots
// @Accept */*
// @Produce json
// @Success 200 {object} domains.MerkleRootsESKPagedResponse
// @Router /chain/merkleroot [get]
// @Param batchSize query string false "Batch size of returned merkleroots"
// @Param lastEvaluatedKey query string false "Last evaluated merkleroot that client has processed"
// @Security Bearer
func (h *handler) merkleroots(c *gin.Context) {
	batchSize := c.DefaultQuery("batchSize", defaultBatchSize)
	lastEvaluatedKey := c.Query("lastEvaluatedKey")

	batchSizeInt, err := strconv.Atoi(batchSize)
	if err != nil || batchSizeInt < 0 {
		bhserrors.ErrorResponse(c, bhserrors.ErrMerklerootInvalidBatchSize.Wrap(err), h.log)
		return
	}

	merkleroots, err := h.service.GetMerkleRoots(batchSizeInt, lastEvaluatedKey)

	if err == nil {
		c.JSON(http.StatusOK, merkleroots)
	} else {
		bhserrors.ErrorResponse(c, err, h.log)
	}
}

// Verify godoc.
//
//	@Summary Verifies Merkle roots inclusion in the longest chain
//	@Tags merkleroots
//	@Accept */*
//	@Produce json
//	@Success 200 {array} merkleroots.ConfirmationsResponse
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
