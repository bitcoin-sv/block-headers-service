// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addrmgr_test

import (
	"errors"
	"net"

	"testing"

	testlog "github.com/libsv/bitcoin-hc/internal/tests/log"
	"github.com/libsv/bitcoin-hc/internal/wire"
	"github.com/libsv/bitcoin-hc/transports/p2p/addrmgr"
)

// naTest is used to describe a test to be performed against the NetAddressKey
// method.
type naTest struct {
	in   wire.NetAddress
	want string
}

// naTests houses all of the tests to be performed against the NetAddressKey
// method.
var naTests = make([]naTest, 0)

// addNaTests.
func addNaTests() {
	// IPv4
	// Localhost
	addNaTest("127.0.0.1", 8333, "127.0.0.1:8333")
	addNaTest("127.0.0.1", 8334, "127.0.0.1:8334")

	// Class A
	addNaTest("1.0.0.1", 8333, "1.0.0.1:8333")
	addNaTest("2.2.2.2", 8334, "2.2.2.2:8334")
	addNaTest("27.253.252.251", 8335, "27.253.252.251:8335")
	addNaTest("123.3.2.1", 8336, "123.3.2.1:8336")

	// Private Class A
	addNaTest("10.0.0.1", 8333, "10.0.0.1:8333")
	addNaTest("10.1.1.1", 8334, "10.1.1.1:8334")
	addNaTest("10.2.2.2", 8335, "10.2.2.2:8335")
	addNaTest("10.10.10.10", 8336, "10.10.10.10:8336")

	// Class B
	addNaTest("128.0.0.1", 8333, "128.0.0.1:8333")
	addNaTest("129.1.1.1", 8334, "129.1.1.1:8334")
	addNaTest("180.2.2.2", 8335, "180.2.2.2:8335")
	addNaTest("191.10.10.10", 8336, "191.10.10.10:8336")

	// Private Class B
	addNaTest("172.16.0.1", 8333, "172.16.0.1:8333")
	addNaTest("172.16.1.1", 8334, "172.16.1.1:8334")
	addNaTest("172.16.2.2", 8335, "172.16.2.2:8335")
	addNaTest("172.16.172.172", 8336, "172.16.172.172:8336")

	// Class C
	addNaTest("193.0.0.1", 8333, "193.0.0.1:8333")
	addNaTest("200.1.1.1", 8334, "200.1.1.1:8334")
	addNaTest("205.2.2.2", 8335, "205.2.2.2:8335")
	addNaTest("223.10.10.10", 8336, "223.10.10.10:8336")

	// Private Class C
	addNaTest("192.168.0.1", 8333, "192.168.0.1:8333")
	addNaTest("192.168.1.1", 8334, "192.168.1.1:8334")
	addNaTest("192.168.2.2", 8335, "192.168.2.2:8335")
	addNaTest("192.168.192.192", 8336, "192.168.192.192:8336")

	// IPv6
	// Localhost
	addNaTest("::1", 8333, "[::1]:8333")
	addNaTest("fe80::1", 8334, "[fe80::1]:8334")

	// Link-local
	addNaTest("fe80::1:1", 8333, "[fe80::1:1]:8333")
	addNaTest("fe91::2:2", 8334, "[fe91::2:2]:8334")
	addNaTest("fea2::3:3", 8335, "[fea2::3:3]:8335")
	addNaTest("feb3::4:4", 8336, "[feb3::4:4]:8336")

	// Site-local
	addNaTest("fec0::1:1", 8333, "[fec0::1:1]:8333")
	addNaTest("fed1::2:2", 8334, "[fed1::2:2]:8334")
	addNaTest("fee2::3:3", 8335, "[fee2::3:3]:8335")
	addNaTest("fef3::4:4", 8336, "[fef3::4:4]:8336")
}

func addNaTest(ip string, port uint16, want string) {
	nip := net.ParseIP(ip)
	na := *wire.NewNetAddressIPPort(nip, port, wire.SFNodeNetwork)
	test := naTest{na, want}
	naTests = append(naTests, test)
}

func lookupFunc(host string) ([]net.IP, error) {
	return nil, errors.New("not implemented")
}

func TestStartStop(t *testing.T) {
	lf := testlog.NewTestLoggerFactory()
	n := addrmgr.New(lookupFunc, lf)
	n.Start()
	n.Stop()
}

