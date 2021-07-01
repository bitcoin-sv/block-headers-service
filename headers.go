package headers

import (
	"context"
)

// BlockHeader defines a single block header, used in SPV validations.
type BlockHeader struct {
	Hash              string  `json:"hash" db:"hash"`
	Versionhex        string  `json:"versionHex" db:"versionHex"`
	Merkleroot        string  `json:"merkleroot" db:"merkleroot"`
	Bits              string  `json:"bits" db:"bits"`
	Chainwork         string  `json:"chainwork" db:"chainwork"`
	Previousblockhash string  `json:"previousblockhash" db:"previousblockhash"`
	Nextblockhash     string  `json:"nextblockhash" db:"nextblockhash"`
	Confirmations     int     `json:"confirmations" db:"confirmations"`
	Height            int     `json:"height" db:"height"`
	Mediantime        int     `json:"mediantime" db:"mediantime"`
	Difficulty        float64 `json:"difficulty" db:"difficulty"`
	Version           uint32  `json:"version" db:"version"`
	Time              uint32  `json:"time" db:"time"`
	Nonce             uint32  `json:"nonce" db:"nonce"`
}

// HeaderArgs are sued to retrieve a single block header.
type HeaderArgs struct {
	Blockhash string `param:"blockhash" db:"blockHash"`
}

// Height contains the current cached height as well as current blockchain height and
// a check for us being in sync or not.
type Height struct{
	Height int `json:"height"`
	NetworkNeight int `json:"networkHeight"`
	Synced bool `json:"synced"`
}

// BlockheaderService enforces validation of arguments and business rules.
type BlockheaderService interface {
	// Header will return a single header by block hash.
	Header(ctx context.Context, args HeaderArgs) (*BlockHeader, error)
	// Create will store a block header in the db.
	Create(ctx context.Context, req BlockHeader) error
	CreateBatch(ctx context.Context, req []*BlockHeader) error
	Height(ctx context.Context) (*Height, error)
}

// BlockheaderReader is used to get header information from a data store.
type BlockheaderReader interface {
	// Header will return a single header by block hash.
	Header(ctx context.Context, args HeaderArgs) (*BlockHeader, error)
	HeightReader
}

// HeightReader defines a contract for reading height data.
type HeightReader interface {
	// Height will return the current block height cached.
	Height(ctx context.Context) (int, error)
}

// BlockheaderWriter will add or modify block header data.
type BlockheaderWriter interface {
	// Create will add a blockheader to the data store.
	Create(ctx context.Context, req BlockHeader) error
	// CreateBatch will add a batch of records to the data store.
	CreateBatch(ctx context.Context, req []*BlockHeader) error
}
