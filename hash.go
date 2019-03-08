package brotli

/* Matches data against static dictionary words, and for each length l,
   for which a match is found, updates matches[l] to be the minimum possible
     (distance << 5) + len_code.
   Returns 1 if matches have been found, otherwise 0.
   Prerequisites:
     matches array is at least BROTLI_MAX_STATIC_DICTIONARY_MATCH_LEN + 1 long
     all elements are initialized to kInvalidMatch */
/* Pointer to hasher data.
 *
 * Excluding initialization and destruction, hasher can be passed as
 * HasherHandle by value.
 *
 * Typically hasher data consists of 3 sections:
 * * HasherCommon structure
 * * private structured hasher data, depending on hasher type
 * * private dynamic hasher data, depending on hasher type and parameters
 *
 */
type HasherCommon struct {
	params           BrotliHasherParams
	is_prepared_     bool
	dict_num_lookups uint
	dict_num_matches uint
}

func (h *HasherCommon) Common() *HasherCommon {
	return h
}

type HasherHandle interface {
	Common() *HasherCommon
}

type score_t uint

var kCutoffTransformsCount uint32 = 10

/*   0,  12,   27,    23,    42,    63,    56,    48,    59,    64 */
/* 0+0, 4+8, 8+19, 12+11, 16+26, 20+43, 24+32, 28+20, 32+27, 36+28 */
var kCutoffTransforms uint64 = 0x071B520ADA2D3200

type HasherSearchResult struct {
	len            uint
	distance       uint
	score          uint
	len_code_delta int
}

/* kHashMul32 multiplier has these properties:
   * The multiplier must be odd. Otherwise we may lose the highest bit.
   * No long streaks of ones or zeros.
   * There is no effort to ensure that it is a prime, the oddity is enough
     for this use.
   * The number has been tuned heuristically against compression benchmarks. */
var kHashMul32 uint32 = 0x1E35A7BD

var kHashMul64 uint64 = 0x1E35A7BD1E35A7BD

var kHashMul64Long uint64 = 0x1FE35A7BD3579BD3

func Hash14(data []byte) uint32 {
	var h uint32 = BROTLI_UNALIGNED_LOAD32LE(data) * kHashMul32

	/* The higher bits contain more mixture from the multiplication,
	   so we take our results from there. */
	return h >> (32 - 14)
}

func PrepareDistanceCache(distance_cache []int, num_distances int) {
	if num_distances > 4 {
		var last_distance int = distance_cache[0]
		distance_cache[4] = last_distance - 1
		distance_cache[5] = last_distance + 1
		distance_cache[6] = last_distance - 2
		distance_cache[7] = last_distance + 2
		distance_cache[8] = last_distance - 3
		distance_cache[9] = last_distance + 3
		if num_distances > 10 {
			var next_last_distance int = distance_cache[1]
			distance_cache[10] = next_last_distance - 1
			distance_cache[11] = next_last_distance + 1
			distance_cache[12] = next_last_distance - 2
			distance_cache[13] = next_last_distance + 2
			distance_cache[14] = next_last_distance - 3
			distance_cache[15] = next_last_distance + 3
		}
	}
}

const BROTLI_LITERAL_BYTE_SCORE = 135

const BROTLI_DISTANCE_BIT_PENALTY = 30

/* Score must be positive after applying maximal penalty. */
const BROTLI_SCORE_BASE = (BROTLI_DISTANCE_BIT_PENALTY * 8 * 8)

/* Usually, we always choose the longest backward reference. This function
   allows for the exception of that rule.

   If we choose a backward reference that is further away, it will
   usually be coded with more bits. We approximate this by assuming
   log2(distance). If the distance can be expressed in terms of the
   last four distances, we use some heuristic constants to estimate
   the bits cost. For the first up to four literals we use the bit
   cost of the literals from the literal cost model, after that we
   use the average bit cost of the cost model.

   This function is used to sometimes discard a longer backward reference
   when it is not much longer and the bit cost for encoding it is more
   than the saved literals.

   backward_reference_offset MUST be positive. */
