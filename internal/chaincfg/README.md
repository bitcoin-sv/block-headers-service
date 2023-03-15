chaincfg
========

[![Build Status](https://travis-ci.org/metasv/bsvd.png?branch=master)](https://travis-ci.org/metasv/bsvd)
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/metasv/mvcd/chaincfg)

Package chaincfg defines chain configuration parameters for the three standard
Bitcoin networks and provides the ability for callers to define their own custom
Bitcoin networks.

Although this package was primarily written for bsvd, it has intentionally been
designed so it can be used as a standalone package for any projects needing to
use parameters for the standard Bitcoin Cash networks or for projects needing to
define their own network.

## Sample Use

```Go
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gignative-solutions/ba-p2p-headers/p2putil"
	"github.com/gignative-solutions/ba-p2p-headers/chaincfg"
)

var testnet = flag.Bool("testnet", false, "operate on the testnet Bitcoin network")

// By default (without -testnet), use mainnet.
var chainParams = &chaincfg.MainNetParams

func main() {
	flag.Parse()

	// Modify active network parameters if operating on testnet.
	if *testnet {
		chainParams = &chaincfg.TestNet3Params
	}

	// later...

	// Create and print new payment address, specific to the active network.
	pubKeyHash := make([]byte, 20)
	addr, err := p2putil.NewAddressPubKeyHash(pubKeyHash, chainParams)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(addr)
}
```

## Installation and Updating

```bash
$ go get -u github.com/metasv/mvcd/chaincfg
```

## License

Package chaincfg is licensed under the [copyfree](http://copyfree.org) ISC
License.
