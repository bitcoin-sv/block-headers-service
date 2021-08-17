package bc

import (
	"encoding/hex"

	"github.com/libsv/go-bt"
)

// A MerkleProof is a structure that proves inclusion of a
// Bitcoin transaction in a block.
type MerkleProof struct {
	Index      uint64   `json:"index"`
	TxOrID     string   `json:"txOrId"`
	Target     string   `json:"target"`
	Nodes      []string `json:"nodes"`
	TargetType string   `json:"targetType,omitempty"`
	ProofType  string   `json:"proofType,omitempty"`
	Composite  bool     `json:"composite,omitempty"`
}

// ToBytes converts the JSON Merkle Proof
// into byte encoding.
//
// Check the following encoding:
//
// flags: 			byte,
// index: 			varint,
// txLength: 		varint, //omitted if flag bit 0 == 0 as it's a fixed length transaction ID
// txOrId: 			byte[32 or variable length],
// target: 			byte[32 or 80], //determined by flag bits 1 and 2
// nodeCount: 	varint,
// nodes: 			node[]
func (mp *MerkleProof) ToBytes() ([]byte, error) {
	index := bt.VarInt(mp.Index)

	txOrID, err := hex.DecodeString(mp.TxOrID)
	if err != nil {
		return nil, err
	}
	txOrID = bt.ReverseBytes(txOrID)

	target, err := hex.DecodeString(mp.Target)
	if err != nil {
		return nil, err
	}
	target = bt.ReverseBytes(target)

	nodeCount := len(mp.Nodes)

	nodes := []byte{}

	for _, n := range mp.Nodes {
		if n == "*" {
			nodes = append(nodes, []byte{1}...)
			continue
		}

		nodes = append(nodes, []byte{0}...)
		nb, err := hex.DecodeString(n)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, bt.ReverseBytes(nb)...)

	}

	var flags uint8

	var txLength []byte
	if len(mp.TxOrID) > 64 { // tx bytes instead of txid
		// set bit at index 0
		flags |= (1 << 0)

		txLength = bt.VarInt(uint64(len(txOrID)))
	}

	if mp.TargetType == "header" {
		// set bit at index 1
		flags |= (1 << 1)
	} else if mp.TargetType == "merkleRoot" {
		// set bit at index 2
		flags |= (1 << 2)
	}

	// ignore proofType and compositeType for this version

	bytes := []byte{}
	bytes = append(bytes, flags)
	bytes = append(bytes, index...)
	bytes = append(bytes, txLength...)
	bytes = append(bytes, txOrID...)
	bytes = append(bytes, target...)
	bytes = append(bytes, byte(nodeCount))
	bytes = append(bytes, nodes...)

	return bytes, nil
}
