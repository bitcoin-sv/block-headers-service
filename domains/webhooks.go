package domains

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
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

// Notify sends notification to webhook.
func (w *Webhook) Notify(h *BlockHeader) (int, string, error) {
	// Prepare the request.
	headerBytes, err := json.Marshal(h)
	if err != nil {
		return 0, "", err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, w.Url, bytes.NewReader(headerBytes))

	if err != nil {
		return 0, "", err
	}

	// Add the necessary headers.
	req.Header.Add(w.TokenHeader, w.Token)
	req.Header.Add("Content-Type", "application/json")

	// Send the request.
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}

	defer res.Body.Close() // nolint: all

	// Read the response.
	body, _ := io.ReadAll(res.Body)

	return res.StatusCode, string(body), nil
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
