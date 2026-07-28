package listmanifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/iputil"
)

func TestGenerateIsDeterministic(t *testing.T) {
	root := t.TempDir()
	for _, rel := range expectedPaths() {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		line := "2001:db8::/32\n"
		if strings.HasPrefix(rel, "ipv4/") {
			line = "192.0.2.0/24\n"
		}
		if err := os.WriteFile(path, []byte(line), 0644); err != nil {
			t.Fatal(err)
		}
	}

	changed, err := Generate(root, time.Time{}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("first generation should write the manifest")
	}
	first, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	changed, err = Generate(root, time.Time{}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("identical lists must not rewrite the manifest")
	}
	second, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("manifest bytes changed without a list change")
	}

	var manifest Manifest
	if err := json.Unmarshal(first, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != SchemaVersion || len(manifest.Files) != 70 {
		t.Fatalf("unexpected manifest: schema=%d files=%d", manifest.SchemaVersion, len(manifest.Files))
	}
	if manifest.GeneratedAt != "" {
		t.Fatal("expected no generated_at with zero time")
	}
	if manifest.Generator != nil {
		t.Fatal("expected nil generator")
	}
	if manifest.Configs != nil {
		t.Fatal("expected nil configs")
	}
	if manifest.Sources != nil {
		t.Fatal("expected nil sources")
	}
}

func TestGenerateWithMetadata(t *testing.T) {
	root := t.TempDir()
	for _, rel := range expectedPaths() {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		line := "2001:db8::/32\n"
		if strings.HasPrefix(rel, "ipv4/") {
			line = "192.0.2.0/24\n"
		}
		if err := os.WriteFile(path, []byte(line), 0644); err != nil {
			t.Fatal(err)
		}
	}

	now := time.Date(2026, 7, 25, 8, 0, 0, 0, time.UTC)
	gen := &GeneratorInfo{Commit: "abc123", Dirty: true}
	configs := map[string]SourceEntry{
		"operators.json": {SHA256: "cfg1hash"},
	}
	sources := map[string]SourceEntry{
		"china.txt": {SHA256: "srchash", URL: "https://example.com/china.txt"},
	}

	changed, err := Generate(root, now, gen, configs, sources)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("first generation should write the manifest")
	}

	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != SchemaVersion {
		t.Fatalf("schema: got %d, want %d", m.SchemaVersion, SchemaVersion)
	}
	if m.GeneratedAt != "2026-07-25T08:00:00Z" {
		t.Fatalf("generated_at: got %q, want %q", m.GeneratedAt, "2026-07-25T08:00:00Z")
	}
	if m.Generator == nil || m.Generator.Commit != "abc123" || !m.Generator.Dirty {
		t.Fatal("unexpected generator field")
	}
	if m.Configs == nil || m.Configs["operators.json"].SHA256 != "cfg1hash" {
		t.Fatal("unexpected configs field")
	}
	if m.Sources == nil || m.Sources["china.txt"].SHA256 != "srchash" {
		t.Fatal("unexpected sources field")
	}
	if len(m.Files) != 70 {
		t.Fatalf("files: got %d, want 70", len(m.Files))
	}
	if m.ContentID == "" {
		t.Fatal("content_id must not be empty")
	}

	// Second call with same lists but different metadata must not rewrite.
	changed, err = Generate(root, time.Now(), &GeneratorInfo{Commit: "def456"}, map[string]SourceEntry{}, map[string]SourceEntry{})
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("metadata-only changes must not rewrite the manifest")
	}
}

func TestInspectRejectsOverlap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "overlap.txt")
	if err := os.WriteFile(path, []byte("192.0.2.0/24\n192.0.2.0/25\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := inspect(path, true); err == nil {
		t.Fatal("expected overlapping CIDRs to be rejected")
	}
}

