package operatorconfig

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/apnicinetnum"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/apnicroute"
)

func TestRepositoryOperatorBoundary(t *testing.T) {
	classifier, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
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
			operator:        "chinatelecom",
			excluded:        true,
			exclusionSource: "explicit_policy",
		},
		{
			name:            "China Unicom CUII dedicated premium backbone",
			asn:             "9929",
			description:     "CUII CHINA UNICOM Industrial Internet Backbone",
			operator:        "chinaunicom",
			excluded:        true,
			exclusionSource: "explicit_policy",
		},
		{
			name:            "CTGNet Hong Kong international network",
			asn:             "23764",
			description:     "CTGNet China Telecom Global",
			operator:        "chinatelecom",
			excluded:        true,
			exclusionSource: "explicit_policy",
		},
		{
			name:            "China Unicom Global Hong Kong international gateway",
			asn:             "10099",
			description:     "UNICOM-Global",
			operator:        "chinaunicom",
			excluded:        true,
			exclusionSource: "explicit_policy",
		},
		{
			name:            "China Mobile International Hong Kong network AS58453",
			asn:             "58453",
			description:     "China Mobile International Limited",
			operator:        "chinamobile",
			excluded:        true,
			exclusionSource: "explicit_policy",
		},
		{
			name:            "China Mobile International Hong Kong network AS58807",
			asn:             "58807",
			description:     "China Mobile International Limited",
			operator:        "chinamobile",
			excluded:        true,
			exclusionSource: "explicit_policy",
		},
		{
			name:        "China Telecom ordinary access origins remain eligible",
			asn:         "4134",
			description: "CHINANET-BACKBONE No.31 Jin-rong Street",
			operator:    "chinatelecom",
		},
		{
			name:        "China Unicom ordinary access origins remain eligible",
			asn:         "4837",
			description: "CHINA169-BACKBONE CHINA UNICOM China169 Backbone",
			operator:    "chinaunicom",
		},
		{
			name:        "CNCGROUP remains a bounded China Unicom identifier",
			asn:         "4837",
			description: "CNCGROUP-BACKBONE China Network Communications Group",
			operator:    "chinaunicom",
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
			operator:    "chinatelecom",
		},
		{
			name:            "dedicated IDC description remains excluded",
			asn:             "23724",
			description:     "IDC China Telecommunications Corporation",
			operator:        "chinatelecom",
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
	c, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
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
	c, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		text     string
		operator string
	}{
		{"CHINANET Zhejiang Province Network", "chinatelecom"},
		{"China Telecom Zhejiang Province Network", "chinatelecom"},
		{"China Mobile Group Zhejiang Co., Ltd.", "chinamobile"},
		{"CMNET-ZHEJIANG", "chinamobile"},
		{"China Unicom Zhejiang Province Network", "chinaunicom"},
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
	c, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
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
	c, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
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
	c, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
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

func TestBackboneInfrastructureAPNICRules(t *testing.T) {
	c, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
	if err != nil {
		t.Fatal(err)
	}

	positives := []struct {
		record  apnicinetnum.Record
		reason  string
		inherit bool
		hard    bool
	}{
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"CHINANET-HK"},
				Descriptions: []string{"CHINANET Hongkong region network", "China Telecom"},
				Country:      "CN",
			},
			reason:  "network outside mainland China",
			inherit: true,
			hard:    true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"CHINANET-US-POP"},
				Descriptions: []string{"Chinanet POP in American", "201 S. Lake Ave. Suite 604, Pasadena, CA 91101"},
				Country:      "CN",
			},
			reason:  "network outside mainland China",
			inherit: true,
			hard:    true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"CHINANET-BB"},
				Descriptions: []string{"CHINANET backbone network", "China Telecom"},
			},
			reason:  "legacy CHINANET backbone pool",
			inherit: true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"CNINFONET-BB"},
				Descriptions: []string{"CNINFONET Backbone", "Data Communication Division", "China Telecom"},
			},
			reason:  "CNINFONET backbone pool",
			inherit: true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"CNINANET-IP-PHONE-BB"},
				Descriptions: []string{"CNINANET IP PHONE Backbone", "Data Communication Division", "China Telecom"},
			},
			reason:  "CNINFONET backbone pool",
			inherit: true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"ChinaUnicom-BACKBONE"},
				Descriptions: []string{"Backbone of China Unicom"},
			},
			reason:  "China Unicom backbone pool",
			inherit: true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"CNCGROUP-BACKBONE"},
				Descriptions: []string{"Backbone of CNC group"},
			},
			reason:  "China Unicom backbone pool",
			inherit: true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"sxipbone-loopback"},
				Descriptions: []string{"shanxi telecom ip bone loopback address"},
			},
			reason:  "infrastructure loopback or IP-backbone link pool",
			inherit: true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"sxtyipbone-links"},
				Descriptions: []string{"shanxi telecom taiyuan branch ip backbone links ip address"},
			},
			reason:  "infrastructure loopback or IP-backbone link pool",
			inherit: true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"SD169LOOPBACK"},
				Descriptions: []string{"Shandong 169 Router LOOPBACK-IP"},
			},
			reason:  "infrastructure loopback or IP-backbone link pool",
			inherit: true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"QZ-CNC-LOOPBACK"},
				Descriptions: []string{"Loopback of Network Device"},
			},
			reason:  "infrastructure loopback or IP-backbone link pool",
			inherit: true,
		},
		{
			record: apnicinetnum.Record{
				Netnames:     []string{"InnerMongoliaHailaer82t64g#2loopback"},
				Descriptions: []string{"InnerMongoliaHailaer82t64g#2loopback"},
			},
			reason:  "infrastructure loopback or IP-backbone link pool",
			inherit: true,
		},
	}
	for _, tt := range positives {
		text := apnicinetnum.SearchText(tt.record)
		result := c.ClassifyAPNICInetnum(text)
		if !result.Excluded {
			t.Fatalf("confirmed backbone infrastructure registration %q was not excluded", text)
		}
		if !strings.Contains(result.Reason, tt.reason) || !strings.HasPrefix(result.MatchedBy, "exclude_apnic_inetnum_rules: ") {
			t.Fatalf("confirmed backbone infrastructure registration %q matched the wrong rule: %+v", text, result)
		}
		if result.Inherit != tt.inherit || result.Hard != tt.hard {
			t.Fatalf("confirmed backbone infrastructure registration %q has wrong hierarchy flags: %+v", text, result)
		}
	}

	negatives := []string{
		"netname=CHINANET-GD | Data Communication Division | China Telecom",
		"netname=CHINANET-ZJ | Zhejiang Province Network | Data Communication Division | China Telecom",
		"CHINANET-BACKBONE No.31 Jin-rong Street",
		"CHINA169-BACKBONE CHINA UNICOM China169 Backbone",
		"CNCGROUP-BACKBONE China Network Communications Group",
		"CHINAMOBILE-CN China Mobile Communications Group Co., Ltd.",
		"CMNET | China Mobile Communications Corporation | Mobile Communications Network Operator in China",
		"AH-IP-BACKBONE-AREA1 | Anhui Unicom IP",
		"Backbone and Customer Address Space",
		"netname=CNINFONET-BB-CUSTOMER | ordinary residential broadband pool",
		"netname=ChinaUnicom-BACKBONE-CUSTOMER | Backbone of China Unicom",
		"netname=CHINANET-US-POP-CUSTOMER | ordinary residential broadband pool",
		"netname=sxtyipbone-links-customer | ordinary residential broadband pool",
		"netname=ORDINARY-ACCESS | customer transferred from CNINFONET-BB",
		"netname=LoopbackNetworks | ordinary residential broadband pool",
	}
	for _, text := range negatives {
		if result := c.ClassifyAPNICInetnum(text); result.Excluded {
			t.Fatalf("ambiguous backbone registration %q was excluded: %+v", text, result)
		}
	}

	for _, text := range []string{
		apnicroute.SearchText(apnicroute.Variant{Descriptions: []string{"CHINANET-BB", "CHINANET backbone network", "Data Communication Division"}}),
		apnicroute.SearchText(apnicroute.Variant{Descriptions: []string{"CNINFONET-BB"}}),
		apnicroute.SearchText(apnicroute.Variant{Descriptions: []string{"ChinaUnicom-BACKBONE", "Backbone of China Unicom"}}),
		apnicroute.SearchText(apnicroute.Variant{Descriptions: []string{"sxtyipbone-links"}}),
		apnicroute.SearchText(apnicroute.Variant{Descriptions: []string{"SD169LOOPBACK", "Shandong 169 Router LOOPBACK-IP"}}),
	} {
		if result := c.ClassifyAPNICInetnum(text); result.Excluded {
			t.Fatalf("route description without an inetnum netname was excluded as backbone infrastructure: %+v", result)
		}
	}
}

