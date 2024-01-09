package matchfinder

import (
	"encoding/binary"
	"math/bits"
	"sort"
)

// MultiHash is an implementation of the MatchFinder
// interface that uses multiple hashes of different lengths.
type MultiHash struct {
	// MaxDistance is the maximum distance (in bytes) to look back for
	// a match. The default is 65535.
	MaxDistance int

	// MinLength is the length of the shortest match to return.
	// The default is 4.
	MinLength int

	// HashLengths is a list of the hashes to use, with the number of
	// bytes to use for each. For example, to to use 4-byte, 7-byte, and
	// 10-byte hashes, set HashLengths to []int{4, 7, 10}.
	// The minimum length is 4.
	HashLengths []int

	// TableBits is the number of bits in the hash table indexes.
	// The default is 17 (128K entries).
	TableBits int

	// DistanceBitCost is used when comparing two matches to see
	// which is better. The comparison is primarily based on the length
	// of the matches, but it can also take the distance into account,
	// in terms of the number of bits needed to represent the distance.
	// One byte of length is given a score of 256, so 32 (256/8) would
	// be a reasonable first guess for the value of one bit.
	// (The default is 0, which bases the comparison solely on length.)
	DistanceBitCost int

	tables [][]uint32

	history []byte
}

func (q *MultiHash) Reset() {
	for _, t := range q.tables {
		for i := range t {
			t[i] = 0
		}
	}
	q.history = q.history[:0]
}

func (q *MultiHash) score(m absoluteMatch) int {
	return (m.End-m.Start)*256 + bits.LeadingZeros32(uint32(m.Start-m.Match))*q.DistanceBitCost
}

func (q *MultiHash) FindMatches(dst []Match, src []byte) []Match {
	if q.MaxDistance == 0 {
		q.MaxDistance = 65535
	}
	if q.MinLength == 0 {
		q.MinLength = 4
	}
	if q.TableBits == 0 {
		q.TableBits = 17
	}
	if len(q.tables) < len(q.HashLengths) {
		q.tables = make([][]uint32, len(q.HashLengths))
		for i := range q.tables {
			q.tables[i] = make([]uint32, 1<<q.TableBits)
		}
	}
	sort.Ints(q.HashLengths)
	maxHashLen := q.HashLengths[len(q.HashLengths)-1]

	e := matchEmitter{Dst: dst}

	if len(q.history) > q.MaxDistance*2 {
		// Trim down the history buffer.
		delta := len(q.history) - q.MaxDistance
		copy(q.history, q.history[delta:])
		q.history = q.history[:q.MaxDistance]

		for _, t := range q.tables {
			for i, v := range t {
				newV := int(v) - delta
				if newV < 0 {
					newV = 0
				}
				t[i] = uint32(newV)
			}
		}
	}

	// Append src to the history buffer.
	e.NextEmit = len(q.history)
	q.history = append(q.history, src...)
	src = q.history

	// matches stores the matches that have been found but not emitted,
	// in reverse order. (matches[0] is the most recent one.)
	var matches [3]absoluteMatch

	candidates := make([]int, len(q.HashLengths))

	for i := e.NextEmit; i < len(src)-maxHashLen; i++ {
		if matches[0] != (absoluteMatch{}) && i >= matches[0].End {
			// We have found some matches, and we're far enough along that we probably
			// won't find overlapping matches, so we might as well emit them.
			if matches[1] != (absoluteMatch{}) {
				e.trim(matches[1], matches[0].Start, q.MinLength)
			}
			e.emit(matches[0])
			matches = [3]absoluteMatch{}
		}

		// Calculate and store the hashes.
		h := uint32(0x811c9dc5) // FNV-32 offset basis
		nb := 0
		for j, hashLen := range q.HashLengths {
			for nb < hashLen {
				h ^= uint32(src[i+nb])
				h *= 0x01000193 // FNV-32 prime
				nb++
			}
			index := h >> (32 - q.TableBits)
			candidates[j] = int(q.tables[j][index])
			q.tables[j][index] = uint32(i)
		}

		// Look for a match.
		var currentMatch absoluteMatch

		if i < matches[0].End {
			// If we're looking for an overlapping match, we only need to check the
			// hash that ends 2 bytes after the end of the previous match.
			for j, candidate := range candidates {
				if i+q.HashLengths[j] != matches[0].End+2 {
					continue
				}
				if candidate == 0 || i-candidate > q.MaxDistance {
					break
				}
				if binary.LittleEndian.Uint32(src[candidate:]) != binary.LittleEndian.Uint32(src[i:]) {
					break
				}
				m := extendMatch2(src, i, candidate, e.NextEmit)
				if m.End-m.Start >= q.HashLengths[j] {
					currentMatch = m
				}
			}
		} else {
			for j, candidate := range candidates {
				if candidate == 0 || i-candidate > q.MaxDistance {
					break
				}
				if i-candidate == matches[0].Start-matches[0].Match {
					// Don't bother to check for the same match we already have.
					continue
				}
				if currentMatch.End-currentMatch.Start > q.HashLengths[j] {
					// Don't bother with hashes that are shorter than the current match.
					continue
				}
				if binary.LittleEndian.Uint32(src[candidate:]) != binary.LittleEndian.Uint32(src[i:]) {
					break
				}
				m := extendMatch2(src, i, candidate, e.NextEmit)
				if m.End-m.Start > q.MinLength && q.score(m) > q.score(currentMatch) {
					currentMatch = m
				}
			}
		}

		if currentMatch == (absoluteMatch{}) || q.score(currentMatch) <= q.score(matches[0]) {
			continue
		}

		matches = [3]absoluteMatch{
			currentMatch,
			matches[0],
			matches[1],
		}

		if matches[2] == (absoluteMatch{}) {
			continue
		}

		// We have three matches, so it's time to emit one and/or eliminate one.
		switch {
		case matches[0].Start < matches[2].End:
			// The first and third matches overlap; discard the one in between.
			matches = [3]absoluteMatch{
				matches[0],
				matches[2],
				absoluteMatch{},
			}

		case matches[0].Start < matches[2].End+q.MinLength:
			// The first and third matches don't overlap, but there's no room for
			// another match between them. Emit the first match and discard the second.
			e.emit(matches[2])
			matches = [3]absoluteMatch{
				matches[0],
				absoluteMatch{},
				absoluteMatch{},
			}

		default:
			// Emit the first match, shortening it if necessary to avoid overlap with the second.
			e.trim(matches[2], matches[1].Start, q.MinLength)
			matches[2] = absoluteMatch{}
		}
	}

	// We've found all the matches now; emit the remaining ones.
	if matches[1] != (absoluteMatch{}) {
		e.trim(matches[1], matches[0].Start, q.MinLength)
	}
	if matches[0] != (absoluteMatch{}) {
		e.emit(matches[0])
	}

	dst = e.Dst
	if e.NextEmit < len(src) {
		dst = append(dst, Match{
			Unmatched: len(src) - e.NextEmit,
		})
	}

	return dst
}
