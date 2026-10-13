package brotli

import (
	"bytes"
	"testing"
)

func TestFindMatchLengthWithLimitEdgeCases(t *testing.T) {
	// Test limit 0 does not panic
	s1 := []byte("hello world")
	s2 := []byte("hello world")
	if n := findMatchLengthWithLimit(s1, s2, 0); n != 0 {
		t.Fatalf("expected 0 for limit 0, got %d", n)
	}

	// Test identical slices across lengths 1 to 64
	buf1 := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz012345"), 4)
	buf2 := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz012345"), 4)
	for limit := uint(1); limit <= uint(len(buf1)); limit++ {
		got := findMatchLengthWithLimit(buf1, buf2, limit)
		if got != limit {
			t.Fatalf("for limit %d: expected %d, got %d", limit, limit, got)
		}
	}

	// Test mismatch at every position from 0 to limit-1
	for limit := uint(1); limit <= 32; limit++ {
		for mismatchPos := uint(0); mismatchPos < limit; mismatchPos++ {
			b1 := make([]byte, limit)
			b2 := make([]byte, limit)
			for i := range b1 {
				b1[i] = byte('A' + (i % 26))
				b2[i] = b1[i]
			}
			b2[mismatchPos] = b1[mismatchPos] ^ 0xFF // invert byte to guarantee mismatch
			got := findMatchLengthWithLimit(b1, b2, limit)
			if got != mismatchPos {
				t.Fatalf("limit=%d, mismatch at %d: got %d, want %d", limit, mismatchPos, got, mismatchPos)
			}
		}
	}
}
