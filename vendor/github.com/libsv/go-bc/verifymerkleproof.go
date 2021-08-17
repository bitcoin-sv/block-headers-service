package bc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/libsv/go-bt"
)

const (
	txOrIDFlag      byte = 1 << iota // 1 << 0 which is 00000001
	targetTypeFlag1                  // 1 << 1 which is 00000010
	targetTypeFlag2                  // 1 << 2 which is 00000100
	proofTypeFlag                    // 1 << 3 which is 00001000
	compositeFlag                    // 1 << 4 which is 00010000

	targetTypeFlags = targetTypeFlag1 | targetTypeFlag2
)

// VerifyMerkleProof verifies a Merkle Proof in standard byte format.
func (spvc *SPVClient) VerifyMerkleProof(ctx context.Context, proof []byte) (valid, isLastInTree bool, err error) {

	mpb, err := parseBinaryMerkleProof(proof)
	if err != nil {
		return false, false, err
	}

	err = validateTxOrID(mpb.flags, mpb.txOrID)
	if err != nil {
		return false, false, err
	}

	txid, err := txidFromTxOrID(mpb.txOrID)
	if err != nil {
		return false, false, err
	}

	var merkleRoot string
	switch mpb.flags & targetTypeFlags {
	// if bits 1 and 2 of flags are NOT set, target should contain a block hash (32 bytes)
	case 0:
		// The `target` field contains a block hash
		blockHeader, err := spvc.bhc.BlockHeader(ctx, mpb.target)
		if err != nil {
			return false, false, err
		}

		merkleRoot = blockHeader.HashMerkleRoot

	// if bit 2 of flags is set, target should contain a merkle root (32 bytes)
	case 4:
		// the `target` field contains a merkle root
		merkleRoot = mpb.target

	// if bit 1 of flags is set, target should contain a block header (80 bytes)
	case 2:
		// The `target` field contains a block header
		var err error
		merkleRoot, err = ExtractMerkleRootFromBlockHeader(mpb.target)
		if err != nil {
			return false, false, err
		}

	default:
		return false, false, errors.New("invalid flags")
	}

	if mpb.flags&proofTypeFlag == 1 {
		return false, false, errors.New("only merkle branch supported in this version") // merkle tree proof type not supported
	}

	if mpb.flags&compositeFlag == 1 {
		return false, false, errors.New("only single proof supported in this version") // composite proof type not supported
	}

	if txid == "" {
		return false, false, errors.New("txid missing")
	}

	if merkleRoot == "" {
		return false, false, errors.New("merkleRoot missing")
	}

	return verifyProof(txid, merkleRoot, mpb.index, mpb.nodes)
}

// VerifyMerkleProofJSON verifies a Merkle Proof in standard JSON format.
func (spvc *SPVClient) VerifyMerkleProofJSON(ctx context.Context, proof *MerkleProof) (bool, bool, error) {

	txid, err := txidFromTxOrID(proof.TxOrID)
	if err != nil {
		return false, false, err
	}

	var merkleRoot string
	if proof.TargetType == "" || proof.TargetType == "hash" {
		// The `target` field contains a block hash

		if len(proof.Target) != 64 {
			return false, false, errors.New("invalid target field")
		}

		blockHeader, err := spvc.bhc.BlockHeader(ctx, proof.Target)
		if err != nil {
			return false, false, err
		}
		merkleRoot = blockHeader.HashMerkleRoot

	} else if proof.TargetType == "header" && len(proof.Target) == 160 {
		// The `target` field contains a block header
		var err error
		merkleRoot, err = ExtractMerkleRootFromBlockHeader(proof.Target)
		if err != nil {
			return false, false, err
		}

	} else if proof.TargetType == "merkleRoot" && len(proof.Target) == 64 {
		// the `target` field contains a merkle root
		merkleRoot = proof.Target

	} else {
		return false, false, errors.New("invalid TargetType or target field")
	}

	if proof.ProofType != "" && proof.ProofType != "branch" {
		return false, false, errors.New("only merkle branch supported in this version") // merkle tree proof type not supported
	}

	if proof.Composite { // OR if (proof.composite && proof.composite != false)
		return false, false, errors.New("only single proof supported in this version") // composite proof type not supported
	}

	if txid == "" {
		return false, false, errors.New("txid missing")
	}

	if merkleRoot == "" {
		return false, false, errors.New("merkleRoot missing")
	}

	return verifyProof(txid, merkleRoot, proof.Index, proof.Nodes)
}

