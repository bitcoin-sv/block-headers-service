package repository

import (
	"context"

	"github.com/bitcoin-sv/block-headers-service/database/sql"
	"github.com/bitcoin-sv/block-headers-service/notification"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
)

// WebhooksRepository provide access to repositories and implements methods for webhooks.
type WebhooksRepository struct {
	db *sql.HeadersDb
}

// AddWebhookToDatabase adds new webhook to db.
func (r *WebhooksRepository) AddWebhookToDatabase(rWebhook *notification.Webhook) error {
	dbWebhook := dto.ToDbWebhook(rWebhook)
	err := r.db.CreateWebhook(context.Background(), dbWebhook)
	return err
}

// DeleteWebhookByUrl deletes webhook by url from db.
func (r *WebhooksRepository) DeleteWebhookByUrl(url string) error {
	err := r.db.DeleteWebhookByUrl(context.Background(), url)
	return err
}

// GetWebhookByUrl returns webhook from db by given url.
func (r *WebhooksRepository) GetWebhookByUrl(url string) (*notification.Webhook, error) {
	w, err := r.db.GetWebhookByUrl(context.Background(), url)
	if err != nil {
		return nil, err
	}
	dbw := w.ToWebhook()
	return dbw, err
}

// GetAllWebhooks returns all webhooks from db.
func (r *WebhooksRepository) GetAllWebhooks() ([]*notification.Webhook, error) {
	webhooks, err := r.db.GetAllWebhooks(context.Background())
	if err != nil {
		return nil, err
	}
	dbWebhooks := make([]*notification.Webhook, 0)
	for _, w := range webhooks {
		dbw := w.ToWebhook()
		dbWebhooks = append(dbWebhooks, dbw)
	}
	return dbWebhooks, err
}

// UpdateWebhook updates webhook in db.
func (r *WebhooksRepository) UpdateWebhook(w *notification.Webhook) error {
	err := r.db.UpdateWebhook(context.Background(), w.Url, w.LastEmitTimestamp, w.LastEmitStatus, w.ErrorsCount, w.Active)
	return err
}

// NewWebhooksRepository creates and returns WebhooksRepository instance.
func NewWebhooksRepository(db *sql.HeadersDb) *WebhooksRepository {
	return &WebhooksRepository{db: db}
}
