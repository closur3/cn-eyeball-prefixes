package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/bits"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

type span struct{ lo, hi uint32 }

type fileMeta struct {
	CIDRCount    int    `json:"cidr_count"`
	AddressCount uint64 `json:"address_count"`
	SHA256       string `json:"sha256"`
}

type sourceMeta struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type listMeta struct {
	Name string `json:"name"`
	Path string `json:"path"`
	fileMeta
}

type manifest struct {
	GeneratedAt string       `json:"generated_at"`
	Scope       string       `json:"scope"`
	Sources     []sourceMeta `json:"sources"`
	Lists       []listMeta   `json:"lists"`
}

type province struct {
	Name string
	Slug string
}

var operators = []string{"chinanet", "cmcc", "unicom"}
var cloudSources = []string{
	"rezmoss_alibaba", "rezmoss_tencent", "rezmoss_huawei", "rezmoss_baidu",
	"ipdata_aliyun", "ipdata_tencent", "ipdata_huawei", "ipdata_ucloud", "ipdata_ksyun", "ipdata_baidu", "ipdata_jdcloud",
}
var provinces = []province{
	{"北京市", "beijing"},
	{"天津市", "tianjin"},
	{"河北省", "hebei"},
	{"山西省", "shanxi"},
	{"内蒙古自治区", "neimenggu"},
	{"辽宁省", "liaoning"},
	{"吉林省", "jilin"},
	{"黑龙江省", "heilongjiang"},
	{"上海市", "shanghai"},
	{"江苏省", "jiangsu"},
	{"浙江省", "zhejiang"},
	{"安徽省", "anhui"},
	{"福建省", "fujian"},
	{"江西省", "jiangxi"},
	{"山东省", "shandong"},
	{"河南省", "henan"},
	{"湖北省", "hubei"},
	{"湖南省", "hunan"},
	{"广东省", "guangdong"},
	{"广西壮族自治区", "guangxi"},
	{"海南省", "hainan"},
	{"重庆市", "chongqing"},
	{"四川省", "sichuan"},
	{"贵州省", "guizhou"},
	{"云南省", "yunnan"},
	{"西藏自治区", "xizang"},
	{"陕西省", "shaanxi"},
	{"甘肃省", "gansu"},
	{"青海省", "qinghai"},
	{"宁夏回族自治区", "ningxia"},
	{"新疆维吾尔自治区", "xinjiang"},
}
var aliases = map[string]string{"北京": "北京市", "天津": "天津市", "上海": "上海市", "重庆": "重庆市", "内蒙古": "内蒙古自治区", "广西": "广西壮族自治区", "宁夏": "宁夏回族自治区", "新疆": "新疆维吾尔自治区", "西藏": "西藏自治区"}
var urls = map[string]string{
	"chinanet":              "https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/chinanet.txt",
	"cmcc":                  "https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/cmcc.txt",
	"unicom":                "https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/unicom.txt",
	"ip2region_ipv4_source": "https://raw.githubusercontent.com/lionsoul2014/ip2region/master/data/ipv4_source.txt",
	"rezmoss_alibaba":       "https://raw.githubusercontent.com/rezmoss/cloud-provider-ip-addresses/main/alibaba/alibaba_ips_merged_v4.txt",
	"rezmoss_tencent":       "https://raw.githubusercontent.com/rezmoss/cloud-provider-ip-addresses/main/tencent/tencent_ips_merged_v4.txt",
	"rezmoss_huawei":        "https://raw.githubusercontent.com/rezmoss/cloud-provider-ip-addresses/main/huawei/huawei_ips_merged_v4.txt",
	"rezmoss_baidu":         "https://raw.githubusercontent.com/rezmoss/cloud-provider-ip-addresses/main/baidu/baidu_ips_merged_v4.txt",
	"ipdata_aliyun":         "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/aliyun-cidr-ipv4.txt",
	"ipdata_tencent":        "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/tencent-cidr-ipv4.txt",
	"ipdata_huawei":         "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/huawei-cidr-ipv4.txt",
	"ipdata_ucloud":         "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/ucloud-cidr-ipv4.txt",
	"ipdata_ksyun":          "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/ksyun-cidr-ipv4.txt",
	"ipdata_baidu":          "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/baidu-cidr-ipv4.txt",
	"ipdata_jdcloud":        "https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/jdcloud-cidr-ipv4.txt",
}

