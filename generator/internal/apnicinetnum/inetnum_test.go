package apnicinetnum

import (
	"reflect"
	"testing"
)

func TestResolveAllUsesMostSpecificRecord(t *testing.T) {
	records := []Record{
		testRecord(100, 199, "parent"),
		testRecord(120, 179, "opaque-child"),
	}

	segments := ResolveAll(records, testPolicyMatch)
	got := testSegmentAt(t, segments, 150)
	if got.Record.Netnames[0] != "opaque-child" {
		t.Fatalf("record at 150 = %q, want most-specific record %q", got.Record.Netnames[0], "opaque-child")
	}
}

func TestResolvePolicy(t *testing.T) {
	tests := []struct {
		name    string
		records []Record
		probes  []testPolicyProbe
	}{
		{
			name: "inherited parent survives opaque child",
			records: []Record{
				testRecord(100, 199, "soft-parent"),
				testRecord(120, 179, "opaque-child"),
			},
			probes: []testPolicyProbe{
				{
					at:           150,
					wantRecord:   "soft-parent",
					wantReason:   "operator-side parent",
					wantInherit:  true,
					wantExcluded: true,
				},
			},
		},
		{
			name: "explicit access child overrides inherited parent",
			records: []Record{
				testRecord(100, 199, "soft-parent"),
				testRecord(120, 179, "access-child"),
			},
			probes: []testPolicyProbe{
				{
					at:           110,
					wantRecord:   "soft-parent",
					wantReason:   "operator-side parent",
					wantInherit:  true,
					wantExcluded: true,
				},
				{
					at:         150,
					wantRecord: "access-child",
					wantAccess: true,
				},
			},
		},
		{
			name: "hard parent cannot be overridden by access child",
			records: []Record{
				testRecord(100, 199, "hard-parent"),
				testRecord(120, 179, "access-child"),
			},
			probes: []testPolicyProbe{
				{
					at:           150,
					wantRecord:   "hard-parent",
					wantReason:   "outside mainland",
					wantInherit:  true,
					wantHard:     true,
					wantExcluded: true,
				},
			},
		},
		{
			name: "neutral grandchild does not cancel access child",
			records: []Record{
				testRecord(100, 199, "soft-parent"),
				testRecord(120, 179, "access-child"),
				testRecord(140, 159, "neutral-grandchild"),
			},
			probes: []testPolicyProbe{
				{at: 150},
			},
		},
		{
			name: "more-specific local exclusion wins over access",
			records: []Record{
				testRecord(100, 199, "soft-parent"),
				testRecord(120, 179, "access-child"),
				testRecord(140, 159, "local-negative"),
			},
			probes: []testPolicyProbe{
				{
					at:         130,
					wantRecord: "access-child",
					wantAccess: true,
				},
				{
					at:           150,
					wantRecord:   "local-negative",
					wantReason:   "local infrastructure",
					wantExcluded: true,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			segments := ResolvePolicy(test.records, testPolicyMatch)
			for _, probe := range test.probes {
				got := testSegmentAt(t, segments, probe.at)
				if probe.wantRecord != "" && got.Record.Netnames[0] != probe.wantRecord {
					t.Errorf("record at %d = %q, want %q", probe.at, got.Record.Netnames[0], probe.wantRecord)
				}
				if got.Match.Reason != probe.wantReason {
					t.Errorf("reason at %d = %q, want %q", probe.at, got.Match.Reason, probe.wantReason)
				}
				if got.Match.Inherit != probe.wantInherit {
					t.Errorf("inherit at %d = %v, want %v", probe.at, got.Match.Inherit, probe.wantInherit)
				}
				if got.Match.Hard != probe.wantHard {
					t.Errorf("hard at %d = %v, want %v", probe.at, got.Match.Hard, probe.wantHard)
				}
				if got.Match.Access != probe.wantAccess {
					t.Errorf("access at %d = %v, want %v", probe.at, got.Match.Access, probe.wantAccess)
				}
				if excluded := got.Match.Reason != ""; excluded != probe.wantExcluded {
					t.Errorf("excluded at %d = %v, want %v", probe.at, excluded, probe.wantExcluded)
				}
			}
		})
	}
}

func TestResolvePolicyAndAllMatchesIndependentResolvers(t *testing.T) {
	records := []Record{
		testRecord(100, 999, "soft-parent"),
		testRecord(150, 949, "opaque-child"),
		testRecord(200, 899, "access-child"),
		testRecord(250, 849, "neutral-grandchild"),
		testRecord(300, 799, "hard-parent"),
		testRecord(350, 749, "access-child"),
		testRecord(400, 699, "local-negative"),
		testRecord(1000, 1099, "opaque-sibling"),
	}
	for i := range records {
		records[i].Descriptions = []string{records[i].Netnames[0] + " description"}
		records[i].Country = "CN"
	}
	classify := func(record Record) Match {
		match := testPolicyMatch(record)
		match.Category = "category:" + record.Netnames[0]
		match.MatchedBy = "netname=" + record.Netnames[0]
		return match
	}

	wantPolicy := ResolvePolicy(records, classify)
	wantAll := ResolveAll(records, func(Record) Match { return Match{} })
	classifyCalls := 0
	gotPolicy, gotAll := ResolvePolicyAndAll(records, func(record Record) Match {
		classifyCalls++
		return classify(record)
	})

	if classifyCalls != len(records) {
		t.Fatalf("classifier calls = %d, want %d", classifyCalls, len(records))
	}
	if !reflect.DeepEqual(gotPolicy, wantPolicy) {
		t.Fatalf("policy segments differ:\ngot:  %#v\nwant: %#v", gotPolicy, wantPolicy)
	}
	if !reflect.DeepEqual(gotAll, wantAll) {
		t.Fatalf("all segments differ:\ngot:  %#v\nwant: %#v", gotAll, wantAll)
	}
}

type testPolicyProbe struct {
	at           uint32
	wantRecord   string
	wantReason   string
	wantInherit  bool
	wantHard     bool
	wantAccess   bool
	wantExcluded bool
}

func testRecord(lo, hi uint32, name string) Record {
	return Record{Lo: lo, Hi: hi, Netnames: []string{name}}
}

func testPolicyMatch(record Record) Match {
	switch record.Netnames[0] {
	case "soft-parent":
		return Match{
			Reason:  "operator-side parent",
			Inherit: true,
		}
	case "hard-parent":
		return Match{
			Reason:  "outside mainland",
			Inherit: true,
			Hard:    true,
		}
	case "access-child":
		return Match{Access: true}
	case "local-negative":
		return Match{Reason: "local infrastructure"}
	default:
		return Match{}
	}
}

func testSegmentAt(t *testing.T, segments []Segment, at uint32) Segment {
	t.Helper()
	for _, segment := range segments {
		if segment.Lo <= at && at <= segment.Hi {
			return segment
		}
	}
	t.Fatalf("no segment covers %d", at)
	return Segment{}
}
