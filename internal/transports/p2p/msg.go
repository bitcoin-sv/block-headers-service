package p2pexp

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/wire"
)

const (
	userAgentName     = "block-headers-service"
	userAgentVersion  = "12.0.0"
	userAgentComments = "experimental"
)

func generateOurVersionMsg() (*wire.MsgVersion, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(9223372036854775807))
	if err != nil {
		panic(err)
	}
	nonce := n.Uint64()

	ourNA := &wire.NetAddress{
		Services:  wire.SFspv,
		Timestamp: time.Now(),
	}

	theirNA := &wire.NetAddress{
		Services:  wire.SFspv,
		Timestamp: time.Now(),
	}

	// TODO: get newest block from DB
	blockNum := int32(0)

	// Version message.
	msg := wire.NewMsgVersion(ourNA, theirNA, nonce, blockNum)
	err = msg.AddUserAgent(userAgentName, userAgentVersion, userAgentComments)
	if err != nil {
		// TODO: log error
		return nil, err
	}

	msg.Services = wire.SFspv
	msg.ProtocolVersion = int32(protocolVersion)

	return msg, nil
}
