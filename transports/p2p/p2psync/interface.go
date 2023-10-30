// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package p2psync

import (
	"github.com/bitcoin-sv/pulse/domains/logging"
	"github.com/bitcoin-sv/pulse/internal/chaincfg"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/pulse/internal/wire"
	"github.com/bitcoin-sv/pulse/service"
	"github.com/bitcoin-sv/pulse/transports/p2p/peer"
)

// PeerNotifier exposes methods to notify peers of status changes to
// transactions, blocks, etc. Currently server (in the main package) implements
// this interface.
type PeerNotifier interface {
	UpdatePeerHeights(latestBlkHash *chainhash.Hash, latestHeight int32, updateSource *peer.Peer)

	RelayInventory(invVect *wire.InvVect, data interface{})

	BanPeer(sp *peer.Peer)
}

// Config is a configuration struct used to initialize a new SyncManager.
type Config struct {
	LoggerFactory logging.LoggerFactory
	PeerNotifier  PeerNotifier
	ChainParams   *chaincfg.Params

	DisableCheckpoints bool
	MaxPeers           int

	MinSyncPeerNetworkSpeed   uint64
	BlocksForForkConfirmation int

	Services    *service.Services
	Checkpoints []chaincfg.Checkpoint
}
