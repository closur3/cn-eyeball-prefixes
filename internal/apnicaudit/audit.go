package apnicaudit

import (
	"fmt"
	"net/netip"
	"sort"
	"strings"

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
	CIDRCount                         int               `json:"cidr_count"`
	FactCount                         int               `json:"fact_count"`
	AddressCount                      uint64            `json:"address_count"`
	RegistryCoveredAddressCount       uint64            `json:"registry_covered_address_count"`
	RegistryCoveragePercent           float64           `json:"registry_coverage_percent"`
	StrongNonPublicSignalAddressCount uint64            `json:"strong_non_public_signal_address_count"`
	Categories                        []CategorySummary `json:"categories"`
}

type Report struct {
	Scope      string       `json:"scope"`
	Notes      []string     `json:"notes"`
	Summary    Summary      `json:"summary"`
	Comparison *Comparison  `json:"comparison,omitempty"`
	CIDRs      []CIDRRecord `json:"cidrs"`
}

type Exclusion struct {
	CIDR         string `json:"cidr"`
	AddressCount uint64 `json:"address_count"`
	Source       string `json:"source"`
	Category     string `json:"category"`
	Operator     string `json:"operator"`
	ASN          string `json:"asn"`
	Registrant   string `json:"registrant,omitempty"`
	Reason       string `json:"reason"`
}

type ExclusionCategory struct {
	Category     string `json:"category"`
	EvidenceRows int    `json:"evidence_rows"`
	AddressCount uint64 `json:"address_count"`
}

type Comparison struct {
	CandidateAddressCount uint64              `json:"candidate_address_count"`
	ExcludedAddressCount  uint64              `json:"excluded_address_count"`
	RetainedAddressCount  uint64              `json:"retained_address_count"`
	Categories            []ExclusionCategory `json:"categories"`
	Exclusions            []Exclusion         `json:"exclusions"`
}

func AttachComparison(report *Report, candidateAddressCount, excludedAddressCount uint64, exclusions []Exclusion) {
	categories := map[string]*ExclusionCategory{}
	for _, exclusion := range exclusions {
		category := categories[exclusion.Category]
		if category == nil {
			category = &ExclusionCategory{Category: exclusion.Category}
			categories[exclusion.Category] = category
		}
		category.EvidenceRows++
		category.AddressCount += exclusion.AddressCount
	}
	names := make([]string, 0, len(categories))
	for name := range categories {
		names = append(names, name)
	}
	sort.Strings(names)
	comparison := &Comparison{CandidateAddressCount: candidateAddressCount, ExcludedAddressCount: excludedAddressCount, RetainedAddressCount: report.Summary.AddressCount, Exclusions: exclusions}
	for _, name := range names {
		comparison.Categories = append(comparison.Categories, *categories[name])
	}
	report.Comparison = comparison
}

