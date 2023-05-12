package domains

import (
	"bytes"
	"context"
	"encoding/json"
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

// Notify sends notification to webhook.
func (w *Webhook) Notify(h *BlockHeader) error {
	// Prepare the request.
	headerBytes, err := json.Marshal(h)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, w.Url, bytes.NewReader(headerBytes))

	if err != nil {
		return err
	}

	// Add the necessary headers.
	req.Header.Add(w.TokenHeader, w.Token)
	req.Header.Add("Content-Type", "application/json")

	// Send the request.
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
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

	// Update the webhook after successfull notification.
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
