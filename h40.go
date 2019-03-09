package brotli

import "encoding/binary"

/* NOLINT(build/header_guard) */
/* Copyright 2016 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* A (forgetful) hash table to the data seen by the compressor, to
   help create backward references to previous data.

   Hashes are stored in chains which are bucketed to groups. Group of chains
   share a storage "bank". When more than "bank size" chain nodes are added,
   oldest nodes are replaced; this way several chains may share a tail. */
func (*H40) HashTypeLength() uint {
	return 4
}

func (*H40) StoreLookahead() uint {
	return 4
}

/* HashBytes is the function that chooses the bucket to place the address in.*/
func HashBytesH40(data []byte) uint {
	var h uint32 = binary.LittleEndian.Uint32(data) * kHashMul32

	/* The higher bits contain more mixture from the multiplication,
	   so we take our results from there. */
	return uint(h >> (32 - 15))
}

type SlotH40 struct {
	delta uint16
	next  uint16
}

type BankH40 struct {
	slots [1 << 16]SlotH40
}

type H40 struct {
	HasherCommon
	addr          [1 << 15]uint32
	head          [1 << 15]uint16
	tiny_hash     [65536]byte
	banks         [1]BankH40
	free_slot_idx [1]uint16
	max_hops      uint
}

func SelfH40(handle HasherHandle) *H40 {
	return handle.(*H40)
}

func (h *H40) Initialize(params *BrotliEncoderParams) {
	var q uint
	if params.quality > 6 {
		q = 7
	} else {
		q = 8
	}
	h.max_hops = q << uint(params.quality-4)
}

func (h *H40) Prepare(one_shot bool, input_size uint, data []byte) {
	var partial_prepare_threshold uint = (1 << 15) >> 6
	/* Partial preparation is 100 times slower (per socket). */
	if one_shot && input_size <= partial_prepare_threshold {
		var i uint
		for i = 0; i < input_size; i++ {
			var bucket uint = HashBytesH40(data[i:])

			/* See InitEmpty comment. */
			h.addr[bucket] = 0xCCCCCCCC

			h.head[bucket] = 0xCCCC
		}
	} else {
		/* Fill |addr| array with 0xCCCCCCCC value. Because of wrapping, position
		   processed by hasher never reaches 3GB + 64M; this makes all new chains
		   to be terminated after the first node. */
		var i int
		for i = 0; i < len(h.addr); i++ {
			h.addr[i] = 0xCCCCCCCC
		}

		h.head = [1 << 15]uint16{}
	}

	h.tiny_hash = [65536]byte{}
	h.free_slot_idx = [1]uint16{}
}

/* Look at 4 bytes at &data[ix & mask]. Compute a hash from these, and prepend
   node to corresponding chain; also update tiny_hash for current position. */
func (h *H40) Store(data []byte, mask uint, ix uint) {
	var key uint = HashBytesH40(data[ix&mask:])
	var bank uint = key & (1 - 1)
	var idx uint
	idx = uint(h.free_slot_idx[bank]) & ((1 << 16) - 1)
	h.free_slot_idx[bank]++
	var delta uint = ix - uint(h.addr[key])
	h.tiny_hash[uint16(ix)] = byte(key)
	if delta > 0xFFFF {
		delta = 0xFFFF
	}
	h.banks[bank].slots[idx].delta = uint16(delta)
	h.banks[bank].slots[idx].next = h.head[key]
	h.addr[key] = uint32(ix)
	h.head[key] = uint16(idx)
}

func (h *H40) StoreRange(data []byte, mask uint, ix_start uint, ix_end uint) {
	var i uint
	for i = ix_start; i < ix_end; i++ {
		h.Store(data, mask, i)
	}
}

func (h *H40) StitchToPreviousBlock(num_bytes uint, position uint, ringbuffer []byte, ring_buffer_mask uint) {
	if num_bytes >= h.HashTypeLength()-1 && position >= 3 {
		/* Prepare the hashes for three last bytes of the last write.
		   These could not be calculated before, since they require knowledge
		   of both the previous and the current block. */
		h.Store(ringbuffer, ring_buffer_mask, position-3)
		h.Store(ringbuffer, ring_buffer_mask, position-2)
		h.Store(ringbuffer, ring_buffer_mask, position-1)
	}
}

