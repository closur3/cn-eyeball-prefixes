package apnicautnum

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

type Record struct {
	ASN, ASName                string
	Organizations, Maintainers []string
}
type Link struct{ ASN, ASName, Via string }
type Index struct {
	byName, byOrg, byMaintainer map[string][]Record
	byASN                       map[string]Record
	active                      map[string]string
}

func Parse(path string) ([]Record, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	fields := map[string][]string{}
	last := ""
	var out []Record
	finish := func() {
		if len(fields["aut-num"]) > 0 {
			asn := strings.TrimPrefix(strings.ToUpper(fields["aut-num"][0]), "AS")
			name := first(fields["as-name"])
			if asn != "" && name != "" {
				maintainers := append(clean(fields["mnt-by"]), clean(fields["mnt-routes"])...)
				out = append(out, Record{asn, name, clean(fields["org"]), clean(maintainers)})
			}
		}
		fields = map[string][]string{}
		last = ""
	}
	s := bufio.NewScanner(z)
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	for s.Scan() {
		line := strings.TrimRight(s.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			finish()
			continue
		}
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "%") {
			continue
		}
		if (line[0] == ' ' || line[0] == '\t' || line[0] == '+') && last != "" {
			continue
		}
		c := strings.IndexByte(line, ':')
		if c <= 0 {
			return nil, fmt.Errorf("%s: malformed RPSL", path)
		}
		last = strings.ToLower(strings.TrimSpace(line[:c]))
		fields[last] = append(fields[last], strings.TrimSpace(line[c+1:]))
	}
	if e := s.Err(); e != nil {
		return nil, e
	}
	finish()
	if len(out) == 0 {
		return nil, fmt.Errorf("%s contains no aut-num records", path)
	}
	return out, nil
}

func NewIndex(records []Record, active map[string]string) *Index {
	return newIndex(records, active, true)
}

// NewRegistryIndex indexes every APNIC aut-num object, including ASNs that are
// not present in the current IPtoASN snapshot. It is used only when another
// independent registry signal is also present.
func NewRegistryIndex(records []Record) *Index {
	return newIndex(records, nil, false)
}

func newIndex(records []Record, active map[string]string, activeOnly bool) *Index {
	x := &Index{map[string][]Record{}, map[string][]Record{}, map[string][]Record{}, map[string]Record{}, active}
	for _, r := range records {
		if activeOnly && active[r.ASN] == "" {
			continue
		}
		x.byASN[r.ASN] = r
		name := Normalize(r.ASName)
		if len(name) >= 5 && name != "UNSPECIFIED" {
			x.byName[name] = append(x.byName[name], r)
		}
		for _, o := range r.Organizations {
			x.byOrg[strings.ToUpper(o)] = append(x.byOrg[strings.ToUpper(o)], r)
		}
		for _, m := range r.Maintainers {
			x.byMaintainer[strings.ToUpper(m)] = append(x.byMaintainer[strings.ToUpper(m)], r)
		}
	}
	return x
}

func (x *Index) Record(asn string) (Record, bool) {
	r, ok := x.byASN[asn]
	return r, ok
}

func (x *Index) DedicatedMaintainer(asn, maintainer string) bool {
	records := x.byMaintainer[strings.ToUpper(maintainer)]
	return len(records) == 1 && records[0].ASN == asn
}

func CommonAll(groups ...[]string) []string {
	if len(groups) == 0 {
		return nil
	}
	counts := map[string]int{}
	original := map[string]string{}
	for i, group := range groups {
		seen := map[string]bool{}
		for _, value := range group {
			key := strings.ToUpper(value)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			if i == 0 {
				original[key] = value
			}
			counts[key]++
		}
	}
	var out []string
	for key, value := range original {
		if counts[key] == len(groups) {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
func (x *Index) Links(netnames, orgs []string) []Link {
	seen := map[string]bool{}
	var out []Link
	add := func(r Record, via string) {
		k := r.ASN + "\x00" + via
		if !seen[k] {
			seen[k] = true
			out = append(out, Link{r.ASN, r.ASName, via})
		}
	}
	for _, n := range netnames {
		key := Normalize(n)
		if len(key) < 5 {
			continue
		}
		for _, r := range x.byName[key] {
			add(r, "netname_as_name")
		}
	}
	for _, o := range orgs {
		for _, r := range x.byOrg[strings.ToUpper(o)] {
			add(r, "organisation_handle")
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ASN != out[j].ASN {
			return out[i].ASN < out[j].ASN
		}
		return out[i].Via < out[j].Via
	})
	return out
}
func Normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
func first(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}
func clean(v []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, x := range v {
		if x != "" && !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}
