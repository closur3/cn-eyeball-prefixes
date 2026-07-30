package ipv6build

import (
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryAllocationConfig(t *testing.T) {
	cfg, err := LoadAllocationConfig(filepath.Join("..", "..", "config", "ipv6-province-prefixes.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Provinces) != 31 {
		t.Fatalf("province count = %d, want 31", len(cfg.Provinces))
	}
	if len(cfg.allocations) != 328 {
		t.Fatalf("allocation count = %d, want 328", len(cfg.allocations))
	}
}

func TestBuildPublicListsCollapsesWithoutLosingProvince(t *testing.T) {
	cfg, err := LoadAllocationConfig(filepath.Join("..", "..", "config", "ipv6-province-prefixes.json"))
	if err != nil {
		t.Fatal(err)
	}
	admitted := map[string][]netip.Prefix{
		"chinatelecom": {
			netip.MustParsePrefix("240e:470::/31"),
			netip.MustParsePrefix("240e:472::/31"),
			netip.MustParsePrefix("240e:470::/32"),
		},
		"chinamobile": {
			netip.MustParsePrefix("2409:8a00::/31"),
		},
		"chinaunicom": {
			netip.MustParsePrefix("2408:8206::/31"),
		},
	}
	lists, err := BuildPublicLists(admitted, cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := []netip.Prefix{netip.MustParsePrefix("240e:470::/30")}
	if !equalPrefixes(lists.Operators["chinatelecom"], want) {
		t.Fatalf("operator list = %v, want %v", lists.Operators["chinatelecom"], want)
	}
	if !equalPrefixes(lists.Provinces["zhejiang"], want) {
		t.Fatalf("Zhejiang list = %v, want %v", lists.Provinces["zhejiang"], want)
	}
	wantCN := []netip.Prefix{
		netip.MustParsePrefix("2408:8206::/31"),
		netip.MustParsePrefix("2409:8a00::/31"),
		netip.MustParsePrefix("240e:470::/30"),
	}
	if !equalPrefixes(lists.CN, wantCN) {
		t.Fatalf("CN list = %v, want %v", lists.CN, wantCN)
	}
}

func TestBuildPublicListsRejectsEmptyOperator(t *testing.T) {
	cfg, err := LoadAllocationConfig(filepath.Join("..", "..", "config", "ipv6-province-prefixes.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, missing := range operatorNames {
		t.Run(missing, func(t *testing.T) {
			admitted := validTestAdmissions()
			delete(admitted, missing)
			_, err := BuildPublicLists(admitted, cfg)
			if err == nil || !strings.Contains(err.Error(), "operator list "+missing+" is empty") {
				t.Fatalf("BuildPublicLists error = %v, want empty %s operator error", err, missing)
			}
		})
	}
}

func TestVerifyPublicLists(t *testing.T) {
	cfg, err := LoadAllocationConfig(filepath.Join("..", "..", "config", "ipv6-province-prefixes.json"))
	if err != nil {
		t.Fatal(err)
	}
	lists, err := BuildPublicLists(validTestAdmissions(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists.Provinces["xizang"]) != 0 {
		t.Fatalf("Xizang list = %v, want empty test fixture", lists.Provinces["xizang"])
	}
	dir := t.TempDir()
	if err := WritePublicLists(dir, lists); err != nil {
		t.Fatal(err)
	}
	if err := VerifyPublicLists(dir, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyPublicListsRejectsEmptyRequiredFile(t *testing.T) {
	cfg, err := LoadAllocationConfig(filepath.Join("..", "..", "config", "ipv6-province-prefixes.json"))
	if err != nil {
		t.Fatal(err)
	}
	lists, err := BuildPublicLists(validTestAdmissions(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, relativePath := range []string{
		"cn.txt",
		"chinatelecom.txt",
		"chinamobile.txt",
		"chinaunicom.txt",
	} {
		t.Run(relativePath, func(t *testing.T) {
			dir := t.TempDir()
			if err := WritePublicLists(dir, lists); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, relativePath), nil, 0o644); err != nil {
				t.Fatal(err)
			}
			err := VerifyPublicLists(dir, cfg)
			if err == nil || !strings.Contains(err.Error(), "contains no IPv6 prefixes") {
				t.Fatalf("VerifyPublicLists error = %v, want empty required file error", err)
			}
		})
	}
}

func TestWritePublicListsRejectsEmptyRequiredList(t *testing.T) {
	prefix := netip.MustParsePrefix("2001:db8::/32")
	tests := append([]string{"cn"}, operatorNames...)
	for _, missing := range tests {
		t.Run(missing, func(t *testing.T) {
			lists := &PublicLists{
				CN: []netip.Prefix{prefix},
				Operators: map[string][]netip.Prefix{
					"chinatelecom": {prefix},
					"chinamobile":  {prefix},
					"chinaunicom":  {prefix},
				},
			}
			if missing == "cn" {
				lists.CN = nil
			} else {
				lists.Operators[missing] = nil
			}
			if err := WritePublicLists(t.TempDir(), lists); err == nil ||
				!strings.Contains(err.Error(), "is empty") {
				t.Fatalf("WritePublicLists error = %v, want empty required-list error", err)
			}
		})
	}
}

func validTestAdmissions() map[string][]netip.Prefix {
	return map[string][]netip.Prefix{
		"chinatelecom": {netip.MustParsePrefix("240e:470::/30")},
		"chinamobile":  {netip.MustParsePrefix("2409:8a00::/31")},
		"chinaunicom":  {netip.MustParsePrefix("2408:8206::/31")},
	}
}
