package testrepository

import (
	"fmt"

	"github.com/bitcoin-sv/block-headers-service/bhserrors"
	"github.com/bitcoin-sv/block-headers-service/notification"
)

// WebhooksTestRepository in memory WebhooksRepository representation for unit testing.
type WebhooksTestRepository struct {
	db *[]notification.Webhook
}

// AddWebhookToDatabase adds new webhook to db.
func (r *WebhooksTestRepository) AddWebhookToDatabase(webhook *notification.Webhook) error {
	for _, w := range *r.db {
		if w.URL == webhook.URL {
			return fmt.Errorf("webhook with url %s already exists", webhook.URL)
		}
	}
	*r.db = append(*r.db, *webhook)
	return nil
}

// DeleteWebhookByURL deletes webhook by url from db.
func (r *WebhooksTestRepository) DeleteWebhookByURL(url string) error {
	for i, w := range *r.db {
		if w.URL == url {
			arr := *r.db
			// Replace the element at index i with the last element.
			arr[i] = arr[len(arr)-1]
			// Assign slice without last element.
			*r.db = arr[:len(arr)-1]
			return nil
		}
	}
	return bhserrors.ErrWebhookNotFound
}

// GetWebhookByURL returns webhook from db by given url.
func (r *WebhooksTestRepository) GetWebhookByURL(_ string) (*notification.Webhook, error) {
	return nil, nil
}

// GetAllWebhooks returns all webhooks from db.
func (r *WebhooksTestRepository) GetAllWebhooks() ([]*notification.Webhook, error) {
	return nil, nil
}

// UpdateWebhook updates webhook in db.
func (r *WebhooksTestRepository) UpdateWebhook(_ *notification.Webhook) error {
	return nil
}

// NewWebhooksTestRepository constructor for WebhooksTestRepository.
func NewWebhooksTestRepository(db *[]notification.Webhook) *WebhooksTestRepository {
	return &WebhooksTestRepository{
		db: db,
	}
}
