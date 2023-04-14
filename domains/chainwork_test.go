package domains

import (
	"fmt"
	"github.com/libsv/bitcoin-hc/internal/tests/assert"
	"math/big"
	"testing"
)

func TestChainWork(t *testing.T) {

	testCases := []struct {
		height            int
		bits              uint32
		expectedChainWork string
	}{
		{
			height:            0,
			bits:              0x1d00ffff,
			expectedChainWork: "4295032833",
		}, {
			height:            100_000,
			bits:              0x1b04864c,
			expectedChainWork: "62209952899966",
		}, {
			height:            200_000,
			bits:              0x1a05db8b,
			expectedChainWork: "12301577519373468",
		}, {
			height:            300_000,
			bits:              0x1900896c,
			expectedChainWork: "34364008516618225545",
		}, {
			height:            400_000,
			bits:              0x1806b99f,
			expectedChainWork: "702202025755488147582",
		}, {
			height:            500_000,
			bits:              0x1809b91a,
			expectedChainWork: "485687622324422197901",
		}, {
			height:            600_000,
			bits:              0x18089116,
			expectedChainWork: "551244161910380757574",
		}, {
			height:            700_000,
			bits:              0x181452d3,
			expectedChainWork: "232359535664858305416",
		}, {
			height:            292_320,
			bits:              0x1900db99,
			expectedChainWork: "21504630620890996935",
		},
	}
	for _, params := range testCases {
		name := fmt.Sprintf("should evaluate bits %d from block %d as chainwork %s",
			params.bits, params.height, params.expectedChainWork)
		t.Run(name, func(t *testing.T) {
			cw := CalculateWork(params.bits)

			assert.Equal(t, cw.BigInt().String(), params.expectedChainWork)
		})
	}
}

func TestCumulatedChainWork(t *testing.T) {
	testCases := []struct {
		height                     int
		bits                       uint32
		previousBlockChainWork     *big.Int
		expectedCumulatedChainWork string
	}{
		{
			height:                     0,
			bits:                       0x1d00ffff,
			previousBlockChainWork:     big.NewInt(0),
			expectedCumulatedChainWork: "4295032833",
		}, {
			height:                     100_000,
			bits:                       0x1b04864c,
			previousBlockChainWork:     bigIntFromHex("64492eaf00f2520"),
			expectedCumulatedChainWork: "451709610344319134",
		}, {
			height:                     200_000,
			bits:                       0x1a05db8b,
			previousBlockChainWork:     bigIntFromHex("1ac0479f335782cb80"),
			expectedCumulatedChainWork: "493482865315456673820",
		}, {
			height:                     300_000,
			bits:                       0x1900896c,
			previousBlockChainWork:     bigIntFromHex("5a795f5d6ede10bc6d60"),
			expectedCumulatedChainWork: "427286275270210023027945",
		}, {
			height:                     400_000,
			bits:                       0x1806b99f,
			previousBlockChainWork:     bigIntFromHex("1229fea679a4cdc26e7460"),
			expectedCumulatedChainWork: "21959681449681744428027102",
		}, {
			height:                     500_000,
			bits:                       0x1809b91a,
			previousBlockChainWork:     bigIntFromHex("7ae4707601d47bc6695487"),
			expectedCumulatedChainWork: "148568209777348817841919764",
		}, {
			height:                     600_000,
			bits:                       0x18089116,
			previousBlockChainWork:     bigIntFromHex("e8f2ea21f069a214067ed7"),
			expectedCumulatedChainWork: "281618473067294323593150749",
		}, {
			height:                     700_000,
			bits:                       0x181452d3,
			previousBlockChainWork:     bigIntFromHex("12f32fb33b26aa239be0fc3"),
			expectedCumulatedChainWork: "366545507884831374927242059",
		}, {
			height:                     292_320,
			bits:                       0x1900db99,
			previousBlockChainWork:     bigIntFromHex("2d66952994737e0a63e0"),
			expectedCumulatedChainWork: "214420312540482753037479",
		},
	}

	for _, params := range testCases {
		name := fmt.Sprintf("should calculate cumulated chainwor for block %d", params.height)
		t.Run(name, func(t *testing.T) {
			cw := CalculateWork(params.bits)
			ccw := CumulatedChainWorkOf(*params.previousBlockChainWork).Add(cw)

			assert.Equal(t, ccw.BigInt().String(), params.expectedCumulatedChainWork)
		})
	}
}

func bigIntFromHex(hex string) *big.Int {
	i := new(big.Int)
	i.SetString(hex, 16)
	return i
}
