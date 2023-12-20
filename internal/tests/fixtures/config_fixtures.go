package fixtures

import "github.com/bitcoin-sv/pulse/config"

// DefaultP2PConfig default p2p config for test purposes.
func DefaultP2PConfig() config.P2PConfig {
	return config.P2PConfig{
		LogLevel:                  config.DefaultLogLevel,
		MaxPeers:                  config.DefaultMaxPeers,
		MaxPeersPerIP:             config.DefaultMaxPeersPerIP,
		MinSyncPeerNetworkSpeed:   config.DefaultMinSyncPeerNetworkSpeed,
		BanDuration:               config.DefaultBanDuration,
		LogDir:                    config.DefaultLogDir,
		ExcessiveBlockSize:        config.DefaultExcessiveBlockSize,
		TrickleInterval:           config.DefaultTrickleInterval,
		BlocksForForkConfirmation: config.DefaultBlocksToConfirmFork,
	}
}
