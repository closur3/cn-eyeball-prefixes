# 三网 IPv6 APNIC 登记颗粒度审计

生成时间：`2026-07-21T19:01:05.858023574Z`

审计对象是 `当前三网 IPv6 Origin ∩ china6`。`inet6num` 按最具体记录解析；`route6` 只统计与当前 BGP Origin 相同的登记。报告包含首轮规则预演，但不生成正式地址列表。

APNIC `inet6num` 记录：**148952**；解析后最具体区间：**194099**；`route6` 前缀：**1014144**。

## 运营商覆盖

| 运营商 | 候选 CIDR | inet6num 覆盖 | 未覆盖 | 同 Origin route6 覆盖 | route6 强非目标信号 |
| --- | ---: | ---: | ---: | ---: | ---: |
| chinanet | 3090 | 99.999993% | 0.000007% | 97.646828% | 0.000000% |
| cmcc | 352 | 100.000000% | 0.000000% | 0.000000% | 0.000000% |
| unicom | 4161 | 100.000000% | 0.000000% | 99.552494% | 0.000000% |

## 首轮准入规则预演

预演排除明确非目标用途、其他运营商登记、独立法定主体登记、无法归属三网的 portable 资源、非 CN 登记及无 APNIC 登记；其余保留。`明确用户侧正证据`只是保留结果中的事实标记，不是唯一准入条件。

| 运营商 | 保留 CIDR | 保留空间 | 排除 CIDR | 排除空间 | 明确用户侧正证据 |
| --- | ---: | ---: | ---: | ---: | ---: |
| chinanet | 771 | 50.095024% | 2320 | 49.904976% | 18.838693% |
| cmcc | 288 | 96.991219% | 64 | 3.008781% | 0.000000% |
| unicom | 4136 | 99.592897% | 25 | 0.407103% | 0.000000% |

## 明确用户侧正证据样本

| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 首轮处理依据 |
| --- | --- | ---: | --- | --- | --- |
| chinanet | `240e:b00::/24` | 6.276026% | CT-IPV6-BROADBAND-ADDRESS; Chinatelecom IPv6 address for fixed broadband | ALLOCATED NON-PORTABLE | Retain: most-specific APNIC registration is attributed to the current operator |
| chinanet | `240e:400::/24` | 6.276011% | CT-IPV6-MOBILE-ADDRESS; Chinatelecom IPv6 address for mobile | ALLOCATED NON-PORTABLE | Retain: most-specific APNIC registration is attributed to the current operator |
| chinanet | `240e:300::/24` | 6.269131% | CT-IPV6-BROADBAND-ADDRESS; Chinatelecom IPv6 address for fixed broadband | ALLOCATED NON-PORTABLE | Retain: most-specific APNIC registration is attributed to the current operator |

## 保留但只有泛化运营商证据的样本

| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 首轮处理依据 |
| --- | --- | ---: | --- | --- | --- |
| cmcc | `2409:8000::/20` | 96.991219% | CMNET-V6-20110823; China Mobile Communications Corporation; Mobile Communications Network Operator in China; Internet Service Provider in China; China Mobile | ALLOCATED PORTABLE | Retain: most-specific APNIC registration is attributed to the current operator |
| chinanet | `240e::/18` | 31.177326% | CT-IPv6-Networks; Chinatelecom networks with tens of high-end routers and switches; Including users who access to Internet through Chinatelecom's networks.; China Telecom | ALLOCATED PORTABLE | Retain: most-specific APNIC registration is attributed to the current operator |
| unicom | `2408:8000::/20` | 99.592897% | CU-CN; China Unicom; No.21, Jin-Rong Street; Beijng 100033 | ALLOCATED PORTABLE | Retain: most-specific APNIC registration is attributed to the current operator |
| chinanet | `240e:16::/31` | 0.047498% | CHINANET-SC-IPv6-USER-ADDRESS; CHINANET Sichuan province network; Data Communication Division; China Telecom | ALLOCATED NON-PORTABLE | Retain: most-specific APNIC registration is attributed to the current operator |
| chinanet | `2001:c68::/32` | 0.024516% | CHINANET-20020830; China Telecom; Internet Service Provider; Beijing,China | ALLOCATED PORTABLE | Retain: most-specific APNIC registration is attributed to the current operator |
| chinanet | `240e:ed::/32` | 0.024516% | China-Telecom-Jiangsu-province-network; China Telecom Jiangsu province network for customer | ALLOCATED NON-PORTABLE | Retain: most-specific APNIC registration is attributed to the current operator |

