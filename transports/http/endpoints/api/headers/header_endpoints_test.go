package headers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/fixtures"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testapp"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/api/headers"
)

var expectedObj = headers.BlockHeaderResponse{
	Hash:             fixtures.HashHeight1.String(),
	Version:          fixtures.HeaderSourceHeight1.Version,
	PreviousBlock:    fixtures.HeaderSourceHeight1.PrevBlock.String(),
	MerkleRoot:       fixtures.HeaderSourceHeight1.MerkleRoot.String(),
	Timestamp:        uint32(fixtures.HeaderSourceHeight1.Timestamp.Unix()),
	DifficultyTarget: fixtures.HeaderSourceHeight1.Bits,
	Nonce:            fixtures.HeaderSourceHeight1.Nonce,
	Work:             strconv.Itoa(fixtures.DefaultChainWork),
}

func TestGetHeaderByHash(t *testing.T) {
	t.Run("failure when authorization on and empty auth header", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t)
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusUnauthorized,
			body: []byte("\"empty auth header\""),
		}

		// when
		res := bhs.API().Call(getHeaderByHash("123"))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})

	t.Run("success", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body headers.BlockHeaderResponse
		}{
			code: http.StatusOK,
			body: expectedObj,
		}

		// when
		res := bhs.API().Call(getHeaderByHash(fixtures.HashHeight1.String()))

		// then
		assert.Equal(t, res.Code, expectedResult.code)

		var header headers.BlockHeaderResponse
		json.NewDecoder(res.Body).Decode(&header)

		assert.Equal(t, header, expectedResult.body)
	})

	t.Run("failure - hash not found", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusBadRequest,
			body: []byte("\"could not find hash\""),
		}

		// when
		res := bhs.API().Call(getHeaderByHash("123"))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})
}

func TestGetHeaderByHeight(t *testing.T) {
	t.Run("failure when authorization on and empty auth header", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t)
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusUnauthorized,
			body: []byte("\"empty auth header\""),
		}

		// when
		res := bhs.API().Call(getHeaderByHeight(123, 1))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})

	t.Run("success", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body headers.BlockHeaderResponse
		}{
			code: http.StatusOK,
			body: expectedObj,
		}

		// when
		res := bhs.API().Call(getHeaderByHeight(1, 1))

		// then
		assert.Equal(t, res.Code, expectedResult.code)

		var header []headers.BlockHeaderResponse
		json.NewDecoder(res.Body).Decode(&header)

		assert.Equal(t, header[0], expectedResult.body)
	})

	t.Run("failure - hash not found", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusBadRequest,
			body: []byte("\"could not find headers in given range\""),
		}

		// when
		res := bhs.API().Call(getHeaderByHeight(123, 1))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})
}

func TestGetHeaderAncestorsByHash(t *testing.T) {
	t.Run("failure when authorization on and empty auth header", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t)
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusUnauthorized,
			body: []byte("\"empty auth header\""),
		}

		// when
		res := bhs.API().Call(getHeaderAncestorsByHash("123", "1234"))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})

	t.Run("success", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body headers.BlockHeaderResponse
		}{
			code: http.StatusOK,
			body: expectedObj,
		}

		// when
		res := bhs.API().Call(getHeaderAncestorsByHash(fixtures.HashHeight2.String(), fixtures.HashHeight1.String()))

		// then
		assert.Equal(t, res.Code, expectedResult.code)

		var header []headers.BlockHeaderResponse
		json.NewDecoder(res.Body).Decode(&header)

		assert.Equal(t, header[0], expectedResult.body)
	})

	t.Run("failure - hash not found", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusBadRequest,
			body: []byte("\"error during getting headers with given hashes\""),
		}

		// when
		res := bhs.API().Call(getHeaderAncestorsByHash("123", "1234"))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})
}

