package brotli

const MAX_HUFFMAN_TREE_SIZE = (2*BROTLI_NUM_COMMAND_SYMBOLS + 1)

/* The maximum size of Huffman dictionary for distances assuming that
   NPOSTFIX = 0 and NDIRECT = 0. */
const MAX_SIMPLE_DISTANCE_ALPHABET_SIZE = 140

/* MAX_SIMPLE_DISTANCE_ALPHABET_SIZE == 140 */

/* Represents the range of values belonging to a prefix code:
   [offset, offset + 2^nbits) */
type PrefixCodeRange struct {
	offset uint32
	nbits  uint32
}

var kBlockLengthPrefixCode = [BROTLI_NUM_BLOCK_LEN_SYMBOLS]PrefixCodeRange{
	PrefixCodeRange{1, 2},
	PrefixCodeRange{5, 2},
	PrefixCodeRange{9, 2},
	PrefixCodeRange{13, 2},
	PrefixCodeRange{17, 3},
	PrefixCodeRange{25, 3},
	PrefixCodeRange{33, 3},
	PrefixCodeRange{41, 3},
	PrefixCodeRange{49, 4},
	PrefixCodeRange{65, 4},
	PrefixCodeRange{81, 4},
	PrefixCodeRange{97, 4},
	PrefixCodeRange{113, 5},
	PrefixCodeRange{145, 5},
	PrefixCodeRange{177, 5},
	PrefixCodeRange{209, 5},
	PrefixCodeRange{241, 6},
	PrefixCodeRange{305, 6},
	PrefixCodeRange{369, 7},
	PrefixCodeRange{497, 8},
	PrefixCodeRange{753, 9},
	PrefixCodeRange{1265, 10},
	PrefixCodeRange{2289, 11},
	PrefixCodeRange{4337, 12},
	PrefixCodeRange{8433, 13},
	PrefixCodeRange{16625, 24},
}

func BlockLengthPrefixCode(len uint32) uint32 {
	var code uint32
	if len >= 177 {
		if len >= 753 {
			code = 20
		} else {
			code = 14
		}
	} else if len >= 41 {
		code = 7
	} else {
		code = 0
	}
	for code < (BROTLI_NUM_BLOCK_LEN_SYMBOLS-1) && len >= kBlockLengthPrefixCode[code+1].offset {
		code++
	}
	return code
}

func GetBlockLengthPrefixCode(len uint32, code *uint, n_extra *uint32, extra *uint32) {
	*code = uint(BlockLengthPrefixCode(uint32(len)))
	*n_extra = kBlockLengthPrefixCode[*code].nbits
	*extra = len - kBlockLengthPrefixCode[*code].offset
}

type BlockTypeCodeCalculator struct {
	last_type        uint
	second_last_type uint
}

func InitBlockTypeCodeCalculator(self *BlockTypeCodeCalculator) {
	self.last_type = 1
	self.second_last_type = 0
}

func NextBlockTypeCode(calculator *BlockTypeCodeCalculator, type_ byte) uint {
	var type_code uint
	if uint(type_) == calculator.last_type+1 {
		type_code = 1
	} else if uint(type_) == calculator.second_last_type {
		type_code = 0
	} else {
		type_code = uint(type_) + 2
	}
	calculator.second_last_type = calculator.last_type
	calculator.last_type = uint(type_)
	return type_code
}

/* |nibblesbits| represents the 2 bits to encode MNIBBLES (0-3)
   REQUIRES: length > 0
   REQUIRES: length <= (1 << 24) */
func BrotliEncodeMlen(length uint, bits *uint64, numbits *uint, nibblesbits *uint64) {
	var lg uint
	if length == 1 {
		lg = 1
	} else {
		lg = uint(Log2FloorNonZero(uint(uint32(length-1)))) + 1
	}
	var tmp uint
	if lg < 16 {
		tmp = 16
	} else {
		tmp = (lg + 3)
	}
	var mnibbles uint = tmp / 4
	assert(length > 0)
	assert(length <= 1<<24)
	assert(lg <= 24)
	*nibblesbits = uint64(mnibbles) - 4
	*numbits = mnibbles * 4
	*bits = uint64(length) - 1
}

func StoreCommandExtra(cmd *Command, storage_ix *uint, storage []byte) {
	var copylen_code uint32 = CommandCopyLenCode(cmd)
	var inscode uint16 = GetInsertLengthCode(uint(cmd.insert_len_))
	var copycode uint16 = GetCopyLengthCode(uint(copylen_code))
	var insnumextra uint32 = GetInsertExtra(inscode)
	var insextraval uint64 = uint64(cmd.insert_len_) - uint64(GetInsertBase(inscode))
	var copyextraval uint64 = uint64(copylen_code) - uint64(GetCopyBase(copycode))
	var bits uint64 = copyextraval<<insnumextra | insextraval
	BrotliWriteBits(uint(insnumextra+GetCopyExtra(copycode)), bits, storage_ix, storage)
}

/* Data structure that stores almost everything that is needed to encode each
   block switch command. */
type BlockSplitCode struct {
	type_code_calculator BlockTypeCodeCalculator
	type_depths          [BROTLI_MAX_BLOCK_TYPE_SYMBOLS]byte
	type_bits            [BROTLI_MAX_BLOCK_TYPE_SYMBOLS]uint16
	length_depths        [BROTLI_NUM_BLOCK_LEN_SYMBOLS]byte
	length_bits          [BROTLI_NUM_BLOCK_LEN_SYMBOLS]uint16
}

/* Stores a number between 0 and 255. */
func StoreVarLenUint8(n uint, storage_ix *uint, storage []byte) {
	if n == 0 {
		BrotliWriteBits(1, 0, storage_ix, storage)
	} else {
		var nbits uint = uint(Log2FloorNonZero(n))
		BrotliWriteBits(1, 1, storage_ix, storage)
		BrotliWriteBits(3, uint64(nbits), storage_ix, storage)
		BrotliWriteBits(nbits, uint64(n)-(uint64(uint(1))<<nbits), storage_ix, storage)
	}
}

/* Stores the compressed meta-block header.
   REQUIRES: length > 0
   REQUIRES: length <= (1 << 24) */
func StoreCompressedMetaBlockHeader(is_final_block bool, length uint, storage_ix *uint, storage []byte) {
	var lenbits uint64
	var nlenbits uint
	var nibblesbits uint64
	var is_final uint64
	if is_final_block {
		is_final = 1
	} else {
		is_final = 0
	}

	/* Write ISLAST bit. */
	BrotliWriteBits(1, is_final, storage_ix, storage)

	/* Write ISEMPTY bit. */
	if is_final_block {
		BrotliWriteBits(1, 0, storage_ix, storage)
	}

	BrotliEncodeMlen(length, &lenbits, &nlenbits, &nibblesbits)
	BrotliWriteBits(2, nibblesbits, storage_ix, storage)
	BrotliWriteBits(nlenbits, lenbits, storage_ix, storage)

	if !is_final_block {
		/* Write ISUNCOMPRESSED bit. */
		BrotliWriteBits(1, 0, storage_ix, storage)
	}
}

