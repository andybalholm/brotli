package brotli

/* NOLINT(build/header_guard) */
/* Copyright 2018 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Rolling hash for long distance long string matches. Stores one position
   per bucket, bucket key is computed over a long region. */
var kRollingHashMul32HROLLING_FAST uint32 = 69069

var kInvalidPosHROLLING_FAST uint32 = 0xffffffff

/* This hasher uses a longer forward length, but returning a higher value here
   will hurt compression by the main hasher when combined with a composite
   hasher. The hasher tests for forward itself instead. */
func HashTypeLengthHROLLING_FAST() uint {
	return 4
}

func StoreLookaheadHROLLING_FAST() uint {
	return 4
}

/* Computes a code from a single byte. A lookup table of 256 values could be
   used, but simply adding 1 works about as good. */
func HashByteHROLLING_FAST(byte byte) uint32 {
	return uint32(byte) + 1
}

func HashRollingFunctionInitialHROLLING_FAST(state uint32, add byte, factor uint32) uint32 {
	return uint32(factor*state + HashByteHROLLING_FAST(add))
}

func HashRollingFunctionHROLLING_FAST(state uint32, add byte, rem byte, factor uint32, factor_remove uint32) uint32 {
	return uint32(factor*state + HashByteHROLLING_FAST(add) - factor_remove*HashByteHROLLING_FAST(rem))
}

type HROLLING_FAST struct {
	state         uint32
	table         []uint32
	next_ix       uint
	chunk_len     uint32
	factor        uint32
	factor_remove uint32
}

func SelfHROLLING_FAST(handle HasherHandle) *HROLLING_FAST {
	return handle.extra.(*HROLLING_FAST)
}

func InitializeHROLLING_FAST(handle HasherHandle, params *BrotliEncoderParams) {
	handle.extra = new(HROLLING_FAST)
	var self *HROLLING_FAST = SelfHROLLING_FAST(handle)
	var i uint
	self.state = 0
	self.next_ix = 0

	self.factor = kRollingHashMul32HROLLING_FAST

	/* Compute the factor of the oldest byte to remove: factor**steps modulo
	   0xffffffff (the multiplications rely on 32-bit overflow) */
	self.factor_remove = 1

	for i = 0; i < 32; i += 4 {
		self.factor_remove *= self.factor
	}

	self.table = make([]uint32, 16777216)
	for i = 0; i < 16777216; i++ {
		self.table[i] = kInvalidPosHROLLING_FAST
	}
}

func PrepareHROLLING_FAST(handle HasherHandle, one_shot bool, input_size uint, data []byte) {
	var self *HROLLING_FAST = SelfHROLLING_FAST(handle)
	var i uint

	/* Too small size, cannot use this hasher. */
	if input_size < 32 {
		return
	}
	self.state = 0
	for i = 0; i < 32; i += 4 {
		self.state = HashRollingFunctionInitialHROLLING_FAST(self.state, data[i], self.factor)
	}
}

func StoreHROLLING_FAST(handle HasherHandle, data []byte, mask uint, ix uint) {
}

func StoreRangeHROLLING_FAST(handle HasherHandle, data []byte, mask uint, ix_start uint, ix_end uint) {
}

func StitchToPreviousBlockHROLLING_FAST(handle HasherHandle, num_bytes uint, position uint, ringbuffer []byte, ring_buffer_mask uint) {
	var self *HROLLING_FAST = SelfHROLLING_FAST(handle)
	var position_masked uint
	/* In this case we must re-initialize the hasher from scratch from the
	   current position. */

	var available uint = num_bytes
	if position&(4-1) != 0 {
		var diff uint = 4 - (position & (4 - 1))
		if diff > available {
			available = 0
		} else {
			available = available - diff
		}
		position += diff
	}

	position_masked = position & ring_buffer_mask

	/* wrapping around ringbuffer not handled. */
	if available > ring_buffer_mask-position_masked {
		available = ring_buffer_mask - position_masked
	}

	PrepareHROLLING_FAST(handle, false, available, ringbuffer[position&ring_buffer_mask:])
	self.next_ix = position
}

func PrepareDistanceCacheHROLLING_FAST(handle HasherHandle, distance_cache *int) {
}

func FindLongestMatchHROLLING_FAST(handle HasherHandle, dictionary *BrotliEncoderDictionary, data []byte, ring_buffer_mask uint, distance_cache *int, cur_ix uint, max_length uint, max_backward uint, gap uint, max_distance uint, out *HasherSearchResult) {
	var self *HROLLING_FAST = SelfHROLLING_FAST(handle)
	var cur_ix_masked uint = cur_ix & ring_buffer_mask
	var pos uint = self.next_ix

	if cur_ix&(4-1) != 0 {
		return
	}

	/* Not enough lookahead */
	if max_length < 32 {
		return
	}

	for pos = self.next_ix; pos <= cur_ix; pos += 4 {
		var code uint32 = self.state & ((16777216 * 64) - 1)
		var rem byte = data[pos&ring_buffer_mask]
		var add byte = data[(pos+32)&ring_buffer_mask]
		var found_ix uint = uint(kInvalidPosHROLLING_FAST)

		self.state = HashRollingFunctionHROLLING_FAST(self.state, add, rem, self.factor, self.factor_remove)

		if code < 16777216 {
			found_ix = uint(self.table[code])
			self.table[code] = uint32(pos)
			if pos == cur_ix && uint32(found_ix) != kInvalidPosHROLLING_FAST {
				/* The cast to 32-bit makes backward distances up to 4GB work even
				   if cur_ix is above 4GB, despite using 32-bit values in the table. */
				var backward uint = uint(uint32(cur_ix - found_ix))
				if backward <= max_backward {
					var found_ix_masked uint = found_ix & ring_buffer_mask
					var len uint = FindMatchLengthWithLimit(data[found_ix_masked:], data[cur_ix_masked:], max_length)
					if len >= 4 && len > out.len {
						var score uint = BackwardReferenceScore(uint(len), backward)
						if score > out.score {
							out.len = uint(len)
							out.distance = backward
							out.score = score
							out.len_code_delta = 0
						}
					}
				}
			}
		}
	}

	self.next_ix = cur_ix + 4
}
