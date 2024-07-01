package merkleroots_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testapp"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/api/merkleroots"
)

func TestReturnSuccessFromVerify(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
	defer cleanup()
	query := []domains.MerkleRootConfirmationRequestItem{
		{
			MerkleRoot:  chaincfg.GenesisMerkleRoot.String(),
			BlockHeight: 0,
		},
	}
	expectedResult := struct {
		code int
		body merkleroots.ConfirmationsResponse
	}{
		code: http.StatusOK,
		body: merkleroots.ConfirmationsResponse{
			ConfirmationState: domains.Confirmed,
			Confirmations: []merkleroots.MerkleRootConfirmation{
				{
					Hash:         chaincfg.GenesisHash.String(),
					BlockHeight:  0,
					MerkleRoot:   chaincfg.GenesisMerkleRoot.String(),
					Confirmation: domains.Confirmed,
				},
			},
		},
	}

	// when
	res := bhs.API().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expectedResult.code)

	var mrcf merkleroots.ConfirmationsResponse
	json.NewDecoder(res.Body).Decode(&mrcf)

	assert.Equal(t, mrcf.ConfirmationState, expectedResult.body.ConfirmationState)
	for i, conf := range mrcf.Confirmations {
		expected := expectedResult.body.Confirmations[i]
		assert.Equal(t, conf.Hash, expected.Hash)
		assert.Equal(t, conf.BlockHeight, expected.BlockHeight)
		assert.Equal(t, conf.MerkleRoot, expected.MerkleRoot)
		assert.Equal(t, conf.Confirmation, expected.Confirmation)
	}
}

func TestReturnFailureFromVerifyWhenAuthorizationIsTurnedOnAndCalledWithoutToken(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t)
	defer cleanup()
	query := []domains.MerkleRootConfirmationRequestItem{}
	expectedResult := struct {
		code int
		body []byte
	}{
		code: http.StatusUnauthorized,
		body: []byte("\"empty auth header\""),
	}

	// when
	res := bhs.API().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expectedResult.code)
	if res.Code != expectedResult.code {
		t.Errorf("Expected to get status %d but instead got %d\n", expectedResult.code, res.Code)
	}
	body, _ := io.ReadAll(res.Body)
	if !bytes.Equal(body, expectedResult.body) {
		t.Errorf("Expected to get body %s but insead got %s\n", expectedResult.body, body)
	}
}

func TestReturnInvalidFromVerify(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
	defer cleanup()
	query := []domains.MerkleRootConfirmationRequestItem{
		{
			MerkleRoot:  chaincfg.GenesisMerkleRoot.String(),
			BlockHeight: 0,
		},
		{
			MerkleRoot:  "invalid_merkle_root",
			BlockHeight: 1,
		},
		{
			MerkleRoot:  "unable_to_verify_merkle_root",
			BlockHeight: 8, // Bigger than top height
		},
	}
	expectedResult := struct {
		code int
		body merkleroots.ConfirmationsResponse
	}{
		code: http.StatusOK,
		body: merkleroots.ConfirmationsResponse{
			ConfirmationState: domains.Invalid,
			Confirmations: []merkleroots.MerkleRootConfirmation{
				{
					Hash:         chaincfg.GenesisHash.String(),
					BlockHeight:  0,
					MerkleRoot:   chaincfg.GenesisMerkleRoot.String(),
					Confirmation: domains.Confirmed,
				},
				{
					Hash:         "",
					BlockHeight:  1,
					MerkleRoot:   "invalid_merkle_root",
					Confirmation: domains.Invalid,
				},
				{
					Hash:         "",
					BlockHeight:  8,
					MerkleRoot:   "unable_to_verify_merkle_root",
					Confirmation: domains.UnableToVerify,
				},
			},
		},
	}

	// when
	res := bhs.API().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expectedResult.code)

	var mrcf merkleroots.ConfirmationsResponse
	json.NewDecoder(res.Body).Decode(&mrcf)

	assert.Equal(t, mrcf.ConfirmationState, expectedResult.body.ConfirmationState)
	for i, conf := range mrcf.Confirmations {
		expected := expectedResult.body.Confirmations[i]
		assert.Equal(t, conf.Hash, expected.Hash)
		assert.Equal(t, conf.BlockHeight, expected.BlockHeight)
		assert.Equal(t, conf.MerkleRoot, expected.MerkleRoot)
		assert.Equal(t, conf.Confirmation, expected.Confirmation)
	}
}

func TestReturnPartialSuccessFromVerify(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
	defer cleanup()
	query := []domains.MerkleRootConfirmationRequestItem{
		{
			MerkleRoot:  chaincfg.GenesisMerkleRoot.String(),
			BlockHeight: 0,
		},
		{
			MerkleRoot:  "unable_to_verify_merkle_root",
			BlockHeight: 8, // Bigger than top height
		},
	}
	expectedResult := struct {
		code int
		body merkleroots.ConfirmationsResponse
	}{
		code: http.StatusOK,
		body: merkleroots.ConfirmationsResponse{
			ConfirmationState: domains.UnableToVerify,
			Confirmations: []merkleroots.MerkleRootConfirmation{
				{
					Hash:         chaincfg.GenesisHash.String(),
					BlockHeight:  0,
					MerkleRoot:   chaincfg.GenesisMerkleRoot.String(),
					Confirmation: domains.Confirmed,
				},
				{
					Hash:         "",
					BlockHeight:  8,
					MerkleRoot:   "unable_to_verify_merkle_root",
					Confirmation: domains.UnableToVerify,
				},
			},
		},
	}

	// when
	res := bhs.API().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expectedResult.code)

	var mrcf merkleroots.ConfirmationsResponse
	json.NewDecoder(res.Body).Decode(&mrcf)

	assert.Equal(t, mrcf.ConfirmationState, expectedResult.body.ConfirmationState)
	for i, conf := range mrcf.Confirmations {
		expected := expectedResult.body.Confirmations[i]
		assert.Equal(t, conf.Hash, expected.Hash)
		assert.Equal(t, conf.BlockHeight, expected.BlockHeight)
		assert.Equal(t, conf.MerkleRoot, expected.MerkleRoot)
		assert.Equal(t, conf.Confirmation, expected.Confirmation)
	}
}

func TestReturnBadRequestErrorFromVerifyWhenGivenEmtpyArray(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
	defer cleanup()
	query := []domains.MerkleRootConfirmationRequestItem{}
	expectedResult := struct {
		code int
		body []byte
	}{
		code: http.StatusBadRequest,
		body: []byte("\"at least one merkleroot is required\""),
	}

	// when
	res := bhs.API().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expectedResult.code)

	body, _ := io.ReadAll(res.Body)
	if !bytes.Equal(body, expectedResult.body) {
		t.Errorf("Expected to get body %s but insead got %s\n", expectedResult.body, body)
	}
}

func verify(request []domains.MerkleRootConfirmationRequestItem) (req *http.Request, err error) {
	query, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(query)
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/chain/merkleroot/verify",
		body,
	)
}
