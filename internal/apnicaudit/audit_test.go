package apnicaudit

import (
	"strings"
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
		{Lo: 0x0a0000f0, Hi: 0x0a0000f7, Record: apnicinetnum.Record{Lo: 0x0a0000f0, Hi: 0x0a0000f7, Descriptions: []string{"China Unicom Zhejiang Province Network"}}},
	}
	report, err := Build("test", []string{"10.0.0.0/24"}, map[string][]Range{"chinanet": {{Lo: 0x0a000000, Hi: 0x0a0000ff}}}, segments, classifier)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.AddressCount != 256 || report.Summary.RegistryCoveredAddressCount != 248 || report.Summary.StrongNonPublicSignalAddressCount != 16 {
		t.Fatalf("unexpected summary: %+v", report.Summary)
	}
	if len(report.CIDRs) != 1 || len(report.CIDRs[0].Facts) != 6 {
		t.Fatalf("unexpected CIDR facts: %+v", report.CIDRs)
	}
	want := []string{"operator_registration", "independent_legal_entity", "other_registration", "strong_non_public_signal", "operator_registration_conflict", "unregistered"}
	for i, classification := range want {
		if report.CIDRs[0].Facts[i].Classification != classification {
			t.Fatalf("fact %d classification=%q want %q", i, report.CIDRs[0].Facts[i].Classification, classification)
		}
	}
}

func TestRenderMarkdownIncludesSummaryAndReviewEvidence(t *testing.T) {
	report := Report{
		Summary: Summary{
			CIDRCount: 1, FactCount: 2, AddressCount: 256, RegistryCoveredAddressCount: 256,
			RegistryCoveragePercent: 100, StrongNonPublicSignalAddressCount: 16,
			Categories: []CategorySummary{
				{Classification: "independent_legal_entity", FactCount: 1, AddressCount: 240, AddressPercent: 93.75},
				{Classification: "strong_non_public_signal", FactCount: 1, AddressCount: 16, AddressPercent: 6.25},
			},
		},
		CIDRs: []CIDRRecord{{
			CIDR: "10.0.0.0/24", AddressCount: 256,
			Facts: []Fact{
				{Start: "10.0.0.0", End: "10.0.0.239", AddressCount: 240, Operator: "chinanet", Classification: "independent_legal_entity", Registry: &Registry{Descriptions: []string{"Example Technology Co., Ltd."}, Range: "10.0.0.0 - 10.0.0.239"}},
				{Start: "10.0.0.240", End: "10.0.0.255", AddressCount: 16, Operator: "chinanet", Classification: "strong_non_public_signal", Reason: "explicit hosting range", Registry: &Registry{Netnames: []string{"EXAMPLE-IDC"}, Range: "10.0.0.240 - 10.0.0.255"}},
			},
		}},
	}
	markdown := RenderMarkdown(report, "zhejiang-apnic.json.gz")
	for _, want := range []string{"# 浙江 IPv4 APNIC 登记事实审计", "全国分层准入规则", "| `strong_non_public_signal` | 1 | 16 | 6.2500% |", "10.0.0.240–10.0.0.255", "Example Technology Co., Ltd.", "zhejiang-apnic.json.gz"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("Markdown report does not contain %q:\n%s", want, markdown)
		}
	}
}

func TestRenderMarkdownExplainsRelaxedBGPAdditions(t *testing.T) {
	report := Report{
		Scope: "Nationwide relaxed-BGP additions versus current dev",
		Summary: Summary{AddressCount: 256, Categories: []CategorySummary{{Classification: "independent_legal_entity", FactCount: 1, AddressCount: 256, AddressPercent: 100}}},
	}
	markdown := RenderMarkdown(report, "bgp-relaxed-added-apnic.json.gz")
	for _, want := range []string{"# 宽松 BGP 新增地址 APNIC 登记事实审计", "取消 APNIC 正向准入", "宽松 BGP 方案会纳入", "尚未进入正式 ACL"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("relaxed-BGP Markdown report does not contain %q:\n%s", want, markdown)
		}
	}
	if strings.Contains(markdown, "# 浙江 IPv4") {
		t.Fatalf("relaxed-BGP Markdown report retained the Zhejiang title:\n%s", markdown)
	}
}
