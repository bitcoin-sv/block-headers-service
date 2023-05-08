package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	webhook "github.com/libsv/bitcoin-hc/transports/http"
)

// registerWebhook godoc.
//
//	@Summary Register new webhook
//	@Tags webshooks
//	@Accept */*
//	@Produce json
//	@Success 200 {object} domains.Webhook
//	@Router /webhook [post]
//
// @Param body body requestBody true "Webhook to register"
// @Security Bearer
func (h *Handler) registerWebhook(c *gin.Context) {
	var reqBody webhook.WebhookRequest
	err := c.Bind(&reqBody)

	if err == nil {
		var tHeader, token string
		if strings.ToLower(reqBody.RequiredAuth.Type) == "custom_header" {
			tHeader = reqBody.RequiredAuth.Header
		} else {
			tHeader = "Authorization"
		}

		token = reqBody.RequiredAuth.Token
		webhook, err := h.services.Webhooks.GenerateWebhook(reqBody.Name, reqBody.Url, tHeader, token)
		if err == nil {
			c.JSON(http.StatusOK, webhook)
		}
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// revokeWebhook godoc.
//
//	@Summary Revoke webhook
//	@Tags webhooks
//	@Accept */*
//	@Success 200
//	@Produce json
//	@Router /webhook/{value} [delete]
//	@Param value path string true "Name or url of webhook to revoke"

// @Security Bearer
func (h *Handler) revokeWebhook(c *gin.Context) {
	value := c.Param("value")
	fmt.Println("value", value)
	err := h.services.Webhooks.DeleteWebhook(value)

	if err == nil {
		c.JSON(http.StatusOK, "Webhook revoked")
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

func (h *Handler) initRegisteredWehooksRoutes(router *gin.RouterGroup) {
	webhooks := router.Group("")
	{
		webhooks.POST("/webhook", h.registerWebhook)
		webhooks.DELETE("/webhook/:value", h.revokeWebhook)
	}
}
