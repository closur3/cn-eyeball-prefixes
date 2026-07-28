package publicverify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/listmanifest"
)

func TestVerifyAcceptsValidPublicLists(t *testing.T) {
	root := createValidPublicTree(t)
	if err := Verify(root); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyRejectsBrokenRelationships(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		mutate func(*testing.T, string)
	}{
		{
			name: "empty IPv4 operator",
			want: "ipv4/chinatelecom.txt contains no prefixes",
			mutate: func(t *testing.T, root string) {
				writeList(t, root, "ipv4/chinatelecom.txt")
			},
		},
		{
			name: "empty IPv6 operator",
			want: "ipv6/chinamobile.txt contains no prefixes",
			mutate: func(t *testing.T, root string) {
				writeList(t, root, "ipv6/chinamobile.txt")
			},
		},
		{
			name: "operator overlap",
			want: "ipv4 operator lists overlap",
			mutate: func(t *testing.T, root string) {
				writeList(t, root, "ipv4/chinamobile.txt", "10.0.0.0/25")
			},
		},
		{
			name: "operator union mismatch",
			want: "ipv4 operator union does not equal ipv4/cn.txt",
			mutate: func(t *testing.T, root string) {
				writeList(t, root, "ipv4/chinatelecom.txt", "10.0.0.0/27")
			},
		},
		{
			name: "province overlap",
			want: "ipv4 province lists overlap",
			mutate: func(t *testing.T, root string) {
				writeList(t, root, "ipv4/provinces/beijing.txt", "10.0.0.0/27")
			},
		},
		{
			name: "province outside cn",
			want: "ipv4/provinces/anhui.txt contains range",
			mutate: func(t *testing.T, root string) {
				writeList(t, root, "ipv4/provinces/anhui.txt", "192.0.2.0/24")
			},
		},
		{
			name: "IPv4 province coverage below threshold",
			want: "ipv4 province lists cover fewer than 90%",
			mutate: func(t *testing.T, root string) {
				for _, relative := range listmanifest.ExpectedPaths() {
					if strings.HasPrefix(relative, "ipv4/provinces/") {
						writeList(t, root, relative)
					}
				}
				writeList(t, root, "ipv4/provinces/anhui.txt", "10.0.0.0/27")
			},
		},
		{
			name: "IPv6 province union mismatch",
			want: "ipv6 province union does not equal ipv6/cn.txt",
			mutate: func(t *testing.T, root string) {
				writeList(t, root, "ipv6/provinces/anhui.txt", "2001:db8::/123")
			},
		},
		{
			name: "IPv4-mapped IPv6 is rejected",
			want: "wrong address family",
			mutate: func(t *testing.T, root string) {
				writeList(t, root, "ipv6/chinatelecom.txt", "::ffff:192.0.2.0/120")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := createValidPublicTree(t)
			test.mutate(t, root)
			err := Verify(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v, want error containing %q", err, test.want)
			}
		})
	}
}

func createValidPublicTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, relative := range listmanifest.ExpectedPaths() {
		writeList(t, root, relative)
	}

	writeList(t, root, "ipv4/cn.txt", "10.0.0.0/24")
	writeList(t, root, "ipv4/chinatelecom.txt", "10.0.0.0/26")
	writeList(t, root, "ipv4/chinamobile.txt", "10.0.0.64/26")
	writeList(t, root, "ipv4/chinaunicom.txt", "10.0.0.128/25")
	// IPv4 province coverage may be a strict subset of cn.txt, but the
	// generator contract requires at least 90% of addresses to be attributed.
	writeList(t, root, "ipv4/provinces/anhui.txt", "10.0.0.0/25")
	writeList(t, root, "ipv4/provinces/beijing.txt", "10.0.0.128/26")
	writeList(t, root, "ipv4/provinces/chongqing.txt", "10.0.0.192/27")
	writeList(t, root, "ipv4/provinces/fujian.txt", "10.0.0.224/29")

	writeList(t, root, "ipv6/cn.txt", "2001:db8::/122")
	writeList(t, root, "ipv6/chinatelecom.txt", "2001:db8::/124")
	writeList(t, root, "ipv6/chinamobile.txt", "2001:db8::10/124")
	writeList(t, root, "ipv6/chinaunicom.txt", "2001:db8::20/123")
	// Empty province files are valid, but their IPv6 union must still equal cn.
	writeList(t, root, "ipv6/provinces/anhui.txt", "2001:db8::/122")
	return root
}

func writeList(t *testing.T, root, relative string, prefixes ...string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	var data []byte
	if len(prefixes) != 0 {
		data = []byte(strings.Join(prefixes, "\n") + "\n")
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
