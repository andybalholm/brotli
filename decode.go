package brotli

/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/
/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/**
 * @file
 * API for Brotli decompression.
 */

/**
 * Result type for ::BrotliDecoderDecompress and
 * ::BrotliDecoderDecompressStream functions.
 */
const (
	BROTLI_DECODER_RESULT_ERROR             = 0
	BROTLI_DECODER_RESULT_SUCCESS           = 1
	BROTLI_DECODER_RESULT_NEEDS_MORE_INPUT  = 2
	BROTLI_DECODER_RESULT_NEEDS_MORE_OUTPUT = 3
)

/**
 * Template that evaluates items of ::BrotliDecoderErrorCode.
 *
 * Example: @code {.cpp}
 * // Log Brotli error code.
 * switch (brotliDecoderErrorCode) {
 * #define CASE_(PREFIX, NAME, CODE) \
 *   case BROTLI_DECODER ## PREFIX ## NAME: \
 *     LOG(INFO) << "error code:" << #NAME; \
 *     break;
 * #define NEWLINE_
 * BROTLI_DECODER_ERROR_CODES_LIST(CASE_, NEWLINE_)
 * #undef CASE_
 * #undef NEWLINE_
 *   default: LOG(FATAL) << "unknown brotli error code";
 * }
 * @endcode
 */

/**
 * Error code for detailed logging / production debugging.
 *
 * See ::BrotliDecoderGetErrorCode and ::BROTLI_LAST_ERROR_CODE.
 */
const (
	BROTLI_DECODER_NO_ERROR                             = 0
	BROTLI_DECODER_SUCCESS                              = 1
	BROTLI_DECODER_NEEDS_MORE_INPUT                     = 2
	BROTLI_DECODER_NEEDS_MORE_OUTPUT                    = 3
	BROTLI_DECODER_ERROR_FORMAT_EXUBERANT_NIBBLE        = -1
	BROTLI_DECODER_ERROR_FORMAT_RESERVED                = -2
	BROTLI_DECODER_ERROR_FORMAT_EXUBERANT_META_NIBBLE   = -3
	BROTLI_DECODER_ERROR_FORMAT_SIMPLE_HUFFMAN_ALPHABET = -4
	BROTLI_DECODER_ERROR_FORMAT_SIMPLE_HUFFMAN_SAME     = -5
	BROTLI_DECODER_ERROR_FORMAT_CL_SPACE                = -6
	BROTLI_DECODER_ERROR_FORMAT_HUFFMAN_SPACE           = -7
	BROTLI_DECODER_ERROR_FORMAT_CONTEXT_MAP_REPEAT      = -8
	BROTLI_DECODER_ERROR_FORMAT_BLOCK_LENGTH_1          = -9
	BROTLI_DECODER_ERROR_FORMAT_BLOCK_LENGTH_2          = -10
	BROTLI_DECODER_ERROR_FORMAT_TRANSFORM               = -11
	BROTLI_DECODER_ERROR_FORMAT_DICTIONARY              = -12
	BROTLI_DECODER_ERROR_FORMAT_WINDOW_BITS             = -13
	BROTLI_DECODER_ERROR_FORMAT_PADDING_1               = -14
	BROTLI_DECODER_ERROR_FORMAT_PADDING_2               = -15
	BROTLI_DECODER_ERROR_FORMAT_DISTANCE                = -16
	BROTLI_DECODER_ERROR_DICTIONARY_NOT_SET             = -19
	BROTLI_DECODER_ERROR_INVALID_ARGUMENTS              = -20
	BROTLI_DECODER_ERROR_ALLOC_CONTEXT_MODES            = -21
	BROTLI_DECODER_ERROR_ALLOC_TREE_GROUPS              = -22
	BROTLI_DECODER_ERROR_ALLOC_CONTEXT_MAP              = -25
	BROTLI_DECODER_ERROR_ALLOC_RING_BUFFER_1            = -26
	BROTLI_DECODER_ERROR_ALLOC_RING_BUFFER_2            = -27
	BROTLI_DECODER_ERROR_ALLOC_BLOCK_TYPE_TREES         = -30
	BROTLI_DECODER_ERROR_UNREACHABLE                    = -31
)

/**
 * The value of the last error code, negative integer.
 *
 * All other error code values are in the range from ::BROTLI_LAST_ERROR_CODE
 * to @c -1. There are also 4 other possible non-error codes @c 0 .. @c 3 in
 * ::BrotliDecoderErrorCode enumeration.
 */
const BROTLI_LAST_ERROR_CODE = BROTLI_DECODER_ERROR_UNREACHABLE

/** Options to be used with ::BrotliDecoderSetParameter. */
const (
	BROTLI_DECODER_PARAM_DISABLE_RING_BUFFER_REALLOCATION = 0
	BROTLI_DECODER_PARAM_LARGE_WINDOW                     = 1
)

const HUFFMAN_TABLE_BITS = 8

const HUFFMAN_TABLE_MASK = 0xFF

/* We need the slack region for the following reasons:
   - doing up to two 16-byte copies for fast backward copying
   - inserting transformed dictionary word (5 prefix + 24 base + 8 suffix) */
var kRingBufferWriteAheadSlack uint32 = 42

var kCodeLengthCodeOrder = [codeLengthCodes]byte{1, 2, 3, 4, 0, 5, 17, 6, 16, 7, 8, 9, 10, 11, 12, 13, 14, 15}

/* Static prefix code for the complex code length code lengths. */
var kCodeLengthPrefixLength = [16]byte{2, 2, 2, 3, 2, 2, 2, 4, 2, 2, 2, 3, 2, 2, 2, 4}

var kCodeLengthPrefixValue = [16]byte{0, 4, 3, 2, 0, 4, 3, 1, 0, 4, 3, 2, 0, 4, 3, 5}

func BrotliDecoderSetParameter(state *Reader, p int, value uint32) bool {
	if state.state != BROTLI_STATE_UNINITED {
		return false
	}
	switch p {
	case BROTLI_DECODER_PARAM_DISABLE_RING_BUFFER_REALLOCATION:
		if !(value == 0) {
			state.canny_ringbuffer_allocation = 0
		} else {
			state.canny_ringbuffer_allocation = 1
		}
		return true

	case BROTLI_DECODER_PARAM_LARGE_WINDOW:
		state.large_window = (!(value == 0))
		return true

	default:
		return false
	}
}

/* Saves error code and converts it to BrotliDecoderResult. */
func SaveErrorCode(s *Reader, e int) int {
	s.error_code = int(e)
	switch e {
	case BROTLI_DECODER_SUCCESS:
		return BROTLI_DECODER_RESULT_SUCCESS

	case BROTLI_DECODER_NEEDS_MORE_INPUT:
		return BROTLI_DECODER_RESULT_NEEDS_MORE_INPUT

	case BROTLI_DECODER_NEEDS_MORE_OUTPUT:
		return BROTLI_DECODER_RESULT_NEEDS_MORE_OUTPUT

	default:
		return BROTLI_DECODER_RESULT_ERROR
	}
}

/* Decodes WBITS by reading 1 - 7 bits, or 0x11 for "Large Window Brotli".
   Precondition: bit-reader accumulator has at least 8 bits. */
func DecodeWindowBits(s *Reader, br *bitReader) int {
	var n uint32
	var large_window bool = s.large_window
	s.large_window = false
	takeBits(br, 1, &n)
	if n == 0 {
		s.window_bits = 16
		return BROTLI_DECODER_SUCCESS
	}

	takeBits(br, 3, &n)
	if n != 0 {
		s.window_bits = 17 + n
		return BROTLI_DECODER_SUCCESS
	}

	takeBits(br, 3, &n)
	if n == 1 {
		if large_window {
			takeBits(br, 1, &n)
			if n == 1 {
				return BROTLI_DECODER_ERROR_FORMAT_WINDOW_BITS
			}

			s.large_window = true
			return BROTLI_DECODER_SUCCESS
		} else {
			return BROTLI_DECODER_ERROR_FORMAT_WINDOW_BITS
		}
	}

	if n != 0 {
		s.window_bits = 8 + n
		return BROTLI_DECODER_SUCCESS
	}

	s.window_bits = 17
	return BROTLI_DECODER_SUCCESS
}

/* Decodes a number in the range [0..255], by reading 1 - 11 bits. */
func DecodeVarLenUint8(s *Reader, br *bitReader, value *uint32) int {
	var bits uint32
	switch s.substate_decode_uint8 {
	case BROTLI_STATE_DECODE_UINT8_NONE:
		if !safeReadBits(br, 1, &bits) {
			return BROTLI_DECODER_NEEDS_MORE_INPUT
		}

		if bits == 0 {
			*value = 0
			return BROTLI_DECODER_SUCCESS
		}
		fallthrough

		/* Fall through. */
	case BROTLI_STATE_DECODE_UINT8_SHORT:
		if !safeReadBits(br, 3, &bits) {
			s.substate_decode_uint8 = BROTLI_STATE_DECODE_UINT8_SHORT
			return BROTLI_DECODER_NEEDS_MORE_INPUT
		}

		if bits == 0 {
			*value = 1
			s.substate_decode_uint8 = BROTLI_STATE_DECODE_UINT8_NONE
			return BROTLI_DECODER_SUCCESS
		}

		/* Use output value as a temporary storage. It MUST be persisted. */
		*value = bits
		fallthrough

		/* Fall through. */
	case BROTLI_STATE_DECODE_UINT8_LONG:
		if !safeReadBits(br, *value, &bits) {
			s.substate_decode_uint8 = BROTLI_STATE_DECODE_UINT8_LONG
			return BROTLI_DECODER_NEEDS_MORE_INPUT
		}

		*value = (1 << *value) + bits
		s.substate_decode_uint8 = BROTLI_STATE_DECODE_UINT8_NONE
		return BROTLI_DECODER_SUCCESS

	default:
		return BROTLI_DECODER_ERROR_UNREACHABLE
	}
}

/* Decodes a metablock length and flags by reading 2 - 31 bits. */
func DecodeMetaBlockLength(s *Reader, br *bitReader) int {
	var bits uint32
	var i int
	for {
		switch s.substate_metablock_header {
		case BROTLI_STATE_METABLOCK_HEADER_NONE:
			if !safeReadBits(br, 1, &bits) {
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}

			if bits != 0 {
				s.is_last_metablock = 1
			} else {
				s.is_last_metablock = 0
			}
			s.meta_block_remaining_len = 0
			s.is_uncompressed = 0
			s.is_metadata = 0
			if s.is_last_metablock == 0 {
				s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_NIBBLES
				break
			}

			s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_EMPTY
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_METABLOCK_HEADER_EMPTY:
			if !safeReadBits(br, 1, &bits) {
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}

			if bits != 0 {
				s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_NONE
				return BROTLI_DECODER_SUCCESS
			}

			s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_NIBBLES
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_METABLOCK_HEADER_NIBBLES:
			if !safeReadBits(br, 2, &bits) {
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}

			s.size_nibbles = uint(byte(bits + 4))
			s.loop_counter = 0
			if bits == 3 {
				s.is_metadata = 1
				s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_RESERVED
				break
			}

			s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_SIZE
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_METABLOCK_HEADER_SIZE:
			i = s.loop_counter

			for ; i < int(s.size_nibbles); i++ {
				if !safeReadBits(br, 4, &bits) {
					s.loop_counter = i
					return BROTLI_DECODER_NEEDS_MORE_INPUT
				}

				if uint(i+1) == s.size_nibbles && s.size_nibbles > 4 && bits == 0 {
					return BROTLI_DECODER_ERROR_FORMAT_EXUBERANT_NIBBLE
				}

				s.meta_block_remaining_len |= int(bits << uint(i*4))
			}

			s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_UNCOMPRESSED
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_METABLOCK_HEADER_UNCOMPRESSED:
			if s.is_last_metablock == 0 {
				if !safeReadBits(br, 1, &bits) {
					return BROTLI_DECODER_NEEDS_MORE_INPUT
				}

				if bits != 0 {
					s.is_uncompressed = 1
				} else {
					s.is_uncompressed = 0
				}
			}

			s.meta_block_remaining_len++
			s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_NONE
			return BROTLI_DECODER_SUCCESS

		case BROTLI_STATE_METABLOCK_HEADER_RESERVED:
			if !safeReadBits(br, 1, &bits) {
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}

			if bits != 0 {
				return BROTLI_DECODER_ERROR_FORMAT_RESERVED
			}

			s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_BYTES
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_METABLOCK_HEADER_BYTES:
			if !safeReadBits(br, 2, &bits) {
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}

			if bits == 0 {
				s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_NONE
				return BROTLI_DECODER_SUCCESS
			}

			s.size_nibbles = uint(byte(bits))
			s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_METADATA
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_METABLOCK_HEADER_METADATA:
			i = s.loop_counter

			for ; i < int(s.size_nibbles); i++ {
				if !safeReadBits(br, 8, &bits) {
					s.loop_counter = i
					return BROTLI_DECODER_NEEDS_MORE_INPUT
				}

				if uint(i+1) == s.size_nibbles && s.size_nibbles > 1 && bits == 0 {
					return BROTLI_DECODER_ERROR_FORMAT_EXUBERANT_META_NIBBLE
				}

				s.meta_block_remaining_len |= int(bits << uint(i*8))
			}

			s.meta_block_remaining_len++
			s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_NONE
			return BROTLI_DECODER_SUCCESS

		default:
			return BROTLI_DECODER_ERROR_UNREACHABLE
		}
	}
}

