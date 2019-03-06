package brotli

/* transforms is a part of ABI, but not API.

   It means that there are some functions that are supposed to be in "common"
   library, but header itself is not placed into include/brotli. This way,
   aforementioned functions will be available only to brotli internals.
*/
const (
	BROTLI_TRANSFORM_IDENTITY        = 0
	BROTLI_TRANSFORM_OMIT_LAST_1     = 1
	BROTLI_TRANSFORM_OMIT_LAST_2     = 2
	BROTLI_TRANSFORM_OMIT_LAST_3     = 3
	BROTLI_TRANSFORM_OMIT_LAST_4     = 4
	BROTLI_TRANSFORM_OMIT_LAST_5     = 5
	BROTLI_TRANSFORM_OMIT_LAST_6     = 6
	BROTLI_TRANSFORM_OMIT_LAST_7     = 7
	BROTLI_TRANSFORM_OMIT_LAST_8     = 8
	BROTLI_TRANSFORM_OMIT_LAST_9     = 9
	BROTLI_TRANSFORM_UPPERCASE_FIRST = 10
	BROTLI_TRANSFORM_UPPERCASE_ALL   = 11
	BROTLI_TRANSFORM_OMIT_FIRST_1    = 12
	BROTLI_TRANSFORM_OMIT_FIRST_2    = 13
	BROTLI_TRANSFORM_OMIT_FIRST_3    = 14
	BROTLI_TRANSFORM_OMIT_FIRST_4    = 15
	BROTLI_TRANSFORM_OMIT_FIRST_5    = 16
	BROTLI_TRANSFORM_OMIT_FIRST_6    = 17
	BROTLI_TRANSFORM_OMIT_FIRST_7    = 18
	BROTLI_TRANSFORM_OMIT_FIRST_8    = 19
	BROTLI_TRANSFORM_OMIT_FIRST_9    = 20
	BROTLI_TRANSFORM_SHIFT_FIRST     = 21
	BROTLI_TRANSFORM_SHIFT_ALL       = 22 + iota - 22
	BROTLI_NUM_TRANSFORM_TYPES
)

const BROTLI_TRANSFORMS_MAX_CUT_OFF = BROTLI_TRANSFORM_OMIT_LAST_9

type BrotliTransforms struct {
	prefix_suffix_size uint16
	prefix_suffix      []byte
	prefix_suffix_map  []uint16
	num_transforms     uint32
	transforms         []byte
	params             []byte
	cutOffTransforms   [BROTLI_TRANSFORMS_MAX_CUT_OFF + 1]int16
}

/* T is BrotliTransforms*; result is uint8_t. */
func BROTLI_TRANSFORM_PREFIX_ID(t *BrotliTransforms, I int) byte {
	return t.transforms[(I*3)+0]
}

func BROTLI_TRANSFORM_TYPE(t *BrotliTransforms, I int) byte {
	return t.transforms[(I*3)+1]
}

func BROTLI_TRANSFORM_SUFFIX_ID(t *BrotliTransforms, I int) byte {
	return t.transforms[(I*3)+2]
}

/* T is BrotliTransforms*; result is const uint8_t*. */
func BROTLI_TRANSFORM_PREFIX(t *BrotliTransforms, I int) []byte {
	return t.prefix_suffix[t.prefix_suffix_map[BROTLI_TRANSFORM_PREFIX_ID(t, I)]:]
}

func BROTLI_TRANSFORM_SUFFIX(t *BrotliTransforms, I int) []byte {
	return t.prefix_suffix[t.prefix_suffix_map[BROTLI_TRANSFORM_SUFFIX_ID(t, I)]:]
}

/* RFC 7932 transforms string data */
var kPrefixSuffix string = "\001 \002, \010 of the \004 of \002s \001.\005 and \004 " + "in \001\"\004 to \002\">\001\n\002. \001]\005 for \003 a \006 " + "that \001'\006 with \006 from \004 by \001(\006. T" + "he \004 on \004 as \004 is \004ing \002\n\t\001:\003ed " + "\002=\"\004 at \003ly \001,\002='\005.com/\007. This \005" + " not \003er \003al \004ful \004ive \005less \004es" + "t \004ize \002\xc2\xa0\004ous \005 the \002e \000"

/* 0x  _0 _2  __5        _E    _3  _6 _8     _E */

/* 2x     _3_ _5    _A_  _D_ _F  _2 _4     _A   _E */

/* 4x       _5_ _7      _E      _5    _A _C */

/* 6x     _3    _8    _D    _2    _7_ _ _A _C */

/* 8x  _0 _ _3    _8   _C _E _ _1     _7       _F */

/* Ax       _5   _9   _D    _2    _7     _D */

