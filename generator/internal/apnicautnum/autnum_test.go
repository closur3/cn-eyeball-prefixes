package apnicautnum

import (
	"strings"
	"testing"
)

func TestDescriptionMapAddsAPNICOnlyOrigin(t *testing.T) {
	base := map[string]string{
		"4134": "CHINANET-BACKBONE - No.31,Jin-rong Street",
	}
	records := []Record{
		{ASN: "137687", ASName: "CHINATELECOM-HENAN-LUOYANG-IDC"},
		{ASN: "4134", ASName: "CHINANET-BACKBONE"},
	}
	got := DescriptionMap(base, records)
	if got["4134"] != base["4134"] {
		t.Fatalf("existing IPtoASN description changed: %q", got["4134"])
	}
	if got["137687"] != "CHINATELECOM-HENAN-LUOYANG-IDC" {
		t.Fatalf("APNIC-only origin description = %q", got["137687"])
	}
	if strings.Count(got["4134"], "CHINANET-BACKBONE") != 1 {
		t.Fatalf("duplicate AS name appended to existing description: %q", got["4134"])
	}
}

func TestDescriptionMapIncludesAutnumDescr(t *testing.T) {
	got := DescriptionMap(nil, []Record{{
		ASN: "64500", ASName: "EXAMPLE-AS",
		Descriptions: []string{"Example IDC network"},
	}})
	if !strings.Contains(got["64500"], "EXAMPLE-AS") || !strings.Contains(got["64500"], "Example IDC network") {
		t.Fatalf("aut-num description map = %q", got["64500"])
	}
}

func TestRegistryIndexIncludesInactiveExactNameLink(t *testing.T) {
	records := []Record{{ASN: "134253", ASName: "BG-DIGITAL"}}
	if links := NewIndex(records, map[string]string{}).Links([]string{"BG-Digital"}, nil); len(links) != 0 {
		t.Fatalf("active index unexpectedly linked inactive ASN: %#v", links)
	}
	links := NewRegistryIndex(records).Links([]string{"BG-Digital"}, nil)
	if len(links) != 1 || links[0].ASN != "134253" || links[0].Via != "netname_as_name" {
		t.Fatalf("registry index did not retain exact inactive ASN link: %#v", links)
	}
}
