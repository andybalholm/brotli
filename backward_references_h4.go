package brotli

/* NOLINT(build/header_guard) */
/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/
func CreateBackwardReferencesNH4(num_bytes uint, position uint, ringbuffer []byte, ringbuffer_mask uint, params *BrotliEncoderParams, hasher HasherHandle, dist_cache []int, last_insert_len *uint, commands []Command, num_commands *uint, num_literals *uint) {
	var max_backward_limit uint = BROTLI_MAX_BACKWARD_LIMIT(params.lgwin)
	var orig_commands []Command = commands
	var insert_length uint = *last_insert_len
	var pos_end uint = position + num_bytes
	var store_end uint
	if num_bytes >= StoreLookaheadH4() {
		store_end = position + num_bytes - StoreLookaheadH4() + 1
	} else {
		store_end = position
	}
	var random_heuristics_window_size uint = LiteralSpreeLengthForSparseSearch(params)
	var apply_random_heuristics uint = position + random_heuristics_window_size
	var gap uint = 0
	/* Set maximum distance, see section 9.1. of the spec. */

	var kMinScore uint = BROTLI_SCORE_BASE + 100

	/* For speed up heuristics for random data. */

	/* Minimum score to accept a backward reference. */
	PrepareDistanceCacheH4(hasher, dist_cache)

	for position+HashTypeLengthH4() < pos_end {
		var max_length uint = pos_end - position
		var max_distance uint = brotli_min_size_t(position, max_backward_limit)
		var sr HasherSearchResult
		sr.len = 0
		sr.len_code_delta = 0
		sr.distance = 0
		sr.score = kMinScore
		FindLongestMatchH4(hasher, &params.dictionary, ringbuffer, ringbuffer_mask, dist_cache, position, max_length, max_distance, gap, params.dist.max_distance, &sr)
		if sr.score > kMinScore {
			/* Found a match. Let's look for something even better ahead. */
			var delayed_backward_references_in_row int = 0
			max_length--
			for ; ; max_length-- {
				var cost_diff_lazy uint = 175
				var sr2 HasherSearchResult
				if params.quality < MIN_QUALITY_FOR_EXTENSIVE_REFERENCE_SEARCH {
					sr2.len = brotli_min_size_t(sr.len-1, max_length)
				} else {
					sr2.len = 0
				}
				sr2.len_code_delta = 0
				sr2.distance = 0
				sr2.score = kMinScore
				max_distance = brotli_min_size_t(position+1, max_backward_limit)
				FindLongestMatchH4(hasher, &params.dictionary, ringbuffer, ringbuffer_mask, dist_cache, position+1, max_length, max_distance, gap, params.dist.max_distance, &sr2)
				if sr2.score >= sr.score+cost_diff_lazy {
					/* Ok, let's just write one byte for now and start a match from the
					   next byte. */
					position++

					insert_length++
					sr = sr2
					delayed_backward_references_in_row++
					if delayed_backward_references_in_row < 4 && position+HashTypeLengthH4() < pos_end {
						continue
					}
				}

				break
			}

			apply_random_heuristics = position + 2*sr.len + random_heuristics_window_size
			max_distance = brotli_min_size_t(position, max_backward_limit)
			{
				/* The first 16 codes are special short-codes,
				   and the minimum offset is 1. */
				var distance_code uint = ComputeDistanceCode(sr.distance, max_distance+gap, dist_cache)
				if (sr.distance <= (max_distance + gap)) && distance_code > 0 {
					dist_cache[3] = dist_cache[2]
					dist_cache[2] = dist_cache[1]
					dist_cache[1] = dist_cache[0]
					dist_cache[0] = int(sr.distance)
					PrepareDistanceCacheH4(hasher, dist_cache)
				}

				InitCommand(&commands[0], &params.dist, insert_length, sr.len, sr.len_code_delta, distance_code)
				commands = commands[1:]
			}

			*num_literals += insert_length
			insert_length = 0
			/* Put the hash keys into the table, if there are enough bytes left.
			   Depending on the hasher implementation, it can push all positions
			   in the given range or only a subset of them.
			   Avoid hash poisoning with RLE data. */
			{
				var range_start uint = position + 2
				var range_end uint = brotli_min_size_t(position+sr.len, store_end)
				if sr.distance < sr.len>>2 {
					range_start = brotli_min_size_t(range_end, brotli_max_size_t(range_start, position+sr.len-(sr.distance<<2)))
				}

				StoreRangeH4(hasher, ringbuffer, ringbuffer_mask, range_start, range_end)
			}

			position += sr.len
		} else {
			insert_length++
			position++

			/* If we have not seen matches for a long time, we can skip some
			   match lookups. Unsuccessful match lookups are very very expensive
			   and this kind of a heuristic speeds up compression quite
			   a lot. */
			if position > apply_random_heuristics {
				/* Going through uncompressible data, jump. */
				if position > apply_random_heuristics+4*random_heuristics_window_size {
					var kMargin uint = brotli_max_size_t(StoreLookaheadH4()-1, 4)
					/* It is quite a long time since we saw a copy, so we assume
					   that this data is not compressible, and store hashes less
					   often. Hashes of non compressible data are less likely to
					   turn out to be useful in the future, too, so we store less of
					   them to not to flood out the hash table of good compressible
					   data. */

					var pos_jump uint = brotli_min_size_t(position+16, pos_end-kMargin)
					for ; position < pos_jump; position += 4 {
						StoreH4(hasher, ringbuffer, ringbuffer_mask, position)
						insert_length += 4
					}
				} else {
					var kMargin uint = brotli_max_size_t(StoreLookaheadH4()-1, 2)
					var pos_jump uint = brotli_min_size_t(position+8, pos_end-kMargin)
					for ; position < pos_jump; position += 2 {
						StoreH4(hasher, ringbuffer, ringbuffer_mask, position)
						insert_length += 2
					}
				}
			}
		}
	}

	insert_length += pos_end - position
	*last_insert_len = insert_length
	*num_commands += uint(-cap(commands) + cap(orig_commands))
}
