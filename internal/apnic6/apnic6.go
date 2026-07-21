package apnic6

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

	"github.com/closur3/cn-eyeball-prefixes/internal/ipset6"
)

type InetRecord struct {
	Prefix            netip.Prefix
	Range             ipset6.Range
	Netnames          []string
	Descriptions      []string
	Organizations     []string
	OrganizationNames []string
	Maintainers       []string
	Country           string
	Status            string
	LastModified      string
}

type Segment struct {
	Range  ipset6.Range
	Record InetRecord
}

type RouteVariant struct {
	Origin            string
	Descriptions      []string
	Organizations     []string
	OrganizationNames []string
	Maintainers       []string
	LastModified      string
}

type RouteRecord struct {
	Prefix   netip.Prefix
	Range    ipset6.Range
	Variants []RouteVariant
}

func ParseInet6num(path string, orgNames map[string]string) ([]InetRecord, error) {
	byPrefix := map[string]*InetRecord{}
	err := parseRPSL(path, func(fields map[string][]string) error {
		if len(fields["inet6num"]) == 0 {
			return nil
		}
		prefix, err := netip.ParsePrefix(first(fields["inet6num"]))
		if err != nil || !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
			return fmt.Errorf("invalid inet6num %q", first(fields["inet6num"]))
		}
		prefix = prefix.Masked()
		row, err := ipset6.FromPrefix(prefix)
		if err != nil {
			return err
		}
		orgs := clean(fields["org"])
		var names []string
		for _, handle := range orgs {
			if name := orgNames[handle]; name != "" {
				names = appendUnique(names, name)
			}
		}
		record := byPrefix[prefix.String()]
		if record == nil {
			record = &InetRecord{Prefix: prefix, Range: row}
			byPrefix[prefix.String()] = record
		}
		record.Netnames = appendUnique(record.Netnames, clean(fields["netname"])...)
		record.Descriptions = appendUnique(record.Descriptions, clean(fields["descr"])...)
		record.Organizations = appendUnique(record.Organizations, orgs...)
		record.OrganizationNames = appendUnique(record.OrganizationNames, names...)
		record.Maintainers = appendUnique(record.Maintainers, clean(fields["mnt-by"])...)
		if record.Country == "" {
			record.Country = first(fields["country"])
		}
		if record.Status == "" {
			record.Status = first(fields["status"])
		}
		if record.LastModified == "" {
			record.LastModified = first(fields["last-modified"])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]InetRecord, 0, len(byPrefix))
	for _, record := range byPrefix {
		out = append(out, *record)
	}
	sort.Slice(out, func(i, j int) bool {
		if c := out[i].Range.Lo.Compare(out[j].Range.Lo); c != 0 {
			return c < 0
		}
		return out[i].Prefix.Bits() < out[j].Prefix.Bits()
	})
	if len(out) == 0 {
		return nil, fmt.Errorf("%s contains no inet6num records", path)
	}
	return out, nil
}

