package service

import (
	peerpkg "github.com/bitcoin-sv/block-headers-service/transports/p2p/peer"
)

// NetworkService represents Network service and provide access to repositories.
type NetworkService struct {
	peers map[*peerpkg.Peer]*peerpkg.SyncState
}

// GetPeers return all currently connected peers.
func (s *NetworkService) GetPeers() []peerpkg.State {
	peerStates := make([]peerpkg.State, 0)

	for peer := range s.peers {
		peerStates = append(peerStates, peer.ToPeerState())
	}
	return peerStates
}

// GetPeersCount return number of currently connected peers.
func (s *NetworkService) GetPeersCount() int {
	if s.peers == nil {
		return 0
	}

	length := len(s.peers)
	return length
}

// NewNetworkService creates and returns NetworkService instance.
func NewNetworkService(peers map[*peerpkg.Peer]*peerpkg.SyncState) *NetworkService {
	return &NetworkService{
		peers: peers,
	}
}
