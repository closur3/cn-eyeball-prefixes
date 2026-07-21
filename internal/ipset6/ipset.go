package ipset6

import (
	"fmt"
	"math/big"
	"net/netip"
	"sort"
)

type Range struct {
	Lo netip.Addr
	Hi netip.Addr
}

func FromPrefix(prefix netip.Prefix) (Range, error) {
	if !prefix.IsValid() || !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
		return Range{}, fmt.Errorf("not an IPv6 prefix: %s", prefix)
	}
	prefix = prefix.Masked()
	bytes := prefix.Addr().As16()
	for bit := prefix.Bits(); bit < 128; bit++ {
		bytes[bit/8] |= byte(1 << uint(7-bit%8))
	}
	return Range{Lo: prefix.Addr(), Hi: netip.AddrFrom16(bytes)}, nil
}

func Merge(rows []Range) []Range {
	rows = append([]Range(nil), rows...)
	sort.Slice(rows, func(i, j int) bool {
		if c := rows[i].Lo.Compare(rows[j].Lo); c != 0 {
			return c < 0
		}
		return rows[i].Hi.Compare(rows[j].Hi) < 0
	})
	out := make([]Range, 0, len(rows))
	for _, row := range rows {
		if !valid(row) {
			panic("invalid IPv6 range")
		}
		if len(out) == 0 {
			out = append(out, row)
			continue
		}
		last := &out[len(out)-1]
		next := last.Hi.Next()
		if row.Lo.Compare(last.Hi) <= 0 || (next.IsValid() && row.Lo == next) {
			if row.Hi.Compare(last.Hi) > 0 {
				last.Hi = row.Hi
			}
			continue
		}
		out = append(out, row)
	}
	return out
}

func Intersect(a, b []Range) []Range {
	a, b = Merge(a), Merge(b)
	out := make([]Range, 0)
	for i, j := 0, 0; i < len(a) && j < len(b); {
		lo := maxAddr(a[i].Lo, b[j].Lo)
		hi := minAddr(a[i].Hi, b[j].Hi)
		if lo.Compare(hi) <= 0 {
			out = append(out, Range{Lo: lo, Hi: hi})
		}
		if a[i].Hi.Compare(b[j].Hi) < 0 {
			i++
		} else {
			j++
		}
	}
	return Merge(out)
}

func Subtract(a, b []Range) []Range {
	a, b = Merge(a), Merge(b)
	out := make([]Range, 0)
	j := 0
	for _, row := range a {
		cursor := row.Lo
		for j < len(b) && b[j].Hi.Compare(cursor) < 0 {
			j++
		}
		for k := j; k < len(b) && b[k].Lo.Compare(row.Hi) <= 0; k++ {
			if b[k].Lo.Compare(cursor) > 0 {
				out = append(out, Range{Lo: cursor, Hi: b[k].Lo.Prev()})
			}
			if b[k].Hi.Compare(row.Hi) >= 0 {
				cursor = netip.Addr{}
				break
			}
			cursor = b[k].Hi.Next()
		}
		if cursor.IsValid() && cursor.Compare(row.Hi) <= 0 {
			out = append(out, Range{Lo: cursor, Hi: row.Hi})
		}
	}
	return Merge(out)
}

func Prefixes(rows []Range) []netip.Prefix {
	rows = Merge(rows)
	var out []netip.Prefix
	for _, row := range rows {
		cursor := row.Lo
		for cursor.Compare(row.Hi) <= 0 {
			chosen := netip.PrefixFrom(cursor, 128)
			chosenRange, _ := FromPrefix(chosen)
			for bits := 0; bits <= 128; bits++ {
				candidate := netip.PrefixFrom(cursor, bits).Masked()
				if candidate.Addr() != cursor {
					continue
				}
				candidateRange, _ := FromPrefix(candidate)
				if candidateRange.Hi.Compare(row.Hi) <= 0 {
					chosen, chosenRange = candidate, candidateRange
					break
				}
			}
			out = append(out, chosen)
			if chosenRange.Hi == row.Hi {
				break
			}
			cursor = chosenRange.Hi.Next()
		}
	}
	return out
}

func AddressCount(rows []Range) *big.Int {
	total := new(big.Int)
	for _, row := range Merge(rows) {
		size := new(big.Int).Sub(addrInt(row.Hi), addrInt(row.Lo))
		size.Add(size, big.NewInt(1))
		total.Add(total, size)
	}
	return total
}

func Slash64Equivalent(rows []Range) string {
	count := AddressCount(rows)
	value := new(big.Rat).SetFrac(count, new(big.Int).Lsh(big.NewInt(1), 64))
	return value.FloatString(4)
}

func valid(row Range) bool {
	return row.Lo.IsValid() && row.Hi.IsValid() && row.Lo.Is6() && row.Hi.Is6() && !row.Lo.Is4In6() && !row.Hi.Is4In6() && row.Lo.Compare(row.Hi) <= 0
}

func addrInt(addr netip.Addr) *big.Int {
	bytes := addr.As16()
	return new(big.Int).SetBytes(bytes[:])
}

func minAddr(a, b netip.Addr) netip.Addr {
	if a.Compare(b) <= 0 {
		return a
	}
	return b
}

func maxAddr(a, b netip.Addr) netip.Addr {
	if a.Compare(b) >= 0 {
		return a
	}
	return b
}
