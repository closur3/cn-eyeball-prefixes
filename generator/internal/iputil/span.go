package iputil

import (
	"math/bits"
	"net/netip"
	"sort"
)

type Span struct{ Lo, Hi uint32 }

func MergeSpans(in []Span) []Span {
	sort.Slice(in, func(i, j int) bool { return in[i].Lo < in[j].Lo })
	out := make([]Span, 0, len(in))
	for _, x := range in {
		if len(out) == 0 || (out[len(out)-1].Hi != ^uint32(0) && x.Lo > out[len(out)-1].Hi+1) {
			out = append(out, x)
			continue
		}
		if x.Hi > out[len(out)-1].Hi {
			out[len(out)-1].Hi = x.Hi
		}
	}
	return out
}

func SubtractSpans(in, excluded []Span) []Span {
	in, excluded = MergeSpans(in), MergeSpans(excluded)
	var out []Span
	j := 0
	for _, r := range in {
		for j < len(excluded) && excluded[j].Hi < r.Lo {
			j++
		}
		pos := r.Lo
		covered := false
		for k := j; k < len(excluded) && excluded[k].Lo <= r.Hi; k++ {
			x := excluded[k]
			if x.Hi < pos {
				continue
			}
			if x.Lo > pos {
				out = append(out, Span{pos, x.Lo - 1})
			}
			if x.Hi >= r.Hi {
				covered = true
				break
			}
			pos = x.Hi + 1
		}
		if !covered {
			out = append(out, Span{pos, r.Hi})
		}
	}
	return out
}

func IntersectSpans(a, b []Span) []Span {
	a, b = MergeSpans(a), MergeSpans(b)
	var out []Span
	for i, j := 0, 0; i < len(a) && j < len(b); {
		lo, hi := a[i].Lo, a[i].Hi
		if b[j].Lo > lo {
			lo = b[j].Lo
		}
		if b[j].Hi < hi {
			hi = b[j].Hi
		}
		if lo <= hi {
			out = append(out, Span{lo, hi})
		}
		if a[i].Hi < b[j].Hi {
			i++
		} else {
			j++
		}
	}
	return out
}

// OverlapsSorted reports whether a normalized, address-sorted span set
// intersects [lo, hi] without repeatedly sorting it for point queries.
func OverlapsSorted(rows []Span, lo, hi uint32) bool {
	i := sort.Search(len(rows), func(i int) bool { return rows[i].Hi >= lo })
	return i < len(rows) && rows[i].Lo <= hi
}

func AddressCount(rows []Span) uint64 {
	var count uint64
	for _, row := range MergeSpans(rows) {
		count += uint64(row.Hi) - uint64(row.Lo) + 1
	}
	return count
}

func SpanCIDRs(rows []Span) []string {
	rows = MergeSpans(rows)
	var lines []string
	for _, row := range rows {
		r := row
		for r.Lo <= r.Hi {
			remaining := uint64(r.Hi) - uint64(r.Lo) + 1
			align := bits.TrailingZeros32(r.Lo)
			if r.Lo == 0 {
				align = 32
			}
			sizeBits := align
			if max := bits.Len64(remaining) - 1; max < sizeBits {
				sizeBits = max
			}
			size := uint64(1) << uint(sizeBits)
			a := netip.AddrFrom4([4]byte{byte(r.Lo >> 24), byte(r.Lo >> 16), byte(r.Lo >> 8), byte(r.Lo)})
			lines = append(lines, netip.PrefixFrom(a, 32-sizeBits).String())
			if size == remaining {
				break
			}
			r.Lo += uint32(size)
		}
	}
	return lines
}
