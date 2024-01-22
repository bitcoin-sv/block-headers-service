package tips_test

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"testing"

	"github.com/bitcoin-sv/pulse/domains"
	"github.com/bitcoin-sv/pulse/internal/tests/assert"
	"github.com/bitcoin-sv/pulse/internal/tests/fixtures"
	"github.com/bitcoin-sv/pulse/internal/tests/testpulse"
	"github.com/bitcoin-sv/pulse/transports/http/endpoints/api/tips"
)

var expected_tip = tips.TipStateResponse{
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
		pulse, cleanup := testpulse.NewTestPulse(t)
		defer cleanup()
		expected_result := struct {
			code int
			body []byte
		}{
			code: http.StatusUnauthorized,
			body: []byte("\"empty auth header\""),
		}

		// when
		res := pulse.Api().Call(getTips())

		// then
		assert.Equal(t, res.Code, expected_result.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expected_result.body)
	})

	t.Run("success", func(t *testing.T) {
		// given
		pulse, cleanup := testpulse.NewTestPulse(t, testpulse.WithLongestChain(), testpulse.WithApiAuthorizationDisabled())
		defer cleanup()
		expected_result := struct {
			code int
			body tips.TipStateResponse
		}{
			code: http.StatusOK,
			body: expected_tip,
		}

		// when
		res := pulse.Api().Call(getTips())

		// then
		assert.Equal(t, res.Code, expected_result.code)

		var tip []tips.TipStateResponse
		json.NewDecoder(res.Body).Decode(&tip)

		assert.Equal(t, len(tip), 1)
		assert.Equal(t, tip[0], expected_result.body)
	})
}

func TestGetTipLongest(t *testing.T) {
	t.Run("failure when authorization on and empty auth header", func(t *testing.T) {
		// given
		pulse, cleanup := testpulse.NewTestPulse(t)
		defer cleanup()
		expected_result := struct {
			code int
			body []byte
		}{
			code: http.StatusUnauthorized,
			body: []byte("\"empty auth header\""),
		}

		// when
		res := pulse.Api().Call(getTipLongestChain())

		// then
		assert.Equal(t, res.Code, expected_result.code)
		body, _ := io.ReadAll(res.Body)
		assert.EqualBytes(t, body, expected_result.body)
	})

	t.Run("success", func(t *testing.T) {
		// given
		pulse, cleanup := testpulse.NewTestPulse(t, testpulse.WithLongestChain(), testpulse.WithApiAuthorizationDisabled())
		defer cleanup()
		expected_result := struct {
			code int
			body tips.TipStateResponse
		}{
			code: http.StatusOK,
			body: expected_tip,
		}

		// when
		res := pulse.Api().Call(getTipLongestChain())

		// then
		assert.Equal(t, res.Code, expected_result.code)

		var tip tips.TipStateResponse
		json.NewDecoder(res.Body).Decode(&tip)

		assert.Equal(t, tip, expected_result.body)
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