/* Decodes the Huffman code.
   This method doesn't read data from the bit reader, BUT drops the amount of
   bits that correspond to the decoded symbol.
   bits MUST contain at least 15 (BROTLI_HUFFMAN_MAX_CODE_LENGTH) valid bits. */
func DecodeSymbol(bits uint32, table []HuffmanCode, br *bitReader) uint32 {
	table = table[bits&HUFFMAN_TABLE_MASK:]
	if table[0].bits > HUFFMAN_TABLE_BITS {
		var nbits uint32 = uint32(table[0].bits) - HUFFMAN_TABLE_BITS
		dropBits(br, HUFFMAN_TABLE_BITS)
		table = table[uint32(table[0].value)+((bits>>HUFFMAN_TABLE_BITS)&bitMask(nbits)):]
	}

	dropBits(br, uint32(table[0].bits))
	return uint32(table[0].value)
}

/* Reads and decodes the next Huffman code from bit-stream.
   This method peeks 16 bits of input and drops 0 - 15 of them. */
func ReadSymbol(table []HuffmanCode, br *bitReader) uint32 {
	return DecodeSymbol(get16BitsUnmasked(br), table, br)
}

/* Same as DecodeSymbol, but it is known that there is less than 15 bits of
   input are currently available. */
func SafeDecodeSymbol(table []HuffmanCode, br *bitReader, result *uint32) bool {
	var val uint32
	var available_bits uint32 = getAvailableBits(br)
	if available_bits == 0 {
		if table[0].bits == 0 {
			*result = uint32(table[0].value)
			return true
		}

		return false /* No valid bits at all. */
	}

	val = uint32(getBitsUnmasked(br))
	table = table[val&HUFFMAN_TABLE_MASK:]
	if table[0].bits <= HUFFMAN_TABLE_BITS {
		if uint32(table[0].bits) <= available_bits {
			dropBits(br, uint32(table[0].bits))
			*result = uint32(table[0].value)
			return true
		} else {
			return false /* Not enough bits for the first level. */
		}
	}

	if available_bits <= HUFFMAN_TABLE_BITS {
		return false /* Not enough bits to move to the second level. */
	}

	/* Speculatively drop HUFFMAN_TABLE_BITS. */
	val = (val & bitMask(uint32(table[0].bits))) >> HUFFMAN_TABLE_BITS

	available_bits -= HUFFMAN_TABLE_BITS
	table = table[uint32(table[0].value)+val:]
	if available_bits < uint32(table[0].bits) {
		return false /* Not enough bits for the second level. */
	}

	dropBits(br, HUFFMAN_TABLE_BITS+uint32(table[0].bits))
	*result = uint32(table[0].value)
	return true
}

func SafeReadSymbol(table []HuffmanCode, br *bitReader, result *uint32) bool {
	var val uint32
	if safeGetBits(br, 15, &val) {
		*result = DecodeSymbol(val, table, br)
		return true
	}

	return SafeDecodeSymbol(table, br, result)
}

/* Makes a look-up in first level Huffman table. Peeks 8 bits. */
func PreloadSymbol(safe int, table []HuffmanCode, br *bitReader, bits *uint32, value *uint32) {
	if safe != 0 {
		return
	}

	table = table[BrotliGetBits(br, HUFFMAN_TABLE_BITS):]
	*bits = uint32(table[0].bits)
	*value = uint32(table[0].value)
}

/* Decodes the next Huffman code using data prepared by PreloadSymbol.
   Reads 0 - 15 bits. Also peeks 8 following bits. */
func ReadPreloadedSymbol(table []HuffmanCode, br *bitReader, bits *uint32, value *uint32) uint32 {
	var result uint32 = *value
	var ext []HuffmanCode
	if *bits > HUFFMAN_TABLE_BITS {
		var val uint32 = get16BitsUnmasked(br)
		ext = table[val&HUFFMAN_TABLE_MASK:][*value:]
		var mask uint32 = bitMask((*bits - HUFFMAN_TABLE_BITS))
		dropBits(br, HUFFMAN_TABLE_BITS)
		ext = ext[(val>>HUFFMAN_TABLE_BITS)&mask:]
		dropBits(br, uint32(ext[0].bits))
		result = uint32(ext[0].value)
	} else {
		dropBits(br, *bits)
	}

	PreloadSymbol(0, table, br, bits, value)
	return result
}

func Log2Floor(x uint32) uint32 {
	var result uint32 = 0
	for x != 0 {
		x >>= 1
		result++
	}

	return result
}

/* Reads (s->symbol + 1) symbols.
   Totally 1..4 symbols are read, 1..11 bits each.
   The list of symbols MUST NOT contain duplicates. */
func ReadSimpleHuffmanSymbols(alphabet_size uint32, max_symbol uint32, s *Reader) int {
	var br *bitReader = &s.br
	var max_bits uint32 = Log2Floor(alphabet_size - 1)
	var i uint32 = s.sub_loop_counter
	/* max_bits == 1..11; symbol == 0..3; 1..44 bits will be read. */

	var num_symbols uint32 = s.symbol
	for i <= num_symbols {
		var v uint32
		if !safeReadBits(br, max_bits, &v) {
			s.sub_loop_counter = i
			s.substate_huffman = BROTLI_STATE_HUFFMAN_SIMPLE_READ
			return BROTLI_DECODER_NEEDS_MORE_INPUT
		}

		if v >= max_symbol {
			return BROTLI_DECODER_ERROR_FORMAT_SIMPLE_HUFFMAN_ALPHABET
		}

		s.symbols_lists_array[i] = uint16(v)
		i++
	}

	for i = 0; i < num_symbols; i++ {
		var k uint32 = i + 1
		for ; k <= num_symbols; k++ {
			if s.symbols_lists_array[i] == s.symbols_lists_array[k] {
				return BROTLI_DECODER_ERROR_FORMAT_SIMPLE_HUFFMAN_SAME
			}
		}
	}

	return BROTLI_DECODER_SUCCESS
}

/* Process single decoded symbol code length:
   A) reset the repeat variable
   B) remember code length (if it is not 0)
   C) extend corresponding index-chain
   D) reduce the Huffman space
   E) update the histogram */
func ProcessSingleCodeLength(code_len uint32, symbol *uint32, repeat *uint32, space *uint32, prev_code_len *uint32, symbol_lists SymbolList, code_length_histo []uint16, next_symbol []int) {
	*repeat = 0
	if code_len != 0 { /* code_len == 1..15 */
		SymbolListPut(symbol_lists, next_symbol[code_len], uint16(*symbol))
		next_symbol[code_len] = int(*symbol)
		*prev_code_len = code_len
		*space -= 32768 >> code_len
		code_length_histo[code_len]++
	}

	(*symbol)++
}

/* Process repeated symbol code length.
    A) Check if it is the extension of previous repeat sequence; if the decoded
       value is not BROTLI_REPEAT_PREVIOUS_CODE_LENGTH, then it is a new
       symbol-skip
    B) Update repeat variable
    C) Check if operation is feasible (fits alphabet)
    D) For each symbol do the same operations as in ProcessSingleCodeLength

   PRECONDITION: code_len == BROTLI_REPEAT_PREVIOUS_CODE_LENGTH or
                 code_len == BROTLI_REPEAT_ZERO_CODE_LENGTH */
func ProcessRepeatedCodeLength(code_len uint32, repeat_delta uint32, alphabet_size uint32, symbol *uint32, repeat *uint32, space *uint32, prev_code_len *uint32, repeat_code_len *uint32, symbol_lists SymbolList, code_length_histo []uint16, next_symbol []int) {
	var old_repeat uint32 /* for BROTLI_REPEAT_ZERO_CODE_LENGTH */ /* for BROTLI_REPEAT_ZERO_CODE_LENGTH */
	var extra_bits uint32 = 3
	var new_len uint32 = 0
	if code_len == repeatPreviousCodeLength {
		new_len = *prev_code_len
		extra_bits = 2
	}

	if *repeat_code_len != new_len {
		*repeat = 0
		*repeat_code_len = new_len
	}

	old_repeat = *repeat
	if *repeat > 0 {
		*repeat -= 2
		*repeat <<= extra_bits
	}

	*repeat += repeat_delta + 3
	repeat_delta = *repeat - old_repeat
	if *symbol+repeat_delta > alphabet_size {
		*symbol = alphabet_size
		*space = 0xFFFFF
		return
	}

	if *repeat_code_len != 0 {
		var last uint = uint(*symbol + repeat_delta)
		var next int = next_symbol[*repeat_code_len]
		for {
			SymbolListPut(symbol_lists, next, uint16(*symbol))
			next = int(*symbol)
			(*symbol)++
			if (*symbol) == uint32(last) {
				break
			}
		}

		next_symbol[*repeat_code_len] = next
		*space -= repeat_delta << (15 - *repeat_code_len)
		code_length_histo[*repeat_code_len] = uint16(uint32(code_length_histo[*repeat_code_len]) + repeat_delta)
	} else {
		*symbol += repeat_delta
	}
}

/* Reads and decodes symbol codelengths. */
func ReadSymbolCodeLengths(alphabet_size uint32, s *Reader) int {
	var br *bitReader = &s.br
	var symbol uint32 = s.symbol
	var repeat uint32 = s.repeat
	var space uint32 = s.space
	var prev_code_len uint32 = s.prev_code_len
	var repeat_code_len uint32 = s.repeat_code_len
	var symbol_lists SymbolList = s.symbol_lists
	var code_length_histo []uint16 = s.code_length_histo[:]
	var next_symbol []int = s.next_symbol[:]
	if !warmupBitReader(br) {
		return BROTLI_DECODER_NEEDS_MORE_INPUT
	}
	var p []HuffmanCode
	for symbol < alphabet_size && space > 0 {
		p = s.table[:]
		var code_len uint32
		if !checkInputAmount(br, shortFillBitWindowRead) {
			s.symbol = symbol
			s.repeat = repeat
			s.prev_code_len = prev_code_len
			s.repeat_code_len = repeat_code_len
			s.space = space
			return BROTLI_DECODER_NEEDS_MORE_INPUT
		}

		fillBitWindow16(br)
		p = p[getBitsUnmasked(br)&uint64(bitMask(BROTLI_HUFFMAN_MAX_CODE_LENGTH_CODE_LENGTH)):]
		dropBits(br, uint32(p[0].bits)) /* Use 1..5 bits. */
		code_len = uint32(p[0].value)   /* code_len == 0..17 */
		if code_len < repeatPreviousCodeLength {
			ProcessSingleCodeLength(code_len, &symbol, &repeat, &space, &prev_code_len, symbol_lists, code_length_histo, next_symbol) /* code_len == 16..17, extra_bits == 2..3 */
		} else {
			var extra_bits uint32
			if code_len == repeatPreviousCodeLength {
				extra_bits = 2
			} else {
				extra_bits = 3
			}
			var repeat_delta uint32 = uint32(getBitsUnmasked(br)) & bitMask(extra_bits)
			dropBits(br, extra_bits)
			ProcessRepeatedCodeLength(code_len, repeat_delta, alphabet_size, &symbol, &repeat, &space, &prev_code_len, &repeat_code_len, symbol_lists, code_length_histo, next_symbol)
		}
	}

	s.space = space
	return BROTLI_DECODER_SUCCESS
}