func TestAPNICTerminalAccessRules(t *testing.T) {
	c, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
	if err != nil {
		t.Fatal(err)
	}

	positives := []apnicinetnum.Record{
		{Netnames: []string{"RESIDENTIAL"}, Descriptions: []string{"ordinary residential broadband IP pool"}},
		{Netnames: []string{"CIXI-ADSL-USERS"}, Descriptions: []string{"ADSL users with static IP address in CIXI"}},
		{Netnames: []string{"CHINANET-NX"}, Descriptions: []string{"yinchuan node adsl ip pool"}},
		{Netnames: []string{"SHANGHAI-TELECOM-DATA"}, Descriptions: []string{"adsl ip pool for shanghai telecom"}},
		{Netnames: []string{"CNCGROUP-GX-GG-BAS-5"}, Descriptions: []string{"CNC GuiGang Wide Band Network-5 pppoe ip pool"}},
		{Netnames: []string{"DZJX-BIP"}, Descriptions: []string{"Shandong Dezhou Jianxiang Residential Area Broad IP Access User"}},
		{Netnames: []string{"MOBILE-USERS"}, Descriptions: []string{"LTE dynamic IP pool for subscribers"}},
	}
	for _, record := range positives {
		text := apnicinetnum.SearchText(record)
		result := c.ClassifyAPNICAccess(text)
		if result.Reason == "" || !strings.HasPrefix(result.MatchedBy, "access_apnic_inetnum_rules: ") {
			t.Fatalf("terminal-access registration %q was not recognized: %+v", text, result)
		}
	}

	negatives := []apnicinetnum.Record{
		{Netnames: []string{"FSKWC"}, Descriptions: []string{"FSKWC NET"}},
		{Netnames: []string{"CHINANET-ZJ"}, Descriptions: []string{"CHINANET Zhejiang province network"}},
		{Netnames: []string{"DQ-ADSL"}, Descriptions: []string{"Da Qing city ADSL device management IP"}},
		{Netnames: []string{"DT-ADSL-TEST"}, Descriptions: []string{"ShanXi Province DaTong City ADSL Test"}},
		{Netnames: []string{"DANCHENG-DANDONG-POLICE-STATE"}, Descriptions: []string{"DanCheng Dandong police station FTTH"}},
		{Netnames: []string{"SUPERHUB-HK"}, Descriptions: []string{"Cloud, Broadband& Hosting Service"}},
		{Netnames: []string{"CZPPPOEPOOL-COM"}, Descriptions: []string{"CZPPPOEPOOL CAR CO., changzhou, JIANGSU province"}},
		{Netnames: []string{"CMNET-WLAN"}, Descriptions: []string{"China Mobile WLAN pool"}},
		{Netnames: []string{"LEASED-ADSL"}, Descriptions: []string{"ADSL leased line"}},
	}
	for _, record := range negatives {
		text := apnicinetnum.SearchText(record)
		if result := c.ClassifyAPNICAccess(text); result.Reason != "" {
			t.Fatalf("non-terminal registration %q was accepted as access evidence: %+v", text, result)
		}
	}

	routeText := apnicroute.SearchText(apnicroute.Variant{Descriptions: []string{"ADSL users", "residential broadband IP pool"}})
	if result := c.ClassifyAPNICAccess(routeText); result.Reason != "" {
		t.Fatalf("route description without an inetnum netname was accepted as access evidence: %+v", result)
	}
}

