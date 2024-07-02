package websocket

import (
	"context"
	"fmt"

	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/centrifugal/centrifuge"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
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
	log            *zerolog.Logger
}

// NewServer creates new websocket server.
func NewServer(log *zerolog.Logger, services *service.Services, isAuthenticationOn bool) (Server, error) {
	websocketLogger := log.With().Str("subservice", "websocket-server").Logger()
	node, err := newNode(&websocketLogger)
	if err != nil {
		return nil, err
	}
	s := &server{
		node:           node,
		isAuthRequired: isAuthenticationOn,
		tokens:         services.Tokens,
		log:            &websocketLogger,
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
	s.log.Info().Msgf("Shutting down a websocket server")
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

func newNode(log *zerolog.Logger) (*centrifuge.Node, error) {
	lh := newLogHandler(log)
	return centrifuge.New(centrifuge.Config{
		Name:       "Block-Headers-Service",
		LogLevel:   lh.Level(),
		LogHandler: lh.Log,
	})
}

func (s *server) setupNode() {
	s.node.OnConnecting(func(_ context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
		s.log.Info().Msg("client connecting")

		if s.isAuthRequired {
			s.log.Debug().Msgf("client connecting with token: %s", event.Token)
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
		s.log.Info().Msgf("user %s connected via %s.", client.UserID(), transport.Name())

		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			s.log.Info().Msgf("user %s subscribes on %s", client.UserID(), e.Channel)
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
			s.log.Info().Msgf("user %s unsubscribed from %s", client.UserID(), e.Channel)
		})

		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			s.log.Info().Msgf("user %s disconnected, disconnect: %s", client.UserID(), e.Disconnect)
		})
	})
}
