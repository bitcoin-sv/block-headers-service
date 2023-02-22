package bitcoin

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-zeromq/zmq4"
)

var allowedTopics = []string{"hashblock", "hashtx", "hashtx2", "rawblock", "rawtx", "rawtx2", "discardedfrommempool", "removedfrommempoolblock", "invalidtx"}

type subscriptionRequest struct {
	topic string
	ch    chan []string
}

// ZMQ struct
type ZMQ struct {
	address            string
	socket             zmq4.Socket
	connected          bool
	err                error
	subscriptions      map[string][]chan []string
	addSubscription    chan subscriptionRequest
	removeSubscription chan subscriptionRequest
	logger             Logger
}

func NewZMQ(host string, port int, optionalLogger ...Logger) *ZMQ {

	zmq := &ZMQ{
		address:            fmt.Sprintf("tcp://%s:%d", host, port),
		subscriptions:      make(map[string][]chan []string),
		addSubscription:    make(chan subscriptionRequest, 10),
		removeSubscription: make(chan subscriptionRequest, 10),
		logger:             &DefaultLogger{},
	}

	if len(optionalLogger) > 0 {
		zmq.logger = optionalLogger[0]
	}

	go zmq.start()

	return zmq
}

func (zmq *ZMQ) Subscribe(topic string, ch chan []string) error {
	if !contains(allowedTopics, topic) {
		return fmt.Errorf("topic must be %+v, received %q", allowedTopics, topic)
	}

	zmq.addSubscription <- subscriptionRequest{
		topic: topic,
		ch:    ch,
	}

	return nil
}

func (zmq *ZMQ) Unsubscribe(topic string, ch chan []string) error {
	if !contains(allowedTopics, topic) {
		return fmt.Errorf("topic must be %+v, received %q", allowedTopics, topic)
	}

	zmq.removeSubscription <- subscriptionRequest{
		topic: topic,
		ch:    ch,
	}

	return nil
}

func (zmq *ZMQ) start() {
	for {
		zmq.socket = zmq4.NewSub(context.Background(), zmq4.WithID(zmq4.SocketIdentity("sub")))
		defer func() {
			if zmq.connected {
				zmq.socket.Close()
				zmq.connected = false
			}
		}()

		if err := zmq.socket.Dial(zmq.address); err != nil {
			zmq.err = err
			zmq.logger.Errorf("Could not dial ZMQ at %s: %v", zmq.address, err)
			zmq.logger.Infof("Attempting to re-establish ZMQ connection in 10 seconds...")
			time.Sleep(10 * time.Second)
			continue
		}

		zmq.logger.Infof("ZMQ: Connecting to %s", zmq.address)

		for topic := range zmq.subscriptions {
			if err := zmq.socket.SetOption(zmq4.OptionSubscribe, topic); err != nil {
				zmq.err = fmt.Errorf("%+v", err)
				return
			}
			zmq.logger.Infof("ZMQ: Subscribed to %s", topic)
		}

	OUT:
		for {
			select {
			case req := <-zmq.addSubscription:
				if err := zmq.socket.SetOption(zmq4.OptionSubscribe, req.topic); err != nil {
					zmq.logger.Errorf("ZMQ: Failed to subscribe to %s", req.topic)
				} else {
					zmq.logger.Infof("ZMQ: Subscribed to %s", req.topic)
				}

				subscribers := zmq.subscriptions[req.topic]
				subscribers = append(subscribers, req.ch)

				zmq.subscriptions[req.topic] = subscribers

			case req := <-zmq.removeSubscription:
				subscribers := zmq.subscriptions[req.topic]
				for i, subscriber := range subscribers {
					if subscriber == req.ch {
						subscribers = append(subscribers[:i], subscribers[i+1:]...)
						zmq.logger.Infof("Removed subscription from %s topic", req.topic)
						break
					}
				}
				zmq.subscriptions[req.topic] = subscribers

			default:
				msg, err := zmq.socket.Recv()
				if err != nil {
					zmq.logger.Errorf("zmq.socket.Recv() - %v\n", err)
					break OUT
				} else {
					if !zmq.connected {
						zmq.connected = true
						zmq.logger.Infof("ZMQ: Connection to %s observed\n", zmq.address)
					}

					subscribers := zmq.subscriptions[string(msg.Frames[0])]

					sequence := "N/A"

					if len(msg.Frames) > 2 && len(msg.Frames[2]) == 4 {
						s := binary.LittleEndian.Uint32(msg.Frames[2])
						sequence = strconv.FormatInt(int64(s), 10)
					}

					for _, subscriber := range subscribers {
						subscriber <- []string{string(msg.Frames[0]), hex.EncodeToString(msg.Frames[1]), sequence}
					}
				}
			}
		}

		if zmq.connected {
			zmq.socket.Close()
			zmq.connected = false
		}
		log.Printf("Attempting to re-establish ZMQ connection in 10 seconds...")
		time.Sleep(10 * time.Second)

	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