/* Stores the uncompressed meta-block header.
   REQUIRES: length > 0
   REQUIRES: length <= (1 << 24) */
func BrotliStoreUncompressedMetaBlockHeader(length uint, storage_ix *uint, storage []byte) {
	var lenbits uint64
	var nlenbits uint
	var nibblesbits uint64

	/* Write ISLAST bit.
	   Uncompressed block cannot be the last one, so set to 0. */
	BrotliWriteBits(1, 0, storage_ix, storage)

	BrotliEncodeMlen(length, &lenbits, &nlenbits, &nibblesbits)
	BrotliWriteBits(2, nibblesbits, storage_ix, storage)
	BrotliWriteBits(nlenbits, lenbits, storage_ix, storage)

	/* Write ISUNCOMPRESSED bit. */
	BrotliWriteBits(1, 1, storage_ix, storage)
}

var BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kStorageOrder = [BROTLI_CODE_LENGTH_CODES]byte{1, 2, 3, 4, 0, 5, 17, 6, 16, 7, 8, 9, 10, 11, 12, 13, 14, 15}

var BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kHuffmanBitLengthHuffmanCodeSymbols = [6]byte{0, 7, 3, 2, 1, 15}
var BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kHuffmanBitLengthHuffmanCodeBitLengths = [6]byte{2, 4, 3, 2, 2, 4}

func BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask(num_codes int, code_length_bitdepth []byte, storage_ix *uint, storage []byte) {
	var skip_some uint = 0
	var codes_to_store uint = BROTLI_CODE_LENGTH_CODES
	/* The bit lengths of the Huffman code over the code length alphabet
	   are compressed with the following static Huffman code:
	     Symbol   Code
	     ------   ----
	     0          00
	     1        1110
	     2         110
	     3          01
	     4          10
	     5        1111 */

	/* Throw away trailing zeros: */
	if num_codes > 1 {
		for ; codes_to_store > 0; codes_to_store-- {
			if code_length_bitdepth[BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kStorageOrder[codes_to_store-1]] != 0 {
				break
			}
		}
	}

	if code_length_bitdepth[BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kStorageOrder[0]] == 0 && code_length_bitdepth[BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kStorageOrder[1]] == 0 {
		skip_some = 2 /* skips two. */
		if code_length_bitdepth[BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kStorageOrder[2]] == 0 {
			skip_some = 3 /* skips three. */
		}
	}

	BrotliWriteBits(2, uint64(skip_some), storage_ix, storage)
	{
		var i uint
		for i = skip_some; i < codes_to_store; i++ {
			var l uint = uint(code_length_bitdepth[BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kStorageOrder[i]])
			BrotliWriteBits(uint(BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kHuffmanBitLengthHuffmanCodeBitLengths[l]), uint64(BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask_kHuffmanBitLengthHuffmanCodeSymbols[l]), storage_ix, storage)
		}
	}
}

func BrotliStoreHuffmanTreeToBitMask(huffman_tree_size uint, huffman_tree []byte, huffman_tree_extra_bits []byte, code_length_bitdepth []byte, code_length_bitdepth_symbols []uint16, storage_ix *uint, storage []byte) {
	var i uint
	for i = 0; i < huffman_tree_size; i++ {
		var ix uint = uint(huffman_tree[i])
		BrotliWriteBits(uint(code_length_bitdepth[ix]), uint64(code_length_bitdepth_symbols[ix]), storage_ix, storage)

		/* Extra bits */
		switch ix {
		case BROTLI_REPEAT_PREVIOUS_CODE_LENGTH:
			BrotliWriteBits(2, uint64(huffman_tree_extra_bits[i]), storage_ix, storage)

		case BROTLI_REPEAT_ZERO_CODE_LENGTH:
			BrotliWriteBits(3, uint64(huffman_tree_extra_bits[i]), storage_ix, storage)
		}
	}
}

func StoreSimpleHuffmanTree(depths []byte, symbols []uint, num_symbols uint, max_bits uint, storage_ix *uint, storage []byte) {
	/* value of 1 indicates a simple Huffman code */
	BrotliWriteBits(2, 1, storage_ix, storage)

	BrotliWriteBits(2, uint64(num_symbols)-1, storage_ix, storage) /* NSYM - 1 */
	{
		/* Sort */
		var i uint
		for i = 0; i < num_symbols; i++ {
			var j uint
			for j = i + 1; j < num_symbols; j++ {
				if depths[symbols[j]] < depths[symbols[i]] {
					var tmp uint = symbols[j]
					symbols[j] = symbols[i]
					symbols[i] = tmp
				}
			}
		}
	}

	if num_symbols == 2 {
		BrotliWriteBits(max_bits, uint64(symbols[0]), storage_ix, storage)
		BrotliWriteBits(max_bits, uint64(symbols[1]), storage_ix, storage)
	} else if num_symbols == 3 {
		BrotliWriteBits(max_bits, uint64(symbols[0]), storage_ix, storage)
		BrotliWriteBits(max_bits, uint64(symbols[1]), storage_ix, storage)
		BrotliWriteBits(max_bits, uint64(symbols[2]), storage_ix, storage)
	} else {
		BrotliWriteBits(max_bits, uint64(symbols[0]), storage_ix, storage)
		BrotliWriteBits(max_bits, uint64(symbols[1]), storage_ix, storage)
		BrotliWriteBits(max_bits, uint64(symbols[2]), storage_ix, storage)
		BrotliWriteBits(max_bits, uint64(symbols[3]), storage_ix, storage)

		/* tree-select */
		var tmp int
		if depths[symbols[0]] == 1 {
			tmp = 1
		} else {
			tmp = 0
		}
		BrotliWriteBits(1, uint64(tmp), storage_ix, storage)
	}
}

/* num = alphabet size
   depths = symbol depths */
