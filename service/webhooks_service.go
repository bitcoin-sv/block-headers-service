package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/repository"
	"github.com/spf13/viper"
)

// WebhookMaxTries is the maximum number of times a webhook will be retried.
const WebhookMaxTries = "webhook.maxTries"

// WebhooksService represents Webhooks service and provide access to repositories.
type WebhooksService struct {
	repo *repository.Repositories
}

// GenerateWebhook generates and save new webhook.
func (s *WebhooksService) GenerateWebhook(url, tHeader, token string) (*domains.Webhook, error) {
	webhook := domains.CreateWebhook(url, tHeader, token)
	err := s.repo.Webhooks.AddWebhookToDatabase(webhook)
	if err != nil {
		w, _ := s.repo.Webhooks.GetWebhookByUrl(url)
		if w != nil && !w.Active {
			err = s.repo.Webhooks.UpdateWebhook(w, w.LastEmitTimestamp, w.LastEmitStatus, 0, true)
			w.Active = true
			w.ErrorsCount = 0
			return w, err
		}
		return nil, errors.New("webhook already exists and is active")
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
			var lastEmitStatus string
			var err error
			timestamp := time.Now()
			statusCode, body, err := webhook.Notify(h)

			if err != nil {
				lastEmitStatus = fmt.Sprint(err)
			} else {
				lastEmitStatus = fmt.Sprint(statusCode, " ", body)
			}

			// If status code is not 200, increment errors count and set active to false if errors count is more than max tries
			if statusCode != http.StatusOK {
				errorsCount := webhook.ErrorsCount + 1
				active := true

				if errorsCount >= viper.GetInt(WebhookMaxTries) {
					active = false
				}

				err = s.repo.Webhooks.UpdateWebhook(webhook, timestamp, lastEmitStatus, errorsCount, active)
			} else {
				// If status code is 200, reset errors count and set active to true
				err = s.repo.Webhooks.UpdateWebhook(webhook, timestamp, lastEmitStatus, 0, true)
			}

			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NewWebhooksService creates and returns WebhooksService instance.
func NewWebhooksService(repo *repository.Repositories) *WebhooksService {
	return &WebhooksService{repo: repo}
}