## 首轮排除样本

| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 首轮处理依据 |
| --- | --- | ---: | --- | --- | --- |
| chinanet | `240e:500::/24` | 6.276026% | CT-IPV6-VOLTE-ADDRESS; Chinatelecom IPv6 address for volte | ALLOCATED NON-PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| chinanet | `240e:a00::/24` | 6.276026% | CT-IPV6-INTERNET-CAFE-ADDRESS; Chinatelecom IPv6 address for Internet cafe leased line | ALLOCATED NON-PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| chinanet | `240e:100::/24` | 6.275040% | CT-IPV6-NETWORK-ADDRESS; Chinatelecom IPv6 address for network | ALLOCATED NON-PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| chinanet | `240e:200::/24` | 6.273824% | CT-IPV6-PLATFORM-ADDRESS; Chinatelecom IPv6 address for own platform | ALLOCATED NON-PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| chinanet | `240e:800::/24` | 6.271502% | CT-IPV6-IOT-ADDRESS; Chinatelecom IPv6 address for IOT | ALLOCATED NON-PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| chinanet | `240e:600::/24` | 6.261335% | CT-IPV6-LEASED-LINE-ADDRESS; Chinatelecom IPv6 address for LEASED LINE | ALLOCATED NON-PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| chinanet | `240e:700::/24` | 6.229961% | CT-IPV6-LEASED-LINE-ADDRESS; Chinatelecom IPv6 address for LEASED LINE | ALLOCATED NON-PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| chinanet | `240e:900::/24` | 5.974659% | CT-IPV6-IDC-ADDRESS; Chinatelecom IPv6 address for IDC & Cloud service | ALLOCATED NON-PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| cmcc | `240a:4000::/21` | 2.937741% | CBN-CN; China Broadcasting Network Corporation Ltd.; No.10 Baiyun Road, Xicheng District, Beijing; China Internet Network Information Center | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| chinanet | `2400:75a0::/28` | 0.024516% | DCCLGOOXNAHP-CN; Digital City Construction Leading Group Office of Xiongan New Area, Hebei Province; Digital City Construction Leading Group Office of Xiongan New Area, Hebei Province | ALLOCATED PORTABLE | Exclude: portable resource cannot be attributed to the current operator |
| cmcc | `2401:1320::/32` | 0.023680% | YUNSILKIPCOM; Silk Road Information Port Cloud Computing Technology Co., Ltd; No. 72 Beibinhe East Road, Chengguan District, Lanzhou , Gansu | ALLOCATED PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| unicom | `2402:18a0::/32` | 0.202218% | SAUNET; Shenyang Aerospace University; Shenyang Aerospace University, No.37 Daoyi South Avenue,; Daoyi District, Shenyang, Liaoning Province, China | ALLOCATED PORTABLE | Exclude: portable resource cannot be attributed to the current operator |
| cmcc | `2402:9a80::/32` | 0.023680% | JSMNET; Jishi Media Co ., Ltd.; No.1027-1, Xinmin Street, Changchun | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| unicom | `2402:ef40::/32` | 0.202218% | DRCSCNET; Development & Research Center of State Council Net.; No.225 Chaonei Street Dongcheng District Beijing China | ALLOCATED PORTABLE | Exclude: portable resource cannot be attributed to the current operator |
| chinanet | `2404:1c80::/32` | 0.024516% | GDXCNET; BeiJing guangdianxinchuang communication &; technology C0.,LTD.; Room 2205,Building 2,Zhubang 2000 Business Center,; No.99 Balizhuang Xili,Chaoyang District | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| cmcc | `2407:37c0::/32` | 0.023680% | JSMNET; Jishi Media Co ., Ltd.; No.1027-1, Xinmin Street, Changchun | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| chinanet | `240e:83:8000::/33` | 0.012162% | BeiJing-Telecom-UserAddress-lowguaranteed; BeiJing Telecom UserAddress for low-guaranteed | ALLOCATED NON-PORTABLE | Exclude: portable resource cannot be attributed to the current operator |
| chinanet | `240e:83::/34` | 0.004597% | BeiJing-Telecom-UserAddress-Highguaranteed; BeiJing Telecom UserAddress for High-guaranteed | ALLOCATED NON-PORTABLE | Exclude: portable resource cannot be attributed to the current operator |
| chinanet | `240e:83::/36` | 0.000766% | BeiJing-Telecom-UserAddress-Highguaranteed; BeiJing Telecom UserAddress for High-guaranteed | ASSIGNED NON-PORTABLE | Exclude: portable resource cannot be attributed to the current operator |
| unicom | `2406:cac0::/32` | 0.001580% | DFM; Dongfeng Communication Technology Co.,Ltd. | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| unicom | `2407:6c40::/32` | 0.000793% | BJ-SHOUZIXIN; Beijing Shougang Automation Information Technology Co.,Ltd; Building 1, Yard 1, Shimen Road, Shijingshan, Beijing | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| chinanet | `2402:f8c0::/32` | 0.000024% | Digital-Guangdong; Digital Guangdong Network Construction Co, Ltd | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| unicom | `2400:cb80::/32` | 0.000105% | BMW-SF-CN; BMW Automotive Finance (China) Co., Ltd.; 22nd Floor, Tower B, Gateway Plaza; No. 18 Xia Guang Li North Road, East Third Ring; Chaoyang District, Beijing 100027, PR China | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| unicom | `2402:dfc0::/32` | 0.000099% | JD-finance-network; Beijing JD Finance Technology Holding Co., Ltd. | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| chinanet | `2406:d440::/32` | 0.000006% | VOLCANO-ENGINE; Beijing Volcano Engine Technology Co., Ltd.; 1309, 13/F, Building 4, Zijin Digital Park, Haidian District, Beijing | ALLOCATED PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| unicom | `2400:89c0::/32` | 0.000049% | SINA; 15F,Ideal Plaza No.58 Bei Si Huan Xi Road Haidian District; Beijing,China,100080 | ALLOCATED PORTABLE | Exclude: portable resource cannot be attributed to the current operator |
| chinanet | `2400:9380::/31` | 0.000001% | CTGI-AS-AP; China Telecom Global Limited; 38/F DAH SING Financial Center; 108 Gloucester Road Wan Chai; China Telecom Global Limited | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration is outside mainland China |
| chinanet | `2401:1d40::/32` | 0.000001% | BJKSCNET; Beijing Kingsoft Cloud Internet Technology Co., Ltd.; Kingsoft Tower,No.33 Xiao Ying West Road,Haidian District,Beijing,China | ALLOCATED PORTABLE | Exclude: APNIC registration contains an explicit non-user-side purpose |
| unicom | `2402:f140::/32` | 0.000009% | AIPO-CN; AIPO Cloud (Guizhou) Technology Co., Ltd. | ALLOCATED PORTABLE | Exclude: most-specific APNIC registration names an independent legal entity |
| chinanet | `2400:9380:8000::/44` | 0.000001% | CTGI-AS-AP; Chinatelecom global limit 38/F., DAH SING Financial Center, 108 Gloucester Road, Wan Chai, Hong Kong | ALLOCATED NON-PORTABLE | Exclude: most-specific APNIC registration is outside mainland China |

