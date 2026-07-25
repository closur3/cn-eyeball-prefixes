package ipv6build

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
	"strconv"
	"strings"
	"time"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/apnic6"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/iputil"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/operatorconfig"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/riswhois6"
)

const (
	// These exact APNIC inet6num descriptions are admission policy. Matching
	// registration prefixes are discovered across the complete current APNIC
	// database on every build; no allocation boundary is a static input.
	fixedBroadbandInet6numDescription = "Chinatelecom IPv6 address for fixed broadband"
	mobileInet6numDescription         = "Chinatelecom IPv6 address for mobile"
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
	OriginASNs           []originMeta   `json:"origin_asns"`
}

type originMeta struct {
	ASN            string   `json:"asn"`
	Description    string   `json:"description"`
	PrefixCount    int      `json:"prefix_count"`
	SamplePrefixes []string `json:"sample_prefixes"`
}

type registryAdmissionMeta struct {
	Descriptions            []string            `json:"descriptions"`
	MatchedInet6numRecords  map[string]int      `json:"matched_inet6num_records"`
	MatchedInet6numPrefixes map[string][]string `json:"matched_inet6num_prefixes"`
	EffectiveRanges         map[string]int      `json:"effective_ranges"`
}

type auditReport struct {
	GeneratedAt       string                         `json:"generated_at"`
	Sources           map[string]sourceMeta          `json:"sources"`
	OutputUnit        string                         `json:"output_unit"`
	RegistryAdmission registryAdmissionMeta          `json:"registry_admission"`
	CuratedAdmission  map[string]map[string][]string `json:"curated_admission"`
	Outputs           map[string]outputMeta          `json:"outputs"`
	Rejected          map[string]map[string]int      `json:"rejected_bgp_prefixes"`
}

type admissionRange struct {
	Lo      netip.Addr
	Hi      netip.Addr
	Purpose string
}

type operatorPolicy struct {
	Name   string
	Ranges []admissionRange
}

type originAccumulator struct {
	Prefixes []string
}

type Options struct {
	RISPath              string
	IPToASNPath          string
	Inet6numPath         string
	OutputDir            string
	AuditReportPath      string
	OperatorConfigPath   string
	AllocationConfigPath string
}

// Main is the IPv6 command entry point used by the repository's unified
// generator wrapper.
func Main() {
	must(RunCLI(os.Args[1:]))
}

// RunCLI parses IPv6 generator arguments without mutating flag.CommandLine.
func RunCLI(args []string) error {
	flags := flag.NewFlagSet("generate ipv6", flag.ContinueOnError)
	risPath := flags.String("ris", "", "RIPE RISWhois IPv6 gzip")
	iptoasnPath := flags.String("iptoasn", "", "IPtoASN IPv6 TSV gzip")
	inet6numPath := flags.String("inet6num", "", "APNIC inet6num gzip")
	outputDir := flags.String("output-dir", "", "staging IPv6 family output directory")
	auditReportPath := flags.String("audit-report", "", "optional detailed build audit JSON (CI artifact, not a public manifest)")
	operatorConfigPath := flags.String("operator-config", "config/operators.json", "operator ASN classification config")
	allocationConfigPath := flags.String("allocation-config", "config/ipv6-province-prefixes.json", "provincial IPv6 allocation config")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *risPath == "" || *iptoasnPath == "" || *inet6numPath == "" {
		return fmt.Errorf("--ris, --iptoasn, and --inet6num are required")
	}
	if *outputDir == "" {
		return fmt.Errorf("--output-dir is required")
	}
	return Run(Options{
		RISPath:              *risPath,
		IPToASNPath:          *iptoasnPath,
		Inet6numPath:         *inet6numPath,
		OutputDir:            *outputDir,
		AuditReportPath:      *auditReportPath,
		OperatorConfigPath:   *operatorConfigPath,
		AllocationConfigPath: *allocationConfigPath,
	})
}