func (*H40) PrepareDistanceCache(distance_cache []int) {
	PrepareDistanceCache(distance_cache, 4)
}

/* Find a longest backward match of &data[cur_ix] up to the length of
   max_length and stores the position cur_ix in the hash table.

   REQUIRES: PrepareDistanceCacheH40 must be invoked for current distance cache
             values; if this method is invoked repeatedly with the same distance
             cache values, it is enough to invoke PrepareDistanceCacheH40 once.

   Does not look for matches longer than max_length.
   Does not look for matches further away than max_backward.
   Writes the best match into |out|.
   |out|->score is updated only if a better match is found. */
func (h *H40) FindLongestMatch(dictionary *BrotliEncoderDictionary, data []byte, ring_buffer_mask uint, distance_cache []int, cur_ix uint, max_length uint, max_backward uint, gap uint, max_distance uint, out *HasherSearchResult) {
	var cur_ix_masked uint = cur_ix & ring_buffer_mask
	var min_score uint = out.score
	var best_score uint = out.score
	var best_len uint = out.len
	var i uint
	var key uint = HashBytesH40(data[cur_ix_masked:])
	var tiny_hash byte = byte(key)
	/* Don't accept a short copy from far away. */
	out.len = 0

	out.len_code_delta = 0

	/* Try last distance first. */
	for i = 0; i < 4; i++ {
		var backward uint = uint(distance_cache[i])
		var prev_ix uint = (cur_ix - backward)

		/* For distance code 0 we want to consider 2-byte matches. */
		if i > 0 && h.tiny_hash[uint16(prev_ix)] != tiny_hash {
			continue
		}
		if prev_ix >= cur_ix || backward > max_backward {
			continue
		}

		prev_ix &= ring_buffer_mask
		{
			var len uint = FindMatchLengthWithLimit(data[prev_ix:], data[cur_ix_masked:], max_length)
			if len >= 2 {
				var score uint = BackwardReferenceScoreUsingLastDistance(uint(len))
				if best_score < score {
					if i != 0 {
						score -= BackwardReferencePenaltyUsingLastDistance(i)
					}
					if best_score < score {
						best_score = score
						best_len = uint(len)
						out.len = best_len
						out.distance = backward
						out.score = best_score
					}
				}
			}
		}
	}
	{
		var bank uint = key & (1 - 1)
		var backward uint = 0
		var hops uint = h.max_hops
		var delta uint = cur_ix - uint(h.addr[key])
		var slot uint = uint(h.head[key])
		for {
			tmp6 := hops
			hops--
			if tmp6 == 0 {
				break
			}
			var prev_ix uint
			var last uint = slot
			backward += delta
			if backward > max_backward {
				break
			}
			prev_ix = (cur_ix - backward) & ring_buffer_mask
			slot = uint(h.banks[bank].slots[last].next)
			delta = uint(h.banks[bank].slots[last].delta)
			if cur_ix_masked+best_len > ring_buffer_mask || prev_ix+best_len > ring_buffer_mask || data[cur_ix_masked+best_len] != data[prev_ix+best_len] {
				continue
			}
			{
				var len uint = FindMatchLengthWithLimit(data[prev_ix:], data[cur_ix_masked:], max_length)
				if len >= 4 {
					/* Comparing for >= 3 does not change the semantics, but just saves
					   for a few unnecessary binary logarithms in backward reference
					   score, since we are not interested in such short matches. */
					var score uint = BackwardReferenceScore(uint(len), backward)
					if best_score < score {
						best_score = score
						best_len = uint(len)
						out.len = best_len
						out.distance = backward
						out.score = best_score
					}
				}
			}
		}

		h.Store(data, ring_buffer_mask, cur_ix)
	}

	if out.score == min_score {
		SearchInStaticDictionary(dictionary, h, data[cur_ix_masked:], max_length, max_backward+gap, max_distance, out, false)
	}
}