func BrotliStoreHuffmanTree(depths []byte, num uint, tree []HuffmanTree, storage_ix *uint, storage []byte) {
	var huffman_tree [BROTLI_NUM_COMMAND_SYMBOLS]byte
	var huffman_tree_extra_bits [BROTLI_NUM_COMMAND_SYMBOLS]byte
	var huffman_tree_size uint = 0
	var code_length_bitdepth = [BROTLI_CODE_LENGTH_CODES]byte{0}
	var code_length_bitdepth_symbols [BROTLI_CODE_LENGTH_CODES]uint16
	var huffman_tree_histogram = [BROTLI_CODE_LENGTH_CODES]uint32{0}
	var i uint
	var num_codes int = 0
	/* Write the Huffman tree into the brotli-representation.
	   The command alphabet is the largest, so this allocation will fit all
	   alphabets. */

	var code uint = 0

	assert(num <= BROTLI_NUM_COMMAND_SYMBOLS)

	BrotliWriteHuffmanTree(depths, num, &huffman_tree_size, huffman_tree[:], huffman_tree_extra_bits[:])

	/* Calculate the statistics of the Huffman tree in brotli-representation. */
	for i = 0; i < huffman_tree_size; i++ {
		huffman_tree_histogram[huffman_tree[i]]++
	}

	for i = 0; i < BROTLI_CODE_LENGTH_CODES; i++ {
		if huffman_tree_histogram[i] != 0 {
			if num_codes == 0 {
				code = i
				num_codes = 1
			} else if num_codes == 1 {
				num_codes = 2
				break
			}
		}
	}

	/* Calculate another Huffman tree to use for compressing both the
	   earlier Huffman tree with. */
	BrotliCreateHuffmanTree(huffman_tree_histogram[:], BROTLI_CODE_LENGTH_CODES, 5, tree, code_length_bitdepth[:])

	BrotliConvertBitDepthsToSymbols(code_length_bitdepth[:], BROTLI_CODE_LENGTH_CODES, code_length_bitdepth_symbols[:])

	/* Now, we have all the data, let's start storing it */
	BrotliStoreHuffmanTreeOfHuffmanTreeToBitMask(num_codes, code_length_bitdepth[:], storage_ix, storage)

	if num_codes == 1 {
		code_length_bitdepth[code] = 0
	}

	/* Store the real Huffman tree now. */
	BrotliStoreHuffmanTreeToBitMask(huffman_tree_size, huffman_tree[:], huffman_tree_extra_bits[:], code_length_bitdepth[:], code_length_bitdepth_symbols[:], storage_ix, storage)
}

/* Builds a Huffman tree from histogram[0:length] into depth[0:length] and
   bits[0:length] and stores the encoded tree to the bit stream. */
func BuildAndStoreHuffmanTree(histogram []uint32, histogram_length uint, alphabet_size uint, tree []HuffmanTree, depth []byte, bits []uint16, storage_ix *uint, storage []byte) {
	var count uint = 0
	var s4 = [4]uint{0}
	var i uint
	var max_bits uint = 0
	for i = 0; i < histogram_length; i++ {
		if histogram[i] != 0 {
			if count < 4 {
				s4[count] = i
			} else if count > 4 {
				break
			}

			count++
		}
	}
	{
		var max_bits_counter uint = alphabet_size - 1
		for max_bits_counter != 0 {
			max_bits_counter >>= 1
			max_bits++
		}
	}

	if count <= 1 {
		BrotliWriteBits(4, 1, storage_ix, storage)
		BrotliWriteBits(max_bits, uint64(s4[0]), storage_ix, storage)
		depth[s4[0]] = 0
		bits[s4[0]] = 0
		return
	}

	for i := 0; i < int(histogram_length); i++ {
		depth[i] = 0
	}
	BrotliCreateHuffmanTree(histogram, histogram_length, 15, tree, depth)
	BrotliConvertBitDepthsToSymbols(depth, histogram_length, bits)

	if count <= 4 {
		StoreSimpleHuffmanTree(depth, s4[:], count, max_bits, storage_ix, storage)
	} else {
		BrotliStoreHuffmanTree(depth, histogram_length, tree, storage_ix, storage)
	}
}

func SortHuffmanTree1(v0 *HuffmanTree, v1 *HuffmanTree) bool {
	return v0.total_count_ < v1.total_count_
}

func BrotliBuildAndStoreHuffmanTreeFast(histogram []uint32, histogram_total uint, max_bits uint, depth []byte, bits []uint16, storage_ix *uint, storage []byte) {
	var count uint = 0
	var symbols = [4]uint{0}
	var length uint = 0
	var total uint = histogram_total
	for total != 0 {
		if histogram[length] != 0 {
			if count < 4 {
				symbols[count] = length
			}

			count++
			total -= uint(histogram[length])
		}

		length++
	}

	if count <= 1 {
		BrotliWriteBits(4, 1, storage_ix, storage)
		BrotliWriteBits(max_bits, uint64(symbols[0]), storage_ix, storage)
		depth[symbols[0]] = 0
		bits[symbols[0]] = 0
		return
	}

	for i := 0; i < int(length); i++ {
		depth[i] = 0
	}
	{
		var max_tree_size uint = 2*length + 1
		var tree []HuffmanTree = make([]HuffmanTree, max_tree_size)
		var count_limit uint32
		for count_limit = 1; ; count_limit *= 2 {
			var node int = 0
			var l uint
			for l = length; l != 0; {
				l--
				if histogram[l] != 0 {
					if histogram[l] >= count_limit {
						InitHuffmanTree(&tree[node:][0], histogram[l], -1, int16(l))
					} else {
						InitHuffmanTree(&tree[node:][0], count_limit, -1, int16(l))
					}

					node++
				}
			}
			{
				var n int = node
				/* Points to the next leaf node. */ /* Points to the next non-leaf node. */
				var sentinel HuffmanTree
				var i int = 0
				var j int = n + 1
				var k int

				SortHuffmanTreeItems(tree, uint(n), HuffmanTreeComparator(SortHuffmanTree1))

				/* The nodes are:
				   [0, n): the sorted leaf nodes that we start with.
				   [n]: we add a sentinel here.
				   [n + 1, 2n): new parent nodes are added here, starting from
				                (n+1). These are naturally in ascending order.
				   [2n]: we add a sentinel at the end as well.
				   There will be (2n+1) elements at the end. */
				InitHuffmanTree(&sentinel, BROTLI_UINT32_MAX, -1, -1)

				tree[node] = sentinel
				node++
				tree[node] = sentinel
				node++

				for k = n - 1; k > 0; k-- {
					var left int
					var right int
					if tree[i].total_count_ <= tree[j].total_count_ {
						left = i
						i++
					} else {
						left = j
						j++
					}

					if tree[i].total_count_ <= tree[j].total_count_ {
						right = i
						i++
					} else {
						right = j
						j++
					}

					/* The sentinel node becomes the parent node. */
					tree[node-1].total_count_ = tree[left].total_count_ + tree[right].total_count_

					tree[node-1].index_left_ = int16(left)
					tree[node-1].index_right_or_value_ = int16(right)

					/* Add back the last sentinel node. */
					tree[node] = sentinel
					node++
				}

				if BrotliSetDepth(2*n-1, tree, depth, 14) {
					/* We need to pack the Huffman tree in 14 bits. If this was not
					   successful, add fake entities to the lowest values and retry. */
					break
				}
			}
		}

		tree = nil
	}

	BrotliConvertBitDepthsToSymbols(depth, length, bits)
	if count <= 4 {
		var i uint

		/* value of 1 indicates a simple Huffman code */
		BrotliWriteBits(2, 1, storage_ix, storage)

		BrotliWriteBits(2, uint64(count)-1, storage_ix, storage) /* NSYM - 1 */

		/* Sort */
		for i = 0; i < count; i++ {
			var j uint
			for j = i + 1; j < count; j++ {
				if depth[symbols[j]] < depth[symbols[i]] {
					var tmp uint = symbols[j]
					symbols[j] = symbols[i]
					symbols[i] = tmp
				}
			}
		}

		if count == 2 {
			BrotliWriteBits(max_bits, uint64(symbols[0]), storage_ix, storage)
			BrotliWriteBits(max_bits, uint64(symbols[1]), storage_ix, storage)
		} else if count == 3 {
			BrotliWriteBits(max_bits, uint64(symbols[0]), storage_ix, storage)
			BrotliWriteBits(max_bits, uint64(symbols[1]), storage_ix, storage)
			BrotliWriteBits(max_bits, uint64(symbols[2]), storage_ix, storage)
		} else {
			BrotliWriteBits(max_bits, uint64(symbols[0]), storage_ix, storage)
			BrotliWriteBits(max_bits, uint64(symbols[1]), storage_ix, storage)
			BrotliWriteBits(max_bits, uint64(symbols[2]), storage_ix, storage)
			BrotliWriteBits(max_bits, uint64(symbols[3]), storage_ix, storage)

			/* tree-select */
			var tmp int
			if depth[symbols[0]] == 1 {
				tmp = 1
			} else {
				tmp = 0
			}
			BrotliWriteBits(1, uint64(tmp), storage_ix, storage)
		}
	} else {
		var previous_value byte = 8
		var i uint

		/* Complex Huffman Tree */
		StoreStaticCodeLengthCode(storage_ix, storage)

		/* Actual RLE coding. */
		for i = 0; i < length; {
			var value byte = depth[i]
			var reps uint = 1
			var k uint
			for k = i + 1; k < length && depth[k] == value; k++ {
				reps++
			}

			i += reps
			if value == 0 {
				BrotliWriteBits(uint(kZeroRepsDepth[reps]), kZeroRepsBits[reps], storage_ix, storage)
			} else {
				if previous_value != value {
					BrotliWriteBits(uint(kCodeLengthDepth[value]), uint64(kCodeLengthBits[value]), storage_ix, storage)
					reps--
				}

				if reps < 3 {
					for reps != 0 {
						reps--
						BrotliWriteBits(uint(kCodeLengthDepth[value]), uint64(kCodeLengthBits[value]), storage_ix, storage)
					}
				} else {
					reps -= 3
					BrotliWriteBits(uint(kNonZeroRepsDepth[reps]), kNonZeroRepsBits[reps], storage_ix, storage)
				}

				previous_value = value
			}
		}
	}
}

