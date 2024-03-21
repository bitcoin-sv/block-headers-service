package p2pexp

import (
	"fmt"
	"net"
	"strings"

	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/rs/zerolog"
)

func SeedFromDNS(seeds []chaincfg.DNSSeed, log *zerolog.Logger) []net.IP {
	allSeedPeers := make([]net.IP, 0)

	for _, seed := range seeds {
		host := seed.Host
		seedpeers, err := bsvdLookup(host)
		if err != nil {
			log.Info().Msgf("DNS discovery failed on seed %s: %v", host, err)
			continue
		}

		allSeedPeers = append(allSeedPeers, seedpeers...)
	}

	return allSeedPeers
}

func bsvdLookup(host string) ([]net.IP, error) {
	if strings.HasSuffix(host, ".onion") {
		return nil, fmt.Errorf("attempt to resolve tor address: %s", host)
	}

	return net.LookupIP(host)
}
