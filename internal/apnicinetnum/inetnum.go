package apnicinetnum

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
	Lo                uint32
	Hi                uint32
	Netnames          []string
	Descriptions      []string
	Organizations     []string
	OrganizationNames []string
	Maintainers       []string
	Country           string
	Status            string
	LastModified      string
}

func AttachOrganizationNames(records []Record, names map[string]string) {
	for i := range records {
		for _, handle := range records[i].Organizations {
			if name := names[handle]; name != "" {
				records[i].OrganizationNames = appendUnique(records[i].OrganizationNames, name)
			}
		}
	}
}

type Match struct {
	Category  string
	Reason    string
	MatchedBy string
}

type Segment struct {
	Lo     uint32
	Hi     uint32
	Record Record
	Match  Match
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

	var records []Record
	fields := map[string][]string{}
	lastKey := ""
	lineNumber := 0
	finish := func() error {
		values := fields["inetnum"]
		if len(values) == 0 {
			fields = map[string][]string{}
			lastKey = ""
			return nil
		}
		parts := strings.Split(values[0], "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid inetnum %q", values[0])
		}
		lo, loErr := netip.ParseAddr(strings.TrimSpace(parts[0]))
		hi, hiErr := netip.ParseAddr(strings.TrimSpace(parts[1]))
		if loErr != nil || hiErr != nil || !lo.Is4() || !hi.Is4() || compare(lo, hi) > 0 {
			return fmt.Errorf("invalid IPv4 inetnum %q", values[0])
		}
		records = append(records, Record{
			Lo:            number(lo),
			Hi:            number(hi),
			Netnames:      clean(fields["netname"]),
			Descriptions:  clean(fields["descr"]),
			Organizations: clean(fields["org"]),
			Maintainers:   clean(fields["mnt-by"]),
			Country:       first(fields["country"]),
			Status:        first(fields["status"]),
			LastModified:  first(fields["last-modified"]),
		})
		fields = map[string][]string{}
		lastKey = ""
		return nil
	}

	scanner := bufio.NewScanner(z)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimRight(scanner.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			if err := finish(); err != nil {
				return nil, fmt.Errorf("%s near line %d: %w", path, lineNumber, err)
			}
			continue
		}
		if strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		if (line[0] == ' ' || line[0] == '\t' || line[0] == '+') && lastKey != "" {
			values := fields[lastKey]
			values[len(values)-1] = strings.TrimSpace(values[len(values)-1] + " " + strings.TrimSpace(strings.TrimPrefix(line, "+")))
			fields[lastKey] = values
			continue
		}
		colon := strings.IndexByte(line, ':')
		if colon <= 0 {
			return nil, fmt.Errorf("%s:%d: malformed RPSL line", path, lineNumber)
		}
		lastKey = strings.ToLower(strings.TrimSpace(line[:colon]))
		fields[lastKey] = append(fields[lastKey], strings.TrimSpace(line[colon+1:]))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := finish(); err != nil {
		return nil, fmt.Errorf("%s at EOF: %w", path, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%s contains no inetnum records", path)
	}
	return mergeExact(records), nil
}

func ResolveAll(records []Record, classify func(Record) Match) []Segment {
	type event struct {
		position uint64
		index    int
		add      bool
	}
	events := make([]event, 0, len(records)*2)
	matches := make([]Match, len(records))
	for i, record := range records {
		matches[i] = classify(record)
		events = append(events, event{position: uint64(record.Lo), index: i, add: true})
		events = append(events, event{position: uint64(record.Hi) + 1, index: i})
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].position != events[j].position {
			return events[i].position < events[j].position
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
		if previous < position && h.Len() > 0 {
			index := h.items[0]
			match := matches[index]
			appendSegment(&out, Segment{Lo: uint32(previous), Hi: uint32(position - 1), Record: records[index], Match: match})
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

func Matched(segments []Segment) []Segment {
	out := make([]Segment, 0, len(segments))
	for _, segment := range segments {
		if segment.Match.Reason != "" {
			out = append(out, segment)
		}
	}
	return out
}

func SearchText(record Record) string {
	parts := make([]string, 0, len(record.Netnames)+len(record.Descriptions)+len(record.OrganizationNames))
	parts = append(parts, record.Netnames...)
	parts = append(parts, record.Descriptions...)
	parts = append(parts, record.OrganizationNames...)
	return strings.Join(parts, " | ")
}

// RegistrantText excludes netname because it is an identifier, not evidence of
// a complete legal entity name.
func RegistrantText(record Record) string {
	parts := make([]string, 0, len(record.Descriptions)+len(record.OrganizationNames))
	parts = append(parts, record.Descriptions...)
	parts = append(parts, record.OrganizationNames...)
	return strings.Join(parts, " | ")
}

type recordHeap struct {
	records []Record
	items   []int
}

func (h recordHeap) Len() int { return len(h.items) }
func (h recordHeap) Less(i, j int) bool {
	a, b := h.records[h.items[i]], h.records[h.items[j]]
	aSize := uint64(a.Hi) - uint64(a.Lo)
	bSize := uint64(b.Hi) - uint64(b.Lo)
	if aSize != bSize {
		return aSize < bSize
	}
	if a.Lo != b.Lo {
		return a.Lo < b.Lo
	}
	if a.Hi != b.Hi {
		return a.Hi < b.Hi
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

func mergeExact(records []Record) []Record {
	sort.Slice(records, func(i, j int) bool {
		if records[i].Lo != records[j].Lo {
			return records[i].Lo < records[j].Lo
		}
		return records[i].Hi < records[j].Hi
	})
	out := make([]Record, 0, len(records))
	for _, record := range records {
		if len(out) == 0 || out[len(out)-1].Lo != record.Lo || out[len(out)-1].Hi != record.Hi {
			out = append(out, record)
			continue
		}
		last := &out[len(out)-1]
		last.Netnames = appendUnique(last.Netnames, record.Netnames...)
		last.Descriptions = appendUnique(last.Descriptions, record.Descriptions...)
		last.Organizations = appendUnique(last.Organizations, record.Organizations...)
		last.OrganizationNames = appendUnique(last.OrganizationNames, record.OrganizationNames...)
		last.Maintainers = appendUnique(last.Maintainers, record.Maintainers...)
		if last.Country == "" {
			last.Country = record.Country
		}
		if last.Status == "" {
			last.Status = record.Status
		}
		if record.LastModified > last.LastModified {
			last.LastModified = record.LastModified
		}
	}
	return out
}

func appendSegment(out *[]Segment, segment Segment) {
	if len(*out) != 0 {
		last := &(*out)[len(*out)-1]
		if last.Hi != ^uint32(0) && last.Hi+1 == segment.Lo && last.Record.Lo == segment.Record.Lo && last.Record.Hi == segment.Record.Hi && last.Match == segment.Match {
			last.Hi = segment.Hi
			return
		}
	}
	*out = append(*out, segment)
}

func appendUnique(values []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		if value != "" && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}

func clean(values []string) []string {
	return appendUnique(nil, values...)
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func number(a netip.Addr) uint32 {
	b := a.As4()
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func compare(a, b netip.Addr) int {
	an, bn := number(a), number(b)
	if an < bn {
		return -1
	}
	if an > bn {
		return 1
	}
	return 0
}
