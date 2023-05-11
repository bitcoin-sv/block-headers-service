package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/tests/testrepository"
	"github.com/libsv/bitcoin-hc/repository"
	"github.com/libsv/bitcoin-hc/service"
	"github.com/libsv/bitcoin-hc/vconfig"
	"github.com/spf13/viper"

	wHttp "github.com/libsv/bitcoin-hc/transports/http"
)

var webhookUrl = "http://localhost:8080/api/v1/webhook/notify"

var preparedWebhook = wHttp.WebhookRequest{
	Url: webhookUrl,
	RequiredAuth: wHttp.RequiredAuth{
		Type:  "BEARER",
		Token: "test-token",
	},
}

// TestCreateWebhookEndpoint tests the webhook registration.
func TestCreateWebhookEndpoint(t *testing.T) {
	r := setUpWebhooks()

	err := createWebhook(r)

	if err != nil {
		t.Fatalf("%v\n", err)
	}
}

// TestMultipleIdenticalWebhooks tests creating mutltiple webhooks with this same Url.
func TestMultipleIdenticalWebhooks(t *testing.T) {
	r := setUpWebhooks()

	err := createWebhook(r)

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	// Call this same webhook endpoint again.
	res2, err := callWebhookEndpoint(r, &preparedWebhook, nil)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if res2.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res2.Code)
	}

	body, _ := io.ReadAll(res2.Body)
	bodyStr := string(body)[1 : len(string(body))-1]

	if bodyStr != "webhook already exists and is active" {
		t.Fatalf("Expected message: 'webhook already exists and is active' but instead got '%s'\n", bodyStr)
	}
}

// TestRevokeWebhookEndpoint tests the webhook revocation.
func TestRevokeWebhookEndpoint(t *testing.T) {
	r := setUpWebhooks()

	err := createWebhook(r)

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	res, err := callWebhookEndpoint(r, nil, &webhookUrl)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}

	body, _ := io.ReadAll(res.Body)
	bodyStr := string(body)[1 : len(string(body))-1]

	if bodyStr != "Webhook revoked" {
		t.Fatalf("Expected message: 'Webhook revoked' but instead got '%s'\n", bodyStr)
	}
}

func setUpWebhooks() (router *gin.Engine) {
	// Load config.
	vconfig.NewViperConfig("pulse-test")

	// Create the handler.
	var webhooksTable []domains.Webhook

	repo := &repository.Repositories{
		Webhooks: testrepository.NewWebhooksTestRepository(&webhooksTable),
	}

	hs := service.NewServices(service.Dept{
		Repositories: repo,
	})

	h := NewHandler(hs)

	// Create the router.
	gin.SetMode(gin.TestMode)
	router = gin.Default()
	prefix := viper.GetString(urlPrefix)
	webhooks := router.Group(prefix)
	webhooks.POST("/webhook", h.registerWebhook)
	webhooks.DELETE("/webhook", h.revokeWebhook)

	return
}

// Call GET or DELETE /webhook endpoint
// If urlToDelete is not nil, DELETE method will be called.
func callWebhookEndpoint(r *gin.Engine, w *wHttp.WebhookRequest, urlToDelete *string) (*httptest.ResponseRecorder, error) {
	// Create the request.
	prefix := viper.GetString(urlPrefix)
	var req *http.Request
	var err error

	if urlToDelete != nil {
		req, err = http.NewRequest(http.MethodDelete, prefix+"/webhook?url="+*urlToDelete, nil)
		if err != nil {
			return nil, errors.New("couldn't create request")
		}
	} else {
		webhookBytes, err := json.Marshal(w)
		if err != nil {
			return nil, fmt.Errorf("couldn't marshal webhook: %w", err)
		}
		req, err = http.NewRequest(http.MethodPost, prefix+"/webhook", bytes.NewReader(webhookBytes))
		if err != nil {
			return nil, errors.New("couldn't create request")
		}
		req.Header.Add("Content-Type", "application/json")
	}

	// Serve the request.
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	return res, nil
}

func createWebhook(r *gin.Engine) error {
	// Try to register new webhook.
	res, err := callWebhookEndpoint(r, &preparedWebhook, nil)
	if err != nil {
		return err
	}

	// Get webhook from the response.
	var webhook domains.Webhook
	err = json.Unmarshal(res.Body.Bytes(), &webhook)
	if err != nil {
		return err
	}

	// Check response status and webhook.
	if res.Code != http.StatusOK {
		return fmt.Errorf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}
	if webhook.Url == "" {
		return fmt.Errorf("Expected to get full webhook but instead got empty one")
	}

	return nil
}