func TestClassifyAPNICPolicyMatchesIndividualClassifiers(t *testing.T) {
	c, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
	if err != nil {
		t.Fatal(err)
	}

	dualMatch := apnicinetnum.SearchText(apnicinetnum.Record{
		Netnames: []string{"CHINANET-BB"},
		Descriptions: []string{
			"CHINANET backbone network",
			"ordinary residential broadband IP pool",
			"Data  Center",
		},
	})
	texts := []string{
		dualMatch,
		apnicinetnum.SearchText(apnicinetnum.Record{
			Netnames:     []string{"RESIDENTIAL"},
			Descriptions: []string{"ordinary residential broadband IP pool"},
		}),
		"Shaoxing Telecom Bureau Data\tCenter",
		"netname=FSKWC | FSKWC NET",
	}
	for _, text := range texts {
		wantExclusion := c.ClassifyAPNICInetnum(text)
		wantAccess := c.ClassifyAPNICAccess(text)
		gotExclusion, gotAccess := c.ClassifyAPNICPolicy(text)
		if gotExclusion != wantExclusion || gotAccess != wantAccess {
			t.Fatalf("ClassifyAPNICPolicy(%q) = (%+v, %+v), want (%+v, %+v)",
				text, gotExclusion, gotAccess, wantExclusion, wantAccess)
		}
	}

	exclusion, access := c.ClassifyAPNICPolicy(dualMatch)
	if !strings.Contains(exclusion.MatchedBy, "netname=chinanet-bb") {
		t.Fatalf("dual-match exclusion did not preserve first-rule priority: %+v", exclusion)
	}
	if access.Reason == "" {
		t.Fatalf("dual-match access evidence was not returned: %+v", access)
	}
}

