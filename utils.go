package weir

import (
	"hash/maphash"
	"math"
)

var seed = maphash.MakeSeed()

func hash(item string) uint64 {
	return maphash.String(seed, item)
}

func toNearestPowerOfTwo(n float64) int {
	if n <= 0 {
		return 1
	}

	power := math.Round(math.Log2(n))

	return int(math.Pow(2, power))
}

func isPowerOfTwo(n int) bool {
	return n > 0 && n&(n-1) == 0
}
