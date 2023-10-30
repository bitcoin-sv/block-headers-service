package domains

import "github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"

// BlockLocator contain slice of header hashes.
type BlockLocator []*chainhash.Hash