func IndexOf(v []byte, v_size uint, value byte) uint {
	var i uint = 0
	for ; i < v_size; i++ {
		if v[i] == value {
			return i
		}
	}

	return i
}

func MoveToFront(v []byte, index uint) {
	var value byte = v[index]
	var i uint
	for i = index; i != 0; i-- {
		v[i] = v[i-1]
	}

	v[0] = value
}

func MoveToFrontTransform(v_in []uint32, v_size uint, v_out []uint32) {
	var i uint
	var mtf [256]byte
	var max_value uint32
	if v_size == 0 {
		return
	}

	max_value = v_in[0]
	for i = 1; i < v_size; i++ {
		if v_in[i] > max_value {
			max_value = v_in[i]
		}
	}

	assert(max_value < 256)
	for i = 0; uint32(i) <= max_value; i++ {
		mtf[i] = byte(i)
	}
	{
		var mtf_size uint = uint(max_value + 1)
		for i = 0; i < v_size; i++ {
			var index uint = IndexOf(mtf[:], mtf_size, byte(v_in[i]))
			assert(index < mtf_size)
			v_out[i] = uint32(index)
			MoveToFront(mtf[:], index)
		}
	}
}

/* Finds runs of zeros in v[0..in_size) and replaces them with a prefix code of
   the run length plus extra bits (lower 9 bits is the prefix code and the rest
   are the extra bits). Non-zero values in v[] are shifted by
   *max_length_prefix. Will not create prefix codes bigger than the initial
   value of *max_run_length_prefix. The prefix code of run length L is simply
   Log2Floor(L) and the number of extra bits is the same as the prefix code. */
func RunLengthCodeZeros(in_size uint, v []uint32, out_size *uint, max_run_length_prefix *uint32) {
	var max_reps uint32 = 0
	var i uint
	var max_prefix uint32
	for i = 0; i < in_size; {
		var reps uint32 = 0
		for ; i < in_size && v[i] != 0; i++ {
		}
		for ; i < in_size && v[i] == 0; i++ {
			reps++
		}

		max_reps = brotli_max_uint32_t(reps, max_reps)
	}

	if max_reps > 0 {
		max_prefix = Log2FloorNonZero(uint(max_reps))
	} else {
		max_prefix = 0
	}
	max_prefix = brotli_min_uint32_t(max_prefix, *max_run_length_prefix)
	*max_run_length_prefix = max_prefix
	*out_size = 0
	for i = 0; i < in_size; {
		assert(*out_size <= i)
		if v[i] != 0 {
			v[*out_size] = v[i] + *max_run_length_prefix
			i++
			(*out_size)++
		} else {
			var reps uint32 = 1
			var k uint
			for k = i + 1; k < in_size && v[k] == 0; k++ {
				reps++
			}

			i += uint(reps)
			for reps != 0 {
				if reps < 2<<max_prefix {
					var run_length_prefix uint32 = Log2FloorNonZero(uint(reps))
					var extra_bits uint32 = reps - (1 << run_length_prefix)
					v[*out_size] = run_length_prefix + (extra_bits << 9)
					(*out_size)++
					break
				} else {
					var extra_bits uint32 = (1 << max_prefix) - 1
					v[*out_size] = max_prefix + (extra_bits << 9)
					reps -= (2 << max_prefix) - 1
					(*out_size)++
				}
			}
		}
	}
}

const SYMBOL_BITS = 9

var EncodeContextMap_kSymbolMask uint32 = (1 << SYMBOL_BITS) - 1

func EncodeContextMap(context_map []uint32, context_map_size uint, num_clusters uint, tree []HuffmanTree, storage_ix *uint, storage []byte) {
	var i uint
	var rle_symbols []uint32
	var max_run_length_prefix uint32 = 6
	var num_rle_symbols uint = 0
	var histogram [BROTLI_MAX_CONTEXT_MAP_SYMBOLS]uint32
	var depths [BROTLI_MAX_CONTEXT_MAP_SYMBOLS]byte
	var bits [BROTLI_MAX_CONTEXT_MAP_SYMBOLS]uint16

	StoreVarLenUint8(num_clusters-1, storage_ix, storage)

	if num_clusters == 1 {
		return
	}

	rle_symbols = make([]uint32, context_map_size)
	MoveToFrontTransform(context_map, context_map_size, rle_symbols)
	RunLengthCodeZeros(context_map_size, rle_symbols, &num_rle_symbols, &max_run_length_prefix)
	histogram = [BROTLI_MAX_CONTEXT_MAP_SYMBOLS]uint32{}
	for i = 0; i < num_rle_symbols; i++ {
		histogram[rle_symbols[i]&EncodeContextMap_kSymbolMask]++
	}
	{
		var use_rle bool = (max_run_length_prefix > 0)
		BrotliWriteSingleBit(use_rle, storage_ix, storage)
		if use_rle {
			BrotliWriteBits(4, uint64(max_run_length_prefix)-1, storage_ix, storage)
		}
	}

	BuildAndStoreHuffmanTree(histogram[:], uint(uint32(num_clusters)+max_run_length_prefix), uint(uint32(num_clusters)+max_run_length_prefix), tree, depths[:], bits[:], storage_ix, storage)
	for i = 0; i < num_rle_symbols; i++ {
		var rle_symbol uint32 = rle_symbols[i] & EncodeContextMap_kSymbolMask
		var extra_bits_val uint32 = rle_symbols[i] >> SYMBOL_BITS
		BrotliWriteBits(uint(depths[rle_symbol]), uint64(bits[rle_symbol]), storage_ix, storage)
		if rle_symbol > 0 && rle_symbol <= max_run_length_prefix {
			BrotliWriteBits(uint(rle_symbol), uint64(extra_bits_val), storage_ix, storage)
		}
	}

	BrotliWriteBits(1, 1, storage_ix, storage) /* use move-to-front */
	rle_symbols = nil
}

