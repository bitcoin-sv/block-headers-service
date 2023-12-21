package config

import "time"

// TrickleInterval is the interval at which a peer will send a getheaders.
const TrickleLinterval = 50 * time.Millisecond

// MaxPeers is the maximum number of peers the server will connect to (inbound and outbound).
const MaxPeers = 125

// MaxPeersPerIP is the maximum number of peers from specific IP.
const MaxPeersPerIP = 5

// MinSyncPeerNetworkSpeed is the minimum network speed required for a peer to be considered for syncing.
const MinSyncPeerNetworkSpeed = 51200

// ExcessiveBlockSize is the maximum block size we can accept.
const ExcessiveBlockSize = 128000000
