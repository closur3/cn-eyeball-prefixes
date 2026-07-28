package ipv6build

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// VerifyMain is the IPv6 verifier command entry point.
func VerifyMain() {
	must(RunVerifyCLI(os.Args[1:]))
}

func RunVerifyCLI(args []string) error {
	flags := flag.NewFlagSet("verify ipv6", flag.ContinueOnError)
	dataDir := flags.String("data", "", "staged IPv6 family directory")
	allocationConfigPath := flags.String("allocation-config", "config/ipv6-province-prefixes.json", "provincial IPv6 allocation config")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *dataDir == "" {
		return fmt.Errorf("--data is required")
	}
	cfg, err := LoadAllocationConfig(*allocationConfigPath)
	if err != nil {
		return err
	}
	if err := VerifyPublicLists(*dataDir, cfg); err != nil {
		return err
	}
	fmt.Println("OK: IPv6 public lists are canonical; operator and province coverage matches cn.txt and the allocation table")
	return nil
}

// VerifyPublicLists independently reloads the public files and checks their
// layout, canonical representation, set relationships and province mapping.
func VerifyPublicLists(dataDir string, cfg *AllocationConfig) error {
	if err := cfg.prepare(); err != nil {
		return err
	}
	expectedPaths := map[string]bool{"cn.txt": true}
	for _, operator := range operatorNames {
		expectedPaths[operator+".txt"] = true
	}
	for _, province := range cfg.Provinces {
		expectedPaths[filepath.ToSlash(filepath.Join("provinces", province.Slug+".txt"))] = true
	}
	if err := rejectUnexpectedPublicFiles(dataDir, expectedPaths); err != nil {
		return err
	}

	cn, err := readPublicPrefixFile(filepath.Join(dataDir, "cn.txt"))
	if err != nil {
		return err
	}
	if len(cn) == 0 {
		return fmt.Errorf("cn.txt contains no IPv6 prefixes")
	}
	byOperator := make(map[string][]netip.Prefix, len(operatorNames))
	var operatorUnion []netip.Prefix
	for _, operator := range operatorNames {
		prefixes, err := readPublicPrefixFile(filepath.Join(dataDir, operator+".txt"))
		if err != nil {
			return err
		}
		if len(prefixes) == 0 {
			return fmt.Errorf("%s.txt contains no IPv6 prefixes", operator)
		}
		byOperator[operator] = prefixes
		operatorUnion = append(operatorUnion, prefixes...)
	}
	if err := rejectCrossOperatorOverlaps(byOperator); err != nil {
		return err
	}
	if !equalPrefixes(CollapsePrefixes(operatorUnion), cn) {
		return fmt.Errorf("operator union does not equal cn.txt")
	}

	expectedByProvince := make(map[string][]netip.Prefix, len(cfg.Provinces))
	for _, province := range cfg.Provinces {
		expectedByProvince[province.Slug] = nil
	}
	for operator, prefixes := range byOperator {
		var allocatedCoverage []netip.Prefix
		for _, prefix := range prefixes {
			for _, allocation := range cfg.allocations {
				if allocation.Operator != operator {
					continue
				}
				intersection, ok := intersectPrefixes(prefix, allocation.Prefix)
				if !ok {
					continue
				}
				allocatedCoverage = append(allocatedCoverage, intersection)
				expectedByProvince[allocation.ProvinceSlug] = append(expectedByProvince[allocation.ProvinceSlug], intersection)
			}
		}
		if !equalPrefixes(CollapsePrefixes(allocatedCoverage), prefixes) {
			return fmt.Errorf("operator %s contains coverage outside the provincial allocation table", operator)
		}
	}

	actualByProvince := make(map[string][]netip.Prefix, len(cfg.Provinces))
	for _, province := range cfg.Provinces {
		path := filepath.Join(dataDir, "provinces", province.Slug+".txt")
		prefixes, err := readPublicPrefixFile(path)
		if err != nil {
			return err
		}
		expected := CollapsePrefixes(expectedByProvince[province.Slug])
		if !equalPrefixes(prefixes, expected) {
			return fmt.Errorf("province %s does not match the allocation-derived operator coverage", province.Slug)
		}
		actualByProvince[province.Slug] = prefixes
	}
	if err := rejectCrossProvinceOverlaps(actualByProvince); err != nil {
		return err
	}
	var provinceUnion []netip.Prefix
	for _, prefixes := range actualByProvince {
		provinceUnion = append(provinceUnion, prefixes...)
	}
	if !equalPrefixes(CollapsePrefixes(provinceUnion), cn) {
		return fmt.Errorf("province union does not equal cn.txt")
	}
	return nil
}

func rejectUnexpectedPublicFiles(dataDir string, expected map[string]bool) error {
	var actual []string
	err := filepath.WalkDir(dataDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".txt") {
			return nil
		}
		rel, err := filepath.Rel(dataDir, path)
		if err != nil {
			return err
		}
		actual = append(actual, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(actual)
	if len(actual) != len(expected) {
		return fmt.Errorf("IPv6 public directory has %d TXT files, want %d", len(actual), len(expected))
	}
	for _, rel := range actual {
		if !expected[rel] {
			return fmt.Errorf("unexpected IPv6 public list: %s", rel)
		}
	}
	return nil
}

func readPublicPrefixFile(path string) ([]netip.Prefix, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var out []netip.Prefix
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return nil, fmt.Errorf("%s contains a blank line", path)
		}
		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if !prefix.Addr().Is6() || prefix.Addr().Is4In6() || prefix != prefix.Masked() || line != prefix.String() {
			return nil, fmt.Errorf("%s contains non-canonical IPv6 prefix %q", path, line)
		}
		if len(out) != 0 {
			previous := out[len(out)-1]
			if prefixOrder(previous, prefix) >= 0 {
				return nil, fmt.Errorf("%s is not strictly sorted at %s", path, prefix)
			}
			if previous.Contains(prefix.Addr()) {
				return nil, fmt.Errorf("%s contains overlapping prefixes %s and %s", path, previous, prefix)
			}
		}
		out = append(out, prefix)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func prefixOrder(a, b netip.Prefix) int {
	if cmp := a.Addr().Compare(b.Addr()); cmp != 0 {
		return cmp
	}
	switch {
	case a.Bits() < b.Bits():
		return -1
	case a.Bits() > b.Bits():
		return 1
	default:
		return 0
	}
}

func intersectPrefixes(a, b netip.Prefix) (netip.Prefix, bool) {
	switch {
	case a.Bits() <= b.Bits() && a.Contains(b.Addr()):
		return b, true
	case b.Bits() <= a.Bits() && b.Contains(a.Addr()):
		return a, true
	default:
		return netip.Prefix{}, false
	}
}

func equalPrefixes(a, b []netip.Prefix) bool {
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

func rejectCrossProvinceOverlaps(byProvince map[string][]netip.Prefix) error {
	type tagged struct {
		province string
		prefix   netip.Prefix
	}
	var values []tagged
	for province, prefixes := range byProvince {
		for _, prefix := range prefixes {
			values = append(values, tagged{province: province, prefix: prefix})
		}
	}
	sort.Slice(values, func(i, j int) bool {
		return prefixOrder(values[i].prefix, values[j].prefix) < 0
	})
	for i := 1; i < len(values); i++ {
		previous, current := values[i-1], values[i]
		if previous.prefix.Contains(current.prefix.Addr()) {
			return fmt.Errorf("cross-province IPv6 overlap: %s %s and %s %s", previous.province, previous.prefix, current.province, current.prefix)
		}
	}
	return nil
}