func TestAddLocalAddress(t *testing.T) {
	var tests = []struct {
		address  wire.NetAddress
		priority addrmgr.AddressPriority
		valid    bool
	}{
		{
			wire.NetAddress{IP: net.ParseIP("192.168.0.100")},
			addrmgr.InterfacePrio,
			false,
		},
		{
			wire.NetAddress{IP: net.ParseIP("204.124.1.1")},
			addrmgr.InterfacePrio,
			true,
		},
		{
			wire.NetAddress{IP: net.ParseIP("204.124.1.1")},
			addrmgr.BoundPrio,
			true,
		},
		{
			wire.NetAddress{IP: net.ParseIP("::1")},
			addrmgr.InterfacePrio,
			false,
		},
		{
			wire.NetAddress{IP: net.ParseIP("fe80::1")},
			addrmgr.InterfacePrio,
			false,
		},
		{
			wire.NetAddress{IP: net.ParseIP("2620:100::1")},
			addrmgr.InterfacePrio,
			true,
		},
	}
	lf := testlog.NewTestLoggerFactory()
	amgr := addrmgr.New(nil, lf)
	for x, test := range tests {
		result := amgr.AddLocalAddress(&test.address, test.priority)
		if result == nil && !test.valid {
			t.Errorf("TestAddLocalAddress test #%d failed: %s should have "+
				"been accepted", x, test.address.IP)
			continue
		}
		if result != nil && test.valid {
			t.Errorf("TestAddLocalAddress test #%d failed: %s should not have "+
				"been accepted", x, test.address.IP)
			continue
		}
	}
}

func TestGetBestLocalAddress(t *testing.T) {
	lf := testlog.NewTestLoggerFactory()
	localAddrs := []wire.NetAddress{
		{IP: net.ParseIP("192.168.0.100")},
		{IP: net.ParseIP("::1")},
		{IP: net.ParseIP("fe80::1")},
		{IP: net.ParseIP("2001:470::1")},
	}

	var tests = []struct {
		remoteAddr wire.NetAddress
		want0      wire.NetAddress
		want1      wire.NetAddress
		want2      wire.NetAddress
		want3      wire.NetAddress
	}{
		{
			// Remote connection from public IPv4
			wire.NetAddress{IP: net.ParseIP("204.124.8.1")},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.ParseIP("204.124.8.100")},
			wire.NetAddress{IP: net.ParseIP("fd87:d87e:eb43:25::1")},
		},
		{
			// Remote connection from private IPv4
			wire.NetAddress{IP: net.ParseIP("172.16.0.254")},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.IPv4zero},
		},
		{
			// Remote connection from public IPv6
			wire.NetAddress{IP: net.ParseIP("2602:100:abcd::102")},
			wire.NetAddress{IP: net.IPv6zero},
			wire.NetAddress{IP: net.ParseIP("2001:470::1")},
			wire.NetAddress{IP: net.ParseIP("2001:470::1")},
			wire.NetAddress{IP: net.ParseIP("2001:470::1")},
		},
		/* XXX
		{
			// Remote connection from Tor
			wire.NetAddress{IP: net.ParseIP("fd87:d87e:eb43::100")},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.ParseIP("204.124.8.100")},
			wire.NetAddress{IP: net.ParseIP("fd87:d87e:eb43:25::1")},
		},
		*/
	}

	amgr := addrmgr.New(nil, lf)

	// Test against default when there's no address
	for x, test := range tests {
		got := amgr.GetBestLocalAddress(&test.remoteAddr)
		if !test.want0.IP.Equal(got.IP) {
			t.Errorf("TestGetBestLocalAddress test1 #%d failed for remote address %s: want %s got %s",
				x, test.remoteAddr.IP, test.want1.IP, got.IP)
			continue
		}
	}

	for _, localAddr := range localAddrs {
		amgr.AddLocalAddress(&localAddr, addrmgr.InterfacePrio)
	}

	// Test against want1
	for x, test := range tests {
		got := amgr.GetBestLocalAddress(&test.remoteAddr)
		if !test.want1.IP.Equal(got.IP) {
			t.Errorf("TestGetBestLocalAddress test1 #%d failed for remote address %s: want %s got %s",
				x, test.remoteAddr.IP, test.want1.IP, got.IP)
			continue
		}
	}

	// Add a public IP to the list of local addresses.
	localAddr := wire.NetAddress{IP: net.ParseIP("204.124.8.100")}
	amgr.AddLocalAddress(&localAddr, addrmgr.InterfacePrio)

	// Test against want2
	for x, test := range tests {
		got := amgr.GetBestLocalAddress(&test.remoteAddr)
		if !test.want2.IP.Equal(got.IP) {
			t.Errorf("TestGetBestLocalAddress test2 #%d failed for remote address %s: want %s got %s",
				x, test.remoteAddr.IP, test.want2.IP, got.IP)
			continue
		}
	}
	/*
		// Add a Tor generated IP address
		localAddr = wire.NetAddress{IP: net.ParseIP("fd87:d87e:eb43:25::1")}
		amgr.AddLocalAddress(&localAddr, addrmgr.ManualPrio)

		// Test against want3
		for x, test := range tests {
			got := amgr.GetBestLocalAddress(&test.remoteAddr)
			if !test.want3.IP.Equal(got.IP) {
				t.Errorf("TestGetBestLocalAddress test3 #%d failed for remote address %s: want %s got %s",
					x, test.remoteAddr.IP, test.want3.IP, got.IP)
				continue
			}
		}
	*/
}

func TestNetAddressKey(t *testing.T) {
	addNaTests()

	t.Logf("Running %d tests", len(naTests))
	for i, test := range naTests {
		key := addrmgr.NetAddressKey(&test.in)
		if key != test.want {
			t.Errorf("NetAddressKey #%d\n got: %s want: %s", i, key, test.want)
			continue
		}
	}

}
