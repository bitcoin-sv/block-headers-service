package http

import (
	"math/big"

	"github.com/libsv/bitcoin-hc/domains"
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
// that has more important informations.
type BlockHeaderStateResponse struct {
	Header        BlockHeaderResponse `json:"header"`
	State         string              `json:"state"`
	ChainWork     *big.Int            `json:"chainWork"  swaggertype:"string"`
	Height        int32               `json:"height"`
	Confirmations int                 `json:"confirmations"`
}

// MapToBlockHeaderReponse maps a domain BlockHeader to a transport BlockHeaderResponse.
func MapToBlockHeaderReponse(header domains.BlockHeader) BlockHeaderResponse {
	return BlockHeaderResponse{
		Hash:             header.Hash.String(),
		Version:          header.Version,
		PreviousBlock:    header.PreviousBlock.String(),
		MerkleRoot:       header.MerkleRoot.String(),
		Timestamp:        uint32(header.Timestamp.Unix()),
		DifficultyTarget: header.DifficultyTarget,
		Nonce:            header.Nonce,
		Work:             header.Chainwork,
	}
}

// MapToBlockHeadersReponse maps a slice of domain BlockHeader to a slice of transport BlockHeaderResponse.
func MapToBlockHeadersReponse(headers []*domains.BlockHeader) []BlockHeaderResponse {
	blockHeadersResponse := make([]BlockHeaderResponse, 0)

	for _, header := range headers {
		blockHeadersResponse = append(blockHeadersResponse, MapToBlockHeaderReponse(*header))
	}

	return blockHeadersResponse
}

// MapToBlockHeaderStateReponse maps a domain BlockHeader to a transport BlockHeaderStateResponse.
func MapToBlockHeaderStateReponse(header domains.BlockHeader, confirmations int) BlockHeaderStateResponse {
	return BlockHeaderStateResponse{
		Header:        MapToBlockHeaderReponse(header),
		State:         header.State.String(),
		ChainWork:     header.CumulatedWork,
		Height:        header.Height,
		Confirmations: confirmations,
	}
}
