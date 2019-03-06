package brotli

/* Copyright 2015 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/
/* Copyright 2015 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Brotli state for partial streaming decoding. */
const (
	BROTLI_STATE_UNINITED = iota
	BROTLI_STATE_LARGE_WINDOW_BITS
	BROTLI_STATE_INITIALIZE
	BROTLI_STATE_METABLOCK_BEGIN
	BROTLI_STATE_METABLOCK_HEADER
	BROTLI_STATE_METABLOCK_HEADER_2
	BROTLI_STATE_CONTEXT_MODES
	BROTLI_STATE_COMMAND_BEGIN
	BROTLI_STATE_COMMAND_INNER
	BROTLI_STATE_COMMAND_POST_DECODE_LITERALS
	BROTLI_STATE_COMMAND_POST_WRAP_COPY
	BROTLI_STATE_UNCOMPRESSED
	BROTLI_STATE_METADATA
	BROTLI_STATE_COMMAND_INNER_WRITE
	BROTLI_STATE_METABLOCK_DONE
	BROTLI_STATE_COMMAND_POST_WRITE_1
	BROTLI_STATE_COMMAND_POST_WRITE_2
	BROTLI_STATE_HUFFMAN_CODE_0
	BROTLI_STATE_HUFFMAN_CODE_1
	BROTLI_STATE_HUFFMAN_CODE_2
	BROTLI_STATE_HUFFMAN_CODE_3
	BROTLI_STATE_CONTEXT_MAP_1
	BROTLI_STATE_CONTEXT_MAP_2
	BROTLI_STATE_TREE_GROUP
	BROTLI_STATE_DONE
)

const (
	BROTLI_STATE_METABLOCK_HEADER_NONE = iota
	BROTLI_STATE_METABLOCK_HEADER_EMPTY
	BROTLI_STATE_METABLOCK_HEADER_NIBBLES
	BROTLI_STATE_METABLOCK_HEADER_SIZE
	BROTLI_STATE_METABLOCK_HEADER_UNCOMPRESSED
	BROTLI_STATE_METABLOCK_HEADER_RESERVED
	BROTLI_STATE_METABLOCK_HEADER_BYTES
	BROTLI_STATE_METABLOCK_HEADER_METADATA
)

const (
	BROTLI_STATE_UNCOMPRESSED_NONE = iota
	BROTLI_STATE_UNCOMPRESSED_WRITE
)

const (
	BROTLI_STATE_TREE_GROUP_NONE = iota
	BROTLI_STATE_TREE_GROUP_LOOP
)

const (
	BROTLI_STATE_CONTEXT_MAP_NONE = iota
	BROTLI_STATE_CONTEXT_MAP_READ_PREFIX
	BROTLI_STATE_CONTEXT_MAP_HUFFMAN
	BROTLI_STATE_CONTEXT_MAP_DECODE
	BROTLI_STATE_CONTEXT_MAP_TRANSFORM
)

const (
	BROTLI_STATE_HUFFMAN_NONE = iota
	BROTLI_STATE_HUFFMAN_SIMPLE_SIZE
	BROTLI_STATE_HUFFMAN_SIMPLE_READ
	BROTLI_STATE_HUFFMAN_SIMPLE_BUILD
	BROTLI_STATE_HUFFMAN_COMPLEX
	BROTLI_STATE_HUFFMAN_LENGTH_SYMBOLS
)

const (
	BROTLI_STATE_DECODE_UINT8_NONE = iota
	BROTLI_STATE_DECODE_UINT8_SHORT
	BROTLI_STATE_DECODE_UINT8_LONG
)

const (
	BROTLI_STATE_READ_BLOCK_LENGTH_NONE = iota
	BROTLI_STATE_READ_BLOCK_LENGTH_SUFFIX
)

