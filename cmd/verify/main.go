package main

import (
	"flag"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type span struct{ lo, hi uint32 }

var cloudSources = []string{
	"rezmoss_alibaba", "rezmoss_tencent", "rezmoss_huawei", "rezmoss_baidu",
	"ipdata_aliyun", "ipdata_tencent", "ipdata_huawei", "ipdata_ucloud", "ipdata_ksyun", "ipdata_baidu", "ipdata_jdcloud",
}

func n(a netip.Addr) uint32 {
	return uint32(a.As4()[0])<<24 | uint32(a.As4()[1])<<16 | uint32(a.As4()[2])<<8 | uint32(a.As4()[3])
}

func readCIDRs(path string, ordered bool) []span {
	b, e := os.ReadFile(path)
	if e != nil {
		panic(e)
	}
	var out []span
	var prev uint32
	first := true
	for _, s := range strings.Fields(string(b)) {
		p, e := netip.ParsePrefix(s)
		if e != nil || !p.Addr().Is4() || p.Addr() != p.Masked().Addr() {
			panic("invalid CIDR: " + path)
		}
		lo := n(p.Addr())
		hi := uint32(uint64(lo) + (uint64(1) << uint(32-p.Bits())) - 1)
		if ordered && !first && lo <= prev {
			panic("unordered or overlapping: " + path)
		}
		first = false
		prev = hi
		out = append(out, span{lo, hi})
	}
	return out
}

func merge(in []span) []span {
	sort.Slice(in, func(i, j int) bool { return in[i].lo < in[j].lo })
	var out []span
	for _, x := range in {
		if len(out) == 0 || (out[len(out)-1].hi != ^uint32(0) && x.lo > out[len(out)-1].hi+1) {
			out = append(out, x)
			continue
		}
		if x.hi > out[len(out)-1].hi {
			out[len(out)-1].hi = x.hi
		}
	}
	return out
}

func assertNoOverlap(a, b []span) {
	for i, j := 0, 0; i < len(a) && j < len(b); {
		if a[i].hi < b[j].lo {
			i++
		} else if b[j].hi < a[i].lo {
			j++
		} else {
			panic("cn.txt overlaps a cloud provider CIDR")
		}
	}
}

func assertContained(a, b []span) {
	for i, j := 0, 0; i < len(a); {
		for j < len(b) && b[j].hi < a[i].lo {
			j++
		}
		if j == len(b) || b[j].lo > a[i].lo || b[j].hi < a[i].hi {
			panic("cn.txt contains a CIDR absent from the origin-only China list")
		}
		i++
	}
}

func main() {
	data := flag.String("data", "data", "data directory")
	sources := flag.String("sources", "", "source directory")
	flag.Parse()
	if *sources == "" {
		panic("--sources is required")
	}

	files, e := filepath.Glob(filepath.Join(*data, "provinces", "*.txt"))
	if e != nil {
		panic(e)
	}
	files = append(files, filepath.Join(*data, "cn.txt"))
	if len(files) != 32 {
		panic("expected cn.txt plus exactly 31 provincial combined lists")
	}

	for _, f := range files {
		readCIDRs(f, true)
	}

	var cloudRanges []span
	for _, source := range cloudSources {
		cloudRanges = append(cloudRanges, readCIDRs(filepath.Join(*sources, source+".txt"), false)...)
	}
	cnRanges := readCIDRs(filepath.Join(*data, "cn.txt"), true)
	assertContained(cnRanges, readCIDRs(filepath.Join(*sources, "china.txt"), false))
	assertNoOverlap(cnRanges, merge(cloudRanges))
	fmt.Println("OK: lists are valid, ordered CIDR lists; cn.txt is contained in the origin-only China list and excludes all cloud provider CIDRs.")
}
