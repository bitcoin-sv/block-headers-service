package testrepository

import (
	"errors"
	"fmt"

	"github.com/libsv/bitcoin-hc/domains"
)

// WebhooksTestRepository in memory WebhooksRepository representation for unit testing.
type WebhooksTestRepository struct {
	db *[]domains.Webhook
}

// AddWebhookToDatabase adds new webhook to db.
func (r *WebhooksTestRepository) AddWebhookToDatabase(webhook *domains.Webhook) error {
	for _, w := range *r.db {
		if w.Url == webhook.Url {
			return fmt.Errorf("webhook with url %s already exists", webhook.Url)
		}
	}
	*r.db = append(*r.db, *webhook)
	return nil
}

// DeleteWebhookByUrl deletes webhook by url from db.
func (r *WebhooksTestRepository) DeleteWebhookByUrl(url string) error {
	for i, w := range *r.db {
		if w.Url == url {
			arr := *r.db
			// Replace the element at index i with the last element.
			arr[i] = arr[len(arr)-1]
			// Assign slice without last element.
			*r.db = arr[:len(arr)-1]
			return nil
		}
	}
	return errors.New("could not find webhook")
}

// GetWebhookByUrl returns webhook from db by given url.
func (r *WebhooksTestRepository) GetWebhookByUrl(url string) (*domains.Webhook, error) {
	return nil, nil
}

// GetAllWebhooks returns all webhooks from db.
func (r *WebhooksTestRepository) GetAllWebhooks() ([]*domains.Webhook, error) {
	return nil, nil
}

// UpdateWebhook updates webhook in db.
func (r *WebhooksTestRepository) UpdateWebhook(w *domains.Webhook) error {
	return nil
}

// NewWebhooksTestRepository constructor for WebhooksTestRepository.
func NewWebhooksTestRepository(db *[]domains.Webhook) *WebhooksTestRepository {
	return &WebhooksTestRepository{
		db: db,
	}
}
