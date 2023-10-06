package fixtures

import "github.com/libsv/bitcoin-hc/config/p2pconfig"

func DefaultP2PConfig() p2pconfig.Config {
	return p2pconfig.Config{
		LogLevel:                  p2pconfig.DefaultLogLevel,
		MaxPeers:                  p2pconfig.DefaultMaxPeers,
		MaxPeersPerIP:             p2pconfig.DefaultMaxPeersPerIP,
		MinSyncPeerNetworkSpeed:   p2pconfig.DefaultMinSyncPeerNetworkSpeed,
		BanDuration:               p2pconfig.DefaultBanDuration,
		LogDir:                    p2pconfig.DefaultLogDir,
		ExcessiveBlockSize:        p2pconfig.DefaultExcessiveBlockSize,
		TrickleInterval:           p2pconfig.DefaultTrickleInterval,
		BlocksForForkConfirmation: p2pconfig.DefaultBlocksToConfirmFork,
		Logger:                    p2pconfig.UseDefaultP2PLogger(),
	}
}