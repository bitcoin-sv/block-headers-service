package domains

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// Webhook represents webhook.
type Webhook struct {
	Url               string    `json:"url"`
	TokenHeader       string    `json:"-"`
	Token             string    `json:"-"`
	CreatedAt         time.Time `json:"createdAt"`
	LastEmitStatus    string    `json:"lastEmitStatus"`
	LastEmitTimestamp time.Time `json:"lastEmitTimestamp"`
	ErrorsCount       int       `json:"errorsCount"`
	Active            bool      `json:"active"`
}

// WebhookMaxTries is the maximum number of times a webhook will be retried.
const WebhookMaxTries = "webhook.maxTries"

// WebhookTargetClient is the interface for the webhooks http calls.
type WebhookTargetClient interface {
	Call(headers map[string]string, method string, url string, body *BlockHeader) (*http.Response, error)
}

// Notify sends notification to webhook.
func (w *Webhook) Notify(h *BlockHeader, client WebhookTargetClient) error {
	// Prepare headers
	headers := map[string]string{
		w.TokenHeader:  w.Token,
		"Content-Type": "application/json",
	}

	res, err := client.Call(headers, http.MethodPost, w.Url, h)

	if err != nil {
		// Update the webhook after failed notification.
		w.updateWebhookAfterNotification(0, "", err)
		return err
	}

	defer res.Body.Close() // nolint: all

	// Read the response.
	body, _ := io.ReadAll(res.Body)
	if err != nil {
		w.updateWebhookAfterNotification(0, "", err)
		return err
	}

	// Update the webhook after successful notification.
	strBody := string(body)
	w.updateWebhookAfterNotification(res.StatusCode, strBody, err)

	return nil
}

func (w *Webhook) updateWebhookAfterNotification(sCode int, body string, err error) {
	w.LastEmitTimestamp = time.Now()

	if err != nil {
		w.LastEmitStatus = fmt.Sprint(err)
	} else {
		w.LastEmitStatus = fmt.Sprint(sCode, " ", body)
	}

	// If status code is not 200, increment errors count and set active to false if errors count is more than max tries
	if sCode != http.StatusOK {
		w.ErrorsCount = w.ErrorsCount + 1

		if w.ErrorsCount >= viper.GetInt(WebhookMaxTries) {
			w.Active = false
		}
	} else {
		// If status code is 200, reset errors count and set active to true
		w.ErrorsCount = 0
		w.Active = true
	}
}

// CreateWebhook creates new webhook.
func CreateWebhook(url, tokenHeader, token string) *Webhook {
	return &Webhook{
		Url:         url,
		TokenHeader: tokenHeader,
		Token:       token,
		CreatedAt:   time.Now(),
		ErrorsCount: 0,
		Active:      true,
	}
}
