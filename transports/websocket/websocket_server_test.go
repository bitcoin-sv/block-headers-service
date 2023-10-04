package websocket_test

import (
	"testing"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/libsv/bitcoin-hc/internal/tests/assert"
	"github.com/libsv/bitcoin-hc/internal/tests/testpulse"
	"github.com/libsv/bitcoin-hc/internal/tests/wait"
)

func TestWebsocketCommunicationWithoutAuthentication(t *testing.T) {
	//setup
	p, cleanup := testpulse.NewTestPulse(t, testpulse.WithoutApiAuthorization())
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
	p, cleanup := testpulse.NewTestPulse(t)
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
	p, cleanup := testpulse.NewTestPulse(t)
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
