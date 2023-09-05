package merkleroots_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/tests/assert"
	"github.com/libsv/bitcoin-hc/internal/tests/testpulse"
	"github.com/libsv/bitcoin-hc/transports/http/endpoints/api/merkleroots"
)

func TestReturnSuccessFromVerify(t *testing.T) {
	// setup
	pulse, cleanup := testpulse.NewTestPulse(t, testpulse.WithLongestChain())
	defer cleanup()
	query := []string{chaincfg.GenesisMerkleRoot.String()}
	expected_result := struct {
		code int
		body merkleroots.MerkleRootsConfirmationsResponse
	}{
		code: http.StatusOK,
		body: merkleroots.MerkleRootsConfirmationsResponse{
			AllConfirmed: true,
			Confirmations: []merkleroots.MerkleRootConfirmation{
				{
					Hash:       chaincfg.GenesisHash.String(),
					MerkleRoot: chaincfg.GenesisMerkleRoot.String(),
					Confirmed:  true,
				},
			},
		},
	}

	// when
	res := pulse.Api().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expected_result.code)

	var mrcf merkleroots.MerkleRootsConfirmationsResponse
	json.NewDecoder(res.Body).Decode(&mrcf)

	assert.Equal(t, mrcf.AllConfirmed, expected_result.body.AllConfirmed)
	for i, conf := range mrcf.Confirmations {
		expected := expected_result.body.Confirmations[i]
		assert.Equal(t, conf.Hash, expected.Hash)
		assert.Equal(t, conf.MerkleRoot, expected.MerkleRoot)
		assert.Equal(t, conf.Confirmed, expected.Confirmed)
	}
}

func TestReturnFailureFromVerifyWhenAuthorizationIsTurnedOnAndCalledWithoutToken(t *testing.T) {
	// setup
	pulse, cleanup := testpulse.NewTestPulse(t, testpulse.WithApiAuthorization())
	defer cleanup()
	query := []string{}
	expected_result := struct {
		code int
		body []byte
	}{
		code: http.StatusUnauthorized,
		body: []byte("\"empty auth header\""),
	}

	// when
	res := pulse.Api().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expected_result.code)
	if res.Code != expected_result.code {
		t.Errorf("Expected to get status %d but instead got %d\n", expected_result.code, res.Code)
	}
	body, _ := io.ReadAll(res.Body)
	if !bytes.Equal(body, expected_result.body) {
		t.Errorf("Expected to get body %s but insead got %s\n", expected_result.body, body)
	}
}

func TestReturnPartialSuccessFromVerify(t *testing.T) {
	// setup
	pulse, cleanup := testpulse.NewTestPulse(t, testpulse.WithLongestChain())
	defer cleanup()
	query := []string{chaincfg.GenesisMerkleRoot.String(), "not_found_merkle_root"}
	expected_result := struct {
		code int
		body merkleroots.MerkleRootsConfirmationsResponse
	}{
		code: http.StatusOK,
		body: merkleroots.MerkleRootsConfirmationsResponse{
			AllConfirmed: false,
			Confirmations: []merkleroots.MerkleRootConfirmation{
				{
					Hash:       chaincfg.GenesisHash.String(),
					MerkleRoot: chaincfg.GenesisMerkleRoot.String(),
					Confirmed:  true,
				},
				{
					Hash:       "",
					MerkleRoot: "not_found_merkle_root",
					Confirmed:  false,
				},
			},
		},
	}

	// when
	res := pulse.Api().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expected_result.code)

	var mrcf merkleroots.MerkleRootsConfirmationsResponse
	json.NewDecoder(res.Body).Decode(&mrcf)

	assert.Equal(t, mrcf.AllConfirmed, expected_result.body.AllConfirmed)
	for i, conf := range mrcf.Confirmations {
		expected := expected_result.body.Confirmations[i]
		assert.Equal(t, conf.Hash, expected.Hash)
		assert.Equal(t, conf.MerkleRoot, expected.MerkleRoot)
		assert.Equal(t, conf.Confirmed, expected.Confirmed)
	}
}

func TestReturnBadRequestErrorFromVerifyWhenGivenEmtpyArray(t *testing.T) {
	// setup
	pulse, cleanup := testpulse.NewTestPulse(t, testpulse.WithLongestChain())
	defer cleanup()
	query := []string{}
	expected_result := struct {
		code int
		body []byte
	}{
		code: http.StatusBadRequest,
		body: []byte("\"At least one merkleroot is required\""),
	}

	// when
	res := pulse.Api().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expected_result.code)

	body, _ := io.ReadAll(res.Body)
	if !bytes.Equal(body, expected_result.body) {
		t.Errorf("Expected to get body %s but insead got %s\n", expected_result.body, body)
	}
}

func verify(merkleroots []string) (req *http.Request, err error) {
	array, err := json.Marshal(merkleroots)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(array)
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/chain/merkleroot/verify",
		body,
	)
}
