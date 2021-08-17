package bc

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/libsv/go-bt"
)

/*
Field 													Purpose 									 														Size (Bytes)
----------------------------------------------------------------------------------------------------
Version 							Block version number 																									4
hashPrevBlock 				256-bit hash of the previous block header 	 													32
hashMerkleRoot 				256-bit hash based on all of the transactions in the block 	 					32
Time 									Current block timestamp as seconds since 1970-01-01T00:00 UTC 				4
Bits 									Current target in compact format 	 																		4
Nonce 								32-bit number (starts at 0) 	 																				4
*/

// A BlockHeader in the Bitcoin blockchain.
type BlockHeader struct {
	Version        uint32
	Time           uint32
	Nonce          uint32
	HashPrevBlock  string
	HashMerkleRoot string
	Bits           string
}

// TODO: make fields private and make getters and setters

// String returns the Block Header encoded as hex string.
func (bh *BlockHeader) String() (string, error) {
	bb, err := bh.Bytes()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bb), nil
}

// Bytes will decode a bitcoin block header struct
// into a byte slice.
// See https://en.bitcoin.it/wiki/Block_hashing_algorithm
func (bh *BlockHeader) Bytes() ([]byte, error) {
	bytes := []byte{}

	v := make([]byte, 4)
	binary.LittleEndian.PutUint32(v, bh.Version)
	bytes = append(bytes, v...)

	p, err := hex.DecodeString(bh.HashPrevBlock)
	if err != nil {
		return nil, err
	}
	p = bt.ReverseBytes(p)
	bytes = append(bytes, p...)

	m, err := hex.DecodeString(bh.HashMerkleRoot)
	if err != nil {
		return nil, err
	}
	m = bt.ReverseBytes(m)
	bytes = append(bytes, m...)

	t := make([]byte, 4)
	binary.LittleEndian.PutUint32(t, bh.Time)
	bytes = append(bytes, t...)

	b, err := hex.DecodeString(bh.Bits)
	if err != nil {
		return nil, err
	}
	b = bt.ReverseBytes(b)
	bytes = append(bytes, b...)

	n := make([]byte, 4)
	binary.LittleEndian.PutUint32(t, bh.Nonce)
	bytes = append(bytes, n...)

	return bytes, nil
}

// EncodeBlockHeaderStr will encode a block header hash
// into the bitcoin block header structure.
// See https://en.bitcoin.it/wiki/Block_hashing_algorithm
func EncodeBlockHeaderStr(headerStr string) (*BlockHeader, error) {
	if len(headerStr) != 160 {
		return nil, errors.New("block header should be 80 bytes long")
	}

	headerBytes, err := hex.DecodeString(headerStr)
	if err != nil {
		return nil, err
	}

	return EncodeBlockHeader(headerBytes)
}

// EncodeBlockHeader will encode a block header byte slice
// into the bitcoin block header structure.
// See https://en.bitcoin.it/wiki/Block_hashing_algorithm
func EncodeBlockHeader(headerBytes []byte) (*BlockHeader, error) {
	if len(headerBytes) != 80 {
		return nil, errors.New("block header should be 80 bytes long")
	}

	return &BlockHeader{
		Version:        binary.LittleEndian.Uint32(headerBytes[:4]),
		HashPrevBlock:  hex.EncodeToString(bt.ReverseBytes(headerBytes[4:36])),
		HashMerkleRoot: hex.EncodeToString(bt.ReverseBytes(headerBytes[36:68])),
		Time:           binary.LittleEndian.Uint32(headerBytes[68:72]),
		Bits:           hex.EncodeToString(bt.ReverseBytes(headerBytes[72:76])),
		Nonce:          binary.LittleEndian.Uint32(headerBytes[76:]),
	}, nil
}

// ExtractMerkleRootFromBlockHeader will take an 80 byte Bitcoin block
// header hex string and return the Merkle Root from it.
func ExtractMerkleRootFromBlockHeader(header string) (string, error) {
	bh, err := EncodeBlockHeaderStr(header)
	if err != nil {
		return "", err
	}
	return bh.HashMerkleRoot, nil
}
