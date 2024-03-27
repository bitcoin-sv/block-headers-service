package p2pexp

import (
	"errors"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/transports/p2p/peer"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/rs/zerolog"
)

type server struct {
	config         *config.P2PConfig
	chainParams    *chaincfg.Params
	headersService service.Headers
	log            *zerolog.Logger

	peer *peer.Peer
}

func NewServer(config *config.P2PConfig, chainParams *chaincfg.Params, headersService service.Headers, log *zerolog.Logger) *server {
	serverLogger := log.With().Str("service", "p2p-experimental").Logger()
	server := &server{
		config:         config,
		chainParams:    chainParams,
		headersService: headersService,
		log:            &serverLogger,
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

	peer, err := peer.NewPeer(seeds[0].String(), s.config, s.chainParams, s.headersService, s.log)
	if err != nil {
		return err
	}

	s.peer = peer

	err = peer.Connect()
	if err != nil {
		return err
	}

	err = peer.Start()
	if err != nil {
		s.log.Error().Msgf("error starting peer, reason: %v", err)
		return err
	}

	return nil
}

func (s *server) Shutdown() error {
	s.peer.Disconnect()
	return nil
}