/* Stores the block switch command with index block_ix to the bit stream. */
func StoreBlockSwitch(code *BlockSplitCode, block_len uint32, block_type byte, is_first_block bool, storage_ix *uint, storage []byte) {
	var typecode uint = NextBlockTypeCode(&code.type_code_calculator, block_type)
	var lencode uint
	var len_nextra uint32
	var len_extra uint32
	if !is_first_block {
		BrotliWriteBits(uint(code.type_depths[typecode]), uint64(code.type_bits[typecode]), storage_ix, storage)
	}

	GetBlockLengthPrefixCode(block_len, &lencode, &len_nextra, &len_extra)

	BrotliWriteBits(uint(code.length_depths[lencode]), uint64(code.length_bits[lencode]), storage_ix, storage)
	BrotliWriteBits(uint(len_nextra), uint64(len_extra), storage_ix, storage)
}

/* Builds a BlockSplitCode data structure from the block split given by the
   vector of block types and block lengths and stores it to the bit stream. */
func BuildAndStoreBlockSplitCode(types []byte, lengths []uint32, num_blocks uint, num_types uint, tree []HuffmanTree, code *BlockSplitCode, storage_ix *uint, storage []byte) {
	var type_histo [BROTLI_MAX_BLOCK_TYPE_SYMBOLS]uint32
	var length_histo [BROTLI_NUM_BLOCK_LEN_SYMBOLS]uint32
	var i uint
	var type_code_calculator BlockTypeCodeCalculator
	for i := 0; i < int(num_types+2); i++ {
		type_histo[i] = 0
	}
	length_histo = [BROTLI_NUM_BLOCK_LEN_SYMBOLS]uint32{}
	InitBlockTypeCodeCalculator(&type_code_calculator)
	for i = 0; i < num_blocks; i++ {
		var type_code uint = NextBlockTypeCode(&type_code_calculator, types[i])
		if i != 0 {
			type_histo[type_code]++
		}
		length_histo[BlockLengthPrefixCode(lengths[i])]++
	}

	StoreVarLenUint8(num_types-1, storage_ix, storage)
	if num_types > 1 { /* TODO: else? could StoreBlockSwitch occur? */
		BuildAndStoreHuffmanTree(type_histo[0:], num_types+2, num_types+2, tree, code.type_depths[0:], code.type_bits[0:], storage_ix, storage)
		BuildAndStoreHuffmanTree(length_histo[0:], BROTLI_NUM_BLOCK_LEN_SYMBOLS, BROTLI_NUM_BLOCK_LEN_SYMBOLS, tree, code.length_depths[0:], code.length_bits[0:], storage_ix, storage)
		StoreBlockSwitch(code, lengths[0], types[0], true, storage_ix, storage)
	}
}

/* Stores a context map where the histogram type is always the block type. */
func StoreTrivialContextMap(num_types uint, context_bits uint, tree []HuffmanTree, storage_ix *uint, storage []byte) {
	StoreVarLenUint8(num_types-1, storage_ix, storage)
	if num_types > 1 {
		var repeat_code uint = context_bits - 1
		var repeat_bits uint = (1 << repeat_code) - 1
		var alphabet_size uint = num_types + repeat_code
		var histogram [BROTLI_MAX_CONTEXT_MAP_SYMBOLS]uint32
		var depths [BROTLI_MAX_CONTEXT_MAP_SYMBOLS]byte
		var bits [BROTLI_MAX_CONTEXT_MAP_SYMBOLS]uint16
		var i uint
		for i := 0; i < int(alphabet_size); i++ {
			histogram[i] = 0
		}

		/* Write RLEMAX. */
		BrotliWriteBits(1, 1, storage_ix, storage)

		BrotliWriteBits(4, uint64(repeat_code)-1, storage_ix, storage)
		histogram[repeat_code] = uint32(num_types)
		histogram[0] = 1
		for i = context_bits; i < alphabet_size; i++ {
			histogram[i] = 1
		}

		BuildAndStoreHuffmanTree(histogram[:], alphabet_size, alphabet_size, tree, depths[:], bits[:], storage_ix, storage)
		for i = 0; i < num_types; i++ {
			var tmp uint
			if i == 0 {
				tmp = 0
			} else {
				tmp = i + context_bits - 1
			}
			var code uint = tmp
			BrotliWriteBits(uint(depths[code]), uint64(bits[code]), storage_ix, storage)
			BrotliWriteBits(uint(depths[repeat_code]), uint64(bits[repeat_code]), storage_ix, storage)
			BrotliWriteBits(repeat_code, uint64(repeat_bits), storage_ix, storage)
		}

		/* Write IMTF (inverse-move-to-front) bit. */
		BrotliWriteBits(1, 1, storage_ix, storage)
	}
}

/* Manages the encoding of one block category (literal, command or distance). */
type BlockEncoder struct {
	histogram_length_ uint
	num_block_types_  uint
	block_types_      []byte
	block_lengths_    []uint32
	num_blocks_       uint
	block_split_code_ BlockSplitCode
	block_ix_         uint
	block_len_        uint
	entropy_ix_       uint
	depths_           []byte
	bits_             []uint16
}

func InitBlockEncoder(self *BlockEncoder, histogram_length uint, num_block_types uint, block_types []byte, block_lengths []uint32, num_blocks uint) {
	self.histogram_length_ = histogram_length
	self.num_block_types_ = num_block_types
	self.block_types_ = block_types
	self.block_lengths_ = block_lengths
	self.num_blocks_ = num_blocks
	InitBlockTypeCodeCalculator(&self.block_split_code_.type_code_calculator)
	self.block_ix_ = 0
	if num_blocks == 0 {
		self.block_len_ = 0
	} else {
		self.block_len_ = uint(block_lengths[0])
	}
	self.entropy_ix_ = 0
	self.depths_ = nil
	self.bits_ = nil
}

func CleanupBlockEncoder(self *BlockEncoder) {
	self.depths_ = nil
	self.bits_ = nil
}

/* Creates entropy codes of block lengths and block types and stores them
   to the bit stream. */