func SafeReadSymbolCodeLengths(alphabet_size uint32, s *Reader) int {
	var br *bitReader = &s.br
	var get_byte bool = false
	var p []HuffmanCode
	for s.symbol < alphabet_size && s.space > 0 {
		p = s.table[:]
		var code_len uint32
		var available_bits uint32
		var bits uint32 = 0
		if get_byte && !pullByte(br) {
			return BROTLI_DECODER_NEEDS_MORE_INPUT
		}
		get_byte = false
		available_bits = getAvailableBits(br)
		if available_bits != 0 {
			bits = uint32(getBitsUnmasked(br))
		}

		p = p[bits&bitMask(BROTLI_HUFFMAN_MAX_CODE_LENGTH_CODE_LENGTH):]
		if uint32(p[0].bits) > available_bits {
			get_byte = true
			continue
		}

		code_len = uint32(p[0].value) /* code_len == 0..17 */
		if code_len < repeatPreviousCodeLength {
			dropBits(br, uint32(p[0].bits))
			ProcessSingleCodeLength(code_len, &s.symbol, &s.repeat, &s.space, &s.prev_code_len, s.symbol_lists, s.code_length_histo[:], s.next_symbol[:]) /* code_len == 16..17, extra_bits == 2..3 */
		} else {
			var extra_bits uint32 = code_len - 14
			var repeat_delta uint32 = (bits >> p[0].bits) & bitMask(extra_bits)
			if available_bits < uint32(p[0].bits)+extra_bits {
				get_byte = true
				continue
			}

			dropBits(br, uint32(p[0].bits)+extra_bits)
			ProcessRepeatedCodeLength(code_len, repeat_delta, alphabet_size, &s.symbol, &s.repeat, &s.space, &s.prev_code_len, &s.repeat_code_len, s.symbol_lists, s.code_length_histo[:], s.next_symbol[:])
		}
	}

	return BROTLI_DECODER_SUCCESS
}

/* Reads and decodes 15..18 codes using static prefix code.
   Each code is 2..4 bits long. In total 30..72 bits are used. */
func ReadCodeLengthCodeLengths(s *Reader) int {
	var br *bitReader = &s.br
	var num_codes uint32 = s.repeat
	var space uint32 = s.space
	var i uint32 = s.sub_loop_counter
	for ; i < codeLengthCodes; i++ {
		var code_len_idx byte = kCodeLengthCodeOrder[i]
		var ix uint32
		var v uint32
		if !safeGetBits(br, 4, &ix) {
			var available_bits uint32 = getAvailableBits(br)
			if available_bits != 0 {
				ix = uint32(getBitsUnmasked(br) & 0xF)
			} else {
				ix = 0
			}

			if uint32(kCodeLengthPrefixLength[ix]) > available_bits {
				s.sub_loop_counter = i
				s.repeat = num_codes
				s.space = space
				s.substate_huffman = BROTLI_STATE_HUFFMAN_COMPLEX
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}
		}

		v = uint32(kCodeLengthPrefixValue[ix])
		dropBits(br, uint32(kCodeLengthPrefixLength[ix]))
		s.code_length_code_lengths[code_len_idx] = byte(v)
		if v != 0 {
			space = space - (32 >> v)
			num_codes++
			s.code_length_histo[v]++
			if space-1 >= 32 {
				/* space is 0 or wrapped around. */
				break
			}
		}
	}

	if num_codes != 1 && space != 0 {
		return BROTLI_DECODER_ERROR_FORMAT_CL_SPACE
	}

	return BROTLI_DECODER_SUCCESS
}

/* Decodes the Huffman tables.
   There are 2 scenarios:
    A) Huffman code contains only few symbols (1..4). Those symbols are read
       directly; their code lengths are defined by the number of symbols.
       For this scenario 4 - 49 bits will be read.

    B) 2-phase decoding:
    B.1) Small Huffman table is decoded; it is specified with code lengths
         encoded with predefined entropy code. 32 - 74 bits are used.
    B.2) Decoded table is used to decode code lengths of symbols in resulting
         Huffman table. In worst case 3520 bits are read. */
func ReadHuffmanCode(alphabet_size uint32, max_symbol uint32, table []HuffmanCode, opt_table_size *uint32, s *Reader) int {
	var br *bitReader = &s.br

	/* Unnecessary masking, but might be good for safety. */
	alphabet_size &= 0x7FF

	/* State machine. */
	for {
		switch s.substate_huffman {
		case BROTLI_STATE_HUFFMAN_NONE:
			if !safeReadBits(br, 2, &s.sub_loop_counter) {
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}

			/* The value is used as follows:
			   1 for simple code;
			   0 for no skipping, 2 skips 2 code lengths, 3 skips 3 code lengths */
			if s.sub_loop_counter != 1 {
				s.space = 32
				s.repeat = 0 /* num_codes */
				var i int
				for i = 0; i <= BROTLI_HUFFMAN_MAX_CODE_LENGTH_CODE_LENGTH; i++ {
					s.code_length_histo[i] = 0
				}

				for i = 0; i < codeLengthCodes; i++ {
					s.code_length_code_lengths[i] = 0
				}

				s.substate_huffman = BROTLI_STATE_HUFFMAN_COMPLEX
				continue
			}
			fallthrough

			/* Read symbols, codes & code lengths directly. */
		/* Fall through. */
		case BROTLI_STATE_HUFFMAN_SIMPLE_SIZE:
			if !safeReadBits(br, 2, &s.symbol) { /* num_symbols */
				s.substate_huffman = BROTLI_STATE_HUFFMAN_SIMPLE_SIZE
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}

			s.sub_loop_counter = 0
			fallthrough
		/* Fall through. */
		case BROTLI_STATE_HUFFMAN_SIMPLE_READ:
			{
				var result int = ReadSimpleHuffmanSymbols(alphabet_size, max_symbol, s)
				if result != BROTLI_DECODER_SUCCESS {
					return result
				}
			}
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_HUFFMAN_SIMPLE_BUILD:
			{
				var table_size uint32
				if s.symbol == 3 {
					var bits uint32
					if !safeReadBits(br, 1, &bits) {
						s.substate_huffman = BROTLI_STATE_HUFFMAN_SIMPLE_BUILD
						return BROTLI_DECODER_NEEDS_MORE_INPUT
					}

					s.symbol += bits
				}

				table_size = BrotliBuildSimpleHuffmanTable(table, HUFFMAN_TABLE_BITS, s.symbols_lists_array[:], s.symbol)
				if opt_table_size != nil {
					*opt_table_size = table_size
				}

				s.substate_huffman = BROTLI_STATE_HUFFMAN_NONE
				return BROTLI_DECODER_SUCCESS
			}
			fallthrough

			/* Decode Huffman-coded code lengths. */
		case BROTLI_STATE_HUFFMAN_COMPLEX:
			{
				var i uint32
				var result int = ReadCodeLengthCodeLengths(s)
				if result != BROTLI_DECODER_SUCCESS {
					return result
				}

				BrotliBuildCodeLengthsHuffmanTable(s.table[:], s.code_length_code_lengths[:], s.code_length_histo[:])
				for i = 0; i < 16; i++ {
					s.code_length_histo[i] = 0
				}

				for i = 0; i <= BROTLI_HUFFMAN_MAX_CODE_LENGTH; i++ {
					s.next_symbol[i] = int(i) - (BROTLI_HUFFMAN_MAX_CODE_LENGTH + 1)
					SymbolListPut(s.symbol_lists, s.next_symbol[i], 0xFFFF)
				}

				s.symbol = 0
				s.prev_code_len = initialRepeatedCodeLength
				s.repeat = 0
				s.repeat_code_len = 0
				s.space = 32768
				s.substate_huffman = BROTLI_STATE_HUFFMAN_LENGTH_SYMBOLS
			}
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_HUFFMAN_LENGTH_SYMBOLS:
			{
				var table_size uint32
				var result int = ReadSymbolCodeLengths(max_symbol, s)
				if result == BROTLI_DECODER_NEEDS_MORE_INPUT {
					result = SafeReadSymbolCodeLengths(max_symbol, s)
				}

				if result != BROTLI_DECODER_SUCCESS {
					return result
				}

				if s.space != 0 {
					return BROTLI_DECODER_ERROR_FORMAT_HUFFMAN_SPACE
				}

				table_size = BrotliBuildHuffmanTable(table, HUFFMAN_TABLE_BITS, s.symbol_lists, s.code_length_histo[:])
				if opt_table_size != nil {
					*opt_table_size = table_size
				}

				s.substate_huffman = BROTLI_STATE_HUFFMAN_NONE
				return BROTLI_DECODER_SUCCESS
			}
			fallthrough

		default:
			return BROTLI_DECODER_ERROR_UNREACHABLE
		}
	}
}

/* Decodes a block length by reading 3..39 bits. */
func ReadBlockLength(table []HuffmanCode, br *bitReader) uint32 {
	var code uint32
	var nbits uint32
	code = ReadSymbol(table, br)
	nbits = uint32(kBlockLengthPrefixCode1[code].nbits) /* nbits == 2..24 */
	return uint32(kBlockLengthPrefixCode1[code].offset) + readBits(br, nbits)
}

/* WARNING: if state is not BROTLI_STATE_READ_BLOCK_LENGTH_NONE, then
   reading can't be continued with ReadBlockLength. */
func SafeReadBlockLength(s *Reader, result *uint32, table []HuffmanCode, br *bitReader) bool {
	var index uint32
	if s.substate_read_block_length == BROTLI_STATE_READ_BLOCK_LENGTH_NONE {
		if !SafeReadSymbol(table, br, &index) {
			return false
		}
	} else {
		index = s.block_length_index
	}
	{
		var bits uint32 /* nbits == 2..24 */
		var nbits uint32 = uint32(kBlockLengthPrefixCode1[index].nbits)
		if !safeReadBits(br, nbits, &bits) {
			s.block_length_index = index
			s.substate_read_block_length = BROTLI_STATE_READ_BLOCK_LENGTH_SUFFIX
			return false
		}

		*result = uint32(kBlockLengthPrefixCode1[index].offset) + bits
		s.substate_read_block_length = BROTLI_STATE_READ_BLOCK_LENGTH_NONE
		return true
	}
}

/* Transform:
    1) initialize list L with values 0, 1,... 255
    2) For each input element X:
    2.1) let Y = L[X]
    2.2) remove X-th element from L
    2.3) prepend Y to L
    2.4) append Y to output

   In most cases max(Y) <= 7, so most of L remains intact.
   To reduce the cost of initialization, we reuse L, remember the upper bound
   of Y values, and reinitialize only first elements in L.

   Most of input values are 0 and 1. To reduce number of branches, we replace
   inner for loop with do-while. */
func InverseMoveToFrontTransform(v []byte, v_len uint32, state *Reader) {
	var mtf [256]byte
	var i int
	for i = 1; i < 256; i++ {
		mtf[i] = byte(i)
	}
	var mtf_1 byte

	/* Transform the input. */
	for i = 0; uint32(i) < v_len; i++ {
		var index int = int(v[i])
		var value byte = mtf[index]
		v[i] = value
		mtf_1 = value
		for index >= 1 {
			index--
			mtf[index+1] = mtf[index]
		}

		mtf[0] = mtf_1
	}
}

