package ipset6

import (
	"net/netip"
	"reflect"
	"testing"
)

func mustRange(prefix string) Range {
	row, err := FromPrefix(netip.MustParsePrefix(prefix))
	if err != nil {
		panic(err)
	}
	return row
}

func TestSetOperations(t *testing.T) {
	a := []Range{mustRange("2001:db8::/124")}
	b := []Range{mustRange("2001:db8::4/126"), mustRange("2001:db8::c/126")}
	want := []netip.Prefix{netip.MustParsePrefix("2001:db8::/126"), netip.MustParsePrefix("2001:db8::8/126")}
	if got := Prefixes(Subtract(a, b)); !reflect.DeepEqual(got, want) {
		t.Fatalf("subtract = %v, want %v", got, want)
	}
	if got := Prefixes(Intersect(a, b)); !reflect.DeepEqual(got, []netip.Prefix{netip.MustParsePrefix("2001:db8::4/126"), netip.MustParsePrefix("2001:db8::c/126")}) {
		t.Fatalf("intersect = %v", got)
	}
}

func TestMergeAndCounts(t *testing.T) {
	rows := []Range{mustRange("240e::/65"), mustRange("240e:0:0:0:8000::/65")}
	if got := Prefixes(rows); !reflect.DeepEqual(got, []netip.Prefix{netip.MustParsePrefix("240e::/64")}) {
		t.Fatalf("prefixes = %v", got)
	}
	if got := Slash64Equivalent(rows); got != "1.0000" {
		t.Fatalf("/64 equivalents = %s", got)
	}
}
