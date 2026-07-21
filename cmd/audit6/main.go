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
)

var operators = []string{"chinanet", "cmcc", "unicom"}

type originRecord struct {
	Range       ipset6.Range
	ASN         string
	Country     string
	Description string
}

type sourceMeta struct {
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type spaceStat struct {
	CIDRCount         int      `json:"cidr_count"`
	AddressCount      string   `json:"address_count"`
	Slash64Equivalent string   `json:"slash64_equivalent"`
	PercentOfChina6   string   `json:"percent_of_china6,omitempty"`
	Samples           []string `json:"samples,omitempty"`
}

type originStat struct {
	ASN               string `json:"asn"`
	Country           string `json:"country"`
	Description       string `json:"description"`
	Operator          string `json:"operator,omitempty"`
	Excluded          bool   `json:"excluded,omitempty"`
	AddressCount      string `json:"address_count"`
	Slash64Equivalent string `json:"slash64_equivalent"`
	PercentOfChina6   string `json:"percent_of_china6"`
	count             *big.Int
}

type operatorCoverage struct {
	Operator          string    `json:"operator"`
	CurrentOrigin     spaceStat `json:"current_origin"`
	InsideChina6      spaceStat `json:"inside_china6"`
	MissingFromChina6 spaceStat `json:"missing_from_china6"`
}

type report struct {
	GeneratedAt           string                       `json:"generated_at"`
	Scope                 string                       `json:"scope"`
	Sources               map[string]sourceMeta        `json:"sources"`
	InputCIDRCount        int                          `json:"input_cidr_count"`
	CanonicalChina6       spaceStat                    `json:"canonical_china6"`
	RISVisibleChina6      spaceStat                    `json:"ris_visible_china6"`
	NotRISVisible         spaceStat                    `json:"not_ris_visible"`
	IPtoASNVisibleChina6  spaceStat                    `json:"iptoasn_visible_china6"`
	NotIPtoASNVisible     spaceStat                    `json:"not_iptoasn_visible"`
	IPtoASNCountrySpace   map[string]spaceStat         `json:"iptoasn_country_space"`
	APNICDelegatedSpace   map[string]spaceStat         `json:"apnic_delegated_space"`
	OperatorCoverage      []operatorCoverage           `json:"operator_coverage"`
	TopOrigins            []originStat                 `json:"top_origins"`
	AllOriginCount        int                          `json:"all_origin_count"`
	ForeignOrUnknownSpace spaceStat                    `json:"foreign_or_unknown_space"`
}

func main() {
	china6Path := flag.String("china6", "", "gaoyifan china6.txt")
	iptoasnPath := flag.String("iptoasn", "", "IPtoASN IPv6 TSV gzip")
	delegatedPath := flag.String("delegated", "", "APNIC delegated-latest file")
	risPath := flag.String("ris", "", "RIPE RISWhois IPv6 dump")
	configPath := flag.String("operator-config", "config/operators.json", "operator config")
	jsonPath := flag.String("json", "reports/ipv6/china6-validation.json", "JSON report path")
	markdownPath := flag.String("markdown", "reports/ipv6/china6-validation.md", "Markdown report path")
	flag.Parse()
	for name, path := range map[string]string{"china6": *china6Path, "iptoasn": *iptoasnPath, "delegated": *delegatedPath, "ris": *risPath} {
		if path == "" {
			panic("--" + name + " is required")
		}
	}

	classifier, err := operatorconfig.Load(*configPath, operators)
	must(err)
	china6, inputCIDRs := readCIDRs(*china6Path)
	origins := readIPtoASN(*iptoasnPath)
	risVisible := readRIS(*risPath)
	delegated := readDelegated(*delegatedPath)
	total := ipset6.AddressCount(china6)

	visibleRanges := make([]ipset6.Range, 0, len(origins))
	countryRanges := map[string][]ipset6.Range{}
	originCounts := map[string]*originStat{}
	operatorRanges := map[string][]ipset6.Range{}
	var foreignOrUnknown []ipset6.Range
	for _, record := range origins {
		if record.ASN == "0" {
			continue
		}
		visibleRanges = append(visibleRanges, record.Range)
		hits := intersectOne(china6, record.Range)
		if len(hits) == 0 {
			result := classifier.Classify(record.ASN, record.Description)
			if result.Operator != "" && !result.Excluded {
				operatorRanges[result.Operator] = append(operatorRanges[result.Operator], record.Range)
			}
			continue
		}
		country := strings.ToUpper(strings.TrimSpace(record.Country))
		if country == "" {
			country = "UNKNOWN"
		}
		countryRanges[country] = append(countryRanges[country], hits...)
		if country != "CN" {
			foreignOrUnknown = append(foreignOrUnknown, hits...)
		}
		result := classifier.Classify(record.ASN, record.Description)
		key := record.ASN + "\x00" + country + "\x00" + record.Description
		entry := originCounts[key]
		if entry == nil {
			entry = &originStat{ASN: record.ASN, Country: country, Description: record.Description, Operator: result.Operator, Excluded: result.Excluded, count: new(big.Int)}
			originCounts[key] = entry
		}
		entry.count.Add(entry.count, ipset6.AddressCount(hits))
		if result.Operator != "" && !result.Excluded {
			operatorRanges[result.Operator] = append(operatorRanges[result.Operator], record.Range)
		}
	}

	allOrigins := make([]originStat, 0, len(originCounts))
	for _, entry := range originCounts {
		entry.AddressCount = entry.count.String()
		entry.Slash64Equivalent = slash64FromCount(entry.count)
		entry.PercentOfChina6 = percent(entry.count, total)
		allOrigins = append(allOrigins, *entry)
	}
	sort.Slice(allOrigins, func(i, j int) bool {
		if c := allOrigins[i].count.Cmp(allOrigins[j].count); c != 0 {
			return c > 0
		}
		return allOrigins[i].ASN < allOrigins[j].ASN
	})
	topOrigins := allOrigins
	if len(topOrigins) > 50 {
		topOrigins = topOrigins[:50]
	}

	countryStats := map[string]spaceStat{}
	for country, rows := range countryRanges {
		countryStats[country] = makeStat(rows, total, 10)
	}
	delegatedStats := map[string]spaceStat{}
	delegatedCovered := []ipset6.Range{}
	for country, rows := range delegated {
		hits := ipset6.Intersect(china6, rows)
		if len(hits) == 0 {
			continue
		}
		delegatedStats[country] = makeStat(hits, total, 10)
		delegatedCovered = append(delegatedCovered, hits...)
	}
	delegatedStats["UNREGISTERED_OR_NON_APNIC"] = makeStat(ipset6.Subtract(china6, delegatedCovered), total, 20)

	operatorCoverageRows := make([]operatorCoverage, 0, len(operators))
	for _, operator := range operators {
		current := ipset6.Merge(operatorRanges[operator])
		inside := ipset6.Intersect(current, china6)
		missing := ipset6.Subtract(current, china6)
		operatorCoverageRows = append(operatorCoverageRows, operatorCoverage{
			Operator: operator, CurrentOrigin: makeStat(current, total, 10), InsideChina6: makeStat(inside, total, 10), MissingFromChina6: makeStat(missing, total, 30),
		})
	}

	reportValue := report{
		GeneratedAt:           time.Now().UTC().Format(time.RFC3339Nano),
		Scope:                 "Factual validation of gaoyifan/china-operator-ip china6.txt against current IPtoASN IPv6 origins, RIPE RISWhois IPv6 visibility, APNIC delegated country registration, and the repository's three-operator ASN policy; no address is admitted or excluded by this audit",
		Sources:               map[string]sourceMeta{},
		InputCIDRCount:        inputCIDRs,
		CanonicalChina6:       makeStat(china6, total, 0),
		RISVisibleChina6:      makeStat(ipset6.Intersect(china6, risVisible), total, 10),
		NotRISVisible:         makeStat(ipset6.Subtract(china6, risVisible), total, 30),
		IPtoASNVisibleChina6:  makeStat(ipset6.Intersect(china6, visibleRanges), total, 10),
		NotIPtoASNVisible:     makeStat(ipset6.Subtract(china6, visibleRanges), total, 30),
		IPtoASNCountrySpace:   countryStats,
		APNICDelegatedSpace:   delegatedStats,
		OperatorCoverage:      operatorCoverageRows,
		TopOrigins:            topOrigins,
		AllOriginCount:        len(allOrigins),
		ForeignOrUnknownSpace: makeStat(foreignOrUnknown, total, 30),
	}
	for name, path := range map[string]string{"china6": *china6Path, "iptoasn_v6": *iptoasnPath, "apnic_delegated": *delegatedPath, "riswhois_ipv6": *risPath, "operator_config": *configPath} {
		meta, err := fileMeta(path)
		must(err)
		reportValue.Sources[name] = meta
	}
	must(writeJSON(*jsonPath, reportValue))
	must(writeFile(*markdownPath, []byte(renderMarkdown(reportValue))))
}

func readCIDRs(path string) ([]ipset6.Range, int) {
	f, err := os.Open(path)
	must(err)
	defer f.Close()
	var rows []ipset6.Range
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if line == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(line)
		if err != nil || !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
			panic("invalid IPv6 prefix in china6: " + line)
		}
		row, err := ipset6.FromPrefix(prefix)
		must(err)
		rows = append(rows, row)
		count++
	}
	must(scanner.Err())
	return ipset6.Merge(rows), count
}