## 最具体 inet6num 分类

| 运营商 | 分类 | CIDR | /64 等价数 | 占运营商候选 |
| --- | --- | ---: | ---: | ---: |
| chinanet | same_operator | 778 | 8776263991296.0000 | 50.095027% |
| chinanet | strong_non_public | 2268 | 8731301838848.0000 | 49.838382% |
| chinanet | other_operator | 0 | 0.0000 | 0.000000% |
| chinanet | independent_legal_entity | 15 | 4299620352.0000 | 0.024542% |
| chinanet | other_or_unclassified | 14 | 7365394432.0000 | 0.042042% |
| cmcc | same_operator | 288 | 17592123785216.0000 | 96.991219% |
| cmcc | strong_non_public | 2 | 4295032832.0000 | 0.023680% |
| cmcc | other_operator | 0 | 0.0000 | 0.000000% |
| cmcc | independent_legal_entity | 62 | 541433266176.0000 | 2.985101% |
| cmcc | other_or_unclassified | 0 | 0.0000 | 0.000000% |
| unicom | same_operator | 4136 | 2115278143488.0000 | 99.592897% |
| unicom | strong_non_public | 0 | 0.0000 | 0.000000% |
| unicom | other_operator | 0 | 0.0000 | 0.000000% |
| unicom | independent_legal_entity | 19 | 55377920.0000 | 0.002607% |
| unicom | other_or_unclassified | 6 | 8591179776.0000 | 0.404495% |

