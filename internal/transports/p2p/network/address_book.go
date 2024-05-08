package network

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/transports/p2p/peer"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
)

// AddressBook represents a collection of known network addresses.
type AddressBook struct {
	banDuration  time.Duration
	addrs        []*knownAddress // addrs is a slice containing known addresses
	addrsLookup  map[string]int  // addrLookup is a map for fast lookup of addresses, maps address key to index in addrs slice
	mu           sync.Mutex
	addrFitlerFn func(*wire.NetAddress) bool
}

// NewAddressBook creates and initializes a new AddressBook instance.
func NewAdressBook(banDuration time.Duration, acceptLocalAddresses bool) *AddressBook {
	// Set the address filter function based on whether local addresses are accepted
	addrFilterFn := IsRoutable
	if acceptLocalAddresses {
		addrFilterFn = IsRoutableWithLocal
	}

	const addressesInitCapacity = 500
	return &AddressBook{
		addrs:        make([]*knownAddress, 0, addressesInitCapacity),
		addrsLookup:  make(map[string]int, addressesInitCapacity),
		banDuration:  banDuration,
		addrFitlerFn: addrFilterFn,
	}
}

// UpsertPeerAddr updates or adds a peer's address.
func (a *AddressBook) UpsertPeerAddr(p *peer.Peer) {
	a.mu.Lock()
	defer a.mu.Unlock()

	pa := p.GetPeerAddr()
	key, ka := a.findAddr(pa)

	if ka != nil {
		ka.peer = p
	} else {
		a.addAddr(key, &knownAddress{addr: pa, peer: p})
	}
}

// UpsertAddrs updates or adds multiple addresses.
func (a *AddressBook) UpsertAddrs(address []*wire.NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, addr := range address {
		if !a.addrFitlerFn(addr) {
			continue
		}

		key, ka := a.findAddr(addr)
		// If the address is not found, add it to the AddressBook.
		if ka == nil {
			a.addAddr(key, &knownAddress{addr: addr})
		} else if addr.Timestamp.After(ka.addr.Timestamp) {
			// Otherwise, update the timestamp if the new one is newer.
			ka.addr.Timestamp = addr.Timestamp
		}
	}
}

// BanAddr bans a network address. Ignores address if doesn't exist in the AddressBook.
func (a *AddressBook) BanAddr(addr *wire.NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ka := a.findAddr(addr)
	if ka != nil {
		now := time.Now()
		ka.banTimestamp = &now
	}
}

// GetRandUnusedAddr returns a randomly chosen unused network address.
func (a *AddressBook) GetRandUnusedAddr(tries uint) *wire.NetAddress {
	a.mu.Lock()
	defer a.mu.Unlock()

	alen := len(a.addrs)
	for i := uint(0); i < tries; i++ {
		// #nosec G404
		ka := a.addrs[rand.Intn(alen)]
		if ka.peer == nil {
			if ka.isBanned(a.banDuration) {
				continue
			}
			return ka.addr
		}
	}

	// return nil if no suitable address is found
	return nil
}

func (a *AddressBook) findAddr(addr *wire.NetAddress) (string, *knownAddress) {
	key := addrKey(addr)
	addrIndex, ok := a.addrsLookup[key]

	if ok {
		return key, a.addrs[addrIndex]
	}
	return key, nil
}

func (a *AddressBook) addAddr(key string, addr *knownAddress) {
	newItemIndex := len(a.addrs)

	a.addrs = append(a.addrs, addr)
	a.addrsLookup[key] = newItemIndex
}

func addrKey(addr *wire.NetAddress) string {
	return fmt.Sprintf("%s:%d", addr.IP, addr.Port)
}

type knownAddress struct {
	addr *wire.NetAddress
	peer *peer.Peer

	banTimestamp *time.Time
}

func (a *knownAddress) isBanned(duration time.Duration) bool {
	return a.banTimestamp != nil && time.Since(*a.banTimestamp) < duration
}