func BackwardReferenceScore(copy_length uint, backward_reference_offset uint) uint {
	return BROTLI_SCORE_BASE + BROTLI_LITERAL_BYTE_SCORE*uint(copy_length) - BROTLI_DISTANCE_BIT_PENALTY*uint(Log2FloorNonZero(backward_reference_offset))
}

func BackwardReferenceScoreUsingLastDistance(copy_length uint) uint {
	return BROTLI_LITERAL_BYTE_SCORE*uint(copy_length) + BROTLI_SCORE_BASE + 15
}

func BackwardReferencePenaltyUsingLastDistance(distance_short_code uint) uint {
	return uint(39) + ((0x1CA10 >> (distance_short_code & 0xE)) & 0xE)
}

func TestStaticDictionaryItem(dictionary *BrotliEncoderDictionary, item uint, data []byte, max_length uint, max_backward uint, max_distance uint, out *HasherSearchResult) bool {
	var len uint
	var word_idx uint
	var offset uint
	var matchlen uint
	var backward uint
	var score uint
	len = item & 0x1F
	word_idx = item >> 5
	offset = uint(dictionary.words.offsets_by_length[len]) + len*word_idx
	if len > max_length {
		return false
	}

	matchlen = FindMatchLengthWithLimit(data, dictionary.words.data[offset:], uint(len))
	if matchlen+uint(dictionary.cutoffTransformsCount) <= len || matchlen == 0 {
		return false
	}
	{
		var cut uint = len - matchlen
		var transform_id uint = (cut << 2) + uint((dictionary.cutoffTransforms>>(cut*6))&0x3F)
		backward = max_backward + 1 + word_idx + (transform_id << dictionary.words.size_bits_by_length[len])
	}

	if backward > max_distance {
		return false
	}

	score = BackwardReferenceScore(matchlen, backward)
	if score < out.score {
		return false
	}

	out.len = matchlen
	out.len_code_delta = int(len) - int(matchlen)
	out.distance = backward
	out.score = score
	return true
}

func SearchInStaticDictionary(dictionary *BrotliEncoderDictionary, handle HasherHandle, data []byte, max_length uint, max_backward uint, max_distance uint, out *HasherSearchResult, shallow bool) {
	var key uint
	var i uint
	var self *HasherCommon = handle.Common()
	if self.dict_num_matches < self.dict_num_lookups>>7 {
		return
	}

	key = uint(Hash14(data) << 1)
	for i = 0; ; (func() { i++; key++ })() {
		var tmp uint
		if shallow {
			tmp = 1
		} else {
			tmp = 2
		}
		if i >= tmp {
			break
		}
		var item uint = uint(dictionary.hash_table[key])
		self.dict_num_lookups++
		if item != 0 {
			var item_matches bool = TestStaticDictionaryItem(dictionary, item, data, max_length, max_backward, max_distance, out)
			if item_matches {
				self.dict_num_matches++
			}
		}
	}
}

type BackwardMatch struct {
	distance        uint32
	length_and_code uint32
}

func InitBackwardMatch(self *BackwardMatch, dist uint, len uint) {
	self.distance = uint32(dist)
	self.length_and_code = uint32(len << 5)
}

func InitDictionaryBackwardMatch(self *BackwardMatch, dist uint, len uint, len_code uint) {
	self.distance = uint32(dist)
	var tmp uint
	if len == len_code {
		tmp = 0
	} else {
		tmp = len_code
	}
	self.length_and_code = uint32(len<<5 | tmp)
}

func BackwardMatchLength(self *BackwardMatch) uint {
	return uint(self.length_and_code >> 5)
}

func BackwardMatchLengthCode(self *BackwardMatch) uint {
	var code uint = uint(self.length_and_code) & 31
	if code != 0 {
		return code
	} else {
		return BackwardMatchLength(self)
	}
}

func DestroyHasher(handle *HasherHandle) {
	if *handle == nil {
		return
	}
	*handle = nil
}

func HasherReset(handle HasherHandle) {
	if handle == nil {
		return
	}
	handle.Common().is_prepared_ = false
}