func readIPtoASN(path string) []originRecord {
	scanner, closeFn := gzipScanner(path)
	defer closeFn()
	var out []originRecord
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 5 {
			continue
		}
		lo, loErr := netip.ParseAddr(fields[0])
		hi, hiErr := netip.ParseAddr(fields[1])
		if loErr != nil || hiErr != nil || !lo.Is6() || !hi.Is6() || lo.Is4In6() || hi.Is4In6() {
			continue
		}
		out = append(out, originRecord{Range: ipset6.Range{Lo: lo, Hi: hi}, ASN: fields[2], Country: fields[3], Description: fields[4]})
	}
	must(scanner.Err())
	sort.Slice(out, func(i, j int) bool { return out[i].Range.Lo.Compare(out[j].Range.Lo) < 0 })
	return out
}

func readRIS(path string) []ipset6.Range {
	scanner, closeFn := gzipScanner(path)
	defer closeFn()
	var out []ipset6.Range
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "route6:") && !strings.HasPrefix(line, "route:") {
			continue
		}
		value := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		prefix, err := netip.ParsePrefix(value)
		if err != nil || !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
			continue
		}
		row, err := ipset6.FromPrefix(prefix)
		must(err)
		out = append(out, row)
	}
	must(scanner.Err())
	return ipset6.Merge(out)
}

