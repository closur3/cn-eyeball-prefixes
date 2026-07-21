package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/closur3/cn-eyeball-prefixes/internal/apnic6"
	"github.com/closur3/cn-eyeball-prefixes/internal/apnicorg"
	"github.com/closur3/cn-eyeball-prefixes/internal/ipset6"
	"github.com/closur3/cn-eyeball-prefixes/internal/operatorconfig"
)

var operators = []string{"chinanet", "cmcc", "unicom"}

type originRecord struct {
	Range       ipset6.Range
	ASN         string
	Country     string
	Description string
	Operator    string
}

type spaceStat struct {
	CIDRCount         int    `json:"cidr_count"`
	AddressCount      string `json:"address_count"`
	Slash64Equivalent string `json:"slash64_equivalent"`
	PercentOfOperator string `json:"percent_of_operator,omitempty"`
}

type categoryStat struct {
	Operator string    `json:"operator"`
	Category string    `json:"category"`
	Space    spaceStat `json:"space"`
}

type registrationFact struct {
	Operator          string   `json:"operator"`
	Category          string   `json:"category"`
	RegistryPrefix    string   `json:"registry_prefix"`
	RegistryPrefixLen int      `json:"registry_prefix_length"`
	Netnames          []string `json:"netnames,omitempty"`
	Descriptions      []string `json:"descriptions,omitempty"`
	Organizations     []string `json:"organizations,omitempty"`
	OrganizationNames []string `json:"organization_names,omitempty"`
	Maintainers       []string `json:"maintainers,omitempty"`
	Country           string   `json:"country,omitempty"`
	Status            string   `json:"status,omitempty"`
	MatchedBy         string   `json:"matched_by,omitempty"`
	Reason            string   `json:"reason,omitempty"`
	AdmissionDecision string   `json:"admission_decision"`
	AdmissionReason   string   `json:"admission_reason"`
	ExplicitUserPurpose bool   `json:"explicit_user_purpose,omitempty"`
	Space             spaceStat `json:"space"`
	count             *big.Int
	ranges            []ipset6.Range
}

type operatorSummary struct {
	Operator                    string    `json:"operator"`
	Candidate                   spaceStat `json:"candidate"`
	MostSpecificRegistryCovered spaceStat `json:"most_specific_registry_covered"`
	RegistryUncovered           spaceStat `json:"registry_uncovered"`
	SameOriginRoute6Covered     spaceStat `json:"same_origin_route6_covered"`
	StrongRoutePurposeSignal    spaceStat `json:"strong_route_purpose_signal"`
	AllowlistAdmitted           spaceStat `json:"allowlist_admitted"`
	NotAdmitted                 spaceStat `json:"not_admitted"`
	ExplicitUserPurpose         spaceStat `json:"explicit_user_purpose"`
}

type report struct {
	GeneratedAt             string             `json:"generated_at"`
	Scope                   string             `json:"scope"`
	Inet6numRecordCount     int                `json:"inet6num_record_count"`
	Inet6numSegmentCount    int                `json:"inet6num_segment_count"`
	Route6RecordCount       int                `json:"route6_record_count"`
	Operators               []operatorSummary  `json:"operators"`
	Categories              []categoryStat     `json:"categories"`
	RegistrationFacts       []registrationFact `json:"registration_facts"`
	PrefixLengthByCategory  map[string]map[int]int `json:"winning_prefix_length_by_category"`
}

type ipv6Manifest struct {
	GeneratedAt string                    `json:"generated_at"`
	Scope       string                    `json:"scope"`
	Status      map[string]string         `json:"status"`
	Operators   map[string]operatorOutput `json:"operators"`
	Evidence    []registrationFact        `json:"admission_evidence"`
}

type operatorOutput struct {
	Path  string    `json:"path,omitempty"`
	Space spaceStat `json:"space"`
}

