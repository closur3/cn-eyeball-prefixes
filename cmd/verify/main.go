package main

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/bits"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/closur3/cn-eyeball-prefixes/internal/apnicautnum"
	"github.com/closur3/cn-eyeball-prefixes/internal/apnicinetnum"
	"github.com/closur3/cn-eyeball-prefixes/internal/apnicorg"
	"github.com/closur3/cn-eyeball-prefixes/internal/apnicroute"
	"github.com/closur3/cn-eyeball-prefixes/internal/operatorconfig"
	"github.com/closur3/cn-eyeball-prefixes/internal/riswhois"
)

type span struct{ lo, hi uint32 }

var cloudSources = []string{
	"ipdata_aliyun", "ipdata_tencent", "ipdata_huawei", "ipdata_ucloud", "ipdata_ksyun", "ipdata_baidu", "ipdata_jdcloud",
}
var operators = []string{"chinanet", "cmcc", "unicom"}

const maxAdmissionCIDRExpansionRatio = 2.0
const maxConflictHealingCIDRRatio = 1.10
const maxConflictHealedAddressRatio = 0.001

type listMeta struct {
	Path         string `json:"path"`
	CIDRCount    int    `json:"cidr_count"`
	AddressCount uint64 `json:"address_count"`
	SHA256       string `json:"sha256"`
}

type sourceMeta struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type stageMeta struct {
	Name         string `json:"name"`
	CIDRCount    int    `json:"cidr_count"`
	AddressCount uint64 `json:"address_count"`
}

type operatorAdmissionMeta struct {
	Mode                      string  `json:"mode"`
	PreCIDRCount              int     `json:"pre_cidr_count"`
	DeniedCIDRCount           int     `json:"denied_cidr_count"`
	HierarchicalCIDRCount     int     `json:"hierarchical_cidr_count"`
	ConflictHealedCIDRCount   int     `json:"conflict_healed_cidr_count"`
	ConflictHealedAddressCount uint64 `json:"conflict_healed_address_count"`
	FinalCIDRCount            int     `json:"final_cidr_count"`
	CIDRExpansionRatio        float64 `json:"cidr_expansion_ratio"`
	MaximumCIDRExpansionRatio float64 `json:"maximum_cidr_expansion_ratio"`
	ConflictHealingCIDRRatio          float64 `json:"conflict_healing_cidr_ratio"`
	MaximumConflictHealingCIDRRatio   float64 `json:"maximum_conflict_healing_cidr_ratio"`
	ConflictHealedAddressRatio        float64 `json:"conflict_healed_address_ratio"`
	MaximumConflictHealedAddressRatio float64 `json:"maximum_conflict_healed_address_ratio"`
}

type cloudSourceMeta struct {
	Source                string `json:"source"`
	SourceCIDRCount       int    `json:"source_cidr_count"`
	SourceAddressCount    uint64 `json:"source_address_count"`
	EffectiveCIDRCount    int    `json:"effective_cidr_count"`
	EffectiveAddressCount uint64 `json:"effective_address_count"`
}

type apnicSourceMeta struct {
	RecordCount                int    `json:"record_count"`
	RelevantRecordCount        int    `json:"relevant_record_count"`
	MatchedWinningSegmentCount int    `json:"matched_winning_segment_count"`
	EffectiveCIDRCount         int    `json:"effective_cidr_count"`
	EffectiveAddressCount      uint64 `json:"effective_address_count"`
}
type portableHolderMeta struct {
	AutnumRecordCount          int    `json:"autnum_record_count"`
	MatchedWinningSegmentCount int    `json:"matched_winning_segment_count"`
	EffectiveCIDRCount         int    `json:"effective_cidr_count"`
	EffectiveAddressCount      uint64 `json:"effective_address_count"`
}

type delegatedHolderMeta struct {
	MatchedWinningSegmentCount int    `json:"matched_winning_segment_count"`
	EffectiveCIDRCount         int    `json:"effective_cidr_count"`
	EffectiveAddressCount      uint64 `json:"effective_address_count"`
}
type routeSourceMeta struct {
	ObjectCount                 int    `json:"object_count"`
	RelevantObjectCount         int    `json:"relevant_object_count"`
	RelevantWinningSegmentCount int    `json:"relevant_winning_segment_count"`
	OriginValidatedMatchCount   int    `json:"origin_validated_match_count"`
	EffectiveCIDRCount          int    `json:"effective_cidr_count"`
	EffectiveAddressCount       uint64 `json:"effective_address_count"`
}

type routeOriginAuditMeta struct {
	Enforced               bool                       `json:"enforced"`
	CandidateEvidenceCount int                        `json:"candidate_evidence_count"`
	CandidateCIDRCount     int                        `json:"candidate_cidr_count"`
	CandidateAddressCount  uint64                     `json:"candidate_address_count"`
	Candidates             []routeOriginCandidateMeta `json:"candidates"`
}

type routeOriginCandidateMeta struct {
	CIDR                      string   `json:"cidr"`
	AddressCount              uint64   `json:"address_count"`
	ASN                       string   `json:"asn"`
	Operator                  string   `json:"operator"`
	ASNDescription            string   `json:"asn_description"`
	RoutePrefix               string   `json:"route_prefix"`
	RouteOriginASN            string   `json:"route_origin_asn"`
	RouteOriginDescription    string   `json:"route_origin_description"`
	RouteOriginASName         string   `json:"route_origin_as_name"`
	Evidence                  string   `json:"evidence"`
	SharedOrganizations       []string `json:"shared_organizations,omitempty"`
	SharedMaintainers         []string `json:"shared_maintainers,omitempty"`
	RegistryNetnames          []string `json:"registry_netnames,omitempty"`
	RegistryDescriptions      []string `json:"registry_descriptions,omitempty"`
	RegistryOrganizations     []string `json:"registry_organizations,omitempty"`
	RegistryOrganizationNames []string `json:"registry_organization_names,omitempty"`
	RegistryMaintainers       []string `json:"registry_maintainers,omitempty"`
	RegistryStatus            string   `json:"registry_status,omitempty"`
	RegistryLastModified      string   `json:"registry_last_modified,omitempty"`
	RouteDescriptions         []string `json:"route_descriptions,omitempty"`
	RouteOrganizations        []string `json:"route_organizations,omitempty"`
	RouteOrganizationNames    []string `json:"route_organization_names,omitempty"`
	RouteMaintainers          []string `json:"route_maintainers,omitempty"`
	RouteLastModified         string   `json:"route_last_modified,omitempty"`
}
type risSourceMeta struct {
	RowCount                          int    `json:"row_count"`
	PrefixCount                       int    `json:"prefix_count"`
	RelevantPrefixCount               int    `json:"relevant_prefix_count"`
	WinningSegmentCount               int    `json:"winning_segment_count"`
	CandidateMOASSegmentCount         int    `json:"candidate_moas_segment_count"`
	StrongEvidenceSegmentCount        int    `json:"strong_evidence_segment_count"`
	RetainedAmbiguousMOASSegmentCount int    `json:"retained_ambiguous_moas_segment_count"`
	EffectiveCIDRCount                int    `json:"effective_cidr_count"`
	EffectiveAddressCount             uint64 `json:"effective_address_count"`
}
type observedOriginMeta struct {
	ASN         string `json:"asn"`
	Description string `json:"description"`
	SeenPeers   int    `json:"seen_peers"`
}
type linkedASNMeta struct {
	ASN         string `json:"asn"`
	Description string `json:"description"`
	ASName      string `json:"as_name"`
	Via         string `json:"via"`
}