// Run applies the current BGP/registry admission policy and writes the compact
// public family layout. Detailed source evidence is optional and belongs in a
// CI artifact rather than lists/manifest.json.
func Run(options Options) error {
	if options.RISPath == "" || options.IPToASNPath == "" || options.Inet6numPath == "" {
		return fmt.Errorf("RIS, IPtoASN, and APNIC inet6num paths are required")
	}
	if options.OutputDir == "" || options.OperatorConfigPath == "" || options.AllocationConfigPath == "" {
		return fmt.Errorf("output, operator config, and allocation config paths are required")
	}

	metadata, err := readASNMetadata(options.IPToASNPath)
	if err != nil {
		return err
	}
	classifier, err := operatorconfig.Load(options.OperatorConfigPath, operatorNames)
	if err != nil {
		return err
	}
	allocationConfig, err := LoadAllocationConfig(options.AllocationConfigPath)
	if err != nil {
		return err
	}
	bgpRecords, err := riswhois6.Parse(options.RISPath)
	if err != nil {
		return err
	}
	registrations, err := apnic6.Parse(options.Inet6numPath)
	if err != nil {
		return err
	}
	// Resolve the current registry hierarchy first. A more-specific APNIC
	// object overrides a broader one and can therefore create a denied hole in
	// an otherwise admitted registration.
	resolvedRegistrations := apnic6.ResolveMostSpecific(registrations)
	admissionRanges := buildAdmissionRanges(resolvedRegistrations)
	curatedAdmission := map[string]map[string][]string{
		"chinamobile": {
			"fixed_broadband": {"2409:8a00::/24"},
			"mobile":          {"2409:8900::/24"},
		},
		"chinaunicom": {
			"fixed_broadband": {"2408:8200::/24", "2408:8300::/24"},
			"mobile":          {"2408:8400::/24"},
		},
	}
	policies := []operatorPolicy{
		{Name: "chinatelecom", Ranges: admissionRanges},
		{Name: "chinamobile", Ranges: rangesFromPrefixes(curatedAdmission["chinamobile"])},
		{Name: "chinaunicom", Ranges: rangesFromPrefixes(curatedAdmission["chinaunicom"])},
	}
	matchedInet6numRecords := map[string]int{"fixed_broadband": 0, "mobile": 0}
	matchedInet6numPrefixes := map[string][]string{"fixed_broadband": {}, "mobile": {}}
	for _, registration := range registrations {
		switch registrationPurpose(registration) {
		case "fixed":
			matchedInet6numRecords["fixed_broadband"]++
			matchedInet6numPrefixes["fixed_broadband"] = append(matchedInet6numPrefixes["fixed_broadband"], registration.Prefix.String())
		case "mobile":
			matchedInet6numRecords["mobile"]++
			matchedInet6numPrefixes["mobile"] = append(matchedInet6numPrefixes["mobile"], registration.Prefix.String())
		}
	}
	sort.Strings(matchedInet6numPrefixes["fixed_broadband"])
	sort.Strings(matchedInet6numPrefixes["mobile"])
	effectiveRanges := map[string]int{"fixed_broadband": 0, "mobile": 0}
	for _, admission := range admissionRanges {
		if admission.Purpose == "fixed" {
			effectiveRanges["fixed_broadband"]++
		} else {
			effectiveRanges["mobile"]++
		}
	}

	admitted := map[string][]netip.Prefix{}
	admissionMatches := map[string]map[string]int{}
	rejected := map[string]map[string]int{}
	origins := map[string]map[string]*originAccumulator{}
	for _, policy := range policies {
		admissionMatches[policy.Name] = map[string]int{"fixed_broadband": 0, "mobile": 0}
		rejected[policy.Name] = map[string]int{}
		origins[policy.Name] = map[string]*originAccumulator{}
	}
	for _, record := range bgpRecords {
		for _, policy := range policies {
			if !overlapsAdmissionRanges(record.Prefix, policy.Ranges) {
				continue
			}
			// A BGP prefix is admitted only when its complete address range is
			// covered by one purpose range and every Origin matches the operator.
			purpose, reason := classifyPrefix(record.Prefix, policy.Ranges)
			if purpose == "" {
				rejected[policy.Name][reason]++
				continue
			}
			if !allOriginsMatch(record.Origins, metadata, classifier, policy.Name) {
				rejected[policy.Name]["non_operator_origin"]++
				continue
			}
			admitted[policy.Name] = append(admitted[policy.Name], record.Prefix)
			if purpose == "fixed" {
				admissionMatches[policy.Name]["fixed_broadband"]++
			} else {
				admissionMatches[policy.Name]["mobile"]++
			}
			for _, asn := range record.Origins {
				entry := origins[policy.Name][asn]
				if entry == nil {
					entry = &originAccumulator{}
					origins[policy.Name][asn] = entry
				}
				entry.Prefixes = append(entry.Prefixes, record.Prefix.String())
			}
		}
	}

	publicLists, err := BuildPublicLists(admitted, allocationConfig)
	if err != nil {
		return err
	}
	if err := WritePublicLists(options.OutputDir, publicLists); err != nil {
		return err
	}

	outputs := map[string]outputMeta{}
	for _, policy := range policies {
		outputs[policy.Name] = outputMeta{
			Path:                 filepath.ToSlash(filepath.Join(options.OutputDir, policy.Name+".txt")),
			PrefixCount:          len(publicLists.Operators[policy.Name]),
			BGPPrefixesByPurpose: admissionMatches[policy.Name],
			OriginASNs:           summarizeOrigins(origins[policy.Name], metadata),
		}
	}
	sources := map[string]sourceMeta{}
	for _, name := range []string{"riswhois_ipv6", "iptoasn_ipv6", "apnic_inet6num"} {
		var path string
		switch name {
		case "riswhois_ipv6":
			path = options.RISPath
		case "iptoasn_ipv6":
			path = options.IPToASNPath
		case "apnic_inet6num":
			path = options.Inet6numPath
		}
		meta, err := fileMetadata(path)
		if err != nil {
			return err
		}
		meta.Source = iputil.SourceURLs[iputil.SourceFilename(name)]
		sources[name] = meta
	}
	result := auditReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Sources:     sources,
		OutputUnit:  "canonical_merged_cidr",
		RegistryAdmission: registryAdmissionMeta{
			Descriptions:            []string{fixedBroadbandInet6numDescription, mobileInet6numDescription},
			MatchedInet6numRecords:  matchedInet6numRecords,
			MatchedInet6numPrefixes: matchedInet6numPrefixes,
			EffectiveRanges:         effectiveRanges,
		},
		CuratedAdmission: curatedAdmission,
		Outputs:          outputs,
		Rejected:         rejected,
	}
	if options.AuditReportPath != "" {
		if err := writeJSON(options.AuditReportPath, result); err != nil {
			return err
		}
	}
	return nil
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

