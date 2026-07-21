package operatorconfig

import "testing"

func TestRepositoryOperatorBoundary(t *testing.T) {
	classifier, err := Load("../../config/operators.json", []string{"chinanet", "cmcc", "unicom"})
	if err != nil {
		t.Fatalf("load repository operator config: %v", err)
	}

	tests := []struct {
		name            string
		asn             string
		description     string
		operator        string
		excluded        bool
		exclusionSource string
	}{
		{
			name:            "China Telecom CN2 dedicated premium backbone",
			asn:             "4809",
			description:     "CHINATELECOM-CORE-WAN-CN2 China Telecom Next Generation Carrier Network",
			operator:        "chinanet",
			excluded:        true,
			exclusionSource: "explicit_policy",
		},
		{
			name:            "China Unicom CUII dedicated premium backbone",
			asn:             "9929",
			description:     "CUII CHINA UNICOM Industrial Internet Backbone",
			operator:        "unicom",
			excluded:        true,
			exclusionSource: "explicit_policy",
		},
		{
			name:        "China Telecom ordinary access origins remain eligible",
			asn:         "4134",
			description: "CHINANET-BACKBONE No.31 Jin-rong Street",
			operator:    "chinanet",
		},
		{
			name:        "China Unicom ordinary access origins remain eligible",
			asn:         "4837",
			description: "CHINA169-BACKBONE CHINA UNICOM China169 Backbone",
			operator:    "unicom",
		},
		{
			name:        "CNCGROUP remains a bounded China Unicom identifier",
			asn:         "4837",
			description: "CNCGROUP-BACKBONE China Network Communications Group",
			operator:    "unicom",
		},
		{
			name:        "embedded cnc typo is not a China Unicom identifier",
			asn:         "64512",
			description: "TIANHE-TELECOM-BRACNCH",
		},
		{
			name:        "Beijing Telecom provincial network exception",
			asn:         "4847",
			description: "China Networks Inter-Exchange",
			operator:    "chinanet",
		},
		{
			name:            "dedicated IDC description remains excluded",
			asn:             "23724",
			description:     "IDC China Telecommunications Corporation",
			operator:        "chinanet",
			excluded:        true,
			exclusionSource: "description_rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.asn, tt.description)
			if result.Operator != tt.operator || result.Excluded != tt.excluded || result.ExclusionSource != tt.exclusionSource {
				t.Fatalf("Classify(%s, %q) = %+v, want operator=%q excluded=%v exclusion_source=%q", tt.asn, tt.description, result, tt.operator, tt.excluded, tt.exclusionSource)
			}
		})
	}
}

func TestIndependentLegalEntityPattern(t *testing.T) {
	c, err := Load("../../config/operators.json", []string{"chinanet", "cmcc", "unicom"})
	if err != nil {
		t.Fatal(err)
	}
	if !c.IsIndependentLegalEntity("Beijing BG Digital Technology Co.. Ltd") {
		t.Fatal("complete BG-Digital legal entity name was not recognized")
	}
	if c.IsIndependentLegalEntity("BG-Digital") {
		t.Fatal("netname alone must not be legal-entity evidence")
	}
	if c.IsIndependentLegalEntity("Ltd") {
		t.Fatal("legal suffix alone must not be legal-entity evidence")
	}
}