type prefixExclusionMeta struct {
	Source                    string               `json:"source"`
	Category                  string               `json:"category"`
	Provider                  string               `json:"provider"`
	CIDR                      string               `json:"cidr"`
	AddressCount              uint64               `json:"address_count"`
	ASN                       string               `json:"asn"`
	Operator                  string               `json:"operator"`
	ASNDescription            string               `json:"asn_description"`
	RegistryNetnames          []string             `json:"registry_netnames"`
	RegistryDescriptions      []string             `json:"registry_descriptions"`
	RegistryOrganizations     []string             `json:"registry_organizations"`
	RegistryOrganizationNames []string             `json:"registry_organization_names"`
	RegistryMaintainers       []string             `json:"registry_maintainers"`
	RegistryStatus            string               `json:"registry_status"`
	RegistryLastModified      string               `json:"registry_last_modified"`
	RoutePrefix               string               `json:"route_prefix"`
	RouteOriginASN            string               `json:"route_origin_asn"`
	RouteOriginDescription    string               `json:"route_origin_description"`
	RouteOriginASName         string               `json:"route_origin_as_name"`
	Evidence                  string               `json:"evidence"`
	SharedOrganizations       []string             `json:"shared_organizations"`
	SharedMaintainers         []string             `json:"shared_maintainers"`
	RouteDescriptions         []string             `json:"route_descriptions"`
	RouteOrganizations        []string             `json:"route_organizations"`
	RouteOrganizationNames    []string             `json:"route_organization_names"`
	RouteMaintainers          []string             `json:"route_maintainers"`
	RouteLastModified         string               `json:"route_last_modified"`
	MatchedBy                 string               `json:"matched_by"`
	Reason                    string               `json:"reason"`
	ObservedOrigins           []observedOriginMeta `json:"observed_origins"`
	LinkedASNs                []linkedASNMeta      `json:"linked_asns"`
}

type manifest struct {
	Sources                       []sourceMeta          `json:"sources"`
	Stages                        []stageMeta           `json:"stages"`
	OperatorAdmission             operatorAdmissionMeta `json:"operator_registration_admission"`
	CloudSources                  []cloudSourceMeta     `json:"cloud_sources"`
	APNICInetnum                  apnicSourceMeta       `json:"apnic_inetnum"`
	APNICPortableHolders          portableHolderMeta    `json:"apnic_portable_holders"`
	APNICDelegatedHolders         delegatedHolderMeta   `json:"apnic_delegated_holders"`
	APNICIndependentLegalEntities delegatedHolderMeta   `json:"apnic_independent_legal_entities"`
	APNICRoute                    routeSourceMeta       `json:"apnic_route"`
	APNICRouteOriginAudit         routeOriginAuditMeta  `json:"apnic_route_origin_audit"`
	RISWhois                      risSourceMeta         `json:"ris_whois"`
	ExcludedPrefixes              []prefixExclusionMeta `json:"excluded_prefixes"`
	Lists                         []listMeta            `json:"lists"`
}

type allowedASNRecord struct {
	operator    string
	description string
	ranges      []span
}

func n(a netip.Addr) uint32 {
	return uint32(a.As4()[0])<<24 | uint32(a.As4()[1])<<16 | uint32(a.As4()[2])<<8 | uint32(a.As4()[3])
}

func readCIDRs(path string, ordered bool) []span {
	b, e := os.ReadFile(path)
	if e != nil {
		panic(e)
	}
	var out []span
	var prev uint32
	first := true
	for _, s := range strings.Fields(string(b)) {
		p, e := netip.ParsePrefix(s)
		if e != nil || !p.Addr().Is4() || p.Addr() != p.Masked().Addr() {
			panic("invalid CIDR: " + path)
		}
		lo := n(p.Addr())
		hi := uint32(uint64(lo) + (uint64(1) << uint(32-p.Bits())) - 1)
		if ordered && !first && lo <= prev {
			panic("unordered or overlapping: " + path)
		}
		first = false
		prev = hi
		out = append(out, span{lo, hi})
	}
	return out
}

func merge(in []span) []span {
	sort.Slice(in, func(i, j int) bool { return in[i].lo < in[j].lo })
	var out []span
	for _, x := range in {
		if len(out) == 0 || (out[len(out)-1].hi != ^uint32(0) && x.lo > out[len(out)-1].hi+1) {
			out = append(out, x)
			continue
		}
		if x.hi > out[len(out)-1].hi {
			out[len(out)-1].hi = x.hi
		}
	}
	return out
}

func intersect(a, b []span) []span {
	a, b = merge(a), merge(b)
	var out []span
	for i, j := 0, 0; i < len(a) && j < len(b); {
		lo, hi := a[i].lo, a[i].hi
		if b[j].lo > lo {
			lo = b[j].lo
		}
		if b[j].hi < hi {
			hi = b[j].hi
		}
		if lo <= hi {
			out = append(out, span{lo, hi})
		}
		if a[i].hi < b[j].hi {
			i++
		} else {
			j++
		}
	}
	return out
}

// overlapsSorted reports whether a normalized, address-sorted span set
// intersects [lo, hi] without repeatedly sorting it for point queries.
func overlapsSorted(rows []span, lo, hi uint32) bool {
	i := sort.Search(len(rows), func(i int) bool { return rows[i].hi >= lo })
	return i < len(rows) && rows[i].lo <= hi
}

func relevantAPNICRecords(records []apnicinetnum.Record, candidates []span) []apnicinetnum.Record {
	out := make([]apnicinetnum.Record, 0, len(records)/8)
	for _, record := range records {
		if overlapsSorted(candidates, record.Lo, record.Hi) {
			out = append(out, record)
		}
	}
	return out
}

func subtract(in, excluded []span) []span {
	in, excluded = merge(in), merge(excluded)
	var out []span
	j := 0
	for _, r := range in {
		for j < len(excluded) && excluded[j].hi < r.lo {
			j++
		}
		pos := r.lo
		covered := false
		for k := j; k < len(excluded) && excluded[k].lo <= r.hi; k++ {
			x := excluded[k]
			if x.hi < pos {
				continue
			}
			if x.lo > pos {
				out = append(out, span{pos, x.lo - 1})
			}
			if x.hi >= r.hi {
				covered = true
				break
			}
			pos = x.hi + 1
		}
		if !covered {
			out = append(out, span{pos, r.hi})
		}
	}
	return out
}

func containingAPNICSegment(segments []apnicinetnum.Segment, row span) *apnicinetnum.Segment {
	for i := range segments {
		if segments[i].Hi < row.lo {
			continue
		}
		if segments[i].Lo > row.lo {
			return nil
		}
		if segments[i].Hi >= row.hi {
			return &segments[i]
		}
	}
	return nil
}

func assertNoOverlap(a, b []span, message string) {
	for i, j := 0, 0; i < len(a) && j < len(b); {
		if a[i].hi < b[j].lo {
			i++
		} else if b[j].hi < a[i].lo {
			j++
		} else {
			panic(message)
		}
	}
}

func assertContained(a, b []span) {
	for i, j := 0, 0; i < len(a); {
		for j < len(b) && b[j].hi < a[i].lo {
			j++
		}
		if j == len(b) || b[j].lo > a[i].lo || b[j].hi < a[i].hi {
			panic("a generated list contains an address absent from its required superset")
		}
		i++
	}
}

func assertEqual(a, b []span, message string) {
	a, b = merge(a), merge(b)
	if len(a) != len(b) {
		panic(fmt.Sprintf("%s (left: %d ranges/%d addresses; right: %d ranges/%d addresses)", message, len(a), addressCount(a), len(b), addressCount(b)))
	}
	for i := range a {
		if a[i] != b[i] {
			panic(fmt.Sprintf("%s (first mismatch: %+v vs %+v)", message, a[i], b[i]))
		}
	}
}

func addressCount(rows []span) uint64 {
	var count uint64
	for _, row := range merge(rows) {
		count += uint64(row.hi) - uint64(row.lo) + 1
	}
	return count
}

func apnicOperatorAdmissionRanges(records []apnicinetnum.Record, classifier *operatorconfig.Classifier) map[string][]span {
	out := map[string][]span{}
	for _, record := range records {
		result := classifier.ClassifyAPNICRegistrant(apnicinetnum.SearchText(record))
		if result.Operator != "" {
			out[result.Operator] = append(out[result.Operator], span{record.Lo, record.Hi})
		}
	}
	for _, operator := range operators {
		out[operator] = merge(out[operator])
	}
	return out
}

func apnicOperatorConflictRanges(segments []apnicinetnum.Segment, classifier *operatorconfig.Classifier) map[string][]span {
	out := map[string][]span{}
	for _, segment := range segments {
		result := classifier.ClassifyAPNICRegistrant(apnicinetnum.SearchText(segment.Record))
		if result.Operator == "" {
			continue
		}
		for _, operator := range operators {
			if operator != result.Operator {
				out[operator] = append(out[operator], span{segment.Lo, segment.Hi})
			}
		}
	}
	for _, operator := range operators {
		out[operator] = merge(out[operator])
	}
	return out
}

