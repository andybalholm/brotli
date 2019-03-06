package brotli

/* Dictionary data (words and transforms) for 1 possible context */
type BrotliEncoderDictionary struct {
	words                 *BrotliDictionary
	cutoffTransformsCount uint32
	cutoffTransforms      uint64
	hash_table            []uint16
	buckets               []uint16
	dict_words            []DictWord
}

func BrotliInitEncoderDictionary(dict *BrotliEncoderDictionary) {
	dict.words = BrotliGetDictionary()

	dict.hash_table = kStaticDictionaryHash[:]
	dict.buckets = kStaticDictionaryBuckets[:]
	dict.dict_words = kStaticDictionaryWords[:]

	dict.cutoffTransformsCount = kCutoffTransformsCount
	dict.cutoffTransforms = kCutoffTransforms
}
