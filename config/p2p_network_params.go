package config

import (
	"github.com/bitcoin-sv/pulse/internal/chaincfg"
)

// ActiveNetParams is a pointer to the parameters specific to the
// currently active bitcoin network.
var ActiveNetParams = updatedMainNetParams(mainNetParams)

// params is used to group parameters for various networks such as the main
// network and test networks.
type params struct {
	*chaincfg.Params
	rpcPort string
}

// mainNetParams contains parameters specific to the main network
// (wire.MainNet).  NOTE: The RPC port is intentionally different than the
// reference implementation because bsvd does not handle wallet requests.  The
// separate wallet process listens on the well-known port and forwards requests
// it does not handle on to bsvd.  This approach allows the wallet process
// to emulate the full reference implementation RPC API.
var mainNetParams = params{
	Params:  &chaincfg.MainNetParams,
	rpcPort: "8334",
}

var mainNetDNSSeeds = []chaincfg.DNSSeed{
	{Host: "seed.bitcoinsv.io", HasFiltering: true},
	{Host: "seed.gorillapool.io", HasFiltering: true},
	{Host: "seed.cascharia.com", HasFiltering: true},
	{Host: "seed.satoshisvision.network", HasFiltering: true},
}

func updatedMainNetParams(p params) *params {
	p.DNSSeeds = mainNetDNSSeeds
	return &p
}