/* Decodes a series of Huffman table using ReadHuffmanCode function. */
func HuffmanTreeGroupDecode(group *HuffmanTreeGroup, s *Reader) int {
	if s.substate_tree_group != BROTLI_STATE_TREE_GROUP_LOOP {
		s.next = group.codes
		s.htree_index = 0
		s.substate_tree_group = BROTLI_STATE_TREE_GROUP_LOOP
	}

	for s.htree_index < int(group.num_htrees) {
		var table_size uint32
		var result int = ReadHuffmanCode(uint32(group.alphabet_size), uint32(group.max_symbol), s.next, &table_size, s)
		if result != BROTLI_DECODER_SUCCESS {
			return result
		}
		group.htrees[s.htree_index] = s.next
		s.next = s.next[table_size:]
		s.htree_index++
	}

	s.substate_tree_group = BROTLI_STATE_TREE_GROUP_NONE
	return BROTLI_DECODER_SUCCESS
}

/* Decodes a context map.
   Decoding is done in 4 phases:
    1) Read auxiliary information (6..16 bits) and allocate memory.
       In case of trivial context map, decoding is finished at this phase.
    2) Decode Huffman table using ReadHuffmanCode function.
       This table will be used for reading context map items.
    3) Read context map items; "0" values could be run-length encoded.
    4) Optionally, apply InverseMoveToFront transform to the resulting map. */
func DecodeContextMap(context_map_size uint32, num_htrees *uint32, context_map_arg *[]byte, s *Reader) int {
	var br *bitReader = &s.br
	var result int = BROTLI_DECODER_SUCCESS

	switch int(s.substate_context_map) {
	case BROTLI_STATE_CONTEXT_MAP_NONE:
		result = DecodeVarLenUint8(s, br, num_htrees)
		if result != BROTLI_DECODER_SUCCESS {
			return result
		}

		(*num_htrees)++
		s.context_index = 0
		*context_map_arg = make([]byte, uint(context_map_size))
		if *context_map_arg == nil {
			return BROTLI_DECODER_ERROR_ALLOC_CONTEXT_MAP
		}

		if *num_htrees <= 1 {
			for i := 0; i < int(context_map_size); i++ {
				(*context_map_arg)[i] = 0
			}
			return BROTLI_DECODER_SUCCESS
		}

		s.substate_context_map = BROTLI_STATE_CONTEXT_MAP_READ_PREFIX
		fallthrough
	/* Fall through. */
	case BROTLI_STATE_CONTEXT_MAP_READ_PREFIX:
		{
			var bits uint32

			/* In next stage ReadHuffmanCode uses at least 4 bits, so it is safe
			   to peek 4 bits ahead. */
			if !safeGetBits(br, 5, &bits) {
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}

			if bits&1 != 0 { /* Use RLE for zeros. */
				s.max_run_length_prefix = (bits >> 1) + 1
				dropBits(br, 5)
			} else {
				s.max_run_length_prefix = 0
				dropBits(br, 1)
			}

			s.substate_context_map = BROTLI_STATE_CONTEXT_MAP_HUFFMAN
		}
		fallthrough

		/* Fall through. */
	case BROTLI_STATE_CONTEXT_MAP_HUFFMAN:
		{
			var alphabet_size uint32 = *num_htrees + s.max_run_length_prefix
			result = ReadHuffmanCode(alphabet_size, alphabet_size, s.context_map_table[:], nil, s)
			if result != BROTLI_DECODER_SUCCESS {
				return result
			}
			s.code = 0xFFFF
			s.substate_context_map = BROTLI_STATE_CONTEXT_MAP_DECODE
		}
		fallthrough

		/* Fall through. */
	case BROTLI_STATE_CONTEXT_MAP_DECODE:
		{
			var context_index uint32 = s.context_index
			var max_run_length_prefix uint32 = s.max_run_length_prefix
			var context_map []byte = *context_map_arg
			var code uint32 = s.code
			var skip_preamble bool = (code != 0xFFFF)
			for context_index < context_map_size || skip_preamble {
				if !skip_preamble {
					if !SafeReadSymbol(s.context_map_table[:], br, &code) {
						s.code = 0xFFFF
						s.context_index = context_index
						return BROTLI_DECODER_NEEDS_MORE_INPUT
					}

					if code == 0 {
						context_map[context_index] = 0
						context_index++
						continue
					}

					if code > max_run_length_prefix {
						context_map[context_index] = byte(code - max_run_length_prefix)
						context_index++
						continue
					}
				} else {
					skip_preamble = false
				}

				/* RLE sub-stage. */
				{
					var reps uint32
					if !safeReadBits(br, code, &reps) {
						s.code = code
						s.context_index = context_index
						return BROTLI_DECODER_NEEDS_MORE_INPUT
					}

					reps += 1 << code
					if context_index+reps > context_map_size {
						return BROTLI_DECODER_ERROR_FORMAT_CONTEXT_MAP_REPEAT
					}

					for {
						context_map[context_index] = 0
						context_index++
						reps--
						if reps == 0 {
							break
						}
					}
				}
			}
		}
		fallthrough

		/* Fall through. */
	case BROTLI_STATE_CONTEXT_MAP_TRANSFORM:
		{
			var bits uint32
			if !safeReadBits(br, 1, &bits) {
				s.substate_context_map = BROTLI_STATE_CONTEXT_MAP_TRANSFORM
				return BROTLI_DECODER_NEEDS_MORE_INPUT
			}

			if bits != 0 {
				InverseMoveToFrontTransform(*context_map_arg, context_map_size, s)
			}

			s.substate_context_map = BROTLI_STATE_CONTEXT_MAP_NONE
			return BROTLI_DECODER_SUCCESS
		}
		fallthrough

	default:
		return BROTLI_DECODER_ERROR_UNREACHABLE
	}
}

/* Decodes a command or literal and updates block type ring-buffer.
   Reads 3..54 bits. */
func DecodeBlockTypeAndLength(safe int, s *Reader, tree_type int) bool {
	var max_block_type uint32 = s.num_block_types[tree_type]
	var type_tree []HuffmanCode
	type_tree = s.block_type_trees[tree_type*BROTLI_HUFFMAN_MAX_SIZE_258:]
	var len_tree []HuffmanCode
	len_tree = s.block_len_trees[tree_type*BROTLI_HUFFMAN_MAX_SIZE_26:]
	var br *bitReader = &s.br
	var ringbuffer []uint32 = s.block_type_rb[tree_type*2:]
	var block_type uint32
	if max_block_type <= 1 {
		return false
	}

	/* Read 0..15 + 3..39 bits. */
	if safe == 0 {
		block_type = ReadSymbol(type_tree, br)
		s.block_length[tree_type] = ReadBlockLength(len_tree, br)
	} else {
		var memento bitReaderState
		bitReaderSaveState(br, &memento)
		if !SafeReadSymbol(type_tree, br, &block_type) {
			return false
		}
		if !SafeReadBlockLength(s, &s.block_length[tree_type], len_tree, br) {
			s.substate_read_block_length = BROTLI_STATE_READ_BLOCK_LENGTH_NONE
			bitReaderRestoreState(br, &memento)
			return false
		}
	}

	if block_type == 1 {
		block_type = ringbuffer[1] + 1
	} else if block_type == 0 {
		block_type = ringbuffer[0]
	} else {
		block_type -= 2
	}

	if block_type >= max_block_type {
		block_type -= max_block_type
	}

	ringbuffer[0] = ringbuffer[1]
	ringbuffer[1] = block_type
	return true
}

func DetectTrivialLiteralBlockTypes(s *Reader) {
	var i uint
	for i = 0; i < 8; i++ {
		s.trivial_literal_contexts[i] = 0
	}
	for i = 0; uint32(i) < s.num_block_types[0]; i++ {
		var offset uint = i << BROTLI_LITERAL_CONTEXT_BITS
		var error uint = 0
		var sample uint = uint(s.context_map[offset])
		var j uint
		for j = 0; j < 1<<BROTLI_LITERAL_CONTEXT_BITS; {
			var k int
			for k = 0; k < 4; k++ {
				error |= uint(s.context_map[offset+j]) ^ sample
				j++
			}
		}

		if error == 0 {
			s.trivial_literal_contexts[i>>5] |= 1 << (i & 31)
		}
	}
}

func PrepareLiteralDecoding(s *Reader) {
	var context_mode byte
	var trivial uint
	var block_type uint32 = s.block_type_rb[1]
	var context_offset uint32 = block_type << BROTLI_LITERAL_CONTEXT_BITS
	s.context_map_slice = s.context_map[context_offset:]
	trivial = uint(s.trivial_literal_contexts[block_type>>5])
	s.trivial_literal_context = int((trivial >> (block_type & 31)) & 1)
	s.literal_htree = []HuffmanCode(s.literal_hgroup.htrees[s.context_map_slice[0]])
	context_mode = s.context_modes[block_type] & 3
	s.context_lookup = BROTLI_CONTEXT_LUT(int(context_mode))
}

/* Decodes the block type and updates the state for literal context.
   Reads 3..54 bits. */
func DecodeLiteralBlockSwitchInternal(safe int, s *Reader) bool {
	if !DecodeBlockTypeAndLength(safe, s, 0) {
		return false
	}

	PrepareLiteralDecoding(s)
	return true
}

func DecodeLiteralBlockSwitch(s *Reader) {
	DecodeLiteralBlockSwitchInternal(0, s)
}

func SafeDecodeLiteralBlockSwitch(s *Reader) bool {
	return DecodeLiteralBlockSwitchInternal(1, s)
}

/* Block switch for insert/copy length.
   Reads 3..54 bits. */
func DecodeCommandBlockSwitchInternal(safe int, s *Reader) bool {
	if !DecodeBlockTypeAndLength(safe, s, 1) {
		return false
	}

	s.htree_command = []HuffmanCode(s.insert_copy_hgroup.htrees[s.block_type_rb[3]])
	return true
}

func DecodeCommandBlockSwitch(s *Reader) {
	DecodeCommandBlockSwitchInternal(0, s)
}

func SafeDecodeCommandBlockSwitch(s *Reader) bool {
	return DecodeCommandBlockSwitchInternal(1, s)
}

/* Block switch for distance codes.
   Reads 3..54 bits. */
func DecodeDistanceBlockSwitchInternal(safe int, s *Reader) bool {
	if !DecodeBlockTypeAndLength(safe, s, 2) {
		return false
	}

	s.dist_context_map_slice = s.dist_context_map[s.block_type_rb[5]<<BROTLI_DISTANCE_CONTEXT_BITS:]
	s.dist_htree_index = s.dist_context_map_slice[s.distance_context]
	return true
}

func DecodeDistanceBlockSwitch(s *Reader) {
	DecodeDistanceBlockSwitchInternal(0, s)
}

func SafeDecodeDistanceBlockSwitch(s *Reader) bool {
	return DecodeDistanceBlockSwitchInternal(1, s)
}

func UnwrittenBytes(s *Reader, wrap bool) uint {
	var pos uint
	if wrap && s.pos > s.ringbuffer_size {
		pos = uint(s.ringbuffer_size)
	} else {
		pos = uint(s.pos)
	}
	var partial_pos_rb uint = (s.rb_roundtrips * uint(s.ringbuffer_size)) + pos
	return partial_pos_rb - s.partial_pos_out
}

/* Dumps output.
   Returns BROTLI_DECODER_NEEDS_MORE_OUTPUT only if there is more output to push
   and either ring-buffer is as big as window size, or |force| is true. */