func TestInspectRejectsIPv4MappedIPv6(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mapped.txt")
	if err := os.WriteFile(path, []byte("::ffff:192.0.2.0/120\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := inspect(path, false); err == nil ||
		!strings.Contains(err.Error(), "wrong address family") {
		t.Fatalf("got %v, want wrong-address-family error", err)
	}
}

func TestVerifyAcceptsCompleteRelease(t *testing.T) {
	root := createListTree(t)
	if _, err := Generate(root, time.Time{}, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := Verify(root); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyRejectsTamperedList(t *testing.T) {
	root := createListTree(t)
	if _, err := Generate(root, time.Time{}, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "ipv4", "cn.txt")
	if err := os.WriteFile(path, []byte("192.0.2.0/25\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(root); err == nil || !strings.Contains(err.Error(), "metadata mismatch for ipv4/cn.txt") {
		t.Fatalf("got %v, want metadata mismatch", err)
	}
}

func TestVerifyRejectsTamperedContentID(t *testing.T) {
	root := createListTree(t)
	if _, err := Generate(root, time.Time{}, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.ContentID = "sha256:tampered"
	data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(root); err == nil || !strings.Contains(err.Error(), "content_id") {
		t.Fatalf("got %v, want content_id mismatch", err)
	}
}

func TestVerifyRequiresCompletePublicationProvenance(t *testing.T) {
	options := VerifyOptions{RequirePublicationProvenance: true}
	root := createListTree(t)
	generatePublicationManifest(t, root, strings.Repeat("a", 40))
	if err := VerifyWithOptions(root, options); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		mutate    func(*Manifest)
		wantError string
	}{
		{
			name: "missing generated time",
			mutate: func(manifest *Manifest) {
				manifest.GeneratedAt = ""
			},
			wantError: "generated_at",
		},
		{
			name: "malformed commit",
			mutate: func(manifest *Manifest) {
				manifest.Generator.Commit = "abc123"
			},
			wantError: "40 lowercase hex digits",
		},
		{
			name: "dirty generator",
			mutate: func(manifest *Manifest) {
				manifest.Generator.Dirty = true
			},
			wantError: "generator must not be dirty",
		},
		{
			name: "missing config",
			mutate: func(manifest *Manifest) {
				delete(manifest.Configs, "operators.json")
			},
			wantError: "configs contains 1 entries",
		},
		{
			name: "malformed config hash",
			mutate: func(manifest *Manifest) {
				manifest.Configs["operators.json"] = SourceEntry{SHA256: "ABC"}
			},
			wantError: "sha256 must be 64 lowercase hex digits",
		},
		{
			name: "missing source",
			mutate: func(manifest *Manifest) {
				delete(manifest.Sources, "china.txt")
			},
			wantError: "sources contains",
		},
		{
			name: "source URL changed",
			mutate: func(manifest *Manifest) {
				entry := manifest.Sources["china.txt"]
				entry.URL = "https://example.invalid/china.txt"
				manifest.Sources["china.txt"] = entry
			},
			wantError: "sources china.txt URL",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := createListTree(t)
			generatePublicationManifest(t, root, strings.Repeat("a", 40))
			path := filepath.Join(root, "manifest.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var manifest Manifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				t.Fatal(err)
			}
			test.mutate(&manifest)
			data, err = json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			if err := VerifyWithOptions(root, options); err == nil ||
				!strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("got %v, want error containing %q", err, test.wantError)
			}
		})
	}
}

func TestVerifyBindsChangedPublicationToWorkflowCommit(t *testing.T) {
	current := createListTree(t)
	oldCommit := strings.Repeat("a", 40)
	newCommit := strings.Repeat("b", 40)
	generatePublicationManifest(t, current, oldCommit)

	unchanged := createListTree(t)
	generatePublicationManifest(t, unchanged, oldCommit)
	options := VerifyOptions{
		RequirePublicationProvenance: true,
		CurrentRoot:                  current,
		ExpectedNewGeneratorCommit:   newCommit,
	}
	if err := VerifyWithOptions(unchanged, options); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(unchanged, "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.GeneratedAt = "2026-07-26T08:00:00Z"
	data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyWithOptions(unchanged, options); err == nil ||
		!strings.Contains(err.Error(), "metadata changed without a public-list content change") {
		t.Fatalf("got %v, want unchanged-content metadata error", err)
	}

	changed := createListTree(t)
	if err := os.WriteFile(
		filepath.Join(changed, "ipv4", "provinces", "anhui.txt"),
		[]byte("198.51.100.0/24\n"),
		0644,
	); err != nil {
		t.Fatal(err)
	}
	generatePublicationManifest(t, changed, newCommit)
	if err := VerifyWithOptions(changed, options); err != nil {
		t.Fatal(err)
	}

	wrongCommit := createListTree(t)
	if err := os.WriteFile(
		filepath.Join(wrongCommit, "ipv4", "provinces", "anhui.txt"),
		[]byte("198.51.100.0/24\n"),
		0644,
	); err != nil {
		t.Fatal(err)
	}
	generatePublicationManifest(t, wrongCommit, strings.Repeat("c", 40))
	if err := VerifyWithOptions(wrongCommit, options); err == nil ||
		!strings.Contains(err.Error(), "want current workflow commit") {
		t.Fatalf("got %v, want generator-commit binding error", err)
	}
}

func TestGenerateRejectsEmptyRequiredList(t *testing.T) {
	root := createListTree(t)
	if err := os.WriteFile(filepath.Join(root, "ipv6", "chinaunicom.txt"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(root, time.Time{}, nil, nil, nil); err == nil ||
		!strings.Contains(err.Error(), "required public list is empty") {
		t.Fatalf("got %v, want empty required-list error", err)
	}
}

func createListTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, rel := range expectedPaths() {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		line := "2001:db8::/32\n"
		if strings.HasPrefix(rel, "ipv4/") {
			line = "192.0.2.0/24\n"
		}
		if err := os.WriteFile(path, []byte(line), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func generatePublicationManifest(t *testing.T, root, commit string) {
	t.Helper()
	hash := strings.Repeat("1", 64)
	configs := map[string]SourceEntry{
		"ipv6-province-prefixes.json": {SHA256: hash},
		"operators.json":              {SHA256: hash},
	}
	sources := make(map[string]SourceEntry, len(iputil.SourceURLs))
	for name, url := range iputil.SourceURLs {
		sources[name] = SourceEntry{SHA256: hash, URL: url}
	}
	changed, err := Generate(
		root,
		time.Date(2026, 7, 25, 8, 0, 0, 0, time.UTC),
		&GeneratorInfo{Commit: commit, Dirty: false},
		configs,
		sources,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected publication manifest to be generated")
	}
}
