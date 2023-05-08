package service

import (
	"fmt"

	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/repository"
)

// WebhooksService represents Webhooks service and provide access to repositories.
type WebhooksService struct {
	repo *repository.Repositories
}

// GenerateWebhook generates and save new webhook.
func (s *WebhooksService) GenerateWebhook(name, url, tHeader, token string) (*domains.Webhook, error) {
	webhook := domains.CreateWebhook(name, url, tHeader, token)
	err := s.repo.Webhooks.AddWebhookToDatabase(webhook)
	fmt.Println(webhook)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	return webhook, nil
}

// DeleteWebhook deletes webhook by name or url.
func (s *WebhooksService) DeleteWebhook(value string) error {
	// //Try to get and delete webhook by name
	_, err := s.repo.Webhooks.GetWebhookByName(value)
	fmt.Println("Error by by name", err)
	if err == nil {
		err = s.repo.Webhooks.DeleteWebhookByName(value)
		if err != nil {
			return err
		}
		return nil
	}

	// If webhook not found by name, try to get and delete webhook by url
	_, err = s.repo.Webhooks.GetWebhookByUrl(value)
	fmt.Println("Error by by url", err)
	if err == nil {
		err = s.repo.Webhooks.DeleteWebhookByUrl(value)
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

// NewWebhooksService creates and returns WebhooksService instance.
func NewWebhooksService(repo *repository.Repositories) *WebhooksService {
	return &WebhooksService{repo: repo}
}