func WriteRingBuffer(s *Reader, available_out *uint, next_out *[]byte, total_out *uint, force bool) int {
	var start []byte
	start = s.ringbuffer[s.partial_pos_out&uint(s.ringbuffer_mask):]
	var to_write uint = UnwrittenBytes(s, true)
	var num_written uint = *available_out
	if num_written > to_write {
		num_written = to_write
	}

	if s.meta_block_remaining_len < 0 {
		return BROTLI_DECODER_ERROR_FORMAT_BLOCK_LENGTH_1
	}

	if next_out != nil && *next_out == nil {
		*next_out = start
	} else {
		if next_out != nil {
			copy(*next_out, start[:num_written])
			*next_out = (*next_out)[num_written:]
		}
	}

	*available_out -= num_written
	s.partial_pos_out += num_written
	if total_out != nil {
		*total_out = s.partial_pos_out
	}

	if num_written < to_write {
		if s.ringbuffer_size == 1<<s.window_bits || force {
			return BROTLI_DECODER_NEEDS_MORE_OUTPUT
		} else {
			return BROTLI_DECODER_SUCCESS
		}
	}

	/* Wrap ring buffer only if it has reached its maximal size. */
	if s.ringbuffer_size == 1<<s.window_bits && s.pos >= s.ringbuffer_size {
		s.pos -= s.ringbuffer_size
		s.rb_roundtrips++
		if uint(s.pos) != 0 {
			s.should_wrap_ringbuffer = 1
		} else {
			s.should_wrap_ringbuffer = 0
		}
	}

	return BROTLI_DECODER_SUCCESS
}

func WrapRingBuffer(s *Reader) {
	if s.should_wrap_ringbuffer != 0 {
		copy(s.ringbuffer, s.ringbuffer_end[:uint(s.pos)])
		s.should_wrap_ringbuffer = 0
	}
}

/* Allocates ring-buffer.

   s->ringbuffer_size MUST be updated by BrotliCalculateRingBufferSize before
   this function is called.

   Last two bytes of ring-buffer are initialized to 0, so context calculation
   could be done uniformly for the first two and all other positions. */
func BrotliEnsureRingBuffer(s *Reader) bool {
	var old_ringbuffer []byte = s.ringbuffer
	if s.ringbuffer_size == s.new_ringbuffer_size {
		return true
	}

	s.ringbuffer = make([]byte, uint(s.new_ringbuffer_size)+uint(kRingBufferWriteAheadSlack))
	if s.ringbuffer == nil {
		/* Restore previous value. */
		s.ringbuffer = old_ringbuffer

		return false
	}

	s.ringbuffer[s.new_ringbuffer_size-2] = 0
	s.ringbuffer[s.new_ringbuffer_size-1] = 0

	if !(old_ringbuffer == nil) {
		copy(s.ringbuffer, old_ringbuffer[:uint(s.pos)])

		old_ringbuffer = nil
	}

	s.ringbuffer_size = s.new_ringbuffer_size
	s.ringbuffer_mask = s.new_ringbuffer_size - 1
	s.ringbuffer_end = s.ringbuffer[s.ringbuffer_size:]

	return true
}

func CopyUncompressedBlockToOutput(available_out *uint, next_out *[]byte, total_out *uint, s *Reader) int {
	/* TODO: avoid allocation for single uncompressed block. */
	if !BrotliEnsureRingBuffer(s) {
		return BROTLI_DECODER_ERROR_ALLOC_RING_BUFFER_1
	}

	/* State machine */
	for {
		switch s.substate_uncompressed {
		case BROTLI_STATE_UNCOMPRESSED_NONE:
			{
				var nbytes int = int(getRemainingBytes(&s.br))
				if nbytes > s.meta_block_remaining_len {
					nbytes = s.meta_block_remaining_len
				}

				if s.pos+nbytes > s.ringbuffer_size {
					nbytes = s.ringbuffer_size - s.pos
				}

				/* Copy remaining bytes from s->br.buf_ to ring-buffer. */
				copyBytes(s.ringbuffer[s.pos:], &s.br, uint(nbytes))

				s.pos += nbytes
				s.meta_block_remaining_len -= nbytes
				if s.pos < 1<<s.window_bits {
					if s.meta_block_remaining_len == 0 {
						return BROTLI_DECODER_SUCCESS
					}

					return BROTLI_DECODER_NEEDS_MORE_INPUT
				}

				s.substate_uncompressed = BROTLI_STATE_UNCOMPRESSED_WRITE
			}
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_UNCOMPRESSED_WRITE:
			{
				var result int
				result = WriteRingBuffer(s, available_out, next_out, total_out, false)
				if result != BROTLI_DECODER_SUCCESS {
					return result
				}

				if s.ringbuffer_size == 1<<s.window_bits {
					s.max_distance = s.max_backward_distance
				}

				s.substate_uncompressed = BROTLI_STATE_UNCOMPRESSED_NONE
				break
			}
		}
	}

	assert(false) /* Unreachable */
	return 0
}

/* Calculates the smallest feasible ring buffer.

   If we know the data size is small, do not allocate more ring buffer
   size than needed to reduce memory usage.

   When this method is called, metablock size and flags MUST be decoded. */
func BrotliCalculateRingBufferSize(s *Reader) {
	var window_size int = 1 << s.window_bits
	var new_ringbuffer_size int = window_size
	var min_size int
	/* We need at least 2 bytes of ring buffer size to get the last two
	   bytes for context from there */
	if s.ringbuffer_size != 0 {
		min_size = s.ringbuffer_size
	} else {
		min_size = 1024
	}
	var output_size int

	/* If maximum is already reached, no further extension is retired. */
	if s.ringbuffer_size == window_size {
		return
	}

	/* Metadata blocks does not touch ring buffer. */
	if s.is_metadata != 0 {
		return
	}

	if s.ringbuffer == nil {
		output_size = 0
	} else {
		output_size = s.pos
	}

	output_size += s.meta_block_remaining_len
	if min_size < output_size {
		min_size = output_size
	} else {
		min_size = min_size
	}

	if !(s.canny_ringbuffer_allocation == 0) {
		/* Reduce ring buffer size to save memory when server is unscrupulous.
		   In worst case memory usage might be 1.5x bigger for a short period of
		   ring buffer reallocation. */
		for new_ringbuffer_size>>1 >= min_size {
			new_ringbuffer_size >>= 1
		}
	}

	s.new_ringbuffer_size = new_ringbuffer_size
}

/* Reads 1..256 2-bit context modes. */
func ReadContextModes(s *Reader) int {
	var br *bitReader = &s.br
	var i int = s.loop_counter

	for i < int(s.num_block_types[0]) {
		var bits uint32
		if !safeReadBits(br, 2, &bits) {
			s.loop_counter = i
			return BROTLI_DECODER_NEEDS_MORE_INPUT
		}

		s.context_modes[i] = byte(bits)
		i++
	}

	return BROTLI_DECODER_SUCCESS
}

func TakeDistanceFromRingBuffer(s *Reader) {
	if s.distance_code == 0 {
		s.dist_rb_idx--
		s.distance_code = s.dist_rb[s.dist_rb_idx&3]

		/* Compensate double distance-ring-buffer roll for dictionary items. */
		s.distance_context = 1
	} else {
		var distance_code int = s.distance_code << 1
		var kDistanceShortCodeIndexOffset uint32 = 0xAAAFFF1B
		var kDistanceShortCodeValueOffset uint32 = 0xFA5FA500
		var v int = (s.dist_rb_idx + int(kDistanceShortCodeIndexOffset>>uint(distance_code))) & 0x3
		/* kDistanceShortCodeIndexOffset has 2-bit values from LSB:
		   3, 2, 1, 0, 3, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2 */

		/* kDistanceShortCodeValueOffset has 2-bit values from LSB:
		   -0, 0,-0, 0,-1, 1,-2, 2,-3, 3,-1, 1,-2, 2,-3, 3 */
		s.distance_code = s.dist_rb[v]

		v = int(kDistanceShortCodeValueOffset>>uint(distance_code)) & 0x3
		if distance_code&0x3 != 0 {
			s.distance_code += v
		} else {
			s.distance_code -= v
			if s.distance_code <= 0 {
				/* A huge distance will cause a () soon.
				   This is a little faster than failing here. */
				s.distance_code = 0x7FFFFFFF
			}
		}
	}
}

func SafeReadBits(br *bitReader, n_bits uint32, val *uint32) bool {
	if n_bits != 0 {
		return safeReadBits(br, n_bits, val)
	} else {
		*val = 0
		return true
	}
}

/* Precondition: s->distance_code < 0. */
func ReadDistanceInternal(safe int, s *Reader, br *bitReader) bool {
	var distval int
	var memento bitReaderState
	var distance_tree []HuffmanCode = []HuffmanCode(s.distance_hgroup.htrees[s.dist_htree_index])
	if safe == 0 {
		s.distance_code = int(ReadSymbol(distance_tree, br))
	} else {
		var code uint32
		bitReaderSaveState(br, &memento)
		if !SafeReadSymbol(distance_tree, br, &code) {
			return false
		}

		s.distance_code = int(code)
	}

	/* Convert the distance code to the actual distance by possibly
	   looking up past distances from the s->ringbuffer. */
	s.distance_context = 0

	if s.distance_code&^0xF == 0 {
		TakeDistanceFromRingBuffer(s)
		s.block_length[2]--
		return true
	}

	distval = s.distance_code - int(s.num_direct_distance_codes)
	if distval >= 0 {
		var nbits uint32
		var postfix int
		var offset int
		if safe == 0 && (s.distance_postfix_bits == 0) {
			nbits = (uint32(distval) >> 1) + 1
			offset = ((2 + (distval & 1)) << nbits) - 4
			s.distance_code = int(s.num_direct_distance_codes) + offset + int(readBits(br, nbits))
		} else {
			/* This branch also works well when s->distance_postfix_bits == 0. */
			var bits uint32
			postfix = distval & s.distance_postfix_mask
			distval >>= s.distance_postfix_bits
			nbits = (uint32(distval) >> 1) + 1
			if safe != 0 {
				if !SafeReadBits(br, nbits, &bits) {
					s.distance_code = -1 /* Restore precondition. */
					bitReaderRestoreState(br, &memento)
					return false
				}
			} else {
				bits = readBits(br, nbits)
			}

			offset = ((2 + (distval & 1)) << nbits) - 4
			s.distance_code = int(s.num_direct_distance_codes) + ((offset + int(bits)) << s.distance_postfix_bits) + postfix
		}
	}

	s.distance_code = s.distance_code - numDistanceShortCodes + 1
	s.block_length[2]--
	return true
}

func ReadDistance(s *Reader, br *bitReader) {
	ReadDistanceInternal(0, s, br)
}

func SafeReadDistance(s *Reader, br *bitReader) bool {
	return ReadDistanceInternal(1, s, br)
}

func ReadCommandInternal(safe int, s *Reader, br *bitReader, insert_length *int) bool {
	var cmd_code uint32
	var insert_len_extra uint32 = 0
	var copy_length uint32
	var v CmdLutElement
	var memento bitReaderState
	if safe == 0 {
		cmd_code = ReadSymbol(s.htree_command, br)
	} else {
		bitReaderSaveState(br, &memento)
		if !SafeReadSymbol(s.htree_command, br, &cmd_code) {
			return false
		}
	}

	v = kCmdLut[cmd_code]
	s.distance_code = int(v.distance_code)
	s.distance_context = int(v.context)
	s.dist_htree_index = s.dist_context_map_slice[s.distance_context]
	*insert_length = int(v.insert_len_offset)
	if safe == 0 {
		if v.insert_len_extra_bits != 0 {
			insert_len_extra = readBits(br, uint32(v.insert_len_extra_bits))
		}

		copy_length = readBits(br, uint32(v.copy_len_extra_bits))
	} else {
		if !SafeReadBits(br, uint32(v.insert_len_extra_bits), &insert_len_extra) || !SafeReadBits(br, uint32(v.copy_len_extra_bits), &copy_length) {
			bitReaderRestoreState(br, &memento)
			return false
		}
	}

	s.copy_length = int(copy_length) + int(v.copy_len_offset)
	s.block_length[1]--
	*insert_length += int(insert_len_extra)
	return true
}

func ReadCommand(s *Reader, br *bitReader, insert_length *int) {
	ReadCommandInternal(0, s, br, insert_length)
}

func SafeReadCommand(s *Reader, br *bitReader, insert_length *int) bool {
	return ReadCommandInternal(1, s, br, insert_length)
}