func n(a netip.Addr) uint32 { return uint32(a.As4()[0])<<24|uint32(a.As4()[1])<<16|uint32(a.As4()[2])<<8|uint32(a.As4()[3]) }
func end(p netip.Prefix) uint32 { return uint32(uint64(n(p.Addr()))+(uint64(1)<<uint(32-p.Bits()))-1) }

func merge(in []span) []span {
	sort.Slice(in, func(i, j int) bool { return in[i].lo < in[j].lo })
	out := []span{}
	for _, x := range in {
		if len(out) == 0 {
			out = append(out, x)
			continue
		}
		last := &out[len(out)-1]
		if last.hi != ^uint32(0) && x.lo > last.hi+1 {
			out = append(out, x)
			continue
		}
		if x.hi > last.hi {
			last.hi = x.hi
		}
	}
	return out
}

func subtract(in, excluded []span) []span {
	in = merge(in)
	excluded = merge(excluded)
	var out []span
	j := 0
	for _, r := range in {
		for j < len(excluded) && excluded[j].hi < r.lo {
			j++
		}
		pos := r.lo
		covered := false
		for k := j; k < len(excluded) && excluded[k].lo <= r.hi; k++ {
			x := excluded[k]
			if x.hi < pos {
				continue
			}
			if x.lo > pos {
				out = append(out, span{pos, x.lo - 1})
			}
			if x.hi >= r.hi {
				covered = true
				break
			}
			pos = x.hi + 1
		}
		if !covered {
			out = append(out, span{pos, r.hi})
		}
	}
	return out
}

func cidrs(path string) ([]span, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var out []span
	for i, s := range strings.Fields(string(b)) {
		p, e := netip.ParsePrefix(s)
		if e != nil || !p.Addr().Is4() || p.Addr() != p.Masked().Addr() {
			return nil, fmt.Errorf("%s:%d", path, i+1)
		}
		out = append(out, span{n(p.Addr()), end(p)})
	}
	return merge(out), nil
}

func sha(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func write(path string, rows []span) (fileMeta, error) {
	rows = merge(rows)
	var lines []string
	var count uint64
	for _, r := range rows {
		for r.lo <= r.hi {
			remaining := uint64(r.hi) - uint64(r.lo) + 1
			align := bits.TrailingZeros32(r.lo)
			if r.lo == 0 {
				align = 32
			}
			sizeBits := align
			if max := bits.Len64(remaining) - 1; max < sizeBits {
				sizeBits = max
			}
			size := uint64(1) << uint(sizeBits)
			a := netip.AddrFrom4([4]byte{byte(r.lo >> 24), byte(r.lo >> 16), byte(r.lo >> 8), byte(r.lo)})
			lines = append(lines, netip.PrefixFrom(a, 32-sizeBits).String())
			count += size
			if size == remaining {
				break
			}
			r.lo += uint32(size)
		}
	}
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return fileMeta{}, e
	}
	if e := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); e != nil {
		return fileMeta{}, e
	}
	sum, e := sha(path)
	if e != nil {
		return fileMeta{}, e
	}
	return fileMeta{len(lines), count, sum}, nil
}

func readManifest(path string) (manifest, bool) {
	b, e := os.ReadFile(path)
	if e != nil {
		return manifest{}, false
	}
	var m manifest
	if json.Unmarshal(b, &m) != nil {
		return manifest{}, false
	}
	return m, true
}

func sameManifestContent(a, b manifest) bool {
	a.GeneratedAt = ""
	b.GeneratedAt = ""
	return reflect.DeepEqual(a, b)
}

func writeManifest(path string, m manifest) {
	b, e := json.MarshalIndent(m, "", "  ")
	if e != nil {
		panic(e)
	}
	if e := os.WriteFile(path, append(b, '\n'), 0644); e != nil {
		panic(e)
	}
}

