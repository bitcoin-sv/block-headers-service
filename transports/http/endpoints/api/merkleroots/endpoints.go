package merkleroots

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/service"
	router "github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
)

const (
	// batchSize is the size of returned merkleroots per request
	batchSize = "2000"
	// lastEvaluatedKey is the last block height that the client proccessed, by default -1 implicating start of the chain
	lastEvaluatedKey = "-1"
)

type handler struct {
	service service.Merkleroots
}

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.APIEndpoints {
	return &handler{service: s.Merkleroots}
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
// @Summary Returns merkleroots for the specified range
// @Tags merkleroots
// @Accept */*
// @Produce json
// @Success 200 {object} merkleroots.MerkleRootsESKPagedResponse.//TODO: map this in swagger docs
// @Router /chain/merkleroot [get]
// @Param batchSize query string false "Batch size of returned merkleroots"
// @Param lastEvaluatedKey query string false "Last evaluated block height that client has processed"
// @Security Bearer
func (h *handler) merkleroots(c *gin.Context) {
	batchSize := c.DefaultQuery("batchSize", batchSize)
	lastEvaluatedKey := c.DefaultQuery("lastEvaluatedKey", lastEvaluatedKey)

	batchSizeInt, errBatchSizeConv := strconv.Atoi(batchSize)
	lastEvaluatedKeyInt, errLastEvaluatedKeyConv := strconv.Atoi(lastEvaluatedKey)

	if errBatchSizeConv != nil {
		c.JSON(http.StatusBadRequest, errors.New("batchSize must be a numeric value").Error())
		return
	}

	if errLastEvaluatedKeyConv != nil {
		c.JSON(http.StatusBadRequest, errors.New("lastEvaluatedKey must be a numeric value").Error())
		return
	}

	merkleroots, err := h.service.GetMerkleRoots(batchSizeInt, lastEvaluatedKeyInt)

	if err == nil {
		c.JSON(http.StatusOK, merkleroots)
	} else {
		c.JSON(http.StatusInternalServerError, err.Error())
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
