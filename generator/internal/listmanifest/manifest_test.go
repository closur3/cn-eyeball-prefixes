package listmanifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

	data, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != SchemaVersion {
		t.Fatalf("schema: got %d, want %d", manifest.SchemaVersion, SchemaVersion)
	}
	if manifest.GeneratedAt != "2026-07-25T08:00:00Z" {
		t.Fatalf("generated_at: got %q, want %q", manifest.GeneratedAt, "2026-07-25T08:00:00Z")
	}
	if manifest.Generator == nil || manifest.Generator.Commit != "abc123" || !manifest.Generator.Dirty {
		t.Fatal("unexpected generator field")
	}
	if manifest.Configs == nil || manifest.Configs["operators.json"].SHA256 != "cfg1hash" {
		t.Fatal("unexpected configs field")
	}
	if manifest.Sources == nil || manifest.Sources["china.txt"].SHA256 != "srchash" {
		t.Fatal("unexpected sources field")
	}
	if len(manifest.Files) != 70 {
		t.Fatalf("files: got %d, want 70", len(manifest.Files))
	}
	if manifest.ContentID == "" {
		t.Fatal("content_id must not be empty")
	}

	changed, err = Generate(
		root,
		time.Now(),
		&GeneratorInfo{Commit: "def456"},
		map[string]SourceEntry{},
		map[string]SourceEntry{},
	)
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
