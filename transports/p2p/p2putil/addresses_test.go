package p2putil

import (
	"net"
	"testing"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/addrmgr"
)

func TestNewAddressFunc(t *testing.T) {
	// given
	na := &wire.NetAddress{
		Timestamp: time.Unix(0x495fab29, 0), // 2009-01-03 12:15:05 -0600 CST
		Services:  wire.SFNodeNetwork,
		IP:        net.ParseIP("127.0.0.1"),
		Port:      8333,
	}
	knownAddr := addrmgr.NewKnownAddress(na, na)
	getAddrFn := func() *addrmgr.KnownAddress {
		return knownAddr
	}
	outboundGroupCount := func(string) int {
		return 0
	}
	lookupFn := func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}
	expectedAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8333,
	}

	t.Run("Success", func(t *testing.T) {
		// when
		findNewAddrFunc := NewAddressFunc(getAddrFn, outboundGroupCount, lookupFn)
		addr, err := findNewAddrFunc()

		// then
		assert.NoError(t, err)
		assert.Equal(t, addr.Network(), expectedAddr.Network())
		assert.Equal(t, addr.String(), expectedAddr.String())
	})

	t.Run("Same group address", func(t *testing.T) {
		// given
		outboundGroupCount := func(string) int {
			return 1 // shouldn't find any address with this return value
		}

		// when
		findNewAddrFunc := NewAddressFunc(getAddrFn, outboundGroupCount, lookupFn)
		addr, err := findNewAddrFunc()

		// then
		assert.IsError(t, err, "no valid connect address")
		assert.Equal(t, addr, nil)
	})
}

func TestAddrStringToNetAddr(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		addr := "127.0.0.1:8333"
		network := "tcp"
		lookupFn := func(string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}

		// when
		netAddr, err := AddrStringToNetAddr(addr, lookupFn)

		// then
		assert.NoError(t, err)
		assert.Equal(t, netAddr.String(), addr)
		assert.Equal(t, netAddr.Network(), network)
	})

	t.Run("Invalid Address", func(t *testing.T) {
		// given
		addr := "invalid"
		lookupFn := func(string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}

		// when
		netAddr, err := AddrStringToNetAddr(addr, lookupFn)

		// then
		assert.IsError(t, err, "address invalid: missing port in address")
		assert.Equal(t, netAddr, nil)
	})

	t.Run("Tor Address", func(t *testing.T) {
		// given
		addr := "tor.onion:8333"
		lookupFn := func(string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}

		// when
		_, err := AddrStringToNetAddr(addr, lookupFn)

		// then
		assert.IsError(t, err, "tor option is not allowed")
	})

	t.Run("No Addresses", func(t *testing.T) {
		// given
		addr := "test_addr:8333"
		lookupFn := func(string) ([]net.IP, error) {
			return []net.IP{}, nil
		}

		// when
		_, err := AddrStringToNetAddr(addr, lookupFn)

		// then
		assert.IsError(t, err, "no addresses found for test_addr")
	})
}
