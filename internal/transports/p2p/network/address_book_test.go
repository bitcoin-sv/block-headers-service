package network

import (
	"net"
	"testing"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/stretchr/testify/require"
)

func TestAddressBook_UpsertAddrs(t *testing.T) {
	t.Run("add new - accept local", func(t *testing.T) {
		// given
		sut := NewAdressbook(0, true)
		local := &wire.NetAddress{
			IP:        net.IPv4(127, 0, 0, 1),
			Port:      8333,
			Timestamp: time.Now(),
		}

		external := &wire.NetAddress{
			IP:        net.IPv4(18, 199, 12, 185),
			Port:      8333,
			Timestamp: time.Now(),
		}

		// when
		sut.UpsertAddrs([]*wire.NetAddress{local, external})

		// then
		require.Len(t, sut.addrs, 2)
	})

	t.Run("add new - do not accept local", func(t *testing.T) {
		// given
		sut := NewAdressbook(0, false)
		local := &wire.NetAddress{
			IP:        net.IPv4(127, 0, 0, 1),
			Port:      8333,
			Timestamp: time.Now(),
		}

		external := &wire.NetAddress{
			IP:        net.IPv4(18, 199, 12, 185),
			Port:      8333,
			Timestamp: time.Now(),
		}

		// when
		sut.UpsertAddrs([]*wire.NetAddress{local, external})

		// then
		require.Len(t, sut.addrs, 1)
	})

	t.Run("add existing", func(t *testing.T) {
		// given
		sut := NewAdressbook(0, true)
		addr := &wire.NetAddress{
			IP:        net.IPv4(18, 199, 12, 185),
			Port:      8333,
			Timestamp: time.Now().Add(-1 * time.Minute),
		}

		sut.UpsertAddrs([]*wire.NetAddress{addr})

		// when
		updated := &wire.NetAddress{
			IP:        addr.IP,
			Port:      addr.Port,
			Timestamp: time.Now(),
		}
		sut.UpsertAddrs([]*wire.NetAddress{updated})

		// then
		require.Len(t, sut.addrs, 1)
		require.Equal(t, updated.Timestamp, sut.addrs[0].addr.Timestamp)
	})
}

func TestAddressBook_BanAddr(t *testing.T) {
	t.Run("ban address", func(t *testing.T) {
		// given
		sut := NewAdressbook(1, false)

		addr := &wire.NetAddress{
			IP:        net.IPv4(18, 199, 12, 185),
			Port:      8333,
			Timestamp: time.Now(),
		}

		sut.UpsertAddrs([]*wire.NetAddress{addr})

		// when
		sut.BanAddr(addr)

		// then
		require.True(t, sut.addrs[0].isBanned(time.Hour))
	})
}

func TestAddressBook_GetRndUnusedAddr(t *testing.T) {
	t.Run("", func(t *testing.T) {
		// given
		sut := NewAdressbook(time.Hour, false)

		addr := &wire.NetAddress{
			IP:        net.IPv4(18, 199, 12, 186),
			Port:      8333,
			Timestamp: time.Now(),
		}

		addr2 := &wire.NetAddress{
			IP:        net.IPv4(18, 199, 12, 185),
			Port:      8333,
			Timestamp: time.Now(),
		}

		sut.UpsertAddrs([]*wire.NetAddress{addr, addr2})
		sut.BanAddr(addr2)

		// when
		r := sut.GetRndUnusedAddr(100)

		// then
		require.Equal(t, addr, r)
	})
}