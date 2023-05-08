package domains

import (
	"math/big"
	"time"
)

// HeaderEventType type of event.
type HeaderEventType string

const (
	// EventHeaderAdded event type for header added.
	EventHeaderAdded HeaderEventType = "ADD"
)

// HeaderEvent represents header event data.
type HeaderEvent struct {
	Operation HeaderEventType     `json:"operation"`
	Header    *HeaderEventDetails `json:"header"`
}

// HeaderEventDetails defines a header as a detailed part of an event.
type HeaderEventDetails struct {
	Height        int32       `json:"height"`
	Hash          string      `json:"hash"`
	Version       int32       `json:"version"`
	MerkleRoot    string      `json:"merkleRoot"`
	Timestamp     time.Time   `json:"creationTimestamp"`
	Nonce         uint32      `json:"nonce"`
	State         HeaderState `json:"state"`
	CumulatedWork *big.Int    `json:"work"`
	PreviousBlock string      `json:"prevBlockHash"`
}

// HeaderAdded makes event from block header.
func HeaderAdded(h *BlockHeader) *HeaderEvent {
	return &HeaderEvent{
		Operation: EventHeaderAdded,
		Header: &HeaderEventDetails{
			Height:        h.Height,
			Hash:          h.Hash.String(),
			Version:       h.Version,
			MerkleRoot:    h.MerkleRoot.String(),
			Timestamp:     h.Timestamp,
			Nonce:         h.Nonce,
			State:         h.State,
			CumulatedWork: h.CumulatedWork,
			PreviousBlock: h.PreviousBlock.String(),
		},
	}
}
