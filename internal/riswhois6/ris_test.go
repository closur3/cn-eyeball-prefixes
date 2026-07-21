package riswhois6

import (
	"strings"
	"testing"
)

func TestParsePreservesExactPrefixesAndMOAS(t *testing.T) {
	input := strings.NewReader(`
% comment
AS4134 240e:16:8000::/35 120
4134 240e:16:8000::/35 118
AS38283 240e:16:8000::/35 4
AS4134 240e:83:800::/37 99
AS4134 1.1.1.0/24 100
`)
	records, stats, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Rows != 4 || stats.IPv6Prefixes != 2 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if got := records[0].Prefix.String(); got != "240e:16:8000::/35" {
		t.Fatalf("prefix was altered: %s", got)
	}
	if len(records[0].Origins) != 2 || records[0].Origins[0].ASN != "38283" || records[0].Origins[1].ASN != "4134" {
		t.Fatalf("MOAS origins were not preserved: %#v", records[0].Origins)
	}
	if records[0].Origins[1].SeenPeers != 120 {
		t.Fatalf("maximum observer count not retained: %#v", records[0].Origins[1])
	}
}
