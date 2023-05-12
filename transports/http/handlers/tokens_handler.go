package handler

import (
	"github.com/libsv/bitcoin-hc/transports/http/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

// getToken godoc.
//
//		@Summary Get information about token
//		@Tags access
//		@Accept */*
//		@Produce json
//		@Success 200 {object} domains.Token
//		@Router /access [get]
//	 @Security Bearer
func (h *Handler) getToken(c *gin.Context) {
	t, exists := c.Get("token")

	if exists {
		c.JSON(http.StatusOK, t)
	} else {
		c.Status(http.StatusBadRequest)
	}
}

// createToken godoc.
//
//		@Summary Creates new token
//		@Tags access
//		@Accept */*
//		@Produce json
//		@Success 200 {object} domains.Token
//		@Router /access [post]
//	 @Security Bearer
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
//		@Summary Gets header state
//		@Tags access
//		@Accept */*
//		@Success 200
//		@Produce json
//		@Router /access/{token} [delete]
//		@Param token path string true "Token to delete"
//	 @Security Bearer
func (h *Handler) revokeToken(c *gin.Context) {
	token := c.Param("token")
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
		tokens.GET("/access", h.getToken)
		tokens.POST("/access", auth.RequireAdmin(h.createToken))
		tokens.DELETE("/access/:token", h.revokeToken)
	}
}
