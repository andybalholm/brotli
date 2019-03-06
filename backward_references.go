package brotli

/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Function to find backward reference copies. */
/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Function to find backward reference copies. */

/* "commands" points to the next output command to write to, "*num_commands" is
   initially the total amount of commands output by previous
   CreateBackwardReferences calls, and must be incremented by the amount written
   by this call. */
func ComputeDistanceCode(distance uint, max_distance uint, dist_cache []int) uint {
	if distance <= max_distance {
		var distance_plus_3 uint = distance + 3
		var offset0 uint = distance_plus_3 - uint(dist_cache[0])
		var offset1 uint = distance_plus_3 - uint(dist_cache[1])
		if distance == uint(dist_cache[0]) {
			return 0
		} else if distance == uint(dist_cache[1]) {
			return 1
		} else if offset0 < 7 {
			return (0x9750468 >> (4 * offset0)) & 0xF
		} else if offset1 < 7 {
			return (0xFDB1ACE >> (4 * offset1)) & 0xF
		} else if distance == uint(dist_cache[2]) {
			return 2
		} else if distance == uint(dist_cache[3]) {
			return 3
		}
	}

	return distance + BROTLI_NUM_DISTANCE_SHORT_CODES - 1
}

func BrotliCreateBackwardReferences(num_bytes uint, position uint, ringbuffer []byte, ringbuffer_mask uint, params *BrotliEncoderParams, hasher HasherHandle, dist_cache []int, last_insert_len *uint, commands []Command, num_commands *uint, num_literals *uint) {
	switch params.hasher.type_ {
	case 2:
		CreateBackwardReferencesNH2(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 3:
		CreateBackwardReferencesNH3(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 4:
		CreateBackwardReferencesNH4(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 5:
		CreateBackwardReferencesNH5(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 6:
		CreateBackwardReferencesNH6(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 40:
		CreateBackwardReferencesNH40(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 41:
		CreateBackwardReferencesNH41(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 42:
		CreateBackwardReferencesNH42(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 54:
		CreateBackwardReferencesNH54(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 35:
		CreateBackwardReferencesNH35(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 55:
		CreateBackwardReferencesNH55(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return
	case 65:
		CreateBackwardReferencesNH65(num_bytes, position, ringbuffer, ringbuffer_mask, params, hasher, dist_cache, last_insert_len, commands, num_commands, num_literals)
		return

	default:
		break
	}
}
