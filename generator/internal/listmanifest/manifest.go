package listmanifest

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/iputil"
)

// SourceEntry records the identity of an upstream data source or local config
// file used to generate the public lists.  Comparing these across commits
// makes it possible to tell which upstream change drove a list change.
type SourceEntry struct {
	SHA256 string `json:"sha256"`
	URL    string `json:"url,omitempty"`
}

// GeneratorInfo describes the version of the generator program that built the
// lists.  Comparing commits tells you whether a list change was caused by a
// code change rather than a data-source or config change.
type GeneratorInfo struct {
	Commit string `json:"commit"`
	Dirty  bool   `json:"dirty"`
}

// FileEntry records the identity of a single output list file.
type FileEntry struct {
	PrefixCount int    `json:"prefix_count"`
	SHA256      string `json:"sha256"`
}

type Manifest struct {
	SchemaVersion int                    `json:"schema_version"`
	ContentID     string                 `json:"content_id"`
	GeneratedAt   string                 `json:"generated_at,omitempty"`
	Generator     *GeneratorInfo         `json:"generator,omitempty"`
	Configs       map[string]SourceEntry `json:"configs,omitempty"`
	Sources       map[string]SourceEntry `json:"sources,omitempty"`
	Files         map[string]FileEntry   `json:"files"`
}

type VerifyOptions struct {
	RequirePublicationProvenance bool
	CurrentRoot                  string
	ExpectedNewGeneratorCommit   string
}

const SchemaVersion = 4

func ComputeSourceHashes(sourceDir string) (map[string]SourceEntry, error) {
	entries := make(map[string]SourceEntry)
	dh, err := os.Open(sourceDir)
	if err != nil {
		return nil, err
	}
	names, err := dh.Readdirnames(-1)
	if closeErr := dh.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(sourceDir, name)
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		hash := sha256.New()
		if _, err = io.Copy(hash, file); err == nil {
			err = file.Close()
		} else {
			_ = file.Close()
		}
		if err != nil {
			return nil, err
		}
		entry := SourceEntry{SHA256: hex.EncodeToString(hash.Sum(nil))}
		url, hasURLMetadata := iputil.SourceURLs[name]
		if !hasURLMetadata {
			return nil, fmt.Errorf("source %s has no declared URL metadata", name)
		}
		entry.URL = url
		entries[name] = entry
	}
	return entries, nil
}

var operators = iputil.Operators

var provinces = []string{
	"anhui",
	"beijing",
	"chongqing",
	"fujian",
	"gansu",
	"guangdong",
	"guangxi",
	"guizhou",
	"hainan",
	"hebei",
	"heilongjiang",
	"henan",
	"hubei",
	"hunan",
	"jiangsu",
	"jiangxi",
	"jilin",
	"liaoning",
	"neimenggu",
	"ningxia",
	"qinghai",
	"shaanxi",
	"shandong",
	"shanghai",
	"shanxi",
	"sichuan",
	"tianjin",
	"xinjiang",
	"xizang",
	"yunnan",
	"zhejiang",
}

// Generate validates the complete public list contract and writes a
// deterministic manifest.  It returns false without touching the file when the
// list content is already current (contentID unchanged), even if metadata
// (sources, configs, generator) have changed.  This keeps manifest updates
// aligned with actual list changes so that metadata-only churn never produces
// a spurious commit.
//
// Optional fields:
//   - generatedAt: zero time skips the generated_at field
//   - gen:         nil skips the generator field
//   - configs:     nil skips the configs field
//   - sources:     nil skips the sources field
func Generate(root string, generatedAt time.Time, gen *GeneratorInfo, configs, sources map[string]SourceEntry) (bool, error) {
	paths, files, err := inspectTree(root, true)
	if err != nil {
		return false, err
	}

	id := contentID(paths, files)

	manifest := Manifest{
		SchemaVersion: SchemaVersion,
		ContentID:     id,
		Files:         files,
		Configs:       configs,
		Sources:       sources,
	}
	if !generatedAt.IsZero() {
		manifest.GeneratedAt = generatedAt.UTC().Format(time.RFC3339Nano)
	}
	if gen != nil {
		manifest.Generator = gen
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return false, err
	}
	data = append(data, '\n')

	path := filepath.Join(root, "manifest.json")
	current, err := os.ReadFile(path)
	if err == nil {
		var prev Manifest
		if json.Unmarshal(current, &prev) == nil && prev.ContentID == id {
			return false, nil
		}
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return false, err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return false, err
	}
	return true, nil
}

// ComputeContentID validates a complete public list tree and returns the same
// deterministic content identifier used by manifest.json.
func ComputeContentID(root string) (string, error) {
	paths, files, err := inspectTree(root, false)
	if err != nil {
		return "", err
	}
	return contentID(paths, files), nil
}

