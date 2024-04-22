package config

import (
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
)

type NetworkType string

const (
	MainNet       NetworkType = "mainnet"
	RegTestNet    NetworkType = "regtest"
	TestNet       NetworkType = "testnet"
	SimulationNet NetworkType = "simnet"
)

func GetNetParams(network NetworkType) *chaincfg.Params {
	switch network {
	case MainNet:
		return &chaincfg.MainNetParams
	case RegTestNet:
		return &chaincfg.RegressionNetParams
	case TestNet:
		return &chaincfg.TestNet3Params
	case SimulationNet:
		return &chaincfg.SimNetParams
	default:
		return &chaincfg.MainNetParams
	}
}

// ActiveNetParams is a pointer to the parameters specific to the
// currently active bitcoin network.
// TODO: remove this after switching to new p2p server
var ActiveNetParams = &chaincfg.MainNetParams
