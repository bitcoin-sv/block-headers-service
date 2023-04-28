package handler

import (
	"encoding/json"
	"errors"
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
)

// Tests for authorization.

// Tests the GET /access endpoint without authorization header.
func TestAccessEndpointWithoutAuthHeader(t *testing.T) {
	r := setUp()

	// Try to get access token without authorization header
	res, err := callEndpoint(r, "", nil)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}
}

// Tests the GET /access endpoint with wrong header.
func TestAccessEndpointWithWrongAuthHeader(t *testing.T) {
	r := setUp()

	// Try to get access token with wrong authorization header
	res, err := callEndpoint(r, "wrongToken", nil)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}
}

// Tests the GET /access endpoint with global auth token.
func TestAccessEndpointWithGlobalAuthHeader(t *testing.T) {
	r := setUp()

	// Try to get access token with global authorization header.
	authToken := viper.GetString("http.server.authToken")
	res, err := callEndpoint(r, authToken, nil)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}
}

// Tests the GET /access endpoint with created auth token.
func TestAccessEndpointWithCreatedAuthHeader(t *testing.T) {
	r := setUp()

	// Try to get access token with global authorization header.
	authToken := viper.GetString("http.server.authToken")
	res, err := callEndpoint(r, authToken, nil)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}

	// Get the token from the response.
	var token domains.Token
	err = json.Unmarshal(res.Body.Bytes(), &token)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	// Try to get access token with created authorization header.
	res2, err := callEndpoint(r, token.Token, nil)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if res2.Code != http.StatusUnauthorized {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res2.Code)
	}
}

// Tests the DELETE method for the /access endpoint for created auth token.
func TestDeleteTokenEndpoint(t *testing.T) {
	r := setUp()

	// Try to get access token with global authorization header.
	authToken := viper.GetString("http.server.authToken")
	res, err := callEndpoint(r, authToken, nil)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}

	// Get the token from the response.
	var token domains.Token
	err = json.Unmarshal(res.Body.Bytes(), &token)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	// Call DELETE endpoint with created token.
	res, err = callEndpoint(r, authToken, &token.Token)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}
}

func setUp() (router *gin.Engine) {
	// Load config.
	vconfig.NewViperConfig("pulse-test").
		WithAuthorization()

	// Create the handler.
	var tokensTable []domains.Token

	repo := &repository.Repositories{
		Tokens: testrepository.NewTokensTestRepository(&tokensTable),
	}

	hs := service.NewServices(service.Dept{
		Repositories: repo,
		Peers:        nil,
	})

	h := NewHandler(hs)

	// Create the router.
	gin.SetMode(gin.TestMode)
	router = gin.Default()
	prefix := viper.GetString(urlPrefix)
	tokens := router.Group(prefix, h.tokenIdentity)
	tokens.GET("/access", h.createToken)
	tokens.DELETE("/access/:token", h.revokeToken)

	return
}

// Call GET or DELETE /access endpoint with given authorization token
// If tokenToDelete is not nil, DELETE method will be called.
func callEndpoint(r *gin.Engine, headerToken string, tokenToDelete *string) (*httptest.ResponseRecorder, error) {
	// Create the request.
	prefix := viper.GetString(urlPrefix)
	var req *http.Request
	var err error

	if tokenToDelete != nil {
		req, err = http.NewRequest(http.MethodDelete, prefix+"/access/"+*tokenToDelete, nil)
		if err != nil {
			return nil, errors.New("Couldn't create request")
		}
	} else {
		req, err = http.NewRequest(http.MethodGet, prefix+"/access", nil)
		if err != nil {
			return nil, errors.New("Couldn't create request")
		}
	}

	// Add the authorization header.
	if authToken != "" {
		req.Header.Add("Authorization", "Bearer "+headerToken)
	}

	// Serve the request.
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	return res, nil
}
