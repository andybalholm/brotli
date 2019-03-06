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
/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* This class models a sequence of literals and a backward reference copy. */
/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Functions for encoding of integers into prefix codes the amount of extra
   bits, and the actual values of the extra bits. */

/* Here distance_code is an intermediate code, i.e. one of the special codes or
   the actual distance increased by BROTLI_NUM_DISTANCE_SHORT_CODES - 1. */
func PrefixEncodeCopyDistance(distance_code uint, num_direct_codes uint, postfix_bits uint, code *uint16, extra_bits *uint32) {
	if distance_code < BROTLI_NUM_DISTANCE_SHORT_CODES+num_direct_codes {
		*code = uint16(distance_code)
		*extra_bits = 0
		return
	} else {
		var dist uint = (uint(1) << (postfix_bits + 2)) + (distance_code - BROTLI_NUM_DISTANCE_SHORT_CODES - num_direct_codes)
		var bucket uint = uint(Log2FloorNonZero(dist) - 1)
		var postfix_mask uint = (1 << postfix_bits) - 1
		var postfix uint = dist & postfix_mask
		var prefix uint = (dist >> bucket) & 1
		var offset uint = (2 + prefix) << bucket
		var nbits uint = bucket - postfix_bits
		*code = uint16(nbits<<10 | (BROTLI_NUM_DISTANCE_SHORT_CODES + num_direct_codes + ((2*(nbits-1) + prefix) << postfix_bits) + postfix))
		*extra_bits = uint32((dist - offset) >> postfix_bits)
	}
}
