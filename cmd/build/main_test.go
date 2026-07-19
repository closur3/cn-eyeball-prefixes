package main

import (
	"reflect"
	"testing"

	"github.com/closur3/cn-operator-allowlist/internal/riswhois"
)

func TestRouteBoundaryCIDRsDoNotAggregateAdjacentAnnouncements(t *testing.T) {
	rows := []span{{lo: 0x67ec7400, hi: 0x67ec75ff}}
	segments := []riswhois.Segment{
		{Lo: 0x67ec7400, Hi: 0x67ec74ff, Record: riswhois.Record{Prefix: "103.236.116.0/24"}},
		{Lo: 0x67ec7500, Hi: 0x67ec75ff, Record: riswhois.Record{Prefix: "103.236.117.0/24"}},
	}
	want := []string{"103.236.116.0/24", "103.236.117.0/24"}
	if got := routeBoundaryCIDRs(rows, segments); !reflect.DeepEqual(got, want) {
		t.Fatalf("route boundaries = %#v, want %#v", got, want)
	}
}
