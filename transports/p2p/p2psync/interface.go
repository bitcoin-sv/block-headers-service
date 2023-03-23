// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package p2psync

import (
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/wire"
	"github.com/libsv/bitcoin-hc/service"
	"github.com/libsv/bitcoin-hc/transports/p2p/p2plog"
	"github.com/libsv/bitcoin-hc/transports/p2p/peer"
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
	Logger       p2plog.Logger
	PeerNotifier PeerNotifier
	ChainParams  *chaincfg.Params

	DisableCheckpoints bool
	MaxPeers           int

	MinSyncPeerNetworkSpeed   uint64
	BlocksForForkConfirmation int

	Services *service.Services
}
