package config

import (
	"fmt"
	"net"
	"strings"
)

// BsvdLookup resolves the IP of the given host using the correct DNS lookup
// function depending on the configuration options.  For example, addresses will
// be resolved using tor when the --proxy flag was specified unless --noonion
// was also specified in which case the normal system DNS resolver will be used.
//
// Any attempt to resolve a tor address (.onion) will return an error since they
// are not intended to be resolved outside of the tor proxy.
func (c *P2PConfig) BsvdLookup(host string) ([]net.IP, error) {
	if strings.HasSuffix(host, ".onion") {
		return nil, fmt.Errorf("attempt to resolve tor address %s", host)
	}

	return Lookup(host)
}

// BsvdDial connects to the address on the named network using the appropriate
// dial function depending on the address and configuration options.  For
// example, .onion addresses will be dialed using the onion specific proxy if
// one was specified, but will otherwise use the normal dial function (which
// could itself use a proxy or not).
func (c *P2PConfig) BsvdDial(addr net.Addr) (net.Conn, error) {
	if strings.Contains(addr.String(), ".onion:") {
		return nil, fmt.Errorf("no possibility to use onion addresses: %v", addr)
	}

	return Dial(addr.Network(), addr.String(), c.DefaultConnectTimeout)
}
