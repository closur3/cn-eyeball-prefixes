package main

import (
	"net/netip"
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
	accepted, rejected := selectPrefixes(records, metadata, classifier)
	if len(accepted["chinanet"]) != 1 || accepted["chinanet"][0].Prefix.String() != "240e:1::/32" {
		t.Fatalf("unexpected accepted prefixes: %#v", accepted["chinanet"])
	}
	if rejected["excluded_origin"] != 1 || rejected["cross_operator_moas"] != 1 {
		t.Fatalf("unexpected rejected reasons: %#v", rejected)
	}
}