func BuildAndStoreBlockSwitchEntropyCodes(self *BlockEncoder, tree []HuffmanTree, storage_ix *uint, storage []byte) {
	BuildAndStoreBlockSplitCode(self.block_types_, self.block_lengths_, self.num_blocks_, self.num_block_types_, tree, &self.block_split_code_, storage_ix, storage)
}

/* Stores the next symbol with the entropy code of the current block type.
   Updates the block type and block length at block boundaries. */
func StoreSymbol(self *BlockEncoder, symbol uint, storage_ix *uint, storage []byte) {
	if self.block_len_ == 0 {
		self.block_ix_++
		var block_ix uint = self.block_ix_
		var block_len uint32 = self.block_lengths_[block_ix]
		var block_type byte = self.block_types_[block_ix]
		self.block_len_ = uint(block_len)
		self.entropy_ix_ = uint(block_type) * self.histogram_length_
		StoreBlockSwitch(&self.block_split_code_, block_len, block_type, false, storage_ix, storage)
	}

	self.block_len_--
	{
		var ix uint = self.entropy_ix_ + symbol
		BrotliWriteBits(uint(self.depths_[ix]), uint64(self.bits_[ix]), storage_ix, storage)
	}
}

/* Stores the next symbol with the entropy code of the current block type and
   context value.
   Updates the block type and block length at block boundaries. */
func StoreSymbolWithContext(self *BlockEncoder, symbol uint, context uint, context_map []uint32, storage_ix *uint, storage []byte, context_bits uint) {
	if self.block_len_ == 0 {
		self.block_ix_++
		var block_ix uint = self.block_ix_
		var block_len uint32 = self.block_lengths_[block_ix]
		var block_type byte = self.block_types_[block_ix]
		self.block_len_ = uint(block_len)
		self.entropy_ix_ = uint(block_type) << context_bits
		StoreBlockSwitch(&self.block_split_code_, block_len, block_type, false, storage_ix, storage)
	}

	self.block_len_--
	{
		var histo_ix uint = uint(context_map[self.entropy_ix_+context])
		var ix uint = histo_ix*self.histogram_length_ + symbol
		BrotliWriteBits(uint(self.depths_[ix]), uint64(self.bits_[ix]), storage_ix, storage)
	}
}

func BuildAndStoreEntropyCodesLiteral(self *BlockEncoder, histograms []HistogramLiteral, histograms_size uint, alphabet_size uint, tree []HuffmanTree, storage_ix *uint, storage []byte) {
	var table_size uint = histograms_size * self.histogram_length_
	self.depths_ = make([]byte, table_size)
	self.bits_ = make([]uint16, table_size)
	{
		var i uint
		for i = 0; i < histograms_size; i++ {
			var ix uint = i * self.histogram_length_
			BuildAndStoreHuffmanTree(histograms[i].data_[0:], self.histogram_length_, alphabet_size, tree, self.depths_[ix:], self.bits_[ix:], storage_ix, storage)
		}
	}
}

func BuildAndStoreEntropyCodesCommand(self *BlockEncoder, histograms []HistogramCommand, histograms_size uint, alphabet_size uint, tree []HuffmanTree, storage_ix *uint, storage []byte) {
	var table_size uint = histograms_size * self.histogram_length_
	self.depths_ = make([]byte, table_size)
	self.bits_ = make([]uint16, table_size)
	{
		var i uint
		for i = 0; i < histograms_size; i++ {
			var ix uint = i * self.histogram_length_
			BuildAndStoreHuffmanTree(histograms[i].data_[0:], self.histogram_length_, alphabet_size, tree, self.depths_[ix:], self.bits_[ix:], storage_ix, storage)
		}
	}
}

func BuildAndStoreEntropyCodesDistance(self *BlockEncoder, histograms []HistogramDistance, histograms_size uint, alphabet_size uint, tree []HuffmanTree, storage_ix *uint, storage []byte) {
	var table_size uint = histograms_size * self.histogram_length_
	self.depths_ = make([]byte, table_size)
	self.bits_ = make([]uint16, table_size)
	{
		var i uint
		for i = 0; i < histograms_size; i++ {
			var ix uint = i * self.histogram_length_
			BuildAndStoreHuffmanTree(histograms[i].data_[0:], self.histogram_length_, alphabet_size, tree, self.depths_[ix:], self.bits_[ix:], storage_ix, storage)
		}
	}
}

func JumpToByteBoundary(storage_ix *uint, storage []byte) {
	*storage_ix = (*storage_ix + 7) &^ 7
	storage[*storage_ix>>3] = 0
}

