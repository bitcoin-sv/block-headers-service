package testapp

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/wait"
	"github.com/bitcoin-sv/block-headers-service/transports/websocket"
	"github.com/centrifugal/centrifuge-go"
)

// Websocket exposes functions to easy testing of block headers service websocket communication.
type Websocket struct {
	*TestBlockHeaderService
}

// WebsocketPublisher component used in tests to publish on websocket chanel.
type WebsocketPublisher struct {
	t         *testing.T
	log       *zerolog.Logger
	publisher websocket.Publisher
}

// Publisher creates WebsocketPublisher.
func (w *Websocket) Publisher() *WebsocketPublisher {
	return &WebsocketPublisher{
		t:         w.t,
		log:       w.log,
		publisher: w.ws.Publisher(),
	}
}

// Publish sends data to websocket channel.
func (p *WebsocketPublisher) Publish(channel string, data string) {
	p.log.Debug().Msgf("Trying to publish to channel %s data: %s", channel, data)
	_, err := p.publisher.Publish(channel, []byte(data))
	if err != nil {
		p.t.Fatalf("Couldn't publish a message")
	}
}

// WebsocketClient component used in tests to subscribe on websocket chanel.
type WebsocketClient struct {
	t            *testing.T
	log          *zerolog.Logger
	client       *centrifuge.Client
	connected    chan bool
	disconnected chan centrifuge.DisconnectedEvent
}

// Client creates WebsocketClient.
func (w *Websocket) Client() *WebsocketClient {
	return w.ClientWithConfig(centrifuge.Config{})
}

// ClientWithConfig creates WebsocketClient using provided config.
func (w *Websocket) ClientWithConfig(config centrifuge.Config) *WebsocketClient {
	client := centrifuge.NewJsonClient("ws://localhost:"+strconv.Itoa(w.port)+"/connection/websocket", config)
	logger := w.log
	return &WebsocketClient{
		t:            w.t,
		log:          logger,
		client:       client,
		connected:    make(chan bool, 1),
		disconnected: make(chan centrifuge.DisconnectedEvent, 1),
	}
}

// Unwrap returns *centrifuge.Client that is wrapped and configured by WebsocketClient.
func (c *WebsocketClient) Unwrap() *centrifuge.Client {
	return c.client
}

// Close closes the websocket client.
func (c *WebsocketClient) Close() {
	c.log.Debug().Msgf("Closing client connection")
	c.Unwrap().Close()
}

// Connect connects client to websocket.
func (c *WebsocketClient) Connect() error {
	c.configureClient()
	if err := c.client.Connect(); err != nil {
		return err
	}
	return nil
}

// Subscribe connects and subscribes to websocket channel.
// If no error then returned <-chan can be used to listen for messages on websocket channel.
func (c *WebsocketClient) Subscribe(channel string) (<-chan string, error) {
	if err := c.Connect(); err != nil {
		c.log.Error().Msgf("Error when trying to connect %v", err)
	}

	sub, subscribed, err := c.subscription(channel)
	if err != nil {
		return nil, err
	}

	receiver := c.subscribeToPublication(sub)

	if err := sub.Subscribe(); err != nil {
		return nil, err
	}

	err = c.waitForSubscribed(subscribed)
	if err != nil && errors.Is(err, wait.TimesOut) {
		err = fmt.Errorf("subscribing take longer then expected. %w", err)
		c.t.Fatal(err)
	}
	return receiver, err
}

func (c *WebsocketClient) waitForSubscribed(subscribed <-chan bool) error {
	timeout := time.Second
	select {
	case <-time.After(timeout):
		return fmt.Errorf("%w when subscribing after %s", wait.TimesOut, timeout)
	case d := <-c.disconnected:
		return errors.New(d.Reason)
	case <-subscribed:
		return nil
	}
}

func (c *WebsocketClient) configureClient() {
	client := c.client
	log := c.log
	client.OnConnecting(func(e centrifuge.ConnectingEvent) {
		log.Debug().Msgf("OnConnecting -> State: %d (%s)", e.Code, e.Reason)
	})
	client.OnConnected(func(e centrifuge.ConnectedEvent) {
		log.Debug().Msgf("OnConnected -> ID %s; data %s", e.ClientID, e.Data)
		c.connected <- true
	})
	client.OnDisconnected(func(e centrifuge.DisconnectedEvent) {
		log.Debug().Msgf("OnDisconnected -> State: %d (%s)", e.Code, e.Reason)
		c.disconnected <- e
	})

	client.OnError(func(e centrifuge.ErrorEvent) {
		log.Debug().Msgf("OnError -> %s", e.Error.Error())
	})

	client.OnMessage(func(e centrifuge.MessageEvent) {
		log.Debug().Msgf("OnMessage -> %s", string(e.Data))
	})

	client.OnSubscribed(func(e centrifuge.ServerSubscribedEvent) {
		log.Debug().Msgf("OnSubscribed -> channel %s: (was recovering: %v, recovered: %v)", e.Channel, e.WasRecovering, e.Recovered)
	})
	client.OnSubscribing(func(e centrifuge.ServerSubscribingEvent) {
		log.Debug().Msgf("OnSubscribing -> channel %s", e.Channel)
	})
	client.OnUnsubscribed(func(e centrifuge.ServerUnsubscribedEvent) {
		log.Debug().Msgf("OnUnsubscribed -> channel %s", e.Channel)
	})

	client.OnPublication(func(e centrifuge.ServerPublicationEvent) {
		log.Debug().Msgf("OnPublication -> channel %s: data: %s (offset %d)", e.Channel, e.Data, e.Offset)
	})
}

func (c *WebsocketClient) subscription(channel string) (*centrifuge.Subscription, <-chan bool, error) {
	subscribed := make(chan bool, 1)
	sub, err := c.client.NewSubscription(channel, centrifuge.SubscriptionConfig{})
	if err != nil {
		return nil, nil, err
	}
	sub.OnSubscribing(func(e centrifuge.SubscribingEvent) {
		c.log.Debug().Msgf("[sub] OnSubscribing -> channel %s: State: %d (%s)", sub.Channel, e.Code, e.Reason)
	})
	sub.OnSubscribed(func(e centrifuge.SubscribedEvent) {
		c.log.Debug().Msgf("[sub] OnSubscribed -> channel %s: (was recovering: %v, recovered: %v)", sub.Channel, e.WasRecovering, e.Recovered)
		subscribed <- true
		close(subscribed)
	})
	sub.OnUnsubscribed(func(e centrifuge.UnsubscribedEvent) {
		c.log.Debug().Msgf("[sub] OnUnsubscribed -> channel %s: State: %d (%s)", sub.Channel, e.Code, e.Reason)
	})

	sub.OnError(func(e centrifuge.SubscriptionErrorEvent) {
		c.log.Debug().Msgf("[sub] OnError -> channel %s: %s", sub.Channel, e.Error)
	})

	sub.OnPublication(func(e centrifuge.PublicationEvent) {
		c.log.Debug().Msgf("[sub] OnPublication -> channel %s: %s (offset %d)", sub.Channel, e.Data, e.Offset)
	})

	return sub, subscribed, err
}

func (c *WebsocketClient) subscribeToPublication(sub *centrifuge.Subscription) <-chan string {
	receiver := make(chan string, 10)
	sub.OnPublication(func(e centrifuge.PublicationEvent) {
		c.log.Debug().Msgf("[sub] OnPublication -> channel %s: %s (offset %d)", sub.Channel, e.Data, e.Offset)
		receiver <- string(e.Data)
	})
	return receiver
}
