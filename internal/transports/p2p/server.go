package p2pexp

import (
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/rs/zerolog"
)

var mainNetDNSSeeds = []chaincfg.DNSSeed{
	{Host: "seed-nodes.bsvb.tech", HasFiltering: true},
}

type server struct {
	peers map[string]string
	log   *zerolog.Logger
}

func NewServer(log *zerolog.Logger) *server {
	serverLogger := log.With().Str("service", "p2p-experimental").Logger()
	return &server{peers: make(map[string]string), log: &serverLogger}
}

func (s *server) Start() error {
	seeds := SeedFromDNS(mainNetDNSSeeds, s.log)
	for _, seed := range seeds {
		s.log.Info().Msgf("Got peer addr: %s", seed.String())
	}
	return nil
}

func (s *server) Shutdown() error {
	return nil
}
