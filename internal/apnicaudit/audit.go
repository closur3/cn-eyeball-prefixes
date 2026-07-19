package apnicaudit

import (
	"fmt"
	"net/netip"
	"sort"

	"github.com/closur3/cn-operator-allowlist/internal/apnicinetnum"
	"github.com/closur3/cn-operator-allowlist/internal/operatorconfig"
)

type Range struct {
	Lo uint32
	Hi uint32
}

type Registry struct {
	Range             string   `json:"range"`
	Netnames          []string `json:"netnames,omitempty"`
	Descriptions      []string `json:"descriptions,omitempty"`
	Organizations     []string `json:"organizations,omitempty"`
	OrganizationNames []string `json:"organization_names,omitempty"`
	Maintainers       []string `json:"maintainers,omitempty"`
	Country           string   `json:"country,omitempty"`
	Status            string   `json:"status,omitempty"`
	LastModified      string   `json:"last_modified,omitempty"`
}

type Fact struct {
	Start          string    `json:"start"`
	End            string    `json:"end"`
	AddressCount   uint64    `json:"address_count"`
	Operator       string    `json:"operator"`
	Classification string    `json:"classification"`
	Reason         string    `json:"reason"`
	MatchedBy      string    `json:"matched_by,omitempty"`
	Registry       *Registry `json:"registry,omitempty"`
}

type CIDRRecord struct {
	CIDR         string `json:"cidr"`
	AddressCount uint64 `json:"address_count"`
	Facts        []Fact `json:"facts"`
}

type CategorySummary struct {
	Classification string  `json:"classification"`
	FactCount      int     `json:"fact_count"`
	AddressCount   uint64  `json:"address_count"`
	AddressPercent float64 `json:"address_percent"`
}

type Summary struct {
	CIDRCount                     int               `json:"cidr_count"`
	FactCount                     int               `json:"fact_count"`
	AddressCount                  uint64            `json:"address_count"`
	RegistryCoveredAddressCount   uint64            `json:"registry_covered_address_count"`
	RegistryCoveragePercent       float64           `json:"registry_coverage_percent"`
	StrongNonPublicSignalAddressCount uint64        `json:"strong_non_public_signal_address_count"`
	Categories                    []CategorySummary `json:"categories"`
}

type Report struct {
	Scope string       `json:"scope"`
	Notes []string     `json:"notes"`
	Summary Summary    `json:"summary"`
	CIDRs []CIDRRecord `json:"cidrs"`
}

func Build(scope string, cidrs []string, operatorRanges map[string][]Range, segments []apnicinetnum.Segment, classifier *operatorconfig.Classifier) (Report, error) {
	report := Report{
		Scope: scope,
		Notes: []string{
			"Every retained address is mapped to the most-specific APNIC inetnum object available in the build snapshot.",
			"An independent legal-entity registration is an audit lead, not proof that the address is outside ordinary Internet user access scope.",
			"The emitted ACL CIDR may be a maximal aggregate and need not itself be visible as a BGP announcement.",
		},
	}
	categoryCounts := map[string]*CategorySummary{}
	operators := make([]string, 0, len(operatorRanges))
	for operator := range operatorRanges {
		operators = append(operators, operator)
		sort.Slice(operatorRanges[operator], func(i, j int) bool { return operatorRanges[operator][i].Lo < operatorRanges[operator][j].Lo })
	}
	sort.Strings(operators)

	for _, cidr := range cidrs {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil || !prefix.Addr().Is4() || prefix != prefix.Masked() {
			return Report{}, fmt.Errorf("invalid canonical IPv4 CIDR %q", cidr)
		}
		lo, hi := number(prefix.Addr()), prefixEnd(prefix)
		entry := CIDRRecord{CIDR: cidr, AddressCount: uint64(hi)-uint64(lo)+1}
		for _, operator := range operators {
			for _, candidate := range overlapping(operatorRanges[operator], lo, hi) {
				start, end := max(lo, candidate.Lo), min(hi, candidate.Hi)
				entry.Facts = append(entry.Facts, registryFacts(start, end, operator, segments, classifier)...)
			}
		}
		sort.Slice(entry.Facts, func(i, j int) bool { return number(netip.MustParseAddr(entry.Facts[i].Start)) < number(netip.MustParseAddr(entry.Facts[j].Start)) })
		var covered uint64
		for _, fact := range entry.Facts {
			covered += fact.AddressCount
			report.Summary.FactCount++
			if fact.Registry != nil {
				report.Summary.RegistryCoveredAddressCount += fact.AddressCount
			}
			if fact.Classification == "strong_non_public_signal" {
				report.Summary.StrongNonPublicSignalAddressCount += fact.AddressCount
			}
			summary := categoryCounts[fact.Classification]
			if summary == nil {
				summary = &CategorySummary{Classification: fact.Classification}
				categoryCounts[fact.Classification] = summary
			}
			summary.FactCount++
			summary.AddressCount += fact.AddressCount
		}
		if covered != entry.AddressCount {
			return Report{}, fmt.Errorf("CIDR %s has %d audited addresses, want %d", cidr, covered, entry.AddressCount)
		}
		report.Summary.AddressCount += entry.AddressCount
		report.CIDRs = append(report.CIDRs, entry)
	}
	report.Summary.CIDRCount = len(report.CIDRs)
	if report.Summary.AddressCount != 0 {
		report.Summary.RegistryCoveragePercent = percent(report.Summary.RegistryCoveredAddressCount, report.Summary.AddressCount)
	}
	order := []string{"operator_registration", "independent_legal_entity", "other_registration", "unregistered", "strong_non_public_signal"}
	for _, name := range order {
		if summary := categoryCounts[name]; summary != nil {
			summary.AddressPercent = percent(summary.AddressCount, report.Summary.AddressCount)
			report.Summary.Categories = append(report.Summary.Categories, *summary)
		}
	}
	return report, nil
}

