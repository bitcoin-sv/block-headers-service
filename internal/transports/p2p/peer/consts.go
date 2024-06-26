package peer

import (
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
)

const (
	userAgentComments            = "experimental"
	initialProtocolVersion       = uint32(70013)
	maxProtocolVersion           = wire.FeeFilterVersion
	minAcceptableProtocolVersion = wire.MultipleAddressVersion
	pingInterval                 = 2 * time.Minute
	writeMsgChannelBufferSize    = 10
	ourServices                  = wire.SFspv
)

var zeroHash chainhash.Hash
