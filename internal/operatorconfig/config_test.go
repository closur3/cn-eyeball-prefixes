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