func TestGetCommonAncestor(t *testing.T) {
	t.Run("failure when authorization on and empty auth header", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t)
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusUnauthorized,
			body: []byte("\"empty auth header\""),
		}

		// when
		res := bhs.API().Call(getCommonAncestors([]string{"123", "1234"}))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})

	t.Run("success", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		genesis := chaincfg.MainNetParams.GenesisBlock.Header
		expectedResponse := headers.BlockHeaderResponse{
			Hash:             genesis.BlockHash().String(),
			Version:          genesis.Version,
			PreviousBlock:    chainhash.Hash{}.String(),
			MerkleRoot:       genesis.MerkleRoot.String(),
			Timestamp:        uint32(genesis.Timestamp.Unix()),
			DifficultyTarget: genesis.Bits,
			Nonce:            genesis.Nonce,
			Work:             domains.CalculateWork(genesis.Bits).BigInt().String(),
		}
		expectedResult := struct {
			code int
			body headers.BlockHeaderResponse
		}{
			code: http.StatusOK,
			body: expectedResponse,
		}

		// when
		res := bhs.API().Call(getCommonAncestors([]string{fixtures.HashHeight2.String(), fixtures.HashHeight1.String()}))

		// then
		assert.Equal(t, res.Code, expectedResult.code)

		var header headers.BlockHeaderResponse
		json.NewDecoder(res.Body).Decode(&header)

		assert.Equal(t, header, expectedResult.body)
	})

	t.Run("failure - hash not found", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusBadRequest,
			body: []byte("\"could not find hash\""),
		}

		// when
		res := bhs.API().Call(getCommonAncestors([]string{"123", "1234"}))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})
}

func TestGetHeadersState(t *testing.T) {
	t.Run("failure when authorization on and empty auth header", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t)
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusUnauthorized,
			body: []byte("\"empty auth header\""),
		}

		// when
		res := bhs.API().Call(getHeadersState("123"))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})

	t.Run("success", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResponse := headers.BlockHeaderStateResponse{
			Header:    expectedObj,
			State:     string(domains.LongestChain),
			ChainWork: strconv.Itoa(fixtures.DefaultChainWork),
			Height:    1,
		}
		expectedResult := struct {
			code int
			body headers.BlockHeaderStateResponse
		}{
			code: http.StatusOK,
			body: expectedResponse,
		}

		// when
		res := bhs.API().Call(getHeadersState(fixtures.HashHeight1.String()))

		// then
		assert.Equal(t, res.Code, expectedResult.code)

		var header headers.BlockHeaderStateResponse
		json.NewDecoder(res.Body).Decode(&header)

		assert.Equal(t, header, expectedResult.body)
	})

	t.Run("failure - hash not found", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body []byte
		}{
			code: http.StatusBadRequest,
			body: []byte("\"could not find hash\""),
		}

		// when
		res := bhs.API().Call(getHeadersState("123"))

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expectedResult.body)
	})
}

func getHeaderByHash(hash string) (req *http.Request, err error) {
	address := fmt.Sprintf("/api/v1/chain/header/%s", hash)
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		address,
		nil,
	)
}

func getHeaderByHeight(height, count int) (req *http.Request, err error) {
	address := fmt.Sprintf("/api/v1/chain/header/byHeight?height=%d&count=%d", height, count)
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		address,
		nil,
	)
}

func getHeaderAncestorsByHash(hash, ancestorHash string) (req *http.Request, err error) {
	address := fmt.Sprintf("/api/v1/chain/header/%s/%s/ancestor", hash, ancestorHash)
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		address,
		nil,
	)
}

func getCommonAncestors(ancestors []string) (req *http.Request, err error) {
	array, err := json.Marshal(ancestors)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(array)
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/chain/header/commonAncestor",
		body,
	)
}

func getHeadersState(hash string) (req *http.Request, err error) {
	address := fmt.Sprintf("/api/v1/chain/header/state/%s", hash)
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		address,
		nil,
	)
}