func readDelegated(path string) map[string][]ipset6.Range {
	f, err := os.Open(path)
	must(err)
	defer f.Close()
	out := map[string][]ipset6.Range{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "|")
		if len(fields) < 7 || fields[0] != "apnic" || fields[2] != "ipv6" || (fields[6] != "allocated" && fields[6] != "assigned") {
			continue
		}
		prefix, err := netip.ParsePrefix(fields[3] + "/" + fields[4])
		if err != nil || !prefix.Addr().Is6() {
			continue
		}
		row, err := ipset6.FromPrefix(prefix)
		must(err)
		country := strings.ToUpper(fields[1])
		out[country] = append(out[country], row)
	}
	must(scanner.Err())
	for country := range out {
		out[country] = ipset6.Merge(out[country])
	}
	return out
}

func gzipScanner(path string) (*bufio.Scanner, func()) {
	f, err := os.Open(path)
	must(err)
	z, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		panic(err)
	}
	scanner := bufio.NewScanner(z)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	return scanner, func() {
		must(z.Close())
		must(f.Close())
	}
}

func intersectOne(rows []ipset6.Range, target ipset6.Range) []ipset6.Range {
	i := sort.Search(len(rows), func(i int) bool { return rows[i].Hi.Compare(target.Lo) >= 0 })
	var out []ipset6.Range
	for ; i < len(rows) && rows[i].Lo.Compare(target.Hi) <= 0; i++ {
		lo, hi := rows[i].Lo, rows[i].Hi
		if lo.Compare(target.Lo) < 0 {
			lo = target.Lo
		}
		if hi.Compare(target.Hi) > 0 {
			hi = target.Hi
		}
		out = append(out, ipset6.Range{Lo: lo, Hi: hi})
	}
	return out
}

func makeStat(rows []ipset6.Range, total *big.Int, sampleLimit int) spaceStat {
	rows = ipset6.Merge(rows)
	count := ipset6.AddressCount(rows)
	prefixes := ipset6.Prefixes(rows)
	stat := spaceStat{CIDRCount: len(prefixes), AddressCount: count.String(), Slash64Equivalent: slash64FromCount(count)}
	if total.Sign() > 0 {
		stat.PercentOfChina6 = percent(count, total)
	}
	for i := 0; i < len(prefixes) && i < sampleLimit; i++ {
		stat.Samples = append(stat.Samples, prefixes[i].String())
	}
	return stat
}

func slash64FromCount(count *big.Int) string {
	return new(big.Rat).SetFrac(new(big.Int).Set(count), new(big.Int).Lsh(big.NewInt(1), 64)).FloatString(4)
}

func percent(count, total *big.Int) string {
	if total.Sign() == 0 {
		return "0.000000%"
	}
	ratio := new(big.Rat).SetFrac(new(big.Int).Mul(new(big.Int).Set(count), big.NewInt(100)), total)
	return ratio.FloatString(6) + "%"
}