## strong_non_public：地址量前 100 项

| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 依据 |
| --- | --- | ---: | --- | --- | --- |
| chinanet | `240e:500::/24` | 6.276026% | CT-IPV6-VOLTE-ADDRESS; Chinatelecom IPv6 address for volte | ALLOCATED NON-PORTABLE | APNIC inet6num explicitly identifies a VoLTE service address pool rather than ordinary Internet access (`ipv6_explicit_service: volte`) |
| chinanet | `240e:a00::/24` | 6.276026% | CT-IPV6-INTERNET-CAFE-ADDRESS; Chinatelecom IPv6 address for Internet cafe leased line | ALLOCATED NON-PORTABLE | APNIC inetnum registration explicitly identifies a dedicated line or circuit (`exclude_apnic_inetnum_rules: \b(?:leased|dedicated|private|special)[ -]?(?:line|circuit)s?\b`) |
| chinanet | `240e:100::/24` | 6.275040% | CT-IPV6-NETWORK-ADDRESS; Chinatelecom IPv6 address for network | ALLOCATED NON-PORTABLE | APNIC inet6num explicitly identifies a generic network address pool without end-user purpose (`ipv6_explicit_service: network`) |
| chinanet | `240e:200::/24` | 6.273824% | CT-IPV6-PLATFORM-ADDRESS; Chinatelecom IPv6 address for own platform | ALLOCATED NON-PORTABLE | APNIC inet6num explicitly identifies the operator's own platform address pool (`ipv6_explicit_service: own platform`) |
| chinanet | `240e:800::/24` | 6.271502% | CT-IPV6-IOT-ADDRESS; Chinatelecom IPv6 address for IOT | ALLOCATED NON-PORTABLE | APNIC inetnum registration explicitly identifies an IoT or M2M network (`exclude_apnic_inetnum_rules: (?:^|[^a-z0-9])(?:iot|m2m)(?:[^a-z0-9]|$)|\binternet of things\b`) |
| chinanet | `240e:600::/24` | 6.261335% | CT-IPV6-LEASED-LINE-ADDRESS; Chinatelecom IPv6 address for LEASED LINE | ALLOCATED NON-PORTABLE | APNIC inetnum registration explicitly identifies a dedicated line or circuit (`exclude_apnic_inetnum_rules: \b(?:leased|dedicated|private|special)[ -]?(?:line|circuit)s?\b`) |
| chinanet | `240e:700::/24` | 6.229961% | CT-IPV6-LEASED-LINE-ADDRESS; Chinatelecom IPv6 address for LEASED LINE | ALLOCATED NON-PORTABLE | APNIC inetnum registration explicitly identifies a dedicated line or circuit (`exclude_apnic_inetnum_rules: \b(?:leased|dedicated|private|special)[ -]?(?:line|circuit)s?\b`) |
| chinanet | `240e:900::/24` | 5.974659% | CT-IPV6-IDC-ADDRESS; Chinatelecom IPv6 address for IDC & Cloud service | ALLOCATED NON-PORTABLE | APNIC inetnum registration explicitly identifies an IDC network (`exclude_apnic_inetnum_rules: idc(?:[^a-z0-9]|$)|\binternet data cent(?:er|re)\b`) |
| cmcc | `2401:1320::/32` | 0.023680% | YUNSILKIPCOM; Silk Road Information Port Cloud Computing Technology Co., Ltd; No. 72 Beibinhe East Road, Chengguan District, Lanzhou , Gansu | ALLOCATED PORTABLE | APNIC inetnum registration explicitly identifies a cloud-computing network (`exclude_apnic_inetnum_rules: \bcloud computing\b`) |
| chinanet | `2406:d440::/32` | 0.000006% | VOLCANO-ENGINE; Beijing Volcano Engine Technology Co., Ltd.; 1309, 13/F, Building 4, Zijin Digital Park, Haidian District, Beijing | ALLOCATED PORTABLE | APNIC inetnum registration explicitly identifies a known cloud or CDN brand (`exclude_apnic_inetnum_rules: (?:^|[^a-z0-9])(?:aliyun|qcloud|ucloud|ksyun|jdcloud|ctyun|qingcloud|volcengine|cloudflare|akamai|wangsu|chinanetcenter|baishancloud|sinnet|westclouddata|nwcd)(?:[^a-z0-9]|$)|\bbeijing guanghuan xinwang(?:[ |.,-]+digital technology)?\b|\b(?:alibaba|huawei|tencent|baidu|kingsoft|jingdong|jd|tianyi|china telecom|china unicom|china mobile|google|oracle|ibm)[ -]?(?:ai[ -]?)?cloud\b|\bamazon web services\b|\bmicrosoft azure\b|\bvolcano engine\b`) |
| chinanet | `2401:1d40::/32` | 0.000001% | BJKSCNET; Beijing Kingsoft Cloud Internet Technology Co., Ltd.; Kingsoft Tower,No.33 Xiao Ying West Road,Haidian District,Beijing,China | ALLOCATED PORTABLE | APNIC inetnum registration explicitly identifies a known cloud or CDN brand (`exclude_apnic_inetnum_rules: (?:^|[^a-z0-9])(?:aliyun|qcloud|ucloud|ksyun|jdcloud|ctyun|qingcloud|volcengine|cloudflare|akamai|wangsu|chinanetcenter|baishancloud|sinnet|westclouddata|nwcd)(?:[^a-z0-9]|$)|\bbeijing guanghuan xinwang(?:[ |.,-]+digital technology)?\b|\b(?:alibaba|huawei|tencent|baidu|kingsoft|jingdong|jd|tianyi|china telecom|china unicom|china mobile|google|oracle|ibm)[ -]?(?:ai[ -]?)?cloud\b|\bamazon web services\b|\bmicrosoft azure\b|\bvolcano engine\b`) |
| chinanet | `2401:3480::/32` | 0.000000% | UCLOUD-NET; Shanghai UCloud Information Technology Company Limited | ALLOCATED PORTABLE | APNIC inetnum registration explicitly identifies a known cloud or CDN brand (`exclude_apnic_inetnum_rules: (?:^|[^a-z0-9])(?:aliyun|qcloud|ucloud|ksyun|jdcloud|ctyun|qingcloud|volcengine|cloudflare|akamai|wangsu|chinanetcenter|baishancloud|sinnet|westclouddata|nwcd)(?:[^a-z0-9]|$)|\bbeijing guanghuan xinwang(?:[ |.,-]+digital technology)?\b|\b(?:alibaba|huawei|tencent|baidu|kingsoft|jingdong|jd|tianyi|china telecom|china unicom|china mobile|google|oracle|ibm)[ -]?(?:ai[ -]?)?cloud\b|\bamazon web services\b|\bmicrosoft azure\b|\bvolcano engine\b`) |
| cmcc | `2401:8be0::/32` | 0.000000% | SIHE-SCCC; Sichuan Sihe Cloud Computing Co., Ltd.; JM-B-15 block, Jiumian Industrial Park, Wudu District,; Hanwang Town, Mianzhu, Deyang, Sichuan | ALLOCATED PORTABLE | APNIC inetnum registration explicitly identifies a cloud-computing network (`exclude_apnic_inetnum_rules: \bcloud computing\b`) |
| chinanet | `2402:92c0::/32` | 0.000000% | YOVOLECLOUD; Beijing Yovole Cloud Computing Technology Company Limited; Room 304,No.2 Building, Fuhai Center,; Daliushu Road, Haidian Distirct, Beijing | ALLOCATED PORTABLE | APNIC inetnum registration explicitly identifies a cloud-computing network (`exclude_apnic_inetnum_rules: \bcloud computing\b`) |
| chinanet | `2403:a200::/32` | 0.000000% | CHINA-21VIANET; 21ViaNet(China),Inc.; BOE Science Park, 10 Jiuxianqiao Road, Chaoyang,; Beijing 100016, China | ALLOCATED PORTABLE | APNIC inetnum registration explicitly identifies 21Vianet or CNISP (`exclude_apnic_inetnum_rules: (?:^|[^a-z0-9])(?:21vianet|cnisp)(?:[^a-z0-9]|$)`) |
| chinanet | `2409:2000::/21` | 0.000000% | HWCSNET; Huawei Public Cloud Service (Huawei Software Technologies Ltd.Co); No.2018 Xuegang Road,Bantian street,Longgang District,; Shenzhen,Guangdong Province, 518129 P.R.China; China Internet Network Information Center | ALLOCATED PORTABLE | APNIC inetnum registration explicitly identifies a cloud-service network (`exclude_apnic_inetnum_rules: \bcloud[ -]?(?:service|platform|data|hosting|server)s?\b`) |