func rangesFromPrefixes(byPurpose map[string][]string) []admissionRange {
	var out []admissionRange
	for purpose, values := range byPurpose {
		purposeName := strings.TrimSuffix(purpose, "_broadband")
		for _, value := range values {
			prefix := netip.MustParsePrefix(value).Masked()
			out = append(out, admissionRange{Lo: prefix.Addr(), Hi: iputil.LastAddress(prefix), Purpose: purposeName})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Lo.Compare(out[j].Lo) < 0 })
	return out
}

func summarizeOrigins(values map[string]*originAccumulator, metadata map[string]asMeta) []originMeta {
	var out []originMeta
	for asn, value := range values {
		sort.Strings(value.Prefixes)
		samples := value.Prefixes
		if len(samples) > 8 {
			samples = samples[:8]
		}
		out = append(out, originMeta{
			ASN:            asn,
			Description:    metadata[asn].Description,
			PrefixCount:    len(value.Prefixes),
			SamplePrefixes: append([]string(nil), samples...),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		a, _ := strconv.ParseUint(out[i].ASN, 10, 32)
		b, _ := strconv.ParseUint(out[j].ASN, 10, 32)
		return a < b
	})
	return out
}

func classifyPrefix(prefix netip.Prefix, ranges []admissionRange) (string, string) {
	lo, hi := prefix.Masked().Addr(), iputil.LastAddress(prefix)
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

func overlapsAdmissionRanges(prefix netip.Prefix, ranges []admissionRange) bool {
	lo, hi := prefix.Masked().Addr(), iputil.LastAddress(prefix)
	i := sort.Search(len(ranges), func(i int) bool { return ranges[i].Hi.Compare(lo) >= 0 })
	return i < len(ranges) && ranges[i].Lo.Compare(hi) <= 0
}

func registrationPurpose(record apnic6.Record) string {
	for _, description := range record.Descriptions {
		switch {
		case strings.EqualFold(strings.TrimSpace(description), fixedBroadbandInet6numDescription):
			return "fixed"
		case strings.EqualFold(strings.TrimSpace(description), mobileInet6numDescription):
			return "mobile"
		}
	}
	return ""
}

func allOriginsMatch(origins []string, metadata map[string]asMeta, classifier *operatorconfig.Classifier, operator string) bool {
	if len(origins) == 0 {
		return false
	}
	for _, asn := range origins {
		meta, ok := metadata[asn]
		if !ok || !strings.EqualFold(meta.Country, "CN") {
			return false
		}
		result := classifier.Classify(asn, meta.Description)
		if result.Operator != operator {
			return false
		}
		// In IPv6, the admitted business range proves that the address space is
		// terminal access space. CN2/CUII may still originate such a route as
		// transport, so their IPv4 address-origin exclusions do not apply here.
		if result.Excluded && asn != "4809" && asn != "9929" {
			return false
		}
	}
	return true
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