func fileMeta(path string) (sourceMeta, error) {
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

func renderMarkdown(r report) string {
	var b strings.Builder
	fmt.Fprintln(&b, "# gaoyifan `china6.txt` 事实审计")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "生成时间：`%s`\n\n", r.GeneratedAt)
	fmt.Fprintln(&b, "本报告只验证 `china6.txt` 的当前 BGP 可见性、国家登记边界和三网 IPv6 Origin 覆盖，不参与正式地址准入或排除。空间占比按精确 IPv6 地址数量计算，`/64 等价数`用于提高可读性。")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## 总览")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 项目 | CIDR | /64 等价数 | 占 china6 空间 |")
	fmt.Fprintln(&b, "| --- | ---: | ---: | ---: |")
	markdownStatRow(&b, "规范化 china6", r.CanonicalChina6)
	markdownStatRow(&b, "IPtoASN 当前可见", r.IPtoASNVisibleChina6)
	markdownStatRow(&b, "IPtoASN 未覆盖", r.NotIPtoASNVisible)
	markdownStatRow(&b, "RIS 当前可见", r.RISVisibleChina6)
	markdownStatRow(&b, "RIS 未观测", r.NotRISVisible)
	markdownStatRow(&b, "IPtoASN 非 CN/未知", r.ForeignOrUnknownSpace)
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "原始 CIDR：**%d**；规范化后 CIDR：**%d**。\n", r.InputCIDRCount, r.CanonicalChina6.CIDRCount)

	fmt.Fprintln(&b, "\n## IPtoASN 国家字段")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 国家/地区 | CIDR | /64 等价数 | 占比 |")
	fmt.Fprintln(&b, "| --- | ---: | ---: | ---: |")
	for _, key := range sortedStatKeys(r.IPtoASNCountrySpace) {
		markdownStatRow(&b, key, r.IPtoASNCountrySpace[key])
	}

	fmt.Fprintln(&b, "\n## APNIC delegated 登记")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 登记国家/地区 | CIDR | /64 等价数 | 占比 |")
	fmt.Fprintln(&b, "| --- | ---: | ---: | ---: |")
	for _, key := range sortedStatKeys(r.APNICDelegatedSpace) {
		markdownStatRow(&b, key, r.APNICDelegatedSpace[key])
	}

	fmt.Fprintln(&b, "\n## 三网当前 Origin 对 china6 的覆盖")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| 运营商 | 当前 Origin CIDR | 已在 china6 | china6 外 | china6 外 /64 等价数 |")
	fmt.Fprintln(&b, "| --- | ---: | ---: | ---: | ---: |")
	for _, row := range r.OperatorCoverage {
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %s |\n", row.Operator, row.CurrentOrigin.CIDRCount, row.InsideChina6.CIDRCount, row.MissingFromChina6.CIDRCount, row.MissingFromChina6.Slash64Equivalent)
		if len(row.MissingFromChina6.Samples) > 0 {
			fmt.Fprintf(&b, "\n%s 的 china6 外样本：`%s`\n\n", row.Operator, strings.Join(row.MissingFromChina6.Samples, "`, `"))
		}
	}

	fmt.Fprintln(&b, "\n## china6 内地址量最大的 Origin")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| ASN | 国家/地区 | 运营商识别 | /64 等价数 | 占比 | 描述 |")
	fmt.Fprintln(&b, "| --- | --- | --- | ---: | ---: | --- |")
	for _, row := range r.TopOrigins {
		operator := row.Operator
		if operator == "" {
			operator = "—"
		}
		if row.Excluded {
			operator = "显式排除"
		}
		fmt.Fprintf(&b, "| AS%s | %s | %s | %s | %s | %s |\n", row.ASN, row.Country, operator, row.Slash64Equivalent, row.PercentOfChina6, strings.ReplaceAll(row.Description, "|", "\\|"))
	}

	appendSamples(&b, "IPtoASN 未覆盖样本", r.NotIPtoASNVisible.Samples)
	appendSamples(&b, "RIS 未观测样本", r.NotRISVisible.Samples)
	appendSamples(&b, "非 CN/未知样本", r.ForeignOrUnknownSpace.Samples)
	return b.String()
}

func markdownStatRow(b *strings.Builder, name string, stat spaceStat) {
	fmt.Fprintf(b, "| %s | %d | %s | %s |\n", name, stat.CIDRCount, stat.Slash64Equivalent, stat.PercentOfChina6)
}

func sortedStatKeys(values map[string]spaceStat) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func appendSamples(b *strings.Builder, title string, samples []string) {
	if len(samples) == 0 {
		return
	}
	fmt.Fprintf(b, "\n## %s\n\n", title)
	for _, sample := range samples {
		fmt.Fprintf(b, "- `%s`\n", sample)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
