package main

import (
	"net/netip"
	"os"
	"path/filepath"
	"testing"

	"github.com/closur3/cn-eyeball-prefixes/internal/operatorconfig"
	"github.com/closur3/cn-eyeball-prefixes/internal/riswhois6"
)

func TestSelectPrefixesKeepsOnlyCompleteSameOperatorOriginSets(t *testing.T) {
	classifier, err := operatorconfig.Parse([]byte(`{
  "operators": {
    "chinanet": {"description_patterns": ["china telecom"], "include_asns": {}},
    "cmcc": {"description_patterns": ["china mobile"], "include_asns": {}},
    "unicom": {"description_patterns": ["unicom"], "include_asns": {}}
  },
  "exclude_description_rules": [{"pattern": "idc", "reason": "IDC"}],
  "exclude_apnic_inetnum_rules": [{"pattern": "idc", "reason": "IDC"}],
  "independent_legal_entity_patterns": ["limited"],
  "exclude_asns": {}
}`), operators)
	if err != nil {
		t.Fatal(err)
	}
	records := []riswhois6.Record{
		{Prefix: netip.MustParsePrefix("240e:1::/32"), Origins: []riswhois6.Origin{{ASN: "4134", SeenPeers: 100}}},
		{Prefix: netip.MustParsePrefix("240e:2::/32"), Origins: []riswhois6.Origin{{ASN: "4134", SeenPeers: 100}, {ASN: "38283", SeenPeers: 10}}},
		{Prefix: netip.MustParsePrefix("240e:3::/32"), Origins: []riswhois6.Origin{{ASN: "4134", SeenPeers: 100}, {ASN: "9808", SeenPeers: 10}}},
	}
	metadata := map[string]asMeta{
		"4134": {Country: "CN", Description: "China Telecom Backbone"},
		"38283": {Country: "CN", Description: "China Telecom IDC"},
		"9808": {Country: "CN", Description: "China Mobile"},
	}
	allocations := map[string][]allocation{
		"chinanet": []allocation{{Prefix: "240e::/18", parsed: netip.MustParsePrefix("240e::/18")}},
	}
	accepted, rejected := selectPrefixes(records, metadata, classifier, allocations)
	if len(accepted["chinanet"]) != 1 || accepted["chinanet"][0].Prefix.String() != "240e:1::/32" {
		t.Fatalf("unexpected accepted prefixes: %#v", accepted["chinanet"])
	}
	if rejected["excluded_origin"] != 1 || rejected["cross_operator_moas"] != 1 {
		t.Fatalf("unexpected rejected reasons: %#v", rejected)
	}
}

func TestValidateAllocationRDAP(t *testing.T) {
	dir := t.TempDir()
	body := `{
  "handle": "240e::/18",
  "startAddress": "240e::",
  "endAddress": "240e:3fff:ffff:ffff:ffff:ffff:ffff:ffff",
  "ipVersion": "v6",
  "name": "CT-IPv6-Networks",
  "type": "ALLOCATED PORTABLE",
  "country": "CN",
  "status": ["active"],
  "entities": [{"handle": "ORG-CT1-AP"}]
}`
	if err := os.WriteFile(filepath.Join(dir, "chinanet.json"), []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	rows := map[string][]allocation{
		"chinanet": {{Prefix: "240e::/18", Netname: "CT-IPv6-Networks", Status: "ALLOCATED PORTABLE", RDAPFile: "chinanet.json", RequiredEntityHandles: []string{"ORG-CT1-AP"}, parsed: netip.MustParsePrefix("240e::/18")}},
		"cmcc": {},
		"unicom": {},
	}
	if err := validateAllocationRDAP(dir, rows); err != nil {
		t.Fatal(err)
	}
	rows["chinanet"][0].Netname = "wrong"
	if err := validateAllocationRDAP(dir, rows); err == nil {
		t.Fatal("registration drift was accepted")
	}
}

func TestSelectPrefixesRejectsLegacyOperatorSpace(t *testing.T) {
	classifier, err := operatorconfig.Parse([]byte(`{
  "operators": {
    "chinanet": {"description_patterns": ["china telecom"], "include_asns": {}},
    "cmcc": {"description_patterns": ["china mobile"], "include_asns": {}},
    "unicom": {"description_patterns": ["unicom"], "include_asns": {}}
  },
  "exclude_description_rules": [{"pattern": "idc", "reason": "IDC"}],
  "exclude_apnic_inetnum_rules": [{"pattern": "idc", "reason": "IDC"}],
  "independent_legal_entity_patterns": ["limited"],
  "exclude_asns": {}
}`), operators)
	if err != nil {
		t.Fatal(err)
	}
	records := []riswhois6.Record{{Prefix: netip.MustParsePrefix("2001:c68::/32"), Origins: []riswhois6.Origin{{ASN: "4134", SeenPeers: 100}}}}
	metadata := map[string]asMeta{"4134": {Country: "CN", Description: "China Telecom Backbone"}}
	allocations := map[string][]allocation{"chinanet": []allocation{{Prefix: "240e::/18", parsed: netip.MustParsePrefix("240e::/18")}}}
	accepted, rejected := selectPrefixes(records, metadata, classifier, allocations)
	if len(accepted["chinanet"]) != 0 || rejected["outside_operator_240x_allocation"] != 1 {
		t.Fatalf("legacy space was admitted: accepted=%#v rejected=%#v", accepted, rejected)
	}
}

func TestClassifyAtlasProbeRequiresExactCurrentAccessPrefix(t *testing.T) {
	probe := atlasProbe{
		ID: 1, ASNv6: 9808, PrefixV6: "2409:8a55:400::/40", IsPublic: true,
		Status: atlasStatus{ID: 1, Name: "Connected"},
		Tags: []atlasTag{{Slug: "home"}, {Slug: "ftth"}},
	}
	current := map[string]map[string]bool{"2409:8a55:400::/40": {"9808": true}}
	allocations := []allocation{{parsed: netip.MustParsePrefix("2409:8000::/20")}}
	got := classifyAtlasProbe("cmcc", probe, current, allocations)
	if !got.AdmissionEligible || got.AccessClass != "fixed_access" || !got.ExactCurrentBGP {
		t.Fatalf("strong exact Atlas access evidence was not eligible: %#v", got)
	}
	probe.PrefixV6 = "2409:8a55::/32"
	got = classifyAtlasProbe("cmcc", probe, current, allocations)
	if got.AdmissionEligible || got.DecisionReason != "not_exact_current_bgp_prefix_and_origin" {
		t.Fatalf("covering Atlas prefix was admitted without an exact BGP match: %#v", got)
	}
}

func TestClassifyAtlasProbeRejectsOfficeOrDatacentre(t *testing.T) {
	probe := atlasProbe{
		ID: 2, ASNv6: 4134, PrefixV6: "240e:6b0::/28", Description: "office data centre", IsPublic: true,
		Status: atlasStatus{ID: 1, Name: "Connected"},
		Tags: []atlasTag{{Slug: "home"}, {Slug: "office"}},
	}
	current := map[string]map[string]bool{"240e:6b0::/28": {"4134": true}}
	allocations := []allocation{{parsed: netip.MustParsePrefix("240e::/18")}}
	got := classifyAtlasProbe("chinanet", probe, current, allocations)
	if got.AdmissionEligible || got.DecisionReason != "explicit_non_target_probe_signal" {
		t.Fatalf("office/datacentre signal did not override a positive tag: %#v", got)
	}
}
