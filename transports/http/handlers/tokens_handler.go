package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// createToken godoc.
//
//	@Summary Creates new token
//	@Tags access
//	@Accept */*
//	@Produce json
//	@Success 200 {object} domains.Token
//	@Router /access [get]
//  @Security Bearer
func (h *Handler) createToken(c *gin.Context) {
	bh, err := h.services.Tokens.GenerateToken()

	if err == nil {
		c.JSON(http.StatusOK, bh)
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// revokeToken godoc.
//
//	@Summary Gets header state
//	@Tags access
//	@Accept */*
//	@Success 200
//	@Produce json
//	@Router /access/{token} [delete]
//	@Param token path string true "Token to delete"
//  @Security Bearer
func (h *Handler) revokeToken(c *gin.Context) {
	token := c.Param("token")
	fmt.Println("token", token)
	err := h.services.Tokens.DeleteToken(token)

	if err == nil {
		c.JSON(http.StatusOK, "Token revoked")
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

func (h *Handler) initAccessRoutes(router *gin.RouterGroup) {
	tokens := router.Group("")
	{
		tokens.GET("/access", h.createToken)
		tokens.DELETE("/access/:token", h.revokeToken)
	}
}
