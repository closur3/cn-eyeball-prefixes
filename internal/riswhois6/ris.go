package riswhois6

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/netip"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Origin struct {
	ASN       string `json:"asn"`
	SeenPeers int    `json:"seen_peers"`
}

type Record struct {
	Prefix  netip.Prefix `json:"-"`
	Origins []Origin     `json:"origins"`
}

type Stats struct {
	Rows         int `json:"rows"`
	IPv6Prefixes int `json:"ipv6_prefixes"`
}

func ParseGzip(path string) ([]Record, Stats, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, Stats{}, err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return nil, Stats{}, err
	}
	defer z.Close()
	return Parse(z)
}

func Parse(r io.Reader) ([]Record, Stats, error) {
	byPrefix := map[string]*Record{}
	stats := Stats{}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		prefix, err := netip.ParsePrefix(fields[1])
		if err != nil || !prefix.Addr().Is6() || prefix.Addr().Is4In6() {
			continue
		}
		prefix = prefix.Masked()
		seenPeers, err := strconv.Atoi(fields[2])
		if err != nil || seenPeers < 1 {
			continue
		}
		asn := strings.TrimPrefix(strings.ToUpper(fields[0]), "AS")
		if _, err := strconv.ParseUint(asn, 10, 32); err != nil || asn == "0" {
			continue
		}
		stats.Rows++
		record := byPrefix[prefix.String()]
		if record == nil {
			record = &Record{Prefix: prefix}
			byPrefix[prefix.String()] = record
		}
		found := false
		for i := range record.Origins {
			if record.Origins[i].ASN == asn {
				if seenPeers > record.Origins[i].SeenPeers {
					record.Origins[i].SeenPeers = seenPeers
				}
				found = true
				break
			}
		}
		if !found {
			record.Origins = append(record.Origins, Origin{ASN: asn, SeenPeers: seenPeers})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, stats, err
	}
	if stats.Rows == 0 {
		return nil, stats, fmt.Errorf("input contains no IPv6 RISWhois rows")
	}
	out := make([]Record, 0, len(byPrefix))
	for _, record := range byPrefix {
		sort.Slice(record.Origins, func(i, j int) bool { return record.Origins[i].ASN < record.Origins[j].ASN })
		out = append(out, *record)
	}
	sort.Slice(out, func(i, j int) bool {
		if c := out[i].Prefix.Addr().Compare(out[j].Prefix.Addr()); c != 0 {
			return c < 0
		}
		return out[i].Prefix.Bits() < out[j].Prefix.Bits()
	})
	stats.IPv6Prefixes = len(out)
	return out, stats, nil
}
