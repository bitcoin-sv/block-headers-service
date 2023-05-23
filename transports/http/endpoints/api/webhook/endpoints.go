package webhook

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/notification"
	"github.com/libsv/bitcoin-hc/service"
	router "github.com/libsv/bitcoin-hc/transports/http/endpoints/routes"
)

// Webhooks is an interface which represents methods required for Webhooks service.
type Webhooks interface {
	CreateWebhook(authType, header, token, url string) (*notification.Webhook, error)
	DeleteWebhook(value string) error
	GetWebhookByUrl(url string) (*notification.Webhook, error)
}

type handler struct {
	service Webhooks
}

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.ApiEndpoints {
	return &handler{service: s.Webhooks}
}

// RegisterApiEndpoints registers routes that are part of service API.
func (h *handler) RegisterApiEndpoints(router *gin.RouterGroup) {
	webhooks := router.Group("/webhook")
	{
		webhooks.POST("", h.registerWebhook)
		webhooks.GET("", h.getWebhook)
		webhooks.DELETE("", h.revokeWebhook)
	}
}

// nolint: godot
// registerWebhook godoc.
//
//	@Summary Register new webhook
//	@Tags webhooks
//	@Accept json
//	@Produce json
//	@Success 200 {object} domains.Webhook
//	@Router /webhook [post]
//	@Param data body webhook.WebhookRequest true "Webhook to register"
//
// @Security Bearer
func (h *handler) registerWebhook(c *gin.Context) {
	var reqBody WebhookRequest
	err := c.Bind(&reqBody)

	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}

	if reqBody.Url == "" {
		c.JSON(http.StatusBadRequest, "Url is required")
		return
	}

	webhook, err := h.service.CreateWebhook(reqBody.RequiredAuth.Type, reqBody.RequiredAuth.Header, reqBody.RequiredAuth.Token, reqBody.Url)
	if err == nil {
		c.JSON(http.StatusOK, webhook)
	} else if webhook == nil {
		c.JSON(http.StatusOK, err.Error())
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
func (h *handler) getWebhook(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, "Url param is required")
		return
	}
	w, err := h.service.GetWebhookByUrl(url)

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
func (h *handler) revokeWebhook(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, "Url param is required")
		return
	}
	err := h.service.DeleteWebhook(url)

	if err == nil {
		c.JSON(http.StatusOK, "Webhook revoked")
	} else {
		c.JSON(http.StatusBadRequest, err.Error())
	}
}