type BrotliDecoderState struct {
	state        int
	loop_counter int
	br           BrotliBitReader
	buffer       struct {
		u64 uint64
		u8  [8]byte
	}
	buffer_length               uint32
	pos                         int
	max_backward_distance       int
	max_distance                int
	ringbuffer_size             int
	ringbuffer_mask             int
	dist_rb_idx                 int
	dist_rb                     [4]int
	error_code                  int
	sub_loop_counter            uint32
	ringbuffer                  []byte
	ringbuffer_end              []byte
	htree_command               []HuffmanCode
	context_lookup              []byte
	context_map_slice           []byte
	dist_context_map_slice      []byte
	literal_hgroup              HuffmanTreeGroup
	insert_copy_hgroup          HuffmanTreeGroup
	distance_hgroup             HuffmanTreeGroup
	block_type_trees            []HuffmanCode
	block_len_trees             []HuffmanCode
	trivial_literal_context     int
	distance_context            int
	meta_block_remaining_len    int
	block_length_index          uint32
	block_length                [3]uint32
	num_block_types             [3]uint32
	block_type_rb               [6]uint32
	distance_postfix_bits       uint32
	num_direct_distance_codes   uint32
	distance_postfix_mask       int
	num_dist_htrees             uint32
	dist_context_map            []byte
	literal_htree               []HuffmanCode
	dist_htree_index            byte
	repeat_code_len             uint32
	prev_code_len               uint32
	copy_length                 int
	distance_code               int
	rb_roundtrips               uint
	partial_pos_out             uint
	symbol                      uint32
	repeat                      uint32
	space                       uint32
	table                       [32]HuffmanCode
	symbol_lists                SymbolList
	symbols_lists_array         [BROTLI_HUFFMAN_MAX_CODE_LENGTH + 1 + BROTLI_NUM_COMMAND_SYMBOLS]uint16
	next_symbol                 [32]int
	code_length_code_lengths    [BROTLI_CODE_LENGTH_CODES]byte
	code_length_histo           [16]uint16
	htree_index                 int
	next                        []HuffmanCode
	context_index               uint32
	max_run_length_prefix       uint32
	code                        uint32
	context_map_table           [BROTLI_HUFFMAN_MAX_SIZE_272]HuffmanCode
	substate_metablock_header   int
	substate_tree_group         int
	substate_context_map        int
	substate_uncompressed       int
	substate_huffman            int
	substate_decode_uint8       int
	substate_read_block_length  int
	is_last_metablock           uint
	is_uncompressed             uint
	is_metadata                 uint
	should_wrap_ringbuffer      uint
	canny_ringbuffer_allocation uint
	large_window                bool
	size_nibbles                uint
	window_bits                 uint32
	new_ringbuffer_size         int
	num_literal_htrees          uint32
	context_map                 []byte
	context_modes               []byte
	dictionary                  *BrotliDictionary
	transforms                  *BrotliTransforms
	trivial_literal_contexts    [8]uint32
}

