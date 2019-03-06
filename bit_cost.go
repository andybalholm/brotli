package brotli

/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Functions to estimate the bit cost of Huffman trees. */
/* Copyright 2013 Google Inc. All Rights Reserved.

   Distributed under MIT license.
   See file LICENSE for detail or copy at https://opensource.org/licenses/MIT
*/

/* Functions to estimate the bit cost of Huffman trees. */
func ShannonEntropy(population []uint32, size uint, total *uint) float64 {
	var sum uint = 0
	var retval float64 = 0
	var population_end []uint32 = population[size:]
	var p uint
	for -cap(population) < -cap(population_end) {
		p = uint(population[0])
		population = population[1:]
		sum += p
		retval -= float64(p) * FastLog2(p)
	}

	if sum != 0 {
		retval += float64(sum) * FastLog2(sum)
	}
	*total = sum
	return retval
}

func BitsEntropy(population []uint32, size uint) float64 {
	var sum uint
	var retval float64 = ShannonEntropy(population, size, &sum)
	if retval < float64(sum) {
		/* At least one bit per literal is needed. */
		retval = float64(sum)
	}

	return retval
}

var BrotliPopulationCostLiteral_kOneSymbolHistogramCost float64 = 12
var BrotliPopulationCostLiteral_kTwoSymbolHistogramCost float64 = 20
var BrotliPopulationCostLiteral_kThreeSymbolHistogramCost float64 = 28
var BrotliPopulationCostLiteral_kFourSymbolHistogramCost float64 = 37

func BrotliPopulationCostLiteral(histogram *HistogramLiteral) float64 {
	var data_size uint = HistogramDataSizeLiteral()
	var count int = 0
	var s [5]uint
	var bits float64 = 0.0
	var i uint
	if histogram.total_count_ == 0 {
		return BrotliPopulationCostLiteral_kOneSymbolHistogramCost
	}

	for i = 0; i < data_size; i++ {
		if histogram.data_[i] > 0 {
			s[count] = i
			count++
			if count > 4 {
				break
			}
		}
	}

	if count == 1 {
		return BrotliPopulationCostLiteral_kOneSymbolHistogramCost
	}

	if count == 2 {
		return BrotliPopulationCostLiteral_kTwoSymbolHistogramCost + float64(histogram.total_count_)
	}

	if count == 3 {
		var histo0 uint32 = histogram.data_[s[0]]
		var histo1 uint32 = histogram.data_[s[1]]
		var histo2 uint32 = histogram.data_[s[2]]
		var histomax uint32 = brotli_max_uint32_t(histo0, brotli_max_uint32_t(histo1, histo2))
		return BrotliPopulationCostLiteral_kThreeSymbolHistogramCost + 2*(float64(histo0)+float64(histo1)+float64(histo2)) - float64(histomax)
	}

	if count == 4 {
		var histo [4]uint32
		var h23 uint32
		var histomax uint32
		for i = 0; i < 4; i++ {
			histo[i] = histogram.data_[s[i]]
		}

		/* Sort */
		for i = 0; i < 4; i++ {
			var j uint
			for j = i + 1; j < 4; j++ {
				if histo[j] > histo[i] {
					var tmp uint32 = histo[j]
					histo[j] = histo[i]
					histo[i] = tmp
				}
			}
		}

		h23 = histo[2] + histo[3]
		histomax = brotli_max_uint32_t(h23, histo[0])
		return BrotliPopulationCostLiteral_kFourSymbolHistogramCost + 3*float64(h23) + 2*(float64(histo[0])+float64(histo[1])) - float64(histomax)
	}
	{
		var max_depth uint = 1
		var depth_histo = [BROTLI_CODE_LENGTH_CODES]uint32{0}
		/* In this loop we compute the entropy of the histogram and simultaneously
		   build a simplified histogram of the code length codes where we use the
		   zero repeat code 17, but we don't use the non-zero repeat code 16. */

		var log2total float64 = FastLog2(histogram.total_count_)
		for i = 0; i < data_size; {
			if histogram.data_[i] > 0 {
				var log2p float64 = log2total - FastLog2(uint(histogram.data_[i]))
				/* Compute -log2(P(symbol)) = -log2(count(symbol)/total_count) =
				   = log2(total_count) - log2(count(symbol)) */

				var depth uint = uint(log2p + 0.5)
				/* Approximate the bit depth by round(-log2(P(symbol))) */
				bits += float64(histogram.data_[i]) * log2p

				if depth > 15 {
					depth = 15
				}

				if depth > max_depth {
					max_depth = depth
				}

				depth_histo[depth]++
				i++
			} else {
				var reps uint32 = 1
				/* Compute the run length of zeros and add the appropriate number of 0
				   and 17 code length codes to the code length code histogram. */

				var k uint
				for k = i + 1; k < data_size && histogram.data_[k] == 0; k++ {
					reps++
				}

				i += uint(reps)
				if i == data_size {
					/* Don't add any cost for the last zero run, since these are encoded
					   only implicitly. */
					break
				}

				if reps < 3 {
					depth_histo[0] += reps
				} else {
					reps -= 2
					for reps > 0 {
						depth_histo[BROTLI_REPEAT_ZERO_CODE_LENGTH]++

						/* Add the 3 extra bits for the 17 code length code. */
						bits += 3

						reps >>= 3
					}
				}
			}
		}

		/* Add the estimated encoding cost of the code length code histogram. */
		bits += float64(18 + 2*max_depth)

		/* Add the entropy of the code length code histogram. */
		bits += BitsEntropy(depth_histo[:], BROTLI_CODE_LENGTH_CODES)
	}

	return bits
}

