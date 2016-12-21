package grumpy

import (
	"math/big"
	"strings"
)

const (
	// Here we calculate the number of bits in a uint and use that to
	// create a typeless constant _maxuint which is the largest value that
	// can be held by a uint. We then use that to create the constants
	// MaxInt and MinInt below.
	// Because these constants are typeless, they can be used wherever
	// a numeric value is needed, without a conversion like int64().
	// A typeless number remains untyped when shifted, even if the shift
	// count is typed.
	// Start with the two's complement of 0 as a uint which is 0xffff...ff.
	// This is the number we are after, but it currently has a type (uint).
	// Dividing it by 0xff gives us 0x0101...01 for the length of a uint.
	// Taking that mod 15 is effectively counting the ones - one for each
	// byte in a uint, so we have either 4 or 8.
	// We multiply by 8 and shift, and now we have 2^32 or 2^64 as a
	// typeless constant number.
	// We subtract 1 from that to get maxuint.
	_maxuint = 1<<(^uint(0)/0xff%15*8) - 1

	// MaxInt is the largest (most positive) number that can be stored as an int.
	MaxInt = _maxuint >> 1
	// MinInt is the smallest (most negative) number that can be stored as an int.
	// The absolute value of MinInt is Maxint+1, thus it can be tricky to deal with.
	MinInt = -(_maxuint + 1) >> 1
)

var (
	maxIntBig = big.NewInt(MaxInt)
	minIntBig = big.NewInt(MinInt)
)

func numParseInteger(z *big.Int, s string, base int) (*big.Int, bool) {
	s = strings.TrimSpace(s)
	if len(s) > 2 && s[0] == '0' {
		switch s[1] {
		case 'b', 'B':
			if base == 0 || base == 2 {
				base = 2
				s = s[2:]
			}
		case 'o', 'O':
			if base == 0 || base == 8 {
				base = 8
				s = s[2:]
			}
		case 'x', 'X':
			if base == 0 || base == 16 {
				base = 16
				s = s[2:]
			}
		default:
			base = 8
		}
	}
	if base == 0 {
		base = 10
	}
	return z.SetString(s, base)
}

func numInIntRange(i *big.Int) bool {
	return i.Cmp(minIntBig) >= 0 && i.Cmp(maxIntBig) <= 0
}