func main() {
	china6Path := flag.String("china6", "", "gaoyifan china6.txt")
	iptoasnPath := flag.String("iptoasn", "", "IPtoASN IPv6 TSV gzip")
	inet6numPath := flag.String("inet6num", "", "APNIC inet6num gzip")
	route6Path := flag.String("route6", "", "APNIC route6 gzip")
	organisationPath := flag.String("organisation", "", "APNIC organisation gzip")
	configPath := flag.String("operator-config", "config/operators.json", "operator config")
	jsonPath := flag.String("json", "reports/ipv6/three-operator-registration.json", "JSON report path")
	markdownPath := flag.String("markdown", "reports/ipv6/three-operator-registration.md", "Markdown report path")
	chinanetOutputPath := flag.String("chinanet-output", "data/ipv6/operators/chinanet.txt", "formal China Telecom IPv6 allowlist path")
	manifestPath := flag.String("manifest", "data/ipv6/manifest.json", "IPv6 manifest path")
	flag.Parse()
	for name, path := range map[string]string{"china6": *china6Path, "iptoasn": *iptoasnPath, "inet6num": *inet6numPath, "route6": *route6Path, "organisation": *organisationPath} {
		if path == "" {
			panic("--" + name + " is required")
		}
	}

	classifier, err := operatorconfig.Load(*configPath, operators)
	must(err)
	china6 := readCIDRs(*china6Path)
	origins := readOrigins(*iptoasnPath, classifier, china6)
	orgNames, err := apnicorg.Parse(*organisationPath)
	must(err)
	inetRecords, err := apnic6.ParseInet6num(*inet6numPath, orgNames)
	must(err)
	segments := apnic6.ResolveMostSpecific(inetRecords)
	routeRecords, err := apnic6.ParseRoute6(*route6Path, orgNames)
	must(err)

	candidateByOperator := map[string][]ipset6.Range{}
	candidateByASN := map[string][]ipset6.Range{}
	operatorByASN := map[string]string{}
	for _, origin := range origins {
		candidateByOperator[origin.Operator] = append(candidateByOperator[origin.Operator], origin.Range)
		candidateByASN[origin.ASN] = append(candidateByASN[origin.ASN], origin.Range)
		operatorByASN[origin.ASN] = origin.Operator
	}
	for operator := range candidateByOperator {
		candidateByOperator[operator] = ipset6.Merge(candidateByOperator[operator])
	}
	for asn := range candidateByASN {
		candidateByASN[asn] = ipset6.Merge(candidateByASN[asn])
	}

	categoryRanges := map[string]map[string][]ipset6.Range{}
	coveredByOperator := map[string][]ipset6.Range{}
	firstPassExcluded := map[string][]ipset6.Range{}
	explicitUserPurpose := map[string][]ipset6.Range{}
	facts := map[string]*registrationFact{}
	prefixLengths := map[string]map[int]int{}
	for _, origin := range origins {
		for _, hit := range intersectSegments(segments, origin.Range) {
			coveredByOperator[origin.Operator] = append(coveredByOperator[origin.Operator], hit.Range)
			category, matchedBy, reason := classifyRegistration(hit.Record, origin.Operator, classifier)
			excluded, exclusionReason := firstPassDecision(category, hit.Record)
			explicitUser := isExplicitUserPurpose(hit.Record)
			decision := "not_admitted"
			admissionReason := "Not admitted: no explicit end-user access purpose evidence"
			if excluded {
				decision = "exclude"
				admissionReason = exclusionReason
			} else if explicitUser {
				decision = "admit"
				admissionReason = "Admit: most-specific APNIC registration explicitly identifies broadband, mobile, or user-address access"
			}
			if categoryRanges[origin.Operator] == nil {
				categoryRanges[origin.Operator] = map[string][]ipset6.Range{}
			}
			categoryRanges[origin.Operator][category] = append(categoryRanges[origin.Operator][category], hit.Range)
			if excluded {
				firstPassExcluded[origin.Operator] = append(firstPassExcluded[origin.Operator], hit.Range)
			}
			if explicitUser {
				explicitUserPurpose[origin.Operator] = append(explicitUserPurpose[origin.Operator], hit.Range)
			}
			lengthKey := origin.Operator + "/" + category
			if prefixLengths[lengthKey] == nil {
				prefixLengths[lengthKey] = map[int]int{}
			}
			prefixLengths[lengthKey][hit.Record.Prefix.Bits()]++
			key := origin.Operator + "\x00" + category + "\x00" + hit.Record.Prefix.String()
			fact := facts[key]
			if fact == nil {
				fact = &registrationFact{
					Operator: origin.Operator, Category: category, RegistryPrefix: hit.Record.Prefix.String(), RegistryPrefixLen: hit.Record.Prefix.Bits(),
					Netnames: hit.Record.Netnames, Descriptions: hit.Record.Descriptions, Organizations: hit.Record.Organizations,
					OrganizationNames: hit.Record.OrganizationNames, Maintainers: hit.Record.Maintainers, Country: hit.Record.Country,
					Status: hit.Record.Status, MatchedBy: matchedBy, Reason: reason,
					AdmissionDecision: decision,
					AdmissionReason: admissionReason, ExplicitUserPurpose: explicitUser, count: new(big.Int),
				}
				facts[key] = fact
			}
			fact.count.Add(fact.count, ipset6.AddressCount([]ipset6.Range{hit.Range}))
			fact.ranges = append(fact.ranges, hit.Range)
		}
	}

	routeCovered := map[string][]ipset6.Range{}
	routeStrong := map[string][]ipset6.Range{}
	for _, record := range routeRecords {
		for _, variant := range record.Variants {
			operator := operatorByASN[variant.Origin]
			if operator == "" {
				continue
			}
			hits := ipset6.Intersect(candidateByASN[variant.Origin], []ipset6.Range{record.Range})
			if len(hits) == 0 {
				continue
			}
			routeCovered[operator] = append(routeCovered[operator], hits...)
			if result := classifier.ClassifyAPNICInetnum(apnic6.RouteSearchText(variant)); result.Excluded {
				routeStrong[operator] = append(routeStrong[operator], hits...)
			}
		}
	}

	generatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	result := report{
		GeneratedAt:            generatedAt,
		Scope:                  "Whitelist admission audit of the most-specific APNIC inet6num evidence covering current mainland China Telecom, China Mobile, and China Unicom IPv6 origins inside gaoyifan china6.txt; only China Telecom currently has sufficient explicit end-user purpose evidence for a formal output",
		Inet6numRecordCount:    len(inetRecords),
		Inet6numSegmentCount:   len(segments),
		Route6RecordCount:      len(routeRecords),
		PrefixLengthByCategory: prefixLengths,
	}
	admittedByOperator := map[string][]ipset6.Range{}
	for _, operator := range operators {
		candidate := candidateByOperator[operator]
		total := ipset6.AddressCount(candidate)
		covered := ipset6.Intersect(candidate, coveredByOperator[operator])
		uncovered := ipset6.Subtract(candidate, covered)
		hardExcluded := ipset6.Intersect(candidate, append(firstPassExcluded[operator], uncovered...))
		admitted := ipset6.Subtract(ipset6.Intersect(candidate, explicitUserPurpose[operator]), hardExcluded)
		admittedByOperator[operator] = admitted
		result.Operators = append(result.Operators, operatorSummary{
			Operator: operator, Candidate: makeStat(candidate, total), MostSpecificRegistryCovered: makeStat(covered, total),
			RegistryUncovered: makeStat(uncovered, total),
			SameOriginRoute6Covered: makeStat(ipset6.Intersect(candidate, routeCovered[operator]), total),
			StrongRoutePurposeSignal: makeStat(ipset6.Intersect(candidate, routeStrong[operator]), total),
			AllowlistAdmitted: makeStat(admitted, total),
			NotAdmitted: makeStat(ipset6.Subtract(candidate, admitted), total),
			ExplicitUserPurpose: makeStat(ipset6.Intersect(candidate, explicitUserPurpose[operator]), total),
		})
		for _, category := range []string{"same_operator", "strong_non_public", "other_operator", "independent_legal_entity", "other_or_unclassified"} {
			result.Categories = append(result.Categories, categoryStat{Operator: operator, Category: category, Space: makeStat(categoryRanges[operator][category], total)})
		}
	}
	for _, fact := range facts {
		total := ipset6.AddressCount(candidateByOperator[fact.Operator])
		fact.Space = makeStat(fact.ranges, total)
		result.RegistrationFacts = append(result.RegistrationFacts, *fact)
	}
	sort.Slice(result.RegistrationFacts, func(i, j int) bool {
		if c := result.RegistrationFacts[i].count.Cmp(result.RegistrationFacts[j].count); c != 0 {
			return c > 0
		}
		return result.RegistrationFacts[i].RegistryPrefix < result.RegistrationFacts[j].RegistryPrefix
	})
	formalChinanet := admittedByOperator["chinanet"]
	must(writePrefixFile(*chinanetOutputPath, formalChinanet))
	manifest := ipv6Manifest{
		GeneratedAt: generatedAt,
		Scope:       "IPv6 end-user access prefix allowlist; only prefixes with explicit positive purpose evidence are admitted",
		Status: map[string]string{
			"chinanet": "ready",
			"cmcc":     "pending_additional_positive_evidence",
			"unicom":   "pending_additional_positive_evidence",
		},
		Operators: map[string]operatorOutput{},
	}
	for _, operator := range operators {
		output := operatorOutput{Space: makeStat(admittedByOperator[operator], ipset6.AddressCount(candidateByOperator[operator]))}
		if operator == "chinanet" {
			output.Path = filepath.ToSlash(*chinanetOutputPath)
		}
		manifest.Operators[operator] = output
	}
	for _, fact := range result.RegistrationFacts {
		if fact.Operator == "chinanet" && fact.AdmissionDecision == "admit" {
			manifest.Evidence = append(manifest.Evidence, fact)
		}
	}

	must(writeJSON(*jsonPath, result))
	must(writeFile(*markdownPath, []byte(renderMarkdown(result))))
	must(writeJSON(*manifestPath, manifest))
}

