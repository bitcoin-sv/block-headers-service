package p2pexp

import (
	"errors"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/rs/zerolog"
)

type server struct {
	config      *config.P2PConfig
	chainParams *chaincfg.Params
	log         *zerolog.Logger
}

func NewServer(config *config.P2PConfig, chainParams *chaincfg.Params, log *zerolog.Logger) *server {
	serverLogger := log.With().Str("service", "p2p-experimental").Logger()
	server := &server{
		config:      config,
		chainParams: chainParams,
		log:         &serverLogger,
	}
	return server
}

func (s *server) Start() error {
	seeds := SeedFromDNS(s.chainParams.DNSSeeds, s.log)
	if len(seeds) == 0 {
		return errors.New("no seeds found")
	}

	for _, seed := range seeds {
		s.log.Info().Msgf("Got peer addr: %s", seed.String())
	}

	peer, err := NewPeer(seeds[0].String(), s.config, s.chainParams, s.log)
	if err != nil {
		return err
	}

	err = peer.Connect()
	if err != nil {
		return err
	}
	defer peer.Disconnect()

	s.log.Info().Msgf("connected to peer: %s", peer.addr.String())

	peer.writeOurVersionMsg()

	return nil
}

func (s *server) Shutdown() error {
	return nil
}
