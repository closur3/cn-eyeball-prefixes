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

type fileMeta struct {
	CIDRCount    int    `json:"cidr_count"`
	AddressCount uint64 `json:"address_count"`
	SHA256       string `json:"sha256"`
}

type sourceMeta struct {
	Name   string `json:"name"`
	URL    string `json:"url,omitempty"`
	Path   string `json:"path,omitempty"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type listMeta struct {
	Name string `json:"name"`
	Path string `json:"path"`
	fileMeta
}

type excludedASNMeta struct {
	ASN                string `json:"asn"`
	Operator           string `json:"operator"`
	Description        string `json:"description"`
	MatchedBy          string `json:"matched_by"`
	ExclusionSource    string `json:"exclusion_source"`
	Reason             string `json:"reason"`
	OriginAddressCount uint64 `json:"origin_address_count"`
}

type includedASNMeta struct {
	ASN                  string `json:"asn"`
	Operator             string `json:"operator"`
	Description          string `json:"description"`
	MatchedBy            string `json:"matched_by"`
	OriginAddressCount   uint64 `json:"origin_address_count"`
	RetainedAddressCount uint64 `json:"retained_address_count"`
}

type includedASNRecord struct {
	meta   includedASNMeta
	ranges []span
}

type asnName struct {
	ASN         string `json:"asn"`
	Description string `json:"description"`
}

type operatorSummary struct {
	Operator             string    `json:"operator"`
	ASNCount             int       `json:"asn_count"`
	RetainedAddressCount uint64    `json:"retained_address_count"`
	ASNs                 []asnName `json:"asns"`
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
	Provider              string `json:"provider"`
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
	Description string `json:"description,omitempty"`
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
	Provider                  string               `json:"provider,omitempty"`
	CIDR                      string               `json:"cidr"`
	AddressCount              uint64               `json:"address_count"`
	ASN                       string               `json:"asn"`
	Operator                  string               `json:"operator"`
	ASNDescription            string               `json:"asn_description"`
	RegistryNetnames          []string             `json:"registry_netnames,omitempty"`
	RegistryDescriptions      []string             `json:"registry_descriptions,omitempty"`
	RegistryOrganizations     []string             `json:"registry_organizations,omitempty"`
	RegistryOrganizationNames []string             `json:"registry_organization_names,omitempty"`
	RegistryMaintainers       []string             `json:"registry_maintainers,omitempty"`
	RegistryStatus            string               `json:"registry_status,omitempty"`
	RegistryLastModified      string               `json:"registry_last_modified,omitempty"`
	RoutePrefix               string               `json:"route_prefix,omitempty"`
	RouteOriginASN            string               `json:"route_origin_asn,omitempty"`
	RouteOriginDescription    string               `json:"route_origin_description,omitempty"`
	RouteOriginASName         string               `json:"route_origin_as_name,omitempty"`
	Evidence                  string               `json:"evidence,omitempty"`
	SharedOrganizations       []string             `json:"shared_organizations,omitempty"`
	SharedMaintainers         []string             `json:"shared_maintainers,omitempty"`
	RouteDescriptions         []string             `json:"route_descriptions,omitempty"`
	RouteOrganizations        []string             `json:"route_organizations,omitempty"`
	RouteOrganizationNames    []string             `json:"route_organization_names,omitempty"`
	RouteMaintainers          []string             `json:"route_maintainers,omitempty"`
	RouteLastModified         string               `json:"route_last_modified,omitempty"`
	MatchedBy                 string               `json:"matched_by"`
	Reason                    string               `json:"reason"`
	ObservedOrigins           []observedOriginMeta `json:"observed_origins,omitempty"`
	LinkedASNs                []linkedASNMeta      `json:"linked_asns,omitempty"`
}

type manifest struct {
	GeneratedAt                   string                `json:"generated_at"`
	Scope                         string                `json:"scope"`
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
	OperatorSummary               []operatorSummary     `json:"operator_summary"`
	IncludedASNs                  []includedASNMeta     `json:"included_asns"`
	ExcludedASNs                  []excludedASNMeta     `json:"excluded_asns"`
	Lists                         []listMeta            `json:"lists"`
}

type province struct {
	Name string
	Slug string
}

var operators = []string{"chinatelecom", "chinamobile", "chinaunicom"}

const maxAdmissionCIDRExpansionRatio = 2.0
const maxConflictHealingCIDRRatio = 1.10
const maxConflictHealedAddressRatio = 0.001

var cloudSources = []string{
	"ipdata_aliyun", "ipdata_tencent", "ipdata_huawei", "ipdata_ucloud", "ipdata_ksyun", "ipdata_baidu", "ipdata_jdcloud",
}
var provinces = []province{
	{"北京市", "beijing"},
	{"天津市", "tianjin"},
	{"河北省", "hebei"},
	{"山西省", "shanxi"},
	{"内蒙古自治区", "neimenggu"},
	{"辽宁省", "liaoning"},
	{"吉林省", "jilin"},
	{"黑龙江省", "heilongjiang"},
	{"上海市", "shanghai"},
	{"江苏省", "jiangsu"},
	{"浙江省", "zhejiang"},
	{"安徽省", "anhui"},
	{"福建省", "fujian"},
	{"江西省", "jiangxi"},
	{"山东省", "shandong"},
	{"河南省", "henan"},
	{"湖北省", "hubei"},
	{"湖南省", "hunan"},
	{"广东省", "guangdong"},
	{"广西壮族自治区", "guangxi"},
	{"海南省", "hainan"},
	{"重庆市", "chongqing"},
	{"四川省", "sichuan"},
	{"贵州省", "guizhou"},
	{"云南省", "yunnan"},
	{"西藏自治区", "xizang"},
	{"陕西省", "shaanxi"},
	{"甘肃省", "gansu"},
	{"青海省", "qinghai"},
	{"宁夏回族自治区", "ningxia"},
	{"新疆维吾尔自治区", "xinjiang"},
}
var aliases = map[string]string{"北京": "北京市", "天津": "天津市", "上海": "上海市", "重庆": "重庆市", "内蒙古": "内蒙古自治区", "广西": "广西壮族自治区", "宁夏": "宁夏回族自治区", "新疆": "新疆维吾尔自治区", "西藏": "西藏自治区"}
var urls = map[string]string{
	"china":                 "https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china.txt",
	"iptoasn_ipv4":          "https://iptoasn.com/data/ip2asn-v4.tsv.gz",
	"ip2region_ipv4_source": "https://raw.githubusercontent.com/lionsoul2014/ip2region/master/data/ipv4_source.txt",
	"apnic_inetnum":         "https://ftp.apnic.net/apnic/whois/apnic.db.inetnum.gz",
	"apnic_autnum":          "https://ftp.apnic.net/apnic/whois/apnic.db.aut-num.gz",
	"apnic_organisation":    "https://ftp.apnic.net/apnic/whois/apnic.db.organisation.gz",
	"apnic_route":           "https://ftp.apnic.net/apnic/whois/apnic.db.route.gz",
	"riswhois_ipv4":         "https://www.ris.ripe.net/dumps/riswhoisdump.IPv4.gz",
	"ipdata_aliyun":         "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/aliyun-cidr-ipv4.txt",
	"ipdata_tencent":        "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/tencent-cidr-ipv4.txt",
	"ipdata_huawei":         "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/huawei-cidr-ipv4.txt",
	"ipdata_ucloud":         "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/ucloud-cidr-ipv4.txt",
	"ipdata_ksyun":          "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/ksyun-cidr-ipv4.txt",
	"ipdata_baidu":          "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/baidu-cidr-ipv4.txt",
	"ipdata_jdcloud":        "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/jdcloud-cidr-ipv4.txt",
}

func n(a netip.Addr) uint32 {
	return uint32(a.As4()[0])<<24 | uint32(a.As4()[1])<<16 | uint32(a.As4()[2])<<8 | uint32(a.As4()[3])
}
func end(p netip.Prefix) uint32 {
	return uint32(uint64(n(p.Addr())) + (uint64(1) << uint(32-p.Bits())) - 1)
}

func merge(in []span) []span {
	sort.Slice(in, func(i, j int) bool { return in[i].lo < in[j].lo })
	out := []span{}
	for _, x := range in {
		if len(out) == 0 {
			out = append(out, x)
			continue
		}
		last := &out[len(out)-1]
		if last.hi != ^uint32(0) && x.lo > last.hi+1 {
			out = append(out, x)
			continue
		}
		if x.hi > last.hi {
			last.hi = x.hi
		}
	}
	return out
}

func subtract(in, excluded []span) []span {
	in = merge(in)
	excluded = merge(excluded)
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

func addressCount(rows []span) uint64 {
	var count uint64
	for _, row := range merge(rows) {
		count += uint64(row.hi) - uint64(row.lo) + 1
	}
	return count
}

func lessOperatorASN(aOperator, aASN, bOperator, bASN string) bool {
	aRank, bRank := len(operators), len(operators)
	for i, operator := range operators {
		if aOperator == operator {
			aRank = i
		}
		if bOperator == operator {
			bRank = i
		}
	}
	if aRank != bRank {
		return aRank < bRank
	}
	if len(aASN) != len(bASN) {
		return len(aASN) < len(bASN)
	}
	return aASN < bASN
}

func cidrs(path string) ([]span, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var out []span
	for i, s := range strings.Fields(string(b)) {
		p, e := netip.ParsePrefix(s)
		if e != nil || !p.Addr().Is4() || p.Addr() != p.Masked().Addr() {
			return nil, fmt.Errorf("%s:%d", path, i+1)
		}
		out = append(out, span{n(p.Addr()), end(p)})
	}
	return merge(out), nil
}

func operatorRanges(path string, classifier *operatorconfig.Classifier) (map[string][]span, map[string]*includedASNRecord, []excludedASNMeta, map[string]string, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, nil, nil, nil, e
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		return nil, nil, nil, nil, e
	}
	defer z.Close()

	out := map[string][]span{}
	included := map[string]*includedASNRecord{}
	excluded := map[string]*excludedASNMeta{}
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
		result := classifier.Classify(x[2], x[4])
		if result.Operator == "" {
			continue
		}
		a, ea := netip.ParseAddr(x[0])
		z, ez := netip.ParseAddr(x[1])
		if ea != nil || ez != nil || !a.Is4() || !z.Is4() {
			return nil, nil, nil, nil, fmt.Errorf("invalid IPtoASN range: %s", scanner.Text())
		}
		row := span{n(a), n(z)}
		if result.Excluded {
			entry := excluded[x[2]]
			if entry == nil {
				entry = &excludedASNMeta{
					ASN: x[2], Operator: result.Operator, Description: x[4], MatchedBy: result.MatchedBy,
					ExclusionSource: result.ExclusionSource, Reason: result.Reason,
				}
				excluded[x[2]] = entry
			}
			entry.OriginAddressCount += uint64(row.hi) - uint64(row.lo) + 1
			continue
		}
		entry := included[x[2]]
		if entry == nil {
			entry = &includedASNRecord{meta: includedASNMeta{ASN: x[2], Operator: result.Operator, Description: x[4], MatchedBy: result.MatchedBy}}
			included[x[2]] = entry
		} else if entry.meta.Operator != result.Operator {
			return nil, nil, nil, nil, fmt.Errorf("ASN %s classified as both %s and %s", x[2], entry.meta.Operator, result.Operator)
		}
		entry.ranges = append(entry.ranges, row)
		entry.meta.OriginAddressCount += uint64(row.hi) - uint64(row.lo) + 1
		out[result.Operator] = append(out[result.Operator], row)
	}
	if e := scanner.Err(); e != nil {
		return nil, nil, nil, nil, e
	}
	for _, o := range operators {
		out[o] = merge(out[o])
		if len(out[o]) == 0 {
			return nil, nil, nil, nil, fmt.Errorf("IPtoASN source produced no ranges for %s", o)
		}
	}
	var excludedList []excludedASNMeta
	for _, entry := range excluded {
		excludedList = append(excludedList, *entry)
	}
	sort.Slice(excludedList, func(i, j int) bool {
		return lessOperatorASN(excludedList[i].Operator, excludedList[i].ASN, excludedList[j].Operator, excludedList[j].ASN)
	})
	return out, included, excludedList, descriptions, nil
}

func includedASNList(records map[string]*includedASNRecord, chinaRanges, excludedRanges []span, admittedByOperator map[string][]span) []includedASNMeta {
	var out []includedASNMeta
	for _, record := range records {
		retained := subtract(intersect(record.ranges, chinaRanges), excludedRanges)
		retained = intersect(retained, admittedByOperator[record.meta.Operator])
		record.meta.RetainedAddressCount = addressCount(retained)
		if record.meta.RetainedAddressCount != 0 {
			out = append(out, record.meta)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return lessOperatorASN(out[i].Operator, out[i].ASN, out[j].Operator, out[j].ASN)
	})
	return out
}

func independentAutnumLinks(record apnicinetnum.Record, index *apnicautnum.Index, classifier *operatorconfig.Classifier, descriptions map[string]string) []linkedASNMeta {
	var out []linkedASNMeta
	for _, link := range index.Links(record.Netnames, record.Organizations) {
		result := classifier.Classify(link.ASN, descriptions[link.ASN])
		if result.Operator == "" || result.Excluded {
			out = append(out, linkedASNMeta{ASN: link.ASN, Description: descriptions[link.ASN], ASName: link.ASName, Via: link.Via})
		}
	}
	return out
}

func auditIndependentRouteOrigins(inetSegments []apnicinetnum.Segment, routeSegments []apnicroute.Segment, finalRanges []span, included map[string]*includedASNRecord, index *apnicautnum.Index, classifier *operatorconfig.Classifier, descriptions map[string]string) ([]routeOriginCandidateMeta, []span) {
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
				for asn, current := range included {
					for _, hit := range intersect(base, current.ranges) {
						for _, cidr := range spanCIDRs([]span{hit}) {
							key := asn + "\x00" + cidr + "\x00" + routeSegment.Record.Prefix + "\x00" + variant.Origin + "\x00" + evidence
							if seen[key] {
								continue
							}
							seen[key] = true
							prefix := netip.MustParsePrefix(cidr)
							candidates = append(candidates, routeOriginCandidateMeta{
								CIDR: cidr, AddressCount: uint64(1) << uint(32-prefix.Bits()), ASN: asn, Operator: current.meta.Operator, ASNDescription: current.meta.Description,
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
		if a.Operator != b.Operator || a.ASN != b.ASN {
			return lessOperatorASN(a.Operator, a.ASN, b.Operator, b.ASN)
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

func summarizeOperators(included []includedASNMeta) []operatorSummary {
	var out []operatorSummary
	for _, operator := range operators {
		summary := operatorSummary{Operator: operator}
		for _, entry := range included {
			if entry.Operator != operator {
				continue
			}
			summary.ASNs = append(summary.ASNs, asnName{ASN: entry.ASN, Description: entry.Description})
			summary.RetainedAddressCount += entry.RetainedAddressCount
		}
		summary.ASNCount = len(summary.ASNs)
		out = append(out, summary)
	}
	return out
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

func sha(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func source(path, name, url, manifestPath string) (sourceMeta, error) {
	sum, err := sha(path)
	if err != nil {
		return sourceMeta{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return sourceMeta{}, err
	}
	return sourceMeta{Name: name, URL: url, Path: manifestPath, Bytes: info.Size(), SHA256: sum}, nil
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

func stage(name string, rows []span) stageMeta {
	return stageMeta{Name: name, CIDRCount: len(spanCIDRs(rows)), AddressCount: addressCount(rows)}
}

func cloudProvider(source string) string {
	return strings.TrimPrefix(source, "ipdata_")
}

func write(path string, rows []span) (fileMeta, error) {
	lines := spanCIDRs(rows)
	return writeLines(path, lines, addressCount(rows))
}

func writeLines(path string, lines []string, addresses uint64) (fileMeta, error) {
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return fileMeta{}, e
	}
	if e := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); e != nil {
		return fileMeta{}, e
	}
	sum, e := sha(path)
	if e != nil {
		return fileMeta{}, e
	}
	return fileMeta{len(lines), addresses, sum}, nil
}

func readManifest(path string) (manifest, bool) {
	b, e := os.ReadFile(path)
	if e != nil {
		return manifest{}, false
	}
	var m manifest
	if json.Unmarshal(b, &m) != nil {
		return manifest{}, false
	}
	return m, true
}

func sameManifestContent(a, b manifest) bool {
	a.GeneratedAt = ""
	b.GeneratedAt = ""
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)
	return aErr == nil && bErr == nil && string(aJSON) == string(bJSON)
}

func writeManifest(path string, m manifest) {
	b, e := json.MarshalIndent(m, "", "  ")
	if e != nil {
		panic(e)
	}
	if e := os.WriteFile(path, append(b, '\n'), 0644); e != nil {
		panic(e)
	}
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

// bgpConflictHealingRanges treats the longest-prefix RIS routing decision as
// the formal healing unit. It returns all retained current three-operator BGP
// addresses and the subset whose complete origin unit has a same-operator
// APNIC parent. Existing strong exclusions may still punch holes afterwards.
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

func provinceSet() map[string]bool {
	out := map[string]bool{}
	for _, p := range provinces {
		out[p.Name] = true
	}
	return out
}

func main() {
	src := flag.String("sources", "", "source directory")
	out := flag.String("output", "data", "output directory")
	operatorConfig := flag.String("operator-config", "config/operators.json", "operator classification config")
	flag.Parse()
	if *src == "" {
		panic("--sources is required")
	}
	pipelineStarted, phaseStarted := time.Now(), time.Now()
	logPhase := func(name string) {
		now := time.Now()
		fmt.Printf("timing: %-28s %s\n", name, now.Sub(phaseStarted).Round(time.Millisecond))
		phaseStarted = now
	}

	oldManifest, hasOldManifest := readManifest(filepath.Join(*out, "manifest.json"))

	classifier, e := operatorconfig.Load(*operatorConfig, operators)
	if e != nil {
		panic(e)
	}
	ranges, includedASNRecords, excludedASNs, asnDescriptions, e := operatorRanges(filepath.Join(*src, "iptoasn_ipv4.tsv.gz"), classifier)
	if e != nil {
		panic(e)
	}
	chinaRanges, e := cidrs(filepath.Join(*src, "china.txt"))
	if e != nil {
		panic(e)
	}
	if addressCount(chinaRanges) < 100000000 {
		panic("origin-only China source contains fewer than 100,000,000 IPv4 addresses")
	}
	var originCandidates []span
	preCloudByOperator := map[string][]span{}
	for _, operator := range operators {
		originCandidates = append(originCandidates, ranges[operator]...)
		preCloudByOperator[operator] = intersect(ranges[operator], chinaRanges)
	}
	var preCloudCandidates []span
	for _, operator := range operators {
		preCloudCandidates = append(preCloudCandidates, preCloudByOperator[operator]...)
	}
	preCloudCandidates = merge(preCloudCandidates)

	var cloudRanges []span
	var cloudSourceSummaries []cloudSourceMeta
	var excludedPrefixes []prefixExclusionMeta
	for _, source := range cloudSources {
		sourceRanges, e := cidrs(filepath.Join(*src, source+".txt"))
		if e != nil {
			panic(e)
		}
		if len(sourceRanges) == 0 {
			panic(source + " contains no IPv4 CIDRs")
		}
		effective := intersect(preCloudCandidates, sourceRanges)
		cloudSourceSummaries = append(cloudSourceSummaries, cloudSourceMeta{
			Source: source, Provider: cloudProvider(source), SourceCIDRCount: len(spanCIDRs(sourceRanges)),
			SourceAddressCount: addressCount(sourceRanges), EffectiveCIDRCount: len(spanCIDRs(effective)),
			EffectiveAddressCount: addressCount(effective),
		})
		for _, record := range includedASNRecords {
			hits := intersect(intersect(record.ranges, chinaRanges), sourceRanges)
			for _, cidr := range spanCIDRs(hits) {
				prefix := netip.MustParsePrefix(cidr)
				excludedPrefixes = append(excludedPrefixes, prefixExclusionMeta{
					Source: source, Category: "cloud_provider_cidr", Provider: cloudProvider(source), CIDR: cidr,
					AddressCount: uint64(1) << uint(32-prefix.Bits()), ASN: record.meta.ASN,
					Operator: record.meta.Operator, ASNDescription: record.meta.Description,
					MatchedBy: "IP-Data provider CIDR", Reason: "Prefix is explicitly listed by IP-Data for " + cloudProvider(source),
				})
			}
		}
		cloudRanges = append(cloudRanges, sourceRanges...)
	}
	cloudRanges = merge(cloudRanges)
	postCloudCandidates := subtract(preCloudCandidates, cloudRanges)
	logPhase("origin and cloud inputs")

	orgNames, e := apnicorg.Parse(filepath.Join(*src, "apnic_organisation.gz"))
	if e != nil {
		panic(e)
	}
	apnicRecords, e := apnicinetnum.Parse(filepath.Join(*src, "apnic_inetnum.gz"))
	if e != nil {
		panic(e)
	}
	if len(apnicRecords) < 10000 {
		panic(fmt.Sprintf("APNIC inetnum source contains only %d records", len(apnicRecords)))
	}
	apnicRecordCount := len(apnicRecords)
	apnicRecords = relevantAPNICRecords(apnicRecords, postCloudCandidates)
	apnicinetnum.AttachOrganizationNames(apnicRecords, orgNames)
	autnumRecords, e := apnicautnum.Parse(filepath.Join(*src, "apnic_autnum.gz"))
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
	postCloudByASN := map[string][]span{}
	for asn, record := range includedASNRecords {
		postCloudByASN[asn] = subtract(intersect(record.ranges, chinaRanges), cloudRanges)
	}
	var apnicRanges, apnicPurposeRanges, portableHolderRanges, delegatedHolderRanges, legalEntityHolderRanges []span
	purposeSegments, portableSegments, delegatedSegments, legalEntityHolderSegments := 0, 0, 0, 0
	for _, segment := range apnicSegments {
		switch segment.Match.Category {
		case "apnic_portable_holder":
			portableSegments++
		case "apnic_delegated_holder":
			delegatedSegments++
		case "apnic_independent_legal_entity_holder":
			legalEntityHolderSegments++
		default:
			purposeSegments++
		}
		segmentRange := []span{{segment.Lo, segment.Hi}}
		effective := intersect(segmentRange, postCloudCandidates)
		if len(effective) == 0 {
			continue
		}
		apnicRanges = append(apnicRanges, effective...)
		switch segment.Match.Category {
		case "apnic_portable_holder":
			portableHolderRanges = append(portableHolderRanges, effective...)
		case "apnic_delegated_holder":
			delegatedHolderRanges = append(delegatedHolderRanges, effective...)
		case "apnic_independent_legal_entity_holder":
			legalEntityHolderRanges = append(legalEntityHolderRanges, effective...)
		default:
			apnicPurposeRanges = append(apnicPurposeRanges, effective...)
		}
		for asn, record := range includedASNRecords {
			hits := intersect(postCloudByASN[asn], effective)
			for _, cidr := range spanCIDRs(hits) {
				prefix := netip.MustParsePrefix(cidr)
				source := "apnic_inetnum"
				if segment.Match.Category == "apnic_portable_holder" || segment.Match.Category == "apnic_delegated_holder" || segment.Match.Category == "apnic_independent_legal_entity_holder" {
					source = "apnic_autnum"
				}
				linkIndex := autnumIndex
				if segment.Match.Category == "apnic_independent_legal_entity_holder" {
					linkIndex = registryAutnumIndex
				}
				excludedPrefixes = append(excludedPrefixes, prefixExclusionMeta{
					Source: source, Category: segment.Match.Category, CIDR: cidr,
					AddressCount: uint64(1) << uint(32-prefix.Bits()), ASN: asn,
					Operator: record.meta.Operator, ASNDescription: record.meta.Description,
					RegistryNetnames: segment.Record.Netnames, RegistryDescriptions: segment.Record.Descriptions,
					RegistryOrganizations: segment.Record.Organizations, RegistryMaintainers: segment.Record.Maintainers,
					RegistryOrganizationNames: segment.Record.OrganizationNames,
					RegistryStatus:            segment.Record.Status, RegistryLastModified: segment.Record.LastModified, MatchedBy: segment.Match.MatchedBy, Reason: segment.Match.Reason,
					LinkedASNs: independentAutnumLinks(segment.Record, linkIndex, classifier, asnDescriptions),
				})
			}
		}
	}
	apnicRanges = merge(apnicRanges)
	apnicPurposeRanges = merge(apnicPurposeRanges)
	portableHolderRanges = merge(portableHolderRanges)
	delegatedHolderRanges = merge(delegatedHolderRanges)
	legalEntityHolderRanges = merge(legalEntityHolderRanges)
	preRouteExcluded := merge(append(append([]span{}, cloudRanges...), apnicRanges...))
	logPhase("APNIC inetnum and aut-num")

	routeRecords, routeObjectCount, relevantRouteObjectCount, e := apnicroute.Parse(filepath.Join(*src, "apnic_route.gz"), orgNames, func(lo, hi uint32) bool { return overlapsSorted(postCloudCandidates, lo, hi) })
	if e != nil {
		panic(e)
	}
	routeSegments := apnicroute.Resolve(routeRecords)
	var routeRanges []span
	routeValidatedMatches := 0
	for _, segment := range routeSegments {
		for _, variant := range segment.Record.Variants {
			record := includedASNRecords[variant.Origin]
			if record == nil {
				continue
			}
			result := classifier.ClassifyAPNICInetnum(apnicroute.SearchText(variant))
			if !result.Excluded {
				continue
			}
			hits := intersect(subtract(intersect(record.ranges, chinaRanges), preRouteExcluded), []span{{segment.Lo, segment.Hi}})
			if len(hits) == 0 {
				continue
			}
			routeValidatedMatches++
			routeRanges = append(routeRanges, hits...)
			for _, cidr := range spanCIDRs(hits) {
				prefix := netip.MustParsePrefix(cidr)
				excludedPrefixes = append(excludedPrefixes, prefixExclusionMeta{Source: "apnic_route", Category: "apnic_route", CIDR: cidr, AddressCount: uint64(1) << uint(32-prefix.Bits()), ASN: variant.Origin, Operator: record.meta.Operator, ASNDescription: record.meta.Description, RegistryDescriptions: variant.Descriptions, RegistryOrganizations: variant.Organizations, RegistryOrganizationNames: variant.OrganizationNames, RegistryMaintainers: variant.Maintainers, RegistryLastModified: variant.LastModified, MatchedBy: result.MatchedBy + "; current BGP origin matches APNIC route origin", Reason: result.Reason})
			}
		}
	}
	routeRanges = merge(routeRanges)
	preRISExcluded := merge(append(append([]span{}, preRouteExcluded...), routeRanges...))
	routeOriginCandidates, routeOriginCandidateRanges := auditIndependentRouteOrigins(apnicAllSegments, routeSegments, subtract(preCloudCandidates, preRISExcluded), includedASNRecords, autnumIndex, classifier, asnDescriptions)
	for _, candidate := range routeOriginCandidates {
		excludedPrefixes = append(excludedPrefixes, routeOriginExclusionMeta(candidate))
	}
	preRISExcluded = merge(append(preRISExcluded, routeOriginCandidateRanges...))
	logPhase("APNIC route")

	risRecords, risStats, e := riswhois.Parse(filepath.Join(*src, "riswhois_ipv4.gz"), func(lo, hi uint32) bool { return overlapsSorted(postCloudCandidates, lo, hi) })
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
		for _, origin := range segment.Record.Origins {
			if origin.SeenPeers > maxPeers {
				maxPeers = origin.SeenPeers
			}
		}
		for asn, record := range includedASNRecords {
			currentSeen := 0
			for _, origin := range segment.Record.Origins {
				if origin.ASN == asn {
					currentSeen = origin.SeenPeers
				}
			}
			if currentSeen < 10 || currentSeen*20 < maxPeers {
				continue
			}
			hits := intersect(subtract(intersect(record.ranges, chinaRanges), preRISExcluded), []span{{segment.Lo, segment.Hi}})
			if len(hits) == 0 {
				continue
			}
			candidateMOAS++
			matchedBy, reason := "", ""
			observed := make([]observedOriginMeta, 0, len(segment.Record.Origins))
			for _, origin := range segment.Record.Origins {
				description := asnDescriptions[origin.ASN]
				observed = append(observed, observedOriginMeta{origin.ASN, description, origin.SeenPeers})
				if origin.ASN == asn || origin.SeenPeers < 10 || origin.SeenPeers*20 < maxPeers {
					continue
				}
				result := classifier.ClassifyAPNICInetnum(description)
				if result.Excluded && matchedBy == "" {
					matchedBy = result.MatchedBy + "; RIPE RIS multi-observer MOAS"
					reason = "Alternate origin AS" + origin.ASN + " is strongly identified as outside ordinary Internet user access scope: " + result.Reason
				}
			}
			if matchedBy == "" {
				continue
			}
			strongMOAS++
			risRanges = append(risRanges, hits...)
			for _, cidr := range spanCIDRs(hits) {
				prefix := netip.MustParsePrefix(cidr)
				excludedPrefixes = append(excludedPrefixes, prefixExclusionMeta{Source: "riswhois_ipv4", Category: "ris_moas", CIDR: cidr, AddressCount: uint64(1) << uint(32-prefix.Bits()), ASN: asn, Operator: record.meta.Operator, ASNDescription: record.meta.Description, MatchedBy: matchedBy, Reason: reason, ObservedOrigins: observed})
			}
		}
	}
	risRanges = merge(risRanges)
	excludedRanges := merge(append(append([]span{}, preRISExcluded...), risRanges...))
	logPhase("RIPE RISWhois")

	sourceRank := map[string]int{}
	for i, source := range cloudSources {
		sourceRank[source] = i
	}
	sourceRank["apnic_inetnum"] = len(cloudSources)
	sourceRank["apnic_autnum"] = len(cloudSources) + 1
	sourceRank["apnic_route"] = len(cloudSources) + 2
	sourceRank["riswhois_ipv4"] = len(cloudSources) + 3
	sort.Slice(excludedPrefixes, func(i, j int) bool {
		a, b := excludedPrefixes[i], excludedPrefixes[j]
		if sourceRank[a.Source] != sourceRank[b.Source] {
			return sourceRank[a.Source] < sourceRank[b.Source]
		}
		if a.Operator != b.Operator || a.ASN != b.ASN {
			return lessOperatorASN(a.Operator, a.ASN, b.Operator, b.ASN)
		}
		aAddress, bAddress := n(netip.MustParsePrefix(a.CIDR).Addr()), n(netip.MustParsePrefix(b.CIDR).Addr())
		if aAddress != bAddress {
			return aAddress < bAddress
		}
		return a.MatchedBy < b.MatchedBy
	})
	operatorAdmissionRanges := apnicOperatorAdmissionRanges(apnicRecords, classifier)
	operatorConflictRanges := apnicOperatorConflictRanges(apnicAllSegments, classifier)
	preAdmissionByOperator := map[string][]span{}
	hierarchicalByOperator := map[string][]span{}
	admissionDeniedByOperator := map[string][]span{}
	var preAdmissionRanges []span
	for _, o := range operators {
		preAdmissionByOperator[o] = subtract(preCloudByOperator[o], excludedRanges)
		parentAdmitted := intersect(preAdmissionByOperator[o], operatorAdmissionRanges[o])
		hierarchicalByOperator[o] = subtract(parentAdmitted, operatorConflictRanges[o])
		preAdmissionRanges = append(preAdmissionRanges, preAdmissionByOperator[o]...)
	}
	preAdmissionRanges = merge(preAdmissionRanges)
	asnOperators := map[string]string{}
	for asn, record := range includedASNRecords {
		asnOperators[asn] = record.meta.Operator
	}
	observedBGPRanges, conflictHealingEligibleRanges := bgpConflictHealingRanges(risSegments, asnOperators, preCloudByOperator, preAdmissionByOperator, operatorAdmissionRanges)
	var hierarchicalRanges []span
	for _, operator := range operators {
		hierarchicalRanges = append(hierarchicalRanges, hierarchicalByOperator[operator]...)
	}
	hierarchicalRanges = merge(hierarchicalRanges)
	ranges = conflictHealedAdmissionByOperator(hierarchicalByOperator, conflictHealingEligibleRanges, preAdmissionByOperator)
	var finalRanges, admissionDeniedRanges, assignedRanges []span
	for _, operator := range operators {
		if len(intersect(ranges[operator], merge(assignedRanges))) != 0 {
			panic("conflict-healed per-operator address sets overlap")
		}
		assignedRanges = append(assignedRanges, ranges[operator]...)
		finalRanges = append(finalRanges, ranges[operator]...)
		admissionDeniedByOperator[operator] = subtract(preAdmissionByOperator[operator], ranges[operator])
		admissionDeniedRanges = append(admissionDeniedRanges, admissionDeniedByOperator[operator]...)
	}
	finalRanges = merge(finalRanges)
	admissionDeniedRanges = merge(admissionDeniedRanges)
	bgpConflictHealedAdded := subtract(finalRanges, hierarchicalRanges)
	bgpConflictHealedRemoved := subtract(hierarchicalRanges, finalRanges)
	if len(bgpConflictHealedRemoved) != 0 {
		panic("BGP conflict healing removes addresses from the hierarchical baseline")
	}
	if len(subtract(bgpConflictHealedAdded, preAdmissionRanges)) != 0 {
		panic("BGP conflict healing adds addresses without a current three-operator origin or after a strong exclusion")
	}
	if len(subtract(bgpConflictHealedAdded, observedBGPRanges)) != 0 {
		panic("BGP conflict healing adds addresses without a current RIS-observed three-operator origin")
	}
	if len(subtract(bgpConflictHealedAdded, conflictHealingEligibleRanges)) != 0 {
		panic("BGP conflict healing adds addresses outside a same-operator APNIC parent")
	}
	var allOperatorConflicts []span
	for _, operator := range operators {
		allOperatorConflicts = append(allOperatorConflicts, operatorConflictRanges[operator]...)
	}
	if len(subtract(bgpConflictHealedAdded, merge(allOperatorConflicts))) != 0 {
		panic("BGP conflict healing adds addresses outside three-operator registration conflicts")
	}
	if len(intersect(finalRanges, excludedRanges)) != 0 {
		panic("BGP conflict healing overlaps an enforced strong exclusion")
	}
	includedASNs := includedASNList(includedASNRecords, chinaRanges, excludedRanges, ranges)
	preAdmissionCIDRCount := len(spanCIDRs(preAdmissionRanges))
	hierarchicalCIDRCount := len(spanCIDRs(hierarchicalRanges))
	finalCIDRCount := len(spanCIDRs(finalRanges))
	if finalCIDRCount > preAdmissionCIDRCount*2 {
		panic(fmt.Sprintf("operator parent-registration admission expands ACL from %d to %d CIDRs, exceeding the 2.0x limit", preAdmissionCIDRCount, finalCIDRCount))
	}
	admissionExpansionRatio := float64(finalCIDRCount) / float64(preAdmissionCIDRCount)
	conflictHealingCIDRRatio := float64(finalCIDRCount) / float64(hierarchicalCIDRCount)
	conflictHealedAddressRatio := float64(addressCount(bgpConflictHealedAdded)) / float64(addressCount(hierarchicalRanges))
	if conflictHealingCIDRRatio > maxConflictHealingCIDRRatio {
		panic(fmt.Sprintf("BGP conflict healing expands ACL from %d to %d CIDRs, exceeding the %.2fx limit", hierarchicalCIDRCount, finalCIDRCount, maxConflictHealingCIDRRatio))
	}
	if conflictHealedAddressRatio > maxConflictHealedAddressRatio {
		panic(fmt.Sprintf("BGP conflict healing adds %.6f of hierarchical addresses, exceeding the %.6f limit", conflictHealedAddressRatio, maxConflictHealedAddressRatio))
	}
	if addressCount(finalRanges) < 100000000 {
		panic("final output contains fewer than 100,000,000 IPv4 addresses")
	}
	by := map[string]map[string][]span{}
	for _, o := range operators {
		by[o] = map[string][]span{}
	}
	provinceSourceRanges := map[string][]span{}
	provinceNames := provinceSet()

	b, e := os.ReadFile(filepath.Join(*src, "ip2region_ipv4_source.txt"))
	if e != nil {
		panic(e)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		x := strings.Split(line, "|")
		if len(x) != 7 || x[2] != "中国" {
			continue
		}
		p := x[3]
		if a, ok := aliases[p]; ok {
			p = a
		}
		if !provinceNames[p] {
			continue
		}
		a, _ := netip.ParseAddr(x[0])
		z, _ := netip.ParseAddr(x[1])
		provinceSourceRanges[p] = append(provinceSourceRanges[p], span{n(a), n(z)})
	}
	for _, p := range provinces {
		provinceRanges := merge(provinceSourceRanges[p.Name])
		for _, o := range operators {
			by[o][p.Name] = intersect(ranges[o], provinceRanges)
		}
	}
	var provinceAttributed []span
	for _, p := range provinces {
		for _, operator := range operators {
			provinceAttributed = append(provinceAttributed, by[operator][p.Name]...)
		}
	}
	provinceAttributed = merge(provinceAttributed)
	if addressCount(provinceAttributed)*100 < addressCount(finalRanges)*90 {
		panic("ip2region attributes fewer than 90% of final output addresses to provinces")
	}
	logPhase("final and province ranges")

	if e := os.RemoveAll(*out); e != nil {
		panic(e)
	}
	if e := os.MkdirAll(*out, 0755); e != nil {
		panic(e)
	}

	m := manifest{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Scope:       "IPv4; mainland China; current China Telecom, China Mobile, or China Unicom BGP origin plus a covering same-operator APNIC registration; current three-operator BGP units heal internal operator-registration conflict holes after all strong exclusions; small ambiguous customer-use fragments are accepted as best-effort error to keep the ACL deployably compact; dedicated premium backbone ASNs, cloud-provider CIDRs, strong APNIC registrations outside ordinary user access scope, independent holders and route origins, and strong RIPE RIS MOAS evidence are excluded",
		OperatorAdmission: operatorAdmissionMeta{Mode: "bgp_registration_conflict_healing_with_strong_exclusions", PreCIDRCount: preAdmissionCIDRCount, DeniedCIDRCount: len(spanCIDRs(admissionDeniedRanges)), HierarchicalCIDRCount: hierarchicalCIDRCount, ConflictHealedCIDRCount: len(spanCIDRs(bgpConflictHealedAdded)), ConflictHealedAddressCount: addressCount(bgpConflictHealedAdded), FinalCIDRCount: finalCIDRCount, CIDRExpansionRatio: admissionExpansionRatio, MaximumCIDRExpansionRatio: maxAdmissionCIDRExpansionRatio, ConflictHealingCIDRRatio: conflictHealingCIDRRatio, MaximumConflictHealingCIDRRatio: maxConflictHealingCIDRRatio, ConflictHealedAddressRatio: conflictHealedAddressRatio, MaximumConflictHealedAddressRatio: maxConflictHealedAddressRatio},
		Stages: []stageMeta{
			stage("operator_origin_candidates", originCandidates),
			stage("china_origin_intersection", preCloudCandidates),
			stage("effective_cloud_prefix_exclusions", intersect(preCloudCandidates, cloudRanges)),
			stage("effective_apnic_prefix_exclusions", apnicPurposeRanges),
			stage("effective_apnic_portable_holder_exclusions", portableHolderRanges),
			stage("effective_apnic_delegated_holder_exclusions", delegatedHolderRanges),
			stage("effective_apnic_independent_legal_entity_exclusions", legalEntityHolderRanges),
			stage("effective_apnic_route_exclusions", routeRanges),
			stage("effective_apnic_independent_route_origin_exclusions", routeOriginCandidateRanges),
			stage("effective_ris_moas_exclusions", risRanges),
			stage("pre_operator_parent_registration_admission", preAdmissionRanges),
			stage("operator_parent_registration_admissions", hierarchicalRanges),
			stage("bgp_conflict_healed_additions", bgpConflictHealedAdded),
			stage("bgp_conflict_healing_admissions", finalRanges),
			stage("operator_registration_admission_denials", admissionDeniedRanges),
			stage("final_output", finalRanges),
			stage("province_attributed_output", provinceAttributed),
		},
		CloudSources: cloudSourceSummaries,
		APNICInetnum: apnicSourceMeta{
			RecordCount: apnicRecordCount, RelevantRecordCount: len(apnicRecords), MatchedWinningSegmentCount: purposeSegments,
			EffectiveCIDRCount: len(spanCIDRs(apnicPurposeRanges)), EffectiveAddressCount: addressCount(apnicPurposeRanges),
		},
		APNICPortableHolders:          portableHolderMeta{AutnumRecordCount: len(autnumRecords), MatchedWinningSegmentCount: portableSegments, EffectiveCIDRCount: len(spanCIDRs(portableHolderRanges)), EffectiveAddressCount: addressCount(portableHolderRanges)},
		APNICDelegatedHolders:         delegatedHolderMeta{MatchedWinningSegmentCount: delegatedSegments, EffectiveCIDRCount: len(spanCIDRs(delegatedHolderRanges)), EffectiveAddressCount: addressCount(delegatedHolderRanges)},
		APNICIndependentLegalEntities: delegatedHolderMeta{MatchedWinningSegmentCount: legalEntityHolderSegments, EffectiveCIDRCount: len(spanCIDRs(legalEntityHolderRanges)), EffectiveAddressCount: addressCount(legalEntityHolderRanges)},
		APNICRoute:                    routeSourceMeta{ObjectCount: routeObjectCount, RelevantObjectCount: relevantRouteObjectCount, RelevantWinningSegmentCount: len(routeSegments), OriginValidatedMatchCount: routeValidatedMatches, EffectiveCIDRCount: len(spanCIDRs(routeRanges)), EffectiveAddressCount: addressCount(routeRanges)},
		APNICRouteOriginAudit:         routeOriginAuditMeta{Enforced: true, CandidateEvidenceCount: len(routeOriginCandidates), CandidateCIDRCount: len(spanCIDRs(routeOriginCandidateRanges)), CandidateAddressCount: addressCount(routeOriginCandidateRanges), Candidates: routeOriginCandidates},
		RISWhois:                      risSourceMeta{RowCount: risStats.Rows, PrefixCount: risStats.Prefixes, RelevantPrefixCount: risStats.RelevantPrefixes, WinningSegmentCount: len(risSegments), CandidateMOASSegmentCount: candidateMOAS, StrongEvidenceSegmentCount: strongMOAS, RetainedAmbiguousMOASSegmentCount: candidateMOAS - strongMOAS, EffectiveCIDRCount: len(spanCIDRs(risRanges)), EffectiveAddressCount: addressCount(risRanges)},
		ExcludedPrefixes:              excludedPrefixes,
		OperatorSummary:               summarizeOperators(includedASNs), IncludedASNs: includedASNs, ExcludedASNs: excludedASNs,
	}
	configSource, e := source(*operatorConfig, "operator_config", "", filepath.ToSlash(*operatorConfig))
	if e != nil {
		panic(e)
	}
	m.Sources = append(m.Sources, configSource)
	for _, o := range append([]string{"china", "iptoasn_ipv4", "apnic_organisation", "apnic_inetnum", "apnic_autnum", "apnic_route", "riswhois_ipv4", "ip2region_ipv4_source"}, cloudSources...) {
		path := sourcePath(*src, o)
		sourceEntry, e := source(path, o, urls[o], "")
		if e != nil {
			panic(e)
		}
		m.Sources = append(m.Sources, sourceEntry)
	}

	for _, o := range operators {
		path := filepath.Join("operators", o+".txt")
		meta, e := write(filepath.Join(*out, path), ranges[o])
		if e != nil {
			panic(e)
		}
		m.Lists = append(m.Lists, listMeta{Name: o, Path: filepath.ToSlash(path), fileMeta: meta})
	}
	cnMeta, e := write(filepath.Join(*out, "cn.txt"), finalRanges)
	if e != nil {
		panic(e)
	}
	var retainedByASN uint64
	for _, entry := range includedASNs {
		retainedByASN += entry.RetainedAddressCount
	}
	if retainedByASN != cnMeta.AddressCount {
		panic(fmt.Sprintf("included ASN address total %d does not match cn.txt address total %d", retainedByASN, cnMeta.AddressCount))
	}
	m.Lists = append(m.Lists, listMeta{Name: "CN", Path: "cn.txt", fileMeta: cnMeta})

	for _, p := range provinces {
		var rows []span
		for _, o := range operators {
			rows = append(rows, by[o][p.Name]...)
		}
		path := filepath.Join("provinces", p.Slug+".txt")
		meta, e := write(filepath.Join(*out, path), rows)
		if e != nil {
			panic(e)
		}
		m.Lists = append(m.Lists, listMeta{Name: p.Name, Path: filepath.ToSlash(path), fileMeta: meta})
	}

	if hasOldManifest && sameManifestContent(oldManifest, m) {
		m.GeneratedAt = oldManifest.GeneratedAt
	}
	writeManifest(filepath.Join(*out, "manifest.json"), m)
	logPhase("write outputs and manifest")
	fmt.Printf("timing: %-28s %s\n", "total", time.Since(pipelineStarted).Round(time.Millisecond))
}
