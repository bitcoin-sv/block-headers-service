package websocket

import (
	"context"
	"fmt"

	"github.com/centrifugal/centrifuge"
	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/service"
)

// Publisher component exposed by server that is providing a way to send messages via websocket.
type Publisher interface {
	Publish(channel string, data []byte, opts ...centrifuge.PublishOption) (centrifuge.PublishResult, error)
}

// Server websocket server controller.
type Server interface {
	Start() error
	Shutdown() error
	SetupEntrypoint(*gin.Engine)
	Publisher() Publisher
}

type server struct {
	node           *centrifuge.Node
	isAuthRequired bool
	tokens         service.Tokens
	log            logging.Logger
}

// NewServer creates new websocket server.
func NewServer(lf logging.LoggerFactory, services *service.Services, isAuthenticationOn bool) (Server, error) {
	node, err := newNode(lf)
	if err != nil {
		return nil, err
	}
	s := &server{
		node:           node,
		isAuthRequired: isAuthenticationOn,
		tokens:         services.Tokens,
		log:            lf.NewLogger("Websocket"),
	}
	return s, nil
}

// Start starts a server.
func (s *server) Start() error {
	s.setupNode()
	if err := s.node.Run(); err != nil {
		return fmt.Errorf("cannot start websocket server: %w", err)
	}
	return nil
}

// Shutdown stoping a server.
func (s *server) Shutdown() error {
	return s.ShutdownWithContext(context.Background())
}

// ShutdownWithContext stoping a server in a provided context.
func (s *server) ShutdownWithContext(ctx context.Context) error {
	s.log.Infof("Shutting down a websocket server")
	if err := s.node.Shutdown(ctx); err != nil {
		return fmt.Errorf("cannot stop websocket server: %w", err)
	}
	return nil
}

// SetupEntrypoint setup gin to init websocket connection.
func (s *server) SetupEntrypoint(engine *gin.Engine) {
	engine.GET("/connection/websocket", gin.WrapH(centrifuge.NewWebsocketHandler(s.node, centrifuge.WebsocketConfig{})))
}

// Publisher returns websocket Publisher component.
func (s *server) Publisher() Publisher {
	return s.node
}

func newNode(lf logging.LoggerFactory) (*centrifuge.Node, error) {
	lh := newLogHandler(lf)
	return centrifuge.New(centrifuge.Config{
		Name:       "Pulse",
		LogLevel:   lh.Level(),
		LogHandler: lh.Log,
	})
}

func (s *server) setupNode() {
	s.node.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
		s.log.Info("client connecting")

		if s.isAuthRequired {
			s.log.Debugf("client connecting with token: %s", event.Token)
			_, err := s.tokens.GetToken(event.Token)
			if err != nil {
				return centrifuge.ConnectReply{}, centrifuge.DisconnectInvalidToken
			}
		}

		return centrifuge.ConnectReply{
			Credentials: &centrifuge.Credentials{
				UserID: "",
			},
		}, nil
	})

	s.node.OnConnect(func(client *centrifuge.Client) {
		transport := client.Transport()
		s.log.Infof("user %s connected via %s.", client.UserID(), transport.Name())

		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			s.log.Infof("user %s subscribes on %s", client.UserID(), e.Channel)
			cb(centrifuge.SubscribeReply{
				Options: centrifuge.SubscribeOptions{
					EnablePositioning: true,
					EnableRecovery:    true,
				},
			}, nil)
		})

		client.OnHistory(func(e centrifuge.HistoryEvent, cb centrifuge.HistoryCallback) {
			if !client.IsSubscribed(e.Channel) {
				cb(centrifuge.HistoryReply{}, centrifuge.ErrorPermissionDenied)
				return
			}
			cb(centrifuge.HistoryReply{}, nil)
		})

		client.OnUnsubscribe(func(e centrifuge.UnsubscribeEvent) {
			s.log.Infof("user %s unsubscribed from %s", client.UserID(), e.Channel)
		})

		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			s.log.Infof("user %s disconnected, disconnect: %s", client.UserID(), e.Disconnect)
		})
	})
}
