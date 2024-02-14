package websocket_test

import (
	"testing"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testapp"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/wait"
	"github.com/centrifugal/centrifuge-go"
)

func TestWebsocketCommunicationWithoutAuthentication(t *testing.T) {
	//setup
	p, cleanup := testapp.NewTestBlockHeaderService(t, testapp.WithApiAuthorizationDisabled())
	defer cleanup()

	//given
	client := p.Websocket().Client()
	defer client.Close()

	publisher := p.Websocket().Publisher()

	//when
	onMsg, err := client.Subscribe("test")

	//then
	assert.NoError(t, err)

	//when
	publisher.Publish("test", `{ "something": "value" }`)

	//then
	msg, err := wait.ForString(onMsg, time.Second)
	assert.NoError(t, err)
	assert.Equal(t, msg, `{ "something": "value" }`)
}

func TestWebsocketCommunicationWithAuthentication(t *testing.T) {
	//setup
	p, cleanup := testapp.NewTestBlockHeaderService(t)
	defer cleanup()

	//given
	client := p.Websocket().ClientWithConfig(centrifuge.Config{
		Token: "mQZQ6WmxURxWz5ch",
	})
	defer client.Close()

	//when
	_, err := client.Subscribe("test")

	//then
	assert.NoError(t, err)
}

func TestWebsocketCommunicationWithInvalidAuthentication(t *testing.T) {
	//setup
	p, cleanup := testapp.NewTestBlockHeaderService(t)
	defer cleanup()

	//given
	client := p.Websocket().ClientWithConfig(centrifuge.Config{
		Token: "invalid_token",
	})
	defer client.Close()

	//when
	_, err := client.Subscribe("test")

	//then
	assert.IsError(t, err, "invalid token")
}
