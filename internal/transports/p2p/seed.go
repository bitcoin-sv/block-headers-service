package p2pexp

import (
	"net"

	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/rs/zerolog"
)

func SeedFromDNS(seeds []chaincfg.DNSSeed, log *zerolog.Logger) []net.IP {
	allSeedPeers := make([]net.IP, 0)

	for _, seed := range seeds {
		host := seed.Host
		seedpeers, err := net.LookupIP(host)
		if err != nil {
			log.Info().Msgf("DNS discovery failed on seed %s: %v", host, err)
			continue
		}

		allSeedPeers = append(allSeedPeers, seedpeers...)
	}

	return allSeedPeers
}
