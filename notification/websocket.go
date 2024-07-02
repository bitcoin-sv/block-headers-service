package notification

import (
	"encoding/json"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/centrifugal/centrifuge"
	"github.com/rs/zerolog"
)

// WebsocketPublisher represents websocket server entrypoint used to publish messages via websocket communication.
type WebsocketPublisher interface {
	Publish(channel string, data []byte, opts ...centrifuge.PublishOption) (centrifuge.PublishResult, error)
}

type wsChan struct {
	publisher      WebsocketPublisher
	log            *zerolog.Logger
	historySize    int
	historySeconds int
}

// NewWebsocketChannel create Channel implementation communicating via websocket.
func NewWebsocketChannel(log *zerolog.Logger, publisher WebsocketPublisher, cfg *config.WebsocketConfig) Channel {
	channelLogger := log.With().Str("subservice", "ws-channel").Logger()
	return &wsChan{
		publisher:      publisher,
		log:            &channelLogger,
		historySize:    cfg.HistoryMax,
		historySeconds: cfg.HistoryTTL,
	}
}

func (w *wsChan) Notify(event Event) {
	bytes, err := json.Marshal(event)
	if err != nil {
		w.log.Error().Msgf("Error when creating json from event %v: %v", event, err)
		return
	}

	if err := w.publishToHeadersChannel(bytes); err != nil {
		w.log.Error().Msgf("Error when sending event %v to channel: %v", event, err)
		return
	}
}

func (w *wsChan) publishToHeadersChannel(bytes []byte) error {
	_, err := w.publisher.Publish("headers", bytes,
		centrifuge.WithHistory(w.historySize, time.Duration(w.historySeconds)*time.Minute))
	return err
}
