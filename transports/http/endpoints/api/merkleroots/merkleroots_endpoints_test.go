package merkleroots_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func TestReturnSuccessFromMerkleRoots(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled(), testapp.WithLongestChain())
	orderedByField := "BlockHeight"
	sortDirection := "ASC"
	batchSize := 2
	lastEvaluatedKey := "1"
	expectedResult := struct {
		Code int
		Body domains.MerkleRootsESKPagedResponse
	}{
		Code: http.StatusOK,
		Body: domains.MerkleRootsESKPagedResponse{
			Content: []*domains.MerkleRootsResponse{
				{
					MerkleRoot:  "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
					BlockHeight: 2,
				},
				{
					MerkleRoot:  "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
					BlockHeight: 3,
				},
			},
			Page: domains.ExclusiveStartKeyPage{
				OrderByField:     &orderedByField,
				SortDirection:    &sortDirection,
				TotalElements:    5,
				Size:             batchSize,
				LastEvaluatedKey: 3,
			},
		},
	}
	defer cleanup()

	// when
	res := bhs.API().Call(getMerkleRoots(strconv.Itoa(batchSize), lastEvaluatedKey))
	var merklerootsResponse domains.MerkleRootsESKPagedResponse
	json.NewDecoder(res.Body).Decode(&merklerootsResponse)

	// then
	assert.Equal(t, res.Code, expectedResult.Code)
	if res.Code != expectedResult.Code {
		t.Errorf("Expected to get status %d but instead got %d\n", expectedResult.Code, res.Code)
	}
	// this conversion is necessary as the default setting of the json decoder when unmarshalling JSON numbers into interface{} takes type of float64
	assert.Equal(t, int(merklerootsResponse.Page.LastEvaluatedKey.(float64)), expectedResult.Body.Page.LastEvaluatedKey.(int))
	assert.Equal(t, merklerootsResponse.Page.OrderByField, expectedResult.Body.Page.OrderByField)
	assert.Equal(t, merklerootsResponse.Page.SortDirection, expectedResult.Body.Page.SortDirection)
	assert.Equal(t, merklerootsResponse.Page.TotalElements, expectedResult.Body.Page.TotalElements)
	assert.Equal(t, merklerootsResponse.Page.Size, expectedResult.Body.Page.Size)
	for i, returnedMerkleroot := range merklerootsResponse.Content {
		expected := expectedResult.Body.Content[i]
		assert.Equal(t, returnedMerkleroot.BlockHeight, expected.BlockHeight)
		assert.Equal(t, returnedMerkleroot.MerkleRoot, expected.MerkleRoot)
	}
}

func TestReturnFailureFromMerkleRootsWhenBatchSizeIsNotInt(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled())
	batchSize := "notIntValue"
	lastEvaluatedKey := "1"
	expectedResult := struct {
		Code int
		Body []byte
	}{
		Code: http.StatusBadRequest,
		Body: []byte("\"batchSize must be a numeric value\""),
	}
	defer cleanup()

	// when
	res := bhs.API().Call(getMerkleRoots(batchSize, lastEvaluatedKey))

	// then
	assert.Equal(t, res.Code, expectedResult.Code)
	if res.Code != expectedResult.Code {
		t.Errorf("Expected to get status %d but instead got %d\n", expectedResult.Code, res.Code)
	}

	body, _ := io.ReadAll(res.Body)
	if !bytes.Equal(body, expectedResult.Body) {
		t.Errorf("Expected to get body %s but instead got %s\n", expectedResult.Body, body)
	}
}

func TestReturnFailureFromMerkleRootsWhenLastEvaluatedKeyIsNotInt(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled())
	batchSize := "3"
	lastEvaluatedKey := "notIntValue"
	expectedResult := struct {
		Code int
		Body []byte
	}{
		Code: http.StatusBadRequest,
		Body: []byte("\"lastEvaluatedKey must be a numeric value\""),
	}
	defer cleanup()

	// when
	res := bhs.API().Call(getMerkleRoots(batchSize, lastEvaluatedKey))

	// then
	assert.Equal(t, res.Code, expectedResult.Code)
	if res.Code != expectedResult.Code {
		t.Errorf("Expected to get status %d but instead got %d\n", expectedResult.Code, res.Code)
	}

	body, _ := io.ReadAll(res.Body)
	if !bytes.Equal(body, expectedResult.Body) {
		t.Errorf("Expected to get body %s but instead got %s\n", expectedResult.Body, body)
	}
}

