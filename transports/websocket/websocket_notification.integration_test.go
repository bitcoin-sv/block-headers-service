package websocket_test

import (
	"testing"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/fixtures"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testapp"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/wait"
	"github.com/kinbiko/jsonassert"
)

func TestShouldNotifyWebsocketAboutNewHeader(t *testing.T) {
	//setup
	p, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithApiAuthorizationDisabled())
	defer cleanup()

	//given
	client := p.Websocket().Client()
	defer client.Close()

	//when
	onMsg, err := client.Subscribe("headers")
	assert.NoError(t, err)

	//and
	err = p.When().NewHeaderReceived(*fixtures.HeaderSourceHeight1)
	assert.NoError(t, err)

	//then
	msg, err := wait.ForString(onMsg, time.Second)
	assert.NoError(t, err)

	expectedEvent := `{
						  "operation": "ADD",
						  "header": {
							"hash": "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048",
							"version": 1,
							"height": 1,
							"merkleRoot": "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
							"creationTimestamp": "2009-01-09T02:54:25Z",
							"nonce": 2573394689,
							"state": "LONGEST_CHAIN",
							"work": 8590065666,
							"prevBlockHash": "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"
						  }
						}`

	json := jsonassert.New(t)
	json.Assertf(msg, expectedEvent)
}
