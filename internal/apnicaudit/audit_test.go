package apnicaudit

import (
	"testing"

	"github.com/closur3/cn-operator-allowlist/internal/apnicinetnum"
	"github.com/closur3/cn-operator-allowlist/internal/operatorconfig"
)

func TestBuildCoversCIDRAndClassifiesMostSpecificRecords(t *testing.T) {
	classifier, err := operatorconfig.Load("../../config/operators.json", []string{"chinanet", "cmcc", "unicom"})
	if err != nil {
		t.Fatal(err)
	}
	segments := []apnicinetnum.Segment{
		{Lo: 0x0a000000, Hi: 0x0a00007f, Record: apnicinetnum.Record{Lo: 0x0a000000, Hi: 0x0a00007f, Descriptions: []string{"CHINANET Zhejiang Province Network"}}},
		{Lo: 0x0a000080, Hi: 0x0a0000bf, Record: apnicinetnum.Record{Lo: 0x0a000080, Hi: 0x0a0000bf, Descriptions: []string{"Example Technology Co., Ltd."}}},
		{Lo: 0x0a0000c0, Hi: 0x0a0000df, Record: apnicinetnum.Record{Lo: 0x0a0000c0, Hi: 0x0a0000df, Netnames: []string{"CAMPUS-POOL"}}},
		{Lo: 0x0a0000e0, Hi: 0x0a0000ef, Record: apnicinetnum.Record{Lo: 0x0a0000e0, Hi: 0x0a0000ef}, Match: apnicinetnum.Match{Reason: "explicit hosting range", MatchedBy: "test"}},
	}
	report, err := Build("test", []string{"10.0.0.0/24"}, map[string][]Range{"chinanet": {{Lo: 0x0a000000, Hi: 0x0a0000ff}}}, segments, classifier)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.AddressCount != 256 || report.Summary.RegistryCoveredAddressCount != 240 || report.Summary.StrongNonPublicSignalAddressCount != 16 {
		t.Fatalf("unexpected summary: %+v", report.Summary)
	}
	if len(report.CIDRs) != 1 || len(report.CIDRs[0].Facts) != 5 {
		t.Fatalf("unexpected CIDR facts: %+v", report.CIDRs)
	}
	want := []string{"operator_registration", "independent_legal_entity", "other_registration", "strong_non_public_signal", "unregistered"}
	for i, classification := range want {
		if report.CIDRs[0].Facts[i].Classification != classification {
			t.Fatalf("fact %d classification=%q want %q", i, report.CIDRs[0].Facts[i].Classification, classification)
		}
	}
}
