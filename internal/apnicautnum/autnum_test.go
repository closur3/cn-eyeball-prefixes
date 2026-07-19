package apnicautnum

import "testing"

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