func verifyProof(c, merkleRoot string, index uint64, nodes []string) (bool, bool, error) {
	isLastInTree := true

	for _, p := range nodes {
		// Check if the node is the left or the right child
		cIsLeft := index%2 == 0

		// Check for duplicate hash - this happens if the node (p) is
		// the last element of an uneven merkle tree layer
		if p == "*" {
			if !cIsLeft { // this shouldn't happen...
				return false, false, errors.New("invalid nodes")
			}
			p = c
		}

		// This check fails at least once if it's not the last element
		if cIsLeft && c != p {
			isLastInTree = false
		}

		var err error
		// Calculate the parent node
		if cIsLeft {
			// Concatenate left leaf (c) with right leaf (p)
			c, err = MerkleTreeParentStr(c, p)
			if err != nil {
				return false, false, err
			}
		} else {
			// Concatenate left leaf (p) with right leaf (c)
			c, err = MerkleTreeParentStr(p, c)
			if err != nil {
				return false, false, err
			}
		}

		// We need integer division here with remainder dropped.
		index = index / 2
	}

	// c is now the calculated merkle root
	return c == merkleRoot, isLastInTree, nil
}

func validateTxOrID(flags byte, txOrID string) error {
	// The `txOrId` field contains a full transaction
	if len(txOrID) > 64 && flags&txOrIDFlag == 0 {
		return errors.New("expecting txid but got tx")
	}

	// The `txOrId` field contains a transaction ID
	if len(txOrID) == 64 && flags&txOrIDFlag == 1 {
		return errors.New("expecting tx but got txid")
	}

	return nil
}

func txidFromTxOrID(txOrID string) (string, error) {

	// The `txOrId` field contains a transaction ID
	if len(txOrID) == 64 {
		return txOrID, nil
	}

	// The `txOrId` field contains a full transaction
	if len(txOrID) > 64 {
		tx, err := bt.NewTxFromString(txOrID)
		if err != nil {
			return "", err
		}

		return tx.GetTxID(), nil
	}

	return "", errors.New("invalid txOrId length - must be at least 64 chars (32 bytes)")
}

type merkleProofBinary struct {
	flags  byte
	index  uint64
	txOrID string
	target string
	nodes  []string
}

func parseBinaryMerkleProof(proof []byte) (*merkleProofBinary, error) {
	mpb := &merkleProofBinary{}

	var offset, size int

	// flags is first byte
	mpb.flags = proof[offset]
	offset++

	// index is the next varint after the 1st byte
	mpb.index, size = bt.DecodeVarInt(proof[offset:])
	offset += size

	var txLength uint64
	// if bit 1 of flags is NOT set, txOrId should contain txid (= 32 bytes)
	if mpb.flags&1 == 0 {
		txLength = 32
	}

	// if bit 1 of flags is set, txOrId should contain tx hex (> 32 bytes)
	if mpb.flags&1 == 1 {
		if txLength <= 32 {
			return nil, errors.New("invalid tx length (should be greater than 32 bytes)")
		}

		// txLength is the next varint after the 1st byte + index size
		txLength, size = bt.DecodeVarInt(proof[offset:])
		offset += size
	}

	// txOrID is the next txLength bytes after 1st byte + index size (+ txLength size)
	mpb.txOrID = hex.EncodeToString(bt.ReverseBytes(proof[offset : offset+int(txLength)]))
	offset += int(txLength)

	switch mpb.flags & targetTypeFlags {
	// if bits 1 and 2 of flags are NOT set, target should contain a block hash (32 bytes)
	// if bit 2 of flags is set, target should contain a merkle root (32 bytes)
	case 0, 4:
		mpb.target = hex.EncodeToString(bt.ReverseBytes(proof[offset : offset+32]))
		offset += 32

	// if bit 1 of flags is set, target should contain a block header (80 bytes)
	case 2:
		mpb.target = hex.EncodeToString(bt.ReverseBytes(proof[offset : offset+80]))
		offset += 80

	default:
		return nil, errors.New("invalid flags")
	}

	nodeCount, size := bt.DecodeVarInt(proof[offset:])
	offset += size

	for i := 0; i < int(nodeCount); i++ {
		t := proof[offset]
		offset++

		var n string
		switch t {
		case 0:
			n = hex.EncodeToString(bt.ReverseBytes(proof[offset : offset+32]))
			offset += 32
		case 1:
			n = "*"

		default:
			return nil, fmt.Errorf("invalid value in node type at index: %q", i)
		}

		mpb.nodes = append(mpb.nodes, n)
	}

	return mpb, nil
}
