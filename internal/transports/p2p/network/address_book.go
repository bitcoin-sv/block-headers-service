package network

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/transports/p2p/peer"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
)

type AddressBook struct {
	banDuration time.Duration
	addrs       []*knownAddress
	keyIndex    map[string]int

	mu           sync.Mutex
	addrFitlerFn func(*wire.NetAddress) bool
}

func NewAdressbook(banDuration time.Duration, acceptLocalAddresses bool) *AddressBook {
	addrFitlerFn := IsRoutable
	if acceptLocalAddresses {
		addrFitlerFn = IsRoutableWithLocal
	}

	return &AddressBook{
		addrs:        make([]*knownAddress, 0, 500),
		keyIndex:     make(map[string]int, 500),
		banDuration:  banDuration,
		addrFitlerFn: addrFitlerFn,
	}
}

func (a *AddressBook) UpsertPeerAddr(p *peer.Peer) {
	a.mu.Lock()
	defer a.mu.Unlock()

	pa := p.GetPeerAddr()
	key, ka := a.internalFind(pa)

	if ka != nil {
		ka.peer = p
	} else {
		a.internalAddAddr(key, &knownAddress{addr: pa, peer: p})
	}
}

func (a *AddressBook) AddAddrs(address []*wire.NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, addr := range address {
		if !a.addrFitlerFn(addr) {
			continue
		}

		key, ka := a.internalFind(addr)
		if ka == nil {
			a.internalAddAddr(key, &knownAddress{addr: addr})
		} else if addr.Timestamp.After(ka.addr.Timestamp) {
			ka.addr.Timestamp = addr.Timestamp
		}
	}
}

func (a *AddressBook) BanAddr(addr *wire.NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ka := a.internalFind(addr)
	if ka != nil {
		now := time.Now()
		ka.banTimestamp = &now
	}
}

func (a *AddressBook) GetRndUnusedAddr(tries uint) *wire.NetAddress {
	a.mu.Lock()
	defer a.mu.Unlock()

	alen := len(a.addrs)
	for i := uint(0); i < tries; i++ {
		ka := a.addrs[rand.Intn(alen)]
		if ka.peer == nil {
			if ka.isBanned(a.banDuration) {
				continue
			}
			return ka.addr
		}
	}

	return nil
}

func (a *AddressBook) internalFind(addr *wire.NetAddress) (string, *knownAddress) {
	key := addrKey(addr)
	addrIndex, ok := a.keyIndex[key]

	if ok {
		return key, a.addrs[addrIndex]
	}
	return key, nil
}

func (a *AddressBook) internalAddAddr(key string, addr *knownAddress) {
	newItemIndex := len(a.addrs)

	a.addrs = append(a.addrs, addr)
	a.keyIndex[key] = newItemIndex
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
