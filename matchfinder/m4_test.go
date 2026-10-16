package matchfinder

import (
	"bytes"
	"testing"
)

func TestM4FindMatchesFidelity(t *testing.T) {
	mf := &M4{
		MaxDistance: 1 << 16,
		ChainLength: 16,
	}

	data := bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 50)
	matches := mf.FindMatches(nil, data)
	if len(matches) == 0 {
		t.Fatal("expected matches to be found in repetitive input")
	}

	// Verify every match points to valid repeating data
	cursor := 0
	for idx, m := range matches {
		cursor += m.Unmatched
		if m.Length > 0 {
			if m.Distance <= 0 {
				t.Fatalf("match %d: invalid distance %d", idx, m.Distance)
			}
			if cursor < m.Distance {
				t.Fatalf("match %d: distance %d exceeds cursor position %d", idx, m.Distance, cursor)
			}
			matchedSlice := data[cursor : cursor+m.Length]
			sourceSlice := data[cursor-m.Distance : cursor-m.Distance+m.Length]
			if !bytes.Equal(matchedSlice, sourceSlice) {
				t.Fatalf("match %d: mismatched content between %q and %q", idx, matchedSlice, sourceSlice)
			}
			cursor += m.Length
		}
	}
}

func TestM4MultipleBlocks(t *testing.T) {
	mf := &M4{
		MaxDistance: 1 << 16,
		ChainLength: 8,
	}

	block1 := []byte("hello world 12345 hello world 12345 ")
	block2 := []byte("hello world 12345 repeating text ")

	matches1 := mf.FindMatches(nil, block1)
	matches2 := mf.FindMatches(nil, block2)

	if len(matches1) == 0 || len(matches2) == 0 {
		t.Fatalf("expected matches in both blocks, got %d and %d", len(matches1), len(matches2))
	}
}

func TestPathfinderFindMatchesFidelity(t *testing.T) {
	pf := &Pathfinder{
		MaxDistance: 1 << 16,
		ChainLength: 8,
	}

	data := bytes.Repeat([]byte("repeating pattern with lots of redundancies in text stream "), 30)
	matches := pf.FindMatches(nil, data)
	if len(matches) == 0 {
		t.Fatal("expected matches to be found in repetitive input")
	}

	cursor := 0
	for idx, m := range matches {
		cursor += m.Unmatched
		if m.Length > 0 {
			if m.Distance <= 0 {
				t.Fatalf("match %d: invalid distance %d", idx, m.Distance)
			}
			if cursor < m.Distance {
				t.Fatalf("match %d: distance %d exceeds cursor %d", idx, m.Distance, cursor)
			}
			matchedSlice := data[cursor : cursor+m.Length]
			sourceSlice := data[cursor-m.Distance : cursor-m.Distance+m.Length]
			if !bytes.Equal(matchedSlice, sourceSlice) {
				t.Fatalf("match %d: mismatched content", idx)
			}
			cursor += m.Length
		}
	}
}

func TestExtendMatchEdgeCases(t *testing.T) {
	s1 := []byte("0123456789abcdef0123456789abcdef")
	// s1 has repeating 16-byte blocks: offset 0 and offset 16
	end := extendMatch(s1, 0, 16)
	if end != len(s1) {
		t.Fatalf("expected extendMatch to match entire slice, got %d, want %d", end, len(s1))
	}

	// Partial match
	s2 := []byte("abcdefghijklmnopabcdefghxxxxxx")
	end2 := extendMatch(s2, 0, 16)
	if end2 != 24 {
		t.Fatalf("expected extendMatch to match 8 bytes, got end %d, want 24", end2)
	}
}
