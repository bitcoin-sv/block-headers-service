package peer

import (
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/wire"
)

const (
	userAgentComments            = "experimental"
	initialProtocolVersion       = uint32(70013)
	maxProtocolVersion           = wire.FeeFilterVersion
	minAcceptableProtocolVersion = wire.MultipleAddressVersion
	pingInterval                 = 15 * time.Second
)
