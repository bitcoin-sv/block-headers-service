package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getHeaderByHash godoc.
//  @Summary Gets header by hash
//  @Tags headers
//  @Accept */*
//  @Success 200 {object} headers.BlockHeader
//  @Produce json
//  @Router /chain/header/{hash} [get]
//  @Param hash path string true "Requested Header Hash"
func (h *Handler) getHeaderByHash(c *gin.Context) {
	hash := c.Param("hash")
	header, err := h.services.Headers.GetHeaderByHash(hash)

	if err == nil {
		c.JSON(http.StatusOK, header)
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// getHeaderByHeight godoc.
//  @Summary Gets header by height
//  @Tags headers
//  @Accept */*
//  @Produce json
//  @Success 200 {object} []headers.BlockHeader
//  @Router /chain/header/byHeight [get]
//  @Param height query int true "Height to start from"
//  @Param count query int false "Headers count (optional)"
func (h *Handler) getHeaderByHeight(c *gin.Context) {
	height, _ := c.GetQuery("height")
	count, _ := c.GetQuery("count")
	heightInt, err := strconv.Atoi(height)
	countInt, err2 := strconv.Atoi(count)

	if err == nil {
		if err2 != nil {
			countInt = 1
		}
		header, err := h.services.Headers.GetHeadersByHeight(heightInt, countInt)
		if err == nil {
			c.JSON(http.StatusOK, header)
		} else {
			c.JSON(http.StatusBadRequest, err.Error())
		}
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// getHeaderAncestorsByHash godoc.
//  @Summary Gets header ancestors
//  @Tags headers
//  @Accept */*
//  @Produce json
//  @Success 200 {object} []headers.BlockHeader
//  @Router /chain/header/{hash}/{ancestorHash}/ancestors [get]
//  @Param hash path string true "Requested Header Hash"
//  @Param ancestorHash path string true "Ancestor Header Hash"
func (h *Handler) getHeaderAncestorsByHash(c *gin.Context) {
	hash := c.Param("hash")
	ancestorHash := c.Param("ancestorHash")
	ancestors, err := h.services.Headers.GetHeaderAncestorsByHash(hash, ancestorHash)

	if err == nil {
		c.JSON(http.StatusOK, ancestors)
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// getCommonAncestors godoc.
//  @Summary Gets common ancestors
//  @Tags headers
//  @Accept */*
//  @Produce json
//  @Success 200 {object} headers.BlockHeader
//  @Router /chain/header/commonAncestor [post]
//  @Param ancesstors body []string true "JSON"
func (h *Handler) getCommonAncestors(c *gin.Context) {
	var body []string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	} else {
		ancestors, err := h.services.Headers.GetCommonAncestors(body)

		if err == nil {
			c.JSON(http.StatusOK, ancestors)
		} else {
			c.JSON(http.StatusBadRequest, err.Error())
		}
	}
}

// getHeadersState godoc.
//  @Summary Gets header state
//  @Tags headers
//  @Accept */*
//  @Produce json
//  @Success 200 {object} headers.BlockHeaderState
//  @Router /chain/header/state/{hash} [get]
//  @Param hash path string true "Requested Header Hash"
func (h *Handler) getHeadersState(c *gin.Context) {
	hash := c.Param("hash")
	state, err := h.services.Headers.GetHeadersState(hash)

	if err == nil {
		c.JSON(http.StatusOK, state)
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

func (h *Handler) initHeadersRoutes(router *gin.RouterGroup) {
	headers := router.Group("/chain/header")
	{
		headers.GET("/:hash", h.getHeaderByHash)
		headers.GET("/byHeight", h.getHeaderByHeight)
		headers.GET("/:hash/:ancestorHash/ancestors", h.getHeaderAncestorsByHash)
		headers.POST("/commonAncestor", h.getCommonAncestors)
		headers.GET("/state/:hash", h.getHeadersState)
	}
}
