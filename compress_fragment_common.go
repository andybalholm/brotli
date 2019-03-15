package brotli

/* Compresses "input" string to the "*storage" buffer as one or more complete
   meta-blocks, and updates the "*storage_ix" bit position.

   If "is_last" is 1, emits an additional empty last meta-block.

   REQUIRES: "input_size" is greater than zero, or "is_last" is 1.
   REQUIRES: "input_size" is less or equal to maximal metablock size (1 << 24).
   REQUIRES: "command_buf" and "literal_buf" point to at least
              kCompressFragmentTwoPassBlockSize long arrays.
   REQUIRES: All elements in "table[0..table_size-1]" are initialized to zero.
   REQUIRES: "table_size" is a power of two
   OUTPUT: maximal copy distance <= |input_size|
   OUTPUT: maximal copy distance <= BROTLI_MAX_BACKWARD_LIMIT(18) */
const maxDistance = 262128

/* kHashMul32_a multiplier has these properties:
   * The multiplier must be odd. Otherwise we may lose the highest bit.
   * No long streaks of ones or zeros.
   * There is no effort to ensure that it is a prime, the oddity is enough
     for this use.
   * The number has been tuned heuristically against compression benchmarks. */
var kHashMul32_a uint32 = 0x1E35A7BD