/* Cx    _2    _7___ ___ _A    _F     _5  _8 */
var kPrefixSuffixMap = [50]uint16{
	0x00,
	0x02,
	0x05,
	0x0E,
	0x13,
	0x16,
	0x18,
	0x1E,
	0x23,
	0x25,
	0x2A,
	0x2D,
	0x2F,
	0x32,
	0x34,
	0x3A,
	0x3E,
	0x45,
	0x47,
	0x4E,
	0x55,
	0x5A,
	0x5C,
	0x63,
	0x68,
	0x6D,
	0x72,
	0x77,
	0x7A,
	0x7C,
	0x80,
	0x83,
	0x88,
	0x8C,
	0x8E,
	0x91,
	0x97,
	0x9F,
	0xA5,
	0xA9,
	0xAD,
	0xB2,
	0xB7,
	0xBD,
	0xC2,
	0xC7,
	0xCA,
	0xCF,
	0xD5,
	0xD8,
}

/* RFC 7932 transforms */
var kTransformsData = []byte{
	49,
	BROTLI_TRANSFORM_IDENTITY,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	0,
	0,
	BROTLI_TRANSFORM_IDENTITY,
	0,
	49,
	BROTLI_TRANSFORM_OMIT_FIRST_1,
	49,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	0,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	47,
	0,
	BROTLI_TRANSFORM_IDENTITY,
	49,
	4,
	BROTLI_TRANSFORM_IDENTITY,
	0,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	3,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	6,
	49,
	BROTLI_TRANSFORM_OMIT_FIRST_2,
	49,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_1,
	49,
	1,
	BROTLI_TRANSFORM_IDENTITY,
	0,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	1,
	0,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	0,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	7,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	9,
	48,
	BROTLI_TRANSFORM_IDENTITY,
	0,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	8,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	5,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	10,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	11,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_3,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	13,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	14,
	49,
	BROTLI_TRANSFORM_OMIT_FIRST_3,
	49,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_2,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	15,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	16,
	0,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	12,
	5,
	BROTLI_TRANSFORM_IDENTITY,
	49,
	0,
	BROTLI_TRANSFORM_IDENTITY,
	1,
	49,
	BROTLI_TRANSFORM_OMIT_FIRST_4,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	18,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	17,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	19,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	20,
	49,
	BROTLI_TRANSFORM_OMIT_FIRST_5,
	49,
	49,
	BROTLI_TRANSFORM_OMIT_FIRST_6,
	49,
	47,
	BROTLI_TRANSFORM_IDENTITY,
	49,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_4,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	22,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	23,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	24,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	25,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_7,
	49,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_1,
	26,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	27,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	28,
	0,
	BROTLI_TRANSFORM_IDENTITY,
	12,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	29,
	49,
	BROTLI_TRANSFORM_OMIT_FIRST_9,
	49,
	49,
	BROTLI_TRANSFORM_OMIT_FIRST_7,
	49,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_6,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	21,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	1,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_8,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	31,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	32,
	47,
	BROTLI_TRANSFORM_IDENTITY,
	3,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_5,
	49,
	49,
	BROTLI_TRANSFORM_OMIT_LAST_9,
	49,
	0,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	1,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	8,
	5,
	BROTLI_TRANSFORM_IDENTITY,
	21,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	0,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	10,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	30,
	0,
	BROTLI_TRANSFORM_IDENTITY,
	5,
	35,
	BROTLI_TRANSFORM_IDENTITY,
	49,
	47,
	BROTLI_TRANSFORM_IDENTITY,
	2,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	17,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	36,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	33,
	5,
	BROTLI_TRANSFORM_IDENTITY,
	0,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	21,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	5,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	37,
	0,
	BROTLI_TRANSFORM_IDENTITY,
	30,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	38,
	0,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	0,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	39,
	0,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	49,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	34,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	8,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	12,
	0,
	BROTLI_TRANSFORM_IDENTITY,
	21,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	40,
	0,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	12,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	41,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	42,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	17,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	43,
	0,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	5,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	10,
	0,
	BROTLI_TRANSFORM_IDENTITY,
	34,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	33,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	44,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	5,
	45,
	BROTLI_TRANSFORM_IDENTITY,
	49,
	0,
	BROTLI_TRANSFORM_IDENTITY,
	33,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	30,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	30,
	49,
	BROTLI_TRANSFORM_IDENTITY,
	46,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	1,
	49,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	34,
	0,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	33,
	0,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	30,
	0,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	1,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	33,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	21,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	12,
	0,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	5,
	49,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	34,
	0,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	12,
	0,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	30,
	0,
	BROTLI_TRANSFORM_UPPERCASE_ALL,
	34,
	0,
	BROTLI_TRANSFORM_UPPERCASE_FIRST,
	34,
}

var kBrotliTransforms = BrotliTransforms{
	217,
	[]byte(kPrefixSuffix),
	kPrefixSuffixMap[:],
	121,
	kTransformsData,
	nil, /* no extra parameters */
	[BROTLI_TRANSFORMS_MAX_CUT_OFF + 1]int16{0, 12, 27, 23, 42, 63, 56, 48, 59, 64},
}

func BrotliGetTransforms() *BrotliTransforms {
	return &kBrotliTransforms
}

func ToUpperCase(p []byte) int {
	if p[0] < 0xC0 {
		if p[0] >= 'a' && p[0] <= 'z' {
			p[0] ^= 32
		}

		return 1
	}

	/* An overly simplified uppercasing model for UTF-8. */
	if p[0] < 0xE0 {
		p[1] ^= 32
		return 2
	}

	/* An arbitrary transform for three byte characters. */
	p[2] ^= 5

	return 3
}

