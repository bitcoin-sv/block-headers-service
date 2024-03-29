package domains

import "math/big"

// calcWork calculate chainwork for header based on given bits.
func calcWork(bits uint32) *big.Int {
	// Return a work value of zero if the passed difficulty bits represent
	// a negative number. Note this should not happen in practice with valid
	// blocks, but an invalid block could trigger it.
	difficultyNum := CompactToBig(bits)
	if difficultyNum.Sign() <= 0 {
		return big.NewInt(0)
	}
	// (1 << 256) / (difficultyNum + 1)
	denominator := new(big.Int).Add(difficultyNum, big.NewInt(1))
	// oneLsh256 is 1 shifted left 256 bits.
	var oneLsh256 = new(big.Int).Lsh(big.NewInt(1), 256)
	return new(big.Int).Div(oneLsh256, denominator)
}

// ChainWork representation of the blockchain work for given block.
type ChainWork big.Int

// BigInt return big.Int representation of chain work.
func (cw *ChainWork) BigInt() *big.Int {
	if cw != nil {
		v := big.Int(*cw)
		return &v
	}
	return big.NewInt(0)
}

// CalculateWork calculate ChainWork based on provided bits.
func CalculateWork(bits uint32) *ChainWork {
	cw := *calcWork(bits)
	return ChainWorkOf(cw)
}

// ChainWorkOf represents big.Int as ChainWork.
func ChainWorkOf(v big.Int) *ChainWork {
	cw := ChainWork(v)
	return &cw
}

// CumulatedChainWork representation of the cumulated blockchain work.
type CumulatedChainWork big.Int

// CumulatedChainWorkOf represents big.Int as CumulatedChainWork.
func CumulatedChainWorkOf(v big.Int) *CumulatedChainWork {
	ccv := CumulatedChainWork(v)
	return &ccv
}

// BigInt return big.Int representation of CumulatedChainWork.
func (ccw *CumulatedChainWork) BigInt() *big.Int {
	if ccw != nil {
		v := big.Int(*ccw)
		return &v
	}
	return big.NewInt(0)
}

// Add returns a CumulatedChainWork as a sum of previous CumulatedChainWork and provided ChainWork.
func (ccw *CumulatedChainWork) Add(cw *ChainWork) CumulatedChainWork {
	sum := big.NewInt(0)
	sum = sum.Add(ccw.BigInt(), cw.BigInt())
	return CumulatedChainWork(*sum)
}

// CompactToBig  takes a compact representation of a 256-bit number used in Bitcoin,
// converts it to a big.Int, and returns the resulting big.Int value.
func CompactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}