func intersectSortedSpan(rows []span, lo, hi uint32) []span {
	i := sort.Search(len(rows), func(i int) bool { return rows[i].hi >= lo })
	var out []span
	for ; i < len(rows) && rows[i].lo <= hi; i++ {
		out = append(out, span{max32(lo, rows[i].lo), min32(hi, rows[i].hi)})
	}
	return out
}

func overlapAddressCountSorted(rows []span, lo, hi uint32) uint64 {
	var count uint64
	for _, hit := range intersectSortedSpan(rows, lo, hi) {
		count += uint64(hit.hi) - uint64(hit.lo) + 1
	}
	return count
}

func bgpConflictHealingRanges(segments []riswhois.Segment, asnOperators map[string]string, originByOperator, retainedByOperator, parentAdmission map[string][]span) (observed, eligible []span) {
	for _, segment := range segments {
		seenOperators := map[string]bool{}
		for _, origin := range segment.Record.Origins {
			operator := asnOperators[origin.ASN]
			if operator == "" || seenOperators[operator] {
				continue
			}
			seenOperators[operator] = true
			unit := intersectSortedSpan(originByOperator[operator], segment.Lo, segment.Hi)
			if len(unit) == 0 {
				continue
			}
			var retained []span
			for _, part := range unit {
				retained = append(retained, intersectSortedSpan(retainedByOperator[operator], part.lo, part.hi)...)
			}
			if len(retained) == 0 {
				continue
			}
			observed = append(observed, retained...)
			var parentPositive, total uint64
			for _, part := range unit {
				parentPositive += overlapAddressCountSorted(parentAdmission[operator], part.lo, part.hi)
				total += uint64(part.hi) - uint64(part.lo) + 1
			}
			if total == 0 {
				continue
			}
			if parentPositive == total {
				eligible = append(eligible, retained...)
			}
		}
	}
	return merge(observed), merge(eligible)
}

func conflictHealedAdmissionByOperator(hierarchicalByOperator map[string][]span, conflictHealingEligible []span, eligibleByOperator map[string][]span) map[string][]span {
	out := map[string][]span{}
	for _, operator := range operators {
		healed := intersect(conflictHealingEligible, eligibleByOperator[operator])
		out[operator] = merge(append(append([]span{}, hierarchicalByOperator[operator]...), healed...))
	}
	return out
}

func cidrCount(rows []span) int {
	count := 0
	for _, row := range merge(rows) {
		r := row
		for r.lo <= r.hi {
			remaining := uint64(r.hi) - uint64(r.lo) + 1
			align := bits.TrailingZeros32(r.lo)
			if r.lo == 0 {
				align = 32
			}
			sizeBits := align
			if max := bits.Len64(remaining) - 1; max < sizeBits {
				sizeBits = max
			}
			size := uint64(1) << uint(sizeBits)
			count++
			if size == remaining {
				break
			}
			r.lo += uint32(size)
		}
	}
	return count
}

func fileSHA(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func sourcePath(dir, name string) string {
	if name == "iptoasn_ipv4" {
		return filepath.Join(dir, name+".tsv.gz")
	}
	if name == "apnic_inetnum" || name == "apnic_autnum" || name == "apnic_organisation" || name == "apnic_route" || name == "riswhois_ipv4" {
		return filepath.Join(dir, name+".gz")
	}
	return filepath.Join(dir, name+".txt")
}

func operatorRanges(path string, classifier *operatorconfig.Classifier) (map[string][]span, map[string]*allowedASNRecord, map[string]string) {
	f, e := os.Open(path)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		panic(e)
	}
	defer z.Close()
	out := map[string][]span{}
	records := map[string]*allowedASNRecord{}
	descriptions := map[string]string{}
	scanner := bufio.NewScanner(z)
	for scanner.Scan() {
		x := strings.SplitN(scanner.Text(), "\t", 5)
		if len(x) != 5 {
			continue
		}
		if descriptions[x[2]] == "" {
			descriptions[x[2]] = x[4]
		}
		if x[3] != "CN" {
			continue
		}
		operator := classifier.Match(x[2], x[4])
		if operator == "" {
			continue
		}
		a, ea := netip.ParseAddr(x[0])
		b, eb := netip.ParseAddr(x[1])
		if ea != nil || eb != nil || !a.Is4() || !b.Is4() {
			panic("invalid IPtoASN range")
		}
		rangeValue := span{n(a), n(b)}
		out[operator] = append(out[operator], rangeValue)
		record := records[x[2]]
		if record == nil {
			record = &allowedASNRecord{operator: operator, description: x[4]}
			records[x[2]] = record
		}
		record.ranges = append(record.ranges, rangeValue)
	}
	if e := scanner.Err(); e != nil {
		panic(e)
	}
	for _, operator := range operators {
		out[operator] = merge(out[operator])
	}
	for _, record := range records {
		record.ranges = merge(record.ranges)
	}
	return out, records, descriptions
}

func independentAutnumLinks(record apnicinetnum.Record, index *apnicautnum.Index, classifier *operatorconfig.Classifier, descriptions map[string]string) []linkedASNMeta {
	var out []linkedASNMeta
	for _, link := range index.Links(record.Netnames, record.Organizations) {
		result := classifier.Classify(link.ASN, descriptions[link.ASN])
		if result.Operator == "" || result.Excluded {
			out = append(out, linkedASNMeta{link.ASN, descriptions[link.ASN], link.ASName, link.Via})
		}
	}
	return out
}

