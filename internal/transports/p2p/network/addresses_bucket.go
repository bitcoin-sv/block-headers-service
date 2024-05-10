package network

type addrBucket struct {
	items  []*knownAddress
	lookup map[string]int
}

func newAddrBucket(initCapacity uint) *addrBucket {
	return &addrBucket{
		items:  make([]*knownAddress, 0, initCapacity),
		lookup: make(map[string]int, initCapacity),
	}
}

func (a *addrBucket) find(key string) *knownAddress {
	addrIndex, ok := a.lookup[key]

	if ok {
		return a.items[addrIndex]
	}
	return nil
}

func (a *addrBucket) add(key string, addr *knownAddress) {
	newItemIndex := len(a.items)

	a.items = append(a.items, addr)
	a.lookup[key] = newItemIndex
}

func (a *addrBucket) rm(key string) {
	addrIndex, ok := a.lookup[key]

	if ok {
		// substitute with last element
		a.items[addrIndex] = a.items[len(a.items)-1]
		// remove last element
		a.items = a.items[:len(a.items)-1]

		delete(a.lookup, key)
	}
}