func CheckInputAmount(safe int, br *bitReader, num uint) bool {
	if safe != 0 {
		return true
	}

	return checkInputAmount(br, num)
}

func ProcessCommandsInternal(safe int, s *Reader) int {
	var pos int = s.pos
	var i int = s.loop_counter
	var result int = BROTLI_DECODER_SUCCESS
	var br *bitReader = &s.br
	var hc []HuffmanCode

	if !CheckInputAmount(safe, br, 28) {
		result = BROTLI_DECODER_NEEDS_MORE_INPUT
		goto saveStateAndReturn
	}

	if safe == 0 {
		warmupBitReader(br)
	}

	/* Jump into state machine. */
	if s.state == BROTLI_STATE_COMMAND_BEGIN {
		goto CommandBegin
	} else if s.state == BROTLI_STATE_COMMAND_INNER {
		goto CommandInner
	} else if s.state == BROTLI_STATE_COMMAND_POST_DECODE_LITERALS {
		goto CommandPostDecodeLiterals
	} else if s.state == BROTLI_STATE_COMMAND_POST_WRAP_COPY {
		goto CommandPostWrapCopy
	} else {
		return BROTLI_DECODER_ERROR_UNREACHABLE
	}

CommandBegin:
	if safe != 0 {
		s.state = BROTLI_STATE_COMMAND_BEGIN
	}

	if !CheckInputAmount(safe, br, 28) { /* 156 bits + 7 bytes */
		s.state = BROTLI_STATE_COMMAND_BEGIN
		result = BROTLI_DECODER_NEEDS_MORE_INPUT
		goto saveStateAndReturn
	}

	if s.block_length[1] == 0 {
		if safe != 0 {
			if !SafeDecodeCommandBlockSwitch(s) {
				result = BROTLI_DECODER_NEEDS_MORE_INPUT
				goto saveStateAndReturn
			}
		} else {
			DecodeCommandBlockSwitch(s)
		}

		goto CommandBegin
	}

	/* Read the insert/copy length in the command. */
	if safe != 0 {
		if !SafeReadCommand(s, br, &i) {
			result = BROTLI_DECODER_NEEDS_MORE_INPUT
			goto saveStateAndReturn
		}
	} else {
		ReadCommand(s, br, &i)
	}

	if i == 0 {
		goto CommandPostDecodeLiterals
	}

	s.meta_block_remaining_len -= i

CommandInner:
	if safe != 0 {
		s.state = BROTLI_STATE_COMMAND_INNER
	}

	/* Read the literals in the command. */
	if s.trivial_literal_context != 0 {
		var bits uint32
		var value uint32
		PreloadSymbol(safe, s.literal_htree, br, &bits, &value)
		for {
			if !CheckInputAmount(safe, br, 28) { /* 162 bits + 7 bytes */
				s.state = BROTLI_STATE_COMMAND_INNER
				result = BROTLI_DECODER_NEEDS_MORE_INPUT
				goto saveStateAndReturn
			}

			if s.block_length[0] == 0 {
				if safe != 0 {
					if !SafeDecodeLiteralBlockSwitch(s) {
						result = BROTLI_DECODER_NEEDS_MORE_INPUT
						goto saveStateAndReturn
					}
				} else {
					DecodeLiteralBlockSwitch(s)
				}

				PreloadSymbol(safe, s.literal_htree, br, &bits, &value)
				if s.trivial_literal_context == 0 {
					goto CommandInner
				}
			}

			if safe == 0 {
				s.ringbuffer[pos] = byte(ReadPreloadedSymbol(s.literal_htree, br, &bits, &value))
			} else {
				var literal uint32
				if !SafeReadSymbol(s.literal_htree, br, &literal) {
					result = BROTLI_DECODER_NEEDS_MORE_INPUT
					goto saveStateAndReturn
				}

				s.ringbuffer[pos] = byte(literal)
			}

			s.block_length[0]--
			pos++
			if pos == s.ringbuffer_size {
				s.state = BROTLI_STATE_COMMAND_INNER_WRITE
				i--
				goto saveStateAndReturn
			}
			i--
			if i == 0 {
				break
			}
		}
	} else {
		var p1 byte = s.ringbuffer[(pos-1)&s.ringbuffer_mask]
		var p2 byte = s.ringbuffer[(pos-2)&s.ringbuffer_mask]
		for {
			var context byte
			if !CheckInputAmount(safe, br, 28) { /* 162 bits + 7 bytes */
				s.state = BROTLI_STATE_COMMAND_INNER
				result = BROTLI_DECODER_NEEDS_MORE_INPUT
				goto saveStateAndReturn
			}

			if s.block_length[0] == 0 {
				if safe != 0 {
					if !SafeDecodeLiteralBlockSwitch(s) {
						result = BROTLI_DECODER_NEEDS_MORE_INPUT
						goto saveStateAndReturn
					}
				} else {
					DecodeLiteralBlockSwitch(s)
				}

				if s.trivial_literal_context != 0 {
					goto CommandInner
				}
			}

			context = BROTLI_CONTEXT(p1, p2, s.context_lookup)
			hc = []HuffmanCode(s.literal_hgroup.htrees[s.context_map_slice[context]])
			p2 = p1
			if safe == 0 {
				p1 = byte(ReadSymbol(hc, br))
			} else {
				var literal uint32
				if !SafeReadSymbol(hc, br, &literal) {
					result = BROTLI_DECODER_NEEDS_MORE_INPUT
					goto saveStateAndReturn
				}

				p1 = byte(literal)
			}

			s.ringbuffer[pos] = p1
			s.block_length[0]--
			pos++
			if pos == s.ringbuffer_size {
				s.state = BROTLI_STATE_COMMAND_INNER_WRITE
				i--
				goto saveStateAndReturn
			}
			i--
			if i == 0 {
				break
			}
		}
	}

	if s.meta_block_remaining_len <= 0 {
		s.state = BROTLI_STATE_METABLOCK_DONE
		goto saveStateAndReturn
	}

CommandPostDecodeLiterals:
	if safe != 0 {
		s.state = BROTLI_STATE_COMMAND_POST_DECODE_LITERALS
	}

	if s.distance_code >= 0 {
		/* Implicit distance case. */
		if s.distance_code != 0 {
			s.distance_context = 0
		} else {
			s.distance_context = 1
		}

		s.dist_rb_idx--
		s.distance_code = s.dist_rb[s.dist_rb_idx&3]
	} else {
		/* Read distance code in the command, unless it was implicitly zero. */
		if s.block_length[2] == 0 {
			if safe != 0 {
				if !SafeDecodeDistanceBlockSwitch(s) {
					result = BROTLI_DECODER_NEEDS_MORE_INPUT
					goto saveStateAndReturn
				}
			} else {
				DecodeDistanceBlockSwitch(s)
			}
		}

		if safe != 0 {
			if !SafeReadDistance(s, br) {
				result = BROTLI_DECODER_NEEDS_MORE_INPUT
				goto saveStateAndReturn
			}
		} else {
			ReadDistance(s, br)
		}
	}

	if s.max_distance != s.max_backward_distance {
		if pos < s.max_backward_distance {
			s.max_distance = pos
		} else {
			s.max_distance = s.max_backward_distance
		}
	}

	i = s.copy_length

	/* Apply copy of LZ77 back-reference, or static dictionary reference if
	   the distance is larger than the max LZ77 distance */
	if s.distance_code > s.max_distance {
		/* The maximum allowed distance is BROTLI_MAX_ALLOWED_DISTANCE = 0x7FFFFFFC.
		   With this choice, no signed overflow can occur after decoding
		   a special distance code (e.g., after adding 3 to the last distance). */
		if s.distance_code > BROTLI_MAX_ALLOWED_DISTANCE {
			return BROTLI_DECODER_ERROR_FORMAT_DISTANCE
		}

		if i >= BROTLI_MIN_DICTIONARY_WORD_LENGTH && i <= BROTLI_MAX_DICTIONARY_WORD_LENGTH {
			var address int = s.distance_code - s.max_distance - 1
			var words *BrotliDictionary = s.dictionary
			var transforms *BrotliTransforms = s.transforms
			var offset int = int(s.dictionary.offsets_by_length[i])
			var shift uint32 = uint32(s.dictionary.size_bits_by_length[i])
			var mask int = int(bitMask(shift))
			var word_idx int = address & mask
			var transform_idx int = address >> shift

			/* Compensate double distance-ring-buffer roll. */
			s.dist_rb_idx += s.distance_context

			offset += word_idx * i
			if words.data == nil {
				return BROTLI_DECODER_ERROR_DICTIONARY_NOT_SET
			}

			if transform_idx < int(transforms.num_transforms) {
				var word []byte
				word = words.data[offset:]
				var len int = i
				if transform_idx == int(transforms.cutOffTransforms[0]) {
					copy(s.ringbuffer[pos:], word[:uint(len)])
				} else {
					len = BrotliTransformDictionaryWord(s.ringbuffer[pos:], word, int(len), transforms, transform_idx)
				}

				pos += int(len)
				s.meta_block_remaining_len -= int(len)
				if pos >= s.ringbuffer_size {
					s.state = BROTLI_STATE_COMMAND_POST_WRITE_1
					goto saveStateAndReturn
				}
			} else {
				return BROTLI_DECODER_ERROR_FORMAT_TRANSFORM
			}
		} else {
			return BROTLI_DECODER_ERROR_FORMAT_DICTIONARY
		}
	} else {
		var src_start int = (pos - s.distance_code) & s.ringbuffer_mask
		var copy_dst []byte
		copy_dst = s.ringbuffer[pos:]
		var copy_src []byte
		copy_src = s.ringbuffer[src_start:]
		var dst_end int = pos + i
		var src_end int = src_start + i

		/* Update the recent distances cache. */
		s.dist_rb[s.dist_rb_idx&3] = s.distance_code

		s.dist_rb_idx++
		s.meta_block_remaining_len -= i

		/* There are 32+ bytes of slack in the ring-buffer allocation.
		   Also, we have 16 short codes, that make these 16 bytes irrelevant
		   in the ring-buffer. Let's copy over them as a first guess. */
		copy(copy_dst, copy_src[:16])

		if src_end > pos && dst_end > src_start {
			/* Regions intersect. */
			goto CommandPostWrapCopy
		}

		if dst_end >= s.ringbuffer_size || src_end >= s.ringbuffer_size {
			/* At least one region wraps. */
			goto CommandPostWrapCopy
		}

		pos += i
		if i > 16 {
			if i > 32 {
				copy(copy_dst[16:], copy_src[16:][:uint(i-16)])
			} else {
				/* This branch covers about 45% cases.
				   Fixed size short copy allows more compiler optimizations. */
				copy(copy_dst[16:], copy_src[16:][:16])
			}
		}
	}

	if s.meta_block_remaining_len <= 0 {
		/* Next metablock, if any. */
		s.state = BROTLI_STATE_METABLOCK_DONE

		goto saveStateAndReturn
	} else {
		goto CommandBegin
	}
CommandPostWrapCopy:
	{
		var wrap_guard int = s.ringbuffer_size - pos
		for {
			i--
			if i < 0 {
				break
			}
			s.ringbuffer[pos] = s.ringbuffer[(pos-s.distance_code)&s.ringbuffer_mask]
			pos++
			wrap_guard--
			if wrap_guard == 0 {
				s.state = BROTLI_STATE_COMMAND_POST_WRITE_2
				goto saveStateAndReturn
			}
		}
	}

	if s.meta_block_remaining_len <= 0 {
		/* Next metablock, if any. */
		s.state = BROTLI_STATE_METABLOCK_DONE

		goto saveStateAndReturn
	} else {
		goto CommandBegin
	}

saveStateAndReturn:
	s.pos = pos
	s.loop_counter = i
	return result
}

func ProcessCommands(s *Reader) int {
	return ProcessCommandsInternal(0, s)
}

func SafeProcessCommands(s *Reader) int {
	return ProcessCommandsInternal(1, s)
}

