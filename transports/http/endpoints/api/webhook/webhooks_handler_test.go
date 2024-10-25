package webhook_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/testapp"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/api/webhook"
	"github.com/stretchr/testify/require"
)

var webhookURL = "http://localhost:8080/api/v1/webhook/notify"

var preparedWebhook = webhook.Request{
	URL: webhookURL,
	RequiredAuth: webhook.RequiredAuth{
		Type:  "BEARER",
		Token: "test-token",
	},
}

// TestCreateWebhookEndpoint tests the webhook registration.
func TestCreateWebhookEndpoint(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled())
	defer cleanup()

	// when
	res := bhs.API().Call(createWebhook())

	// then
	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}
}

// TestMultipleIdenticalWebhooks tests creating mutltiple webhooks with this same URL.
func TestMultipleIdenticalWebhooks(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled())
	defer cleanup()
	expectedBodyResponse := "{\"code\":\"ErrRefreshWebhook\",\"message\":\"webhook already exists and is active\"}"

	// when
	res := bhs.API().Call(createWebhook())

	// then
	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}

	// when
	res2 := bhs.API().Call(createWebhook())

	if res2.Code != http.StatusBadRequest {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res2.Code)
	}

	require.JSONEq(t, expectedBodyResponse, res2.Body.String())
}

// TestRevokeWebhookEndpoint tests the webhook revocation.
func TestRevokeWebhookEndpoint(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled())
	defer cleanup()

	// when
	res := bhs.API().Call(createWebhook())

	// then
	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}

	res2 := bhs.API().Call(revokeWebhook(webhookURL))

	if res2.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res2.Code)
	}

	body, _ := io.ReadAll(res2.Body)
	bodyStr := string(body)[1 : len(string(body))-1]

	if bodyStr != "Webhook revoked" {
		t.Fatalf("Expected message: 'Webhook revoked' but instead got '%s'\n", bodyStr)
	}
}

func createWebhook() (req *http.Request, err error) {
	webhookBytes, err := json.Marshal(&preparedWebhook)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal webhook: %w", err)
	}
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/webhook", bytes.NewReader(webhookBytes))
	req.Header.Add("Content-Type", "application/json")
	return
}

func revokeWebhook(url string) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/v1/webhook?url="+url, nil)
	return
}
