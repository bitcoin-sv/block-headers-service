package service

import (
	peerpkg "github.com/libsv/bitcoin-hc/transports/p2p/peer"
)

type NetworkService struct {
	peers map[*peerpkg.Peer]*peerpkg.PeerSyncState
}

func (s *NetworkService) GetPeers() []peerpkg.PeerState {
	peerStates := make([]peerpkg.PeerState, 0)

	for peer := range s.peers {
		peerStates = append(peerStates, peer.ToPeerState())
	}
	return peerStates
}

func (s *NetworkService) GetPeersCount() int {
	if s.peers == nil {
		return 0
	}

	length := len(s.peers)
	return length
}

func NewNetworkService(peers map[*peerpkg.Peer]*peerpkg.PeerSyncState) *NetworkService {
	return &NetworkService{
		peers: peers,
	}
}
