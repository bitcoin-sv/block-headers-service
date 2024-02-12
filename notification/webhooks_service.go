package notification

import (
	"errors"
	"strings"

	"github.com/rs/zerolog"

	"github.com/bitcoin-sv/block-headers-service/config"
)

// WebhooksService represents Webhooks service and provide access to repositories.
type WebhooksService struct {
	webhooks Webhooks
	client   WebhookTargetClient
	log      *zerolog.Logger
	cfg      *config.WebhookConfig
}

// NewWebhooksService creates and returns WebhooksService instance.
func NewWebhooksService(repo Webhooks, client WebhookTargetClient, log *zerolog.Logger, cfg *config.WebhookConfig) *WebhooksService {
	webhhoksLogger := log.With().Str("service", "webhooks").Logger()
	return &WebhooksService{
		webhooks: repo,
		client:   client,
		log:      &webhhoksLogger,
		cfg:      cfg,
	}
}

// CreateWebhook creates and save new webhook.
func (s *WebhooksService) CreateWebhook(authType, header, token, url string) (*Webhook, error) {
	// If custom header is specified, use it, otherwise use default
	if strings.ToLower(authType) == "bearer" {
		header = "Authorization"
		token = "Bearer " + token
	}

	webhook := CreateWebhook(url, header, token, s.cfg.MaxTries)

	err := s.webhooks.AddWebhookToDatabase(webhook)
	if err != nil {
		return s.refreshWebhook(url)
	}
	return webhook, nil
}

// DeleteWebhook deletes webhook by name or url.
func (s *WebhooksService) DeleteWebhook(value string) error {
	// Try to get and delete webhook by url
	_, err := s.webhooks.GetWebhookByUrl(value)
	if err == nil {
		err = s.webhooks.DeleteWebhookByUrl(value)
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

// Notify notifies all active webhooks.
func (s *WebhooksService) Notify(event Event) {
	webhooks, err := s.webhooks.GetAllWebhooks()

	if err != nil {
		s.log.Error().Msgf("Cannot load webhooks to notify. %v", err)
		return
	}

	// Notify all active webhooks
	for _, webhook := range webhooks {
		if webhook.Active {
			if err := webhook.Notify(event, s.client); err != nil {
				s.log.Warn().Msgf("Error during notification of the webhook: %v", err)
			}

			if err := s.webhooks.UpdateWebhook(webhook); err != nil {
				s.log.Error().Msgf("Error has happened during updating webhook state: %v", err)
			}
		}
	}
}

// GetWebhookByUrl returns webhook by url.
func (s *WebhooksService) GetWebhookByUrl(url string) (*Webhook, error) {
	return s.webhooks.GetWebhookByUrl(url)
}

// refreshWebhook refresh webhook by resetting ErrorsCount and Active fields.
func (s *WebhooksService) refreshWebhook(url string) (*Webhook, error) {
	w, err := s.webhooks.GetWebhookByUrl(url)
	if err != nil {
		return nil, err
	}

	if w != nil && !w.Active {
		w.Active = true
		w.ErrorsCount = 0
		err = s.webhooks.UpdateWebhook(w)
		if err != nil {
			return nil, err
		}
		return w, nil
	}
	return nil, errors.New("webhook already exists and is active")
}
