package peer

import (
	"errors"
	"sync"
)

type PeersCollection struct {
	peers []*Peer
	mu    sync.Mutex
}

func NewPeersCollection(size uint) *PeersCollection {
	return &PeersCollection{
		peers: make([]*Peer, size),
	}
}

func (col *PeersCollection) AddPeer(p *Peer) error {
	col.mu.Lock()
	defer col.mu.Unlock()

	for i, pp := range col.peers {
		if pp == nil {
			col.peers[i] = p
			return nil
		}
	}

	return errors.New("no space for new peer")
}

func (col *PeersCollection) RmPeer(p *Peer) {
	col.mu.Lock()
	defer col.mu.Unlock()

	for i, pp := range col.peers {
		if pp == p {
			col.peers[i] = nil
			return
		}
	}
}

func (col *PeersCollection) Space() uint {
	space := uint(0)

	col.mu.Lock()
	defer col.mu.Unlock()

	for _, p := range col.peers {
		if p == nil {
			space++
		}
	}

	return space
}

func (col *PeersCollection) Enumerate() []*Peer {
	col.mu.Lock()
	defer col.mu.Unlock()

	return col.peers[:]
}