func classifyRegistration(record apnic6.InetRecord, operator string, classifier *operatorconfig.Classifier) (string, string, string) {
	text := apnic6.InetSearchText(record)
	lower := strings.ToLower(text)
	for _, rule := range []struct{ needle, label, reason string }{
		{"ct-ipv6-volte-address", "ipv6_explicit_service: volte", "APNIC inet6num explicitly identifies a VoLTE service address pool rather than ordinary Internet access"},
		{"ipv6 address for volte", "ipv6_explicit_service: volte", "APNIC inet6num explicitly identifies a VoLTE service address pool rather than ordinary Internet access"},
		{"ct-ipv6-platform-address", "ipv6_explicit_service: own platform", "APNIC inet6num explicitly identifies the operator's own platform address pool"},
		{"ipv6 address for own platform", "ipv6_explicit_service: own platform", "APNIC inet6num explicitly identifies the operator's own platform address pool"},
		{"ct-ipv6-network-address", "ipv6_explicit_service: network", "APNIC inet6num explicitly identifies a generic network address pool without end-user purpose"},
		{"ipv6 address for network", "ipv6_explicit_service: network", "APNIC inet6num explicitly identifies a generic network address pool without end-user purpose"},
	} {
		if strings.Contains(lower, rule.needle) {
			return "strong_non_public", rule.label, rule.reason
		}
	}
	if result := classifier.ClassifyAPNICInetnum(text); result.Excluded {
		return "strong_non_public", result.MatchedBy, result.Reason
	}
	if result := classifier.ClassifyAPNICRegistrant(text); result.Operator != "" {
		if result.Operator == operator {
			return "same_operator", result.MatchedBy, "Most-specific APNIC registration is attributed to the current BGP operator"
		}
		return "other_operator", result.MatchedBy, "Most-specific APNIC registration is attributed to another Chinese operator"
	}
	if classifier.IsIndependentLegalEntity(apnic6.InetRegistrantText(record)) {
		return "independent_legal_entity", "independent_legal_entity_patterns", "Most-specific APNIC registration names an independent legal entity without operator attribution"
	}
	return "other_or_unclassified", "", "Most-specific APNIC registration has no current strong operator or non-public classification"
}