/* Returns the maximum number of distance symbols which can only represent
   distances not exceeding BROTLI_MAX_ALLOWED_DISTANCE. */

var BrotliMaxDistanceSymbol_bound = [maxNpostfix + 1]uint32{0, 4, 12, 28}
var BrotliMaxDistanceSymbol_diff = [maxNpostfix + 1]uint32{73, 126, 228, 424}

func BrotliMaxDistanceSymbol(ndirect uint32, npostfix uint32) uint32 {
	var postfix uint32 = 1 << npostfix
	if ndirect < BrotliMaxDistanceSymbol_bound[npostfix] {
		return ndirect + BrotliMaxDistanceSymbol_diff[npostfix] + postfix
	} else if ndirect > BrotliMaxDistanceSymbol_bound[npostfix]+postfix {
		return ndirect + BrotliMaxDistanceSymbol_diff[npostfix]
	} else {
		return BrotliMaxDistanceSymbol_bound[npostfix] + BrotliMaxDistanceSymbol_diff[npostfix] + postfix
	}
}

/* Invariant: input stream is never overconsumed:
   - invalid input implies that the whole stream is invalid -> any amount of
     input could be read and discarded
   - when result is "needs more input", then at least one more byte is REQUIRED
     to complete decoding; all input data MUST be consumed by decoder, so
     client could swap the input buffer
   - when result is "needs more output" decoder MUST ensure that it doesn't
     hold more than 7 bits in bit reader; this saves client from swapping input
     buffer ahead of time
   - when result is "success" decoder MUST return all unused data back to input
     buffer; this is possible because the invariant is held on enter */
func BrotliDecoderDecompressStream(s *Reader, available_in *uint, next_in *[]byte, available_out *uint, next_out *[]byte) int {
	var result int = BROTLI_DECODER_SUCCESS
	var br *bitReader = &s.br

	/* Do not try to process further in a case of unrecoverable error. */
	if int(s.error_code) < 0 {
		return BROTLI_DECODER_RESULT_ERROR
	}

	if *available_out != 0 && (next_out == nil || *next_out == nil) {
		return SaveErrorCode(s, BROTLI_DECODER_ERROR_INVALID_ARGUMENTS)
	}

	if *available_out == 0 {
		next_out = nil
	}
	if s.buffer_length == 0 { /* Just connect bit reader to input stream. */
		br.input_len = *available_in
		br.input = *next_in
		br.byte_pos = 0
	} else {
		/* At least one byte of input is required. More than one byte of input may
		   be required to complete the transaction -> reading more data must be
		   done in a loop -> do it in a main loop. */
		result = BROTLI_DECODER_NEEDS_MORE_INPUT

		br.input = s.buffer.u8[:]
		br.byte_pos = 0
	}

	/* State machine */
	for {
		if result != BROTLI_DECODER_SUCCESS {
			/* Error, needs more input/output. */
			if result == BROTLI_DECODER_NEEDS_MORE_INPUT {
				if s.ringbuffer != nil { /* Pro-actively push output. */
					var intermediate_result int = WriteRingBuffer(s, available_out, next_out, nil, true)

					/* WriteRingBuffer checks s->meta_block_remaining_len validity. */
					if int(intermediate_result) < 0 {
						result = intermediate_result
						break
					}
				}

				if s.buffer_length != 0 { /* Used with internal buffer. */
					if br.byte_pos == br.input_len {
						/* Successfully finished read transaction.
						   Accumulator contains less than 8 bits, because internal buffer
						   is expanded byte-by-byte until it is enough to complete read. */
						s.buffer_length = 0

						/* Switch to input stream and restart. */
						result = BROTLI_DECODER_SUCCESS

						br.input_len = *available_in
						br.input = *next_in
						br.byte_pos = 0
						continue
					} else if *available_in != 0 {
						/* Not enough data in buffer, but can take one more byte from
						   input stream. */
						result = BROTLI_DECODER_SUCCESS

						s.buffer.u8[s.buffer_length] = (*next_in)[0]
						s.buffer_length++
						br.input_len = uint(s.buffer_length)
						*next_in = (*next_in)[1:]
						(*available_in)--

						/* Retry with more data in buffer. */
						continue
					}

					/* Can't finish reading and no more input. */
					break
					/* Input stream doesn't contain enough input. */
				} else {
					/* Copy tail to internal buffer and return. */
					*next_in = br.input[br.byte_pos:]

					*available_in = br.input_len - br.byte_pos
					for *available_in != 0 {
						s.buffer.u8[s.buffer_length] = (*next_in)[0]
						s.buffer_length++
						*next_in = (*next_in)[1:]
						(*available_in)--
					}

					break
				}
			}

			/* Unreachable. */

			/* Fail or needs more output. */
			if s.buffer_length != 0 {
				/* Just consumed the buffered input and produced some output. Otherwise
				   it would result in "needs more input". Reset internal buffer. */
				s.buffer_length = 0
			} else {
				/* Using input stream in last iteration. When decoder switches to input
				   stream it has less than 8 bits in accumulator, so it is safe to
				   return unused accumulator bits there. */
				bitReaderUnload(br)

				*available_in = br.input_len - br.byte_pos
				*next_in = br.input[br.byte_pos:]
			}

			break
		}

		switch s.state {
		/* Prepare to the first read. */
		case BROTLI_STATE_UNINITED:
			if !warmupBitReader(br) {
				result = BROTLI_DECODER_NEEDS_MORE_INPUT
				break
			}

			/* Decode window size. */
			result = DecodeWindowBits(s, br) /* Reads 1..8 bits. */
			if result != BROTLI_DECODER_SUCCESS {
				break
			}

			if s.large_window {
				s.state = BROTLI_STATE_LARGE_WINDOW_BITS
				break
			}

			s.state = BROTLI_STATE_INITIALIZE

		case BROTLI_STATE_LARGE_WINDOW_BITS:
			if !safeReadBits(br, 6, &s.window_bits) {
				result = BROTLI_DECODER_NEEDS_MORE_INPUT
				break
			}

			if s.window_bits < largeMinWbits || s.window_bits > largeMaxWbits {
				result = BROTLI_DECODER_ERROR_FORMAT_WINDOW_BITS
				break
			}

			s.state = BROTLI_STATE_INITIALIZE
			fallthrough

			/* Maximum distance, see section 9.1. of the spec. */
		/* Fall through. */
		case BROTLI_STATE_INITIALIZE:
			s.max_backward_distance = (1 << s.window_bits) - BROTLI_WINDOW_GAP

			/* Allocate memory for both block_type_trees and block_len_trees. */
			s.block_type_trees = make([]HuffmanCode, (3 * (BROTLI_HUFFMAN_MAX_SIZE_258 + BROTLI_HUFFMAN_MAX_SIZE_26)))

			if s.block_type_trees == nil {
				result = BROTLI_DECODER_ERROR_ALLOC_BLOCK_TYPE_TREES
				break
			}

			s.block_len_trees = s.block_type_trees[3*BROTLI_HUFFMAN_MAX_SIZE_258:]

			s.state = BROTLI_STATE_METABLOCK_BEGIN
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_METABLOCK_BEGIN:
			BrotliDecoderStateMetablockBegin(s)

			s.state = BROTLI_STATE_METABLOCK_HEADER
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_METABLOCK_HEADER:
			result = DecodeMetaBlockLength(s, br)
			/* Reads 2 - 31 bits. */
			if result != BROTLI_DECODER_SUCCESS {
				break
			}

			if s.is_metadata != 0 || s.is_uncompressed != 0 {
				if !bitReaderJumpToByteBoundary(br) {
					result = BROTLI_DECODER_ERROR_FORMAT_PADDING_1
					break
				}
			}

			if s.is_metadata != 0 {
				s.state = BROTLI_STATE_METADATA
				break
			}

			if s.meta_block_remaining_len == 0 {
				s.state = BROTLI_STATE_METABLOCK_DONE
				break
			}

			BrotliCalculateRingBufferSize(s)
			if s.is_uncompressed != 0 {
				s.state = BROTLI_STATE_UNCOMPRESSED
				break
			}

			s.loop_counter = 0
			s.state = BROTLI_STATE_HUFFMAN_CODE_0
		case BROTLI_STATE_UNCOMPRESSED:
			{
				result = CopyUncompressedBlockToOutput(available_out, next_out, nil, s)
				if result != BROTLI_DECODER_SUCCESS {
					break
				}

				s.state = BROTLI_STATE_METABLOCK_DONE
				break
			}
			fallthrough

		case BROTLI_STATE_METADATA:
			for ; s.meta_block_remaining_len > 0; s.meta_block_remaining_len-- {
				var bits uint32

				/* Read one byte and ignore it. */
				if !safeReadBits(br, 8, &bits) {
					result = BROTLI_DECODER_NEEDS_MORE_INPUT
					break
				}
			}

			if result == BROTLI_DECODER_SUCCESS {
				s.state = BROTLI_STATE_METABLOCK_DONE
			}

		case BROTLI_STATE_HUFFMAN_CODE_0:
			if s.loop_counter >= 3 {
				s.state = BROTLI_STATE_METABLOCK_HEADER_2
				break
			}

			/* Reads 1..11 bits. */
			result = DecodeVarLenUint8(s, br, &s.num_block_types[s.loop_counter])

			if result != BROTLI_DECODER_SUCCESS {
				break
			}

			s.num_block_types[s.loop_counter]++
			if s.num_block_types[s.loop_counter] < 2 {
				s.loop_counter++
				break
			}

			s.state = BROTLI_STATE_HUFFMAN_CODE_1
			fallthrough
		/* Fall through. */
		case BROTLI_STATE_HUFFMAN_CODE_1:
			{
				var alphabet_size uint32 = s.num_block_types[s.loop_counter] + 2
				var tree_offset int = s.loop_counter * BROTLI_HUFFMAN_MAX_SIZE_258
				result = ReadHuffmanCode(alphabet_size, alphabet_size, s.block_type_trees[tree_offset:], nil, s)
				if result != BROTLI_DECODER_SUCCESS {
					break
				}
				s.state = BROTLI_STATE_HUFFMAN_CODE_2
			}
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_HUFFMAN_CODE_2:
			{
				var alphabet_size uint32 = numBlockLenSymbols
				var tree_offset int = s.loop_counter * BROTLI_HUFFMAN_MAX_SIZE_26
				result = ReadHuffmanCode(alphabet_size, alphabet_size, s.block_len_trees[tree_offset:], nil, s)
				if result != BROTLI_DECODER_SUCCESS {
					break
				}
				s.state = BROTLI_STATE_HUFFMAN_CODE_3
			}
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_HUFFMAN_CODE_3:
			{
				var tree_offset int = s.loop_counter * BROTLI_HUFFMAN_MAX_SIZE_26
				if !SafeReadBlockLength(s, &s.block_length[s.loop_counter], s.block_len_trees[tree_offset:], br) {
					result = BROTLI_DECODER_NEEDS_MORE_INPUT
					break
				}

				s.loop_counter++
				s.state = BROTLI_STATE_HUFFMAN_CODE_0
				break
			}
			fallthrough
		case BROTLI_STATE_METABLOCK_HEADER_2:
			{
				var bits uint32
				if !safeReadBits(br, 6, &bits) {
					result = BROTLI_DECODER_NEEDS_MORE_INPUT
					break
				}

				s.distance_postfix_bits = bits & bitMask(2)
				bits >>= 2
				s.num_direct_distance_codes = numDistanceShortCodes + (bits << s.distance_postfix_bits)
				s.distance_postfix_mask = int(bitMask(s.distance_postfix_bits))
				s.context_modes = make([]byte, uint(s.num_block_types[0]))
				if s.context_modes == nil {
					result = BROTLI_DECODER_ERROR_ALLOC_CONTEXT_MODES
					break
				}

				s.loop_counter = 0
				s.state = BROTLI_STATE_CONTEXT_MODES
			}
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_CONTEXT_MODES:
			result = ReadContextModes(s)

			if result != BROTLI_DECODER_SUCCESS {
				break
			}

			s.state = BROTLI_STATE_CONTEXT_MAP_1
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_CONTEXT_MAP_1:
			result = DecodeContextMap(s.num_block_types[0]<<BROTLI_LITERAL_CONTEXT_BITS, &s.num_literal_htrees, &s.context_map, s)

			if result != BROTLI_DECODER_SUCCESS {
				break
			}

			DetectTrivialLiteralBlockTypes(s)
			s.state = BROTLI_STATE_CONTEXT_MAP_2
			fallthrough
		/* Fall through. */
		case BROTLI_STATE_CONTEXT_MAP_2:
			{
				var num_direct_codes uint32 = s.num_direct_distance_codes - numDistanceShortCodes
				var num_distance_codes uint32
				var max_distance_symbol uint32
				if s.large_window {
					num_distance_codes = uint32(distanceAlphabetSize(uint(s.distance_postfix_bits), uint(num_direct_codes), largeMaxDistanceBits))
					max_distance_symbol = BrotliMaxDistanceSymbol(num_direct_codes, s.distance_postfix_bits)
				} else {
					num_distance_codes = uint32(distanceAlphabetSize(uint(s.distance_postfix_bits), uint(num_direct_codes), maxDistanceBits))
					max_distance_symbol = num_distance_codes
				}
				var allocation_success bool = true
				result = DecodeContextMap(s.num_block_types[2]<<BROTLI_DISTANCE_CONTEXT_BITS, &s.num_dist_htrees, &s.dist_context_map, s)
				if result != BROTLI_DECODER_SUCCESS {
					break
				}

				if !BrotliDecoderHuffmanTreeGroupInit(s, &s.literal_hgroup, numLiteralSymbols, numLiteralSymbols, s.num_literal_htrees) {
					allocation_success = false
				}

				if !BrotliDecoderHuffmanTreeGroupInit(s, &s.insert_copy_hgroup, numCommandSymbols, numCommandSymbols, s.num_block_types[1]) {
					allocation_success = false
				}

				if !BrotliDecoderHuffmanTreeGroupInit(s, &s.distance_hgroup, num_distance_codes, max_distance_symbol, s.num_dist_htrees) {
					allocation_success = false
				}

				if !allocation_success {
					return SaveErrorCode(s, BROTLI_DECODER_ERROR_ALLOC_TREE_GROUPS)
				}

				s.loop_counter = 0
				s.state = BROTLI_STATE_TREE_GROUP
			}
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_TREE_GROUP:
			{
				var hgroup *HuffmanTreeGroup = nil
				switch s.loop_counter {
				case 0:
					hgroup = &s.literal_hgroup
				case 1:
					hgroup = &s.insert_copy_hgroup
				case 2:
					hgroup = &s.distance_hgroup
				default:
					return SaveErrorCode(s, BROTLI_DECODER_ERROR_UNREACHABLE)
				}

				result = HuffmanTreeGroupDecode(hgroup, s)
				if result != BROTLI_DECODER_SUCCESS {
					break
				}
				s.loop_counter++
				if s.loop_counter >= 3 {
					PrepareLiteralDecoding(s)
					s.dist_context_map_slice = s.dist_context_map
					s.htree_command = []HuffmanCode(s.insert_copy_hgroup.htrees[0])
					if !BrotliEnsureRingBuffer(s) {
						result = BROTLI_DECODER_ERROR_ALLOC_RING_BUFFER_2
						break
					}

					s.state = BROTLI_STATE_COMMAND_BEGIN
				}

				break
			}
			fallthrough

		case BROTLI_STATE_COMMAND_BEGIN,
			/* Fall through. */
			BROTLI_STATE_COMMAND_INNER,

			/* Fall through. */
			BROTLI_STATE_COMMAND_POST_DECODE_LITERALS,

			/* Fall through. */
			BROTLI_STATE_COMMAND_POST_WRAP_COPY:
			result = ProcessCommands(s)

			if result == BROTLI_DECODER_NEEDS_MORE_INPUT {
				result = SafeProcessCommands(s)
			}

		case BROTLI_STATE_COMMAND_INNER_WRITE,
			/* Fall through. */
			BROTLI_STATE_COMMAND_POST_WRITE_1,

			/* Fall through. */
			BROTLI_STATE_COMMAND_POST_WRITE_2:
			result = WriteRingBuffer(s, available_out, next_out, nil, false)

			if result != BROTLI_DECODER_SUCCESS {
				break
			}

			WrapRingBuffer(s)
			if s.ringbuffer_size == 1<<s.window_bits {
				s.max_distance = s.max_backward_distance
			}

			if s.state == BROTLI_STATE_COMMAND_POST_WRITE_1 {
				if s.meta_block_remaining_len == 0 {
					/* Next metablock, if any. */
					s.state = BROTLI_STATE_METABLOCK_DONE
				} else {
					s.state = BROTLI_STATE_COMMAND_BEGIN
				}

				break
			} else if s.state == BROTLI_STATE_COMMAND_POST_WRITE_2 {
				s.state = BROTLI_STATE_COMMAND_POST_WRAP_COPY /* BROTLI_STATE_COMMAND_INNER_WRITE */
			} else {
				if s.loop_counter == 0 {
					if s.meta_block_remaining_len == 0 {
						s.state = BROTLI_STATE_METABLOCK_DONE
					} else {
						s.state = BROTLI_STATE_COMMAND_POST_DECODE_LITERALS
					}

					break
				}

				s.state = BROTLI_STATE_COMMAND_INNER
			}

		case BROTLI_STATE_METABLOCK_DONE:
			if s.meta_block_remaining_len < 0 {
				result = BROTLI_DECODER_ERROR_FORMAT_BLOCK_LENGTH_2
				break
			}

			BrotliDecoderStateCleanupAfterMetablock(s)
			if s.is_last_metablock == 0 {
				s.state = BROTLI_STATE_METABLOCK_BEGIN
				break
			}

			if !bitReaderJumpToByteBoundary(br) {
				result = BROTLI_DECODER_ERROR_FORMAT_PADDING_2
				break
			}

			if s.buffer_length == 0 {
				bitReaderUnload(br)
				*available_in = br.input_len - br.byte_pos
				*next_in = br.input[br.byte_pos:]
			}

			s.state = BROTLI_STATE_DONE
			fallthrough

			/* Fall through. */
		case BROTLI_STATE_DONE:
			if s.ringbuffer != nil {
				result = WriteRingBuffer(s, available_out, next_out, nil, true)
				if result != BROTLI_DECODER_SUCCESS {
					break
				}
			}

			return SaveErrorCode(s, result)
		}
	}

	return SaveErrorCode(s, result)
}

