package main

import (
	"testing"

	"github.com/closur3/cn-eyeball-prefixes/internal/apnicinetnum"
	"github.com/closur3/cn-eyeball-prefixes/internal/operatorconfig"
	"github.com/closur3/cn-eyeball-prefixes/internal/riswhois"
)

func TestOverlapsSorted(t *testing.T) {
	rows := []span{{10, 19}, {30, 39}, {50, 50}}
	for _, tt := range []struct {
		lo, hi uint32
		want   bool
	}{
		{0, 9, false}, {9, 10, true}, {12, 14, true}, {20, 29, false},
		{39, 49, true}, {50, 50, true}, {51, 100, false},
	} {
		if got := overlapsSorted(rows, tt.lo, tt.hi); got != tt.want {
			t.Fatalf("overlapsSorted(%d, %d) = %v, want %v", tt.lo, tt.hi, got, tt.want)
		}
	}
	if overlapsSorted(nil, 0, 100) {
		t.Fatal("empty span set overlaps")
	}
}

func TestParentOperatorRegistrationAdmitsMoreSpecificCustomerRecord(t *testing.T) {
	classifier, err := operatorconfig.Load("../../config/operators.json", operators)
	if err != nil {
		t.Fatal(err)
	}
	records := []apnicinetnum.Record{
		{Lo: 0, Hi: 255, Descriptions: []string{"CHINANET Zhejiang province network"}},
		{Lo: 64, Hi: 127, Descriptions: []string{"Example customer assignment"}},
	}
	admitted := apnicOperatorAdmissionRanges(records, classifier)["chinatelecom"]
	if len(admitted) != 1 || admitted[0] != (span{0, 255}) {
		t.Fatalf("unexpected parent admission ranges: %#v", admitted)
	}
	segments := apnicinetnum.ResolveAll(records, func(apnicinetnum.Record) apnicinetnum.Match { return apnicinetnum.Match{} })
	conflicts := apnicOperatorConflictRanges(segments, classifier)
	if len(conflicts["chinatelecom"]) != 0 {
		t.Fatalf("independent customer label unexpectedly became an operator conflict: %#v", conflicts["chinatelecom"])
	}
}

func TestRelevantAPNICRecords(t *testing.T) {
	records := []apnicinetnum.Record{{Lo: 0, Hi: 9}, {Lo: 10, Hi: 19}, {Lo: 20, Hi: 29}}
	got := relevantAPNICRecords(records, []span{{12, 15}, {25, 25}})
	if len(got) != 2 || got[0].Lo != 10 || got[1].Lo != 20 {
		t.Fatalf("unexpected relevant records: %#v", got)
	}
}

func TestBGPConflictHealingKeepsTheRouteUnitAtomic(t *testing.T) {
	segments := []riswhois.Segment{{
		Lo: 0, Hi: 255,
		Record: riswhois.Record{Lo: 0, Hi: 255, Prefix: "0.0.0.0/24", Origins: []riswhois.Origin{{ASN: "4134", SeenPeers: 100}}},
	}}
	observed, eligible := bgpConflictHealingRanges(
		segments,
		map[string]string{"4134": "chinatelecom"},
		map[string][]span{"chinatelecom": {{0, 255}}},
		map[string][]span{"chinatelecom": {{0, 127}, {144, 255}}},
		map[string][]span{"chinatelecom": {{0, 255}}},
	)
	if len(observed) != 2 || observed[0] != (span{0, 127}) || observed[1] != (span{144, 255}) {
		t.Fatalf("unexpected RIS-observed retained ranges: %#v", observed)
	}
	if len(eligible) != 2 || eligible[0] != (span{0, 127}) || eligible[1] != (span{144, 255}) {
		t.Fatalf("same-operator parent did not make the retained BGP unit eligible for conflict healing: %#v", eligible)
	}
}

func TestBGPConflictHealingRequiresAPNICParent(t *testing.T) {
	segments := []riswhois.Segment{{
		Lo: 0, Hi: 255,
		Record: riswhois.Record{Lo: 0, Hi: 255, Prefix: "0.0.0.0/24", Origins: []riswhois.Origin{{ASN: "4134", SeenPeers: 100}}},
	}}
	observed, eligible := bgpConflictHealingRanges(
		segments,
		map[string]string{"4134": "chinatelecom"},
		map[string][]span{"chinatelecom": {{0, 255}}},
		map[string][]span{"chinatelecom": {{0, 255}}},
		map[string][]span{"chinatelecom": nil},
	)
	if len(observed) != 1 || observed[0] != (span{0, 255}) {
		t.Fatalf("RIS observation unexpectedly depended on APNIC parent evidence: %#v", observed)
	}
	if len(eligible) != 0 {
		t.Fatalf("conflict healing unexpectedly admitted a route without an APNIC operator parent: %#v", eligible)
	}
}

func TestConflictHealedAdmissionHealsOnlyEligibleOperatorRanges(t *testing.T) {
	hierarchical := map[string][]span{
		"chinatelecom": {{0, 63}},
		"chinamobile":     {{128, 191}},
		"chinaunicom":   nil,
	}
	eligible := map[string][]span{
		"chinatelecom": {{0, 127}},
		"chinamobile":     {{128, 255}},
		"chinaunicom":   nil,
	}
	got := conflictHealedAdmissionByOperator(hierarchical, []span{{64, 223}}, eligible)
	if len(got["chinatelecom"]) != 1 || got["chinatelecom"][0] != (span{0, 127}) {
		t.Fatalf("chinatelecom conflict healing did not heal its eligible conflict hole: %#v", got["chinatelecom"])
	}
	if len(got["chinamobile"]) != 1 || got["chinamobile"][0] != (span{128, 223}) {
		t.Fatalf("chinamobile conflict healing escaped its BGP-covered eligible range: %#v", got["chinamobile"])
	}
	if len(got["chinaunicom"]) != 0 {
		t.Fatalf("conflict healing invented an ineligible operator range: %#v", got["chinaunicom"])
	}
}
