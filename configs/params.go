// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package configs

import (
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/wire"
)

// ActiveNetParams is a pointer to the parameters specific to the
// currently active bitcoin network.
// var ActiveNetParams = &mainNetParams
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
	{"seed.bitcoinsv.io", true},
	{"seed.cascharia.com", true},
	{"seed.satoshisvision.network", true},
	// {"btccash-seeder.bitcoinunlimited.info", false},
}

var bsvCheckpoints = []chaincfg.Checkpoint{
	{11111, newHashFromStr("0000000069e244f73d78e8fd29ba2fd2ed618bd6fa2ee92559f542fdb26e7c1d")},
	{33333, newHashFromStr("000000002dd5588a74784eaa7ab0507a18ad16a236e7b1ce69f00d7ddfb5d0a6")},
	{74000, newHashFromStr("0000000000573993a3c9e41ce34471c079dcf5f52a0e824a81e7f953b8661a20")},
	{105000, newHashFromStr("00000000000291ce28027faea320c8d2b054b2e0fe44a773f3eefb151d6bdc97")},
	{134444, newHashFromStr("00000000000005b12ffd4cd315cd34ffd4a594f430ac814c91184a0d42d2b0fe")},
	{168000, newHashFromStr("000000000000099e61ea72015e79632f216fe6cb33d7899acb35b75c8303b763")},
	{193000, newHashFromStr("000000000000059f452a5f7340de6682a977387c17010ff6e6c3bd83ca8b1317")},
	{210000, newHashFromStr("000000000000048b95347e83192f69cf0366076336c639f9b7228e9ba171342e")},
	{216116, newHashFromStr("00000000000001b4f4b433e81ee46494af945cf96014816a4e2370f11b23df4e")},
	{225430, newHashFromStr("00000000000001c108384350f74090433e7fcf79a606b8e797f065b130575932")},
	{250000, newHashFromStr("000000000000003887df1f29024b06fc2200b55f8af8f35453d7be294df2d214")},
	{279000, newHashFromStr("0000000000000001ae8c72a0b0c301f67e3afca10e819efa9041e458e9bd7e40")},
	{295000, newHashFromStr("00000000000000004d9b4ef50f0f9d686fd69db2e03af35a100370c64632a983")},
	// UAHF fork block.
	{478558, newHashFromStr("0000000000000000011865af4122fe3b144e2cbeea86142e8ff2fb4107352d43")},
	// Nov, 13 DAA activation block.
	{504031, newHashFromStr("0000000000000000011ebf65b60d0a3de80b8175be709d653b4c1a1beeb6ab9c")},
	// Monolith activation.
	{530359, newHashFromStr("0000000000000000011ada8bd08f46074f44a8f155396f43e38acf9501c49103")},
}

// newHashFromStr converts the passed big-endian hex string into a
// chainhash.Hash.  It only differs from the one available in chainhash in that
// it panics on an error since it will only (and must only) be called with
// hard-coded, and therefore known good, hashes.
func newHashFromStr(hexStr string) *chainhash.Hash {
	hash, err := chainhash.NewHashFromStr(hexStr)
	if err != nil {
		// Ordinarily I don't like panics in library code since it
		// can take applications down without them having a chance to
		// recover which is extremely annoying, however an exception is
		// being made in this case because the only way this can panic
		// is if there is an error in the hard-coded hashes.  Thus it
		// will only ever potentially panic on init and therefore is
		// 100% predictable.
		panic(err)
	}
	return hash
}

func updatedMainNetParams(p params) *params {
	p.DNSSeeds = mainNetDNSSeeds
	// p.Checkpoints = bsvCheckpoints
	return &p
}

// regressionNetParams contains parameters specific to the regression test
// network (wire.TestNet).  NOTE: The RPC port is intentionally different
// than the reference implementation - see the mainNetParams comment for
// details.
var regressionNetParams = params{
	Params:  &chaincfg.RegressionNetParams,
	rpcPort: "18334",
}

// testNet3Params contains parameters specific to the test network (version 3)
// (wire.TestNet3).  NOTE: The RPC port is intentionally different than the
// reference implementation - see the mainNetParams comment for details.
var testNet3Params = params{
	Params:  &chaincfg.TestNet3Params,
	rpcPort: "18334",
}

// simNetParams contains parameters specific to the simulation test network
// (wire.SimNet).
var simNetParams = params{
	Params:  &chaincfg.SimNetParams,
	rpcPort: "18556",
}

// netName returns the name used when referring to a bitcoin network.  At the
// time of writing, bsvd currently places blocks for testnet version 3 in the
// data and log directory "testnet", which does not match the Name field of the
// chaincfg parameters.  This function can be used to override this directory
// name as "testnet" when the passed active network matches wire.TestNet3.
//
// A proper upgrade to move the data and log directories for this network to
// "testnet3" is planned for the future, at which point this function can be
// removed and the network parameter's name used instead.
func netName(chainParams *params) string {
	switch chainParams.Net {
	case wire.TestNet3:
		return "testnet"
	default:
		return chainParams.Name
	}
}
