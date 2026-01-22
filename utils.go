package weir

import (
	"hash/maphash"
)

var seed = maphash.MakeSeed()

func hash(item string) uint64 {
	return maphash.String(seed, item)
}

func isPowerOfTwo(n int) bool {
	return n > 0 && n&(n-1) == 0
}
