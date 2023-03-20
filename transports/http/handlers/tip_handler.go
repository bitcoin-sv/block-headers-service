package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTips godoc.
//  @Summary Gets all tips
//  @Tags tip
//  @Accept */*
//  @Produce json
//  @Success 200 {object} []headers.BlockHeaderState
//  @Router /chain/tips [get]
func (h *Handler) getTips(c *gin.Context) {
	tips, err := h.services.Tip.GetTips()

	if err == nil {
		c.JSON(http.StatusOK, tips)
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// PruneTip godoc.
//  @Summary Prune tip
//  @Tags tip
//  @Accept */*
//  @Produce json
//  @Success 200 {object} string
//  @Router /chain/tips/prune/{hash} [get]
//  @Param hash path string true "Requested Header Hash"
func (h *Handler) pruneTip(c *gin.Context) {
	param := c.Param("hash")
	fmt.Println(param)
	tip, err := h.services.Tip.PruneTip()

	if err == nil {
		c.JSON(http.StatusOK, tip)
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

func (h *Handler) initTipRoutes(router *gin.RouterGroup) {
	tip := router.Group("/chain")
	{
		tip.GET("/tips", h.getTips)
		tip.GET("/tips/prune/:hash", h.pruneTip)
	}
}
