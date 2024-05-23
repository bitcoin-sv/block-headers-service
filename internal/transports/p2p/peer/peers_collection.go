package peer

import (
	"errors"
	"slices"
	"sync"
)

// PeersCollection represents a fixed size collection of peer objects with concurrency-safe operations.
type PeersCollection struct {
	peers []*Peer
	size  uint
	mu    sync.Mutex
}

// NewPeersCollection creates and initializes a new PeersCollection instance with the specified, fixed size.
func NewPeersCollection(size uint) *PeersCollection {
	return &PeersCollection{
		size:  size,
		peers: make([]*Peer, 0, size),
	}
}

// AddPeer adds a new peer to the PeersCollection.
// Returns an error if there is no space available for the new peer.
func (col *PeersCollection) AddPeer(p *Peer) error {
	col.mu.Lock()
	defer col.mu.Unlock()

	if len(col.peers) == int(col.size) {
		return errors.New("no space available for new peer")
	}

	col.peers = append(col.peers, p)
	return nil
}

// RmPeer removes the specified peer from the PeersCollection. Ignores address if doesn't exist in the PeersCollection.
func (col *PeersCollection) RmPeer(p *Peer) {
	col.mu.Lock()
	defer col.mu.Unlock()

	// find index of peer
	pIndex := slices.Index(col.peers, p)
	if pIndex == -1 {
		return
	}

	// substitute with last element
	col.peers[pIndex] = col.peers[len(col.peers)-1]

	// remove last element
	col.peers = col.peers[:len(col.peers)-1]
}

// Space returns the number of available slots for new peers in the PeersCollection.
func (col *PeersCollection) Space() uint {
	col.mu.Lock()
	defer col.mu.Unlock()

	return col.size - uint(len(col.peers))
}

// Enumerate returns a slice containing all non-nil peers in the PeersCollection. Order of peers in the returned slice is not guaranteed.
func (col *PeersCollection) Enumerate() []*Peer {
	col.mu.Lock()
	defer col.mu.Unlock()

	// copy slice
	res := make([]*Peer, len(col.peers))
	copy(res, col.peers)

	return res
}
