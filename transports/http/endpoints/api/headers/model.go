package headers

import (
	"math/big"

	"github.com/libsv/bitcoin-hc/domains"
)

// blockHeaderResponse defines a single block header.
type blockHeaderResponse struct {
	Hash             string   `json:"hash"`
	Version          int32    `json:"version"`
	PreviousBlock    string   `json:"prevBlockHash"`
	MerkleRoot       string   `json:"merkleRoot"`
	Timestamp        uint32   `json:"creationTimestamp"`
	DifficultyTarget uint32   `json:"difficultyTarget"`
	Nonce            uint32   `json:"nonce"`
	Work             *big.Int `json:"work" swaggertype:"string"`
}

// blockHeaderStateResponse is an extended version of the blockHeaderResponse
// that has more important information.
type blockHeaderStateResponse struct {
	Header    blockHeaderResponse `json:"header"`
	State     string              `json:"state"`
	ChainWork *big.Int            `json:"chainWork"  swaggertype:"string"`
	Height    int32               `json:"height"`
}

// newBlockHeaderResponse maps a domain BlockHeader to a transport blockHeaderResponse.
func newBlockHeaderResponse(header *domains.BlockHeader) blockHeaderResponse {
	return blockHeaderResponse{
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

// mapToBlockHeadersResponses maps a slice of domain BlockHeader to a slice of transport blockHeaderResponse.
func mapToBlockHeadersResponses(headers []*domains.BlockHeader) []blockHeaderResponse {
	blockHeadersResponse := make([]blockHeaderResponse, 0)

	for _, header := range headers {
		blockHeadersResponse = append(blockHeadersResponse, newBlockHeaderResponse(header))
	}

	return blockHeadersResponse
}

// newBlockHeaderStateResponse maps a domain BlockHeader to a transport blockHeaderStateResponse.
func newBlockHeaderStateResponse(header *domains.BlockHeader) blockHeaderStateResponse {
	return blockHeaderStateResponse{
		Header:    newBlockHeaderResponse(header),
		State:     header.State.String(),
		ChainWork: header.CumulatedWork,
		Height:    header.Height,
	}
}
