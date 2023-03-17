package domains

import "github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"

// BlockLocator contain slice of header hashes.
type BlockLocator []*chainhash.Hash