// Verify checks that the manifest and public list tree form one complete,
// self-consistent release. It does not rewrite any files.
func Verify(root string) error {
	return VerifyWithOptions(root, VerifyOptions{})
}

// VerifyWithOptions applies Verify plus publication-specific provenance
// requirements.
func VerifyWithOptions(root string, options VerifyOptions) error {
	paths, files, err := inspectTree(root, true)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		return err
	}
	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return fmt.Errorf("manifest.json: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return fmt.Errorf("manifest.json: %w", err)
	}
	if manifest.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"manifest.json: schema_version is %d, want %d",
			manifest.SchemaVersion,
			SchemaVersion,
		)
	}
	if options.RequirePublicationProvenance ||
		options.CurrentRoot != "" ||
		options.ExpectedNewGeneratorCommit != "" {
		if err := validatePublicationProvenance(&manifest); err != nil {
			return err
		}
	}
	if len(manifest.Files) != len(paths) {
		return fmt.Errorf(
			"manifest.json: contains %d file entries, want %d",
			len(manifest.Files),
			len(paths),
		)
	}

	expected := make(map[string]bool, len(paths))
	for _, rel := range paths {
		expected[rel] = true
		recorded, ok := manifest.Files[rel]
		if !ok {
			return fmt.Errorf("manifest.json: missing file entry %s", rel)
		}
		actual := files[rel]
		if recorded != actual {
			return fmt.Errorf(
				"manifest.json: metadata mismatch for %s: got count=%d sha256=%s, want count=%d sha256=%s",
				rel,
				recorded.PrefixCount,
				recorded.SHA256,
				actual.PrefixCount,
				actual.SHA256,
			)
		}
	}
	for rel := range manifest.Files {
		if !expected[rel] {
			return fmt.Errorf("manifest.json: unexpected file entry %s", rel)
		}
	}

	actualID := contentID(paths, files)
	if manifest.ContentID != actualID {
		return fmt.Errorf(
			"manifest.json: content_id is %q, want %q",
			manifest.ContentID,
			actualID,
		)
	}
	if options.CurrentRoot != "" || options.ExpectedNewGeneratorCommit != "" {
		if options.CurrentRoot == "" || options.ExpectedNewGeneratorCommit == "" {
			return fmt.Errorf(
				"publication comparison requires both current root and expected new generator commit",
			)
		}
		if !isLowerHex(options.ExpectedNewGeneratorCommit, 40) {
			return fmt.Errorf("expected new generator commit must be 40 lowercase hex digits")
		}
		if err := Verify(options.CurrentRoot); err != nil {
			return fmt.Errorf("current release: %w", err)
		}
		currentData, err := os.ReadFile(filepath.Join(options.CurrentRoot, "manifest.json"))
		if err != nil {
			return fmt.Errorf("current release manifest: %w", err)
		}
		var current Manifest
		if err := json.Unmarshal(currentData, &current); err != nil {
			return fmt.Errorf("current release manifest: %w", err)
		}
		if manifest.ContentID == current.ContentID {
			if !bytes.Equal(data, currentData) {
				return fmt.Errorf(
					"manifest.json: metadata changed without a public-list content change",
				)
			}
		} else if manifest.Generator.Commit != options.ExpectedNewGeneratorCommit {
			return fmt.Errorf(
				"manifest.json: generator commit is %s, want current workflow commit %s",
				manifest.Generator.Commit,
				options.ExpectedNewGeneratorCommit,
			)
		}
	}
	return nil
}

func validatePublicationProvenance(manifest *Manifest) error {
	generatedAt, err := time.Parse(time.RFC3339Nano, manifest.GeneratedAt)
	if err != nil || generatedAt.UTC().Format(time.RFC3339Nano) != manifest.GeneratedAt {
		return fmt.Errorf("manifest.json: generated_at must be canonical UTC RFC3339")
	}
	switch {
	case manifest.Generator == nil:
		return fmt.Errorf("manifest.json: generator metadata is required")
	case !isLowerHex(manifest.Generator.Commit, 40):
		return fmt.Errorf("manifest.json: generator commit must be 40 lowercase hex digits")
	case manifest.Generator.Dirty:
		return fmt.Errorf("manifest.json: generator must not be dirty")
	}

	expectedConfigs := map[string]string{
		"ipv6-province-prefixes.json": "",
		"operators.json":              "",
	}
	if err := validateSourceEntries("configs", manifest.Configs, expectedConfigs); err != nil {
		return err
	}
	if err := validateSourceEntries("sources", manifest.Sources, iputil.SourceURLs); err != nil {
		return err
	}
	return nil
}

