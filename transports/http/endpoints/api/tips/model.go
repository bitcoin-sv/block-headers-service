package tips

import (
	"math/big"

	"github.com/libsv/bitcoin-hc/domains"
)

// tipResponse defines a single block header.
type tipResponse struct {
	Hash             string   `json:"hash"`
	Version          int32    `json:"version"`
	PreviousBlock    string   `json:"prevBlockHash"`
	MerkleRoot       string   `json:"merkleRoot"`
	Timestamp        uint32   `json:"creationTimestamp"`
	DifficultyTarget uint32   `json:"difficultyTarget"`
	Nonce            uint32   `json:"nonce"`
	Work             *big.Int `json:"work" swaggertype:"string"`
}

// tipStateResponse is an extended version of the tipResponse
// that has more important information.
type tipStateResponse struct {
	Header    tipResponse `json:"header"`
	State     string      `json:"state"`
	ChainWork *big.Int    `json:"chainWork"  swaggertype:"string"`
	Height    int32       `json:"height"`
}

// newTipResponse maps a domain BlockHeader to a transport tipResponse.
func newTipResponse(header *domains.BlockHeader) tipResponse {
	return tipResponse{
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

// newTipStateResponse maps a domain BlockHeader to a transport tipStateResponse.
func newTipStateResponse(header *domains.BlockHeader) tipStateResponse {
	return tipStateResponse{
		Header:    newTipResponse(header),
		State:     header.State.String(),
		ChainWork: header.CumulatedWork,
		Height:    header.Height,
	}
}

// mapToTipStateResponse maps a slice of domain BlockHeader to a slice of transport tipStateResponse.
func mapToTipStateResponse(headers []*domains.BlockHeader) []tipStateResponse {
	blockHeaderStatesResponse := make([]tipStateResponse, 0)

	for _, header := range headers {
		blockHeaderStatesResponse = append(blockHeaderStatesResponse, newTipStateResponse(header))
	}

	return blockHeaderStatesResponse
}
