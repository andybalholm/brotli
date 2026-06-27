package matchfinder

import (
	"os"
	"testing"
)

func BenchmarkTrio(b *testing.B) {
	data, err := os.ReadFile("../testdata/Isaac.Newton-Opticks.txt")
	if err != nil {
		b.Fatal(err)
	}

	const blockSize = 1 << 16
	trio := &Trio{MaxDistance: 1 << 20}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	var matches []Match
	for i := 0; i < b.N; i++ {
		trio.Reset()
		for off := 0; off < len(data); off += blockSize {
			end := min(off+blockSize, len(data))
			matches = trio.FindMatches(matches[:0], data[off:end])
		}
	}
}