func registryFacts(lo, hi uint32, operator string, segments []apnicinetnum.Segment, classifier *operatorconfig.Classifier) []Fact {
	i := sort.Search(len(segments), func(i int) bool { return segments[i].Hi >= lo })
	cursor := uint64(lo)
	limit := uint64(hi)
	var out []Fact
	for i < len(segments) && segments[i].Lo <= hi {
		segment := segments[i]
		start, end := max(lo, segment.Lo), min(hi, segment.Hi)
		if cursor < uint64(start) {
			out = append(out, uncoveredFact(uint32(cursor), start-1, operator))
		}
		classification, reason, matchedBy := classify(segment, classifier)
		out = append(out, Fact{
			Start: addr(start), End: addr(end), AddressCount: uint64(end)-uint64(start)+1,
			Operator: operator, Classification: classification, Reason: reason, MatchedBy: matchedBy,
			Registry: registry(segment.Record),
		})
		cursor = uint64(end) + 1
		i++
	}
	if cursor <= limit {
		out = append(out, uncoveredFact(uint32(cursor), hi, operator))
	}
	return out
}

func classify(segment apnicinetnum.Segment, classifier *operatorconfig.Classifier) (string, string, string) {
	if segment.Match.Reason != "" {
		return "strong_non_public_signal", segment.Match.Reason, segment.Match.MatchedBy
	}
	text := apnicinetnum.SearchText(segment.Record)
	registrant := classifier.Classify("0", text)
	if registrant.Excluded {
		return "strong_non_public_signal", registrant.Reason, registrant.MatchedBy
	}
	if registrant.Operator != "" {
		return "operator_registration", "APNIC registrant text matches "+registrant.Operator, registrant.MatchedBy
	}
	if classifier.IsIndependentLegalEntity(apnicinetnum.RegistrantText(segment.Record)) {
		return "independent_legal_entity", "APNIC registrant text names an independent legal entity; retained because registration alone is not sufficient exclusion evidence", "independent_legal_entity_patterns"
	}
	return "other_registration", "APNIC registration does not match an operator or a complete independent legal-entity pattern", ""
}

func registry(record apnicinetnum.Record) *Registry {
	return &Registry{
		Range: addr(record.Lo) + " - " + addr(record.Hi), Netnames: record.Netnames,
		Descriptions: record.Descriptions, Organizations: record.Organizations,
		OrganizationNames: record.OrganizationNames, Maintainers: record.Maintainers,
		Country: record.Country, Status: record.Status, LastModified: record.LastModified,
	}
}

func uncoveredFact(lo, hi uint32, operator string) Fact {
	return Fact{Start: addr(lo), End: addr(hi), AddressCount: uint64(hi)-uint64(lo)+1, Operator: operator, Classification: "unregistered", Reason: "No APNIC inetnum object covers this address range in the build snapshot"}
}

func overlapping(rows []Range, lo, hi uint32) []Range {
	i := sort.Search(len(rows), func(i int) bool { return rows[i].Hi >= lo })
	start := i
	for i < len(rows) && rows[i].Lo <= hi {
		i++
	}
	return rows[start:i]
}

func prefixEnd(prefix netip.Prefix) uint32 {
	lo := uint64(number(prefix.Addr()))
	size := uint64(1) << uint(32-prefix.Bits())
	return uint32(lo + size - 1)
}

func number(a netip.Addr) uint32 {
	b := a.As4()
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func addr(value uint32) string {
	return netip.AddrFrom4([4]byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)}).String()
}

func percent(part, total uint64) float64 {
	return float64(part) * 100 / float64(total)
}

func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}
