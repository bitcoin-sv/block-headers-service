package service

import (
	peerpkg "github.com/libsv/bitcoin-hc/transports/p2p/peer"
)

// NetworkService represents Network service and provide access to repositories.
type NetworkService struct {
	peers map[*peerpkg.Peer]*peerpkg.PeerSyncState
}

// GetPeers return all currently connected peers.
func (s *NetworkService) GetPeers() []peerpkg.PeerState {
	peerStates := make([]peerpkg.PeerState, 0)

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
func NewNetworkService(peers map[*peerpkg.Peer]*peerpkg.PeerSyncState) *NetworkService {
	return &NetworkService{
		peers: peers,
	}
}