func ResolveMostSpecific(records []InetRecord) []Segment {
	type event struct {
		position netip.Addr
		index    int
		add      bool
	}
	events := make([]event, 0, len(records)*2)
	for i, record := range records {
		events = append(events, event{position: record.Range.Lo, index: i, add: true})
		if next := record.Range.Hi.Next(); next.IsValid() {
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
	var out []Segment
	previous := events[0].position
	for i := 0; i < len(events); {
		position := events[i].position
		for h.Len() > 0 && !active[h.items[0]] {
			heap.Pop(h)
		}
		if previous.Compare(position) < 0 && h.Len() > 0 {
			index := h.items[0]
			appendSegment(&out, Segment{Range: ipset6.Range{Lo: previous, Hi: position.Prev()}, Record: records[index]})
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

func ParseRoute6(path string, orgNames map[string]string) ([]RouteRecord, error) {
	byPrefix := map[string]*RouteRecord{}
	err := parseRPSL(path, func(fields map[string][]string) error {
		if len(fields["route6"]) == 0 {
			return nil
		}
		prefix, err := netip.ParsePrefix(first(fields["route6"]))
		if err != nil || !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
			return fmt.Errorf("invalid route6 %q", first(fields["route6"]))
		}
		prefix = prefix.Masked()
		origin := strings.TrimPrefix(strings.ToUpper(first(fields["origin"])), "AS")
		if _, err := strconv.ParseUint(origin, 10, 32); err != nil {
			return fmt.Errorf("invalid route6 origin %q", first(fields["origin"]))
		}
		row, err := ipset6.FromPrefix(prefix)
		if err != nil {
			return err
		}
		orgs := clean(fields["org"])
		var names []string
		for _, handle := range orgs {
			if name := orgNames[handle]; name != "" {
				names = appendUnique(names, name)
			}
		}
		maintainers := append(clean(fields["mnt-by"]), clean(fields["mnt-routes"])...)
		variant := RouteVariant{Origin: origin, Descriptions: clean(fields["descr"]), Organizations: orgs, OrganizationNames: names, Maintainers: clean(maintainers), LastModified: first(fields["last-modified"])}
		record := byPrefix[prefix.String()]
		if record == nil {
			record = &RouteRecord{Prefix: prefix, Range: row}
			byPrefix[prefix.String()] = record
		}
		record.Variants = append(record.Variants, variant)
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]RouteRecord, 0, len(byPrefix))
	for _, record := range byPrefix {
		out = append(out, *record)
	}
	sort.Slice(out, func(i, j int) bool {
		if c := out[i].Range.Lo.Compare(out[j].Range.Lo); c != 0 {
			return c < 0
		}
		return out[i].Prefix.Bits() < out[j].Prefix.Bits()
	})
	if len(out) == 0 {
		return nil, fmt.Errorf("%s contains no route6 records", path)
	}
	return out, nil
}

func InetSearchText(record InetRecord) string {
	parts := append([]string{}, record.Netnames...)
	parts = append(parts, record.Descriptions...)
	parts = append(parts, record.OrganizationNames...)
	return strings.Join(parts, " | ")
}

func InetRegistrantText(record InetRecord) string {
	parts := append([]string{}, record.Descriptions...)
	parts = append(parts, record.OrganizationNames...)
	return strings.Join(parts, " | ")
}

func RouteSearchText(variant RouteVariant) string {
	parts := append([]string{}, variant.Descriptions...)
	parts = append(parts, variant.OrganizationNames...)
	return strings.Join(parts, " | ")
}

type recordHeap struct {
	records []InetRecord
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
func (h recordHeap) Swap(i, j int)   { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *recordHeap) Push(value any) { h.items = append(h.items, value.(int)) }
func (h *recordHeap) Pop() any {
	last := len(h.items) - 1
	value := h.items[last]
	h.items = h.items[:last]
	return value
}

func appendSegment(out *[]Segment, segment Segment) {
	if len(*out) > 0 {
		last := &(*out)[len(*out)-1]
		if last.Record.Prefix == segment.Record.Prefix && last.Range.Hi.Next() == segment.Range.Lo {
			last.Range.Hi = segment.Range.Hi
			return
		}
	}
	*out = append(*out, segment)
}

func parseRPSL(path string, finish func(map[string][]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer z.Close()
	fields := map[string][]string{}
	last := ""
	flush := func() error {
		if len(fields) == 0 {
			return nil
		}
		if err := finish(fields); err != nil {
			return err
		}
		fields, last = map[string][]string{}, ""
		return nil
	}
	scanner := bufio.NewScanner(z)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		if (line[0] == ' ' || line[0] == '\t' || line[0] == '+') && last != "" {
			values := fields[last]
			values[len(values)-1] = strings.TrimSpace(values[len(values)-1] + " " + strings.TrimSpace(strings.TrimPrefix(line, "+")))
			fields[last] = values
			continue
		}
		colon := strings.IndexByte(line, ':')
		if colon <= 0 {
			return fmt.Errorf("%s: malformed RPSL line", path)
		}
		last = strings.ToLower(strings.TrimSpace(line[:colon]))
		fields[last] = append(fields[last], strings.TrimSpace(line[colon+1:]))
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return flush()
}

func clean(values []string) []string { return appendUnique(nil, values...) }
func appendUnique(values []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}
func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