func auditIndependentRouteOrigins(inetSegments []apnicinetnum.Segment, routeSegments []apnicroute.Segment, finalRanges []span, allowed map[string]*allowedASNRecord, index *apnicautnum.Index, classifier *operatorconfig.Classifier, descriptions map[string]string) ([]routeOriginCandidateMeta, []span) {
	var candidates []routeOriginCandidateMeta
	var candidateRanges []span
	seen := map[string]bool{}
	inetStart := 0
	for _, routeSegment := range routeSegments {
		for inetStart < len(inetSegments) && inetSegments[inetStart].Hi < routeSegment.Lo {
			inetStart++
		}
		for i := inetStart; i < len(inetSegments) && inetSegments[i].Lo <= routeSegment.Hi; i++ {
			inetSegment := inetSegments[i]
			lo, hi := max32(inetSegment.Lo, routeSegment.Lo), min32(inetSegment.Hi, routeSegment.Hi)
			base := intersect(finalRanges, []span{{lo: lo, hi: hi}})
			if len(base) == 0 {
				continue
			}
			for _, variant := range routeSegment.Record.Variants {
				autnum, ok := index.Record(variant.Origin)
				if !ok {
					continue
				}
				classification := classifier.Classify(variant.Origin, descriptions[variant.Origin])
				if classification.Operator != "" && !classification.Excluded {
					continue
				}
				sharedOrganizations := apnicautnum.CommonAll(inetSegment.Record.Organizations, variant.Organizations, autnum.Organizations)
				var sharedMaintainers []string
				for _, maintainer := range apnicautnum.CommonAll(inetSegment.Record.Maintainers, variant.Maintainers, autnum.Maintainers) {
					if index.DedicatedMaintainer(variant.Origin, maintainer) {
						sharedMaintainers = append(sharedMaintainers, maintainer)
					}
				}
				if len(sharedOrganizations) == 0 && len(sharedMaintainers) == 0 {
					continue
				}
				evidence := "shared_organisation_handle"
				if len(sharedOrganizations) == 0 {
					evidence = "shared_dedicated_maintainer"
				} else if len(sharedMaintainers) != 0 {
					evidence = "shared_organisation_handle_and_dedicated_maintainer"
				}
				for asn, current := range allowed {
					for _, hit := range intersect(base, current.ranges) {
						for _, cidr := range spanCIDRs([]span{hit}) {
							key := asn + "\x00" + cidr + "\x00" + routeSegment.Record.Prefix + "\x00" + variant.Origin + "\x00" + evidence
							if seen[key] {
								continue
							}
							seen[key] = true
							prefix := netip.MustParsePrefix(cidr)
							candidates = append(candidates, routeOriginCandidateMeta{
								CIDR: cidr, AddressCount: uint64(1) << uint(32-prefix.Bits()), ASN: asn, Operator: current.operator, ASNDescription: current.description,
								RoutePrefix: routeSegment.Record.Prefix, RouteOriginASN: variant.Origin, RouteOriginDescription: descriptions[variant.Origin], RouteOriginASName: autnum.ASName,
								Evidence: evidence, SharedOrganizations: sharedOrganizations, SharedMaintainers: sharedMaintainers,
								RegistryNetnames: inetSegment.Record.Netnames, RegistryDescriptions: inetSegment.Record.Descriptions, RegistryOrganizations: inetSegment.Record.Organizations, RegistryOrganizationNames: inetSegment.Record.OrganizationNames, RegistryMaintainers: inetSegment.Record.Maintainers, RegistryStatus: inetSegment.Record.Status, RegistryLastModified: inetSegment.Record.LastModified,
								RouteDescriptions: variant.Descriptions, RouteOrganizations: variant.Organizations, RouteOrganizationNames: variant.OrganizationNames, RouteMaintainers: variant.Maintainers, RouteLastModified: variant.LastModified,
							})
							candidateRanges = append(candidateRanges, hit)
						}
					}
				}
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.Operator != b.Operator {
			return a.Operator < b.Operator
		}
		if a.ASN != b.ASN {
			if len(a.ASN) != len(b.ASN) {
				return len(a.ASN) < len(b.ASN)
			}
			return a.ASN < b.ASN
		}
		ai, bi := n(netip.MustParsePrefix(a.CIDR).Addr()), n(netip.MustParsePrefix(b.CIDR).Addr())
		if ai != bi {
			return ai < bi
		}
		if a.RouteOriginASN != b.RouteOriginASN {
			return a.RouteOriginASN < b.RouteOriginASN
		}
		return a.Evidence < b.Evidence
	})
	return candidates, merge(candidateRanges)
}

func routeOriginExclusionMeta(candidate routeOriginCandidateMeta) prefixExclusionMeta {
	return prefixExclusionMeta{
		Source: "apnic_route", Category: "apnic_route_origin", CIDR: candidate.CIDR, AddressCount: candidate.AddressCount,
		ASN: candidate.ASN, Operator: candidate.Operator, ASNDescription: candidate.ASNDescription,
		RegistryNetnames: candidate.RegistryNetnames, RegistryDescriptions: candidate.RegistryDescriptions,
		RegistryOrganizations: candidate.RegistryOrganizations, RegistryOrganizationNames: candidate.RegistryOrganizationNames,
		RegistryMaintainers: candidate.RegistryMaintainers, RegistryStatus: candidate.RegistryStatus, RegistryLastModified: candidate.RegistryLastModified,
		RoutePrefix: candidate.RoutePrefix, RouteOriginASN: candidate.RouteOriginASN,
		RouteOriginDescription: candidate.RouteOriginDescription, RouteOriginASName: candidate.RouteOriginASName,
		Evidence: candidate.Evidence, SharedOrganizations: candidate.SharedOrganizations, SharedMaintainers: candidate.SharedMaintainers,
		RouteDescriptions: candidate.RouteDescriptions, RouteOrganizations: candidate.RouteOrganizations,
		RouteOrganizationNames: candidate.RouteOrganizationNames, RouteMaintainers: candidate.RouteMaintainers, RouteLastModified: candidate.RouteLastModified,
		MatchedBy: "APNIC independent route origin: " + candidate.Evidence,
		Reason:    "APNIC inetnum, route, and aut-num records strongly link the prefix to active independent AS" + candidate.RouteOriginASN,
	}
}

func spanCIDRs(rows []span) []string {
	rows = merge(rows)
	var lines []string
	for _, row := range rows {
		r := row
		for r.lo <= r.hi {
			remaining := uint64(r.hi) - uint64(r.lo) + 1
			align := bits.TrailingZeros32(r.lo)
			if r.lo == 0 {
				align = 32
			}
			sizeBits := align
			if max := bits.Len64(remaining) - 1; max < sizeBits {
				sizeBits = max
			}
			size := uint64(1) << uint(sizeBits)
			a := netip.AddrFrom4([4]byte{byte(r.lo >> 24), byte(r.lo >> 16), byte(r.lo >> 8), byte(r.lo)})
			lines = append(lines, netip.PrefixFrom(a, 32-sizeBits).String())
			if size == remaining {
				break
			}
			r.lo += uint32(size)
		}
	}
	return lines
}

func min32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func main() {
	data := flag.String("data", "data", "data directory")
	sources := flag.String("sources", "", "source directory")
	operatorConfig := flag.String("operator-config", "config/operators.json", "operator classification config")
	flag.Parse()
	if *sources == "" {
		panic("--sources is required")
	}
	pipelineStarted, phaseStarted := time.Now(), time.Now()
	logPhase := func(name string) {
		now := time.Now()
		fmt.Printf("timing: %-28s %s\n", name, now.Sub(phaseStarted).Round(time.Millisecond))
		phaseStarted = now
	}

	provinceFiles, e := filepath.Glob(filepath.Join(*data, "provinces", "*.txt"))
	if e != nil {
		panic(e)
	}
	if len(provinceFiles) != 31 {
		panic("expected cn.txt plus exactly 31 provincial combined lists")
	}
	operatorFiles, e := filepath.Glob(filepath.Join(*data, "operators", "*.txt"))
	if e != nil {
		panic(e)
	}
	if len(operatorFiles) != len(operators) {
		panic("expected exactly three per-operator lists")
	}
	for _, obsolete := range []string{
		filepath.Join(*data, "experiments"),
		filepath.Join(*data, "audits"),
	} {
		if _, e := os.Stat(obsolete); !os.IsNotExist(e) {
			panic("formal output contains an obsolete experiment artifact: " + obsolete)
		}
	}
	files := append(append(append([]string{}, provinceFiles...), operatorFiles...), filepath.Join(*data, "cn.txt"))

	for _, f := range files {
		readCIDRs(f, true)
	}

	var cloudRanges []span
	cloudBySource := map[string][]span{}
	for _, source := range cloudSources {
		cloudBySource[source] = readCIDRs(filepath.Join(*sources, source+".txt"), false)
		cloudRanges = append(cloudRanges, cloudBySource[source]...)
	}
	cnRanges := readCIDRs(filepath.Join(*data, "cn.txt"), true)
	classifier, e := operatorconfig.Load(*operatorConfig, operators)
	if e != nil {
		panic(e)
	}
	chinaRanges := readCIDRs(filepath.Join(*sources, "china.txt"), false)
	assertContained(cnRanges, chinaRanges)
	allowedByOperator, allowedByASN, asnDescriptions := operatorRanges(filepath.Join(*sources, "iptoasn_ipv4.tsv.gz"), classifier)
	var allowedOperators []span
	for _, operator := range operators {
		allowedOperators = append(allowedOperators, allowedByOperator[operator]...)
	}
	assertContained(cnRanges, merge(allowedOperators))
	preCloudCandidates := intersect(allowedOperators, chinaRanges)
	postCloudCandidates := subtract(preCloudCandidates, cloudRanges)
	logPhase("outputs, origin, and cloud")
	orgNames, e := apnicorg.Parse(filepath.Join(*sources, "apnic_organisation.gz"))
	if e != nil {
		panic(e)
	}
	apnicRecords, e := apnicinetnum.Parse(filepath.Join(*sources, "apnic_inetnum.gz"))
	if e != nil {
		panic(e)
	}
	apnicRecordCount := len(apnicRecords)
	apnicRecords = relevantAPNICRecords(apnicRecords, postCloudCandidates)
	apnicinetnum.AttachOrganizationNames(apnicRecords, orgNames)
	autnumRecords, e := apnicautnum.Parse(filepath.Join(*sources, "apnic_autnum.gz"))
	if e != nil {
		panic(e)
	}
	autnumIndex := apnicautnum.NewIndex(autnumRecords, asnDescriptions)
	registryAutnumIndex := apnicautnum.NewRegistryIndex(autnumRecords)
	apnicAllSegments := apnicinetnum.ResolveAll(apnicRecords, func(record apnicinetnum.Record) apnicinetnum.Match {
		result := classifier.ClassifyAPNICInetnum(apnicinetnum.SearchText(record))
		if result.Excluded {
			return apnicinetnum.Match{Category: "apnic_inetnum", Reason: result.Reason, MatchedBy: result.MatchedBy}
		}
		status := strings.ToUpper(record.Status)
		registrant := classifier.Classify("0", apnicinetnum.SearchText(record))
		independent := (registrant.Operator == "" || registrant.Excluded) && len(independentAutnumLinks(record, autnumIndex, classifier, asnDescriptions)) != 0
		if (status == "ALLOCATED PORTABLE" || status == "ASSIGNED PORTABLE") && independent {
			return apnicinetnum.Match{Category: "apnic_portable_holder", Reason: "Most-specific APNIC portable registration is linked to a currently active independent ASN", MatchedBy: "APNIC portable holder linked through aut-num"}
		}
		if (status == "ALLOCATED NON-PORTABLE" || status == "ASSIGNED NON-PORTABLE") && independent {
			return apnicinetnum.Match{Category: "apnic_delegated_holder", Reason: "Most-specific APNIC non-portable registration is linked to a currently active independent ASN", MatchedBy: "APNIC delegated holder linked through aut-num"}
		}
		registryLinks := independentAutnumLinks(record, registryAutnumIndex, classifier, asnDescriptions)
		if (registrant.Operator == "" || registrant.Excluded) && classifier.IsIndependentLegalEntity(apnicinetnum.RegistrantText(record)) && len(registryLinks) != 0 {
			return apnicinetnum.Match{Category: "apnic_independent_legal_entity_holder", Reason: "Most-specific APNIC registration names an independent legal entity and is exactly linked to an APNIC aut-num", MatchedBy: "APNIC legal entity plus exact aut-num org/netname link"}
		}
		return apnicinetnum.Match{}
	})
	apnicSegments := apnicinetnum.Matched(apnicAllSegments)
	var matchedAPNICRanges, matchedPurposeRanges, matchedPortableRanges, matchedDelegatedRanges, matchedLegalEntityRanges []span
	purposeSegments, portableSegments, delegatedSegments, legalEntityHolderSegments := 0, 0, 0, 0
	for _, segment := range apnicSegments {
		matchedAPNICRanges = append(matchedAPNICRanges, span{segment.Lo, segment.Hi})
		switch segment.Match.Category {
		case "apnic_portable_holder":
			portableSegments++
			matchedPortableRanges = append(matchedPortableRanges, span{segment.Lo, segment.Hi})
		case "apnic_delegated_holder":
			delegatedSegments++
			matchedDelegatedRanges = append(matchedDelegatedRanges, span{segment.Lo, segment.Hi})
		case "apnic_independent_legal_entity_holder":
			legalEntityHolderSegments++
			matchedLegalEntityRanges = append(matchedLegalEntityRanges, span{segment.Lo, segment.Hi})
		default:
			purposeSegments++
			matchedPurposeRanges = append(matchedPurposeRanges, span{segment.Lo, segment.Hi})
		}
	}
	apnicRanges := intersect(postCloudCandidates, matchedAPNICRanges)
	apnicPurposeRanges := intersect(postCloudCandidates, matchedPurposeRanges)
	portableHolderRanges := intersect(postCloudCandidates, matchedPortableRanges)
	delegatedHolderRanges := intersect(postCloudCandidates, matchedDelegatedRanges)
	legalEntityHolderRanges := intersect(postCloudCandidates, matchedLegalEntityRanges)
	preRouteExcluded := merge(append(append([]span{}, cloudRanges...), apnicRanges...))
	logPhase("APNIC inetnum and aut-num")
	routeRecords, routeObjects, relevantRouteObjects, e := apnicroute.Parse(filepath.Join(*sources, "apnic_route.gz"), orgNames, func(lo, hi uint32) bool { return overlapsSorted(postCloudCandidates, lo, hi) })
	if e != nil {
		panic(e)
	}
	routeSegments := apnicroute.Resolve(routeRecords)
	var routeRanges []span
	routeMatches := 0
	for _, segment := range routeSegments {
		for _, variant := range segment.Record.Variants {
			record := allowedByASN[variant.Origin]
			if record == nil {
				continue
			}
			result := classifier.ClassifyAPNICInetnum(apnicroute.SearchText(variant))
			if !result.Excluded {
				continue
			}
			hits := intersect(subtract(intersect(record.ranges, chinaRanges), preRouteExcluded), []span{{segment.Lo, segment.Hi}})
			if len(hits) > 0 {
				routeMatches++
				routeRanges = append(routeRanges, hits...)
			}
		}
	}
	routeRanges = merge(routeRanges)
	preRISExcluded := merge(append(append([]span{}, preRouteExcluded...), routeRanges...))
	routeOriginCandidates, routeOriginCandidateRanges := auditIndependentRouteOrigins(apnicAllSegments, routeSegments, subtract(preCloudCandidates, preRISExcluded), allowedByASN, autnumIndex, classifier, asnDescriptions)
	preRISExcluded = merge(append(preRISExcluded, routeOriginCandidateRanges...))
	logPhase("APNIC route")
	risRecords, risStats, e := riswhois.Parse(filepath.Join(*sources, "riswhois_ipv4.gz"), func(lo, hi uint32) bool { return overlapsSorted(postCloudCandidates, lo, hi) })
	if e != nil {
		panic(e)
	}
	risSegments := riswhois.Resolve(risRecords)
	var risRanges []span
	candidateMOAS, strongMOAS := 0, 0
	for _, segment := range risSegments {
		if len(segment.Record.Origins) < 2 {
			continue
		}
		maxPeers := 0
		for _, o := range segment.Record.Origins {
			if o.SeenPeers > maxPeers {
				maxPeers = o.SeenPeers
			}
		}
		for asn, record := range allowedByASN {
			current := 0
			for _, o := range segment.Record.Origins {
				if o.ASN == asn {
					current = o.SeenPeers
				}
			}
			if current < 10 || current*20 < maxPeers {
				continue
			}
			hits := intersect(subtract(intersect(record.ranges, chinaRanges), preRISExcluded), []span{{segment.Lo, segment.Hi}})
			if len(hits) == 0 {
				continue
			}
			candidateMOAS++
			strong := false
			for _, o := range segment.Record.Origins {
				if o.ASN == asn || o.SeenPeers < 10 || o.SeenPeers*20 < maxPeers {
					continue
				}
				if classifier.ClassifyAPNICInetnum(asnDescriptions[o.ASN]).Excluded {
					strong = true
					break
				}
			}
			if strong {
				strongMOAS++
				risRanges = append(risRanges, hits...)
			}
		}
	}
	risRanges = merge(risRanges)
	excludedRanges := merge(append(append([]span{}, preRISExcluded...), risRanges...))
	logPhase("RIPE RISWhois")
	assertNoOverlap(cnRanges, excludedRanges, "cn.txt overlaps an explicit cloud, APNIC, independent route-origin, or strong RIS MOAS exclusion")
	operatorAdmissionRanges := apnicOperatorAdmissionRanges(apnicRecords, classifier)
	operatorConflictRanges := apnicOperatorConflictRanges(apnicAllSegments, classifier)
	preAdmissionByOperator := map[string][]span{}
	hierarchicalByOperator := map[string][]span{}
	admissionDeniedByOperator := map[string][]span{}
	expectedByOperator := map[string][]span{}
	var preAdmissionRanges, hierarchicalRanges []span
	for _, operator := range operators {
		preAdmissionByOperator[operator] = subtract(intersect(allowedByOperator[operator], chinaRanges), excludedRanges)
		parentAdmitted := intersect(preAdmissionByOperator[operator], operatorAdmissionRanges[operator])
		hierarchicalByOperator[operator] = subtract(parentAdmitted, operatorConflictRanges[operator])
		preAdmissionRanges = append(preAdmissionRanges, preAdmissionByOperator[operator]...)
		hierarchicalRanges = append(hierarchicalRanges, hierarchicalByOperator[operator]...)
	}
	preAdmissionRanges = merge(preAdmissionRanges)
	hierarchicalRanges = merge(hierarchicalRanges)
	asnOperators := map[string]string{}
	for asn, record := range allowedByASN {
		asnOperators[asn] = record.operator
	}
	originByOperator := map[string][]span{}
	for _, operator := range operators {
		originByOperator[operator] = intersect(allowedByOperator[operator], chinaRanges)
	}
	observedBGPRanges, conflictHealingEligibleRanges := bgpConflictHealingRanges(risSegments, asnOperators, originByOperator, preAdmissionByOperator, operatorAdmissionRanges)
	expectedByOperator = conflictHealedAdmissionByOperator(hierarchicalByOperator, conflictHealingEligibleRanges, preAdmissionByOperator)
	var admissionDeniedRanges, expectedCN, assignedRanges []span
	for _, operator := range operators {
		assertNoOverlap(expectedByOperator[operator], merge(assignedRanges), "conflict-healed per-operator address sets overlap")
		assignedRanges = append(assignedRanges, expectedByOperator[operator]...)
		expectedCN = append(expectedCN, expectedByOperator[operator]...)
		admissionDeniedByOperator[operator] = subtract(preAdmissionByOperator[operator], expectedByOperator[operator])
		admissionDeniedRanges = append(admissionDeniedRanges, admissionDeniedByOperator[operator]...)
	}
	expectedCN = merge(expectedCN)
	admissionDeniedRanges = merge(admissionDeniedRanges)
	bgpConflictHealedAdded := subtract(expectedCN, hierarchicalRanges)
	bgpConflictHealedRemoved := subtract(hierarchicalRanges, expectedCN)
	if len(bgpConflictHealedRemoved) != 0 {
		panic("BGP conflict healing removes addresses from the hierarchical baseline")
	}
	assertContained(bgpConflictHealedAdded, preAdmissionRanges)
	assertContained(bgpConflictHealedAdded, observedBGPRanges)
	assertContained(bgpConflictHealedAdded, conflictHealingEligibleRanges)
	var allOperatorConflicts []span
	for _, operator := range operators {
		allOperatorConflicts = append(allOperatorConflicts, operatorConflictRanges[operator]...)
	}
	assertContained(bgpConflictHealedAdded, merge(allOperatorConflicts))
	assertNoOverlap(expectedCN, excludedRanges, "BGP conflict healing overlaps an enforced strong exclusion")
	preAdmissionCIDRCount := cidrCount(preAdmissionRanges)
	hierarchicalCIDRCount := cidrCount(hierarchicalRanges)
	finalCIDRCount := cidrCount(expectedCN)
	if finalCIDRCount > preAdmissionCIDRCount*2 {
		panic("operator parent-registration admission exceeds the 2.0x CIDR expansion limit")
	}
	admissionExpansionRatio := float64(finalCIDRCount) / float64(preAdmissionCIDRCount)
	conflictHealingCIDRRatio := float64(finalCIDRCount) / float64(hierarchicalCIDRCount)
	conflictHealedAddressRatio := float64(addressCount(bgpConflictHealedAdded)) / float64(addressCount(hierarchicalRanges))
	if conflictHealingCIDRRatio > maxConflictHealingCIDRRatio {
		panic("BGP conflict healing exceeds the 1.10x CIDR expansion limit")
	}
	if conflictHealedAddressRatio > maxConflictHealedAddressRatio {
		panic("BGP conflict healing exceeds the 0.1% added-address limit")
	}
	assertEqual(cnRanges, expectedCN, "cn.txt address set does not equal the recomputed final output")
	var generatedOperators []span
	generatedByOperator := map[string][]span{}
	for _, operator := range operators {
		path := filepath.Join(*data, "operators", operator+".txt")
		ranges := readCIDRs(path, true)
		generatedByOperator[operator] = ranges
		expected := expectedByOperator[operator]
		assertEqual(ranges, expected, "operator address set does not recompute: "+operator)
		assertContained(ranges, cnRanges)
		assertContained(ranges, allowedByOperator[operator])
		assertNoOverlap(ranges, merge(generatedOperators), "per-operator lists overlap")
		generatedOperators = append(generatedOperators, ranges...)
	}
	assertEqual(generatedOperators, cnRanges, "the union of per-operator lists does not equal cn.txt")
	var provincialRanges []span
	for _, f := range provinceFiles {
		ranges := readCIDRs(f, true)
		assertContained(ranges, cnRanges)
		assertNoOverlap(ranges, merge(provincialRanges), "provincial lists overlap")
		provincialRanges = append(provincialRanges, ranges...)
	}
	manifestPath := filepath.Join(*data, "manifest.json")
	b, e := os.ReadFile(manifestPath)
	if e != nil {
		panic(e)
	}
	var m manifest
	if e := json.Unmarshal(b, &m); e != nil {
		panic(e)
	}
	if m.OperatorAdmission.Mode != "bgp_registration_conflict_healing_with_strong_exclusions" || m.OperatorAdmission.PreCIDRCount != preAdmissionCIDRCount || m.OperatorAdmission.DeniedCIDRCount != cidrCount(admissionDeniedRanges) || m.OperatorAdmission.HierarchicalCIDRCount != hierarchicalCIDRCount || m.OperatorAdmission.ConflictHealedCIDRCount != cidrCount(bgpConflictHealedAdded) || m.OperatorAdmission.ConflictHealedAddressCount != addressCount(bgpConflictHealedAdded) || m.OperatorAdmission.FinalCIDRCount != finalCIDRCount || m.OperatorAdmission.CIDRExpansionRatio != admissionExpansionRatio || m.OperatorAdmission.MaximumCIDRExpansionRatio != maxAdmissionCIDRExpansionRatio || m.OperatorAdmission.ConflictHealingCIDRRatio != conflictHealingCIDRRatio || m.OperatorAdmission.MaximumConflictHealingCIDRRatio != maxConflictHealingCIDRRatio || m.OperatorAdmission.ConflictHealedAddressRatio != conflictHealedAddressRatio || m.OperatorAdmission.MaximumConflictHealedAddressRatio != maxConflictHealedAddressRatio {
		panic("manifest BGP conflict-healing metadata mismatch")
	}
	expectedSourceNames := append([]string{"china", "iptoasn_ipv4", "apnic_organisation", "apnic_inetnum", "apnic_autnum", "apnic_route", "riswhois_ipv4", "ip2region_ipv4_source"}, cloudSources...)
	if len(m.Sources) != len(expectedSourceNames)+1 {
		panic("manifest source count mismatch")
	}
	for i, entry := range m.Sources {
		var path string
		if i == 0 {
			if entry.Name != "operator_config" || entry.Path != filepath.ToSlash(*operatorConfig) {
				panic("manifest operator config source mismatch")
			}
			path = *operatorConfig
		} else {
			if entry.Name != expectedSourceNames[i-1] {
				panic("manifest source order mismatch")
			}
			path = sourcePath(*sources, entry.Name)
		}
		info, e := os.Stat(path)
		if e != nil || entry.Bytes != info.Size() || entry.SHA256 != fileSHA(path) {
			panic("manifest source metadata mismatch for " + entry.Name)
		}
	}
	var originCandidates []span
	for _, operator := range operators {
		originCandidates = append(originCandidates, allowedByOperator[operator]...)
	}
	expectedStages := []struct {
		name string
		rows []span
	}{
		{"operator_origin_candidates", originCandidates},
		{"china_origin_intersection", preCloudCandidates},
		{"effective_cloud_prefix_exclusions", intersect(preCloudCandidates, cloudRanges)},
		{"effective_apnic_prefix_exclusions", apnicPurposeRanges},
		{"effective_apnic_portable_holder_exclusions", portableHolderRanges},
		{"effective_apnic_delegated_holder_exclusions", delegatedHolderRanges},
		{"effective_apnic_independent_legal_entity_exclusions", legalEntityHolderRanges},
		{"effective_apnic_route_exclusions", routeRanges},
		{"effective_apnic_independent_route_origin_exclusions", routeOriginCandidateRanges},
		{"effective_ris_moas_exclusions", risRanges},
		{"pre_operator_parent_registration_admission", preAdmissionRanges},
		{"operator_parent_registration_admissions", hierarchicalRanges},
		{"bgp_conflict_healed_additions", bgpConflictHealedAdded},
		{"bgp_conflict_healing_admissions", cnRanges},
		{"operator_registration_admission_denials", admissionDeniedRanges},
		{"final_output", cnRanges},
		{"province_attributed_output", provincialRanges},
	}
	if len(m.Stages) != len(expectedStages) {
		panic("manifest stage count mismatch")
	}
	for i, expected := range expectedStages {
		entry := m.Stages[i]
		if entry.Name != expected.name || entry.CIDRCount != cidrCount(expected.rows) || entry.AddressCount != addressCount(expected.rows) {
			panic("manifest stage metadata mismatch for " + expected.name)
		}
	}
	if len(m.CloudSources) != len(cloudSources) {
		panic("manifest cloud source count mismatch")
	}
	for i, name := range cloudSources {
		entry := m.CloudSources[i]
		ranges := cloudBySource[name]
		effective := intersect(preCloudCandidates, ranges)
		if entry.Source != name || entry.SourceCIDRCount != cidrCount(ranges) || entry.SourceAddressCount != addressCount(ranges) || entry.EffectiveCIDRCount != cidrCount(effective) || entry.EffectiveAddressCount != addressCount(effective) {
			panic("manifest cloud source metadata mismatch for " + name)
		}
	}
	if m.APNICInetnum.RecordCount != apnicRecordCount || m.APNICInetnum.RelevantRecordCount != len(apnicRecords) || m.APNICInetnum.MatchedWinningSegmentCount != purposeSegments || m.APNICInetnum.EffectiveCIDRCount != cidrCount(apnicPurposeRanges) || m.APNICInetnum.EffectiveAddressCount != addressCount(apnicPurposeRanges) {
		panic("manifest APNIC inetnum metadata mismatch")
	}
	if m.APNICPortableHolders.AutnumRecordCount != len(autnumRecords) || m.APNICPortableHolders.MatchedWinningSegmentCount != portableSegments || m.APNICPortableHolders.EffectiveCIDRCount != cidrCount(portableHolderRanges) || m.APNICPortableHolders.EffectiveAddressCount != addressCount(portableHolderRanges) {
		panic("manifest APNIC portable-holder metadata mismatch")
	}
	if m.APNICDelegatedHolders.MatchedWinningSegmentCount != delegatedSegments || m.APNICDelegatedHolders.EffectiveCIDRCount != cidrCount(delegatedHolderRanges) || m.APNICDelegatedHolders.EffectiveAddressCount != addressCount(delegatedHolderRanges) {
		panic("manifest APNIC delegated-holder metadata mismatch")
	}
	if m.APNICIndependentLegalEntities.MatchedWinningSegmentCount != legalEntityHolderSegments || m.APNICIndependentLegalEntities.EffectiveCIDRCount != cidrCount(legalEntityHolderRanges) || m.APNICIndependentLegalEntities.EffectiveAddressCount != addressCount(legalEntityHolderRanges) {
		panic("manifest APNIC independent legal-entity metadata mismatch")
	}
	if m.APNICRoute.ObjectCount != routeObjects || m.APNICRoute.RelevantObjectCount != relevantRouteObjects || m.APNICRoute.RelevantWinningSegmentCount != len(routeSegments) || m.APNICRoute.OriginValidatedMatchCount != routeMatches || m.APNICRoute.EffectiveCIDRCount != cidrCount(routeRanges) || m.APNICRoute.EffectiveAddressCount != addressCount(routeRanges) {
		panic("manifest APNIC route metadata mismatch")
	}
	if !m.APNICRouteOriginAudit.Enforced || m.APNICRouteOriginAudit.CandidateEvidenceCount != len(routeOriginCandidates) || m.APNICRouteOriginAudit.CandidateCIDRCount != cidrCount(routeOriginCandidateRanges) || m.APNICRouteOriginAudit.CandidateAddressCount != addressCount(routeOriginCandidateRanges) || !reflect.DeepEqual(m.APNICRouteOriginAudit.Candidates, routeOriginCandidates) {
		panic("manifest APNIC independent route-origin audit mismatch")
	}
	if m.RISWhois.RowCount != risStats.Rows || m.RISWhois.PrefixCount != risStats.Prefixes || m.RISWhois.RelevantPrefixCount != risStats.RelevantPrefixes || m.RISWhois.WinningSegmentCount != len(risSegments) || m.RISWhois.CandidateMOASSegmentCount != candidateMOAS || m.RISWhois.StrongEvidenceSegmentCount != strongMOAS || m.RISWhois.RetainedAmbiguousMOASSegmentCount != candidateMOAS-strongMOAS || m.RISWhois.EffectiveCIDRCount != cidrCount(risRanges) || m.RISWhois.EffectiveAddressCount != addressCount(risRanges) {
		panic("manifest RIPE RIS metadata mismatch")
	}
	var manifestExcludedRanges []span
	for _, entry := range m.ExcludedPrefixes {
		prefix, e := netip.ParsePrefix(entry.CIDR)
		if e != nil || !prefix.Addr().Is4() || prefix != prefix.Masked() {
			panic("invalid excluded prefix in manifest: " + entry.CIDR)
		}
		row := span{n(prefix.Addr()), uint32(uint64(n(prefix.Addr())) + (uint64(1) << uint(32-prefix.Bits())) - 1)}
		if entry.AddressCount != uint64(row.hi)-uint64(row.lo)+1 {
			panic("excluded prefix address count mismatch: " + entry.CIDR)
		}
		asnRecord := allowedByASN[entry.ASN]
		if asnRecord == nil || entry.Operator != asnRecord.operator || entry.ASNDescription != asnRecord.description {
			panic("excluded prefix ASN metadata mismatch: " + entry.CIDR)
		}
		assertContained([]span{row}, intersect(asnRecord.ranges, chinaRanges))
		switch entry.Category {
		case "cloud_provider_cidr":
			sourceRanges, ok := cloudBySource[entry.Source]
			provider := strings.TrimPrefix(entry.Source, "ipdata_")
			if !ok || entry.Provider != provider || entry.MatchedBy != "IP-Data provider CIDR" || entry.Reason != "Prefix is explicitly listed by IP-Data for "+provider {
				panic("excluded cloud prefix metadata mismatch: " + entry.CIDR)
			}
			assertContained([]span{row}, merge(sourceRanges))
		case "apnic_inetnum":
			if entry.Source != "apnic_inetnum" || entry.Provider != "" {
				panic("excluded APNIC prefix source mismatch: " + entry.CIDR)
			}
			segment := containingAPNICSegment(apnicSegments, row)
			if segment == nil || entry.MatchedBy != segment.Match.MatchedBy || entry.Reason != segment.Match.Reason || !reflect.DeepEqual(entry.RegistryNetnames, segment.Record.Netnames) || !reflect.DeepEqual(entry.RegistryDescriptions, segment.Record.Descriptions) || !reflect.DeepEqual(entry.RegistryOrganizations, segment.Record.Organizations) || !reflect.DeepEqual(entry.RegistryOrganizationNames, segment.Record.OrganizationNames) || !reflect.DeepEqual(entry.RegistryMaintainers, segment.Record.Maintainers) || entry.RegistryStatus != segment.Record.Status || entry.RegistryLastModified != segment.Record.LastModified {
				panic("excluded APNIC prefix registration metadata mismatch: " + entry.CIDR)
			}
			assertContained([]span{row}, apnicPurposeRanges)
		case "apnic_portable_holder":
			if entry.Source != "apnic_autnum" || entry.Provider != "" || entry.MatchedBy != "APNIC portable holder linked through aut-num" || entry.Reason != "Most-specific APNIC portable registration is linked to a currently active independent ASN" {
				panic("excluded APNIC portable-holder metadata mismatch: " + entry.CIDR)
			}
			segment := containingAPNICSegment(apnicSegments, row)
			if segment == nil || segment.Match.Category != "apnic_portable_holder" || entry.RegistryStatus != segment.Record.Status || !equalStrings(entry.RegistryNetnames, segment.Record.Netnames) || !equalStrings(entry.RegistryDescriptions, segment.Record.Descriptions) || !equalStrings(entry.RegistryOrganizations, segment.Record.Organizations) || !equalStrings(entry.RegistryOrganizationNames, segment.Record.OrganizationNames) || !equalStrings(entry.RegistryMaintainers, segment.Record.Maintainers) || entry.RegistryLastModified != segment.Record.LastModified || !reflect.DeepEqual(entry.LinkedASNs, independentAutnumLinks(segment.Record, autnumIndex, classifier, asnDescriptions)) {
				panic("excluded APNIC portable-holder evidence does not recompute: " + entry.CIDR)
			}
			assertContained([]span{row}, portableHolderRanges)
		case "apnic_delegated_holder":
			if entry.Source != "apnic_autnum" || entry.Provider != "" || entry.MatchedBy != "APNIC delegated holder linked through aut-num" || entry.Reason != "Most-specific APNIC non-portable registration is linked to a currently active independent ASN" {
				panic("excluded APNIC delegated-holder metadata mismatch: " + entry.CIDR)
			}
			segment := containingAPNICSegment(apnicSegments, row)
			if segment == nil || segment.Match.Category != "apnic_delegated_holder" || entry.RegistryStatus != segment.Record.Status || !equalStrings(entry.RegistryNetnames, segment.Record.Netnames) || !equalStrings(entry.RegistryDescriptions, segment.Record.Descriptions) || !equalStrings(entry.RegistryOrganizations, segment.Record.Organizations) || !equalStrings(entry.RegistryOrganizationNames, segment.Record.OrganizationNames) || !equalStrings(entry.RegistryMaintainers, segment.Record.Maintainers) || entry.RegistryLastModified != segment.Record.LastModified || !reflect.DeepEqual(entry.LinkedASNs, independentAutnumLinks(segment.Record, autnumIndex, classifier, asnDescriptions)) {
				panic("excluded APNIC delegated-holder evidence does not recompute: " + entry.CIDR)
			}
			assertContained([]span{row}, delegatedHolderRanges)
		case "apnic_independent_legal_entity_holder":
			if entry.Source != "apnic_autnum" || entry.Provider != "" || entry.MatchedBy != "APNIC legal entity plus exact aut-num org/netname link" || entry.Reason != "Most-specific APNIC registration names an independent legal entity and is exactly linked to an APNIC aut-num" {
				panic("excluded APNIC independent legal-entity metadata mismatch: " + entry.CIDR)
			}
			segment := containingAPNICSegment(apnicSegments, row)
			if segment == nil || segment.Match.Category != "apnic_independent_legal_entity_holder" || !classifier.IsIndependentLegalEntity(apnicinetnum.RegistrantText(segment.Record)) || entry.RegistryStatus != segment.Record.Status || !equalStrings(entry.RegistryNetnames, segment.Record.Netnames) || !equalStrings(entry.RegistryDescriptions, segment.Record.Descriptions) || !equalStrings(entry.RegistryOrganizations, segment.Record.Organizations) || !equalStrings(entry.RegistryOrganizationNames, segment.Record.OrganizationNames) || !equalStrings(entry.RegistryMaintainers, segment.Record.Maintainers) || entry.RegistryLastModified != segment.Record.LastModified || !reflect.DeepEqual(entry.LinkedASNs, independentAutnumLinks(segment.Record, registryAutnumIndex, classifier, asnDescriptions)) {
				panic("excluded APNIC independent legal-entity evidence does not recompute: " + entry.CIDR)
			}
			assertContained([]span{row}, legalEntityHolderRanges)
		case "apnic_route":
			if entry.Source != "apnic_route" || entry.Provider != "" {
				panic("excluded APNIC route metadata mismatch: " + entry.CIDR)
			}
			matched := false
			for _, segment := range routeSegments {
				if segment.Lo > row.lo || segment.Hi < row.hi {
					continue
				}
				for _, variant := range segment.Record.Variants {
					if variant.Origin != entry.ASN {
						continue
					}
					result := classifier.ClassifyAPNICInetnum(apnicroute.SearchText(variant))
					if result.Excluded && entry.MatchedBy == result.MatchedBy+"; current BGP origin matches APNIC route origin" && entry.Reason == result.Reason && equalStrings(entry.RegistryDescriptions, variant.Descriptions) && equalStrings(entry.RegistryOrganizations, variant.Organizations) && equalStrings(entry.RegistryOrganizationNames, variant.OrganizationNames) && equalStrings(entry.RegistryMaintainers, variant.Maintainers) && entry.RegistryLastModified == variant.LastModified {
						matched = true
					}
				}
			}
			if !matched {
				panic("excluded APNIC route evidence does not recompute: " + entry.CIDR)
			}
			assertContained([]span{row}, routeRanges)
		case "apnic_route_origin":
			if entry.Source != "apnic_route" || entry.Provider != "" {
				panic("excluded APNIC independent route-origin metadata mismatch: " + entry.CIDR)
			}
			matched := false
			for _, candidate := range routeOriginCandidates {
				if reflect.DeepEqual(entry, routeOriginExclusionMeta(candidate)) {
					matched = true
					break
				}
			}
			if !matched {
				panic("excluded APNIC independent route-origin evidence does not recompute: " + entry.CIDR)
			}
			assertContained([]span{row}, routeOriginCandidateRanges)
		case "ris_moas":
			if entry.Source != "riswhois_ipv4" || entry.Provider != "" || len(entry.ObservedOrigins) < 2 {
				panic("excluded RIS MOAS metadata mismatch: " + entry.CIDR)
			}
			matched := false
			for _, segment := range risSegments {
				if segment.Lo > row.lo || segment.Hi < row.hi {
					continue
				}
				maxPeers := 0
				currentSeen := 0
				observed := make([]observedOriginMeta, 0, len(segment.Record.Origins))
				for _, origin := range segment.Record.Origins {
					if origin.SeenPeers > maxPeers {
						maxPeers = origin.SeenPeers
					}
					if origin.ASN == entry.ASN {
						currentSeen = origin.SeenPeers
					}
					observed = append(observed, observedOriginMeta{origin.ASN, asnDescriptions[origin.ASN], origin.SeenPeers})
				}
				if currentSeen < 10 || currentSeen*20 < maxPeers || !reflect.DeepEqual(observed, entry.ObservedOrigins) {
					continue
				}
				for _, origin := range segment.Record.Origins {
					if origin.ASN == entry.ASN || origin.SeenPeers < 10 || origin.SeenPeers*20 < maxPeers {
						continue
					}
					result := classifier.ClassifyAPNICInetnum(asnDescriptions[origin.ASN])
					if result.Excluded && entry.MatchedBy == result.MatchedBy+"; RIPE RIS multi-observer MOAS" && entry.Reason == "Alternate origin AS"+origin.ASN+" is strongly identified as outside ordinary Internet user access scope: "+result.Reason {
						matched = true
						break
					}
				}
			}
			if !matched {
				panic("excluded RIS MOAS evidence does not recompute: " + entry.CIDR)
			}
			assertContained([]span{row}, risRanges)
		default:
			panic("unknown excluded prefix category: " + entry.Category)
		}
		manifestExcludedRanges = append(manifestExcludedRanges, row)
	}
	assertEqual(manifestExcludedRanges, intersect(preCloudCandidates, excludedRanges), "manifest excluded-prefix union mismatch")
	if len(m.Lists) != len(files) {
		panic("manifest list count does not match generated files")
	}
	seen := map[string]bool{}
	for _, entry := range m.Lists {
		path := filepath.Join(*data, filepath.FromSlash(entry.Path))
		ranges := readCIDRs(path, true)
		if seen[entry.Path] || entry.CIDRCount != len(strings.Fields(string(mustRead(path)))) || entry.AddressCount != addressCount(ranges) || entry.SHA256 != fileSHA(path) {
			panic("manifest metadata mismatch for " + entry.Path)
		}
		seen[entry.Path] = true
	}
	for _, path := range files {
		rel, e := filepath.Rel(*data, path)
		if e != nil || !seen[filepath.ToSlash(rel)] {
			panic("generated file is missing from manifest: " + path)
		}
	}
	logPhase("output and manifest checks")
	fmt.Printf("timing: %-28s %s\n", "total", time.Since(pipelineStarted).Round(time.Millisecond))
	fmt.Println("OK: all lists and manifest metadata are valid; operator/province relations, China boundary, ASN policy, cloud CIDRs, APNIC inetnum, portable/delegated-holder aut-num, origin-validated and strongly linked independent route-origin exclusions, and conservative RIPE RIS MOAS exclusions hold.")
}

func mustRead(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
