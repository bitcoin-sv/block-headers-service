package tips_test

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"testing"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/fixtures"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testapp"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/api/tips"
	"github.com/stretchr/testify/require"
)

var expectedTip = tips.TipStateResponse{
	Header: tips.TipResponse{
		Hash:             fixtures.HashHeight4.String(),
		Version:          fixtures.HeaderSourceHeight4.Version,
		PreviousBlock:    fixtures.HeaderSourceHeight4.PrevBlock.String(),
		MerkleRoot:       fixtures.HeaderSourceHeight4.MerkleRoot.String(),
		Timestamp:        uint32(fixtures.HeaderSourceHeight4.Timestamp.Unix()),
		DifficultyTarget: fixtures.HeaderSourceHeight4.Bits,
		Nonce:            fixtures.HeaderSourceHeight4.Nonce,
		Work:             big.NewInt(fixtures.DefaultChainWork),
	},
	State:     string(domains.LongestChain),
	ChainWork: big.NewInt(17180131332),
	Height:    4,
}

func TestGetTips(t *testing.T) {
	t.Run("failure when authorization on and empty auth header", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t)
		defer cleanup()
		expectedResult := struct {
			code int
			body string
		}{
			code: http.StatusUnauthorized,
			body: "{\"code\":\"ErrMissingAuthHeader\",\"message\":\"Empty auth header\"}",
		}

		// when
		res := bhs.API().Call(getTips())

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		require.JSONEq(t, expectedResult.body, res.Body.String())
	})

	t.Run("success", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body tips.TipStateResponse
		}{
			code: http.StatusOK,
			body: expectedTip,
		}

		// when
		res := bhs.API().Call(getTips())

		// then
		assert.Equal(t, res.Code, expectedResult.code)

		var tip []tips.TipStateResponse
		json.NewDecoder(res.Body).Decode(&tip)

		assert.Equal(t, len(tip), 1)
		assert.Equal(t, tip[0], expectedResult.body)
	})
}

func TestGetTipLongest(t *testing.T) {
	t.Run("failure when authorization on and empty auth header", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t)
		defer cleanup()
		expectedResult := struct {
			code int
			body string
		}{
			code: http.StatusUnauthorized,
			body: "{\"code\":\"ErrMissingAuthHeader\",\"message\":\"Empty auth header\"}",
		}

		// when
		res := bhs.API().Call(getTipLongestChain())

		// then
		assert.Equal(t, res.Code, expectedResult.code)
		require.JSONEq(t, expectedResult.body, res.Body.String())
	})

	t.Run("success", func(t *testing.T) {
		// given
		bhs, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithLongestChain(), testapp.WithAPIAuthorizationDisabled())
		defer cleanup()
		expectedResult := struct {
			code int
			body tips.TipStateResponse
		}{
			code: http.StatusOK,
			body: expectedTip,
		}

		// when
		res := bhs.API().Call(getTipLongestChain())

		// then
		assert.Equal(t, res.Code, expectedResult.code)

		var tip tips.TipStateResponse
		json.NewDecoder(res.Body).Decode(&tip)

		assert.Equal(t, tip, expectedResult.body)
	})
}

func getTips() (req *http.Request, err error) {
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/chain/tip",
		nil,
	)
}

func getTipLongestChain() (req *http.Request, err error) {
	return http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/chain/tip/longest",
		nil,
	)
}
