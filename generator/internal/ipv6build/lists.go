package ipv6build

import (
	"bytes"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/iputil"
)

// PublicLists is the public, access-type-neutral IPv6 output. Fixed/mobile
// distinctions are consumed during validation and province assignment only.
type PublicLists struct {
	CN        []netip.Prefix
	Operators map[string][]netip.Prefix
	Provinces map[string][]netip.Prefix
}

// BuildPublicLists assigns every admitted BGP prefix to exactly one configured
// province, then produces canonical, sorted, non-overlapping minimal CIDR sets.
func BuildPublicLists(admitted map[string][]netip.Prefix, cfg *AllocationConfig) (*PublicLists, error) {
	if err := cfg.prepare(); err != nil {
		return nil, err
	}
	for operator := range admitted {
		if !isKnownOperator(operator) {
			return nil, fmt.Errorf("admitted prefixes contain unknown operator %q", operator)
		}
	}

	result := &PublicLists{
		Operators: make(map[string][]netip.Prefix, len(operatorNames)),
		Provinces: make(map[string][]netip.Prefix, len(cfg.Provinces)),
	}
	for _, province := range cfg.Provinces {
		result.Provinces[province.Slug] = nil
	}

	var all []netip.Prefix
	for _, operator := range operatorNames {
		var operatorPrefixes []netip.Prefix
		for _, raw := range admitted[operator] {
			prefix, err := canonicalIPv6Prefix(raw)
			if err != nil {
				return nil, fmt.Errorf("operator %s: %w", operator, err)
			}
			allocation, ok := cfg.allocationFor(operator, prefix)
			if !ok {
				return nil, fmt.Errorf("operator %s prefix %s is not contained by exactly one provincial allocation", operator, prefix)
			}
			operatorPrefixes = append(operatorPrefixes, prefix)
			result.Provinces[allocation.ProvinceSlug] = append(result.Provinces[allocation.ProvinceSlug], prefix)
		}
		result.Operators[operator] = CollapsePrefixes(operatorPrefixes)
		all = append(all, result.Operators[operator]...)
	}
	if err := rejectCrossOperatorOverlaps(result.Operators); err != nil {
		return nil, err
	}
	result.CN = CollapsePrefixes(all)
	for slug, prefixes := range result.Provinces {
		result.Provinces[slug] = CollapsePrefixes(prefixes)
	}
	if err := validateRequiredPublicLists(result); err != nil {
		return nil, err
	}
	return result, nil
}

// validateRequiredPublicLists enforces the availability contract for public
// IPv6 output. Provinces are intentionally allowed to be empty: a province may
// temporarily have no admitted BGP prefix even while the national and operator
// lists remain healthy.
func validateRequiredPublicLists(lists *PublicLists) error {
	if lists == nil {
		return fmt.Errorf("public IPv6 lists are nil")
	}
	if len(lists.CN) == 0 {
		return fmt.Errorf("public IPv6 cn list is empty")
	}
	for _, operator := range operatorNames {
		if len(lists.Operators[operator]) == 0 {
			return fmt.Errorf("public IPv6 operator list %s is empty", operator)
		}
	}
	return nil
}

func canonicalIPv6Prefix(prefix netip.Prefix) (netip.Prefix, error) {
	if !prefix.IsValid() {
		return netip.Prefix{}, fmt.Errorf("invalid prefix")
	}
	prefix = prefix.Masked()
	if !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
		return netip.Prefix{}, fmt.Errorf("prefix %s is not IPv6", prefix)
	}
	return prefix, nil
}