func BrotliStoreMetaBlock(input []byte, start_pos uint, length uint, mask uint, prev_byte byte, prev_byte2 byte, is_last bool, params *BrotliEncoderParams, literal_context_mode int, commands []Command, n_commands uint, mb *MetaBlockSplit, storage_ix *uint, storage []byte) {
	var pos uint = start_pos
	var i uint
	var num_distance_symbols uint32 = params.dist.alphabet_size
	var num_effective_distance_symbols uint32 = num_distance_symbols
	var tree []HuffmanTree
	var literal_context_lut ContextLut = BROTLI_CONTEXT_LUT(literal_context_mode)
	var literal_enc BlockEncoder
	var command_enc BlockEncoder
	var distance_enc BlockEncoder
	var dist *BrotliDistanceParams = &params.dist
	if params.large_window && num_effective_distance_symbols > BROTLI_NUM_HISTOGRAM_DISTANCE_SYMBOLS {
		num_effective_distance_symbols = BROTLI_NUM_HISTOGRAM_DISTANCE_SYMBOLS
	}

	StoreCompressedMetaBlockHeader(is_last, length, storage_ix, storage)

	tree = make([]HuffmanTree, MAX_HUFFMAN_TREE_SIZE)
	InitBlockEncoder(&literal_enc, BROTLI_NUM_LITERAL_SYMBOLS, mb.literal_split.num_types, mb.literal_split.types, mb.literal_split.lengths, mb.literal_split.num_blocks)
	InitBlockEncoder(&command_enc, BROTLI_NUM_COMMAND_SYMBOLS, mb.command_split.num_types, mb.command_split.types, mb.command_split.lengths, mb.command_split.num_blocks)
	InitBlockEncoder(&distance_enc, uint(num_effective_distance_symbols), mb.distance_split.num_types, mb.distance_split.types, mb.distance_split.lengths, mb.distance_split.num_blocks)

	BuildAndStoreBlockSwitchEntropyCodes(&literal_enc, tree, storage_ix, storage)
	BuildAndStoreBlockSwitchEntropyCodes(&command_enc, tree, storage_ix, storage)
	BuildAndStoreBlockSwitchEntropyCodes(&distance_enc, tree, storage_ix, storage)

	BrotliWriteBits(2, uint64(dist.distance_postfix_bits), storage_ix, storage)
	BrotliWriteBits(4, uint64(dist.num_direct_distance_codes)>>dist.distance_postfix_bits, storage_ix, storage)
	for i = 0; i < mb.literal_split.num_types; i++ {
		BrotliWriteBits(2, uint64(literal_context_mode), storage_ix, storage)
	}

	if mb.literal_context_map_size == 0 {
		StoreTrivialContextMap(mb.literal_histograms_size, BROTLI_LITERAL_CONTEXT_BITS, tree, storage_ix, storage)
	} else {
		EncodeContextMap(mb.literal_context_map, mb.literal_context_map_size, mb.literal_histograms_size, tree, storage_ix, storage)
	}

	if mb.distance_context_map_size == 0 {
		StoreTrivialContextMap(mb.distance_histograms_size, BROTLI_DISTANCE_CONTEXT_BITS, tree, storage_ix, storage)
	} else {
		EncodeContextMap(mb.distance_context_map, mb.distance_context_map_size, mb.distance_histograms_size, tree, storage_ix, storage)
	}

	BuildAndStoreEntropyCodesLiteral(&literal_enc, mb.literal_histograms, mb.literal_histograms_size, BROTLI_NUM_LITERAL_SYMBOLS, tree, storage_ix, storage)
	BuildAndStoreEntropyCodesCommand(&command_enc, mb.command_histograms, mb.command_histograms_size, BROTLI_NUM_COMMAND_SYMBOLS, tree, storage_ix, storage)
	BuildAndStoreEntropyCodesDistance(&distance_enc, mb.distance_histograms, mb.distance_histograms_size, uint(num_distance_symbols), tree, storage_ix, storage)
	tree = nil

	for i = 0; i < n_commands; i++ {
		var cmd Command = commands[i]
		var cmd_code uint = uint(cmd.cmd_prefix_)
		StoreSymbol(&command_enc, cmd_code, storage_ix, storage)
		StoreCommandExtra(&cmd, storage_ix, storage)
		if mb.literal_context_map_size == 0 {
			var j uint
			for j = uint(cmd.insert_len_); j != 0; j-- {
				StoreSymbol(&literal_enc, uint(input[pos&mask]), storage_ix, storage)
				pos++
			}
		} else {
			var j uint
			for j = uint(cmd.insert_len_); j != 0; j-- {
				var context uint = uint(BROTLI_CONTEXT(prev_byte, prev_byte2, literal_context_lut))
				var literal byte = input[pos&mask]
				StoreSymbolWithContext(&literal_enc, uint(literal), context, mb.literal_context_map, storage_ix, storage, BROTLI_LITERAL_CONTEXT_BITS)
				prev_byte2 = prev_byte
				prev_byte = literal
				pos++
			}
		}

		pos += uint(CommandCopyLen(&cmd))
		if CommandCopyLen(&cmd) != 0 {
			prev_byte2 = input[(pos-2)&mask]
			prev_byte = input[(pos-1)&mask]
			if cmd.cmd_prefix_ >= 128 {
				var dist_code uint = uint(cmd.dist_prefix_) & 0x3FF
				var distnumextra uint32 = uint32(cmd.dist_prefix_) >> 10
				var distextra uint64 = uint64(cmd.dist_extra_)
				if mb.distance_context_map_size == 0 {
					StoreSymbol(&distance_enc, dist_code, storage_ix, storage)
				} else {
					var context uint = uint(CommandDistanceContext(&cmd))
					StoreSymbolWithContext(&distance_enc, dist_code, context, mb.distance_context_map, storage_ix, storage, BROTLI_DISTANCE_CONTEXT_BITS)
				}

				BrotliWriteBits(uint(distnumextra), distextra, storage_ix, storage)
			}
		}
	}

	CleanupBlockEncoder(&distance_enc)
	CleanupBlockEncoder(&command_enc)
	CleanupBlockEncoder(&literal_enc)
	if is_last {
		JumpToByteBoundary(storage_ix, storage)
	}
}

func BuildHistograms(input []byte, start_pos uint, mask uint, commands []Command, n_commands uint, lit_histo *HistogramLiteral, cmd_histo *HistogramCommand, dist_histo *HistogramDistance) {
	var pos uint = start_pos
	var i uint
	for i = 0; i < n_commands; i++ {
		var cmd Command = commands[i]
		var j uint
		HistogramAddCommand(cmd_histo, uint(cmd.cmd_prefix_))
		for j = uint(cmd.insert_len_); j != 0; j-- {
			HistogramAddLiteral(lit_histo, uint(input[pos&mask]))
			pos++
		}

		pos += uint(CommandCopyLen(&cmd))
		if CommandCopyLen(&cmd) != 0 && cmd.cmd_prefix_ >= 128 {
			HistogramAddDistance(dist_histo, uint(cmd.dist_prefix_)&0x3FF)
		}
	}
}

func StoreDataWithHuffmanCodes(input []byte, start_pos uint, mask uint, commands []Command, n_commands uint, lit_depth []byte, lit_bits []uint16, cmd_depth []byte, cmd_bits []uint16, dist_depth []byte, dist_bits []uint16, storage_ix *uint, storage []byte) {
	var pos uint = start_pos
	var i uint
	for i = 0; i < n_commands; i++ {
		var cmd Command = commands[i]
		var cmd_code uint = uint(cmd.cmd_prefix_)
		var j uint
		BrotliWriteBits(uint(cmd_depth[cmd_code]), uint64(cmd_bits[cmd_code]), storage_ix, storage)
		StoreCommandExtra(&cmd, storage_ix, storage)
		for j = uint(cmd.insert_len_); j != 0; j-- {
			var literal byte = input[pos&mask]
			BrotliWriteBits(uint(lit_depth[literal]), uint64(lit_bits[literal]), storage_ix, storage)
			pos++
		}

		pos += uint(CommandCopyLen(&cmd))
		if CommandCopyLen(&cmd) != 0 && cmd.cmd_prefix_ >= 128 {
			var dist_code uint = uint(cmd.dist_prefix_) & 0x3FF
			var distnumextra uint32 = uint32(cmd.dist_prefix_) >> 10
			var distextra uint32 = cmd.dist_extra_
			BrotliWriteBits(uint(dist_depth[dist_code]), uint64(dist_bits[dist_code]), storage_ix, storage)
			BrotliWriteBits(uint(distnumextra), uint64(distextra), storage_ix, storage)
		}
	}
}

