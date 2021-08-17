package zmq

import (
	"log"

	"github.com/ordishs/go-bitcoin"
	"github.com/pkg/errors"

	"github.com/libsv/bitcoin-hc/config"
)

// Setup will setup the bitcoin node and zmq connections.
func Setup(c *config.BitcoinNode) (*bitcoin.ZMQ, *bitcoin.Bitcoind) {
	zmqConnection := bitcoin.NewZMQWithRaw("localhost", 28332)

	bn, err := bitcoin.New(c.Host, c.Port, c.Username, c.Password, c.UseSSL)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "failed to setup bitcoin node"))
	}
	return zmqConnection, bn

}
