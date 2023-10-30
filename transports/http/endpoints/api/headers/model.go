package headers

import (
	"math/big"

	"github.com/bitcoin-sv/pulse/domains"
)

// BlockHeaderResponse defines a single block header.
type BlockHeaderResponse struct {
	Hash             string   `json:"hash"`
	Version          int32    `json:"version"`
	PreviousBlock    string   `json:"prevBlockHash"`
	MerkleRoot       string   `json:"merkleRoot"`
	Timestamp        uint32   `json:"creationTimestamp"`
	DifficultyTarget uint32   `json:"difficultyTarget"`
	Nonce            uint32   `json:"nonce"`
	Work             *big.Int `json:"work" swaggertype:"string"`
}

// BlockHeaderStateResponse is an extended version of the BlockHeaderResponse
// that has more important information.
type BlockHeaderStateResponse struct {
	Header    BlockHeaderResponse `json:"header"`
	State     string              `json:"state"`
	ChainWork *big.Int            `json:"chainWork"  swaggertype:"string"`
	Height    int32               `json:"height"`
}

// newBlockHeaderResponse maps a domain BlockHeader to a transport BlockHeaderResponse.
func newBlockHeaderResponse(header *domains.BlockHeader) BlockHeaderResponse {
	return BlockHeaderResponse{
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

// mapToBlockHeadersResponses maps a slice of domain BlockHeader to a slice of transport BlockHeaderResponse.
func mapToBlockHeadersResponses(headers []*domains.BlockHeader) []BlockHeaderResponse {
	blockHeadersResponse := make([]BlockHeaderResponse, 0)

	for _, header := range headers {
		blockHeadersResponse = append(blockHeadersResponse, newBlockHeaderResponse(header))
	}

	return blockHeadersResponse
}

// newBlockHeaderStateResponse maps a domain BlockHeader to a transport BlockHeaderStateResponse.
func newBlockHeaderStateResponse(header *domains.BlockHeader) BlockHeaderStateResponse {
	return BlockHeaderStateResponse{
		Header:    newBlockHeaderResponse(header),
		State:     header.State.String(),
		ChainWork: header.CumulatedWork,
		Height:    header.Height,
	}
}