func validateSourceEntries(
	label string,
	entries map[string]SourceEntry,
	expectedURLs map[string]string,
) error {
	if len(entries) != len(expectedURLs) {
		return fmt.Errorf(
			"manifest.json: %s contains %d entries, want %d",
			label,
			len(entries),
			len(expectedURLs),
		)
	}
	for name, expectedURL := range expectedURLs {
		entry, ok := entries[name]
		if !ok {
			return fmt.Errorf("manifest.json: %s is missing %s", label, name)
		}
		if !isLowerHex(entry.SHA256, 64) {
			return fmt.Errorf(
				"manifest.json: %s %s sha256 must be 64 lowercase hex digits",
				label,
				name,
			)
		}
		if entry.URL != expectedURL {
			return fmt.Errorf(
				"manifest.json: %s %s URL is %q, want %q",
				label,
				name,
				entry.URL,
				expectedURL,
			)
		}
	}
	return nil
}

func isLowerHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func inspectTree(root string, requireCoreLists bool) ([]string, map[string]FileEntry, error) {
	paths := expectedPaths()
	if err := rejectUnexpectedLists(root, paths); err != nil {
		return nil, nil, err
	}
	files := make(map[string]FileEntry, len(paths))
	for _, rel := range paths {
		meta, err := inspect(filepath.Join(root, filepath.FromSlash(rel)), strings.HasPrefix(rel, "ipv4/"))
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", rel, err)
		}
		if requireCoreLists && isRequiredList(rel) && meta.PrefixCount == 0 {
			return nil, nil, fmt.Errorf("%s: required public list is empty", rel)
		}
		files[rel] = meta
	}
	return paths, files, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("contains multiple JSON values")
		}
		return err
	}
	return nil
}

func expectedPaths() []string {
	var paths []string
	for _, family := range []string{"ipv4", "ipv6"} {
		paths = append(paths, family+"/cn.txt")
		for _, operator := range operators {
			paths = append(paths, family+"/"+operator+".txt")
		}
		for _, province := range provinces {
			paths = append(paths, family+"/provinces/"+province+".txt")
		}
	}
	sort.Strings(paths)
	return paths
}

// ExpectedPaths returns the complete, sorted public list contract. Callers
// receive a fresh slice and may modify it without affecting later calls.
func ExpectedPaths() []string {
	return expectedPaths()
}

func isRequiredList(relative string) bool {
	return strings.Count(filepath.ToSlash(relative), "/") == 1
}

func rejectUnexpectedLists(root string, expected []string) error {
	want := make(map[string]bool, len(expected))
	for _, rel := range expected {
		want[rel] = true
	}
	var unexpected []string
	for _, family := range []string{"ipv4", "ipv6"} {
		base := filepath.Join(root, family)
		err := filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".txt") {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if !want[rel] {
				unexpected = append(unexpected, rel)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	if len(unexpected) != 0 {
		sort.Strings(unexpected)
		return fmt.Errorf("unexpected public list files: %s", strings.Join(unexpected, ", "))
	}
	return nil
}

func inspect(path string, ipv4 bool) (FileEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileEntry{}, err
	}
	sum := sha256.Sum256(data)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var previous netip.Prefix
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return FileEntry{}, fmt.Errorf("blank line at line %d", count+1)
		}
		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return FileEntry{}, fmt.Errorf("invalid CIDR at line %d: %w", count+1, err)
		}
		if prefix != prefix.Masked() {
			return FileEntry{}, fmt.Errorf("non-canonical CIDR at line %d: %s", count+1, line)
		}
		if line != prefix.String() {
			return FileEntry{}, fmt.Errorf("non-canonical CIDR text at line %d: use %s", count+1, prefix)
		}
		if ipv4 != prefix.Addr().Is4() ||
			(!ipv4 && (!prefix.Addr().Is6() || prefix.Addr().Is4In6())) {
			return FileEntry{}, fmt.Errorf("wrong address family at line %d: %s", count+1, line)
		}
		if count != 0 {
			if compare(previous, prefix) >= 0 {
				return FileEntry{}, fmt.Errorf("CIDRs are not strictly sorted at line %d", count+1)
			}
			if previous.Contains(prefix.Addr()) {
				return FileEntry{}, fmt.Errorf("CIDRs overlap at line %d: %s contains %s", count+1, previous, prefix)
			}
		}
		previous = prefix
		count++
	}
	if err := scanner.Err(); err != nil {
		return FileEntry{}, err
	}
	return FileEntry{PrefixCount: count, SHA256: hex.EncodeToString(sum[:])}, nil
}

func compare(a, b netip.Prefix) int {
	if cmp := a.Addr().Compare(b.Addr()); cmp != 0 {
		return cmp
	}
	switch {
	case a.Bits() < b.Bits():
		return -1
	case a.Bits() > b.Bits():
		return 1
	default:
		return 0
	}
}

func contentID(paths []string, files map[string]FileEntry) string {
	hash := sha256.New()
	fmt.Fprintf(hash, "schema=%d\n", SchemaVersion)
	for _, path := range paths {
		meta := files[path]
		fmt.Fprintf(hash, "%s\x00%d\x00%s\n", path, meta.PrefixCount, meta.SHA256)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}
