package config

import (
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
)

// NetworkType is a string that represents the network type.
type NetworkType string

const (
	// MainNet represents the main network.
	MainNet NetworkType = "mainnet"
	// RegTestNet represents the regression test network.
	RegTestNet NetworkType = "regtest"
	// TestNet represents the test network.
	TestNet NetworkType = "testnet"
	// SimulationNet represents the simulation network.
	SimulationNet NetworkType = "simnet"
)

// GetNetParams returns the network parameters for current network.
func (c *P2PConfig) GetNetParams() *chaincfg.Params {
	switch c.ChainNetType {
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
// TODO: remove this after switching to new p2p server.
var ActiveNetParams = &chaincfg.MainNetParams
