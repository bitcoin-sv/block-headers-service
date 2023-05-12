package access

import (
	"net/http"

	"github.com/libsv/bitcoin-hc/service"
	"github.com/libsv/bitcoin-hc/transports/http/auth"
	router "github.com/libsv/bitcoin-hc/transports/http/endpoints/routes"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service service.Tokens
}

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.ApiEndpoints {
	return &handler{service: s.Tokens}
}

// RegisterApiEndpoints registers routes that are part of service API.
func (h *handler) RegisterApiEndpoints(router *gin.RouterGroup) {
	tokens := router.Group("/access")
	{
		tokens.GET("", h.getToken)
		tokens.POST("", auth.RequireAdmin(h.createToken))
		tokens.DELETE("/:token", h.revokeToken)
	}
}

// getToken godoc.
//
//		@Summary Get information about token
//		@Tags access
//		@Accept */*
//		@Produce json
//		@Success 200 {object} domains.Token
//		@Router /access [get]
//	 @Security Bearer
func (h *handler) getToken(c *gin.Context) {
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
func (h *handler) createToken(c *gin.Context) {
	bh, err := h.service.GenerateToken()

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
func (h *handler) revokeToken(c *gin.Context) {
	token := c.Param("token")
	err := h.service.DeleteToken(token)

	if err == nil {
		c.JSON(http.StatusOK, "Token revoked")
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}