## independent_legal_entity：地址量前 100 项

| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 依据 |
| --- | --- | ---: | --- | --- | --- |
| cmcc | `240a:4000::/21` | 2.937741% | CBN-CN; China Broadcasting Network Corporation Ltd.; No.10 Baiyun Road, Xicheng District, Beijing; China Internet Network Information Center | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| cmcc | `2402:9a80::/32` | 0.023680% | JSMNET; Jishi Media Co ., Ltd.; No.1027-1, Xinmin Street, Changchun | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2404:1c80::/32` | 0.024516% | GDXCNET; BeiJing guangdianxinchuang communication &; technology C0.,LTD.; Room 2205,Building 2,Zhubang 2000 Business Center,; No.99 Balizhuang Xili,Chaoyang District | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| cmcc | `2407:37c0::/32` | 0.023680% | JSMNET; Jishi Media Co ., Ltd.; No.1027-1, Xinmin Street, Changchun | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2406:cac0::/32` | 0.001580% | DFM; Dongfeng Communication Technology Co.,Ltd. | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2407:6c40::/32` | 0.000793% | BJ-SHOUZIXIN; Beijing Shougang Automation Information Technology Co.,Ltd; Building 1, Yard 1, Shimen Road, Shijingshan, Beijing | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2402:f8c0::/32` | 0.000024% | Digital-Guangdong; Digital Guangdong Network Construction Co, Ltd | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2400:cb80::/32` | 0.000105% | BMW-SF-CN; BMW Automotive Finance (China) Co., Ltd.; 22nd Floor, Tower B, Gateway Plaza; No. 18 Xia Guang Li North Road, East Third Ring; Chaoyang District, Beijing 100027, PR China | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2402:dfc0::/32` | 0.000099% | JD-finance-network; Beijing JD Finance Technology Holding Co., Ltd. | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2402:f140::/32` | 0.000009% | AIPO-CN; AIPO Cloud (Guizhou) Technology Co., Ltd. | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2400:9600::/32` | 0.000000% | DMTNET; Shanghai DMT Information Network cor.,LTD.; 23F YUNHUA Science & Tech.Building, No.912; Gonghexin Road, Shanghai, P.R.China | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2400:be00::/32` | 0.000000% | ZSPNET; BEIJING ZHONGGUANCUN SOFTWARE PARK DEVELOPMENT CO.,Ltd.; P.O.Box 5118,Zhongguancun Software Park,; Haidian District, Beijing P.R.C. | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2401:a140::/32` | 0.000000% | CLOUDTIMES; Bei Jing Cloud Times Technology Co.,Ltd; Room 2804,Tianchang Park, Building 7, Beiyuan Road Olympic Media Village,; Chaoyang District, Beijing,China | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2402:f140::/32` | 0.000000% | AIPO-CN; AIPO Cloud (Guizhou) Technology Co., Ltd. | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2403:1ec0::/32` | 0.000000% | JD; Beijing Jingdong Shangke Information Technology Co. Ltd. | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2403:6740::/32` | 0.000003% | LDST; Beijing Dicai Network Communications Technology Co., Ltd. | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2403:8080::/32` | 0.000000% | DXTNET; Beijing Teletron Telecom Engineering Co., Ltd.; Jian Guo Road, Chaoyang District, Beijing, PR.China | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2403:e7c0::/32` | 0.000003% | TaiJi-ZY; Taiji Computer Corporation Limited | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2404:6500::/32` | 0.000003% | IFLYTEK; ANHUI USTC iFLYTEK Co., Ltd.; No. 666 West Wangjiang Road, Hefei, Anhui | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2404:7600::/32` | 0.000003% | DSNET; Shanghai Data Solution Co., Ltd.; 2F,NO.4Buliding 498 Guoshoujing Rd.Shanghai ZJ.Hi-Tech Park | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2405:1480::/32` | 0.000000% | SKBJNET; Beijing Sankuai Technology Co.,Ltd.; Wangjing International R&D Park Phase 3,No.6 Wangjing East Road,; Chaoyang District,Beijing 100102,PRC | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2405:1480::/32` | 0.000003% | SKBJNET; Beijing Sankuai Technology Co.,Ltd.; Wangjing International R&D Park Phase 3,No.6 Wangjing East Road,; Chaoyang District,Beijing 100102,PRC | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2405:7040::/32` | 0.000003% | COSCOSHIPPING; CHINA COSCO SHIPPING CORPORATION LIMITED | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| chinanet | `2405:7040::/32` | 0.000000% | COSCOSHIPPING; CHINA COSCO SHIPPING CORPORATION LIMITED | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| unicom | `2405:a900::/32` | 0.000003% | QIHOO; Beijing Qihu Technology Company Limited; 112 Room, D buliding , Deshengyuan square,; No.28 xinjiekouwaiwai,Xicheng District; Beijing,China | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| cmcc | `2407:6c40::/32` | 0.000000% | BJ-SHOUZIXIN; Beijing Shougang Automation Information Technology Co.,Ltd; Building 1, Yard 1, Shimen Road, Shijingshan, Beijing | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |
| cmcc | `2407:8f40::/32` | 0.000000% | CNIXP; ShenZhen QianHai New-Type Internet Exchange Point Co.,Ltd; Group A 4F, Qianhai Shenzhen-Hong Kong Innovation Center,; Menghai Rd. 4008, Qianhai Cooperation Zone, Shenzhen | ALLOCATED PORTABLE | Most-specific APNIC registration names an independent legal entity without operator attribution (`independent_legal_entity_patterns`) |

