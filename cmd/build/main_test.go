package main

import (
	"testing"

	"github.com/closur3/cn-operator-allowlist/internal/apnicinetnum"
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

func TestRelevantAPNICRecords(t *testing.T) {
	records := []apnicinetnum.Record{{Lo: 0, Hi: 9}, {Lo: 10, Hi: 19}, {Lo: 20, Hi: 29}}
	got := relevantAPNICRecords(records, []span{{12, 15}, {25, 25}})
	if len(got) != 2 || got[0].Lo != 10 || got[1].Lo != 20 {
		t.Fatalf("unexpected relevant records: %#v", got)
	}
}
