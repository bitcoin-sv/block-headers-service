package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	webhook "github.com/libsv/bitcoin-hc/transports/http"
)

// nolint: godot
// registerWebhook godoc.
//
//	@Summary Register new webhook
//	@Tags webhooks
//	@Accept json
//	@Produce json
//	@Success 200 {object} domains.Webhook
//	@Router /webhook [post]
//	@Param data body http.WebhookRequest true "Webhook to register"
//
// @Security Bearer
func (h *Handler) registerWebhook(c *gin.Context) {
	var reqBody webhook.WebhookRequest
	err := c.Bind(&reqBody)

	if err == nil {
		if reqBody.Url == "" {
			c.JSON(http.StatusBadRequest, "Url is required")
			return
		}

		webhook, err := h.services.Webhooks.CreateWebhook(reqBody)
		if err == nil {
			c.JSON(http.StatusOK, webhook)
		} else if webhook == nil {
			c.JSON(http.StatusOK, err.Error())
		}
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// nolint: godot
// getWebhook godoc.
//
//	@Summary Get webhook
//	@Tags webhooks
//	@Accept */*
//	@Produce json
//	@Success 200 {object} domains.Webhook
//	@Router /webhook [get]
//	@Param url query string true "Url of webhook to check"
//
// @Security Bearer
func (h *Handler) getWebhook(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, "Url param is required")
		return
	}
	w, err := h.services.Webhooks.GetWebhookByUrl(url)

	if err == nil {
		c.JSON(http.StatusOK, w)
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}

// nolint: godot
// revokeWebhook godoc.
//
//	@Summary Revoke webhook
//	@Tags webhooks
//	@Accept */*
//	@Success 200
//	@Produce json
//	@Router /webhook [delete]
//	@Param url query string true "Url of webhook to revoke"
//
// @Security Bearer
func (h *Handler) revokeWebhook(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, "Url param is required")
		return
	}
	err := h.services.Webhooks.DeleteWebhook(url)

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
		webhooks.GET("/webhook", h.getWebhook)
		webhooks.DELETE("/webhook", h.revokeWebhook)
	}
}
