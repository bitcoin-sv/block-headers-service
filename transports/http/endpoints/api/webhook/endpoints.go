package webhook

import (
	"net/http"

	"github.com/bitcoin-sv/block-headers-service/bhserrors"
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/notification"
	"github.com/bitcoin-sv/block-headers-service/service"
	router "github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/routes"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Webhooks is an interface which represents methods required for Webhooks service.
type Webhooks interface {
	CreateWebhook(authType, header, token, url string) (*notification.Webhook, error)
	DeleteWebhook(value string) error
	GetWebhookByURL(url string) (*notification.Webhook, error)
}

type handler struct {
	service Webhooks
	log     *zerolog.Logger
}

// NewHandler creates new endpoint handler.
func NewHandler(s *service.Services) router.APIEndpoints {
	return &handler{service: s.Webhooks, log: s.Logger}
}

// RegisterAPIEndpoints registers routes that are part of service API.
func (h *handler) RegisterAPIEndpoints(router *gin.RouterGroup, _ *config.HTTPConfig) {
	webhooks := router.Group("/webhook")
	{
		webhooks.POST("", h.registerWebhook)
		webhooks.GET("", h.getWebhook)
		webhooks.DELETE("", h.revokeWebhook)
	}
}

// registerWebhook godoc.
//
//	@Summary Register new webhook
//	@Tags webhooks
//	@Accept json
//	@Produce json
//	@Success 200 {object} notification.Webhook
//	@Router /webhook [post]
//	@Param data body webhook.Request true "Webhook to register"
//
// @Security Bearer
func (h *handler) registerWebhook(c *gin.Context) {
	var reqBody Request
	err := c.Bind(&reqBody)

	if err != nil {
		bhserrors.ErrorResponse(c, bhserrors.ErrBindBody.Wrap(err), h.log)
	}

	if reqBody.URL == "" {
		bhserrors.ErrorResponse(c, bhserrors.ErrURLBodyRequired, h.log)
		return
	}

	webhook, err := h.service.CreateWebhook(reqBody.RequiredAuth.Type, reqBody.RequiredAuth.Header, reqBody.RequiredAuth.Token, reqBody.URL)
	if err == nil {
		c.JSON(http.StatusOK, webhook)
	} else {
		bhserrors.ErrorResponse(c, err, h.log)
	}
}

// getWebhook godoc.
//
//	@Summary Get webhook
//	@Tags webhooks
//	@Accept */*
//	@Produce json
//	@Success 200 {object} notification.Webhook
//	@Router /webhook [get]
//	@Param url query string true "URL of webhook to check"
//
// @Security Bearer
func (h *handler) getWebhook(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		bhserrors.ErrorResponse(c, bhserrors.ErrURLParamRequired, h.log)
		return
	}
	w, err := h.service.GetWebhookByURL(url)

	if err == nil {
		c.JSON(http.StatusOK, w)
	} else {
		bhserrors.ErrorResponse(c, err, h.log)
	}
}

// revokeWebhook godoc.
//
//	@Summary Revoke webhook
//	@Tags webhooks
//	@Accept */*
//	@Success 200
//	@Produce json
//	@Router /webhook [delete]
//	@Param url query string true "URL of webhook to revoke"
//
// @Security Bearer
func (h *handler) revokeWebhook(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		bhserrors.ErrorResponse(c, bhserrors.ErrURLParamRequired, h.log)
		return
	}
	err := h.service.DeleteWebhook(url)

	if err == nil {
		c.JSON(http.StatusOK, "Webhook revoked")
	} else {
		bhserrors.ErrorResponse(c, err, h.log)
	}
}
