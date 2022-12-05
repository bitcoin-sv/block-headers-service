package cryptolib

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
)

// GetBlockSubsidyForHeight func
func GetBlockSubsidyForHeight(height uint64) uint64 {
	halvings := height / 210000
	// Force block reward to zero when right shift is undefined.
	if halvings >= 64 {
		return 0
	}

	subsidy := uint64(50 * 1e8)

	// Subsidy is cut in half every 210,000 blocks which will occur approximately every 4 years.
	subsidy >>= halvings
	return subsidy
}

// DifficultyFromBits returns the mining difficulty from the nBits field in the block header.
func DifficultyFromBits(bits string) (float64, error) {
	b, _ := hex.DecodeString(bits)
	ib := binary.BigEndian.Uint32(b)
	return targetToDifficulty(toCompactSize(ib))
}

func toCompactSize(bits uint32) *big.Int {
	t := big.NewInt(int64(bits % 0x01000000))
	t.Mul(t, big.NewInt(2).Exp(big.NewInt(2), big.NewInt(8*(int64(bits/0x01000000)-3)), nil))

	return t
}

func targetToDifficulty(target *big.Int) (float64, error) {
	a := float64(0xFFFF0000000000000000000000000000000000000000000000000000) // genesis difficulty
	b, err := strconv.ParseFloat(target.String(), 64)
	if err != nil {
		return 0.0, err
	}
	return a / b, nil
}

// GetLittleEndianBytes returns a byte array in little endian from an unsigned integer of 32 bytes.
func GetLittleEndianBytes(v uint32, l uint32) []byte {
	// TODO: is v hex encoded?
	buf := make([]byte, l)

	binary.LittleEndian.PutUint32(buf, v)

	return buf
}

// VarInt takes an unsiged integer and  returns a byte array in VarInt format.
// See http://learnmeabitcoin.com/glossary/varint
func VarInt(i uint64) []byte {
	b := make([]byte, 9)
	if i < 0xfd {
		b[0] = byte(i)
		return b[:1]
	}
	if i < 0x10000 {
		b[0] = 0xfd
		binary.LittleEndian.PutUint16(b[1:3], uint16(i))
		return b[:3]
	}
	if i < 0x100000000 {
		b[0] = 0xfe
		binary.LittleEndian.PutUint32(b[1:5], uint32(i))
		return b[:5]
	}
	b[0] = 0xff
	binary.LittleEndian.PutUint64(b[1:9], i)
	return b
}

// DecodeVarInt takes a byte array in VarInt format and returns the
// decoded unsiged integer value and it's size in bytes.
// See http://learnmeabitcoin.com/glossary/varint
func DecodeVarInt(b []byte) (result uint64, size int) {
	switch b[0] {
	case 0xff:
		result = binary.LittleEndian.Uint64(b[1:9])
		size = 9

	case 0xfe:
		result = uint64(binary.LittleEndian.Uint32(b[1:5]))
		size = 5

	case 0xfd:
		result = uint64(binary.LittleEndian.Uint16(b[1:3]))
		size = 3

	default:
		result = uint64(binary.LittleEndian.Uint16([]byte{b[0], 0x00}))
		size = 1
	}

	return
}

