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
	Path   string `json:"path"`
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
	Constraints []string                  `json:"constraints"`
	RIS         riswhois6.Stats           `json:"ris"`
	Operators   map[string]operatorOutput `json:"operators"`
}

type audit struct {
	GeneratedAt      string         `json:"generated_at"`
	Scope            string         `json:"scope"`
	RIS              riswhois6.Stats `json:"ris"`
	AcceptedByOperator map[string]int `json:"accepted_prefixes_by_operator"`
	RejectedByReason map[string]int `json:"rejected_prefixes_by_reason"`
	ChinanetASNs     []asnStat      `json:"chinanet_origin_asns"`
}

func main() {
	risPath := flag.String("ris", "", "RIPE RISWhois IPv6 dump")
	iptoasnPath := flag.String("iptoasn", "", "IPtoASN IPv6 TSV gzip, used only for ASN names")
	configPath := flag.String("operator-config", "config/operators.json", "operator config")
	allocationPath := flag.String("allocation-config", "config/ipv6_allocations.json", "official operator IPv6 allocation boundaries")
	chinanetOutput := flag.String("chinanet-output", "data/ipv6/operators/chinanet.txt", "China Telecom exact BGP candidate prefixes")
	manifestPath := flag.String("manifest", "data/ipv6/manifest.json", "IPv6 manifest")
	auditJSONPath := flag.String("audit-json", "reports/ipv6/bgp-candidates.json", "BGP candidate audit JSON")
	auditMarkdownPath := flag.String("audit-markdown", "reports/ipv6/bgp-candidates.md", "BGP candidate audit Markdown")
	flag.Parse()
	if *risPath == "" || *iptoasnPath == "" {
		panic("--ris and --iptoasn are required")
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
	chinanet := accepted["chinanet"]
	must(writePrefixes(*chinanetOutput, chinanet))

	generatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	chinanetASNs := summarizeASNs(chinanet, metadata)
	sources := map[string]sourceMeta{}
	for name, path := range map[string]string{
		"riswhois_ipv6": *risPath,
		"iptoasn_v6_asn_metadata_only": *iptoasnPath,
		"operator_config": *configPath,
		"operator_ipv6_allocations": *allocationPath,
	} {
		meta, err := fileMetadata(path)
		must(err)
		sources[name] = meta
	}
	result := manifest{
		GeneratedAt: generatedAt,
		Scope: "Current exact IPv6 BGP prefixes fully contained by each operator's official APNIC 240x allocation and whose complete observed Origin set is attributed to that operator",
		OutputKind: "exact_bgp_prefix_candidates",
		Sources: sources,
		Constraints: []string{
			"No gaoyifan china6 intersection",
			"No APNIC inet6num input",
			"Admission boundaries are the official allocations 240e::/18, 2409:8000::/20, and 2408:8000::/20 recorded in versioned configuration",
			"Legacy 2001:: and 2400:: operator prefixes are outside the admission boundary",
			"Every emitted CIDR is an exact prefix present in the RIPE RISWhois IPv6 dump",
			"No address subtraction, sibling merging, or synthetic CIDR aggregation",
			"Prefixes with unknown, excluded, or cross-operator Origins are not emitted",
			"BGP candidates are not a terminal-user allowlist until independent positive access-purpose evidence is defined",
		},
		RIS: risStats,
		Operators: map[string]operatorOutput{
			"chinanet": {
				Status: "bgp_candidates_require_access_purpose_evidence",
				Path: filepath.ToSlash(*chinanetOutput),
				PrefixCount: len(chinanet),
				Slash64Equivalent: uniqueSlash64(chinanet),
				OriginASNs: chinanetASNs,
				AdmissionBlocks: publicAllocations(allocations["chinanet"]),
			},
			"cmcc": {Status: "pending_additional_positive_evidence", Slash64Equivalent: "0.0000", AdmissionBlocks: publicAllocations(allocations["cmcc"])},
			"unicom": {Status: "pending_additional_positive_evidence", Slash64Equivalent: "0.0000", AdmissionBlocks: publicAllocations(allocations["unicom"])},
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
		ChinanetASNs: chinanetASNs,
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
	return sourceMeta{Path: filepath.ToSlash(path), Bytes: n, SHA256: hex.EncodeToString(h.Sum(nil))}, nil
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
	fmt.Fprintln(&b, "本报告直接以 RIPE RISWhois 的当前原始 IPv6 宣告前缀为单位；只准入完整位于电信 `240e::/18`、移动 `2409:8000::/20`、联通 `2408:8000::/20` 相应母段内的前缀。不使用 `china6`，不读取 APNIC `inet6num`，不进行地址切除或 CIDR 再聚合。")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "BGP 只能证明前缀和 Origin，不能证明终端用户接入用途。因此电信输出是待继续验证的 BGP 候选，不是已经完成用途证明的正式白名单。")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "RIS 行：**%d**；原始 IPv6 前缀：**%d**。\n", r.RIS.Rows, r.RIS.IPv6Prefixes)
	fmt.Fprintln(&b, "\n## 三网 Origin 候选")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 运营商 | 原始 BGP 前缀 | 状态 |")
	fmt.Fprintln(&b, "| --- | ---: | --- |")
	fmt.Fprintf(&b, "| chinanet | %d | 已输出候选，仍需终端用途正证据 |\n", r.AcceptedByOperator["chinanet"])
	fmt.Fprintf(&b, "| cmcc | %d | 不输出，等待额外正证据 |\n", r.AcceptedByOperator["cmcc"])
	fmt.Fprintf(&b, "| unicom | %d | 不输出，等待额外正证据 |\n", r.AcceptedByOperator["unicom"])
	fmt.Fprintln(&b, "\n## 电信候选 Origin ASN")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| ASN | 原始 BGP 前缀 | 描述 |")
	fmt.Fprintln(&b, "| --- | ---: | --- |")
	for _, row := range r.ChinanetASNs {
		fmt.Fprintf(&b, "| AS%s | %d | %s |\n", row.ASN, row.PrefixCount, strings.ReplaceAll(row.Description, "|", "\\|"))
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
