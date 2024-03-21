package p2pexp

import (
	"errors"
	"net"
	"reflect"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/rs/zerolog"
)

type server struct {
	peers map[string]string
	log   *zerolog.Logger
}

func NewServer(log *zerolog.Logger) *server {
	serverLogger := log.With().Str("service", "p2p-experimental").Logger()
	return &server{peers: make(map[string]string), log: &serverLogger}
}

func (s *server) Start() error {
	seeds := SeedFromDNS(config.ActiveNetParams.DNSSeeds, s.log)
	if len(seeds) == 0 {
		return errors.New("no seeds found")
	}

	for _, seed := range seeds {
		s.log.Info().Msgf("Got peer addr: %s", seed.String())
	}

	firstPeer := seeds[0].String() + ":" + config.ActiveNetParams.DefaultPort
	conn, err := net.Dial("tcp", firstPeer)
	if err != nil {
		return err
	}
	defer conn.Close()

	s.log.Info().Msgf("connected to peer: %s", firstPeer)

	rmsg, _, err := wire.ReadMessage(conn, uint32(70013), config.ActiveNetParams.Net)
	if err != nil {
		return err
	}

	s.log.Info().Msgf("received msg type: %s", reflect.TypeOf(rmsg))

	switch msg := rmsg.(type) {
	case *wire.MsgVersion:
		s.log.Info().Msgf("got version msg: %v", msg)
	}

	return nil
}

func (s *server) Shutdown() error {
	return nil
}
