package dto

import (
	"time"

	"github.com/libsv/bitcoin-hc/domains"
)

// DbWebhook represent webhook saved in db.
type DbWebhook struct {
	Url               string    `db:"url"`
	TokenHeader       string    `db:"tokenHeader"`
	Token             string    `db:"token"`
	CreatedAt         time.Time `db:"createdAt"`
	LastEmitStatus    string    `db:"lastEmitStatus"`
	LastEmitTimestamp time.Time `db:"lastEmitTimestamp"`
	ErrorsCount       int       `db:"errorsCount"`
	Active            bool      `db:"active"`
}

// ToWebhook converts DbWebhook to Webhook.
func (dbt *DbWebhook) ToWebhook() *domains.Webhook {
	return &domains.Webhook{
		Url:         dbt.Url,
		TokenHeader: dbt.TokenHeader,
		Token:       dbt.Token,
		CreatedAt:   dbt.CreatedAt,
		ErrorsCount: dbt.ErrorsCount,
		Active:      dbt.Active,
	}
}

// ToDbWebhook converts Webhook to DbWebhook.
func ToDbWebhook(t *domains.Webhook) *DbWebhook {
	return &DbWebhook{
		Url:         t.Url,
		TokenHeader: t.TokenHeader,
		Token:       t.Token,
		CreatedAt:   t.CreatedAt,
		ErrorsCount: t.ErrorsCount,
		Active:      t.Active,
	}
}
