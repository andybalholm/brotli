package brotli

const FAST_ONE_PASS_COMPRESSION_QUALITY = 0

const FAST_TWO_PASS_COMPRESSION_QUALITY = 1

const ZOPFLIFICATION_QUALITY = 10

const HQ_ZOPFLIFICATION_QUALITY = 11

const MAX_QUALITY_FOR_STATIC_ENTROPY_CODES = 2

const MIN_QUALITY_FOR_BLOCK_SPLIT = 4

const MIN_QUALITY_FOR_NONZERO_DISTANCE_PARAMS = 4

const MIN_QUALITY_FOR_OPTIMIZE_HISTOGRAMS = 4

const MIN_QUALITY_FOR_EXTENSIVE_REFERENCE_SEARCH = 5

const MIN_QUALITY_FOR_CONTEXT_MODELING = 5

const MIN_QUALITY_FOR_HQ_CONTEXT_MODELING = 7

const MIN_QUALITY_FOR_HQ_BLOCK_SPLITTING = 10

/* For quality below MIN_QUALITY_FOR_BLOCK_SPLIT there is no block splitting,
   so we buffer at most this much literals and commands. */
const MAX_NUM_DELAYED_SYMBOLS = 0x2FFF

/* Returns hash-table size for quality levels 0 and 1. */
func MaxHashTableSize(quality int) uint {
	if quality == FAST_ONE_PASS_COMPRESSION_QUALITY {
		return 1 << 15
	} else {
		return 1 << 17
	}
}

/* The maximum length for which the zopflification uses distinct distances. */
const MAX_ZOPFLI_LEN_QUALITY_10 = 150

const MAX_ZOPFLI_LEN_QUALITY_11 = 325

/* Do not thoroughly search when a long copy is found. */
const BROTLI_LONG_COPY_QUICK_STEP = 16384

func MaxZopfliLen(params *BrotliEncoderParams) uint {
	if params.quality <= 10 {
		return MAX_ZOPFLI_LEN_QUALITY_10
	} else {
		return MAX_ZOPFLI_LEN_QUALITY_11
	}
}

/* Number of best candidates to evaluate to expand Zopfli chain. */
func MaxZopfliCandidates(params *BrotliEncoderParams) uint {
	if params.quality <= 10 {
		return 1
	} else {
		return 5
	}
}

func SanitizeParams(params *BrotliEncoderParams) {
	params.quality = brotli_min_int(BROTLI_MAX_QUALITY, brotli_max_int(BROTLI_MIN_QUALITY, params.quality))
	if params.quality <= MAX_QUALITY_FOR_STATIC_ENTROPY_CODES {
		params.large_window = false
	}

	if params.lgwin < BROTLI_MIN_WINDOW_BITS {
		params.lgwin = BROTLI_MIN_WINDOW_BITS
	} else {
		var max_lgwin int
		if params.large_window {
			max_lgwin = BROTLI_LARGE_MAX_WINDOW_BITS
		} else {
			max_lgwin = BROTLI_MAX_WINDOW_BITS
		}
		if params.lgwin > uint(max_lgwin) {
			params.lgwin = uint(max_lgwin)
		}
	}
}

/* Returns optimized lg_block value. */
func ComputeLgBlock(params *BrotliEncoderParams) int {
	var lgblock int = params.lgblock
	if params.quality == FAST_ONE_PASS_COMPRESSION_QUALITY || params.quality == FAST_TWO_PASS_COMPRESSION_QUALITY {
		lgblock = int(params.lgwin)
	} else if params.quality < MIN_QUALITY_FOR_BLOCK_SPLIT {
		lgblock = 14
	} else if lgblock == 0 {
		lgblock = 16
		if params.quality >= 9 && params.lgwin > uint(lgblock) {
			lgblock = brotli_min_int(18, int(params.lgwin))
		}
	} else {
		lgblock = brotli_min_int(BROTLI_MAX_INPUT_BLOCK_BITS, brotli_max_int(BROTLI_MIN_INPUT_BLOCK_BITS, lgblock))
	}

	return lgblock
}

/* Returns log2 of the size of main ring buffer area.
   Allocate at least lgwin + 1 bits for the ring buffer so that the newly
   added block fits there completely and we still get lgwin bits and at least
   read_block_size_bits + 1 bits because the copy tail length needs to be
   smaller than ring-buffer size. */
func ComputeRbBits(params *BrotliEncoderParams) int {
	return 1 + brotli_max_int(int(params.lgwin), params.lgblock)
}

func MaxMetablockSize(params *BrotliEncoderParams) uint {
	var bits int = brotli_min_int(ComputeRbBits(params), BROTLI_MAX_INPUT_BLOCK_BITS)
	return uint(1) << uint(bits)
}

/* When searching for backward references and have not seen matches for a long
   time, we can skip some match lookups. Unsuccessful match lookups are very
   expensive and this kind of a heuristic speeds up compression quite a lot.
   At first 8 byte strides are taken and every second byte is put to hasher.
   After 4x more literals stride by 16 bytes, every put 4-th byte to hasher.
   Applied only to qualities 2 to 9. */
func LiteralSpreeLengthForSparseSearch(params *BrotliEncoderParams) uint {
	if params.quality < 9 {
		return 64
	} else {
		return 512
	}
}

func ChooseHasher(params *BrotliEncoderParams, hparams *BrotliHasherParams) {
	if params.quality > 9 {
		hparams.type_ = 10
	} else if params.quality == 4 && params.size_hint >= 1<<20 {
		hparams.type_ = 54
	} else if params.quality < 5 {
		hparams.type_ = params.quality
	} else if params.lgwin <= 16 {
		if params.quality < 7 {
			hparams.type_ = 40
		} else if params.quality < 9 {
			hparams.type_ = 41
		} else {
			hparams.type_ = 42
		}
	} else if params.size_hint >= 1<<20 && params.lgwin >= 19 {
		hparams.type_ = 6
		hparams.block_bits = params.quality - 1
		hparams.bucket_bits = 15
		hparams.hash_len = 5
		if params.quality < 7 {
			hparams.num_last_distances_to_check = 4
		} else if params.quality < 9 {
			hparams.num_last_distances_to_check = 10
		} else {
			hparams.num_last_distances_to_check = 16
		}
	} else {
		hparams.type_ = 5
		hparams.block_bits = params.quality - 1
		if params.quality < 7 {
			hparams.bucket_bits = 14
		} else {
			hparams.bucket_bits = 15
		}
		if params.quality < 7 {
			hparams.num_last_distances_to_check = 4
		} else if params.quality < 9 {
			hparams.num_last_distances_to_check = 10
		} else {
			hparams.num_last_distances_to_check = 16
		}
	}

	if params.lgwin > 24 {
		/* Different hashers for large window brotli: not for qualities <= 2,
		   these are too fast for large window. Not for qualities >= 10: their
		   hasher already works well with large window. So the changes are:
		   H3 --> H35: for quality 3.
		   H54 --> H55: for quality 4 with size hint > 1MB
		   H6 --> H65: for qualities 5, 6, 7, 8, 9. */
		if hparams.type_ == 3 {
			hparams.type_ = 35
		}

		if hparams.type_ == 54 {
			hparams.type_ = 55
		}

		if hparams.type_ == 6 {
			hparams.type_ = 65
		}
	}
}
