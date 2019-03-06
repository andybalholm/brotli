package brotli

/* Uses the slow shortest-path block splitter and does context clustering.
   The distance parameters are dynamically selected based on the commands
   which get recomputed under the new distance parameters. The new distance
   parameters are stored into *params. */

/* Uses a fast greedy block splitter that tries to merge current block with the
   last or the second last block and uses a static context clustering which
   is the same for all block types. */
/* All Store functions here will use a storage_ix, which is always the bit
   position for the current storage. */

/* REQUIRES: length > 0 */
/* REQUIRES: length <= (1 << 24) */

/* Stores the meta-block without doing any block splitting, just collects
   one histogram per block category and uses that for entropy coding.
   REQUIRES: length > 0
   REQUIRES: length <= (1 << 24) */

/* Same as above, but uses static prefix codes for histograms with a only a few
   symbols, and uses static code length prefix codes for all other histograms.
   REQUIRES: length > 0
   REQUIRES: length <= (1 << 24) */

/* This is for storing uncompressed blocks (simple raw storage of
   bytes-as-bytes).
   REQUIRES: length > 0
   REQUIRES: length <= (1 << 24) */
/* Copyright 2015 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Static entropy codes used for faster meta-block encoding. */
/* Copyright 2010 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Write bits into a byte array. */

/*#define BIT_WRITER_DEBUG */

/* This function writes bits into bytes in increasing addresses, and within
   a byte least-significant-bit first.

   The function can write up to 56 bits in one go with WriteBits
   Example: let's assume that 3 bits (Rs below) have been written already:

   BYTE-0     BYTE+1       BYTE+2

   0000 0RRR    0000 0000    0000 0000

   Now, we could write 5 or less bits in MSB by just sifting by 3
   and OR'ing to BYTE-0.

   For n bits, we take the last 5 bits, OR that with high bits in BYTE-0,
   and locate the rest in BYTE+1, BYTE+2, etc. */
func BrotliWriteBits(n_bits uint, bits uint64, pos *uint, array []byte) {
	var array_pos []byte = array[*pos>>3:]
	var bits_reserved_in_first_byte uint = (*pos & 7)
	/* implicit & 0xFF is assumed for uint8_t arithmetics */

	var bits_left_to_write uint
	bits <<= bits_reserved_in_first_byte
	array_pos[0] |= byte(bits)
	array_pos = array_pos[1:]
	for bits_left_to_write = n_bits + bits_reserved_in_first_byte; bits_left_to_write >= 9; bits_left_to_write -= 8 {
		bits >>= 8
		array_pos[0] = byte(bits)
		array_pos = array_pos[1:]
	}

	array_pos[0] = 0
	*pos += n_bits
}

func BrotliWriteSingleBit(bit bool, pos *uint, array []byte) {
	if bit {
		BrotliWriteBits(1, 1, pos, array)
	} else {
		BrotliWriteBits(1, 0, pos, array)
	}
}

func BrotliWriteBitsPrepareStorage(pos uint, array []byte) {
	assert(pos&7 == 0)
	array[pos>>3] = 0
}
