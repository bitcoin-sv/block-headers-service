package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/domains"
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

		// If custom header is specified, use it, otherwise use default
		if strings.ToLower(reqBody.RequiredAuth.Type) == "custom_header" {
			tHeader = reqBody.RequiredAuth.Header
			token = reqBody.RequiredAuth.Token
		} else {
			tHeader = "Authorization"
			token = "Bearer " + reqBody.RequiredAuth.Token
		}

		webhook, err := h.services.Webhooks.GenerateWebhook(reqBody.Url, tHeader, token)
		fmt.Println("ERROR: ", err)
		if err == nil {
			fmt.Println("WEBHOOK2: ", webhook)
			c.JSON(http.StatusOK, webhook)
		} else if webhook == nil {
			c.JSON(http.StatusOK, err.Error())
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
	err := h.services.Webhooks.DeleteWebhook(value)

	if err == nil {
		c.JSON(http.StatusOK, "Webhook revoked")
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

func (h *Handler) notify(c *gin.Context) {
	bh, _ := h.services.Headers.GetHeaderByHash("000000007bc154e0fa7ea32218a72fe2c1bb9f86cf8c9ebf9a715ed27fdb229a")
	err := h.services.Webhooks.NotifyWebhooks(bh)

	if err == nil {
		c.JSON(http.StatusOK, "Webhook revoked")
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

func (h *Handler) newHeader(c *gin.Context) {
	var bh domains.BlockHeader
	err := c.Bind(&bh)
	if err == nil {
		c.JSON(http.StatusOK, "Webhook updated")
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

func (h *Handler) initRegisteredWehooksRoutes(router *gin.RouterGroup) {
	webhooks := router.Group("")
	{
		webhooks.POST("/webhook", h.registerWebhook)
		webhooks.DELETE("/webhook/:value", h.revokeWebhook)
		webhooks.GET("/webhook/notify", h.notify)
		webhooks.POST("/webhook/new-header", h.newHeader)
	}
}
