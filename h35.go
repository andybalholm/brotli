package brotli

/* NOTE: this hasher does not search in the dictionary. It is used as
   backup-hasher, the main hasher already searches in it. */
/* NOLINT(build/header_guard) */
/* Copyright 2018 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Composite hasher: This hasher allows to combine two other hashers, HASHER_A
   and HASHER_B. */
func HashTypeLengthH35() uint {
	var a uint = HashTypeLengthH3()
	var b uint = HashTypeLengthHROLLING_FAST()
	if a > b {
		return a
	} else {
		return b
	}
}

func StoreLookaheadH35() uint {
	var a uint = StoreLookaheadH3()
	var b uint = StoreLookaheadHROLLING_FAST()
	if a > b {
		return a
	} else {
		return b
	}
}

type H35 struct {
	HasherCommon
	ha     HasherHandle
	hb     HasherHandle
	params *BrotliEncoderParams
}

func SelfH35(handle HasherHandle) *H35 {
	return handle.(*H35)
}

func InitializeH35(handle HasherHandle, params *BrotliEncoderParams) {
	var self *H35 = SelfH35(handle)
	self.ha = nil
	self.hb = nil
	self.params = params
}

/* TODO: Initialize of the hashers is defered to Prepare (and params
   remembered here) because we don't get the one_shot and input_size params
   here that are needed to know the memory size of them. Instead provide
   those params to all hashers InitializeH35 */
func PrepareH35(handle HasherHandle, one_shot bool, input_size uint, data []byte) {
	var self *H35 = SelfH35(handle)
	if self.ha == nil {
		var common_a *HasherCommon
		var common_b *HasherCommon

		self.ha = new(H3)
		common_a = self.ha.Common()
		common_a.params = self.params.hasher
		common_a.is_prepared_ = false
		common_a.dict_num_lookups = 0
		common_a.dict_num_matches = 0
		InitializeH3(self.ha, self.params)

		self.hb = new(HROLLING_FAST)
		common_b = self.hb.Common()
		common_b.params = self.params.hasher
		common_b.is_prepared_ = false
		common_b.dict_num_lookups = 0
		common_b.dict_num_matches = 0
		InitializeHROLLING_FAST(self.hb, self.params)
	}

	PrepareH3(self.ha, one_shot, input_size, data)
	PrepareHROLLING_FAST(self.hb, one_shot, input_size, data)
}

func StoreH35(handle HasherHandle, data []byte, mask uint, ix uint) {
	var self *H35 = SelfH35(handle)
	StoreH3(self.ha, data, mask, ix)
	StoreHROLLING_FAST(self.hb, data, mask, ix)
}

func StoreRangeH35(handle HasherHandle, data []byte, mask uint, ix_start uint, ix_end uint) {
	var self *H35 = SelfH35(handle)
	StoreRangeH3(self.ha, data, mask, ix_start, ix_end)
	StoreRangeHROLLING_FAST(self.hb, data, mask, ix_start, ix_end)
}

func StitchToPreviousBlockH35(handle HasherHandle, num_bytes uint, position uint, ringbuffer []byte, ring_buffer_mask uint) {
	var self *H35 = SelfH35(handle)
	StitchToPreviousBlockH3(self.ha, num_bytes, position, ringbuffer, ring_buffer_mask)
	StitchToPreviousBlockHROLLING_FAST(self.hb, num_bytes, position, ringbuffer, ring_buffer_mask)
}

func PrepareDistanceCacheH35(handle HasherHandle, distance_cache []int) {
	var self *H35 = SelfH35(handle)
	PrepareDistanceCacheH3(self.ha, distance_cache)
	PrepareDistanceCacheHROLLING_FAST(self.hb, &distance_cache[0])
}

func FindLongestMatchH35(handle HasherHandle, dictionary *BrotliEncoderDictionary, data []byte, ring_buffer_mask uint, distance_cache []int, cur_ix uint, max_length uint, max_backward uint, gap uint, max_distance uint, out *HasherSearchResult) {
	var self *H35 = SelfH35(handle)
	FindLongestMatchH3(self.ha, dictionary, data, ring_buffer_mask, distance_cache, cur_ix, max_length, max_backward, gap, max_distance, out)
	FindLongestMatchHROLLING_FAST(self.hb, dictionary, data, ring_buffer_mask, &distance_cache[0], cur_ix, max_length, max_backward, gap, max_distance, out)
}