func TestReturnSuccessFromMerkleRootsWhenLastEvaluatedKeyAndBatchSizeAreNotProvided(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled(), testapp.WithLongestChain())
	orderedByField := "BlockHeight"
	sortDirection := "ASC"
	expectedResult := struct {
		Code int
		Body domains.MerkleRootsESKPagedResponse
	}{
		Code: http.StatusOK,
		Body: domains.MerkleRootsESKPagedResponse{
			Content: []*domains.MerkleRootsResponse{
				{
					MerkleRoot:  "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
					BlockHeight: 0,
				}, {
					MerkleRoot:  "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
					BlockHeight: 1,
				}, {
					MerkleRoot:  "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
					BlockHeight: 2,
				}, {
					MerkleRoot:  "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
					BlockHeight: 3,
				}, {
					MerkleRoot:  "df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a",
					BlockHeight: 4,
				},
			},
			Page: domains.ExclusiveStartKeyPage{
				OrderByField:     &orderedByField,
				SortDirection:    &sortDirection,
				TotalElements:    5,
				Size:             5,
				LastEvaluatedKey: 4,
			},
		},
	}
	defer cleanup()

	// when
	res := bhs.API().Call(getMerkleRoots("", ""))
	var merklerootsResponse domains.MerkleRootsESKPagedResponse
	json.NewDecoder(res.Body).Decode(&merklerootsResponse)

	// then
	assert.Equal(t, res.Code, expectedResult.Code)
	if res.Code != expectedResult.Code {
		t.Errorf("Expected to get status %d but instead got %d\n", expectedResult.Code, res.Code)
	}

	// this conversion is necessary as the default setting of the json decoder when unmarshalling JSON numbers into interface{} takes type of float64
	assert.Equal(t, int(merklerootsResponse.Page.LastEvaluatedKey.(float64)), expectedResult.Body.Page.LastEvaluatedKey.(int))
	assert.Equal(t, merklerootsResponse.Page.OrderByField, expectedResult.Body.Page.OrderByField)
	assert.Equal(t, merklerootsResponse.Page.SortDirection, expectedResult.Body.Page.SortDirection)
	assert.Equal(t, merklerootsResponse.Page.TotalElements, expectedResult.Body.Page.TotalElements)
	assert.Equal(t, merklerootsResponse.Page.Size, expectedResult.Body.Page.Size)
	for i, returnedMerkleroot := range merklerootsResponse.Content {
		expected := expectedResult.Body.Content[i]
		assert.Equal(t, returnedMerkleroot.BlockHeight, expected.BlockHeight)
		assert.Equal(t, returnedMerkleroot.MerkleRoot, expected.MerkleRoot)
	}
}

func TestReturnSuccessFromMerkleRootsWhenLastEvaluatedKeyIsNotProvided(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled(), testapp.WithLongestChain())
	orderedByField := "BlockHeight"
	sortDirection := "ASC"
	batchSize := "3"
	expectedResult := struct {
		Code int
		Body domains.MerkleRootsESKPagedResponse
	}{
		Code: http.StatusOK,
		Body: domains.MerkleRootsESKPagedResponse{
			Content: []*domains.MerkleRootsResponse{
				{
					MerkleRoot:  "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
					BlockHeight: 0,
				}, {
					MerkleRoot:  "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
					BlockHeight: 1,
				}, {
					MerkleRoot:  "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
					BlockHeight: 2,
				},
			},
			Page: domains.ExclusiveStartKeyPage{
				OrderByField:     &orderedByField,
				SortDirection:    &sortDirection,
				TotalElements:    5,
				Size:             3,
				LastEvaluatedKey: 2,
			},
		},
	}
	defer cleanup()

	// when
	res := bhs.API().Call(getMerkleRoots(batchSize, ""))
	var merklerootsResponse domains.MerkleRootsESKPagedResponse
	json.NewDecoder(res.Body).Decode(&merklerootsResponse)

	// then
	assert.Equal(t, res.Code, expectedResult.Code)
	if res.Code != expectedResult.Code {
		t.Errorf("Expected to get status %d but instead got %d\n", expectedResult.Code, res.Code)
	}

	// this conversion is necessary as the default setting of the json decoder when unmarshalling JSON numbers into interface{} takes type of float64
	assert.Equal(t, int(merklerootsResponse.Page.LastEvaluatedKey.(float64)), expectedResult.Body.Page.LastEvaluatedKey.(int))
	assert.Equal(t, merklerootsResponse.Page.OrderByField, expectedResult.Body.Page.OrderByField)
	assert.Equal(t, merklerootsResponse.Page.SortDirection, expectedResult.Body.Page.SortDirection)
	assert.Equal(t, merklerootsResponse.Page.TotalElements, expectedResult.Body.Page.TotalElements)
	assert.Equal(t, merklerootsResponse.Page.Size, expectedResult.Body.Page.Size)
	for i, returnedMerkleroot := range merklerootsResponse.Content {
		expected := expectedResult.Body.Content[i]
		assert.Equal(t, returnedMerkleroot.BlockHeight, expected.BlockHeight)
		assert.Equal(t, returnedMerkleroot.MerkleRoot, expected.MerkleRoot)
	}
}

