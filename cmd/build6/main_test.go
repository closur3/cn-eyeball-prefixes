package main

import (
	"net/netip"
	"testing"

	"github.com/closur3/cn-eyeball-prefixes/internal/apnic6"
)

func TestRegistrationPurposeAcceptsOnlyTwoDescriptions(t *testing.T) {
	for _, test := range []struct {
		description string
		want        string
	}{
		{fixedDescription, "fixed"},
		{mobileDescription, "mobile"},
		{"BeiJing Telecom UserAddress", ""},
		{"Chinatelecom IPv6 address for network", ""},
		{"Chinatelecom IPv6 address for own platform", ""},
	} {
		got := registrationPurpose(apnic6.Record{Descriptions: []string{test.description}})
		if got != test.want {
			t.Fatalf("description %q: got %q, want %q", test.description, got, test.want)
		}
	}
}

func TestClassifyPrefixRequiresUniformPositiveCoverage(t *testing.T) {
	prefix := netip.MustParsePrefix("240e:400::/32")
	fixed := apnic6.Record{Descriptions: []string{fixedDescription}}
	mobile := apnic6.Record{Descriptions: []string{mobileDescription}}
	segments := []apnic6.Segment{{Lo: prefix.Addr(), Hi: lastAddress(prefix), Record: mobile}}
	if got, reason := classifyPrefix(prefix, buildAdmissionRanges(segments)); got != "mobile" || reason != "" {
		t.Fatalf("uniform mobile prefix: got=%q reason=%q", got, reason)
	}
	middle := netip.MustParseAddr("240e:400:8000::")
	segments = []apnic6.Segment{
		{Lo: prefix.Addr(), Hi: middle.Prev(), Record: mobile},
		{Lo: middle, Hi: lastAddress(prefix), Record: fixed},
	}
	if got, reason := classifyPrefix(prefix, buildAdmissionRanges(segments)); got != "" || reason != "mixed_access_purpose" {
		t.Fatalf("mixed prefix was admitted: got=%q reason=%q", got, reason)
	}
}

func TestNonAdmittedMoreSpecificRegistrationCreatesAdmissionGap(t *testing.T) {
	prefix := netip.MustParsePrefix("240e:300::/32")
	middle := netip.MustParseAddr("240e:300:8000::")
	fixed := apnic6.Record{Descriptions: []string{fixedDescription}}
	nonAccess := apnic6.Record{Descriptions: []string{"Chinatelecom IPv6 address for IDC"}}
	segments := []apnic6.Segment{
		{Lo: prefix.Addr(), Hi: middle.Prev(), Record: fixed},
		{Lo: middle, Hi: lastAddress(prefix), Record: nonAccess},
	}
	if got, reason := classifyPrefix(prefix, buildAdmissionRanges(segments)); got != "" || reason != "outside_admitted_registry" {
		t.Fatalf("prefix crossing a non-access registration was admitted: got=%q reason=%q", got, reason)
	}
}

func TestTelecomOriginRecognitionDoesNotApplyIPv4Exclusions(t *testing.T) {
	metadata := map[string]asMeta{
		"4134": {Country: "CN", Description: "China Telecom Backbone"},
		"4809": {Country: "CN", Description: "China Telecom Next Generation Carrier Network"},
	}
	if !allTelecomOrigins([]string{"4134", "4809"}, metadata) {
		t.Fatal("China Telecom transport origin rejected an explicitly labelled access prefix")
	}
}
