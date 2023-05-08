package domains

import (
	"time"
)

// Webhook represents webhook.
type Webhook struct {
	Name              string    `json:"name"`
	Url               string    `json:"url"`
	TokenHeader       string    `json:"-"`
	Token             string    `json:"-"`
	CreatedAt         time.Time `json:"createdAt"`
	LastEmitStatus    string    `json:"lastEmitStatus"`
	LastEmitTimestamp time.Time `json:"lastEmitTimestamp"`
	ErrorsCount       int       `json:"errorsCount"`
}

// CreateWebhook creates new webhook.
func CreateWebhook(name, url, tokenHeader, token string) *Webhook {
	return &Webhook{
		Name:        name,
		Url:         url,
		TokenHeader: tokenHeader,
		Token:       token,
		CreatedAt:   time.Now(),
	}
}
