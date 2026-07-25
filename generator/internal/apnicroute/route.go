package apnicroute

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

type Variant struct {
	Origin                                                      string
	Descriptions, Organizations, OrganizationNames, Maintainers []string
	LastModified                                                string
}
type Record struct {
	Lo, Hi   uint32
	Prefix   string
	Variants []Variant
}

func (r Record) GetLo() uint32 { return r.Lo }
func (r Record) GetHi() uint32 { return r.Hi }
type Segment struct {
	Lo, Hi uint32
	Record Record
}

func Parse(path string, orgNames map[string]string, relevant func(uint32, uint32) bool) ([]Record, int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return nil, 0, 0, err
	}
	defer z.Close()
	fields := map[string][]string{}
	last := ""
	objects := 0
	relevantObjects := 0
	var raw []Record
	finish := func() error {
		defer func() { fields, last = map[string][]string{}, "" }()
		if len(fields["route"]) == 0 {
			return nil
		}
		p, e := netip.ParsePrefix(fields["route"][0])
		if e != nil || !p.Addr().Is4() {
			return fmt.Errorf("invalid route %q", fields["route"][0])
		}
		p = p.Masked()
		origin := strings.TrimPrefix(strings.ToUpper(iputil.First(fields["origin"])), "AS")
		if _, e = strconv.ParseUint(origin, 10, 32); e != nil {
			return fmt.Errorf("invalid route origin %q", iputil.First(fields["origin"]))
		}
		objects++
		lo, hi := iputil.Number(p.Addr()), iputil.End(p)
		if relevant != nil && !relevant(lo, hi) {
			return nil
		}
		relevantObjects++
		orgs := iputil.Clean(fields["org"])
		var names []string
		for _, h := range orgs {
			if n := orgNames[h]; n != "" {
				names = iputil.AppendUnique(names, n)
			}
		}
		maintainers := append(iputil.Clean(fields["mnt-by"]), iputil.Clean(fields["mnt-routes"])...)
		raw = append(raw, Record{lo, hi, p.String(), []Variant{{origin, iputil.Clean(fields["descr"]), orgs, names, iputil.Clean(maintainers), iputil.First(fields["last-modified"])}}})
		return nil
	}
	s := bufio.NewScanner(z)
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	for s.Scan() {
		line := strings.TrimRight(s.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			if e := finish(); e != nil {
				return nil, objects, relevantObjects, e
			}
			continue
		}
		if strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		if (line[0] == ' ' || line[0] == '\t' || line[0] == '+') && last != "" {
			v := fields[last]
			v[len(v)-1] = strings.TrimSpace(v[len(v)-1] + " " + strings.TrimSpace(strings.TrimPrefix(line, "+")))
			fields[last] = v
			continue
		}
		c := strings.IndexByte(line, ':')
		if c <= 0 {
			return nil, objects, relevantObjects, fmt.Errorf("%s: malformed RPSL", path)
		}
		last = strings.ToLower(strings.TrimSpace(line[:c]))
		fields[last] = append(fields[last], strings.TrimSpace(line[c+1:]))
	}
	if e := s.Err(); e != nil {
		return nil, objects, relevantObjects, e
	}
	if e := finish(); e != nil {
		return nil, objects, relevantObjects, e
	}
	if objects == 0 {
		return nil, 0, 0, fmt.Errorf("%s contains no route records", path)
	}
	sort.Slice(raw, func(i, j int) bool {
		if raw[i].Lo != raw[j].Lo {
			return raw[i].Lo < raw[j].Lo
		}
		return raw[i].Hi < raw[j].Hi
	})
	out := []Record{}
	for _, r := range raw {
		if len(out) == 0 || out[len(out)-1].Lo != r.Lo || out[len(out)-1].Hi != r.Hi {
			out = append(out, r)
			continue
		}
		out[len(out)-1].Variants = append(out[len(out)-1].Variants, r.Variants...)
	}
	return out, objects, relevantObjects, nil
}

func Resolve(records []Record) []Segment {
	if len(records) == 0 {
		return nil
	}
	type event struct {
		pos uint64
		idx int
		add bool
	}
	ev := make([]event, 0, len(records)*2)
	for i, r := range records {
		ev = append(ev, event{uint64(r.Lo), i, true}, event{uint64(r.Hi) + 1, i, false})
	}
	sort.Slice(ev, func(i, j int) bool {
		if ev[i].pos != ev[j].pos {
			return ev[i].pos < ev[j].pos
		}
		return !ev[i].add && ev[j].add
	})
	active := make([]bool, len(records))
	h := &iputil.SpanHeap[Record]{Data: records}
	heap.Init(h)
	out := []Segment{}
	prev := ev[0].pos
	for i := 0; i < len(ev); {
		pos := ev[i].pos
		for h.Len() > 0 && !active[h.Top()] {
			heap.Pop(h)
		}
		if prev < pos && h.Len() > 0 {
			r := records[h.Top()]
			out = append(out, Segment{uint32(prev), uint32(pos - 1), r})
		}
		for i < len(ev) && ev[i].pos == pos {
			e := ev[i]
			active[e.idx] = e.add
			if e.add {
				heap.Push(h, e.idx)
			}
			i++
		}
		prev = pos
	}
	return out
}
func SearchText(v Variant) string {
	x := append([]string{}, v.Descriptions...)
	x = append(x, v.OrganizationNames...)
	return strings.Join(x, " | ")
}


