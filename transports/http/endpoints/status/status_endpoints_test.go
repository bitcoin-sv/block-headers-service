package status_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/testbhs"
)

func TestReturnSuccessFromStatus(t *testing.T) {
	//setup
	bhs, cleanup := testbhs.NewTestBHS(t)
	defer cleanup()

	//when
	res := bhs.Api().Call(getStatus())

	//then
	if res.Code != http.StatusOK {
		t.Errorf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}
}

func TestReturnSuccessFromStatusWhenAuthorizationIsTurnedOnAndCalledWithoutToken(t *testing.T) {
	//setup
	bhs, cleanup := testbhs.NewTestBHS(t)
	defer cleanup()

	//when
	res := bhs.Api().Call(getStatus())

	//then
	if res.Code != http.StatusOK {
		t.Errorf("Expected to get status %d but instead got %d\n", http.StatusOK, res.Code)
	}
}

func getStatus() (req *http.Request, err error) {
	return http.NewRequestWithContext(context.Background(), http.MethodGet, "/status", nil)
}
