package ipv6build

import (
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"os"
	"sort"
	"strings"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/iputil"
)

const (
	allocationSchemaVersion = 1
	expectedProvinceCount   = 31
)

var operatorNames = iputil.Operators

// AllocationConfig is the machine-readable source of truth for provincial
// IPv6 access allocations. Access type remains available to admission and
// validation code, but is intentionally not part of the public list layout.
type AllocationConfig struct {
	SchemaVersion int              `json:"schema_version"`
	AddressFamily string           `json:"address_family"`
	Provinces     []ProvinceConfig `json:"provinces"`

	allocations []allocation
}

type ProvinceConfig struct {
	Code      string                        `json:"code"`
	Name      string                        `json:"name"`
	Slug      string                        `json:"slug"`
	Operators map[string]AccessPrefixConfig `json:"operators"`
}

type AccessPrefixConfig struct {
	Fixed  []string `json:"fixed"`
	Mobile []string `json:"mobile"`
}

type allocation struct {
	Prefix       netip.Prefix
	ProvinceCode string
	ProvinceName string
	ProvinceSlug string
	Operator     string
	AccessType   string
}

// LoadAllocationConfig decodes and validates the allocation table. Unknown
// fields are rejected so a misspelled key cannot silently remove a province or
// access family from generated lists.
func LoadAllocationConfig(path string) (*AllocationConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg AllocationConfig
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode IPv6 allocation config: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, err
	}
	if err := cfg.prepare(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode IPv6 allocation config: multiple JSON values")
		}
		return fmt.Errorf("decode IPv6 allocation config: %w", err)
	}
	return nil
}

func (cfg *AllocationConfig) prepare() error {
	if cfg == nil {
		return fmt.Errorf("IPv6 allocation config is nil")
	}
	if cfg.SchemaVersion != allocationSchemaVersion {
		return fmt.Errorf("unsupported IPv6 allocation schema_version %d", cfg.SchemaVersion)
	}
	if cfg.AddressFamily != "ipv6" {
		return fmt.Errorf("address_family must be %q, got %q", "ipv6", cfg.AddressFamily)
	}
	if len(cfg.Provinces) != expectedProvinceCount {
		return fmt.Errorf("IPv6 allocation config has %d provinces, want %d", len(cfg.Provinces), expectedProvinceCount)
	}

	codes := make(map[string]struct{}, len(cfg.Provinces))
	names := make(map[string]struct{}, len(cfg.Provinces))
	slugs := make(map[string]struct{}, len(cfg.Provinces))
	var allocations []allocation
	for provinceIndex, province := range cfg.Provinces {
		if !validAdminCode(province.Code) {
			return fmt.Errorf("province %d has invalid administrative code %q", provinceIndex, province.Code)
		}
		if strings.TrimSpace(province.Name) == "" {
			return fmt.Errorf("province %s has an empty name", province.Code)
		}
		if !validSlug(province.Slug) {
			return fmt.Errorf("province %s has invalid slug %q", province.Code, province.Slug)
		}
		if _, exists := codes[province.Code]; exists {
			return fmt.Errorf("duplicate province administrative code %q", province.Code)
		}
		if _, exists := names[province.Name]; exists {
			return fmt.Errorf("duplicate province name %q", province.Name)
		}
		if _, exists := slugs[province.Slug]; exists {
			return fmt.Errorf("duplicate province slug %q", province.Slug)
		}
		codes[province.Code] = struct{}{}
		names[province.Name] = struct{}{}
		slugs[province.Slug] = struct{}{}

		if len(province.Operators) != len(operatorNames) {
			return fmt.Errorf("province %s has %d operators, want %d", province.Code, len(province.Operators), len(operatorNames))
		}
		for operator := range province.Operators {
			if !isKnownOperator(operator) {
				return fmt.Errorf("province %s has unknown operator %q", province.Code, operator)
			}
		}
		for _, operator := range operatorNames {
			access := province.Operators[operator]
			for _, item := range []struct {
				name   string
				values []string
			}{
				{name: "fixed", values: access.Fixed},
				{name: "mobile", values: access.Mobile},
			} {
				if len(item.values) == 0 {
					return fmt.Errorf("province %s operator %s has no %s prefixes", province.Code, operator, item.name)
				}
				for _, value := range item.values {
					prefix, err := netip.ParsePrefix(value)
					if err != nil {
						return fmt.Errorf("province %s operator %s %s prefix %q: %w", province.Code, operator, item.name, value, err)
					}
					prefix = prefix.Masked()
					if !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
						return fmt.Errorf("province %s operator %s %s prefix %q is not IPv6", province.Code, operator, item.name, value)
					}
					if value != prefix.String() {
						return fmt.Errorf("province %s operator %s %s prefix %q is not canonical; use %q", province.Code, operator, item.name, value, prefix)
					}
					allocations = append(allocations, allocation{
						Prefix:       prefix,
						ProvinceCode: province.Code,
						ProvinceName: province.Name,
						ProvinceSlug: province.Slug,
						Operator:     operator,
						AccessType:   item.name,
					})
				}
			}
		}
	}

	sort.Slice(allocations, func(i, j int) bool {
		if cmp := allocations[i].Prefix.Addr().Compare(allocations[j].Prefix.Addr()); cmp != 0 {
			return cmp < 0
		}
		return allocations[i].Prefix.Bits() < allocations[j].Prefix.Bits()
	})
	for i := 1; i < len(allocations); i++ {
		previous, current := allocations[i-1], allocations[i]
		if prefixesOverlap(previous.Prefix, current.Prefix) {
			return fmt.Errorf(
				"overlapping IPv6 allocations: %s/%s/%s %s and %s/%s/%s %s",
				previous.ProvinceCode, previous.Operator, previous.AccessType, previous.Prefix,
				current.ProvinceCode, current.Operator, current.AccessType, current.Prefix,
			)
		}
	}
	cfg.allocations = allocations
	return nil
}

func (cfg *AllocationConfig) allocationFor(operator string, prefix netip.Prefix) (allocation, bool) {
	for _, item := range cfg.allocations {
		if item.Operator == operator &&
			item.Prefix.Bits() <= prefix.Bits() &&
			item.Prefix.Contains(prefix.Addr()) {
			return item, true
		}
	}
	return allocation{}, false
}

func validAdminCode(value string) bool {
	return len(value) == 2 &&
		value[0] >= '0' && value[0] <= '9' &&
		value[1] >= '0' && value[1] <= '9'
}

func validSlug(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func isKnownOperator(value string) bool {
	for _, operator := range operatorNames {
		if value == operator {
			return true
		}
	}
	return false
}

func prefixesOverlap(a, b netip.Prefix) bool {
	return a.Contains(b.Addr()) || b.Contains(a.Addr())
}