func TestReturnSuccessFromMerkleRootsWhenBatchSizeIsNotProvided(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithAPIAuthorizationDisabled(), testapp.WithLongestChain())
	orderedByField := "BlockHeight"
	sortDirection := "ASC"
	lastEvaluatedKey := "1"
	expectedResult := struct {
		Code int
		Body domains.MerkleRootsESKPagedResponse
	}{
		Code: http.StatusOK,
		Body: domains.MerkleRootsESKPagedResponse{
			Content: []*domains.MerkleRootsResponse{
				{
					MerkleRoot:  "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
					BlockHeight: 2,
				}, {
					MerkleRoot:  "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
					BlockHeight: 3,
				}, {
					MerkleRoot:  "df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a",
					BlockHeight: 4,
				},
			},
			Page: domains.ExclusiveStartKeyPage{
				OrderByField:     &orderedByField,
				SortDirection:    &sortDirection,
				TotalElements:    5,
				Size:             3,
				LastEvaluatedKey: 4,
			},
		},
	}
	defer cleanup()

	// when
	res := bhs.API().Call(getMerkleRoots("", lastEvaluatedKey))
	var merklerootsResponse domains.MerkleRootsESKPagedResponse
	json.NewDecoder(res.Body).Decode(&merklerootsResponse)

	// then
	assert.Equal(t, res.Code, expectedResult.Code)
	if res.Code != expectedResult.Code {
		t.Errorf("Expected to get status %d but instead got %d\n", expectedResult.Code, res.Code)
	}

	// this conversion is necessary as the default setting of the json decoder when unmarshalling JSON numbers into interface{} takes type of float64
	assert.Equal(t, int(merklerootsResponse.Page.LastEvaluatedKey.(float64)), expectedResult.Body.Page.LastEvaluatedKey.(int))
	assert.Equal(t, merklerootsResponse.Page.OrderByField, expectedResult.Body.Page.OrderByField)
	assert.Equal(t, merklerootsResponse.Page.SortDirection, expectedResult.Body.Page.SortDirection)
	assert.Equal(t, merklerootsResponse.Page.TotalElements, expectedResult.Body.Page.TotalElements)
	assert.Equal(t, merklerootsResponse.Page.Size, expectedResult.Body.Page.Size)
	for i, returnedMerkleroot := range merklerootsResponse.Content {
		expected := expectedResult.Body.Content[i]
		assert.Equal(t, returnedMerkleroot.BlockHeight, expected.BlockHeight)
		assert.Equal(t, returnedMerkleroot.MerkleRoot, expected.MerkleRoot)
	}
}

func TestReturnFailureFromMerkleRootsWhenAuthorizationIsTurnedOnAndCalledWithoutToken(t *testing.T) {
	// setup
	bhs, cleanup := testapp.NewTestBlockHeaderService(t)
	batchSize := "2"
	lastEvaluatedKey := "1"
	expectedResult := struct {
		Code int
		Body []byte
	}{
		Code: http.StatusUnauthorized,
		Body: []byte("\"empty auth header\""),
	}
	defer cleanup()

	// when
	res := bhs.API().Call(getMerkleRoots(batchSize, lastEvaluatedKey))

	// then
	assert.Equal(t, res.Code, expectedResult.Code)
	if res.Code != expectedResult.Code {
		t.Errorf("Expected to get status %d but instead got %d\n", expectedResult.Code, res.Code)
	}

	body, _ := io.ReadAll(res.Body)
	if !bytes.Equal(body, expectedResult.Body) {
		t.Errorf("Expected to get body %s but instead got %s\n", expectedResult.Body, body)
	}
}

// getMerkleRoots creates http request to fetch /chain/merkleroot it accepts two params
// batchSize and lastEvaluatedKey of type string, we can omit any of them to simulate
// user not passing these values and need to pass empty string in this place
func getMerkleRoots(batchSize, lastEvaluatedKey string) (req *http.Request, err error) {
	address := "/api/v1/chain/merkleroot"

	queryParams := []string{}
	if batchSize != "" {
		queryParams = append(queryParams, "batchSize="+batchSize)
	}
	if lastEvaluatedKey != "" {
		queryParams = append(queryParams, "lastEvaluatedKey="+lastEvaluatedKey)
	}

	if len(queryParams) > 0 {
		address = fmt.Sprintf("%s?%s", address, strings.Join(queryParams, "&"))
	}

	return http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		address,
		nil,
	)
}