// CollapsePrefixes returns the minimal equivalent CIDR set. It masks host
// bits, removes duplicates and covered subprefixes, and recursively combines
// sibling prefixes.
func CollapsePrefixes(values []netip.Prefix) []netip.Prefix {
	set := make(map[netip.Prefix]struct{}, len(values))
	for _, value := range values {
		if !value.IsValid() {
			continue
		}
		value = value.Masked()
		if !value.Addr().Is6() || value.Addr().Is4In6() {
			continue
		}
		set[value] = struct{}{}
	}

	ordered := make([]netip.Prefix, 0, len(set))
	for prefix := range set {
		ordered = append(ordered, prefix)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Bits() != ordered[j].Bits() {
			return ordered[i].Bits() < ordered[j].Bits()
		}
		return ordered[i].Addr().Compare(ordered[j].Addr()) < 0
	})

	kept := make(map[netip.Prefix]struct{}, len(ordered))
	for _, prefix := range ordered {
		covered := false
		for bits := 0; bits < prefix.Bits(); bits++ {
			parent := netip.PrefixFrom(prefix.Addr(), bits).Masked()
			if _, ok := kept[parent]; ok {
				covered = true
				break
			}
		}
		if !covered {
			kept[prefix] = struct{}{}
		}
	}

	for bits := 128; bits > 0; bits-- {
		childrenByParent := make(map[netip.Prefix][]netip.Prefix)
		for prefix := range kept {
			if prefix.Bits() != bits {
				continue
			}
			parent := netip.PrefixFrom(prefix.Addr(), bits-1).Masked()
			childrenByParent[parent] = append(childrenByParent[parent], prefix)
		}
		for parent, children := range childrenByParent {
			if len(children) != 2 {
				continue
			}
			delete(kept, children[0])
			delete(kept, children[1])
			kept[parent] = struct{}{}
		}
	}

	out := make([]netip.Prefix, 0, len(kept))
	for prefix := range kept {
		out = append(out, prefix)
	}
	sort.Slice(out, func(i, j int) bool {
		if cmp := out[i].Addr().Compare(out[j].Addr()); cmp != 0 {
			return cmp < 0
		}
		return out[i].Bits() < out[j].Bits()
	})
	return out
}

func rejectCrossOperatorOverlaps(byOperator map[string][]netip.Prefix) error {
	type taggedPrefix struct {
		prefix   netip.Prefix
		operator string
	}
	var values []taggedPrefix
	for operator, prefixes := range byOperator {
		for _, prefix := range prefixes {
			values = append(values, taggedPrefix{prefix: prefix, operator: operator})
		}
	}
	sort.Slice(values, func(i, j int) bool {
		if cmp := values[i].prefix.Addr().Compare(values[j].prefix.Addr()); cmp != 0 {
			return cmp < 0
		}
		return values[i].prefix.Bits() < values[j].prefix.Bits()
	})
	if len(values) == 0 {
		return nil
	}
	active := values[0]
	activeHi := iputil.LastAddress(active.prefix)
	for _, current := range values[1:] {
		if current.prefix.Addr().Compare(activeHi) <= 0 {
			return fmt.Errorf("cross-operator IPv6 overlap: %s %s and %s %s", active.operator, active.prefix, current.operator, current.prefix)
		}
		currentHi := iputil.LastAddress(current.prefix)
		if currentHi.Compare(activeHi) > 0 {
			active, activeHi = current, currentHi
		}
	}
	return nil
}

// Files returns the stable relative-path contract for one address-family
// directory.
func (lists *PublicLists) Files() map[string][]netip.Prefix {
	files := map[string][]netip.Prefix{"cn.txt": lists.CN}
	for _, operator := range operatorNames {
		files[operator+".txt"] = lists.Operators[operator]
	}
	for slug, prefixes := range lists.Provinces {
		files[filepath.ToSlash(filepath.Join("provinces", slug+".txt"))] = prefixes
	}
	return files
}

// WritePublicLists writes cn.txt, three operator files, and all 31 province
// files directly below the supplied IPv6 family directory.
func WritePublicLists(outputDir string, lists *PublicLists) error {
	if err := validateRequiredPublicLists(lists); err != nil {
		return err
	}
	for relativePath, prefixes := range lists.Files() {
		var content bytes.Buffer
		for _, prefix := range prefixes {
			fmt.Fprintln(&content, prefix)
		}
		path := filepath.Join(outputDir, filepath.FromSlash(relativePath))
		if err := writeFileIfChanged(path, content.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func writeFileIfChanged(path string, content []byte) error {
	current, err := os.ReadFile(path)
	if err == nil && bytes.Equal(current, content) {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}
