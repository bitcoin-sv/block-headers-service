package tips

import (
	"math/big"

	"github.com/bitcoin-sv/block-headers-service/domains"
)

// TipResponse defines a single block header.
type TipResponse struct {
	Hash             string   `json:"hash"`
	Version          int32    `json:"version"`
	PreviousBlock    string   `json:"prevBlockHash"`
	MerkleRoot       string   `json:"merkleRoot"`
	Timestamp        uint32   `json:"creationTimestamp"`
	DifficultyTarget uint32   `json:"difficultyTarget"`
	Nonce            uint32   `json:"nonce"`
	Work             *big.Int `json:"work" swaggertype:"string"`
}

// TipStateResponse is an extended version of the TipResponse
// that has more important information.
type TipStateResponse struct {
	Header    TipResponse `json:"header"`
	State     string      `json:"state"`
	ChainWork *big.Int    `json:"chainWork"  swaggertype:"string"`
	Height    int32       `json:"height"`
}

// newTipResponse maps a domain BlockHeader to a transport TipResponse.
func newTipResponse(header *domains.BlockHeader) TipResponse {
	return TipResponse{
		Hash:             header.Hash.String(),
		Version:          header.Version,
		PreviousBlock:    header.PreviousBlock.String(),
		MerkleRoot:       header.MerkleRoot.String(),
		Timestamp:        uint32(header.Timestamp.Unix()),
		DifficultyTarget: header.Bits,
		Nonce:            header.Nonce,
		Work:             header.Chainwork,
	}
}

// newTipStateResponse maps a domain BlockHeader to a transport TipStateResponse.
func newTipStateResponse(header *domains.BlockHeader) TipStateResponse {
	return TipStateResponse{
		Header:    newTipResponse(header),
		State:     header.State.String(),
		ChainWork: header.CumulatedWork,
		Height:    header.Height,
	}
}

// mapToTipStateResponse maps a slice of domain BlockHeader to a slice of transport TipStateResponse.
func mapToTipStateResponse(headers []*domains.BlockHeader) []TipStateResponse {
	blockHeaderStatesResponse := make([]TipStateResponse, 0)

	for _, header := range headers {
		blockHeaderStatesResponse = append(blockHeaderStatesResponse, newTipStateResponse(header))
	}

	return blockHeaderStatesResponse
}