func BrotliStoreMetaBlockTrivial(input []byte, start_pos uint, length uint, mask uint, is_last bool, params *BrotliEncoderParams, commands []Command, n_commands uint, storage_ix *uint, storage []byte) {
	var lit_histo HistogramLiteral
	var cmd_histo HistogramCommand
	var dist_histo HistogramDistance
	var lit_depth [BROTLI_NUM_LITERAL_SYMBOLS]byte
	var lit_bits [BROTLI_NUM_LITERAL_SYMBOLS]uint16
	var cmd_depth [BROTLI_NUM_COMMAND_SYMBOLS]byte
	var cmd_bits [BROTLI_NUM_COMMAND_SYMBOLS]uint16
	var dist_depth [MAX_SIMPLE_DISTANCE_ALPHABET_SIZE]byte
	var dist_bits [MAX_SIMPLE_DISTANCE_ALPHABET_SIZE]uint16
	var tree []HuffmanTree
	var num_distance_symbols uint32 = params.dist.alphabet_size

	StoreCompressedMetaBlockHeader(is_last, length, storage_ix, storage)

	HistogramClearLiteral(&lit_histo)
	HistogramClearCommand(&cmd_histo)
	HistogramClearDistance(&dist_histo)

	BuildHistograms(input, start_pos, mask, commands, n_commands, &lit_histo, &cmd_histo, &dist_histo)

	BrotliWriteBits(13, 0, storage_ix, storage)

	tree = make([]HuffmanTree, MAX_HUFFMAN_TREE_SIZE)
	BuildAndStoreHuffmanTree(lit_histo.data_[:], BROTLI_NUM_LITERAL_SYMBOLS, BROTLI_NUM_LITERAL_SYMBOLS, tree, lit_depth[:], lit_bits[:], storage_ix, storage)
	BuildAndStoreHuffmanTree(cmd_histo.data_[:], BROTLI_NUM_COMMAND_SYMBOLS, BROTLI_NUM_COMMAND_SYMBOLS, tree, cmd_depth[:], cmd_bits[:], storage_ix, storage)
	BuildAndStoreHuffmanTree(dist_histo.data_[:], MAX_SIMPLE_DISTANCE_ALPHABET_SIZE, uint(num_distance_symbols), tree, dist_depth[:], dist_bits[:], storage_ix, storage)
	tree = nil
	StoreDataWithHuffmanCodes(input, start_pos, mask, commands, n_commands, lit_depth[:], lit_bits[:], cmd_depth[:], cmd_bits[:], dist_depth[:], dist_bits[:], storage_ix, storage)
	if is_last {
		JumpToByteBoundary(storage_ix, storage)
	}
}

func BrotliStoreMetaBlockFast(input []byte, start_pos uint, length uint, mask uint, is_last bool, params *BrotliEncoderParams, commands []Command, n_commands uint, storage_ix *uint, storage []byte) {
	var num_distance_symbols uint32 = params.dist.alphabet_size
	var distance_alphabet_bits uint32 = Log2FloorNonZero(uint(num_distance_symbols-1)) + 1

	StoreCompressedMetaBlockHeader(is_last, length, storage_ix, storage)

	BrotliWriteBits(13, 0, storage_ix, storage)

	if n_commands <= 128 {
		var histogram = [BROTLI_NUM_LITERAL_SYMBOLS]uint32{0}
		var pos uint = start_pos
		var num_literals uint = 0
		var i uint
		var lit_depth [BROTLI_NUM_LITERAL_SYMBOLS]byte
		var lit_bits [BROTLI_NUM_LITERAL_SYMBOLS]uint16
		for i = 0; i < n_commands; i++ {
			var cmd Command = commands[i]
			var j uint
			for j = uint(cmd.insert_len_); j != 0; j-- {
				histogram[input[pos&mask]]++
				pos++
			}

			num_literals += uint(cmd.insert_len_)
			pos += uint(CommandCopyLen(&cmd))
		}

		BrotliBuildAndStoreHuffmanTreeFast(histogram[:], num_literals, /* max_bits = */
			8, lit_depth[:], lit_bits[:], storage_ix, storage)

		StoreStaticCommandHuffmanTree(storage_ix, storage)
		StoreStaticDistanceHuffmanTree(storage_ix, storage)
		StoreDataWithHuffmanCodes(input, start_pos, mask, commands, n_commands, lit_depth[:], lit_bits[:], kStaticCommandCodeDepth[:], kStaticCommandCodeBits[:], kStaticDistanceCodeDepth[:], kStaticDistanceCodeBits[:], storage_ix, storage)
	} else {
		var lit_histo HistogramLiteral
		var cmd_histo HistogramCommand
		var dist_histo HistogramDistance
		var lit_depth [BROTLI_NUM_LITERAL_SYMBOLS]byte
		var lit_bits [BROTLI_NUM_LITERAL_SYMBOLS]uint16
		var cmd_depth [BROTLI_NUM_COMMAND_SYMBOLS]byte
		var cmd_bits [BROTLI_NUM_COMMAND_SYMBOLS]uint16
		var dist_depth [MAX_SIMPLE_DISTANCE_ALPHABET_SIZE]byte
		var dist_bits [MAX_SIMPLE_DISTANCE_ALPHABET_SIZE]uint16
		HistogramClearLiteral(&lit_histo)
		HistogramClearCommand(&cmd_histo)
		HistogramClearDistance(&dist_histo)
		BuildHistograms(input, start_pos, mask, commands, n_commands, &lit_histo, &cmd_histo, &dist_histo)
		BrotliBuildAndStoreHuffmanTreeFast(lit_histo.data_[:], lit_histo.total_count_, /* max_bits = */
			8, lit_depth[:], lit_bits[:], storage_ix, storage)

		BrotliBuildAndStoreHuffmanTreeFast(cmd_histo.data_[:], cmd_histo.total_count_, /* max_bits = */
			10, cmd_depth[:], cmd_bits[:], storage_ix, storage)

		BrotliBuildAndStoreHuffmanTreeFast(dist_histo.data_[:], dist_histo.total_count_, /* max_bits = */
			uint(distance_alphabet_bits), dist_depth[:], dist_bits[:], storage_ix, storage)

		StoreDataWithHuffmanCodes(input, start_pos, mask, commands, n_commands, lit_depth[:], lit_bits[:], cmd_depth[:], cmd_bits[:], dist_depth[:], dist_bits[:], storage_ix, storage)
	}

	if is_last {
		JumpToByteBoundary(storage_ix, storage)
	}
}

/* This is for storing uncompressed blocks (simple raw storage of
   bytes-as-bytes). */
func BrotliStoreUncompressedMetaBlock(is_final_block bool, input []byte, position uint, mask uint, len uint, storage_ix *uint, storage []byte) {
	var masked_pos uint = position & mask
	BrotliStoreUncompressedMetaBlockHeader(uint(len), storage_ix, storage)
	JumpToByteBoundary(storage_ix, storage)

	if masked_pos+len > mask+1 {
		var len1 uint = mask + 1 - masked_pos
		copy(storage[*storage_ix>>3:], input[masked_pos:][:len1])
		*storage_ix += len1 << 3
		len -= len1
		masked_pos = 0
	}

	copy(storage[*storage_ix>>3:], input[masked_pos:][:len])
	*storage_ix += uint(len << 3)

	/* We need to clear the next 4 bytes to continue to be
	   compatible with BrotliWriteBits. */
	BrotliWriteBitsPrepareStorage(*storage_ix, storage)

	/* Since the uncompressed block itself may not be the final block, add an
	   empty one after this. */
	if is_final_block {
		BrotliWriteBits(1, 1, storage_ix, storage) /* islast */
		BrotliWriteBits(1, 1, storage_ix, storage) /* isempty */
		JumpToByteBoundary(storage_ix, storage)
	}
}
