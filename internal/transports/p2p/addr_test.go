package p2pexp

import (
	"net"
	"testing"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
)

func TestParseAddress(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		// given
		cases := []struct {
			addr     string
			port     string
			expected *net.TCPAddr
		}{
			{
				addr: "127.0.0.1",
				port: "8333",
				expected: &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8333,
				},
			},
			{
				addr: "::ffff:192.0.2.1",
				port: "8333",
				expected: &net.TCPAddr{
					IP:   net.ParseIP("::ffff:192.0.2.1"),
					Port: 8333,
				},
			},
			{
				addr: "2001:db8::68",
				port: "8333",
				expected: &net.TCPAddr{
					IP:   net.ParseIP("2001:db8::68"),
					Port: 8333,
				},
			},
		}

		for _, c := range cases {
			// when
			addr, err := parseAddress(c.addr, c.port)

			// then
			assert.Equal(t, addr, c.expected)
			assert.NoError(t, err)
		}
	})

	t.Run("test errors", func(t *testing.T) {
		// given
		cases := []struct {
			addr          string
			port          string
			expectedError string
		}{
			{
				addr:          "127.0.0.1",
				port:          "wrong_port",
				expectedError: "could not parse port: strconv.Atoi: parsing \"wrong_port\": invalid syntax",
			},
			{
				addr:          "wrong_ip",
				port:          "8333",
				expectedError: "could not parse peer IP",
			},
		}

		for _, c := range cases {
			// when
			addr, err := parseAddress(c.addr, c.port)

			// then
			assert.Equal(t, addr, nil)
			assert.IsError(t, err, c.expectedError)
		}
	})
}
