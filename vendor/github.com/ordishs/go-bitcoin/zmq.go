package bitcoin

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-zeromq/zmq4"
)

// ZMQ struct
type ZMQ struct {
	mu            sync.RWMutex
	address       string
	socket        zmq4.Socket
	connected     bool
	err           error
	topics        []string
	subscriptions map[string][]chan []string
}

// NewZMQ comment
func NewZMQ(host string, port int) *ZMQ {
	return newZMQ(host, port, false, "hash")
}

// NewZMQWithRaw creates a bitcoin ZMQ listener with raw enabled
func NewZMQWithRaw(host string, port int) *ZMQ {
	return newZMQ(host, port, true, "hash")
}

// NewZMQWithSubscribeOptionValue creates a bitcoin ZMQ listener with subscribe option value
func NewZMQWithSubscribeOptionValue(host string, port int, optionValue string) *ZMQ {
	return newZMQ(host, port, false, optionValue)
}

func newZMQ(host string, port int, rawRequired bool, optionValue string) *ZMQ {
	zmq := &ZMQ{
		address:       fmt.Sprintf("tcp://%s:%d", host, port),
		subscriptions: make(map[string][]chan []string),
		topics:        []string{"hashblock", "hashtx", "discardedfrommempool", "removedfrommempoolblock"},
	}

	if rawRequired {
		zmq.topics = append(zmq.topics, "rawblock")
		zmq.topics = append(zmq.topics, "rawtx")
	}

	go func() {
		for {
			zmq.socket = zmq4.NewSub(context.Background(), zmq4.WithID(zmq4.SocketIdentity("sub")))
			defer func() {
				if zmq.connected {
					zmq.socket.Close()
					zmq.connected = false
				}
			}()

			if err := zmq.socket.Dial(zmq.address); err != nil {
				zmq.mu.Lock()
				zmq.err = err
				zmq.mu.Unlock()
				log.Printf("Attempting to re-establish ZMQ connection in 5 seconds...")
				time.Sleep(10 * time.Second)
				continue
			}

			if err := zmq.socket.SetOption(zmq4.OptionSubscribe, optionValue); err != nil {
				zmq.err = fmt.Errorf("%+v", err)
				return
			}

			if rawRequired {
				if err := zmq.socket.SetOption(zmq4.OptionSubscribe, "raw"); err != nil {
					zmq.err = fmt.Errorf("%+v", err)
					return
				}
			}

			log.Printf("ZMQ: Subscribing to %s", zmq.address)

			//  0MQ is so fast, we need to wait a while...
			time.Sleep(time.Second)

			for {
				msg, err := zmq.socket.Recv()
				if err != nil {
					log.Printf("ERROR from zmq.socket.Recv() - %v\n", err)
					break
				} else {
					if !zmq.connected {
						zmq.connected = true
						log.Printf("ZMQ: Subscription to %s established\n", zmq.address)
					}

					// log.Printf("%s: %s", string(msg.Frames[0]), hex.EncodeToString(msg.Frames[1]))
					zmq.mu.RLock()
					subscribers := zmq.subscriptions[string(msg.Frames[0])]
					for _, subscriber := range subscribers {
						subscriber <- []string{string(msg.Frames[0]), hex.EncodeToString(msg.Frames[1])}
					}
					zmq.mu.RUnlock()
				}
			}
			if zmq.connected {
				zmq.socket.Close()
				zmq.connected = false
			}
			log.Printf("Attempting to re-establish ZMQ connection in 10 seconds...")
			time.Sleep(10 * time.Second)
		}
	}()

	return zmq
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Subscribe comment
func (zmq *ZMQ) Subscribe(topic string, ch chan []string) error {
	if !contains(zmq.topics, topic) {
		return fmt.Errorf("topic must be %+v, received %q", zmq.topics, topic)
	}

	zmq.mu.Lock()
	defer zmq.mu.Unlock()

	if zmq.err != nil {
		return fmt.Errorf("ERROR: ZMQ failed: %v", zmq.err)
	}

	subscribers := zmq.subscriptions[topic]
	subscribers = append(subscribers, ch)

	zmq.subscriptions[topic] = subscribers

	return nil
}

// Unsubscribe comment
func (zmq *ZMQ) Unsubscribe(topic string, ch chan []string) error {
	if !contains(zmq.topics, topic) {
		return fmt.Errorf("topic must be %+v, received %q", zmq.topics, topic)
	}

	zmq.mu.Lock()
	defer zmq.mu.Unlock()

	subscribers := zmq.subscriptions[topic]
	for i, channel := range subscribers {
		if channel == ch {
			subscribers = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}
	zmq.subscriptions[topic] = subscribers

	return nil
}
