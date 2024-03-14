package p2putil

import (
	"fmt"
	"net"
	"runtime"
	"strings"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/rs/zerolog"
)

// InitListeners initializes the configured net listeners and adds any bound
// addresses to the address manager. Returns the listeners and a NAT interface,
// which is non-nil if UPnP is in use.
func InitListeners(log *zerolog.Logger) ([]net.Listener, error) {
	listenAddrs := []string{
		net.JoinHostPort("", config.ActiveNetParams.DefaultPort),
	}

	// Listen for TCP connections at the configured addresses
	netAddrs, err := parseListeners(listenAddrs)
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := net.Listen(addr.Network(), addr.String())
		if err != nil {
			log.Warn().Msgf("Can't listen on %s: %v", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)

	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			na, err := net.ResolveTCPAddr("tcp4", addr)
			if err != nil {
				return nil, err
			}
			netAddrs = append(netAddrs, na)
			na, err = net.ResolveTCPAddr("tcp6", addr)
			if err != nil {
				return nil, err
			}
			netAddrs = append(netAddrs, na)
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}

		if ip.To4() == nil {
			na, err := net.ResolveTCPAddr("tcp6", addr)
			if err != nil {
				return nil, err
			}
			netAddrs = append(netAddrs, na)
		} else {
			na, err := net.ResolveTCPAddr("tcp4", addr)
			if err != nil {
				return nil, err
			}
			netAddrs = append(netAddrs, na)
		}
	}

	return netAddrs, nil
}