func TestRepositoryReviewedBackboneIPv4Prefixes(t *testing.T) {
	c, err := Load("../../config/operators.json", []string{"chinatelecom", "chinamobile", "chinaunicom"})
	if err != nil {
		t.Fatal(err)
	}

	got := c.ReviewedBackboneIPv4Prefixes()
	if len(got) == 0 {
		t.Fatal("ReviewedBackboneIPv4Prefixes() returned no entries")
	}
	if got[0].CIDR != "59.43.0.0/16" || got[0].Operator != "chinatelecom" || got[0].Reason == "" || len(got[0].EvidenceURLs) != 2 {
		t.Fatalf("ReviewedBackboneIPv4Prefixes()[0] = %+v", got[0])
	}

	got[0].CIDR = "192.0.2.0/24"
	got[0].EvidenceURLs[0] = "https://example.com/mutated"
	fresh := c.ReviewedBackboneIPv4Prefixes()
	if fresh[0].CIDR != "59.43.0.0/16" || fresh[0].EvidenceURLs[0] != "https://rdap.apnic.net/ip/59.43.0.0/16" {
		t.Fatalf("ReviewedBackboneIPv4Prefixes returned mutable classifier state: %+v", fresh[0])
	}
}

func TestBackboneIPv4PrefixValidation(t *testing.T) {
	valid := BackboneIPv4Prefix{
		CIDR:         "59.43.0.0/16",
		Operator:     "chinatelecom",
		Reason:       "APNIC registration explicitly identifies a backbone pool",
		EvidenceURLs: []string{"https://rdap.apnic.net/ip/59.43.0.0/16"},
	}

	tests := []struct {
		name    string
		entries []BackboneIPv4Prefix
	}{
		{
			name: "non-canonical CIDR",
			entries: []BackboneIPv4Prefix{{
				CIDR:         "59.43.1.0/16",
				Operator:     valid.Operator,
				Reason:       valid.Reason,
				EvidenceURLs: valid.EvidenceURLs,
			}},
		},
		{
			name: "IPv6 CIDR",
			entries: []BackboneIPv4Prefix{{
				CIDR:         "2001:db8::/32",
				Operator:     valid.Operator,
				Reason:       valid.Reason,
				EvidenceURLs: valid.EvidenceURLs,
			}},
		},
		{
			name: "unknown operator",
			entries: []BackboneIPv4Prefix{{
				CIDR:         valid.CIDR,
				Operator:     "unknown",
				Reason:       valid.Reason,
				EvidenceURLs: valid.EvidenceURLs,
			}},
		},
		{
			name: "empty reason",
			entries: []BackboneIPv4Prefix{{
				CIDR:         valid.CIDR,
				Operator:     valid.Operator,
				Reason:       " ",
				EvidenceURLs: valid.EvidenceURLs,
			}},
		},
		{
			name: "no evidence",
			entries: []BackboneIPv4Prefix{{
				CIDR:     valid.CIDR,
				Operator: valid.Operator,
				Reason:   valid.Reason,
			}},
		},
		{
			name: "non-HTTPS evidence",
			entries: []BackboneIPv4Prefix{{
				CIDR:         valid.CIDR,
				Operator:     valid.Operator,
				Reason:       valid.Reason,
				EvidenceURLs: []string{"http://rdap.apnic.net/ip/59.43.0.0/16"},
			}},
		},
		{
			name:    "duplicate CIDR",
			entries: []BackboneIPv4Prefix{valid, valid},
		},
		{
			name: "overlapping CIDRs",
			entries: []BackboneIPv4Prefix{
				valid,
				{
					CIDR:         "59.43.0.0/17",
					Operator:     valid.Operator,
					Reason:       valid.Reason,
					EvidenceURLs: valid.EvidenceURLs,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseBackboneConfigForTest(t, tt.entries); err == nil {
				t.Fatal("Parse() succeeded, want validation error")
			}
		})
	}
}

func parseBackboneConfigForTest(t *testing.T, entries []BackboneIPv4Prefix) (*Classifier, error) {
	t.Helper()
	cfg := configFile{
		Operators: map[string]operator{
			"chinatelecom": {DescriptionPatterns: []string{"chinatelecom"}},
			"chinamobile":  {DescriptionPatterns: []string{"chinamobile"}},
			"chinaunicom":  {DescriptionPatterns: []string{"chinaunicom"}},
		},
		ExcludeAPNICInetnumRules:       []descriptionRule{{Pattern: "never-match", Reason: "test exclusion rule"}},
		AccessAPNICInetnumRules:        []descriptionRule{{Pattern: "never-match", Reason: "test access rule"}},
		ExcludeBackboneIPv4Prefixes:    entries,
		IndependentLegalEntityPatterns: []string{"never-match"},
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal test config: %v", err)
	}
	return Parse(b, []string{"chinatelecom", "chinamobile", "chinaunicom"})
}
