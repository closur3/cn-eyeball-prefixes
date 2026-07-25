package iputil

import "path/filepath"

var SourceURLs = map[string]string{
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

func SourceFilename(id string) string {
	switch id {
	case "iptoasn_ipv4", "iptoasn_ipv6":
		return id + ".tsv.gz"
	case "apnic_inetnum", "apnic_inet6num", "apnic_autnum", "apnic_organisation", "apnic_route", "riswhois_ipv4", "riswhois_ipv6":
		return id + ".gz"
	default:
		return id + ".txt"
	}
}

func SourcePath(dir, id string) string {
	return filepath.Join(dir, SourceFilename(id))
}