func Build(scope string, cidrs []string, operatorRanges map[string][]Range, segments []apnicinetnum.Segment, classifier *operatorconfig.Classifier) (Report, error) {
	report := Report{
		Scope: scope,
		Notes: []string{
			"Every retained address is mapped to the most-specific APNIC inetnum object available in the build snapshot.",
			"Nationwide admission requires a covering APNIC inetnum object attributable to the same operator as the current BGP origin.",
			"A more-specific customer or legacy registration is retained unless it supplies an enforced strong exclusion or identifies a conflicting operator.",
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
		entry := CIDRRecord{CIDR: cidr, AddressCount: uint64(hi) - uint64(lo) + 1}
		for _, operator := range operators {
			for _, candidate := range overlapping(operatorRanges[operator], lo, hi) {
				start, end := max(lo, candidate.Lo), min(hi, candidate.Hi)
				entry.Facts = append(entry.Facts, registryFacts(start, end, operator, segments, classifier)...)
			}
		}
		sort.Slice(entry.Facts, func(i, j int) bool {
			return number(netip.MustParseAddr(entry.Facts[i].Start)) < number(netip.MustParseAddr(entry.Facts[j].Start))
		})
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
	order := []string{"operator_registration", "operator_registration_conflict", "independent_legal_entity", "other_registration", "unregistered", "strong_non_public_signal"}
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
		classification, reason, matchedBy := classify(segment, operator, classifier)
		out = append(out, Fact{
			Start: addr(start), End: addr(end), AddressCount: uint64(end) - uint64(start) + 1,
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

func classify(segment apnicinetnum.Segment, operator string, classifier *operatorconfig.Classifier) (string, string, string) {
	if segment.Match.Reason != "" {
		return "strong_non_public_signal", segment.Match.Reason, segment.Match.MatchedBy
	}
	text := apnicinetnum.SearchText(segment.Record)
	if prefix := classifier.ClassifyAPNICInetnum(text); prefix.Excluded {
		return "strong_non_public_signal", prefix.Reason, prefix.MatchedBy
	}
	registrant := classifier.ClassifyAPNICRegistrant(text)
	if registrant.Operator == operator {
		return "operator_registration", "APNIC registrant text matches " + registrant.Operator, registrant.MatchedBy
	}
	if registrant.Operator != "" {
		return "operator_registration_conflict", "APNIC registrant text matches " + registrant.Operator + " but BGP origin candidate is " + operator, registrant.MatchedBy
	}
	if classifier.IsIndependentLegalEntity(apnicinetnum.RegistrantText(segment.Record)) {
		return "independent_legal_entity", "Most-specific APNIC registrant text names an independent legal entity; this fact is retained for audit and is excluded only when an enforced second signal links it to an independent network", "independent_legal_entity_patterns"
	}
	return "other_registration", "Most-specific APNIC registration is not attributable to the BGP origin operator; this leaf fact alone is not an exclusion", ""
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
	return Fact{Start: addr(lo), End: addr(hi), AddressCount: uint64(hi) - uint64(lo) + 1, Operator: operator, Classification: "unregistered", Reason: "No APNIC inetnum object covers this address range"}
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

// RenderMarkdown turns the complete machine-readable evidence into a compact
// review report.
func RenderMarkdown(report Report, evidencePath string) string {
	var b strings.Builder
	relaxedAdditions := report.Scope == "Nationwide relaxed-BGP additions versus current dev"
	if relaxedAdditions {
		b.WriteString("# 宽松 BGP 新增地址 APNIC 登记事实审计\n\n")
		b.WriteString("本报告反查真正宽松 BGP 方案相对当前 `dev` 新增的全国地址：当前三网 BGP Origin 直接作为准入证据，不再要求 APNIC 上层登记归属于同一运营商。新增地址已经通过全部现行强排除。完整逐地址事实保存在 [`")
	} else {
		b.WriteString("# 浙江 IPv4 APNIC 登记事实审计\n\n")
		b.WriteString("本报告以浙江输出为样本，复核全国分层准入规则：当前 BGP Origin 必须属于三网，并由同一家运营商的 APNIC 上层登记覆盖；最具体登记用于审计及强证据排除，不再单独决定准入。完整逐地址事实保存在 [`")
	}
	b.WriteString(markdownText(evidencePath))
	b.WriteString("`](./")
	b.WriteString(markdownText(evidencePath))
	b.WriteString(")。\n\n")

	b.WriteString("## 总览\n\n")
	b.WriteString("| 指标 | 数值 |\n|---|---:|\n")
	fmt.Fprintf(&b, "| 最大聚合 ACL CIDR | %s |\n", formatUint(uint64(report.Summary.CIDRCount)))
	fmt.Fprintf(&b, "| IPv4 地址 | %s |\n", formatUint(report.Summary.AddressCount))
	fmt.Fprintf(&b, "| 最具体 APNIC 事实片段 | %s |\n", formatUint(uint64(report.Summary.FactCount)))
	fmt.Fprintf(&b, "| APNIC 登记覆盖 | %s（%.4f%%） |\n", formatUint(report.Summary.RegistryCoveredAddressCount), report.Summary.RegistryCoveragePercent)
	fmt.Fprintf(&b, "| 构建规则仍识别出的强非公众信号 | %s |\n\n", formatUint(report.Summary.StrongNonPublicSignalAddressCount))
	renderComparison(&b, report)

	b.WriteString("## 登记分类\n\n")
	b.WriteString("| 分类 | 事实片段 | 地址 | 占全部地址 | 含义 |\n|---|---:|---:|---:|---|\n")
	meaning := map[string]string{
		"operator_registration":    "登记文本可归属于三网运营商",
		"operator_registration_conflict": "最具体登记运营商与 BGP Origin 运营商不一致，拒绝",
		"independent_legal_entity": "最具体登记为独立法定主体；仅在第二项强证据成立时排除",
		"other_registration":       "最具体登记无法归属对应三网；在上层准入成立时保留审计",
		"unregistered":             "构建快照内没有覆盖该范围的 inetnum",
		"strong_non_public_signal": "命中当前明确非公众用途规则；应优先复核",
	}
	if relaxedAdditions {
		meaning["operator_registration_conflict"] = "最具体登记属于另一家三网运营商；宽松 BGP 方案仍会纳入"
		meaning["independent_legal_entity"] = "最具体登记为独立法定主体；现行强关联不足，但宽松 BGP 方案会纳入"
		meaning["other_registration"] = "最具体登记无法归属对应三网；宽松 BGP 方案会纳入"
	}
	for _, category := range report.Summary.Categories {
		fmt.Fprintf(&b, "| `%s` | %s | %s | %.4f%% | %s |\n", category.Classification, formatUint(uint64(category.FactCount)), formatUint(category.AddressCount), category.AddressPercent, meaning[category.Classification])
	}

	b.WriteString("\n## 怎样阅读\n\n")
	b.WriteString("- ACL 文件采用最大 CIDR 聚合；表中的“保留范围”才是与 APNIC 登记边界对齐后的精确地址范围。\n")
	if relaxedAdditions {
		b.WriteString("- 这些地址尚未进入正式 ACL；它们专门展示取消 APNIC 正向准入后会增加的范围。\n")
	} else {
		b.WriteString("- 全国输出统一采用上层运营商登记准入；最具体登记中的运营商冲突会被拒绝，其他独立主体或归属不明记录只有命中既有强证据规则才会删除。\n")
	}
	b.WriteString("- 排名按覆盖地址量排列，用来优先投入人工审查，不代表风险评分。\n")
	b.WriteString("- 下方索引只负责让主要事实可读；完整证据、全部小片段和全部字段仍以 gzip JSON 为准。\n\n")

	renderStrongSignals(&b, report)
	renderReviewGroups(&b, report, "independent_legal_entity", "独立法定主体登记：地址量前 100 项", 100)
	renderReviewGroups(&b, report, "other_registration", "其他登记：地址量前 100 项", 100)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderComparison(b *strings.Builder, report Report) {
	if report.Comparison == nil {
		return
	}
	c := report.Comparison
	b.WriteString("## 前缀清洗前后对照\n\n")
	b.WriteString("这里的“准入前候选”指已满足三网 Origin 与中国边界、但尚未执行云 CIDR、APNIC、route、MOAS 排除及全国同运营商上层 APNIC 登记准入的浙江地址。分类地址数按证据行累加，多个上游命中同一地址时可能重复；总未准入地址数按地址并集计算。\n\n")
	b.WriteString("| 阶段 | 地址 |\n|---|---:|\n")
	fmt.Fprintf(b, "| 准入前候选 | %s |\n", formatUint(c.CandidateAddressCount))
	fmt.Fprintf(b, "| 未准入（并集） | %s |\n", formatUint(c.ExcludedAddressCount))
	fmt.Fprintf(b, "| 最终保留 | %s |\n\n", formatUint(c.RetainedAddressCount))
	b.WriteString("| 排除类别 | 证据范围 | 地址（可重复） |\n|---|---:|---:|\n")
	for _, category := range c.Categories {
		fmt.Fprintf(b, "| `%s` | %s | %s |\n", markdownText(category.Category), formatUint(uint64(category.EvidenceRows)), formatUint(category.AddressCount))
	}
	b.WriteString("\n### 排除证据样本\n\n")
	b.WriteString("| CIDR | 地址 | 类别 | 运营商 / ASN | APNIC 登记主体 | 原因 |\n|---|---:|---|---|---|---|\n")
	limit := minInt(200, len(c.Exclusions))
	for _, exclusion := range c.Exclusions[:limit] {
		fmt.Fprintf(b, "| `%s` | %s | `%s` | `%s / AS%s` | %s | %s |\n", exclusion.CIDR, formatUint(exclusion.AddressCount), markdownText(exclusion.Category), markdownText(exclusion.Operator), markdownText(exclusion.ASN), markdownText(exclusion.Registrant), markdownText(exclusion.Reason))
	}
	if len(c.Exclusions) > limit {
		fmt.Fprintf(b, "\n其余 %s 条排除证据未在 Markdown 展开；完整内容保存在 gzip JSON 与 manifest。\n", formatUint(uint64(len(c.Exclusions)-limit)))
	}
	b.WriteString("\n")
}

type reviewGroup struct {
	Label        string
	AddressCount uint64
	FactCount    int
	Samples      []string
}

func renderStrongSignals(b *strings.Builder, report Report) {
	b.WriteString("## 当前规则仍识别出的强非公众信号\n\n")
	b.WriteString("这些条目已处于最终 ACL 的登记事实中，应优先检查生成边界为何仍保留它们。\n\n")
	b.WriteString("| 保留范围 | 所属 ACL CIDR | 运营商 | APNIC 登记主体 | APNIC 登记范围 | 命中原因 |\n|---|---|---|---|---|---|\n")
	count := 0
	for _, cidr := range report.CIDRs {
		for _, fact := range cidr.Facts {
			if fact.Classification != "strong_non_public_signal" {
				continue
			}
			count++
			registryRange := "—"
			if fact.Registry != nil {
				registryRange = fact.Registry.Range
			}
			fmt.Fprintf(b, "| `%s` | `%s` | `%s` | %s | `%s` | %s |\n", factRange(fact), cidr.CIDR, fact.Operator, markdownText(registrantLabel(fact.Registry)), markdownText(registryRange), markdownText(fact.Reason))
		}
	}
	if count == 0 {
		b.WriteString("| — | — | — | 当前没有残留强信号 | — | — |\n")
	}
	b.WriteString("\n")
}

func renderReviewGroups(b *strings.Builder, report Report, classification, title string, limit int) {
	groups := map[string]*reviewGroup{}
	for _, cidr := range report.CIDRs {
		for _, fact := range cidr.Facts {
			if fact.Classification != classification {
				continue
			}
			label := registrantLabel(fact.Registry)
			group := groups[label]
			if group == nil {
				group = &reviewGroup{Label: label}
				groups[label] = group
			}
			group.AddressCount += fact.AddressCount
			group.FactCount++
			if len(group.Samples) < 3 {
				sample := "`" + factRange(fact) + "` in `" + cidr.CIDR + "`"
				if len(group.Samples) == 0 || group.Samples[len(group.Samples)-1] != sample {
					group.Samples = append(group.Samples, sample)
				}
			}
		}
	}
	rows := make([]reviewGroup, 0, len(groups))
	for _, group := range groups {
		rows = append(rows, *group)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].AddressCount != rows[j].AddressCount {
			return rows[i].AddressCount > rows[j].AddressCount
		}
		return rows[i].Label < rows[j].Label
	})

	fmt.Fprintf(b, "## %s\n\n", title)
	fmt.Fprintf(b, "共 %s 个登记主体标签；下表展示前 %d 项。标签优先取 APNIC organisation name，其次取 description、netname 或 organisation handle。\n\n", formatUint(uint64(len(rows))), minInt(limit, len(rows)))
	b.WriteString("| # | APNIC 登记主体 | 地址 | 占全部地址 | 事实片段 | 保留范围样本 / 所属 ACL CIDR |\n|---:|---|---:|---:|---:|---|\n")
	if len(rows) == 0 {
		b.WriteString("| — | 无 | 0 | 0% | 0 | — |\n\n")
		return
	}
	for i, group := range rows[:minInt(limit, len(rows))] {
		fmt.Fprintf(b, "| %d | %s | %s | %.4f%% | %s | %s |\n", i+1, markdownText(group.Label), formatUint(group.AddressCount), percent(group.AddressCount, report.Summary.AddressCount), formatUint(uint64(group.FactCount)), strings.Join(group.Samples, "<br>"))
	}
	if len(rows) > limit {
		fmt.Fprintf(b, "\n其余 %s 个较小登记主体标签未在 Markdown 展开，可在完整 gzip JSON 中查询。\n", formatUint(uint64(len(rows)-limit)))
	}
	b.WriteString("\n")
}

func registrantLabel(registry *Registry) string {
	if registry == nil {
		return "（无 APNIC 登记）"
	}
	for _, values := range [][]string{registry.OrganizationNames, registry.Descriptions, registry.Netnames, registry.Organizations} {
		for _, value := range values {
			if value = strings.TrimSpace(value); value != "" {
				return value
			}
		}
	}
	return "（登记主体字段为空）"
}

func factRange(fact Fact) string {
	if fact.Start == fact.End {
		return fact.Start
	}
	return fact.Start + "–" + fact.End
}

func markdownText(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

func formatUint(value uint64) string {
	s := fmt.Sprintf("%d", value)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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