func provinceSet() map[string]bool {
	out := map[string]bool{}
	for _, p := range provinces {
		out[p.Name] = true
	}
	return out
}

func main() {
	src := flag.String("sources", "", "source directory")
	out := flag.String("output", "data", "output directory")
	flag.Parse()
	if *src == "" {
		panic("--sources is required")
	}

	oldManifest, hasOldManifest := readManifest(filepath.Join(*out, "manifest.json"))

	ranges := map[string][]span{}
	for _, o := range operators {
		r, e := cidrs(filepath.Join(*src, o+".txt"))
		if e != nil {
			panic(e)
		}
		ranges[o] = r
	}
	var cloudRanges []span
	for _, source := range cloudSources {
		r, e := cidrs(filepath.Join(*src, source+".txt"))
		if e != nil {
			panic(e)
		}
		cloudRanges = append(cloudRanges, r...)
	}
	cloudRanges = merge(cloudRanges)
	for _, o := range operators {
		ranges[o] = subtract(ranges[o], cloudRanges)
	}

	by := map[string]map[string][]span{}
	for _, o := range operators {
		by[o] = map[string][]span{}
	}
	provinceNames := provinceSet()

	b, e := os.ReadFile(filepath.Join(*src, "ip2region_ipv4_source.txt"))
	if e != nil {
		panic(e)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		x := strings.Split(line, "|")
		if len(x) != 7 || x[2] != "中国" {
			continue
		}
		p := x[3]
		if a, ok := aliases[p]; ok {
			p = a
		}
		if !provinceNames[p] {
			continue
		}
		a, _ := netip.ParseAddr(x[0])
		z, _ := netip.ParseAddr(x[1])
		lo, hi := n(a), n(z)
		for _, o := range operators {
			for _, r := range ranges[o] {
				l, h := lo, hi
				if r.lo > l {
					l = r.lo
				}
				if r.hi < h {
					h = r.hi
				}
				if l <= h {
					by[o][p] = append(by[o][p], span{l, h})
				}
			}
		}
	}

	if e := os.RemoveAll(*out); e != nil {
		panic(e)
	}
	if e := os.MkdirAll(*out, 0755); e != nil {
		panic(e)
	}

	m := manifest{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Scope:       "ACL list; IPv4; mainland China; retains China Telecom, China Mobile, and China Unicom only after excluding CIDRs listed by either cloud-provider source: rezmoss (Alibaba, Tencent, Huawei, Baidu) or IP-Data (Alibaba, Tencent, Huawei, UCloud, Kingsoft, Baidu, JD Cloud); does not attempt to identify or exclude IDC addresses within operators' address space",
	}
	for _, o := range append(append(operators, "ip2region_ipv4_source"), cloudSources...) {
		sourcePath := filepath.Join(*src, o+".txt")
		sum, e := sha(sourcePath)
		if e != nil {
			panic(e)
		}
		m.Sources = append(m.Sources, sourceMeta{o, urls[o], sum})
	}

	var all []span
	for _, o := range operators {
		all = append(all, ranges[o]...)
	}
	cnMeta, e := write(filepath.Join(*out, "cn.txt"), all)
	if e != nil {
		panic(e)
	}
	m.Lists = append(m.Lists, listMeta{Name: "CN", Path: "cn.txt", fileMeta: cnMeta})

	for _, p := range provinces {
		var rows []span
		for _, o := range operators {
			rows = append(rows, by[o][p.Name]...)
		}
		path := filepath.Join("provinces", p.Slug+".txt")
		meta, e := write(filepath.Join(*out, path), rows)
		if e != nil {
			panic(e)
		}
		m.Lists = append(m.Lists, listMeta{Name: p.Name, Path: path, fileMeta: meta})
	}

	if hasOldManifest && sameManifestContent(oldManifest, m) {
		m.GeneratedAt = oldManifest.GeneratedAt
	}
	writeManifest(filepath.Join(*out, "manifest.json"), m)
}
