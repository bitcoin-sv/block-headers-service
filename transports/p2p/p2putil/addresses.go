package p2putil

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/addrmgr"
)

type (
	//revive:disable:exported
	GetAddressFn       func() *addrmgr.KnownAddress
	OutboundGroupCount func(string) int
	LookupFn           func(string) ([]net.IP, error)
	FindNewAddrFn      func() (net.Addr, error)
	//revive:enable:exported
)

// NewAddressFunc returns a function closure that can be used to get new working address
// Only setup a function to return new addresses to connect to when
// not running in connect-only mode.  The simulation network is always
// in connect-only mode since it is only intended to connect to
// specified peers and actively avoid advertising and connecting to
// discovered peers in order to prevent it from becoming a public test
// network.
func NewAddressFunc(getAddressFn GetAddressFn, outboundGroupCount OutboundGroupCount, lookupFn LookupFn) FindNewAddrFn {
	newAddressFunc := func() (net.Addr, error) {
		for tries := 0; tries < 100; tries++ {
			addr := getAddressFn()
			// Peer log
			// srvrconfigs.Log.Infof("[Server] newServer addr: %#v", addr)
			if addr == nil {
				break
			}

			// Address will not be invalid, local or unroutable
			// because addrmanager rejects those on addition.
			// Just check that we don't already have an address
			// in the same group so that we are not connecting
			// to the same network segment at the expense of
			// others.
			key := addrmgr.GroupKey(addr.NetAddress())
			if outboundGroupCount(key) != 0 {
				continue
			}

			// only allow recent nodes (10mins) after we failed 30
			// times
			if tries < 30 && time.Since(addr.LastAttempt()) < 10*time.Minute {
				continue
			}

			// allow nondefault ports after 50 failed tries.
			if tries < 50 && fmt.Sprintf("%d", addr.NetAddress().Port) !=
				config.ActiveNetParams.DefaultPort {
				continue
			}

			addrString := addrmgr.NetAddressKey(addr.NetAddress())
			return AddrStringToNetAddr(addrString, lookupFn)
		}

		return nil, errors.New("no valid connect address")
	}

	return newAddressFunc
}

// AddrStringToNetAddr takes an address in the form of 'host:port' and returns
// a net.Addr which maps to the original address with any host names resolved
// to IP addresses.  It also handles tor addresses properly by returning a
// net.Addr that encapsulates the address.
func AddrStringToNetAddr(addr string, lookupFn func(string) ([]net.IP, error)) (net.Addr, error) {
	host, strPort, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(strPort)
	if err != nil {
		return nil, err
	}

	// Skip if host is already an IP address.
	if ip := net.ParseIP(host); ip != nil {
		return &net.TCPAddr{
			IP:   ip,
			Port: port,
		}, nil
	}

	// Tor addresses cannot be resolved to an IP, so just return an onion
	// address instead.
	if strings.HasSuffix(host, ".onion") {
		return nil, errors.New("tor option is not allowed")
	}

	// Attempt to look up an IP address associated with the parsed host.
	ips, err := lookupFn(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses found for %s", host)
	}

	return &net.TCPAddr{
		IP:   ips[0],
		Port: port,
	}, nil
}
