package releasecheck

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/listmanifest"
)

func TestCheckAcceptsUnchangedLists(t *testing.T) {
	current, candidate := createFixture(t)
	report, err := Check(testOptions(current, candidate))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", report.Warnings)
	}
	if len(report.Files) != 70 {
		t.Fatalf("files: got %d, want 70", len(report.Files))
	}
	for _, entry := range report.Files {
		if entry.JaccardSimilarity != 1 {
			t.Fatalf("%s jaccard: got %f, want 1", entry.Path, entry.JaccardSimilarity)
		}
	}
}

func TestCheckRejectsEmptyRequiredList(t *testing.T) {
	current, candidate := createFixture(t)
	path := filepath.Join(candidate, "ipv6", "chinatelecom.txt")
	if err := os.WriteFile(path, nil, 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Check(testOptions(current, candidate)); err == nil ||
		!strings.Contains(err.Error(), "empty") {
		t.Fatalf("got %v, want empty-candidate error", err)
	}
}

func TestCheckAllowsUnchangedEmptyProvince(t *testing.T) {
	current, candidate := createFixture(t)
	for _, root := range []string{current, candidate} {
		path := filepath.Join(root, "ipv6", "provinces", "xizang.txt")
		if err := os.WriteFile(path, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	report, err := Check(testOptions(current, candidate))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", report.Warnings)
	}
	for _, entry := range report.Files {
		if entry.Path == "ipv6/provinces/xizang.txt" && entry.JaccardSimilarity != 1 {
			t.Fatalf("empty province jaccard: got %f, want 1", entry.JaccardSimilarity)
		}
	}
}

func TestCheckWarnsOnLargeAddressRemoval(t *testing.T) {
	current, candidate := createFixture(t)
	path := filepath.Join(candidate, "ipv4", "chinamobile.txt")
	if err := os.WriteFile(path, []byte("192.0.2.0/25\n"), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := Check(testOptions(current, candidate))
	if err != nil {
		t.Fatal(err)
	}
	assertWarningContains(t, report, "ipv4/chinamobile.txt removes 0.500000")
}

func TestCheckWarnsOnLargeAddressAddition(t *testing.T) {
	current, candidate := createFixture(t)
	currentPath := filepath.Join(current, "ipv6", "chinaunicom.txt")
	candidatePath := filepath.Join(candidate, "ipv6", "chinaunicom.txt")
	if err := os.WriteFile(currentPath, []byte("2001:db8::/33\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(candidatePath, []byte("2001:db8::/32\n"), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := Check(testOptions(current, candidate))
	if err != nil {
		t.Fatal(err)
	}
	assertWarningContains(t, report, "ipv6/chinaunicom.txt adds 1.000000")
}

func TestCheckWritesWarningsReport(t *testing.T) {
	current, candidate := createFixture(t)
	path := filepath.Join(candidate, "ipv4", "provinces", "xizang.txt")
	if err := os.WriteFile(path, []byte("192.0.2.0/23\n"), 0644); err != nil {
		t.Fatal(err)
	}

	options := testOptions(current, candidate)
	options.ReportPath = filepath.Join(t.TempDir(), "release-change-report.json")
	if _, err := Check(options); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(options.ReportPath)
	if err != nil {
		t.Fatal(err)
	}
	var persisted Report
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.SchemaVersion != ReportSchemaVersion {
		t.Fatalf("schema version: got %d, want %d", persisted.SchemaVersion, ReportSchemaVersion)
	}
	assertWarningContains(t, &persisted, "ipv4/provinces/xizang.txt adds 1.000000")
}

func TestCheckAllowsExactAddressChangeLimits(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		candidate string
	}{
		{
			name: "removal",
			current: strings.Join([]string{
				"192.0.2.0/29",
				"192.0.2.8/31",
			}, "\n") + "\n",
			candidate: strings.Join([]string{
				"192.0.2.0/29",
				"192.0.2.8/32",
			}, "\n") + "\n",
		},
		{
			name: "addition",
			current: strings.Join([]string{
				"192.0.2.0/29",
				"192.0.2.8/31",
			}, "\n") + "\n",
			candidate: strings.Join([]string{
				"192.0.2.0/29",
				"192.0.2.8/31",
				"192.0.2.10/32",
			}, "\n") + "\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current, candidate := createFixture(t)
			relative := filepath.Join("ipv4", "provinces", "xizang.txt")
			if err := os.WriteFile(filepath.Join(current, relative), []byte(test.current), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(candidate, relative), []byte(test.candidate), 0644); err != nil {
				t.Fatal(err)
			}
			options := testOptions(current, candidate)
			options.MaxRemovedAddressRatio = 0.10
			options.MaxAddedAddressRatio = 0.10
			report, err := Check(options)
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Warnings) != 0 {
				t.Fatalf("unexpected warnings at exact threshold: %v", report.Warnings)
			}
		})
	}
}

func TestCheckAllowsReviewedAdditionAboveCurrentSize(t *testing.T) {
	current, candidate := createFixture(t)
	currentPath := filepath.Join(current, "ipv6", "chinaunicom.txt")
	candidatePath := filepath.Join(candidate, "ipv6", "chinaunicom.txt")
	if err := os.WriteFile(currentPath, []byte("2001:db8::/34\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(candidatePath, []byte("2001:db8::/32\n"), 0644); err != nil {
		t.Fatal(err)
	}

	options := testOptions(current, candidate)
	options.MaxAddedAddressRatio = 3
	report, err := Check(options)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", report.Warnings)
	}
}

func TestCheckWarnsOnEmptyBaseline(t *testing.T) {
	current, candidate := createFixture(t)
	path := filepath.Join(current, "ipv6", "provinces", "xizang.txt")
	if err := os.WriteFile(path, nil, 0644); err != nil {
		t.Fatal(err)
	}

	options := testOptions(current, candidate)
	report, err := Check(options)
	if err != nil {
		t.Fatal(err)
	}
	assertWarningContains(t, report, "has no historical address baseline")

	var found bool
	for _, entry := range report.Files {
		if entry.Path != "ipv6/provinces/xizang.txt" {
			continue
		}
		found = true
		if !entry.BaselineEmpty {
			t.Fatal("expected baseline_empty to be true")
		}
		if entry.AddedAddressRatio != nil || entry.PrefixCountGrowthRatio != nil {
			t.Fatal("zero-baseline ratios must be null")
		}
	}
	if !found {
		t.Fatal("missing xizang report entry")
	}
	if !strings.HasPrefix(report.CandidateContentID, "sha256:") ||
		len(report.CandidateContentID) != len("sha256:")+64 {
		t.Fatalf("invalid candidate content ID: %q", report.CandidateContentID)
	}
}

func TestCheckAllowsExactPrefixGrowthLimit(t *testing.T) {
	current, candidate := createFixture(t)
	path := filepath.Join(candidate, "ipv4", "cn.txt")
	data := strings.Join([]string{
		"192.0.2.0/26",
		"192.0.2.64/26",
		"192.0.2.128/26",
		"192.0.2.192/26",
	}, "\n") + "\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := Check(testOptions(current, candidate))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", report.Warnings)
	}
}

