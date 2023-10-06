package notification

import (
	"encoding/json"
	"time"

	"github.com/centrifugal/centrifuge"
	"github.com/libsv/bitcoin-hc/config"
	"github.com/libsv/bitcoin-hc/domains/logging"
)

// WebsocketPublisher represents websocket server entrypoint used to publish messages via websocket communication.
type WebsocketPublisher interface {
	Publish(channel string, data []byte, opts ...centrifuge.PublishOption) (centrifuge.PublishResult, error)
}

type wsChan struct {
	publisher      WebsocketPublisher
	log            logging.Logger
	historySize    int
	historySeconds int
}

// NewWebsocketChannel create Channel implementation communicating via websocket.
func NewWebsocketChannel(lf logging.LoggerFactory, publisher WebsocketPublisher, cfg *config.Websocket) Channel {
	return &wsChan{
		publisher:      publisher,
		log:            lf.NewLogger("ws-channel"),
		historySize:    cfg.HistoryMax,
		historySeconds: cfg.HistoryTTL,
	}
}

func (w *wsChan) Notify(event Event) {
	bytes, err := json.Marshal(event)
	if err != nil {
		w.log.Errorf("Error when creating json from event %v: %v", event, err)
		return
	}

	if err := w.publishToHeadersChannel(bytes); err != nil {
		w.log.Errorf("Error when sending event %v to channel: %v", event, err)
		return
	}
}

func (w *wsChan) publishToHeadersChannel(bytes []byte) error {
	_, err := w.publisher.Publish("headers", bytes,
		centrifuge.WithHistory(w.historySize, time.Duration(w.historySeconds)*time.Minute))
	return err
}
