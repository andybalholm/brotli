package matchfinder

import (
	"encoding/binary"
	"math"
	"math/bits"
	"slices"
)

// Pathfinder is a MatchFinder that uses hash chains to find matches, and a
// shortest-path optimizer to choose which matches to use.
type Pathfinder struct {
	// MaxDistance is the maximum distance (in bytes) to look back for
	// a match. The default is 65535.
	MaxDistance int

	// MinLength is the length of the shortest match to return.
	// The default is 4.
	MinLength int

	// HashLen is the number of bytes to use to calculate the hashes.
	// The maximum is 8 and the default is 6.
	HashLen int

	// TableBits is the number of bits in the hash table indexes.
	// The default is 17 (128K entries).
	TableBits int

	// ChainLength is how many entries to search on the "match chain" of older
	// locations with the same hash as the current location.
	ChainLength int

	table []uint32
	chain []uint16

	history  []byte
	arrivals []arrival
	matches  []Match
}

func (q *Pathfinder) Reset() {
	for i := range q.table {
		q.table[i] = 0
	}
	q.history = q.history[:0]
	q.chain = q.chain[:0]
}

// An arrival represents how we got to a certain byte position. If length==0,
// it is a literal, otherwise it is a match. The cost is the total cost (in
// bits) to get there from the beginning of the block. On literals, distance
// is set to the previous match distance.
type arrival struct {
	length   uint32
	distance uint32
	cost     float32
}

const (
	baseMatchCost   float32 = 5
	repeatMatchCost float32 = 6
)

func (q *Pathfinder) FindMatches(dst []Match, src []byte) []Match {
	if q.MaxDistance == 0 {
		q.MaxDistance = 65535
	}
	if q.MinLength == 0 {
		q.MinLength = 4
	}
	if q.HashLen == 0 {
		q.HashLen = 6
	}
	if q.TableBits == 0 {
		q.TableBits = 17
	}
	if len(q.table) < 1<<q.TableBits {
		q.table = make([]uint32, 1<<q.TableBits)
	}

	var histogram [256]uint32
	for _, b := range src {
		histogram[b]++
	}
	var byteCost [256]float32
	for b, n := range histogram {
		cost := max(math.Log2(float64(len(src))/float64(n)), 1)
		byteCost[b] = float32(cost)
	}

	// Each element in arrivals corresponds to the position just after
	// the corresponding byte in src.
	arrivals := q.arrivals
	if len(arrivals) < len(src) {
		arrivals = make([]arrival, len(src))
		q.arrivals = arrivals
	} else {
		arrivals = arrivals[:len(src)]
		for i := range arrivals {
			arrivals[i] = arrival{}
		}
	}

	if len(q.history) > q.MaxDistance*2 {
		// Trim down the history buffer.
		delta := len(q.history) - q.MaxDistance
		copy(q.history, q.history[delta:])
		q.history = q.history[:q.MaxDistance]
		if q.ChainLength > 0 {
			q.chain = q.chain[:q.MaxDistance]
		}

		for i, v := range q.table {
			newV := max(int(v)-delta, 0)
			q.table[i] = uint32(newV)
		}
	}

	// Append src to the history buffer.
	historyLen := len(q.history)
	q.history = append(q.history, src...)
	if q.ChainLength > 0 {
		q.chain = append(q.chain, make([]uint16, len(src))...)
	}
	src = q.history

	for i := historyLen; i < len(src); i++ {
		var arrivedHere arrival
		if i > historyLen {
			arrivedHere = arrivals[i-historyLen-1]
		}

		literalCost := byteCost[src[i]]
		nextArrival := &arrivals[i-historyLen]
		if nextArrival.cost == 0 || arrivedHere.cost+literalCost < nextArrival.cost {
			*nextArrival = arrival{
				cost:     arrivedHere.cost + literalCost,
				distance: arrivedHere.distance,
			}
		}

		if i > len(src)-8 {
			continue
		}

		if arrivedHere.length == 0 && arrivedHere.distance != 0 {
			// Look for a repeated match.
			prevDistance := int(arrivedHere.distance)
			if binary.LittleEndian.Uint32(src[i:]) == binary.LittleEndian.Uint32(src[i-prevDistance:]) {
				length := extendMatch(src, i-prevDistance+4, i+4) - i
				for j := q.MinLength; j <= length; j++ {
					a := &arrivals[i-historyLen-1+j]
					if a.cost == 0 || arrivedHere.cost+repeatMatchCost < a.cost {
						*a = arrival{
							length:   uint32(j),
							distance: arrivedHere.distance,
							cost:     arrivedHere.cost + repeatMatchCost,
						}
					}
				}
			}
		}

		// Calculate and store the hash.
		h := ((binary.LittleEndian.Uint64(src[i:]) & (1<<(8*q.HashLen) - 1)) * hashMul64) >> (64 - q.TableBits)
		candidate := int(q.table[h])
		q.table[h] = uint32(i)
		if q.ChainLength > 0 && candidate != 0 {
			delta := i - candidate
			if delta < 1<<16 {
				q.chain[i] = uint16(delta)
			}
		}
		if candidate == 0 || i-candidate > q.MaxDistance {
			continue
		}

		prevLength := 0
		if binary.LittleEndian.Uint32(src[candidate:]) == binary.LittleEndian.Uint32(src[i:]) {
			length := extendMatch(src, candidate+4, i+4) - i
			matchCost := baseMatchCost + float32(bits.Len(uint(i-candidate)))
			for j := q.MinLength; j <= length; j++ {
				a := &arrivals[i-historyLen-1+j]
				if a.cost == 0 || arrivedHere.cost+matchCost < a.cost {
					*a = arrival{
						length:   uint32(j),
						distance: uint32(i - candidate),
						cost:     arrivedHere.cost + matchCost,
					}
				}
			}
			prevLength = length
		}

		for range q.ChainLength {
			delta := q.chain[candidate]
			if delta == 0 {
				break
			}
			candidate -= int(delta)
			if candidate <= 0 || i-candidate > q.MaxDistance {
				break
			}
			if binary.LittleEndian.Uint32(src[candidate:]) == binary.LittleEndian.Uint32(src[i:]) {
				length := extendMatch(src, candidate+4, i+4) - i
				if length > prevLength {
					matchCost := baseMatchCost + float32(bits.Len(uint(i-candidate)))
					for j := q.MinLength; j <= length; j++ {
						a := &arrivals[i-historyLen-1+j]
						if a.cost == 0 || arrivedHere.cost+matchCost < a.cost {
							*a = arrival{
								length:   uint32(j),
								distance: uint32(i - candidate),
								cost:     arrivedHere.cost + matchCost,
							}
						}
					}
					prevLength = length
				}
			}
		}
	}

	// We've found the shortest path; now walk it backward and store the matches.
	matches := q.matches[:0]
	i := len(arrivals) - 1
	for i >= 0 {
		a := arrivals[i]
		if a.length > 0 {
			matches = append(matches, Match{
				Length:   int(a.length),
				Distance: int(a.distance),
			})
			i -= int(a.length)
		} else {
			if len(matches) == 0 {
				matches = append(matches, Match{})
			}
			matches[len(matches)-1].Unmatched++
			i--
		}
	}
	q.matches = matches

	slices.Reverse(matches)

	return append(dst, matches...)
}