func firstPassDecision(category string, record apnic6.InetRecord) (bool, string) {
	if record.Country != "" && !strings.EqualFold(record.Country, "CN") {
		return true, "Exclude: most-specific APNIC registration is outside mainland China"
	}
	switch category {
	case "strong_non_public":
		return true, "Exclude: APNIC registration contains an explicit non-user-side purpose"
	case "other_operator":
		return true, "Exclude: most-specific APNIC registration belongs to another operator"
	case "independent_legal_entity":
		return true, "Exclude: most-specific APNIC registration names an independent legal entity"
	case "other_or_unclassified":
		if isPortableStatus(record.Status) {
			return true, "Exclude: portable resource cannot be attributed to the current operator"
		}
		return false, "Retain: non-portable assignment under the current three-operator Origin"
	default:
		if strings.EqualFold(strings.TrimSpace(record.Status), "ALLOCATED PORTABLE") {
			return false, "Retain: operator allocation parent; no more-specific APNIC purpose registration covers this candidate space"
		}
		return false, "Retain: most-specific APNIC registration is attributed to the current operator"
	}
}

func isPortableStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "ALLOCATED PORTABLE", "ASSIGNED PORTABLE":
		return true
	default:
		return false
	}
}

func isExplicitUserPurpose(record apnic6.InetRecord) bool {
	text := strings.ToLower(apnic6.InetSearchText(record))
	for _, phrase := range []string{"fixed broadband", "user address", "ipv6 address for mobile"} {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	compact := strings.NewReplacer("-", "", "_", "", " ", "").Replace(text)
	return strings.Contains(compact, "useraddress")
}

func readCIDRs(path string) []ipset6.Range {
	f, err := os.Open(path)
	must(err)
	defer f.Close()
	var rows []ipset6.Range
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if line == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(line)
		if err != nil || !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
			panic("invalid IPv6 prefix: " + line)
		}
		row, err := ipset6.FromPrefix(prefix)
		must(err)
		rows = append(rows, row)
	}
	must(scanner.Err())
	return ipset6.Merge(rows)
}

