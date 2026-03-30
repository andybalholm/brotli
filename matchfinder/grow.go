package matchfinder

import "slices"

func growClearUint32(dst []uint32, n int) []uint32 {
	if n <= 0 {
		return dst
	}

	oldLen := len(dst)
	dst = slices.Grow(dst, n)
	dst = dst[:oldLen+n]
	clear(dst[oldLen:])
	return dst
}