func HasherSetup(handle *HasherHandle, params *BrotliEncoderParams, data []byte, position uint, input_size uint, is_last bool) {
	var self HasherHandle = nil
	var common *HasherCommon = nil
	var one_shot bool = (position == 0 && is_last)
	if *handle == nil {
		ChooseHasher(params, &params.hasher)
		switch params.hasher.type_ {
		case 2:
			self = new(H2)
		case 3:
			self = new(H3)
		case 4:
			self = new(H4)
		case 5:
			self = new(H5)
		case 6:
			self = new(H6)
		case 40:
			self = new(H40)
		case 41:
			self = new(H41)
		case 42:
			self = new(H42)
		case 54:
			self = new(H54)
		case 35:
			self = new(H35)
		case 55:
			self = new(H55)
		case 65:
			self = new(H65)
		case 10:
			self = new(H10)
		}

		*handle = self
		common = self.Common()
		common.params = params.hasher
		switch common.params.type_ {
		case 2:
			InitializeH2(*handle, params)
		case 3:
			InitializeH3(*handle, params)
		case 4:
			InitializeH4(*handle, params)
		case 5:
			InitializeH5(*handle, params)
		case 6:
			InitializeH6(*handle, params)
		case 40:
			InitializeH40(*handle, params)
		case 41:
			InitializeH41(*handle, params)
		case 42:
			InitializeH42(*handle, params)
		case 54:
			InitializeH54(*handle, params)
		case 35:
			InitializeH35(*handle, params)
		case 55:
			InitializeH55(*handle, params)
		case 65:
			InitializeH65(*handle, params)
		case 10:
			InitializeH10(*handle, params)

		default:
			break
		}

		HasherReset(*handle)
	}

	self = *handle
	common = self.Common()
	if !common.is_prepared_ {
		switch common.params.type_ {
		case 2:
			PrepareH2(self, one_shot, input_size, data)
		case 3:
			PrepareH3(self, one_shot, input_size, data)
		case 4:
			PrepareH4(self, one_shot, input_size, data)
		case 5:
			PrepareH5(self, one_shot, input_size, data)
		case 6:
			PrepareH6(self, one_shot, input_size, data)
		case 40:
			PrepareH40(self, one_shot, input_size, data)
		case 41:
			PrepareH41(self, one_shot, input_size, data)
		case 42:
			PrepareH42(self, one_shot, input_size, data)
		case 54:
			PrepareH54(self, one_shot, input_size, data)
		case 35:
			PrepareH35(self, one_shot, input_size, data)
		case 55:
			PrepareH55(self, one_shot, input_size, data)
		case 65:
			PrepareH65(self, one_shot, input_size, data)
		case 10:
			PrepareH10(self, one_shot, input_size, data)

		default:
			break
		}

		if position == 0 {
			common.dict_num_lookups = 0
			common.dict_num_matches = 0
		}

		common.is_prepared_ = true
	}
}

func InitOrStitchToPreviousBlock(handle *HasherHandle, data []byte, mask uint, params *BrotliEncoderParams, position uint, input_size uint, is_last bool) {
	var self HasherHandle
	HasherSetup(handle, params, data, position, input_size, is_last)
	self = *handle
	switch self.Common().params.type_ {
	case 2:
		StitchToPreviousBlockH2(self, input_size, position, data, mask)
	case 3:
		StitchToPreviousBlockH3(self, input_size, position, data, mask)
	case 4:
		StitchToPreviousBlockH4(self, input_size, position, data, mask)
	case 5:
		StitchToPreviousBlockH5(self, input_size, position, data, mask)
	case 6:
		StitchToPreviousBlockH6(self, input_size, position, data, mask)
	case 40:
		StitchToPreviousBlockH40(self, input_size, position, data, mask)
	case 41:
		StitchToPreviousBlockH41(self, input_size, position, data, mask)
	case 42:
		StitchToPreviousBlockH42(self, input_size, position, data, mask)
	case 54:
		StitchToPreviousBlockH54(self, input_size, position, data, mask)
	case 35:
		StitchToPreviousBlockH35(self, input_size, position, data, mask)
	case 55:
		StitchToPreviousBlockH55(self, input_size, position, data, mask)
	case 65:
		StitchToPreviousBlockH65(self, input_size, position, data, mask)
	case 10:
		StitchToPreviousBlockH10(self, input_size, position, data, mask)

	default:
		break
	}
}
