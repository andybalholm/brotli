package brotli

/* NOLINT(build/header_guard) */
/* Copyright 2018 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Composite hasher: This hasher allows to combine two other hashers, HASHER_A
   and HASHER_B. */
func HashTypeLengthH65() uint {
	var a uint = HashTypeLengthH6()
	var b uint = HashTypeLengthHROLLING()
	if a > b {
		return a
	} else {
		return b
	}
}

func StoreLookaheadH65() uint {
	var a uint = StoreLookaheadH6()
	var b uint = StoreLookaheadHROLLING()
	if a > b {
		return a
	} else {
		return b
	}
}

type H65 struct {
	HasherCommon
	ha     HasherHandle
	hb     HasherHandle
	params *BrotliEncoderParams
}

func SelfH65(handle HasherHandle) *H65 {
	return handle.(*H65)
}

func InitializeH65(handle HasherHandle, params *BrotliEncoderParams) {
	var self *H65 = SelfH65(handle)
	self.ha = nil
	self.hb = nil
	self.params = params
}

/* TODO: Initialize of the hashers is defered to Prepare (and params
   remembered here) because we don't get the one_shot and input_size params
   here that are needed to know the memory size of them. Instead provide
   those params to all hashers InitializeH65 */
func PrepareH65(handle HasherHandle, one_shot bool, input_size uint, data []byte) {
	var self *H65 = SelfH65(handle)
	if self.ha == nil {
		var common_a *HasherCommon
		var common_b *HasherCommon

		self.ha = new(H6)
		common_a = self.ha.Common()
		common_a.params = self.params.hasher
		common_a.is_prepared_ = false
		common_a.dict_num_lookups = 0
		common_a.dict_num_matches = 0
		InitializeH6(self.ha, self.params)

		self.hb = new(HROLLING)
		common_b = self.hb.Common()
		common_b.params = self.params.hasher
		common_b.is_prepared_ = false
		common_b.dict_num_lookups = 0
		common_b.dict_num_matches = 0
		InitializeHROLLING(self.hb, self.params)
	}

	PrepareH6(self.ha, one_shot, input_size, data)
	PrepareHROLLING(self.hb, one_shot, input_size, data)
}

func StoreH65(handle HasherHandle, data []byte, mask uint, ix uint) {
	var self *H65 = SelfH65(handle)
	StoreH6(self.ha, data, mask, ix)
	StoreHROLLING(self.hb, data, mask, ix)
}

func StoreRangeH65(handle HasherHandle, data []byte, mask uint, ix_start uint, ix_end uint) {
	var self *H65 = SelfH65(handle)
	StoreRangeH6(self.ha, data, mask, ix_start, ix_end)
	StoreRangeHROLLING(self.hb, data, mask, ix_start, ix_end)
}

func StitchToPreviousBlockH65(handle HasherHandle, num_bytes uint, position uint, ringbuffer []byte, ring_buffer_mask uint) {
	var self *H65 = SelfH65(handle)
	StitchToPreviousBlockH6(self.ha, num_bytes, position, ringbuffer, ring_buffer_mask)
	StitchToPreviousBlockHROLLING(self.hb, num_bytes, position, ringbuffer, ring_buffer_mask)
}

func PrepareDistanceCacheH65(handle HasherHandle, distance_cache []int) {
	var self *H65 = SelfH65(handle)
	PrepareDistanceCacheH6(self.ha, distance_cache)
	PrepareDistanceCacheHROLLING(self.hb, &distance_cache[0])
}

func FindLongestMatchH65(handle HasherHandle, dictionary *BrotliEncoderDictionary, data []byte, ring_buffer_mask uint, distance_cache []int, cur_ix uint, max_length uint, max_backward uint, gap uint, max_distance uint, out *HasherSearchResult) {
	var self *H65 = SelfH65(handle)
	FindLongestMatchH6(self.ha, dictionary, data, ring_buffer_mask, distance_cache, cur_ix, max_length, max_backward, gap, max_distance, out)
	FindLongestMatchHROLLING(self.hb, dictionary, data, ring_buffer_mask, &distance_cache[0], cur_ix, max_length, max_backward, gap, max_distance, out)
}
