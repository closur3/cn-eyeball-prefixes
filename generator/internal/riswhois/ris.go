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

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/iputil"
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

func (r Record) GetLo() uint32 { return r.Lo }
func (r Record) GetHi() uint32 { return r.Hi }

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
			lo, hi := iputil.Number(p.Addr()), iputil.End(p)
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
	h := &iputil.SpanHeap[Record]{Data: records}
	heap.Init(h)
	prev := ev[0].p
	out := []Segment{}
	for i := 0; i < len(ev); {
		p := ev[i].p
		for h.Len() > 0 && !active[h.Top()] {
			heap.Pop(h)
		}
		if prev < p && h.Len() > 0 {
			out = append(out, Segment{uint32(prev), uint32(p - 1), records[h.Top()]})
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