func Shift(word []byte, word_len int, parameter uint16) int {
	/* Limited sign extension: scalar < (1 << 24). */
	var scalar uint32 = (uint32(parameter) & 0x7FFF) + (0x1000000 - (uint32(parameter) & 0x8000))
	if word[0] < 0x80 {
		/* 1-byte rune / 0sssssss / 7 bit scalar (ASCII). */
		scalar += uint32(word[0])

		word[0] = byte(scalar & 0x7F)
		return 1
	} else if word[0] < 0xC0 {
		/* Continuation / 10AAAAAA. */
		return 1
	} else if word[0] < 0xE0 {
		/* 2-byte rune / 110sssss AAssssss / 11 bit scalar. */
		if word_len < 2 {
			return 1
		}
		scalar += uint32(word[1]&0x3F | (word[0]&0x1F)<<6)
		word[0] = byte(0xC0 | (scalar>>6)&0x1F)
		word[1] = byte(uint32(word[1]&0xC0) | scalar&0x3F)
		return 2
	} else if word[0] < 0xF0 {
		/* 3-byte rune / 1110ssss AAssssss BBssssss / 16 bit scalar. */
		if word_len < 3 {
			return word_len
		}
		scalar += uint32(word[2]&0x3F | (word[1]&0x3F)<<6 | (word[0]&0x0F)<<12)
		word[0] = byte(0xE0 | (scalar>>12)&0x0F)
		word[1] = byte(uint32(word[1]&0xC0) | (scalar>>6)&0x3F)
		word[2] = byte(uint32(word[2]&0xC0) | scalar&0x3F)
		return 3
	} else if word[0] < 0xF8 {
		/* 4-byte rune / 11110sss AAssssss BBssssss CCssssss / 21 bit scalar. */
		if word_len < 4 {
			return word_len
		}
		scalar += uint32(word[3]&0x3F | (word[2]&0x3F)<<6 | (word[1]&0x3F)<<12 | (word[0]&0x07)<<18)
		word[0] = byte(0xF0 | (scalar>>18)&0x07)
		word[1] = byte(uint32(word[1]&0xC0) | (scalar>>12)&0x3F)
		word[2] = byte(uint32(word[2]&0xC0) | (scalar>>6)&0x3F)
		word[3] = byte(uint32(word[3]&0xC0) | scalar&0x3F)
		return 4
	}

	return 1
}

func BrotliTransformDictionaryWord(dst []byte, word []byte, len int, transforms *BrotliTransforms, transform_idx int) int {
	var idx int = 0
	var prefix []byte = BROTLI_TRANSFORM_PREFIX(transforms, transform_idx)
	var type_ byte = BROTLI_TRANSFORM_TYPE(transforms, transform_idx)
	var suffix []byte = BROTLI_TRANSFORM_SUFFIX(transforms, transform_idx)
	{
		var prefix_len int = int(prefix[0])
		prefix = prefix[1:]
		for {
			tmp1 := prefix_len
			prefix_len--
			if tmp1 == 0 {
				break
			}
			dst[idx] = prefix[0]
			idx++
			prefix = prefix[1:]
		}
	}
	{
		var t int = int(type_)
		var i int = 0
		if t <= BROTLI_TRANSFORM_OMIT_LAST_9 {
			len -= t
		} else if t >= BROTLI_TRANSFORM_OMIT_FIRST_1 && t <= BROTLI_TRANSFORM_OMIT_FIRST_9 {
			var skip int = t - (BROTLI_TRANSFORM_OMIT_FIRST_1 - 1)
			word = word[skip:]
			len -= skip
		}

		for i < len {
			dst[idx] = word[i]
			idx++
			i++
		}
		if t == BROTLI_TRANSFORM_UPPERCASE_FIRST {
			ToUpperCase(dst[idx-len:])
		} else if t == BROTLI_TRANSFORM_UPPERCASE_ALL {
			var uppercase []byte = dst
			uppercase = uppercase[idx-len:]
			for len > 0 {
				var step int = ToUpperCase(uppercase)
				uppercase = uppercase[step:]
				len -= step
			}
		} else if t == BROTLI_TRANSFORM_SHIFT_FIRST {
			var param uint16 = uint16(transforms.params[transform_idx*2] + (transforms.params[transform_idx*2+1] << 8))
			Shift(dst[idx-len:], int(len), param)
		} else if t == BROTLI_TRANSFORM_SHIFT_ALL {
			var param uint16 = uint16(transforms.params[transform_idx*2] + (transforms.params[transform_idx*2+1] << 8))
			var shift []byte = dst
			shift = shift[idx-len:]
			for len > 0 {
				var step int = Shift(shift, int(len), param)
				shift = shift[step:]
				len -= step
			}
		}
	}
	{
		var suffix_len int = int(suffix[0])
		suffix = suffix[1:]
		for {
			tmp2 := suffix_len
			suffix_len--
			if tmp2 == 0 {
				break
			}
			dst[idx] = suffix[0]
			idx++
			suffix = suffix[1:]
		}
		return idx
	}
}
