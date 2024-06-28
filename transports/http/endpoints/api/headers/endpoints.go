package headers

import (
	"net/http"
	"strconv"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/service"
	router "github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/routes"
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
func (h *handler) RegisterApiEndpoints(router *gin.RouterGroup, cfg *config.HTTPConfig) {
	headers := router.Group("/chain/header")
	{
		headers.GET("/:hash", h.getHeaderByHash)
		headers.GET("/byHeight", h.getHeaderByHeight)
		headers.GET("/:hash/:ancestorHash/ancestor", h.getHeaderAncestorsByHash)
		headers.POST("/commonAncestor", h.getCommonAncestor)
		headers.GET("/state/:hash", h.getHeadersState)
	}
}

// getHeaderByHash godoc.
//
//		@Summary Gets header by hash
//		@Tags headers
//		@Accept */*
//		@Success 200 {object} BlockHeaderResponse
//		@Produce json
//		@Router /chain/header/{hash} [get]
//		@Param hash path string true "Requested Header Hash"
//	 @Security Bearer
func (h *handler) getHeaderByHash(c *gin.Context) {
	hash := c.Param("hash")
	bh, err := h.service.GetHeaderByHash(hash)

	if err == nil {
		c.JSON(http.StatusOK, newBlockHeaderResponse(bh))
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// getHeaderByHeight godoc.
//
//		@Summary Gets header by height
//		@Tags headers
//		@Accept */*
//		@Produce json
//		@Success 200 {object} []BlockHeaderResponse
//		@Router /chain/header/byHeight [get]
//		@Param height query int true "Height to start from"
//		@Param count query int false "Headers count (optional)"
//	 @Security Bearer
func (h *handler) getHeaderByHeight(c *gin.Context) {
	height, _ := c.GetQuery("height")
	count, _ := c.GetQuery("count")
	heightInt, err := strconv.Atoi(height)
	countInt, err2 := strconv.Atoi(count)

	if err == nil {
		if err2 != nil {
			countInt = 1
		}
		bh, err := h.service.GetHeadersByHeight(heightInt, countInt)
		if err == nil {
			c.JSON(http.StatusOK, mapToBlockHeadersResponses(bh))
		} else {
			c.JSON(http.StatusBadRequest, err.Error())
		}
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// getHeaderAncestorsByHash godoc.
//
//		@Summary Gets header ancestors
//		@Tags headers
//		@Accept */*
//		@Produce json
//		@Success 200 {object} []BlockHeaderResponse
//		@Router /chain/header/{hash}/{ancestorHash}/ancestor [get]
//		@Param hash path string true "Requested Header Hash"
//		@Param ancestorHash path string true "Ancestor Header Hash"
//	 @Security Bearer
func (h *handler) getHeaderAncestorsByHash(c *gin.Context) {
	hash := c.Param("hash")
	ancestorHash := c.Param("ancestorHash")
	ancestors, err := h.service.GetHeaderAncestorsByHash(hash, ancestorHash)

	if err == nil {
		c.JSON(http.StatusOK, mapToBlockHeadersResponses(ancestors))
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// getCommonAncestors godoc.
//
//		@Summary Gets common ancestors
//		@Tags headers
//		@Accept */*
//		@Produce json
//		@Success 200 {object} BlockHeaderResponse
//		@Router /chain/header/commonAncestor [post]
//		@Param ancesstors body []string true "JSON"
//	 @Security Bearer
func (h *handler) getCommonAncestor(c *gin.Context) {
	var body []string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	} else {
		ancestor, err := h.service.GetCommonAncestor(body)

		if err == nil {
			c.JSON(http.StatusOK, newBlockHeaderResponse(ancestor))
		} else {
			c.JSON(http.StatusBadRequest, err.Error())
		}
	}
}

// getHeadersState godoc.
//
//		@Summary Gets header state
//		@Tags headers
//		@Accept */*
//		@Produce json
//		@Success 200 {object} BlockHeaderStateResponse
//		@Router /chain/header/state/{hash} [get]
//		@Param hash path string true "Requested Header Hash"
//	 @Security Bearer
func (h *handler) getHeadersState(c *gin.Context) {
	hash := c.Param("hash")
	bh, err := h.service.GetHeaderByHash(hash)

	if err == nil {
		headerStateResponse := newBlockHeaderStateResponse(bh)
		c.JSON(http.StatusOK, headerStateResponse)
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}