func BrotliDecoderStateInit(s *BrotliDecoderState) bool {
	s.error_code = 0 /* BROTLI_DECODER_NO_ERROR */

	BrotliInitBitReader(&s.br)
	s.state = BROTLI_STATE_UNINITED
	s.large_window = false
	s.substate_metablock_header = BROTLI_STATE_METABLOCK_HEADER_NONE
	s.substate_tree_group = BROTLI_STATE_TREE_GROUP_NONE
	s.substate_context_map = BROTLI_STATE_CONTEXT_MAP_NONE
	s.substate_uncompressed = BROTLI_STATE_UNCOMPRESSED_NONE
	s.substate_huffman = BROTLI_STATE_HUFFMAN_NONE
	s.substate_decode_uint8 = BROTLI_STATE_DECODE_UINT8_NONE
	s.substate_read_block_length = BROTLI_STATE_READ_BLOCK_LENGTH_NONE

	s.buffer_length = 0
	s.loop_counter = 0
	s.pos = 0
	s.rb_roundtrips = 0
	s.partial_pos_out = 0

	s.block_type_trees = nil
	s.block_len_trees = nil
	s.ringbuffer = nil
	s.ringbuffer_size = 0
	s.new_ringbuffer_size = 0
	s.ringbuffer_mask = 0

	s.context_map = nil
	s.context_modes = nil
	s.dist_context_map = nil
	s.context_map_slice = nil
	s.dist_context_map_slice = nil

	s.sub_loop_counter = 0

	s.literal_hgroup.codes = nil
	s.literal_hgroup.htrees = nil
	s.insert_copy_hgroup.codes = nil
	s.insert_copy_hgroup.htrees = nil
	s.distance_hgroup.codes = nil
	s.distance_hgroup.htrees = nil

	s.is_last_metablock = 0
	s.is_uncompressed = 0
	s.is_metadata = 0
	s.should_wrap_ringbuffer = 0
	s.canny_ringbuffer_allocation = 1

	s.window_bits = 0
	s.max_distance = 0
	s.dist_rb[0] = 16
	s.dist_rb[1] = 15
	s.dist_rb[2] = 11
	s.dist_rb[3] = 4
	s.dist_rb_idx = 0
	s.block_type_trees = nil
	s.block_len_trees = nil

	s.symbol_lists.storage = s.symbols_lists_array[:]
	s.symbol_lists.offset = BROTLI_HUFFMAN_MAX_CODE_LENGTH + 1

	s.dictionary = BrotliGetDictionary()
	s.transforms = BrotliGetTransforms()

	return true
}

func BrotliDecoderStateMetablockBegin(s *BrotliDecoderState) {
	s.meta_block_remaining_len = 0
	s.block_length[0] = 1 << 24
	s.block_length[1] = 1 << 24
	s.block_length[2] = 1 << 24
	s.num_block_types[0] = 1
	s.num_block_types[1] = 1
	s.num_block_types[2] = 1
	s.block_type_rb[0] = 1
	s.block_type_rb[1] = 0
	s.block_type_rb[2] = 1
	s.block_type_rb[3] = 0
	s.block_type_rb[4] = 1
	s.block_type_rb[5] = 0
	s.context_map = nil
	s.context_modes = nil
	s.dist_context_map = nil
	s.context_map_slice = nil
	s.literal_htree = nil
	s.dist_context_map_slice = nil
	s.dist_htree_index = 0
	s.context_lookup = nil
	s.literal_hgroup.codes = nil
	s.literal_hgroup.htrees = nil
	s.insert_copy_hgroup.codes = nil
	s.insert_copy_hgroup.htrees = nil
	s.distance_hgroup.codes = nil
	s.distance_hgroup.htrees = nil
}

func BrotliDecoderStateCleanupAfterMetablock(s *BrotliDecoderState) {
	s.context_modes = nil
	s.context_map = nil
	s.dist_context_map = nil
	s.literal_hgroup.htrees = nil
	s.insert_copy_hgroup.htrees = nil
	s.distance_hgroup.htrees = nil
}

func BrotliDecoderStateCleanup(s *BrotliDecoderState) {
	BrotliDecoderStateCleanupAfterMetablock(s)

	s.ringbuffer = nil
	s.block_type_trees = nil
}

func BrotliDecoderHuffmanTreeGroupInit(s *BrotliDecoderState, group *HuffmanTreeGroup, alphabet_size uint32, max_symbol uint32, ntrees uint32) bool {
	var max_table_size uint = uint(kMaxHuffmanTableSize[(alphabet_size+31)>>5])
	group.alphabet_size = uint16(alphabet_size)
	group.max_symbol = uint16(max_symbol)
	group.num_htrees = uint16(ntrees)
	group.htrees = make([][]HuffmanCode, ntrees)
	group.codes = make([]HuffmanCode, (uint(ntrees) * max_table_size))
	return !(group.codes == nil)
}
