package releasecheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/listmanifest"
)

func TestCheckAcceptsUnchangedLists(t *testing.T) {
	current, candidate := createFixture(t)
	warnings, err := Check(testOptions(current, candidate))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}

func TestCheckWarnsOnLargeAddressChanges(t *testing.T) {
	tests := []struct {
		name      string
		relative  string
		current   string
		candidate string
		want      string
	}{
		{
			name:      "removal",
			relative:  filepath.Join("ipv4", "chinamobile.txt"),
			current:   "192.0.2.0/24\n",
			candidate: "192.0.2.0/25\n",
			want:      "removes 50.00%",
		},
		{
			name:      "addition",
			relative:  filepath.Join("ipv6", "chinaunicom.txt"),
			current:   "2001:db8::/33\n",
			candidate: "2001:db8::/32\n",
			want:      "adds 100.00%",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current, candidate := createFixture(t)
			if err := os.WriteFile(filepath.Join(current, test.relative), []byte(test.current), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(candidate, test.relative), []byte(test.candidate), 0644); err != nil {
				t.Fatal(err)
			}

			warnings, err := Check(testOptions(current, candidate))
			if err != nil {
				t.Fatal(err)
			}
			assertWarningContains(t, warnings, test.want)
		})
	}
}

func TestCheckWarnsOnPrefixGrowth(t *testing.T) {
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

	options := testOptions(current, candidate)
	options.MaxPrefixCountGrowthRatio = 3
	warnings, err := Check(options)
	if err != nil {
		t.Fatal(err)
	}
	assertWarningContains(t, warnings, "prefix count grows 4.00x")
}

func TestCheckWarnsWhenEmptyBaselineBecomesNonEmpty(t *testing.T) {
	current, candidate := createFixture(t)
	path := filepath.Join(current, "ipv6", "provinces", "xizang.txt")
	if err := os.WriteFile(path, nil, 0644); err != nil {
		t.Fatal(err)
	}

	warnings, err := Check(testOptions(current, candidate))
	if err != nil {
		t.Fatal(err)
	}
	assertWarningContains(t, warnings, "has no historical baseline")
}

func TestCheckRejectsEmptyRequiredList(t *testing.T) {
	current, candidate := createFixture(t)
	path := filepath.Join(candidate, "ipv6", "chinatelecom.txt")
	if err := os.WriteFile(path, nil, 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Check(testOptions(current, candidate)); err == nil ||
		!strings.Contains(err.Error(), "is empty") {
		t.Fatalf("got %v, want empty-list error", err)
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

func assertWarningContains(t *testing.T, warnings []string, substring string) {
	t.Helper()
	for _, warning := range warnings {
		if strings.Contains(warning, substring) {
			return
		}
	}
	t.Fatalf("no warning contains %q: %v", substring, warnings)
}
