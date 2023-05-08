package repository

import (
	"context"

	"github.com/libsv/bitcoin-hc/data/sql"
	"github.com/libsv/bitcoin-hc/domains"
	dto "github.com/libsv/bitcoin-hc/repository/dto"
)

// WebhooksRepository provide access to repositories and implements methods for webhooks.
type WebhooksRepository struct {
	db *sql.HeadersDb
}

// AddWebhooksToDatabase adds new webhook to db.
func (r *WebhooksRepository) AddWebhookToDatabase(rWebhook *domains.Webhook) error {
	dbWebhook := dto.ToDbWebhook(rWebhook)
	err := r.db.CreateWebhook(context.Background(), dbWebhook)
	return err
}

// DeleteWebhookByName deletes webhook by name from db.
func (r *WebhooksRepository) DeleteWebhookByName(name string) error {
	err := r.db.DeleteWebhookByName(context.Background(), name)
	return err
}

// DeleteWebhookByUrl deletes webhook by url from db.
func (r *WebhooksRepository) DeleteWebhookByUrl(url string) error {
	err := r.db.DeleteWebhookByUrl(context.Background(), url)
	return err
}

// GetWebhookByName returns webhook from db by given name.
func (r *WebhooksRepository) GetWebhookByName(name string) (*domains.Webhook, error) {
	w, err := r.db.GetWebhookByName(context.Background(), name)
	if err != nil {
		return nil, err
	}
	dbw := w.ToWebhook()
	return dbw, err
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

// NewWebhooksRepository creates and returns WebhooksRepository instance.
func NewWebhooksRepository(db *sql.HeadersDb) *WebhooksRepository {
	return &WebhooksRepository{db: db}
}
