package domains

import (
	"bytes"
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

// Nofify sends notification to webhook.
func (w *Webhook) Notify(h *BlockHeader) (int, string, error) {
	// Prepare the request.
	headerBytes, _ := json.Marshal(h)
	req, err := http.NewRequest(http.MethodPost, w.Url, bytes.NewReader(headerBytes))

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

	defer res.Body.Close()

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
