package merkleroots_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testapp"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/api/merkleroots"
	"github.com/stretchr/testify/require"
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
		body string
	}{
		code: http.StatusUnauthorized,
		body: "{\"code\":\"ErrMissingAuthHeader\",\"message\":\"Empty auth header\"}",
	}

	// when
	res := bhs.API().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expectedResult.code)
	if res.Code != expectedResult.code {
		t.Errorf("Expected to get status %d but instead got %d\n", expectedResult.code, res.Code)
	}
	require.JSONEq(t, expectedResult.body, res.Body.String())
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
		body string
	}{
		code: http.StatusBadRequest,
		body: "{\"code\":\"ErrVerifyMerklerootsBadBody\",\"message\":\"At least one merkleroot is required\"}",
	}

	// when
	res := bhs.API().Call(verify(query))

	// then
	assert.Equal(t, res.Code, expectedResult.code)
	require.JSONEq(t, expectedResult.body, res.Body.String())
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

func TestMerkleRootsSuccess(t *testing.T) {
	tests := map[string]struct {
		batchSize     string
		evaluationKey string
		expectedBody  string
		expectedCode  int
	}{
		"return page from 2nd element": {
			batchSize:     "2",
			evaluationKey: "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
			expectedCode:  http.StatusOK,
			expectedBody: `{
		                     "content": [
		                        {
		                          "merkleRoot": "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
		                          "blockHeight": 2
		                        },
		                        {
		                          "merkleRoot": "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
		                          "blockHeight": 3
		                        }
		                      ],
		                      "page": {
		                        "totalElements": 5,
		                        "size": 2,
		                        "lastEvaluatedKey": "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644"
		                     }
		                  }`,
		},
		"return page without providing any params": {
			batchSize:     "",
			evaluationKey: "",
			expectedCode:  http.StatusOK,
			expectedBody: `{
		                    "content": [
		                      {
		                        "merkleRoot": "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
		                        "blockHeight": 0
		                      },
		                      {
		                        "merkleRoot": "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
		                        "blockHeight": 1
		                      },
		                      {
		                        "merkleRoot": "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
		                        "blockHeight": 2
		                      },
		                      {
		                        "merkleRoot": "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
		                        "blockHeight": 3
		                      },
		                      {
		                        "merkleRoot": "df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a",
		                        "blockHeight": 4
		                      }
		                    ],
		                    "page": {
		                      "totalElements": 5,
		                      "size": 5,
		                      "lastEvaluatedKey": ""
		                    }
		                  }`,
		},
		"return page without providing evaluationKey": {
			batchSize:     "2",
			evaluationKey: "",
			expectedCode:  http.StatusOK,
			expectedBody: `{
		                      "content": [
		                        {
		                          "merkleRoot": "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
		                          "blockHeight": 0
		                        },
		                        {
		                          "merkleRoot": "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
		                          "blockHeight": 1
		                        }
		                      ],
		                      "page": {
		                        "totalElements": 5,
		                        "size": 2,
		                        "lastEvaluatedKey": "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098"
		                      }
		                   }`,
		},
		"return page without providing batchSize": {
			batchSize:     "",
			evaluationKey: "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
			expectedCode:  http.StatusOK,
			expectedBody: `{
                        "content": [
		                       {
		                         "merkleRoot": "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
		                         "blockHeight": 3
		                       },
		                       {
		                         "merkleRoot": "df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a",
		                         "blockHeight": 4
		                       }
                        ],
                        "page": {
                          "totalElements": 5,
                          "size": 2,
                          "lastEvaluatedKey": ""
                        }
                     }`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			// setup
			bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled(), testapp.WithLongestChain())
			defer cleanup()

			// when
			res := bhs.API().Call(getMerkleRoots(test.batchSize, test.evaluationKey))

			// then
			require.Equal(t, test.expectedCode, res.Code)
			require.JSONEq(t, test.expectedBody, res.Body.String())
		})
	}
}

func TestMerkleRootsFailure(t *testing.T) {
	tests := map[string]struct {
		batchSize     string
		evaluationKey string
		expectedBody  string
		expectedCode  int
	}{
		"return error when batchSize is not positive number": {
			batchSize:     "notIntValue",
			evaluationKey: "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
			expectedCode:  http.StatusBadRequest,
			expectedBody: `{
		                   "code": "ErrInvalidBatchSize",
		                   "message": "batchSize must be 0 or a positive integer"
		                  }`,
		},
		"return error when evaluationKey doesn't exist": {
			batchSize:     "2",
			evaluationKey: "keyNotExisting",
			expectedCode:  http.StatusNotFound,
			expectedBody: `{
		                   "code": "ErrMerkleRootNotFound",
		                   "message": "No block with provided merkleroot was found"
		                  }`,
		},
		"return error when evaluationKey merkleroot is from stale chain": {
			batchSize:     "2",
			evaluationKey: "88d2a4e04a96b45e3ba04637098a92fd0786daf3fc8ff88314f8e739a9918bf3",
			expectedCode:  http.StatusConflict,
			expectedBody: `{
		                   "code": "ErrMerkleRootNotInLongestChain",
		                   "message": "Provided merkleroot is not part of the longest chain"
		                  }`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			// setup
			bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChainFork(), testapp.WithAPIAuthorizationDisabled())
			defer cleanup()

			// when
			res := bhs.API().Call(getMerkleRoots(test.batchSize, test.evaluationKey))

			// then
			require.Equal(t, test.expectedCode, res.Code)
			require.JSONEq(t, test.expectedBody, res.Body.String())
		})
	}
}

// getMerkleRoots creates http request to fetch /chain/merkleroot it accepts two params
// batchSize and lastEvaluatedKey of type string, we can omit any of them to simulate
// user not passing these values and need to pass empty string in this place
func getMerkleRoots(batchSize, lastEvaluatedKey string) (req *http.Request, err error) {
	address, err := url.Parse("/api/v1/chain/merkleroot")
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	if batchSize != "" {
		query.Add("batchSize", batchSize)
	}
	if lastEvaluatedKey != "" {
		query.Add("lastEvaluatedKey", lastEvaluatedKey)
	}

	address.RawQuery = query.Encode()
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		address.String(),
		nil,
	)
}
