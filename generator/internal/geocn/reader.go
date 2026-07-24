package geocn

import (
	"fmt"
	"net/netip"

	"github.com/oschwald/maxminddb-golang/v2"
)

type Record struct {
	DivisionCode uint32 `maxminddb:"division_code"`
	ISP          string `maxminddb:"isp"`
}

func ProvinceRanges(path string) (map[string][]netip.Prefix, error) {
	reader, err := maxminddb.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open GeoCN mmdb: %w", err)
	}
	defer reader.Close()

	out := make(map[string][]netip.Prefix)
	for result := range reader.Networks() {
		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("iterate GeoCN network: %w", err)
		}
		var record Record
		if err := result.Decode(&record); err != nil {
			return nil, fmt.Errorf("decode GeoCN record: %w", err)
		}
		prefix := result.Prefix()
		if !prefix.Addr().Is4() || prefix.Addr().Is4In6() {
			continue
		}
		if record.DivisionCode == 0 {
			continue
		}
		province := ProvinceFromCode(record.DivisionCode)
		if province == "" {
			continue
		}
		out[province] = append(out[province], prefix.Masked())
	}
	return out, nil
}

var codeToProvince = map[uint32]string{
	11: "北京市",
	12: "天津市",
	13: "河北省",
	14: "山西省",
	15: "内蒙古自治区",
	21: "辽宁省",
	22: "吉林省",
	23: "黑龙江省",
	31: "上海市",
	32: "江苏省",
	33: "浙江省",
	34: "安徽省",
	35: "福建省",
	36: "江西省",
	37: "山东省",
	41: "河南省",
	42: "湖北省",
	43: "湖南省",
	44: "广东省",
	45: "广西壮族自治区",
	46: "海南省",
	50: "重庆市",
	51: "四川省",
	52: "贵州省",
	53: "云南省",
	54: "西藏自治区",
	61: "陕西省",
	62: "甘肃省",
	63: "青海省",
	64: "宁夏回族自治区",
	65: "新疆维吾尔自治区",
}

func ProvinceFromCode(code uint32) string {
	return codeToProvince[code/10000]
}
