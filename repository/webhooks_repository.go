package repository

import (
	"context"
	"time"

	"github.com/libsv/bitcoin-hc/data/sql"
	"github.com/libsv/bitcoin-hc/domains"
	dto "github.com/libsv/bitcoin-hc/repository/dto"
)

// WebhooksRepository provide access to repositories and implements methods for webhooks.
type WebhooksRepository struct {
	db *sql.HeadersDb
}

// AddWebhookToDatabase adds new webhook to db.
func (r *WebhooksRepository) AddWebhookToDatabase(rWebhook *domains.Webhook) error {
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
func (r *WebhooksRepository) GetWebhookByUrl(url string) (*domains.Webhook, error) {
	w, err := r.db.GetWebhookByUrl(context.Background(), url)
	if err != nil {
		return nil, err
	}
	dbw := w.ToWebhook()
	return dbw, err
}

// GetAllWebhooks returns all webhooks from db.
func (r *WebhooksRepository) GetAllWebhooks() ([]*domains.Webhook, error) {
	webhooks, err := r.db.GetAllWebhooks(context.Background())
	if err != nil {
		return nil, err
	}
	dbWebhooks := make([]*domains.Webhook, 0)
	for _, w := range webhooks {
		dbw := w.ToWebhook()
		dbWebhooks = append(dbWebhooks, dbw)
	}
	return dbWebhooks, err
}

// UpdateWebhook updates webhook in db.
func (r *WebhooksRepository) UpdateWebhook(w *domains.Webhook, lastEmitTimestamp time.Time, lastEmitStatus string, errorsCount int, active bool) error {
	err := r.db.UpdateWebhook(context.Background(), w.Url, lastEmitTimestamp, lastEmitStatus, errorsCount, active)
	return err
}

// NewWebhooksRepository creates and returns WebhooksRepository instance.
func NewWebhooksRepository(db *sql.HeadersDb) *WebhooksRepository {
	return &WebhooksRepository{db: db}
}
