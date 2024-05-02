package peer

import (
	"errors"
	"sync"
)

// PeersCollection represents a fixed size collection of peer objects with concurrency-safe operations.
type PeersCollection struct {
	peers []*Peer
	mu    sync.Mutex
}

// NewPeersCollection creates and initializes a new PeersCollection instance with the specified, fixed size.
func NewPeersCollection(size uint) *PeersCollection {
	return &PeersCollection{
		peers: make([]*Peer, size),
	}
}

// AddPeer adds a new peer to the PeersCollection.
// Returns an error if there is no space available for the new peer.
func (col *PeersCollection) AddPeer(p *Peer) error {
	col.mu.Lock()
	defer col.mu.Unlock()

	for i, pp := range col.peers {
		if pp == nil {
			col.peers[i] = p
			return nil
		}
	}

	return errors.New("no space available for new peer")
}

// RmPeer removes the specified peer from the PeersCollection. Ignores address if doesn't exist in the PeersCollection.
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

// Space returns the number of available slots for new peers in the PeersCollection.
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

// Enumerate returns a slice containing all non-nil peers in the PeersCollection. Order of peers in the returned slice is not guaranteed.
func (col *PeersCollection) Enumerate() []*Peer {
	col.mu.Lock()
	defer col.mu.Unlock()

	res := make([]*Peer, 0, len(col.peers))
	for _, p := range col.peers {
		if p != nil {
			res = append(res, p)
		}
	}

	return res
}
