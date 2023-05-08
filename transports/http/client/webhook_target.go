package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/bitcoin-hc/notification"
)

type webhookTargetClientFunc func(headers map[string]string, method string, url string, body any) (*http.Response, error)

func (f webhookTargetClientFunc) Call(headers map[string]string, method string, url string, body any) (*http.Response, error) {
	return f(headers, method, url, body)
}

// NewWebhookTargetClient returns a new WebhookTargetClient.
func NewWebhookTargetClient() notification.WebhookTargetClient {
	return webhookTargetClientFunc(callRequest)
}

func callRequest(headers map[string]string, method string, url string, body any) (*http.Response, error) {
	bBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewReader(bBytes))

	if err != nil {
		return nil, err
	}

	for header, value := range headers {
		req.Header.Add(header, value)
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return res, nil
}
