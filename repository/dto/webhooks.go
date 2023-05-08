package dto

import (
	"time"

	"github.com/libsv/bitcoin-hc/domains"
)

// DbWebhook represent webhook saved in db.
type DbWebhook struct {
	Name              string    `db:"name"`
	Url               string    `db:"url"`
	TokenHeader       string    `db:"tokenHeader"`
	Token             string    `db:"token"`
	CreatedAt         time.Time `db:"createdAt"`
	LastEmitStatus    string    `db:"lastEmitStatus"`
	LastEmitTimestamp time.Time `db:"lastEmitTimestamp"`
	ErrorsCount       int       `db:"errorsCount"`
}

// ToWebhook converts DbWebhook to Webhook.
func (dbt *DbWebhook) ToWebhook() *domains.Webhook {
	return &domains.Webhook{
		Name:        dbt.Name,
		Url:         dbt.Url,
		TokenHeader: dbt.TokenHeader,
		Token:       dbt.Token,
		CreatedAt:   dbt.CreatedAt,
		ErrorsCount: dbt.ErrorsCount,
	}
}

// ToDbWebhook converts Webhook to DbWebhook.
func ToDbWebhook(t *domains.Webhook) *DbWebhook {
	return &DbWebhook{
		Name:        t.Name,
		Url:         t.Url,
		TokenHeader: t.TokenHeader,
		Token:       t.Token,
		CreatedAt:   t.CreatedAt,
		ErrorsCount: t.ErrorsCount,
	}
}