// Read the stream and return the varint value only consuming the correct number of bytes....
func DecodeVarIntFromReader(r *bufio.Reader) (uint64, []byte, error) {
	b := make([]byte, 1)
	if n, err := io.ReadFull(r, b); n != 1 || err != nil {
		return 0, nil, fmt.Errorf("Could not read varint type, got %d bytes and err: %v", n, err)
	}

	bytes := make([]byte, 0)
	bytes = append(bytes, b...)

	switch b[0] {
	case 0xff:
		bb := make([]byte, 8)
		if n, err := io.ReadFull(r, bb); n != 8 || err != nil {
			return 0, nil, fmt.Errorf("Could not read varint(8), got %d bytes and err: %v", n, err)
		}
		bytes = append(bytes, bb...)
		return binary.LittleEndian.Uint64(bb), bytes, nil

	case 0xfe:
		bb := make([]byte, 4)
		if n, err := io.ReadFull(r, bb); n != 4 || err != nil {
			return 0, nil, fmt.Errorf("Could not read varint(4), got %d bytes and err: %v", n, err)
		}
		bytes = append(bytes, bb...)
		return uint64(binary.LittleEndian.Uint32(bb)), bytes, nil

	case 0xfd:
		bb := make([]byte, 2)
		if n, err := io.ReadFull(r, bb); n != 2 || err != nil {
			return 0, nil, fmt.Errorf("Could not read varint(2), got %d bytes and err: %v", n, err)
		}
		bytes = append(bytes, bb...)
		return uint64(binary.LittleEndian.Uint16(bb)), bytes, nil

	default:
		return uint64(binary.LittleEndian.Uint16([]byte{b[0], 0x00})), bytes, nil
	}
}

// EncodeParts takes a slice of slices and returns a single slice with the appropriate OP_PUSH commands embedded.
func EncodeParts(parts [][]byte) ([]byte, error) {
	b := make([]byte, 0)

	for i, part := range parts {
		l := int64(len(part))

		if l <= 75 {
			b = append(b, byte(len(part)))
			b = append(b, part...)

		} else if l <= 0xFF {
			b = append(b, 0x4c) // OP_PUSHDATA1
			b = append(b, byte(len(part)))
			b = append(b, part...)

		} else if l <= 0xFFFF {
			b = append(b, 0x4d) // OP_PUSHDATA2
			lenBuf := make([]byte, 2)
			binary.LittleEndian.PutUint16(lenBuf, uint16(len(part)))
			b = append(b, lenBuf...)
			b = append(b, part...)

		} else if l <= 0xFFFFFFFF {
			b = append(b, 0x4e) // OP_PUSHDATA4
			lenBuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(lenBuf, uint32(len(part)))
			b = append(b, lenBuf...)
			b = append(b, part...)

		} else {
			return nil, fmt.Errorf("Part %d is too big", i)
		}
	}

	return b, nil
}

// DecodeStringParts calls DecodeParts.
func DecodeStringParts(s string) ([][]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return DecodeParts(b)
}

// DecodeParts returns an array of strings...
func DecodeParts(b []byte) ([][]byte, error) {
	var r [][]byte
	for len(b) > 0 {
		// Handle OP codes
		switch b[0] {
		case OpPUSHDATA1:
			if len(b) < 2 {
				return r, errors.New("Not enough data")
			}
			l := uint64(b[1])
			if len(b) < int(2+l) {
				return r, errors.New("Not enough data")
			}
			part := b[2 : 2+l]
			r = append(r, part)
			b = b[2+l:]

		case OpPUSHDATA2:
			if len(b) < 3 {
				return r, errors.New("Not enough data")
			}
			l := binary.LittleEndian.Uint16(b[1:])
			if len(b) < int(3+l) {
				return r, errors.New("Not enough data")
			}
			part := b[3 : 3+l]
			r = append(r, part)
			b = b[3+l:]

		case OpPUSHDATA4:
			if len(b) < 5 {
				return r, errors.New("Not enough data")
			}
			l := binary.LittleEndian.Uint32(b[1:])
			if len(b) < int(5+l) {
				return r, errors.New("Not enough data")
			}
			part := b[5 : 5+l]
			r = append(r, part)
			b = b[5+l:]

		default:
			if b[0] >= 0x01 && b[0] <= 0x4e {
				l := uint8(b[0])
				if len(b) < int(1+l) {
					return r, errors.New("Not enough data")
				}
				part := b[1 : l+1]
				r = append(r, part)
				b = b[1+l:]
			} else {
				r = append(r, []byte{b[0]})
				b = b[1:]
			}
		}
	}

	return r, nil
}