func TestNationwideAPNICRegistrantAdmission(t *testing.T) {
	c, err := Load("../../config/operators.json", []string{"chinanet", "cmcc", "unicom"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		text     string
		operator string
	}{
		{"CHINANET Zhejiang Province Network", "chinanet"},
		{"China Telecom Zhejiang Province Network", "chinanet"},
		{"China Mobile Group Zhejiang Co., Ltd.", "cmcc"},
		{"CMNET-ZHEJIANG", "cmcc"},
		{"China Unicom Zhejiang Province Network", "unicom"},
	}
	for _, tt := range tests {
		if result := c.ClassifyAPNICRegistrant(tt.text); result.Operator != tt.operator {
			t.Fatalf("ClassifyAPNICRegistrant(%q) = %+v, want %s", tt.text, result, tt.operator)
		}
	}
	for _, text := range []string{
		"Ningbo Telecom Co.ltd",
		"Zhejiang Telecommunication Shaoxing Ltd",
		"QuZhou Mobile Communications Co.,Ltd.(QZMCC)",
		"HANGZHOU DIFO TELECOMMUNICATION CO.LTD",
		"Shanghai Great Wall Broadband Network Service Co., Ltd.",
		"Jiaxingshi Xinda Dianzi Keji Co.,Ltd",
		"Hangzhou Network Technology Co., Ltd. Bank of Internet",
	} {
		if result := c.ClassifyAPNICRegistrant(text); result.Operator != "" {
			t.Fatalf("independent registrant %q was admitted as %+v", text, result)
		}
	}
}

func TestNetEaseAndWangyinAPNICRules(t *testing.T) {
	c, err := Load("../../config/operators.json", []string{"chinanet", "cmcc", "unicom"})
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"GUANGZHOUWANGYIHZ | GUANGZHOUWANGYI,HANGZHOU,ZHEJIANG",
		"SHANGHAIWANGYIHZ | SHANGHAIWANGYI,HANGZHOU,ZHEJIANG",
		"GUANGZHOU-WANGYI-LTD | Guangzhou Wangyi Computer Systems Co.,Ltd.",
		"Guangzhou NetEase Computer System Co., Ltd.",
	} {
		if result := c.ClassifyAPNICInetnum(text); !result.Excluded {
			t.Fatalf("NetEase registration %q was not excluded", text)
		}
	}
	for _, text := range []string{
		"WANGYINHULIAN,HANGZHOU,ZHEJIANG",
		"WANGYINHULIANZHEJIANGHENGHUA,HANGZHOU,ZHEJIANG",
		"SHIJIYITENGWANGYINHULIAN,HANGZHOU,ZHEJIANG",
		"HangZhou Netbank Interlink Technolgies CO.,LTD",
	} {
		if result := c.ClassifyAPNICInetnum(text); !result.Excluded {
			t.Fatalf("Wangyin Hulian registration %q was not excluded", text)
		}
	}
	if result := c.ClassifyAPNICInetnum("ordinary residential broadband IP pool"); result.Excluded {
		t.Fatalf("ordinary access pool was excluded: %+v", result)
	}
}

func TestConfirmedZhejiangAPNICRules(t *testing.T) {
	c, err := Load("../../config/operators.json", []string{"chinanet", "cmcc", "unicom"})
	if err != nil {
		t.Fatal(err)
	}
	positives := []string{
		"Zhejiang-Provincial-Bureau-of-Data | Zhejiang Provincial Bureau of Data",
		"NINGBO-GOVERNMENT-NETWORK | Ningbo Electronic Government Network",
		"NINGBO-PEOPLE-GOV | Ningbo Municipal People's Government",
		"IDCBEIYONGJH | IDCBEIYONG,JINHUA,ZHEJIANG",
		"YuYaoIDCYeWuDiZhiDuanVLAN511ChinaunicomNingboChina",
		"ZHUANXIANDIZHIBEIYONGJH | ZHUANXIANDIZHIBEIYONG,JINHUA,ZHEJIANG",
		"CHINATELLECOM-CLOUD-COMPANY",
		"CLOUD-INTERFACE-ADDRESS | Cloud interface address",
		"Beijing Jinshan cloud Network Technology Co., Ltd.",
		"ZHEJIANGZHIYUN | Zhejiang zhi cloud information technology co., LTD",
		"HANGZHOU-YOUPAIYUN-LTD | Hangzhou beat cloud Technology Co. Ltd.",
	}
	for _, text := range positives {
		if result := c.ClassifyAPNICInetnum(text); !result.Excluded {
			t.Fatalf("confirmed Zhejiang non-public registration %q was not excluded", text)
		}
	}
	negatives := []string{
		"IDCCeShi,ZheJiang,Wenzhou",
		"ZHEJIANG-IDCARD-CENTRE | Zhejiang TELECOM",
		"ZHEJIANGZHIYUNXINXI",
		"Hangzhou Office of Ningbo Municipal People's Government",
		"ordinary residential broadband IP pool",
	}
	for _, text := range negatives {
		if result := c.ClassifyAPNICInetnum(text); result.Excluded {
			t.Fatalf("unconfirmed control registration %q was excluded: %+v", text, result)
		}
	}
}

func TestAPNICInetnumRulesNormalizeWhitespace(t *testing.T) {
	c, err := Load("../../config/operators.json", []string{"chinanet", "cmcc", "unicom"})
	if err != nil {
		t.Fatal(err)
	}

	for _, text := range []string{
		"Shaoxing Telecom Bureau Data  Center",
		"Shaoxing Telecom Bureau Data\tCenter",
	} {
		result := c.ClassifyAPNICInetnum(text)
		if !result.Excluded {
			t.Fatalf("APNIC inetnum registration %q was not excluded after whitespace normalization", text)
		}
	}
}
