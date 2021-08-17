package bc

import (
	"encoding/hex"

	"github.com/libsv/go-bt"
	"github.com/libsv/go-bt/crypto"
)

// MerkleTreeParentStr returns the Merkle Tree parent of two Merkle
// Tree children using hex strings instead of just bytes.
func MerkleTreeParentStr(leftNode, rightNode string) (string, error) {
	l, err := hex.DecodeString(leftNode)
	if err != nil {
		return "", err
	}
	r, err := hex.DecodeString(rightNode)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(MerkleTreeParent(l, r)), nil
}

// MerkleTreeParent returns the Merkle Tree parent of two Merkle
// Tree children.
func MerkleTreeParent(leftNode, rightNode []byte) []byte {
	// swap endianness before concatenating
	l := bt.ReverseBytes(leftNode)
	r := bt.ReverseBytes(rightNode)

	// concatenate leaves
	concat := append(l, r...)

	// hash the concatenation
	hash := crypto.Sha256d(concat)

	// swap endianness at the end and convert to hex string
	return bt.ReverseBytes(hash)
}