## other_operator：地址量前 100 项

| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 依据 |
| --- | --- | ---: | --- | --- | --- |
| — | — | — | 无 | — | — |

## other_or_unclassified：地址量前 100 项

| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 依据 |
| --- | --- | ---: | --- | --- | --- |
| chinanet | `2400:75a0::/28` | 0.024516% | DCCLGOOXNAHP-CN; Digital City Construction Leading Group Office of Xiongan New Area, Hebei Province; Digital City Construction Leading Group Office of Xiongan New Area, Hebei Province | ALLOCATED PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| unicom | `2402:18a0::/32` | 0.202218% | SAUNET; Shenyang Aerospace University; Shenyang Aerospace University, No.37 Daoyi South Avenue,; Daoyi District, Shenyang, Liaoning Province, China | ALLOCATED PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| unicom | `2402:ef40::/32` | 0.202218% | DRCSCNET; Development & Research Center of State Council Net.; No.225 Chaonei Street Dongcheng District Beijing China | ALLOCATED PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| chinanet | `240e:83:8000::/33` | 0.012162% | BeiJing-Telecom-UserAddress-lowguaranteed; BeiJing Telecom UserAddress for low-guaranteed | ALLOCATED NON-PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| chinanet | `240e:83::/34` | 0.004597% | BeiJing-Telecom-UserAddress-Highguaranteed; BeiJing Telecom UserAddress for High-guaranteed | ALLOCATED NON-PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| chinanet | `240e:83::/36` | 0.000766% | BeiJing-Telecom-UserAddress-Highguaranteed; BeiJing Telecom UserAddress for High-guaranteed | ASSIGNED NON-PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| unicom | `2400:89c0::/32` | 0.000049% | SINA; 15F,Ideal Plaza No.58 Bei Si Huan Xi Road Haidian District; Beijing,China,100080 | ALLOCATED PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| unicom | `2405:6f00::/32` | 0.000006% | SPINET; State Power Information Net; No.1,Baiguanglu No.2 Street,Xuanwu District,Beijing | ALLOCATED PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| chinanet | `2400:6e60::/32` | 0.000000% | SINOCHEM; Sinochem Corporation | ALLOCATED PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| unicom | `2403:4c80::/32` | 0.000003% | YWNET; Beijing Yiwangxin Technology Co;Ltd.; Unit2201, Zhongyu Plaza Number jia-6 Gongtibei Road; Chaoyang District, Beijing China | ALLOCATED PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| chinanet | `2403:f4c0::/32` | 0.000000% | AGRI-NET; Information center of MARA; No.11 Nongzhanguan Nanli,Chaoyang District,Beijing | ALLOCATED PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |
| chinanet | `240a:6000::/24` | 0.000000% | CWINET; Information Center Ministry of Water Resources; No2,Lane2,Baiguang Road,Xicheng District,Beijing; China Internet Network Information Center | ALLOCATED PORTABLE | Most-specific APNIC registration has no current strong operator or non-public classification |

