package apnic6

import (
	"bufio"
	"compress/gzip"
	"container/heap"
	"fmt"
	"net/netip"
	"os"
	"sort"
	"strings"
)

type Record struct {
	Prefix       netip.Prefix
	Lo           netip.Addr
	Hi           netip.Addr
	Descriptions []string
}

type Segment struct {
	Lo     netip.Addr
	Hi     netip.Addr
	Record Record
}

func Parse(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	byPrefix := map[string]*Record{}
	fields := map[string][]string{}
	lastField := ""
	flush := func() error {
		defer func() { fields, lastField = map[string][]string{}, "" }()
		if len(fields["inet6num"]) == 0 {
			return nil
		}
		prefix, err := netip.ParsePrefix(fields["inet6num"][0])
		if err != nil || !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
			return fmt.Errorf("invalid inet6num %q", fields["inet6num"][0])
		}
		prefix = prefix.Masked()
		lo, hi := prefix.Addr(), lastAddress(prefix)
		record := byPrefix[prefix.String()]
		if record == nil {
			record = &Record{Prefix: prefix, Lo: lo, Hi: hi}
			byPrefix[prefix.String()] = record
		}
		for _, description := range fields["descr"] {
			description = strings.TrimSpace(description)
			if description != "" && !contains(record.Descriptions, description) {
				record.Descriptions = append(record.Descriptions, description)
			}
		}
		return nil
	}
	scanner := bufio.NewScanner(z)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			if err := flush(); err != nil {
				return nil, err
			}
			continue
		}
		if strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		if (line[0] == ' ' || line[0] == '\t' || line[0] == '+') && lastField != "" {
			values := fields[lastField]
			values[len(values)-1] = strings.TrimSpace(values[len(values)-1] + " " + strings.TrimSpace(strings.TrimPrefix(line, "+")))
			fields[lastField] = values
			continue
		}
		colon := strings.IndexByte(line, ':')
		if colon <= 0 {
			return nil, fmt.Errorf("malformed APNIC RPSL line")
		}
		lastField = strings.ToLower(strings.TrimSpace(line[:colon]))
		fields[lastField] = append(fields[lastField], strings.TrimSpace(line[colon+1:]))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := flush(); err != nil {
		return nil, err
	}
	out := make([]Record, 0, len(byPrefix))
	for _, record := range byPrefix {
		out = append(out, *record)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no IPv6 inet6num records")
	}
	return out, nil
}

func ResolveMostSpecific(records []Record) []Segment {
	type event struct {
		position netip.Addr
		index    int
		add      bool
	}
	var events []event
	for i, record := range records {
		events = append(events, event{position: record.Lo, index: i, add: true})
		if next := record.Hi.Next(); next.IsValid() {
			events = append(events, event{position: next, index: i})
		}
	}
	sort.Slice(events, func(i, j int) bool {
		if c := events[i].position.Compare(events[j].position); c != 0 {
			return c < 0
		}
		return !events[i].add && events[j].add
	})
	active := make([]bool, len(records))
	h := &recordHeap{records: records}
	heap.Init(h)
	previous := events[0].position
	var out []Segment
	for i := 0; i < len(events); {
		position := events[i].position
		for h.Len() > 0 && !active[h.items[0]] {
			heap.Pop(h)
		}
		if previous.Compare(position) < 0 && h.Len() > 0 {
			index := h.items[0]
			out = append(out, Segment{Lo: previous, Hi: position.Prev(), Record: records[index]})
		}
		for i < len(events) && events[i].position == position {
			e := events[i]
			active[e.index] = e.add
			if e.add {
				heap.Push(h, e.index)
			}
			i++
		}
		previous = position
	}
	return out
}

func lastAddress(prefix netip.Prefix) netip.Addr {
	b := prefix.Masked().Addr().As16()
	for bit := prefix.Bits(); bit < 128; bit++ {
		b[bit/8] |= 1 << uint(7-bit%8)
	}
	return netip.AddrFrom16(b)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

type recordHeap struct {
	records []Record
	items   []int
}

func (h recordHeap) Len() int { return len(h.items) }
func (h recordHeap) Less(i, j int) bool {
	a, b := h.records[h.items[i]], h.records[h.items[j]]
	if a.Prefix.Bits() != b.Prefix.Bits() {
		return a.Prefix.Bits() > b.Prefix.Bits()
	}
	return h.items[i] < h.items[j]
}
func (h recordHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *recordHeap) Push(value any) { h.items = append(h.items, value.(int)) }
func (h *recordHeap) Pop() any {
	last := len(h.items) - 1
	value := h.items[last]
	h.items = h.items[:last]
	return value
}
