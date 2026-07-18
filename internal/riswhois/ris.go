package riswhois

import (
	"bufio"
	"compress/gzip"
	"container/heap"
	"fmt"
	"net/netip"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Origin struct {
	ASN       string `json:"asn"`
	SeenPeers int    `json:"seen_peers"`
}
type Record struct {
	Lo, Hi  uint32
	Prefix  string
	Origins []Origin
}
type Segment struct {
	Lo, Hi uint32
	Record Record
}
type Stats struct{ Rows, Prefixes, RelevantPrefixes int }

func Parse(path string, relevant func(uint32, uint32) bool) ([]Record, Stats, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, Stats{}, e
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		return nil, Stats{}, e
	}
	defer z.Close()
	group := map[string]*Record{}
	stats := Stats{}
	s := bufio.NewScanner(z)
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		x := strings.Fields(line)
		if len(x) != 3 {
			continue
		}
		p, e := netip.ParsePrefix(x[1])
		if e != nil || !p.Addr().Is4() {
			continue
		}
		p = p.Masked()
		peers, e := strconv.Atoi(x[2])
		if e != nil {
			continue
		}
		stats.Rows++
		r := group[p.String()]
		if r == nil {
			lo, hi := number(p.Addr()), end(p)
			if relevant != nil && !relevant(lo, hi) {
				group[p.String()] = &Record{Prefix: "-"}
				continue
			}
			r = &Record{Lo: lo, Hi: hi, Prefix: p.String()}
			group[p.String()] = r
		}
		if r.Prefix == "-" {
			continue
		}
		asn := strings.TrimPrefix(strings.ToUpper(x[0]), "AS")
		found := false
		for i := range r.Origins {
			if r.Origins[i].ASN == asn {
				if peers > r.Origins[i].SeenPeers {
					r.Origins[i].SeenPeers = peers
				}
				found = true
				break
			}
		}
		if !found {
			r.Origins = append(r.Origins, Origin{asn, peers})
		}
	}
	if e := s.Err(); e != nil {
		return nil, stats, e
	}
	stats.Prefixes = len(group)
	out := []Record{}
	for _, r := range group {
		if r.Prefix != "-" {
			sort.Slice(r.Origins, func(i, j int) bool { return r.Origins[i].ASN < r.Origins[j].ASN })
			out = append(out, *r)
		}
	}
	stats.RelevantPrefixes = len(out)
	if stats.Rows == 0 {
		return nil, stats, fmt.Errorf("%s contains no RIS rows", path)
	}
	return out, stats, nil
}
func Resolve(records []Record) []Segment {
	if len(records) == 0 {
		return nil
	}
	type event struct {
		p uint64
		i int
		a bool
	}
	ev := make([]event, 0, len(records)*2)
	for i, r := range records {
		ev = append(ev, event{uint64(r.Lo), i, true}, event{uint64(r.Hi) + 1, i, false})
	}
	sort.Slice(ev, func(i, j int) bool {
		if ev[i].p != ev[j].p {
			return ev[i].p < ev[j].p
		}
		return !ev[i].a && ev[j].a
	})
	active := make([]bool, len(records))
	h := &rh{r: records}
	heap.Init(h)
	prev := ev[0].p
	out := []Segment{}
	for i := 0; i < len(ev); {
		p := ev[i].p
		for h.Len() > 0 && !active[h.x[0]] {
			heap.Pop(h)
		}
		if prev < p && h.Len() > 0 {
			out = append(out, Segment{uint32(prev), uint32(p - 1), records[h.x[0]]})
		}
		for i < len(ev) && ev[i].p == p {
			q := ev[i]
			active[q.i] = q.a
			if q.a {
				heap.Push(h, q.i)
			}
			i++
		}
		prev = p
	}
	return out
}

type rh struct {
	r []Record
	x []int
}

func (h rh) Len() int { return len(h.x) }
func (h rh) Less(i, j int) bool {
	a, b := h.r[h.x[i]], h.r[h.x[j]]
	as, bs := uint64(a.Hi)-uint64(a.Lo), uint64(b.Hi)-uint64(b.Lo)
	if as != bs {
		return as < bs
	}
	return h.x[i] < h.x[j]
}
func (h rh) Swap(i, j int) { h.x[i], h.x[j] = h.x[j], h.x[i] }
func (h *rh) Push(v any)   { h.x = append(h.x, v.(int)) }
func (h *rh) Pop() any     { x := h.x[len(h.x)-1]; h.x = h.x[:len(h.x)-1]; return x }
func number(a netip.Addr) uint32 {
	b := a.As4()
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
func end(p netip.Prefix) uint32 {
	return uint32(uint64(number(p.Addr())) + (uint64(1) << uint(32-p.Bits())) - 1)
}
