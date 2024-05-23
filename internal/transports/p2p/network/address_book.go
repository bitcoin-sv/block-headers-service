package network

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/wire"
)

type addressBucketType string

const (
	freeBucket   addressBucketType = "free"
	usedBucket   addressBucketType = "used"
	bannedBucket addressBucketType = "banned"
)

// AddressBook represents a collection of known network addresses.
type AddressBook struct {
	banDuration  time.Duration
	addrs        map[addressBucketType]*addrBucket
	mu           sync.Mutex
	addrFitlerFn func(*wire.NetAddress) bool
}

// NewAddressBook creates and initializes a new AddressBook instance.
func NewAddressBook(banDuration time.Duration, acceptLocalAddresses bool) *AddressBook {
	// Set the address filter function based on whether local addresses are accepted
	addrFilterFn := wire.IsRoutable
	if acceptLocalAddresses {
		addrFilterFn = wire.IsRoutableWithLocal
	}

	const addressesInitCapacity = 500
	const usedAddressesInitCapacity = 8

	knownAddress := make(map[addressBucketType]*addrBucket, 3)
	knownAddress[freeBucket] = newAddrBucket(addressesInitCapacity)
	knownAddress[bannedBucket] = newAddrBucket(addressesInitCapacity)
	knownAddress[usedBucket] = newAddrBucket(usedAddressesInitCapacity)

	return &AddressBook{
		banDuration:  banDuration,
		addrFitlerFn: addrFilterFn,
		addrs:        knownAddress,
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

		key, ka, _ := a.findAddr(addr)
		// If the address is not found, add it to the AddressBook.
		if ka == nil {
			a.addrs[freeBucket].add(key, &knownAddress{addr: addr})
		} else if addr.Timestamp.After(ka.addr.Timestamp) {
			// Otherwise, update the timestamp if the new one is newer.
			ka.addr.Timestamp = addr.Timestamp
		}
	}
}

// MarkUsedAddr updates or adds a peer's address.
func (a *AddressBook) MarkUsedAddr(pa *wire.NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()

	key := addrKey(pa)
	// remove from free if exists
	a.addrs[freeBucket].rm(key)
	// add to used
	a.addrs[usedBucket].add(key, &knownAddress{addr: pa})

}

// BanAddr bans a network address. Ignores address if doesn't exist in the AddressBook.
func (a *AddressBook) BanAddr(addr *wire.NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if key, ka, bucket := a.findAddr(addr); ka != nil {
		switch bucket {
		case freeBucket:
			a.ban(bucket, key, ka)
		case usedBucket:
			a.ban(bucket, key, ka)
		case bannedBucket:
		default:
			// Do nothing
		}
	}
}

// GetRandFreeAddr returns a randomly chosen unused network address.
func (a *AddressBook) GetRandFreeAddr() *wire.NetAddress {
	a.mu.Lock()
	defer a.mu.Unlock()

	freeAddres := a.addrs[freeBucket].items
	fLen := len(freeAddres)
	if fLen == 0 {
		return nil
	}

	// #nosec G404
	randIndx := rand.Intn(fLen)
	return freeAddres[randIndx].addr
}

func (a *AddressBook) findAddr(addr *wire.NetAddress) (key string, ka *knownAddress, bucket addressBucketType) {
	key = addrKey(addr)

	// search in free addresses
	if ka = a.addrs[freeBucket].find(key); ka != nil {
		bucket = freeBucket
		return
	}

	// search in used
	if ka = a.addrs[usedBucket].find(key); ka != nil {
		bucket = usedBucket
		return
	}

	// search in banned
	if ka = a.addrs[bannedBucket].find(key); ka != nil {
		bucket = bannedBucket
		return
	}

	return key, nil, ""
}

func (a *AddressBook) ban(bucket addressBucketType, key string, ka *knownAddress) {
	a.addrs[bucket].rm(key)
	a.addrs[bannedBucket].add(key, ka)
	go a.unban(key, ka)
}

func (a *AddressBook) unban(key string, ka *knownAddress) {
	time.Sleep(a.banDuration)

	a.mu.Lock()
	defer a.mu.Unlock()

	a.addrs[bannedBucket].rm(key)
	a.addrs[freeBucket].add(key, ka)
}

func addrKey(addr *wire.NetAddress) string {
	return fmt.Sprintf("%s:%d", addr.IP, addr.Port)
}

type knownAddress struct {
	addr *wire.NetAddress
}
