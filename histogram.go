package brotli

/* The distance symbols effectively used by "Large Window Brotli" (32-bit). */
const BROTLI_NUM_HISTOGRAM_DISTANCE_SYMBOLS = 544

type HistogramLiteral struct {
	data_        [BROTLI_NUM_LITERAL_SYMBOLS]uint32
	total_count_ uint
	bit_cost_    float64
}

func HistogramClearLiteral(self *HistogramLiteral) {
	self.data_ = [BROTLI_NUM_LITERAL_SYMBOLS]uint32{}
	self.total_count_ = 0
	self.bit_cost_ = HUGE_VAL
}

func ClearHistogramsLiteral(array []HistogramLiteral, length uint) {
	var i uint
	for i = 0; i < length; i++ {
		HistogramClearLiteral(&array[i:][0])
	}
}

func HistogramAddLiteral(self *HistogramLiteral, val uint) {
	self.data_[val]++
	self.total_count_++
}

func HistogramAddVectorLiteral(self *HistogramLiteral, p []byte, n uint) {
	self.total_count_ += n
	n += 1
	for {
		n--
		if n == 0 {
			break
		}
		self.data_[p[0]]++
		p = p[1:]
	}
}

func HistogramAddHistogramLiteral(self *HistogramLiteral, v *HistogramLiteral) {
	var i uint
	self.total_count_ += v.total_count_
	for i = 0; i < BROTLI_NUM_LITERAL_SYMBOLS; i++ {
		self.data_[i] += v.data_[i]
	}
}

func HistogramDataSizeLiteral() uint {
	return BROTLI_NUM_LITERAL_SYMBOLS
}

type HistogramCommand struct {
	data_        [BROTLI_NUM_COMMAND_SYMBOLS]uint32
	total_count_ uint
	bit_cost_    float64
}

func HistogramClearCommand(self *HistogramCommand) {
	self.data_ = [BROTLI_NUM_COMMAND_SYMBOLS]uint32{}
	self.total_count_ = 0
	self.bit_cost_ = HUGE_VAL
}

func ClearHistogramsCommand(array []HistogramCommand, length uint) {
	var i uint
	for i = 0; i < length; i++ {
		HistogramClearCommand(&array[i:][0])
	}
}

func HistogramAddCommand(self *HistogramCommand, val uint) {
	self.data_[val]++
	self.total_count_++
}

func HistogramAddVectorCommand(self *HistogramCommand, p []uint16, n uint) {
	self.total_count_ += n
	n += 1
	for {
		n--
		if n == 0 {
			break
		}
		self.data_[p[0]]++
		p = p[1:]
	}
}

func HistogramAddHistogramCommand(self *HistogramCommand, v *HistogramCommand) {
	var i uint
	self.total_count_ += v.total_count_
	for i = 0; i < BROTLI_NUM_COMMAND_SYMBOLS; i++ {
		self.data_[i] += v.data_[i]
	}
}

func HistogramDataSizeCommand() uint {
	return BROTLI_NUM_COMMAND_SYMBOLS
}

type HistogramDistance struct {
	data_        [BROTLI_NUM_DISTANCE_SYMBOLS]uint32
	total_count_ uint
	bit_cost_    float64
}

func HistogramClearDistance(self *HistogramDistance) {
	self.data_ = [BROTLI_NUM_DISTANCE_SYMBOLS]uint32{}
	self.total_count_ = 0
	self.bit_cost_ = HUGE_VAL
}

func ClearHistogramsDistance(array []HistogramDistance, length uint) {
	var i uint
	for i = 0; i < length; i++ {
		HistogramClearDistance(&array[i:][0])
	}
}

func HistogramAddDistance(self *HistogramDistance, val uint) {
	self.data_[val]++
	self.total_count_++
}

func HistogramAddVectorDistance(self *HistogramDistance, p []uint16, n uint) {
	self.total_count_ += n
	n += 1
	for {
		n--
		if n == 0 {
			break
		}
		self.data_[p[0]]++
		p = p[1:]
	}
}

func HistogramAddHistogramDistance(self *HistogramDistance, v *HistogramDistance) {
	var i uint
	self.total_count_ += v.total_count_
	for i = 0; i < BROTLI_NUM_DISTANCE_SYMBOLS; i++ {
		self.data_[i] += v.data_[i]
	}
}

func HistogramDataSizeDistance() uint {
	return BROTLI_NUM_DISTANCE_SYMBOLS
}

type BlockSplitIterator struct {
	split_  *BlockSplit
	idx_    uint
	type_   uint
	length_ uint
}

func InitBlockSplitIterator(self *BlockSplitIterator, split *BlockSplit) {
	self.split_ = split
	self.idx_ = 0
	self.type_ = 0
	if split.lengths != nil {
		self.length_ = uint(split.lengths[0])
	} else {
		self.length_ = 0
	}
}

func BlockSplitIteratorNext(self *BlockSplitIterator) {
	if self.length_ == 0 {
		self.idx_++
		self.type_ = uint(self.split_.types[self.idx_])
		self.length_ = uint(self.split_.lengths[self.idx_])
	}

	self.length_--
}

func BrotliBuildHistogramsWithContext(cmds []Command, num_commands uint, literal_split *BlockSplit, insert_and_copy_split *BlockSplit, dist_split *BlockSplit, ringbuffer []byte, start_pos uint, mask uint, prev_byte byte, prev_byte2 byte, context_modes []int, literal_histograms []HistogramLiteral, insert_and_copy_histograms []HistogramCommand, copy_dist_histograms []HistogramDistance) {
	var pos uint = start_pos
	var literal_it BlockSplitIterator
	var insert_and_copy_it BlockSplitIterator
	var dist_it BlockSplitIterator
	var i uint

	InitBlockSplitIterator(&literal_it, literal_split)
	InitBlockSplitIterator(&insert_and_copy_it, insert_and_copy_split)
	InitBlockSplitIterator(&dist_it, dist_split)
	for i = 0; i < num_commands; i++ {
		var cmd *Command = &cmds[i]
		var j uint
		BlockSplitIteratorNext(&insert_and_copy_it)
		HistogramAddCommand(&insert_and_copy_histograms[insert_and_copy_it.type_], uint(cmd.cmd_prefix_))

		/* TODO: unwrap iterator blocks. */
		for j = uint(cmd.insert_len_); j != 0; j-- {
			var context uint
			BlockSplitIteratorNext(&literal_it)
			context = literal_it.type_
			if context_modes != nil {
				var lut ContextLut = BROTLI_CONTEXT_LUT(context_modes[context])
				context = (context << BROTLI_LITERAL_CONTEXT_BITS) + uint(BROTLI_CONTEXT(prev_byte, prev_byte2, lut))
			}

			HistogramAddLiteral(&literal_histograms[context], uint(ringbuffer[pos&mask]))
			prev_byte2 = prev_byte
			prev_byte = ringbuffer[pos&mask]
			pos++
		}

		pos += uint(CommandCopyLen(cmd))
		if CommandCopyLen(cmd) != 0 {
			prev_byte2 = ringbuffer[(pos-2)&mask]
			prev_byte = ringbuffer[(pos-1)&mask]
			if cmd.cmd_prefix_ >= 128 {
				var context uint
				BlockSplitIteratorNext(&dist_it)
				context = uint(uint32(dist_it.type_<<BROTLI_DISTANCE_CONTEXT_BITS) + CommandDistanceContext(cmd))
				HistogramAddDistance(&copy_dist_histograms[context], uint(cmd.dist_prefix_)&0x3FF)
			}
		}
	}
}
