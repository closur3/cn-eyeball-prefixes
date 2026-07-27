package ipv4build

import (
	"reflect"
	"testing"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/apnicinetnum"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/operatorconfig"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/riswhois"
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

func TestMoreSpecificAccessRegistrationOverridesBackboneParent(t *testing.T) {
	classifier, err := operatorconfig.Load("../../config/operators.json", operators)
	if err != nil {
		t.Fatal(err)
	}
	records := []apnicinetnum.Record{
		{
			Lo:           0,
			Hi:           255,
			Netnames:     []string{"CHINANET-BB"},
			Descriptions: []string{"CHINANET backbone network", "Data Communication Division"},
		},
		{
			Lo:           64,
			Hi:           127,
			Netnames:     []string{"CHINANET-ZJ"},
			Descriptions: []string{"ordinary residential broadband IP pool"},
		},
	}
	segments := apnicinetnum.ResolvePolicy(records, func(record apnicinetnum.Record) apnicinetnum.Match {
		text := apnicinetnum.SearchText(record)
		result := classifier.ClassifyAPNICInetnum(text)
		access := classifier.ClassifyAPNICAccess(text)
		return apnicinetnum.Match{
			Reason: result.Reason, MatchedBy: result.MatchedBy,
			Inherit: result.Inherit, Hard: result.Hard, Access: access.Reason != "",
		}
	})
	if len(segments) != 3 {
		t.Fatalf("unexpected resolved segment count: %#v", segments)
	}
	if segments[0].Match.Reason == "" || segments[2].Match.Reason == "" {
		t.Fatalf("backbone parent was not excluded outside its child registration: %#v", segments)
	}
	if segments[1].Lo != 64 || segments[1].Hi != 127 || segments[1].Match.Reason != "" {
		t.Fatalf("most-specific ordinary access child inherited the backbone exclusion: %#v", segments[1])
	}
}

func TestUnknownChildCannotOverrideBackboneParent(t *testing.T) {
	classifier, err := operatorconfig.Load("../../config/operators.json", operators)
	if err != nil {
		t.Fatal(err)
	}
	records := []apnicinetnum.Record{
		{
			Lo:           0,
			Hi:           255,
			Netnames:     []string{"CHINANET-BB"},
			Descriptions: []string{"CHINANET backbone network", "Data Communication Division"},
		},
		{
			Lo:           64,
			Hi:           127,
			Netnames:     []string{"FSKWC"},
			Descriptions: []string{"FSKWC NET"},
		},
	}
	segments := apnicinetnum.ResolvePolicy(records, func(record apnicinetnum.Record) apnicinetnum.Match {
		result := classifier.ClassifyAPNICInetnum(apnicinetnum.SearchText(record))
		return apnicinetnum.Match{
			Reason: result.Reason, MatchedBy: result.MatchedBy,
			Inherit: result.Inherit, Hard: result.Hard,
		}
	})
	for _, segment := range segments {
		if segment.Lo <= 64 && 64 <= segment.Hi {
			if segment.Match.Reason == "" || len(segment.Record.Netnames) == 0 || segment.Record.Netnames[0] != "CHINANET-BB" {
				t.Fatalf("opaque FSKWC child overrode its backbone parent: %#v", segment)
			}
			return
		}
	}
	t.Fatal("resolved policy did not cover the FSKWC child")
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
		"chinamobile":  {{128, 191}},
		"chinaunicom":  nil,
	}
	eligible := map[string][]span{
		"chinatelecom": {{0, 127}},
		"chinamobile":  {{128, 255}},
		"chinaunicom":  nil,
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

func TestReviewedBackboneExclusionsRequireCorroboratingEvidence(t *testing.T) {
	classifier, err := operatorconfig.Load("../../config/operators.json", operators)
	if err != nil {
		t.Fatal(err)
	}
	parent := apnicinetnum.Record{
		Lo:           0,
		Hi:           255,
		Netnames:     []string{"CHINANET-ZJ"},
		Descriptions: []string{"CHINANET Zhejiang province network"},
	}
	policy := operatorconfig.BackboneIPv4Prefix{
		CIDR:         "0.0.0.0/24",
		Operator:     "chinatelecom",
		Reason:       "reviewed test backbone",
		EvidenceURLs: []string{"https://example.com/evidence"},
	}

	tests := []struct {
		name             string
		records          []apnicinetnum.Record
		originByOperator map[string][]span
		want             []span
	}{
		{
			name: "unknown child does not override reviewed backbone parent",
			records: []apnicinetnum.Record{
				parent,
				{Lo: 64, Hi: 127, Netnames: []string{"FSKWC"}, Descriptions: []string{"FSKWC NET"}},
			},
			originByOperator: map[string][]span{"chinatelecom": {{0, 255}}},
			want:             []span{{0, 255}},
		},
		{
			name: "explicit more-specific access child remains included",
			records: []apnicinetnum.Record{
				parent,
				{Lo: 64, Hi: 127, Netnames: []string{"RESIDENTIAL"}, Descriptions: []string{"ordinary residential broadband IP pool"}},
			},
			originByOperator: map[string][]span{"chinatelecom": {{0, 255}}},
			want:             []span{{0, 63}, {128, 255}},
		},
		{
			name: "more-specific registration for another operator remains included",
			records: []apnicinetnum.Record{
				parent,
				{Lo: 64, Hi: 127, Netnames: []string{"CMNET-ZHEJIANG"}, Descriptions: []string{"China Mobile Group Zhejiang Co., Ltd."}},
			},
			originByOperator: map[string][]span{"chinatelecom": {{0, 255}}},
			want:             []span{{0, 63}, {128, 255}},
		},
		{
			name:             "missing same-operator origin disables policy",
			records:          []apnicinetnum.Record{parent},
			originByOperator: map[string][]span{"chinamobile": {{0, 255}}},
		},
		{
			name: "missing same-operator APNIC parent disables policy",
			records: []apnicinetnum.Record{
				{Lo: 0, Hi: 255, Netnames: []string{"EXAMPLE"}, Descriptions: []string{"Example allocation"}},
			},
			originByOperator: map[string][]span{"chinatelecom": {{0, 255}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments := apnicinetnum.ResolveAll(tt.records, func(apnicinetnum.Record) apnicinetnum.Match {
				return apnicinetnum.Match{}
			})
			index := reviewedBackboneIndexForTest([]operatorconfig.BackboneIPv4Prefix{policy}, tt.records, segments, classifier)
			got, _ := reviewedBackboneExclusions(
				[]operatorconfig.BackboneIPv4Prefix{policy},
				tt.originByOperator,
				[]span{{0, 255}},
				index,
				nil,
				nil,
			)
			assertReviewedBackboneSpans(t, got, tt.want)
		})
	}
}

func TestReviewedBackboneIndexMatchesLegacyAPNICRanges(t *testing.T) {
	classifier, err := operatorconfig.Load("../../config/operators.json", operators)
	if err != nil {
		t.Fatal(err)
	}
	policies := []operatorconfig.BackboneIPv4Prefix{{
		CIDR:         "0.0.0.0/24",
		Operator:     "chinatelecom",
		Reason:       "reviewed test backbone",
		EvidenceURLs: []string{"https://example.com/evidence"},
	}}
	records := []apnicinetnum.Record{
		{Lo: 0, Hi: 255, Netnames: []string{"CHINANET-ZJ"}, Descriptions: []string{"CHINANET Zhejiang province network"}},
		{Lo: 64, Hi: 127, Netnames: []string{"CMNET-ZHEJIANG"}, Descriptions: []string{"China Mobile Group Zhejiang Co., Ltd."}},
		{Lo: 96, Hi: 111, Netnames: []string{"FSKWC"}, Descriptions: []string{"FSKWC NET"}},
	}
	segments := apnicinetnum.ResolveAll(records, func(apnicinetnum.Record) apnicinetnum.Match {
		return apnicinetnum.Match{}
	})
	index := reviewedBackboneIndexForTest(policies, records, segments, classifier)
	legacyAdmission := apnicOperatorAdmissionRanges(records, classifier)
	legacyConflicts := apnicOperatorConflictRanges(segments, classifier)
	for _, operator := range operators {
		assertReviewedBackboneSpans(t, index.admissionRanges[operator], legacyAdmission[operator])
		assertReviewedBackboneSpans(t, index.conflictRanges[operator], legacyConflicts[operator])
	}
}

func TestReviewedBackboneIndexPolicyScopedParentsMatchLegacyResolver(t *testing.T) {
	classifier, err := operatorconfig.Load("../../config/operators.json", operators)
	if err != nil {
		t.Fatal(err)
	}
	policies := []operatorconfig.BackboneIPv4Prefix{
		{CIDR: "0.0.0.0/24", Operator: "chinatelecom", Reason: "first policy", EvidenceURLs: []string{"https://example.com/first"}},
		{CIDR: "0.0.2.0/24", Operator: "chinatelecom", Reason: "second policy", EvidenceURLs: []string{"https://example.com/second"}},
	}
	records := []apnicinetnum.Record{
		{Lo: 0, Hi: 1023, Netnames: []string{"CHINANET-ZJ"}, Descriptions: []string{"CHINANET Zhejiang province network"}},
		{Lo: 64, Hi: 127, Netnames: []string{"CHINANET-HZ"}, Descriptions: []string{"CHINANET Hangzhou network"}},
		{Lo: 576, Hi: 639, Netnames: []string{"CHINANET-NB"}, Descriptions: []string{"CHINANET Ningbo network"}},
		{Lo: 2048, Hi: 2303, Netnames: []string{"CHINANET-JS"}, Descriptions: []string{"CHINANET Jiangsu province network"}},
	}
	segments := apnicinetnum.ResolveAll(records, func(apnicinetnum.Record) apnicinetnum.Match {
		return apnicinetnum.Match{}
	})
	index := reviewedBackboneIndexForTest(policies, records, segments, classifier)

	var legacyRecords []apnicinetnum.Record
	for _, record := range records {
		if classifier.ClassifyAPNICRegistrant(apnicinetnum.SearchText(record)).Operator == "chinatelecom" {
			legacyRecords = append(legacyRecords, record)
		}
	}
	legacy := apnicinetnum.ResolveAll(legacyRecords, func(apnicinetnum.Record) apnicinetnum.Match {
		return apnicinetnum.Match{}
	})
	for _, tt := range []struct {
		cidr   string
		lo, hi uint32
	}{
		{cidr: "0.0.0.0/24", lo: 0, hi: 255},
		{cidr: "0.0.2.0/24", lo: 512, hi: 767},
	} {
		got := trimReviewedBackboneParentSegments(index.parentSegments[tt.cidr], tt.lo, tt.hi)
		want := trimReviewedBackboneParentSegments(legacy, tt.lo, tt.hi)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("policy-scoped parent segments for %s = %#v, want %#v", tt.cidr, got, want)
		}
	}
}

func reviewedBackboneIndexForTest(policies []operatorconfig.BackboneIPv4Prefix, records []apnicinetnum.Record, segments []apnicinetnum.Segment, classifier *operatorconfig.Classifier) *reviewedBackboneIndex {
	index := newReviewedBackboneIndex(policies)
	for _, record := range records {
		text := apnicinetnum.SearchText(record)
		_, access := classifier.ClassifyAPNICPolicy(text)
		index.Observe(record, classifier.Classify("0", text).Operator, access)
	}
	index.Finalize(segments)
	return index
}

func trimReviewedBackboneParentSegments(segments []apnicinetnum.Segment, lo, hi uint32) []apnicinetnum.Segment {
	var out []apnicinetnum.Segment
	for _, segment := range segments {
		if segment.Hi < lo || segment.Lo > hi {
			continue
		}
		segment.Lo = max(segment.Lo, lo)
		segment.Hi = min(segment.Hi, hi)
		out = append(out, segment)
	}
	return out
}

func assertReviewedBackboneSpans(t *testing.T, got, want []span) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("reviewed backbone exclusions = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("reviewed backbone exclusions = %#v, want %#v", got, want)
		}
	}
}