## same_operator：地址量前 100 项

| 运营商 | APNIC 前缀 | 占运营商候选 | netname / description / org | status | 依据 |
| --- | --- | ---: | --- | --- | --- |
| cmcc | `2409:8000::/20` | 96.991219% | CMNET-V6-20110823; China Mobile Communications Corporation; Mobile Communications Network Operator in China; Internet Service Provider in China; China Mobile | ALLOCATED PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china mobile`) |
| chinanet | `240e::/18` | 31.177326% | CT-IPv6-Networks; Chinatelecom networks with tens of high-end routers and switches; Including users who access to Internet through Chinatelecom's networks.; China Telecom | ALLOCATED PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| unicom | `2408:8000::/20` | 99.592897% | CU-CN; China Unicom; No.21, Jin-Rong Street; Beijng 100033 | ALLOCATED PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: unicom`) |
| chinanet | `240e:b00::/24` | 6.276026% | CT-IPV6-BROADBAND-ADDRESS; Chinatelecom IPv6 address for fixed broadband | ALLOCATED NON-PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| chinanet | `240e:400::/24` | 6.276011% | CT-IPV6-MOBILE-ADDRESS; Chinatelecom IPv6 address for mobile | ALLOCATED NON-PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| chinanet | `240e:300::/24` | 6.269131% | CT-IPV6-BROADBAND-ADDRESS; Chinatelecom IPv6 address for fixed broadband | ALLOCATED NON-PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| chinanet | `240e:16::/31` | 0.047498% | CHINANET-SC-IPv6-USER-ADDRESS; CHINANET Sichuan province network; Data Communication Division; China Telecom | ALLOCATED NON-PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| chinanet | `2001:c68::/32` | 0.024516% | CHINANET-20020830; China Telecom; Internet Service Provider; Beijing,China | ALLOCATED PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| chinanet | `240e:ed::/32` | 0.024516% | China-Telecom-Jiangsu-province-network; China Telecom Jiangsu province network for customer | ALLOCATED NON-PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| chinanet | `2400:9380::/31` | 0.000001% | CTGI-AS-AP; China Telecom Global Limited; 38/F DAH SING Financial Center; 108 Gloucester Road Wan Chai; China Telecom Global Limited | ALLOCATED PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| chinanet | `2400:9380:8000::/44` | 0.000001% | CTGI-AS-AP; Chinatelecom global limit 38/F., DAH SING Financial Center, 108 Gloucester Road, Wan Chai, Hong Kong | ALLOCATED NON-PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| chinanet | `2400:9380:8020::/44` | 0.000000% | CTGI-AS-AP; Chinatelecom global litie Singpore site | ALLOCATED NON-PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
| chinanet | `2400:9380:8040::/44` | 0.000000% | CTGI-AS-AP; Chinatelecom global litie Japan site | ALLOCATED NON-PORTABLE | Most-specific APNIC registration is attributed to the current BGP operator (`description_patterns: china ?telecom`) |
