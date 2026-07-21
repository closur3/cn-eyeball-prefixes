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
	"math/big"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/closur3/cn-eyeball-prefixes/internal/ipset6"
	"github.com/closur3/cn-eyeball-prefixes/internal/operatorconfig"
	"github.com/closur3/cn-eyeball-prefixes/internal/riswhois6"
)

var operators = []string{"chinanet", "cmcc", "unicom"}

type asMeta struct {
	Country     string `json:"country,omitempty"`
	Description string `json:"description,omitempty"`
}

type allocationConfig struct {
	Operators map[string][]allocation `json:"operators"`
}

type allocation struct {
	Prefix       string `json:"prefix"`
	Registry     string `json:"registry"`
	Netname      string `json:"netname"`
	Organization string `json:"organization"`
	Status       string `json:"status"`
	Source       string `json:"source"`
	parsed       netip.Prefix
}

type sourceMeta struct {
	Source string `json:"source"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type asnStat struct {
	ASN         string `json:"asn"`
	Description string `json:"description"`
	PrefixCount int    `json:"prefix_count"`
}

type operatorOutput struct {
	Status            string    `json:"status"`
	Path              string    `json:"path,omitempty"`
	PrefixCount       int       `json:"prefix_count"`
	Slash64Equivalent string    `json:"unique_slash64_equivalent"`
	OriginASNs        []asnStat `json:"origin_asns,omitempty"`
	AdmissionBlocks   []allocation `json:"admission_blocks,omitempty"`
}

type manifest struct {
	GeneratedAt string                    `json:"generated_at"`
	Scope       string                    `json:"scope"`
	OutputKind  string                    `json:"output_kind"`
	Sources     map[string]sourceMeta     `json:"sources"`
	Policy      buildPolicy               `json:"policy"`
	RIS         riswhois6.Stats           `json:"ris"`
	Operators   map[string]operatorOutput `json:"operators"`
}

type buildPolicy struct {
	AddressFamily         string `json:"address_family"`
	AllocationAuthority  string `json:"allocation_authority"`
	AllocationCheck      string `json:"allocation_check"`
	RoutingSnapshot      string `json:"routing_snapshot"`
	OutputUnit           string `json:"output_unit"`
	OriginPolicy         string `json:"origin_policy"`
	SyntheticAggregation bool  `json:"synthetic_aggregation"`
}

type audit struct {
	GeneratedAt        string                    `json:"generated_at"`
	Scope              string                    `json:"scope"`
	RIS                riswhois6.Stats           `json:"ris"`
	AcceptedByOperator map[string]int `json:"accepted_prefixes_by_operator"`
	RejectedByReason   map[string]int             `json:"rejected_prefixes_by_reason"`
	OriginASNs         map[string][]asnStat       `json:"origin_asns"`
	Atlas              atlasAudit                 `json:"ripe_atlas_access_evidence"`
}

type atlasResponse struct {
	Count   int          `json:"count"`
	Results []atlasProbe `json:"results"`
}

type atlasProbe struct {
	ID          int         `json:"id"`
	ASNv6       int         `json:"asn_v6"`
	PrefixV6    string      `json:"prefix_v6"`
	Description string      `json:"description"`
	IsAnchor    bool        `json:"is_anchor"`
	IsPublic    bool        `json:"is_public"`
	Status      atlasStatus `json:"status"`
	Tags        []atlasTag  `json:"tags"`
}

type atlasStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type atlasTag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type atlasEvidence struct {
	Operator          string   `json:"operator"`
	ProbeID           int      `json:"probe_id"`
	Prefix            string   `json:"prefix"`
	ASN               string   `json:"asn"`
	Status            string   `json:"status"`
	AccessClass       string   `json:"access_class,omitempty"`
	PositiveSignals   []string `json:"positive_signals,omitempty"`
	NegativeSignals   []string `json:"negative_signals,omitempty"`
	ExactCurrentBGP   bool     `json:"exact_current_bgp_prefix"`
	AdmissionEligible bool     `json:"admission_eligible"`
	DecisionReason    string   `json:"decision_reason"`
}

type atlasAudit struct {
	SourceProbeCount    map[string]int  `json:"source_probe_count"`
	ConnectedProbeCount map[string]int  `json:"connected_probe_count"`
	EligiblePrefixCount map[string]int  `json:"eligible_prefix_count"`
	RejectedByReason    map[string]int  `json:"rejected_by_reason"`
	Evidence            []atlasEvidence `json:"evidence"`
}

func main() {
	risPath := flag.String("ris", "", "RIPE RISWhois IPv6 dump")
	iptoasnPath := flag.String("iptoasn", "", "IPtoASN IPv6 TSV gzip, used only for ASN names")
	configPath := flag.String("operator-config", "config/operators.json", "operator config")
	allocationPath := flag.String("allocation-config", "config/ipv6_allocations.json", "official operator IPv6 allocation boundaries")
	atlasDir := flag.String("atlas-dir", "", "directory containing RIPE Atlas probe API responses for every operator allocation")
	chinanetOutput := flag.String("chinanet-output", "data/ipv6/operators/chinanet.txt", "China Telecom exact BGP candidate prefixes")
	manifestPath := flag.String("manifest", "data/ipv6/manifest.json", "IPv6 manifest")
	auditJSONPath := flag.String("audit-json", "reports/ipv6/bgp-candidates.json", "BGP candidate audit JSON")
	auditMarkdownPath := flag.String("audit-markdown", "reports/ipv6/bgp-candidates.md", "BGP candidate audit Markdown")
	flag.Parse()
	if *risPath == "" || *iptoasnPath == "" || *atlasDir == "" {
		panic("--ris, --iptoasn, and --atlas-dir are required")
	}

	classifier, err := operatorconfig.Load(*configPath, operators)
	must(err)
	allocations, err := readAllocations(*allocationPath)
	must(err)
	metadata, err := readASNMetadata(*iptoasnPath)
	must(err)
	records, risStats, err := riswhois6.ParseGzip(*risPath)
	must(err)

	accepted, rejected := selectPrefixes(records, metadata, classifier, allocations)
	atlasResult, err := auditAtlasEvidence(*atlasDir, accepted, allocations)
	must(err)
	chinanet := accepted["chinanet"]
	must(writePrefixes(*chinanetOutput, chinanet))

	generatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	chinanetASNs := summarizeASNs(chinanet, metadata)
	sources := map[string]sourceMeta{}
	for name, item := range map[string]struct{ path, source string }{
		"riswhois_ipv6":                  {*risPath, "https://www.ris.ripe.net/dumps/riswhoisdump.IPv6.gz"},
		"iptoasn_v6_asn_metadata_only":   {*iptoasnPath, "https://iptoasn.com/data/ip2asn-v6.tsv.gz"},
		"operator_config":                {*configPath, filepath.ToSlash(*configPath)},
		"operator_ipv6_allocations":      {*allocationPath, filepath.ToSlash(*allocationPath)},
	} {
		meta, err := fileMetadata(item.path)
		must(err)
		meta.Source = item.source
		sources[name] = meta
	}
	for _, operator := range operators {
		path := filepath.Join(*atlasDir, operator+".json")
		meta, err := fileMetadata(path)
		must(err)
		meta.Source = atlasURL(operator)
		sources["ripe_atlas_probes_"+operator] = meta
	}
	result := manifest{
		GeneratedAt: generatedAt,
		Scope: "Current exact IPv6 BGP prefixes fully contained by each operator's official APNIC 240x allocation and whose complete observed Origin set is attributed to that operator",
		OutputKind: "exact_bgp_prefix_candidates",
		Sources: sources,
		Policy: buildPolicy{
			AddressFamily: "IPv6", AllocationAuthority: "APNIC",
			AllocationCheck: "versioned_static_config", RoutingSnapshot: "RIPE_RISWhois",
			OutputUnit: "exact_observed_bgp_prefix", OriginPolicy: "complete_origin_set_same_operator",
			SyntheticAggregation: false,
		},
		RIS: risStats,
		Operators: map[string]operatorOutput{
			"chinanet": {
				Status: "candidate",
				Path: filepath.ToSlash(*chinanetOutput),
				PrefixCount: len(chinanet),
				Slash64Equivalent: uniqueSlash64(chinanet),
				OriginASNs: chinanetASNs,
				AdmissionBlocks: publicAllocations(allocations["chinanet"]),
			},
			"cmcc": {Status: "not_emitted", Slash64Equivalent: "0.0000", AdmissionBlocks: publicAllocations(allocations["cmcc"])},
			"unicom": {Status: "not_emitted", Slash64Equivalent: "0.0000", AdmissionBlocks: publicAllocations(allocations["unicom"])},
		},
	}
	auditValue := audit{
		GeneratedAt: generatedAt,
		Scope: result.Scope,
		RIS: risStats,
		AcceptedByOperator: map[string]int{
			"chinanet": len(accepted["chinanet"]),
			"cmcc": len(accepted["cmcc"]),
			"unicom": len(accepted["unicom"]),
		},
		RejectedByReason: rejected,
		OriginASNs: map[string][]asnStat{
			"chinanet": summarizeASNs(accepted["chinanet"], metadata),
			"cmcc": summarizeASNs(accepted["cmcc"], metadata),
			"unicom": summarizeASNs(accepted["unicom"], metadata),
		},
		Atlas: atlasResult,
	}
	must(writeJSON(*manifestPath, result))
	must(writeJSON(*auditJSONPath, auditValue))
	must(writeFile(*auditMarkdownPath, []byte(renderMarkdown(auditValue))))
}

func selectPrefixes(records []riswhois6.Record, metadata map[string]asMeta, classifier *operatorconfig.Classifier, allocations map[string][]allocation) (map[string][]riswhois6.Record, map[string]int) {
	accepted := map[string][]riswhois6.Record{}
	rejected := map[string]int{}
	for _, record := range records {
		operator := ""
		reason := ""
		for _, origin := range record.Origins {
			meta, ok := metadata[origin.ASN]
			if !ok || meta.Description == "" {
				reason = "missing_asn_metadata"
				break
			}
			if !strings.EqualFold(meta.Country, "CN") {
				reason = "non_cn_origin"
				break
			}
			result := classifier.Classify(origin.ASN, meta.Description)
			if result.Excluded {
				reason = "excluded_origin"
				break
			}
			if result.Operator == "" {
				reason = "non_operator_origin"
				break
			}
			if operator != "" && operator != result.Operator {
				reason = "cross_operator_moas"
				break
			}
			operator = result.Operator
		}
		if reason != "" || operator == "" {
			if reason == "" {
				reason = "no_origin"
			}
			rejected[reason]++
			continue
		}
		if !insideAllocation(record.Prefix, allocations[operator]) {
			rejected["outside_operator_240x_allocation"]++
			continue
		}
		accepted[operator] = append(accepted[operator], record)
	}
	return accepted, rejected
}

func auditAtlasEvidence(dir string, candidates map[string][]riswhois6.Record, allocations map[string][]allocation) (atlasAudit, error) {
	result := atlasAudit{
		SourceProbeCount:    map[string]int{},
		ConnectedProbeCount: map[string]int{},
		EligiblePrefixCount: map[string]int{},
		RejectedByReason:    map[string]int{},
	}
	for _, operator := range operators {
		b, err := os.ReadFile(filepath.Join(dir, operator+".json"))
		if err != nil {
			return result, fmt.Errorf("read RIPE Atlas evidence for %s: %w", operator, err)
		}
		var response atlasResponse
		if err := json.Unmarshal(b, &response); err != nil {
			return result, fmt.Errorf("parse RIPE Atlas evidence for %s: %w", operator, err)
		}
		if response.Count != len(response.Results) {
			return result, fmt.Errorf("RIPE Atlas pagination is incomplete for %s: count=%d results=%d", operator, response.Count, len(response.Results))
		}
		result.SourceProbeCount[operator] = response.Count
		current := map[string]map[string]bool{}
		for _, record := range candidates[operator] {
			origins := map[string]bool{}
			for _, origin := range record.Origins {
				origins[origin.ASN] = true
			}
			current[record.Prefix.String()] = origins
		}
		eligiblePrefixes := map[string]bool{}
		for _, probe := range response.Results {
			evidence := classifyAtlasProbe(operator, probe, current, allocations[operator])
			if probe.Status.ID == 1 {
				result.ConnectedProbeCount[operator]++
			}
			if evidence.AdmissionEligible {
				eligiblePrefixes[evidence.Prefix] = true
			} else {
				result.RejectedByReason[evidence.DecisionReason]++
			}
			result.Evidence = append(result.Evidence, evidence)
		}
		result.EligiblePrefixCount[operator] = len(eligiblePrefixes)
	}
	sort.Slice(result.Evidence, func(i, j int) bool {
		if result.Evidence[i].Operator != result.Evidence[j].Operator {
			return result.Evidence[i].Operator < result.Evidence[j].Operator
		}
		if result.Evidence[i].AdmissionEligible != result.Evidence[j].AdmissionEligible {
			return result.Evidence[i].AdmissionEligible
		}
		if result.Evidence[i].Prefix != result.Evidence[j].Prefix {
			return result.Evidence[i].Prefix < result.Evidence[j].Prefix
		}
		return result.Evidence[i].ProbeID < result.Evidence[j].ProbeID
	})
	return result, nil
}

func classifyAtlasProbe(operator string, probe atlasProbe, current map[string]map[string]bool, allocations []allocation) atlasEvidence {
	evidence := atlasEvidence{
		Operator: operator,
		ProbeID: probe.ID,
		Prefix: probe.PrefixV6,
		ASN: fmt.Sprintf("%d", probe.ASNv6),
		Status: probe.Status.Name,
	}
	prefix, err := netip.ParsePrefix(probe.PrefixV6)
	if err != nil || !prefix.Addr().Is6() || prefix.Addr().Is4In6() || prefix != prefix.Masked() || !insideAllocation(prefix, allocations) {
		evidence.DecisionReason = "invalid_or_outside_operator_allocation"
		return evidence
	}
	positiveFixed := map[string]bool{"home": true, "ftth": true, "gpon": true, "pppoe": true, "dsl": true, "residential": true}
	positiveMobile := map[string]bool{"mobile": true, "cellular": true, "lte": true, "4g": true, "5g": true}
	negative := map[string]bool{"office": true, "datacentre": true, "data-centre": true, "data-center": true, "hosting": true, "cloud": true, "server": true, "anchor": true}
	for _, tag := range probe.Tags {
		slug := strings.ToLower(strings.TrimSpace(tag.Slug))
		if positiveFixed[slug] {
			evidence.PositiveSignals = append(evidence.PositiveSignals, "tag:"+slug)
			evidence.AccessClass = "fixed_access"
		}
		if positiveMobile[slug] {
			evidence.PositiveSignals = append(evidence.PositiveSignals, "tag:"+slug)
			evidence.AccessClass = "mobile_access"
		}
		if negative[slug] {
			evidence.NegativeSignals = append(evidence.NegativeSignals, "tag:"+slug)
		}
	}
	description := strings.ToLower(probe.Description)
	for _, phrase := range []string{"home", "residential", "ftth", "gpon", "pppoe"} {
		if strings.Contains(description, phrase) {
			evidence.PositiveSignals = append(evidence.PositiveSignals, "description:"+phrase)
			evidence.AccessClass = "fixed_access"
		}
	}
	for _, phrase := range []string{"mobile access", "mobile broadband", "cellular", " lte", " 4g", " 5g"} {
		if strings.Contains(description, phrase) {
			evidence.PositiveSignals = append(evidence.PositiveSignals, "description:"+strings.TrimSpace(phrase))
			evidence.AccessClass = "mobile_access"
		}
	}
	for _, phrase := range []string{"office", "datacentre", "data centre", "data center", "hosting", "cloud", "server"} {
		if strings.Contains(description, phrase) {
			evidence.NegativeSignals = append(evidence.NegativeSignals, "description:"+phrase)
		}
	}
	origins, exact := current[prefix.String()]
	evidence.ExactCurrentBGP = exact && origins[evidence.ASN]
	switch {
	case probe.Status.ID != 1:
		evidence.DecisionReason = "probe_not_currently_connected"
	case probe.IsAnchor:
		evidence.DecisionReason = "atlas_anchor"
	case !probe.IsPublic:
		evidence.DecisionReason = "probe_not_public"
	case len(evidence.NegativeSignals) > 0:
		evidence.DecisionReason = "explicit_non_target_probe_signal"
	case len(evidence.PositiveSignals) == 0:
		evidence.DecisionReason = "no_explicit_access_signal"
	case !evidence.ExactCurrentBGP:
		evidence.DecisionReason = "not_exact_current_bgp_prefix_and_origin"
	default:
		evidence.AdmissionEligible = true
		evidence.DecisionReason = "eligible_exact_current_bgp_prefix_with_explicit_access_signal"
	}
	return evidence
}

func readAllocations(path string) (map[string][]allocation, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg allocationConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse IPv6 allocation config: %w", err)
	}
	if len(cfg.Operators) != len(operators) {
		return nil, fmt.Errorf("IPv6 allocation config has %d operators, want %d", len(cfg.Operators), len(operators))
	}
	for _, operator := range operators {
		rows := cfg.Operators[operator]
		if len(rows) == 0 {
			return nil, fmt.Errorf("IPv6 allocation config has no blocks for %s", operator)
		}
		for i := range rows {
			prefix, err := netip.ParsePrefix(rows[i].Prefix)
			if err != nil || !prefix.Addr().Is6() || prefix.Addr().Is4In6() || prefix != prefix.Masked() {
				return nil, fmt.Errorf("invalid IPv6 allocation %q for %s", rows[i].Prefix, operator)
			}
			if rows[i].Registry != "APNIC" || rows[i].Organization == "" || rows[i].Source == "" {
				return nil, fmt.Errorf("IPv6 allocation %q for %s lacks APNIC evidence", rows[i].Prefix, operator)
			}
			rows[i].parsed = prefix
		}
		cfg.Operators[operator] = rows
	}
	return cfg.Operators, nil
}

func insideAllocation(prefix netip.Prefix, allocations []allocation) bool {
	prefix = prefix.Masked()
	last := lastAddress(prefix)
	for _, allocation := range allocations {
		if allocation.parsed.Contains(prefix.Addr()) && allocation.parsed.Contains(last) {
			return true
		}
	}
	return false
}

func lastAddress(prefix netip.Prefix) netip.Addr {
	b := prefix.Masked().Addr().As16()
	for bit := prefix.Bits(); bit < 128; bit++ {
		b[bit/8] |= 1 << uint(7-bit%8)
	}
	return netip.AddrFrom16(b)
}

func publicAllocations(rows []allocation) []allocation {
	out := append([]allocation(nil), rows...)
	for i := range out {
		out[i].parsed = netip.Prefix{}
	}
	return out
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

func summarizeASNs(records []riswhois6.Record, metadata map[string]asMeta) []asnStat {
	counts := map[string]int{}
	for _, record := range records {
		for _, origin := range record.Origins {
			counts[origin.ASN]++
		}
	}
	out := make([]asnStat, 0, len(counts))
	for asn, count := range counts {
		out = append(out, asnStat{ASN: asn, Description: metadata[asn].Description, PrefixCount: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PrefixCount != out[j].PrefixCount {
			return out[i].PrefixCount > out[j].PrefixCount
		}
		return out[i].ASN < out[j].ASN
	})
	return out
}

func uniqueSlash64(records []riswhois6.Record) string {
	rows := make([]ipset6.Range, 0, len(records))
	for _, record := range records {
		row, err := ipset6.FromPrefix(record.Prefix)
		must(err)
		rows = append(rows, row)
	}
	count := ipset6.AddressCount(ipset6.Merge(rows))
	return new(big.Rat).SetFrac(count, new(big.Int).Lsh(big.NewInt(1), 64)).FloatString(4)
}

func writePrefixes(path string, records []riswhois6.Record) error {
	var b strings.Builder
	for _, record := range records {
		fmt.Fprintln(&b, record.Prefix)
	}
	return writeFile(path, []byte(b.String()))
}

func atlasURL(operator string) string {
	prefix := map[string]string{
		"chinanet": "240e%3A%3A%2F18",
		"cmcc":     "2409%3A8000%3A%3A%2F20",
		"unicom":   "2408%3A8000%3A%3A%2F20",
	}[operator]
	return "https://atlas.ripe.net/api/v2/probes/?country_code=CN&prefix_v6=" + prefix + "&page_size=500"
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

func renderMarkdown(r audit) string {
	var b strings.Builder
	fmt.Fprintln(&b, "# 三网 IPv6 原始 BGP 前缀审计")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "生成时间：`%s`\n\n", r.GeneratedAt)
	fmt.Fprintln(&b, "输入边界由 APNIC RDAP 实时校验；路由状态来自 RIPE RISWhois；统计单位为当前观测到的原始 BGP 前缀。")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "BGP 只能证明前缀和 Origin，不能证明终端用户接入用途。RIPE Atlas 探针在本报告中仅作为正向接入证据；本轮只审计，不自动扩大正式白名单。")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "RIS 行：**%d**；原始 IPv6 前缀：**%d**。\n", r.RIS.Rows, r.RIS.IPv6Prefixes)
	fmt.Fprintln(&b, "\n## 三网 Origin 候选")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 运营商 | 原始 BGP 前缀 | 状态 |")
	fmt.Fprintln(&b, "| --- | ---: | --- |")
	fmt.Fprintf(&b, "| chinanet | %d | candidate |\n", r.AcceptedByOperator["chinanet"])
	fmt.Fprintf(&b, "| cmcc | %d | not_emitted |\n", r.AcceptedByOperator["cmcc"])
	fmt.Fprintf(&b, "| unicom | %d | not_emitted |\n", r.AcceptedByOperator["unicom"])
	fmt.Fprintln(&b, "\n## RIPE Atlas 正向接入证据")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "只有当前在线、公开、非 Anchor、带明确终端接入信号、无办公/机房反证，且 Atlas 前缀与当前 RIS BGP 前缀及 Origin 精确一致的样本，才标记为可提升准入。")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 运营商 | API 探针 | 当前在线 | 可提升的精确 BGP 前缀 |")
	fmt.Fprintln(&b, "| --- | ---: | ---: | ---: |")
	for _, operator := range operators {
		fmt.Fprintf(&b, "| %s | %d | %d | %d |\n", operator, r.Atlas.SourceProbeCount[operator], r.Atlas.ConnectedProbeCount[operator], r.Atlas.EligiblePrefixCount[operator])
	}
	fmt.Fprintln(&b, "\n### 可提升样本")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 运营商 | BGP 前缀 | Origin | 类型 | 证据 | 探针 |")
	fmt.Fprintln(&b, "| --- | --- | ---: | --- | --- | ---: |")
	for _, row := range r.Atlas.Evidence {
		if !row.AdmissionEligible {
			continue
		}
		fmt.Fprintf(&b, "| %s | `%s` | AS%s | %s | %s | %d |\n", row.Operator, row.Prefix, row.ASN, row.AccessClass, strings.Join(row.PositiveSignals, ", "), row.ProbeID)
	}
	fmt.Fprintln(&b, "\n## 三网候选 Origin ASN")
	for _, operator := range operators {
		fmt.Fprintf(&b, "\n### %s\n\n", operator)
		fmt.Fprintln(&b, "| ASN | 原始 BGP 前缀 | 描述 |")
		fmt.Fprintln(&b, "| --- | ---: | --- |")
		for _, row := range r.OriginASNs[operator] {
			fmt.Fprintf(&b, "| AS%s | %d | %s |\n", row.ASN, row.PrefixCount, strings.ReplaceAll(row.Description, "|", "\\|"))
		}
	}
	fmt.Fprintln(&b, "\n## 未进入三网候选的原因")
	fmt.Fprintln(&b)
	keys := make([]string, 0, len(r.RejectedByReason))
	for key := range r.RejectedByReason {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(&b, "- `%s`: %d\n", key, r.RejectedByReason[key])
	}
	return b.String()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
