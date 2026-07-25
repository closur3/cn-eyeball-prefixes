package listmanifest

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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

const SchemaVersion = 4

var sourceURLs = map[string]string{
	"china.txt":              "https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china.txt",
	"iptoasn_ipv4.tsv.gz":    "https://iptoasn.com/data/ip2asn-v4.tsv.gz",
	"iptoasn_ipv6.tsv.gz":    "https://iptoasn.com/data/ip2asn-v6.tsv.gz",
	"apnic_inetnum.gz":       "https://ftp.apnic.net/apnic/whois/apnic.db.inetnum.gz",
	"apnic_inet6num.gz":      "https://ftp.apnic.net/apnic/whois/apnic.db.inet6num.gz",
	"apnic_autnum.gz":        "https://ftp.apnic.net/apnic/whois/apnic.db.aut-num.gz",
	"apnic_organisation.gz":  "https://ftp.apnic.net/apnic/whois/apnic.db.organisation.gz",
	"apnic_route.gz":         "https://ftp.apnic.net/apnic/whois/apnic.db.route.gz",
	"riswhois_ipv4.gz":       "https://www.ris.ripe.net/dumps/riswhoisdump.IPv4.gz",
	"riswhois_ipv6.gz":       "https://www.ris.ripe.net/dumps/riswhoisdump.IPv6.gz",
	"ip2region_ipv4_source.txt": "https://raw.githubusercontent.com/lionsoul2014/ip2region/master/data/ipv4_source.txt",
	"ipdata_aliyun.txt":      "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/aliyun-cidr-ipv4.txt",
	"ipdata_tencent.txt":     "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/tencent-cidr-ipv4.txt",
	"ipdata_huawei.txt":      "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/huawei-cidr-ipv4.txt",
	"ipdata_ucloud.txt":      "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/ucloud-cidr-ipv4.txt",
	"ipdata_ksyun.txt":       "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/ksyun-cidr-ipv4.txt",
	"ipdata_baidu.txt":       "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/baidu-cidr-ipv4.txt",
	"ipdata_jdcloud.txt":     "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/jdcloud-cidr-ipv4.txt",
}

func ComputeSourceHashes(sourceDir string) map[string]SourceEntry {
	entries := make(map[string]SourceEntry)
	dh, err := os.Open(sourceDir)
	if err != nil {
		return nil
	}
	names, _ := dh.Readdirnames(-1)
	dh.Close()
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(sourceDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		h := sha256.Sum256(data)
		entry := SourceEntry{SHA256: hex.EncodeToString(h[:])}
		if url, ok := sourceURLs[name]; ok {
			entry.URL = url
		}
		entries[name] = entry
	}
	return entries
}

var operators = []string{"chinatelecom", "chinamobile", "chinaunicom"}

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
	paths := expectedPaths()
	if err := rejectUnexpectedLists(root, paths); err != nil {
		return false, err
	}

	files := make(map[string]FileEntry, len(paths))
	for _, rel := range paths {
		meta, err := inspect(filepath.Join(root, filepath.FromSlash(rel)), strings.HasPrefix(rel, "ipv4/"))
		if err != nil {
			return false, fmt.Errorf("%s: %w", rel, err)
		}
		files[rel] = meta
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
		if prefix.Addr().Is4() != ipv4 {
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