func TestCheckWarnsOnPrefixExplosionWithSameCoverage(t *testing.T) {
	current, candidate := createFixture(t)
	var lines []string
	for i := 0; i < 256; i++ {
		lines = append(lines, "192.0.2."+itoa(i)+"/32")
	}
	path := filepath.Join(candidate, "ipv4", "cn.txt")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := Check(testOptions(current, candidate))
	if err != nil {
		t.Fatal(err)
	}
	assertWarningContains(t, report, "ipv4/cn.txt grows from 1 to 256 prefixes")
}

func TestCheckRejectsNonFiniteThresholds(t *testing.T) {
	current, candidate := createFixture(t)
	tests := []struct {
		name   string
		mutate func(*Options)
	}{
		{
			name: "nan removed",
			mutate: func(options *Options) {
				options.MaxRemovedAddressRatio = math.NaN()
			},
		},
		{
			name: "positive infinity added",
			mutate: func(options *Options) {
				options.MaxAddedAddressRatio = math.Inf(1)
			},
		},
		{
			name: "negative infinity prefix growth",
			mutate: func(options *Options) {
				options.MaxPrefixCountGrowthRatio = math.Inf(-1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := testOptions(current, candidate)
			test.mutate(&options)
			if _, err := Check(options); err == nil || !strings.Contains(err.Error(), "must be finite") {
				t.Fatalf("got %v, want finite-threshold error", err)
			}
		})
	}
}

func testOptions(current, candidate string) Options {
	return Options{
		CurrentRoot:               current,
		CandidateRoot:             candidate,
		MaxRemovedAddressRatio:    0.20,
		MaxAddedAddressRatio:      0.20,
		MaxPrefixCountGrowthRatio: 4,
	}
}

func createFixture(t *testing.T) (string, string) {
	t.Helper()
	current := t.TempDir()
	candidate := t.TempDir()
	for _, root := range []string{current, candidate} {
		for _, relative := range listmanifest.ExpectedPaths() {
			path := filepath.Join(root, filepath.FromSlash(relative))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			line := "2001:db8::/32\n"
			if strings.HasPrefix(relative, "ipv4/") {
				line = "192.0.2.0/24\n"
			}
			if err := os.WriteFile(path, []byte(line), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return current, candidate
}

func assertWarningContains(t *testing.T, report *Report, substring string) {
	t.Helper()
	for _, warning := range report.Warnings {
		if strings.Contains(warning, substring) {
			return
		}
	}
	t.Fatalf("no warning contains %q: %v", substring, report.Warnings)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [3]byte
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[position:])
}