func BrotliDecoderHasMoreOutput(s *Reader) bool {
	/* After unrecoverable error remaining output is considered nonsensical. */
	if int(s.error_code) < 0 {
		return false
	}

	return s.ringbuffer != nil && UnwrittenBytes(s, false) != 0
}

func BrotliDecoderTakeOutput(s *Reader, size *uint) []byte {
	var result []byte = nil
	var available_out uint
	if *size != 0 {
		available_out = *size
	} else {
		available_out = 1 << 24
	}
	var requested_out uint = available_out
	var status int
	if (s.ringbuffer == nil) || (int(s.error_code) < 0) {
		*size = 0
		return nil
	}

	WrapRingBuffer(s)
	status = WriteRingBuffer(s, &available_out, &result, nil, true)

	/* Either WriteRingBuffer returns those "success" codes... */
	if status == BROTLI_DECODER_SUCCESS || status == BROTLI_DECODER_NEEDS_MORE_OUTPUT {
		*size = requested_out - available_out
	} else {
		/* ... or stream is broken. Normally this should be caught by
		   BrotliDecoderDecompressStream, this is just a safeguard. */
		if int(status) < 0 {
			SaveErrorCode(s, status)
		}
		*size = 0
		result = nil
	}

	return result
}

func BrotliDecoderIsUsed(s *Reader) bool {
	return s.state != BROTLI_STATE_UNINITED || getAvailableBits(&s.br) != 0
}

func BrotliDecoderIsFinished(s *Reader) bool {
	return (s.state == BROTLI_STATE_DONE) && !BrotliDecoderHasMoreOutput(s)
}

func BrotliDecoderGetErrorCode(s *Reader) int {
	return int(s.error_code)
}

func BrotliDecoderErrorString(c int) string {
	switch c {
	case BROTLI_DECODER_NO_ERROR:
		return "NO_ERROR"
	case BROTLI_DECODER_SUCCESS:
		return "SUCCESS"
	case BROTLI_DECODER_NEEDS_MORE_INPUT:
		return "NEEDS_MORE_INPUT"
	case BROTLI_DECODER_NEEDS_MORE_OUTPUT:
		return "NEEDS_MORE_OUTPUT"
	case BROTLI_DECODER_ERROR_FORMAT_EXUBERANT_NIBBLE:
		return "EXUBERANT_NIBBLE"
	case BROTLI_DECODER_ERROR_FORMAT_RESERVED:
		return "RESERVED"
	case BROTLI_DECODER_ERROR_FORMAT_EXUBERANT_META_NIBBLE:
		return "EXUBERANT_META_NIBBLE"
	case BROTLI_DECODER_ERROR_FORMAT_SIMPLE_HUFFMAN_ALPHABET:
		return "SIMPLE_HUFFMAN_ALPHABET"
	case BROTLI_DECODER_ERROR_FORMAT_SIMPLE_HUFFMAN_SAME:
		return "SIMPLE_HUFFMAN_SAME"
	case BROTLI_DECODER_ERROR_FORMAT_CL_SPACE:
		return "CL_SPACE"
	case BROTLI_DECODER_ERROR_FORMAT_HUFFMAN_SPACE:
		return "HUFFMAN_SPACE"
	case BROTLI_DECODER_ERROR_FORMAT_CONTEXT_MAP_REPEAT:
		return "CONTEXT_MAP_REPEAT"
	case BROTLI_DECODER_ERROR_FORMAT_BLOCK_LENGTH_1:
		return "BLOCK_LENGTH_1"
	case BROTLI_DECODER_ERROR_FORMAT_BLOCK_LENGTH_2:
		return "BLOCK_LENGTH_2"
	case BROTLI_DECODER_ERROR_FORMAT_TRANSFORM:
		return "TRANSFORM"
	case BROTLI_DECODER_ERROR_FORMAT_DICTIONARY:
		return "DICTIONARY"
	case BROTLI_DECODER_ERROR_FORMAT_WINDOW_BITS:
		return "WINDOW_BITS"
	case BROTLI_DECODER_ERROR_FORMAT_PADDING_1:
		return "PADDING_1"
	case BROTLI_DECODER_ERROR_FORMAT_PADDING_2:
		return "PADDING_2"
	case BROTLI_DECODER_ERROR_FORMAT_DISTANCE:
		return "DISTANCE"
	case BROTLI_DECODER_ERROR_DICTIONARY_NOT_SET:
		return "DICTIONARY_NOT_SET"
	case BROTLI_DECODER_ERROR_INVALID_ARGUMENTS:
		return "INVALID_ARGUMENTS"
	case BROTLI_DECODER_ERROR_ALLOC_CONTEXT_MODES:
		return "CONTEXT_MODES"
	case BROTLI_DECODER_ERROR_ALLOC_TREE_GROUPS:
		return "TREE_GROUPS"
	case BROTLI_DECODER_ERROR_ALLOC_CONTEXT_MAP:
		return "CONTEXT_MAP"
	case BROTLI_DECODER_ERROR_ALLOC_RING_BUFFER_1:
		return "RING_BUFFER_1"
	case BROTLI_DECODER_ERROR_ALLOC_RING_BUFFER_2:
		return "RING_BUFFER_2"
	case BROTLI_DECODER_ERROR_ALLOC_BLOCK_TYPE_TREES:
		return "BLOCK_TYPE_TREES"
	case BROTLI_DECODER_ERROR_UNREACHABLE:
		return "UNREACHABLE"
	default:
		return "INVALID"
	}
}

func BrotliDecoderVersion() uint32 {
	return BROTLI_VERSION
}