var BrotliPopulationCostCommand_kOneSymbolHistogramCost float64 = 12
var BrotliPopulationCostCommand_kTwoSymbolHistogramCost float64 = 20
var BrotliPopulationCostCommand_kThreeSymbolHistogramCost float64 = 28
var BrotliPopulationCostCommand_kFourSymbolHistogramCost float64 = 37

func BrotliPopulationCostCommand(histogram *HistogramCommand) float64 {
	var data_size uint = HistogramDataSizeCommand()
	var count int = 0
	var s [5]uint
	var bits float64 = 0.0
	var i uint
	if histogram.total_count_ == 0 {
		return BrotliPopulationCostCommand_kOneSymbolHistogramCost
	}

	for i = 0; i < data_size; i++ {
		if histogram.data_[i] > 0 {
			s[count] = i
			count++
			if count > 4 {
				break
			}
		}
	}

	if count == 1 {
		return BrotliPopulationCostCommand_kOneSymbolHistogramCost
	}

	if count == 2 {
		return BrotliPopulationCostCommand_kTwoSymbolHistogramCost + float64(histogram.total_count_)
	}

	if count == 3 {
		var histo0 uint32 = histogram.data_[s[0]]
		var histo1 uint32 = histogram.data_[s[1]]
		var histo2 uint32 = histogram.data_[s[2]]
		var histomax uint32 = brotli_max_uint32_t(histo0, brotli_max_uint32_t(histo1, histo2))
		return BrotliPopulationCostCommand_kThreeSymbolHistogramCost + 2*(float64(histo0)+float64(histo1)+float64(histo2)) - float64(histomax)
	}

	if count == 4 {
		var histo [4]uint32
		var h23 uint32
		var histomax uint32
		for i = 0; i < 4; i++ {
			histo[i] = histogram.data_[s[i]]
		}

		/* Sort */
		for i = 0; i < 4; i++ {
			var j uint
			for j = i + 1; j < 4; j++ {
				if histo[j] > histo[i] {
					var tmp uint32 = histo[j]
					histo[j] = histo[i]
					histo[i] = tmp
				}
			}
		}

		h23 = histo[2] + histo[3]
		histomax = brotli_max_uint32_t(h23, histo[0])
		return BrotliPopulationCostCommand_kFourSymbolHistogramCost + 3*float64(h23) + 2*(float64(histo[0])+float64(histo[1])) - float64(histomax)
	}
	{
		var max_depth uint = 1
		var depth_histo = [BROTLI_CODE_LENGTH_CODES]uint32{0}
		/* In this loop we compute the entropy of the histogram and simultaneously
		   build a simplified histogram of the code length codes where we use the
		   zero repeat code 17, but we don't use the non-zero repeat code 16. */

		var log2total float64 = FastLog2(histogram.total_count_)
		for i = 0; i < data_size; {
			if histogram.data_[i] > 0 {
				var log2p float64 = log2total - FastLog2(uint(histogram.data_[i]))
				/* Compute -log2(P(symbol)) = -log2(count(symbol)/total_count) =
				   = log2(total_count) - log2(count(symbol)) */

				var depth uint = uint(log2p + 0.5)
				/* Approximate the bit depth by round(-log2(P(symbol))) */
				bits += float64(histogram.data_[i]) * log2p

				if depth > 15 {
					depth = 15
				}

				if depth > max_depth {
					max_depth = depth
				}

				depth_histo[depth]++
				i++
			} else {
				var reps uint32 = 1
				/* Compute the run length of zeros and add the appropriate number of 0
				   and 17 code length codes to the code length code histogram. */

				var k uint
				for k = i + 1; k < data_size && histogram.data_[k] == 0; k++ {
					reps++
				}

				i += uint(reps)
				if i == data_size {
					/* Don't add any cost for the last zero run, since these are encoded
					   only implicitly. */
					break
				}

				if reps < 3 {
					depth_histo[0] += reps
				} else {
					reps -= 2
					for reps > 0 {
						depth_histo[BROTLI_REPEAT_ZERO_CODE_LENGTH]++

						/* Add the 3 extra bits for the 17 code length code. */
						bits += 3

						reps >>= 3
					}
				}
			}
		}

		/* Add the estimated encoding cost of the code length code histogram. */
		bits += float64(18 + 2*max_depth)

		/* Add the entropy of the code length code histogram. */
		bits += BitsEntropy(depth_histo[:], BROTLI_CODE_LENGTH_CODES)
	}

	return bits
}

