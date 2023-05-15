package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/repository"
	"github.com/libsv/bitcoin-hc/transports/http/client"
)

// WebhookMaxTries is the maximum number of times a webhook will be retried.
const WebhookMaxTries = "webhook.maxTries"

// WebhooksService represents Webhooks service and provide access to repositories.
type WebhooksService struct {
	repo   *repository.Repositories
	client domains.WebhookTargetClient
}

// CreateWebhook creates and save new webhook.
func (s *WebhooksService) CreateWebhook(authType, header, token, url string) (*domains.Webhook, error) {
	// If custom header is specified, use it, otherwise use default
	if strings.ToLower(authType) == "bearer" {
		header = "Authorization"
		token = "Bearer " + token
	}

	webhook := domains.CreateWebhook(url, header, token)

	err := s.repo.Webhooks.AddWebhookToDatabase(webhook)
	if err != nil {
		return s.refreshWebhook(url)
	}
	return webhook, nil
}

// DeleteWebhook deletes webhook by name or url.
func (s *WebhooksService) DeleteWebhook(value string) error {
	// Try to get and delete webhook by url
	_, err := s.repo.Webhooks.GetWebhookByUrl(value)
	if err == nil {
		err = s.repo.Webhooks.DeleteWebhookByUrl(value)
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

// NotifyWebhooks notifies all active webhooks.
func (s *WebhooksService) NotifyWebhooks(h *domains.BlockHeader) error {
	webhooks, err := s.repo.Webhooks.GetAllWebhooks()

	if err != nil {
		return err
	}

	// Notify all active webhooks
	for _, webhook := range webhooks {
		if webhook.Active {
			err := webhook.Notify(h, s.client)

			if err != nil {
				fmt.Println(err)
			}

			err = s.repo.Webhooks.UpdateWebhook(webhook)

			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// GetWebhookByUrl returns webhook by url.
func (s *WebhooksService) GetWebhookByUrl(url string) (*domains.Webhook, error) {
	return s.repo.Webhooks.GetWebhookByUrl(url)
}

// refreshWebhook refresh webhook by resetting ErrorsCount and Active fields.
func (s *WebhooksService) refreshWebhook(url string) (*domains.Webhook, error) {
	w, err := s.repo.Webhooks.GetWebhookByUrl(url)
	if err != nil {
		return nil, err
	}

	if w != nil && !w.Active {
		w.Active = true
		w.ErrorsCount = 0
		err = s.repo.Webhooks.UpdateWebhook(w)
		return w, err
	}
	return nil, errors.New("webhook already exists and is active")
}

// NewWebhooksService creates and returns WebhooksService instance.
func NewWebhooksService(repo *repository.Repositories) *WebhooksService {
	return &WebhooksService{
		repo:   repo,
		client: client.NewWebhookTargetClient(),
	}
}
