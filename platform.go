package brotli

/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Bit reading helpers */
/* Copyright 2016 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Macros for compiler / platform specific features and build options.

   Build options are:
    * BROTLI_BUILD_32_BIT disables 64-bit optimizations
    * BROTLI_BUILD_64_BIT forces to use 64-bit optimizations
    * BROTLI_BUILD_BIG_ENDIAN forces to use big-endian optimizations
    * BROTLI_BUILD_ENDIAN_NEUTRAL disables endian-aware optimizations
    * BROTLI_BUILD_LITTLE_ENDIAN forces to use little-endian optimizations
    * BROTLI_BUILD_PORTABLE disables dangerous optimizations, like unaligned
      read and overlapping memcpy; this reduces decompression speed by 5%
    * BROTLI_BUILD_NO_RBIT disables "rbit" optimization for ARM CPUs
    * BROTLI_DEBUG dumps file name and line number when decoder detects stream
      or memory error
    * BROTLI_ENABLE_LOG enables asserts and dumps various state information
*/
type brotli_reg_t uint64

func brotli_min_double(a float64, b float64) float64 {
	if a < b {
		return a
	} else {
		return b
	}
}

func brotli_max_double(a float64, b float64) float64 {
	if a > b {
		return a
	} else {
		return b
	}
}

func brotli_min_float(a float32, b float32) float32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func brotli_max_float(a float32, b float32) float32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func brotli_min_int(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func brotli_max_int(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func brotli_min_size_t(a uint, b uint) uint {
	if a < b {
		return a
	} else {
		return b
	}
}

func brotli_max_size_t(a uint, b uint) uint {
	if a > b {
		return a
	} else {
		return b
	}
}

func brotli_min_uint32_t(a uint32, b uint32) uint32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func brotli_max_uint32_t(a uint32, b uint32) uint32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func brotli_min_uint8_t(a byte, b byte) byte {
	if a < b {
		return a
	} else {
		return b
	}
}

func brotli_max_uint8_t(a byte, b byte) byte {
	if a > b {
		return a
	} else {
		return b
	}
}
