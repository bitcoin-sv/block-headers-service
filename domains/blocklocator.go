package domains

import "github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"

// BlockLocator contain slice of header hashes.
type BlockLocator []*chainhash.Hash
