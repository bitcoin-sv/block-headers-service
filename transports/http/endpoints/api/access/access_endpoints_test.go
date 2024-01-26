package access_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/internal/tests/assert"
	"github.com/bitcoin-sv/pulse/internal/tests/testpulse"

	"github.com/bitcoin-sv/pulse/domains"
)

const EmptyToken = ""

// Tests the GET /access endpoint without authorization header.
func TestAccessEndpointWithoutAuthHeader(t *testing.T) {
	//setup
	pulse, cleanup := testpulse.NewTestPulse(t)
	defer cleanup()

	//when
	res := pulse.Api().Call(getTokenInfo(EmptyToken))

	//then
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusUnauthorized, res.Code)
	}
}

// Tests the GET /access endpoint with global auth header.
func TestAccessEndpointWithGlobalAuthHeader(t *testing.T) {
	//setup
	cfg := config.GetDefaultAppConfig()
	pulse, cleanup := testpulse.NewTestPulse(t)
	defer cleanup()

	//given
	authToken := cfg.HTTP.AuthToken

	//when
	res := pulse.Api().Call(getTokenInfo(authToken))

	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}

	body := tokenFromResponse(t, res)
	if !body.IsAdmin {
		t.Fatalf("Expected to get admin token")
	}
}

// Tests the GET /access endpoint with wrong header.
func TestAccessEndpointWithWrongAuthHeader(t *testing.T) {
	//setup
	pulse, cleanup := testpulse.NewTestPulse(t)
	defer cleanup()

	//when
	res := pulse.Api().Call(getTokenInfo("wrong_token"))

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusUnauthorized, res.Code)
	}
}

// Tests the POST /access endpoint with created auth token.
func TestAccessEndpointWithCreatedAuthHeader(t *testing.T) {
	//setup
	cfg := config.GetDefaultAppConfig()
	pulse, cleanup := testpulse.NewTestPulse(t)
	defer cleanup()

	//when
	res := pulse.Api().Call(createToken(cfg.HTTP.AuthToken))

	//then
	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}

	//and
	body := tokenFromResponse(t, res)
	if body.IsAdmin {
		t.Fatalf("Expected to get non admin token")
	}
	token := body.Token

	//when
	res = pulse.Api().Call(getTokenInfo(token))

	//then
	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}

	//and
	body = tokenFromResponse(t, res)
	if body.IsAdmin {
		t.Fatalf("Expected to get non admin token")
	}

	//and
	assert.Equal(t, body.Token, token)
}

// Tests the DELETE method for the /access endpoint for created auth token.
func TestDeleteTokenEndpoint(t *testing.T) {
	//setup
	cfg := config.GetDefaultAppConfig()
	pulse, cleanup := testpulse.NewTestPulse(t)
	defer cleanup()

	//when
	res := pulse.Api().Call(createToken(cfg.HTTP.AuthToken))

	//then
	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}

	//given
	var body domains.Token
	err := json.Unmarshal(res.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	//when
	res = pulse.Api().Call(deleteToken(cfg.HTTP.AuthToken, body.Token))

	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}
}

func tokenFromResponse(t *testing.T, res *httptest.ResponseRecorder) domains.Token {
	var body domains.Token
	err := json.Unmarshal(res.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	return body
}

func getTokenInfo(headerToken string) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/access", nil)
	if headerToken != "" && err == nil {
		req.Header.Add("Authorization", "Bearer "+headerToken)
	}
	return
}

func createToken(headerToken string) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/access", nil)
	if headerToken != "" && err == nil {
		req.Header.Add("Authorization", "Bearer "+headerToken)
	}
	return
}

func deleteToken(headerToken string, tokenToDelete string) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/v1/access/"+tokenToDelete, nil)
	if headerToken != "" && err == nil {
		req.Header.Add("Authorization", "Bearer "+headerToken)
	}
	return
}