var BrotliPopulationCostDistance_kOneSymbolHistogramCost float64 = 12
var BrotliPopulationCostDistance_kTwoSymbolHistogramCost float64 = 20
var BrotliPopulationCostDistance_kThreeSymbolHistogramCost float64 = 28
var BrotliPopulationCostDistance_kFourSymbolHistogramCost float64 = 37

func BrotliPopulationCostDistance(histogram *HistogramDistance) float64 {
	var data_size uint = HistogramDataSizeDistance()
	var count int = 0
	var s [5]uint
	var bits float64 = 0.0
	var i uint
	if histogram.total_count_ == 0 {
		return BrotliPopulationCostDistance_kOneSymbolHistogramCost
	}

	for i = 0; i < data_size; i++ {
		if histogram.data_[i] > 0 {
			s[count] = i
			count++
			if count > 4 {
				break
			}
		}
	}

	if count == 1 {
		return BrotliPopulationCostDistance_kOneSymbolHistogramCost
	}

	if count == 2 {
		return BrotliPopulationCostDistance_kTwoSymbolHistogramCost + float64(histogram.total_count_)
	}

	if count == 3 {
		var histo0 uint32 = histogram.data_[s[0]]
		var histo1 uint32 = histogram.data_[s[1]]
		var histo2 uint32 = histogram.data_[s[2]]
		var histomax uint32 = brotli_max_uint32_t(histo0, brotli_max_uint32_t(histo1, histo2))
		return BrotliPopulationCostDistance_kThreeSymbolHistogramCost + 2*(float64(histo0)+float64(histo1)+float64(histo2)) - float64(histomax)
	}

	if count == 4 {
		var histo [4]uint32
		var h23 uint32
		var histomax uint32
		for i = 0; i < 4; i++ {
			histo[i] = histogram.data_[s[i]]
		}

		/* Sort */
		for i = 0; i < 4; i++ {
			var j uint
			for j = i + 1; j < 4; j++ {
				if histo[j] > histo[i] {
					var tmp uint32 = histo[j]
					histo[j] = histo[i]
					histo[i] = tmp
				}
			}
		}

		h23 = histo[2] + histo[3]
		histomax = brotli_max_uint32_t(h23, histo[0])
		return BrotliPopulationCostDistance_kFourSymbolHistogramCost + 3*float64(h23) + 2*(float64(histo[0])+float64(histo[1])) - float64(histomax)
	}
	{
		var max_depth uint = 1
		var depth_histo = [BROTLI_CODE_LENGTH_CODES]uint32{0}
		/* In this loop we compute the entropy of the histogram and simultaneously
		   build a simplified histogram of the code length codes where we use the
		   zero repeat code 17, but we don't use the non-zero repeat code 16. */

		var log2total float64 = FastLog2(histogram.total_count_)
		for i = 0; i < data_size; {
			if histogram.data_[i] > 0 {
				var log2p float64 = log2total - FastLog2(uint(histogram.data_[i]))
				/* Compute -log2(P(symbol)) = -log2(count(symbol)/total_count) =
				   = log2(total_count) - log2(count(symbol)) */

				var depth uint = uint(log2p + 0.5)
				/* Approximate the bit depth by round(-log2(P(symbol))) */
				bits += float64(histogram.data_[i]) * log2p

				if depth > 15 {
					depth = 15
				}

				if depth > max_depth {
					max_depth = depth
				}

				depth_histo[depth]++
				i++
			} else {
				var reps uint32 = 1
				/* Compute the run length of zeros and add the appropriate number of 0
				   and 17 code length codes to the code length code histogram. */

				var k uint
				for k = i + 1; k < data_size && histogram.data_[k] == 0; k++ {
					reps++
				}

				i += uint(reps)
				if i == data_size {
					/* Don't add any cost for the last zero run, since these are encoded
					   only implicitly. */
					break
				}

				if reps < 3 {
					depth_histo[0] += reps
				} else {
					reps -= 2
					for reps > 0 {
						depth_histo[BROTLI_REPEAT_ZERO_CODE_LENGTH]++

						/* Add the 3 extra bits for the 17 code length code. */
						bits += 3

						reps >>= 3
					}
				}
			}
		}

		/* Add the estimated encoding cost of the code length code histogram. */
		bits += float64(18 + 2*max_depth)

		/* Add the entropy of the code length code histogram. */
		bits += BitsEntropy(depth_histo[:], BROTLI_CODE_LENGTH_CODES)
	}

	return bits
}
