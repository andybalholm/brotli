package brotli

import "math"

const HUGE_VAL = math.MaxFloat64

func assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}

func log2(n float64) float64 {
	return math.Log2(n)
}
