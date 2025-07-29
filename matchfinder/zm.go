package matchfinder

import "encoding/binary"

const (
	zmTableBits = 15
	zmTableSize = 1 << zmTableBits
)

// ZM is a MatchFinder that combines the cache table of ZFast with the
// overlap-based parsing of M4.
type ZM struct {
	MaxDistance int
	history     []byte
	table       [zmTableSize]tableEntry
}

func (z *ZM) Reset() {
	z.table = [zmTableSize]tableEntry{}
	z.history = z.history[:0]
}

func (z *ZM) FindMatches(dst []Match, src []byte) []Match {
	if z.MaxDistance == 0 {
		z.MaxDistance = 1 << 16
	}

	if len(z.history) > z.MaxDistance*2 {
		delta := len(z.history) - z.MaxDistance
		copy(z.history, z.history[delta:])
		z.history = z.history[:z.MaxDistance]

		for i := range z.table {
			v := z.table[i].offset
			v -= int32(delta)
			if v < 0 {
				z.table[i] = tableEntry{}
			} else {
				z.table[i].offset = v
			}
		}
	}

	e := matchEmitter{
		Dst:      dst,
		NextEmit: len(z.history),
	}
	z.history = append(z.history, src...)
	src = z.history

	if len(src) < 10 {
		return append(dst, Match{
			Unmatched: len(src),
		})
	}

	// matches stores the matches that have been found but not emitted,
	// in reverse order. (matches[0] is the most recent one.)
	var matches [3]absoluteMatch

	sLimit := int32(len(src)) - 8

mainLoop:
	for {
		// Search for a match, starting after the last match emitted.
		s := int32(e.NextEmit)
		if s > sLimit {
			break mainLoop
		}
		cv := binary.LittleEndian.Uint64(src[s:])

		for {
			nextHash := z.hash(cv)
			nextHash2 := z.hash(cv >> 8)
			candidate := z.table[nextHash]
			candidate2 := z.table[nextHash2]
			z.table[nextHash] = tableEntry{offset: s, val: uint32(cv)}
			z.table[nextHash2] = tableEntry{offset: s + 1, val: uint32(cv >> 8)}

			// Look for a repeat match two bytes after the current position.
			if len(e.Dst) > 0 {
				prevDistance := int32(e.Dst[len(e.Dst)-1].Distance)
				if prevDistance != 0 {
					repIndex := s - prevDistance + 2
					if repIndex >= 0 && binary.LittleEndian.Uint32(src[repIndex:]) == uint32(cv>>16) {
						// There is a repeated match at s+2.
						matches[0] = extendMatch2(src, int(s+2), int(repIndex), e.NextEmit+1)
						break
					}
				}
			}

			if candidate.offset < s && s-candidate.offset < int32(z.MaxDistance) && uint32(cv) == candidate.val {
				// There is a match at s.
				matches[0] = extendMatch2(src, int(s), int(candidate.offset), e.NextEmit)
				break
			}
			if candidate2.offset < s+1 && s+1-candidate2.offset < int32(z.MaxDistance) && uint32(cv>>8) == candidate2.val {
				// There is a match at s+1.
				matches[0] = extendMatch2(src, int(s+1), int(candidate2.offset), e.NextEmit)
				break
			}

			s += 2 + ((s - int32(e.NextEmit)) >> 5)
			if s > sLimit {
				break mainLoop
			}
			cv = binary.LittleEndian.Uint64(src[s:])
		}

		// We have a match in matches[0].
		// Now look for overlapping matches.

		for {
			if matches[0].End > int(sLimit) {
				break
			}
			s = int32(matches[0].End - 4)
			cv = binary.LittleEndian.Uint64(src[s:])
			nextHash := z.hash(cv)
			nextHash2 := z.hash(cv >> 8)
			candidate := z.table[nextHash]
			candidate2 := z.table[nextHash2]
			z.table[nextHash] = tableEntry{offset: s, val: uint32(cv)}
			z.table[nextHash2] = tableEntry{offset: s + 1, val: uint32(cv >> 8)}

			var newMatch absoluteMatch
			if candidate.offset < s && s-candidate.offset < int32(z.MaxDistance) && uint32(cv) == candidate.val {
				// There is a match at s.
				newMatch = extendMatch2(src, int(s), int(candidate.offset), e.NextEmit)
			} else if candidate2.offset < s && s+1-candidate2.offset < int32(z.MaxDistance) && uint32(cv>>8) == candidate2.val {
				// There is a match at s+1.
				newMatch = extendMatch2(src, int(s+1), int(candidate2.offset), e.NextEmit)
				break
			}

			if newMatch.End-newMatch.Start <= matches[0].End-matches[0].Start {
				// The new match isn't longer than the old one, so we break out of the loop
				// of looking for overlapping matches.
				break
			}

			matches = [3]absoluteMatch{
				newMatch,
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

			case matches[0].Start < matches[2].End+4:
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
				if matches[2].End > matches[1].Start {
					matches[2].End = matches[1].Start
				}
				if matches[2].End-matches[2].Start >= 4 {
					e.emit(matches[2])
				}
				matches[2] = absoluteMatch{}
			}
		}

		// We're done looking for overlapping matches; emit the ones we have.

		if matches[1] != (absoluteMatch{}) {
			if matches[1].End > matches[0].Start {
				matches[1].End = matches[0].Start
			}
			if matches[1].End-matches[1].Start >= 4 {
				e.emit(matches[1])
			}
		}
		e.emit(matches[0])
		matches = [3]absoluteMatch{}
	}

	dst = e.Dst
	if e.NextEmit < len(src) {
		dst = append(dst, Match{
			Unmatched: len(src) - e.NextEmit,
		})
	}

	return dst
}

func (z *ZM) hash(u uint64) uint32 {
	return uint32(((u << 16) * 227718039650203) >> (64 - zmTableBits))
}
