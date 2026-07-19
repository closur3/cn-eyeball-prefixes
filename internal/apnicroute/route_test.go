package apnicroute

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestParseFiltersIrrelevantObjectsButKeepsCounts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "route.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	z := gzip.NewWriter(f)
	_, err = z.Write([]byte("route: 10.0.0.0/8\norigin: AS64500\ndescr: relevant\n\nroute: 192.0.2.0/24\norigin: AS64501\ndescr: irrelevant\n\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	records, objects, relevantObjects, err := Parse(path, nil, func(lo, hi uint32) bool {
		return lo <= 0x0affffff && hi >= 0x0a000000
	})
	if err != nil {
		t.Fatal(err)
	}
	if objects != 2 || relevantObjects != 1 || len(records) != 1 || records[0].Prefix != "10.0.0.0/8" {
		t.Fatalf("objects=%d relevant=%d records=%#v", objects, relevantObjects, records)
	}
}