func readOrigins(path string, classifier *operatorconfig.Classifier, china6 []ipset6.Range) []originRecord {
	f, err := os.Open(path)
	must(err)
	defer f.Close()
	z, err := gzip.NewReader(f)
	must(err)
	defer z.Close()
	var out []originRecord
	scanner := bufio.NewScanner(z)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 5 || fields[2] == "0" {
			continue
		}
		lo, loErr := netip.ParseAddr(fields[0])
		hi, hiErr := netip.ParseAddr(fields[1])
		if loErr != nil || hiErr != nil || !lo.Is6() || !hi.Is6() || lo.Is4In6() || hi.Is4In6() {
			continue
		}
		result := classifier.Classify(fields[2], fields[4])
		if result.Operator == "" || result.Excluded {
			continue
		}
		for _, hit := range ipset6.Intersect([]ipset6.Range{{Lo: lo, Hi: hi}}, china6) {
			out = append(out, originRecord{Range: hit, ASN: fields[2], Country: fields[3], Description: fields[4], Operator: result.Operator})
		}
	}
	must(scanner.Err())
	sort.Slice(out, func(i, j int) bool { return out[i].Range.Lo.Compare(out[j].Range.Lo) < 0 })
	return out
}

func intersectSegments(segments []apnic6.Segment, target ipset6.Range) []apnic6.Segment {
	i := sort.Search(len(segments), func(i int) bool { return segments[i].Range.Hi.Compare(target.Lo) >= 0 })
	var out []apnic6.Segment
	for ; i < len(segments) && segments[i].Range.Lo.Compare(target.Hi) <= 0; i++ {
		lo, hi := segments[i].Range.Lo, segments[i].Range.Hi
		if lo.Compare(target.Lo) < 0 {
			lo = target.Lo
		}
		if hi.Compare(target.Hi) > 0 {
			hi = target.Hi
		}
		out = append(out, apnic6.Segment{Range: ipset6.Range{Lo: lo, Hi: hi}, Record: segments[i].Record})
	}
	return out
}

