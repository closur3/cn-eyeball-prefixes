package main

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/closur3/cn-eyeball-prefixes/internal/apnic6"
	"github.com/closur3/cn-eyeball-prefixes/internal/riswhois6"
)

const (
	telecomBlock      = "240e::/18"
	fixedDescription  = "Chinatelecom IPv6 address for fixed broadband"
	mobileDescription = "Chinatelecom IPv6 address for mobile"
)

type asMeta struct {
	Country     string
	Description string
}

type sourceMeta struct {
	Source string `json:"source"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type outputMeta struct {
	Path                 string         `json:"path"`
	PrefixCount          int            `json:"prefix_count"`
	BGPPrefixesByPurpose map[string]int `json:"bgp_prefixes_by_purpose"`
}

type registryAdmissionMeta struct {
	Descriptions          []string       `json:"descriptions"`
	MatchedInet6numRecords map[string]int `json:"matched_inet6num_records"`
	EffectiveRanges        map[string]int `json:"effective_ranges"`
}

type manifest struct {
	GeneratedAt       string                `json:"generated_at"`
	Sources           map[string]sourceMeta `json:"sources"`
	Boundary          string                `json:"telecom_boundary"`
	OutputUnit        string                `json:"output_unit"`
	RegistryAdmission registryAdmissionMeta `json:"registry_admission"`
	Outputs           map[string]outputMeta `json:"outputs"`
	Rejected          map[string]int        `json:"rejected_bgp_prefixes"`
}

type admissionRange struct {
	Lo      netip.Addr
	Hi      netip.Addr
	Purpose string
}

func main() {
	risPath := flag.String("ris", "", "RIPE RISWhois IPv6 gzip")
	iptoasnPath := flag.String("iptoasn", "", "IPtoASN IPv6 TSV gzip")
	inet6numPath := flag.String("inet6num", "", "APNIC inet6num gzip")
	outputPath := flag.String("output", "data/ipv6/operators/chinatelecom.txt", "China Telecom access prefix output")
	manifestPath := flag.String("manifest", "data/ipv6/manifest.json", "manifest path")
	flag.Parse()
	if *risPath == "" || *iptoasnPath == "" || *inet6numPath == "" {
		panic("--ris, --iptoasn, and --inet6num are required")
	}

	boundary := netip.MustParsePrefix(telecomBlock)
	metadata, err := readASNMetadata(*iptoasnPath)
	must(err)
	bgpRecords, err := riswhois6.Parse(*risPath)
	must(err)
	registrations, err := apnic6.Parse(*inet6numPath, boundary)
	must(err)
	resolvedRegistrations := apnic6.ResolveMostSpecific(registrations)
	admissionRanges := buildAdmissionRanges(resolvedRegistrations)
	matchedInet6numRecords := map[string]int{"fixed_broadband": 0, "mobile": 0}
	for _, registration := range registrations {
		switch registrationPurpose(registration) {
		case "fixed":
			matchedInet6numRecords["fixed_broadband"]++
		case "mobile":
			matchedInet6numRecords["mobile"]++
		}
	}
	effectiveRanges := map[string]int{"fixed_broadband": 0, "mobile": 0}
	for _, admission := range admissionRanges {
		if admission.Purpose == "fixed" {
			effectiveRanges["fixed_broadband"]++
		} else {
			effectiveRanges["mobile"]++
		}
	}

	var admitted []netip.Prefix
	admissionMatches := map[string]int{"fixed_broadband": 0, "mobile": 0}
	rejected := map[string]int{}
	for _, record := range bgpRecords {
		if !inside(record.Prefix, boundary) {
			continue
		}
		if !allTelecomOrigins(record.Origins, metadata) {
			rejected["non_telecom_origin"]++
			continue
		}
		purpose, reason := classifyPrefix(record.Prefix, admissionRanges)
		switch purpose {
		case "fixed":
			admitted = append(admitted, record.Prefix)
			admissionMatches["fixed_broadband"]++
		case "mobile":
			admitted = append(admitted, record.Prefix)
			admissionMatches["mobile"]++
		default:
			rejected[reason]++
		}
	}

	sort.Slice(admitted, func(i, j int) bool {
		if c := admitted[i].Addr().Compare(admitted[j].Addr()); c != 0 {
			return c < 0
		}
		return admitted[i].Bits() < admitted[j].Bits()
	})
	must(writePrefixes(*outputPath, admitted))
	sources := map[string]sourceMeta{}
	for name, item := range map[string]struct{ path, source string }{
		"riswhois_ipv6": {*risPath, "https://www.ris.ripe.net/dumps/riswhoisdump.IPv6.gz"},
		"iptoasn_ipv6":  {*iptoasnPath, "https://iptoasn.com/data/ip2asn-v6.tsv.gz"},
		"apnic_inet6num": {*inet6numPath, "https://ftp.apnic.net/apnic/whois/apnic.db.inet6num.gz"},
	} {
		meta, err := fileMetadata(item.path)
		must(err)
		meta.Source = item.source
		sources[name] = meta
	}
	result := manifest{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Sources: sources,
		Boundary: telecomBlock,
		OutputUnit: "exact_current_bgp_prefix",
		RegistryAdmission: registryAdmissionMeta{
			Descriptions:           []string{fixedDescription, mobileDescription},
			MatchedInet6numRecords: matchedInet6numRecords,
			EffectiveRanges:        effectiveRanges,
		},
		Outputs: map[string]outputMeta{
			"chinatelecom": {
				Path:                 filepath.ToSlash(*outputPath),
				PrefixCount:          len(admitted),
				BGPPrefixesByPurpose: admissionMatches,
			},
		},
		Rejected: rejected,
	}
	must(writeJSON(*manifestPath, result))
}

func buildAdmissionRanges(segments []apnic6.Segment) []admissionRange {
	var out []admissionRange
	for _, segment := range segments {
		purpose := registrationPurpose(segment.Record)
		if purpose == "" {
			continue
		}
		if len(out) > 0 && out[len(out)-1].Purpose == purpose && out[len(out)-1].Hi.Next() == segment.Lo {
			out[len(out)-1].Hi = segment.Hi
			continue
		}
		out = append(out, admissionRange{Lo: segment.Lo, Hi: segment.Hi, Purpose: purpose})
	}
	return out
}

func classifyPrefix(prefix netip.Prefix, ranges []admissionRange) (string, string) {
	lo, hi := prefix.Masked().Addr(), lastAddress(prefix)
	i := sort.Search(len(ranges), func(i int) bool { return ranges[i].Hi.Compare(lo) >= 0 })
	cursor := lo
	purpose := ""
	for ; i < len(ranges) && ranges[i].Lo.Compare(hi) <= 0; i++ {
		admission := ranges[i]
		if admission.Lo.Compare(cursor) > 0 {
			return "", "outside_admitted_registry"
		}
		if purpose != "" && purpose != admission.Purpose {
			return "", "mixed_access_purpose"
		}
		purpose = admission.Purpose
		if admission.Hi.Compare(hi) >= 0 {
			return purpose, ""
		}
		cursor = admission.Hi.Next()
	}
	return "", "outside_admitted_registry"
}

func registrationPurpose(record apnic6.Record) string {
	for _, description := range record.Descriptions {
		switch {
		case strings.EqualFold(strings.TrimSpace(description), fixedDescription):
			return "fixed"
		case strings.EqualFold(strings.TrimSpace(description), mobileDescription):
			return "mobile"
		}
	}
	return ""
}

func allTelecomOrigins(origins []string, metadata map[string]asMeta) bool {
	if len(origins) == 0 {
		return false
	}
	for _, asn := range origins {
		meta, ok := metadata[asn]
		if !ok || !strings.EqualFold(meta.Country, "CN") || !isTelecom(asn, meta.Description) {
			return false
		}
	}
	return true
}

func isTelecom(asn, description string) bool {
	if asn == "4847" {
		return true
	}
	value := strings.ToLower(description)
	return strings.Contains(value, "china telecom") || strings.Contains(value, "chinatelecom") || strings.Contains(value, "chinanet")
}

func readASNMetadata(path string) (map[string]asMeta, error) {
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
	type choice struct {
		meta  asMeta
		count int
	}
	choices := map[string]map[string]*choice{}
	scanner := bufio.NewScanner(z)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 5 || fields[2] == "0" {
			continue
		}
		key := fields[3] + "\x00" + fields[4]
		if choices[fields[2]] == nil {
			choices[fields[2]] = map[string]*choice{}
		}
		entry := choices[fields[2]][key]
		if entry == nil {
			entry = &choice{meta: asMeta{Country: fields[3], Description: fields[4]}}
			choices[fields[2]][key] = entry
		}
		entry.count++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	out := map[string]asMeta{}
	for asn, variants := range choices {
		best := &choice{}
		for _, candidate := range variants {
			if candidate.count > best.count || (candidate.count == best.count && candidate.meta.Description < best.meta.Description) {
				best = candidate
			}
		}
		out[asn] = best.meta
	}
	return out, nil
}

func inside(prefix, boundary netip.Prefix) bool {
	return boundary.Contains(prefix.Masked().Addr()) && boundary.Contains(lastAddress(prefix))
}

func lastAddress(prefix netip.Prefix) netip.Addr {
	b := prefix.Masked().Addr().As16()
	for bit := prefix.Bits(); bit < 128; bit++ {
		b[bit/8] |= 1 << uint(7-bit%8)
	}
	return netip.AddrFrom16(b)
}

func writePrefixes(path string, prefixes []netip.Prefix) error {
	if len(prefixes) == 0 {
		return fmt.Errorf("refusing to write empty prefix list: %s", path)
	}
	var b strings.Builder
	for _, prefix := range prefixes {
		fmt.Fprintln(&b, prefix)
	}
	return writeFile(path, []byte(b.String()))
}

func fileMetadata(path string) (sourceMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return sourceMeta{}, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return sourceMeta{}, err
	}
	return sourceMeta{Bytes: n, SHA256: hex.EncodeToString(h.Sum(nil))}, nil
}

func writeJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, append(b, '\n'))
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
