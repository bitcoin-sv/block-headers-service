package domains

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
	CumulatedWork    *big.Int `json:"work"`
}

// BlockHeaderStateResponse is an extended version of the BlockHeaderResponse
// that has more important informations.
type BlockHeaderStateResponse struct {
	Header        BlockHeaderResponse `json:"header"`
	State         string              `json:"state"`
	ChainWork     uint64              `json:"chainWork"`
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
		CumulatedWork:    header.CumulatedWork,
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
		ChainWork:     header.Chainwork,
		Height:        header.Height,
		Confirmations: confirmations,
	}
}