func makeStat(rows []ipset6.Range, total *big.Int) spaceStat {
	rows = ipset6.Merge(rows)
	count := ipset6.AddressCount(rows)
	stat := countStat(count, total)
	stat.CIDRCount = len(ipset6.Prefixes(rows))
	return stat
}

func countStat(count, total *big.Int) spaceStat {
	stat := spaceStat{AddressCount: count.String(), Slash64Equivalent: slash64(count)}
	if total.Sign() > 0 {
		stat.PercentOfOperator = percent(count, total)
	}
	return stat
}

func slash64(count *big.Int) string {
	return new(big.Rat).SetFrac(new(big.Int).Set(count), new(big.Int).Lsh(big.NewInt(1), 64)).FloatString(4)
}

func percent(count, total *big.Int) string {
	if total.Sign() == 0 {
		return "0.000000%"
	}
	return new(big.Rat).SetFrac(new(big.Int).Mul(new(big.Int).Set(count), big.NewInt(100)), total).FloatString(6) + "%"
}

func writeJSON(path string, value any) error {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, append(bytes, '\n'))
}

func writeFile(path string, bytes []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

func writePrefixFile(path string, rows []ipset6.Range) error {
	var b strings.Builder
	for _, prefix := range ipset6.Prefixes(rows) {
		fmt.Fprintln(&b, prefix)
	}
	return writeFile(path, []byte(b.String()))
}

func renderMarkdown(r report) string {
	var b strings.Builder
	fmt.Fprintln(&b, "# 三网 IPv6 APNIC 登记颗粒度审计")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "生成时间：`%s`\n\n", r.GeneratedAt)
	fmt.Fprintln(&b, "审计对象是 `当前三网 IPv6 Origin ∩ china6`。三网统一采用白名单准入：只有明确终端用户接入正证据才可进入正式结果。当前只生成证据充分的中国电信名单；移动、联通保持待补充状态。")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "APNIC `inet6num` 记录：**%d**；解析后最具体区间：**%d**；`route6` 前缀：**%d**。\n", r.Inet6numRecordCount, r.Inet6numSegmentCount, r.Route6RecordCount)

	fmt.Fprintln(&b, "\n## 运营商覆盖")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 运营商 | 候选 CIDR | inet6num 覆盖 | 未覆盖 | 同 Origin route6 覆盖 | route6 强非目标信号 |")
	fmt.Fprintln(&b, "| --- | ---: | ---: | ---: | ---: | ---: |")
	for _, row := range r.Operators {
		fmt.Fprintf(&b, "| %s | %d | %s | %s | %s | %s |\n", row.Operator, row.Candidate.CIDRCount, row.MostSpecificRegistryCovered.PercentOfOperator, row.RegistryUncovered.PercentOfOperator, row.SameOriginRoute6Covered.PercentOfOperator, row.StrongRoutePurposeSignal.PercentOfOperator)
	}

	fmt.Fprintln(&b, "\n## 白名单准入结果")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "准入要求最具体 APNIC 登记明确写明 `fixed broadband`、`mobile` 或 `UserAddress`，并同时位于当前同运营商 Origin 与 `china6`。运营商总分配、`for customer`、普通 ISP 描述和历史使用案例均不构成准入证据。")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 运营商 | 准入 CIDR | 准入空间 | 未准入 CIDR | 未准入空间 | 状态 |")
	fmt.Fprintln(&b, "| --- | ---: | ---: | ---: | ---: | --- |")
	for _, row := range r.Operators {
		status := "待补充正向证据"
		if row.Operator == "chinanet" {
			status = "已生成正式名单"
		}
		fmt.Fprintf(&b, "| %s | %d | %s | %d | %s | %s |\n", row.Operator, row.AllowlistAdmitted.CIDRCount, row.AllowlistAdmitted.PercentOfOperator, row.NotAdmitted.CIDRCount, row.NotAdmitted.PercentOfOperator, status)
	}

	renderFactSamples(&b, "白名单准入证据", r.RegistrationFacts, 100, func(fact registrationFact) bool {
		return fact.AdmissionDecision == "admit"
	})
	fmt.Fprintln(&b, "\n> `ALLOCATED PORTABLE` 的运营商总分配对象只证明地址资源归属，不提供终端用户或业务用途证据。候选地址回落到这类父级对象，表示 APNIC 没有覆盖它的更具体用途登记，不能把父级 Description 解读为该地址的实际用途。")
	renderFactSamples(&b, "缺少正向证据而未准入的样本", r.RegistrationFacts, 30, func(fact registrationFact) bool {
		return fact.AdmissionDecision == "not_admitted"
	})
	renderFactSamples(&b, "明确排除样本", r.RegistrationFacts, 30, func(fact registrationFact) bool {
		return fact.AdmissionDecision == "exclude"
	})

	fmt.Fprintln(&b, "\n## 最具体 inet6num 分类")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 运营商 | 分类 | CIDR | /64 等价数 | 占运营商候选 |")
	fmt.Fprintln(&b, "| --- | --- | ---: | ---: | ---: |")
	for _, row := range r.Categories {
		fmt.Fprintf(&b, "| %s | %s | %d | %s | %s |\n", row.Operator, row.Category, row.Space.CIDRCount, row.Space.Slash64Equivalent, row.Space.PercentOfOperator)
	}

	for _, category := range []string{"strong_non_public", "independent_legal_entity", "other_operator", "other_or_unclassified", "same_operator"} {
		fmt.Fprintf(&b, "\n## %s：地址量前 100 项\n\n", category)
		fmt.Fprintln(&b, "| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 依据 |")
		fmt.Fprintln(&b, "| --- | --- | ---: | --- | --- | --- |")
		shown := 0
		for _, fact := range r.RegistrationFacts {
			if fact.Category != category || shown >= 100 {
				continue
			}
			labelParts := append([]string{}, fact.Netnames...)
			labelParts = append(labelParts, fact.Descriptions...)
			labelParts = append(labelParts, fact.OrganizationNames...)
			label := strings.ReplaceAll(strings.Join(labelParts, "; "), "|", "\\|")
			evidence := fact.Reason
			if fact.MatchedBy != "" {
				evidence += " (`" + strings.ReplaceAll(fact.MatchedBy, "`", "'") + "`)"
			}
			fmt.Fprintf(&b, "| %s | `%s` | %s | %s | %s | %s |\n", fact.Operator, fact.RegistryPrefix, fact.Space.PercentOfOperator, label, fact.Status, evidence)
			shown++
		}
		if shown == 0 {
			fmt.Fprintln(&b, "| — | — | — | 无 | — | — |")
		}
	}
	return b.String()
}

func renderFactSamples(b *strings.Builder, title string, facts []registrationFact, limit int, include func(registrationFact) bool) {
	fmt.Fprintf(b, "\n## %s\n\n", title)
	fmt.Fprintln(b, "| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 白名单处理依据 |")
	fmt.Fprintln(b, "| --- | --- | ---: | --- | --- | --- |")
	shown := 0
	for _, fact := range facts {
		if !include(fact) || shown >= limit {
			continue
		}
		labelParts := append([]string{}, fact.Netnames...)
		labelParts = append(labelParts, fact.Descriptions...)
		labelParts = append(labelParts, fact.OrganizationNames...)
		label := strings.ReplaceAll(strings.Join(labelParts, "; "), "|", "\\|")
		fmt.Fprintf(b, "| %s | `%s` | %s | %s | %s | %s |\n", fact.Operator, fact.RegistryPrefix, fact.Space.PercentOfOperator, label, fact.Status, fact.AdmissionReason)
		shown++
	}
	if shown == 0 {
		fmt.Fprintln(b, "| — | — | — | 无 | — | — |")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
