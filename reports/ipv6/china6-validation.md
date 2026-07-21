# gaoyifan `china6.txt` 事实审计

生成时间：`2026-07-21T18:49:36.267217287Z`

本报告只验证 `china6.txt` 的当前 BGP 可见性、国家登记边界和三网 IPv6 Origin 覆盖，不参与正式地址准入或排除。空间占比按精确 IPv6 地址数量计算，`/64 等价数`用于提高可读性。

## 总览

| 项目 | CIDR | /64 等价数 | 占 china6 空间 |
| --- | ---: | ---: | ---: |
| 规范化 china6 | 1594 | 109076809908224.0000 | 100.000000% |
| IPtoASN 当前可见 | 1546 | 108982769876992.0000 | 99.913785% |
| IPtoASN 未覆盖 | 73 | 94040031232.0000 | 0.086215% |
| RIS 当前可见 | 1594 | 109076809908224.0000 | 100.000000% |
| RIS 未观测 | 0 | 0.0000 | 0.000000% |
| IPtoASN 明确非 CN Origin | 37 | 10092544.0000 | 0.000009% |
| IPtoASN 国家未知 | 11 | 16501297971200.0000 | 15.128145% |

原始 CIDR：**1594**；规范化后 CIDR：**1594**。

## IPtoASN 国家字段

| 国家/地区 | CIDR | /64 等价数 | 占比 |
| --- | ---: | ---: | ---: |
| CN | 1514 | 92481461813248.0000 | 84.785631% |
| HK | 3 | 196608.0000 | 0.000000% |
| PL | 20 | 7929856.0000 | 0.000007% |
| RO | 1 | 65536.0000 | 0.000000% |
| UNKNOWN | 11 | 16501297971200.0000 | 15.128145% |
| US | 13 | 1900544.0000 | 0.000002% |

## APNIC delegated 登记

| 登记国家/地区 | CIDR | /64 等价数 | 占比 |
| --- | ---: | ---: | ---: |
| CN | 1272 | 109050324254720.0000 | 99.975718% |
| HK | 27 | 2686976.0000 | 0.000002% |
| SG | 1 | 65536.0000 | 0.000000% |
| UNREGISTERED_OR_NON_APNIC | 294 | 26482900992.0000 | 0.024279% |

## APNIC 登记 CN 的三网当前 Origin 对 china6 的覆盖

| 运营商 | 当前 Origin CIDR | 已在 china6 | china6 外 | china6 外 /64 等价数 |
| --- | ---: | ---: | ---: | ---: |
| chinanet | 3067 | 3067 | 0 | 0.0000 |
| cmcc | 352 | 352 | 0 | 0.0000 |
| unicom | 4161 | 4161 | 0 | 0.0000 |

## gaoyifan 三网 IPv6 分表与当前 Origin

分表只作为独立交叉验证。`一致率`以相应上游分表为分母；`覆盖率`以 `APNIC delegated CN + 当前同运营商 Origin` 为分母。

| 运营商 | 上游 CIDR | 当前 Origin 一致 | 一致率 | 当前 Origin 覆盖率 | 另一运营商 Origin | 无当前三网 Origin | china6 外 | 当前 Origin 漏项 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| chinanet | 854 | 3067 | 78.115950% | 100.000000% | 1 | 1620 | 22 | 0 |
| cmcc | 128 | 352 | 50.715132% | 100.000000% | 2 | 91 | 1 | 0 |
| unicom | 1326 | 4161 | 10.510140% | 100.000000% | 35 | 2084 | 0 | 0 |

## chinanet：另一运营商 Origin 冲突样本

- `2407:8f40:2::/48`

## chinanet：无当前三网 Origin 样本

- `2001:250:2::/48`
- `2001:250:205::/48`
- `2001:250:206::/47`
- `2001:250:208::/46`
- `2001:250:20c::/48`
- `2001:250:20f::/48`
- `2001:250:210::/48`
- `2001:250:214::/47`
- `2001:250:217::/48`
- `2001:250:218::/46`
- `2001:250:21e::/47`
- `2001:250:223::/48`
- `2001:250:228::/48`
- `2001:250:22b::/48`
- `2001:250:22d::/48`
- `2001:250:22e::/48`
- `2001:250:230::/47`
- `2001:250:234::/48`
- `2001:250:236::/47`
- `2001:250:238::/48`

## cmcc：另一运营商 Origin 冲突样本

- `2405:6f00:c101::/48`
- `2405:6f00:c170::/48`

## cmcc：无当前三网 Origin 样本

- `2400:9020:f012::/47`
- `2400:95e0::/48`
- `2400:a860:1::/48`
- `2400:a860:2::/47`
- `2400:a860:4::/47`
- `2400:a860:6::/48`
- `2400:ae00:1981::/48`
- `2400:ee00:ffec::/46`
- `2400:ee00:fff0::/44`
- `2401:2a00:f000::/43`
- `2401:9a00::/44`
- `2401:9a00:10::/46`
- `2401:c020:6::/48`
- `2401:c020:8::/47`
- `2401:c020:14::/48`
- `2402:1440:30::/48`
- `2402:1440:200::/39`
- `2402:1440:400::/38`
- `2402:1440:800::/37`
- `2402:1440:1000::/36`

## unicom：另一运营商 Origin 冲突样本

- `2401:a140:1::/48`
- `2402:92c0::/48`
- `2403:1ec0:1610::/48`
- `2403:a200:a1ff::/48`
- `2405:1480:1000::/48`
- `2407:8f40:2::/48`
- `2409:27fb::/48`
- `240e:b:e000::/40`
- `240e:b:f800::/37`
- `240e:c:800::/37`
- `240e:c:1800::/37`
- `240e:4b:c000::/37`
- `240e:4c::/36`
- `240e:b1:8800::/37`
- `240e:b1:9800::/37`
- `240e:108:10f3::/48`
- `240e:108:10f4::/48`
- `240e:23d::/34`
- `240e:93c:1000::/36`
- `240e:93d::/36`

## unicom：无当前三网 Origin 样本

- `2001:250:2::/48`
- `2001:250:205::/48`
- `2001:250:206::/47`
- `2001:250:208::/46`
- `2001:250:20c::/48`
- `2001:250:20f::/48`
- `2001:250:210::/48`
- `2001:250:214::/47`
- `2001:250:218::/47`
- `2001:250:21b::/48`
- `2001:250:21e::/47`
- `2001:250:223::/48`
- `2001:250:228::/48`
- `2001:250:22b::/48`
- `2001:250:22d::/48`
- `2001:250:22e::/48`
- `2001:250:230::/47`
- `2001:250:234::/48`
- `2001:250:236::/47`
- `2001:250:238::/48`

## china6 内地址量最大的 Origin

| ASN | 国家/地区 | 运营商识别 | /64 等价数 | 占比 | 描述 |
| --- | --- | --- | ---: | ---: | --- |
| AS137726 | CN | — | 17592186044416.0000 | 16.128255% | SINOPEC-NET China Petroleum & Chemical Corporation |
| AS23910 | CN | — | 17433159925760.0000 | 15.982462% | CNGI-CERNET2-AS-AP China Next Generation Internet CERNET2 |
| AS4134 | CN | chinanet | 17106975391744.0000 | 15.683421% | CHINANET-BACKBONE No.31,Jin-rong Street |
| AS9808 | CN | cmcc | 16131965255680.0000 | 14.789546% | CHINAMOBILE-CN China Mobile Communications Group Co., Ltd. |
| AS133111 | CN | — | 8830436048896.0000 | 8.095613% | CNT-NORTHCHINA CERNET New Technology Co., Ltd |
| AS37963 | CN | — | 4423815135232.0000 | 4.055688% | ALIBABA-CN-NET Hangzhou Alibaba Advertising Co.,Ltd. |
| AS38365 | CN | — | 4402339905536.0000 | 4.036000% | BAIDU Beijing Baidu Netcom Science and Technology Co., Ltd. |
| AS4837 | CN | unicom | 1709288128512.0000 | 1.567050% | CHINA169-BACKBONE CHINA UNICOM China169 Backbone |
| AS24445 | CN | cmcc | 296385970176.0000 | 0.271722% | CMNET-V4HENAN-AS-AP Henan Mobile Communications Co.,Ltd |
| AS56040 | CN | cmcc | 280288493568.0000 | 0.256964% | CMNET-GUANGDONG-AP China Mobile communications corporation |
| AS4812 | CN | chinanet | 208003989504.0000 | 0.190695% | CHINANET-SH-AP China Telecom Group |
| AS17816 | CN | unicom | 179605864448.0000 | 0.164660% | CHINA169-GZ China Unicom IP network China169 Guangdong province |
| AS56041 | CN | cmcc | 163423518720.0000 | 0.149824% | CMNET-ZHEJIANG-AP China Mobile communications corporation |
| AS56042 | CN | cmcc | 154721517568.0000 | 0.141846% | CMNET-SHANXI-AP China Mobile communications corporation |
| AS24400 | CN | cmcc | 154638876672.0000 | 0.141771% | CMNET-V4SHANGHAI-AS-AP Shanghai Mobile Communications Co.,Ltd. |
| AS56044 | CN | cmcc | 153662717952.0000 | 0.140876% | CMNET-AS-LIAONING China Mobile communications corporation |
| AS56047 | CN | cmcc | 133917507584.0000 | 0.122774% | CMNET-HUNAN-AP China Mobile communications corporation |
| AS132525 | CN | cmcc | 128884408320.0000 | 0.118159% | CMNET-HEILONGJIANG-CN HeiLongJiang Mobile Communication Company Limited |
| AS56046 | CN | cmcc | 115449987072.0000 | 0.105843% | CMNET-JIANGSU-AP China Mobile communications corporation |
| AS4808 | CN | unicom | 113266655232.0000 | 0.103841% | CHINA169-BJ China Unicom Beijing Province Network |
| AS24547 | CN | cmcc | 103231193088.0000 | 0.094641% | CMNET-V4HEBEI-AS-AP Hebei Mobile Communication Company Limited |
| AS17621 | CN | unicom | 102155812864.0000 | 0.093655% | CNCGROUP-SH China Unicom Shanghai network |
| AS134810 | CN | cmcc | 96672415744.0000 | 0.088628% | CMNET-JILIN-AS-AP China Mobile Group JiLin communications corporation |
| AS7497 | CN | — | 77309411328.0000 | 0.070876% | CSTNET-AS-AP Computer Network Information Center of Chinese Academy of Sciences CNIC-CAS |
| AS59016 | CN | — | 73014444032.0000 | 0.066939% | HACN China Broadcast Henan Network Co., Ltd |
| AS24363 | CN | — | 68719542272.0000 | 0.063001% | CNGI-JNN-IX-AS-AP CERNET2 IX at Shandong University |
| AS56048 | CN | cmcc | 64846561280.0000 | 0.059450% | CMNET-BEIJING-AP China Mobile Communicaitons Corporation |
| AS140061 | CN | chinanet | 63100026880.0000 | 0.057849% | CHINANET-QINGHAI-AS-AP Qinghai Telecom |
| AS24444 | CN | cmcc | 59961835520.0000 | 0.054972% | CMNET-V4SHANDONG-AS-AP Shandong Mobile Communication Company Limited |
| AS56045 | CN | cmcc | 56108580864.0000 | 0.051440% | CMNET-JIANGXI-AP China Mobile communications corporation |
| AS38019 | CN | cmcc | 43690033152.0000 | 0.040054% | CMNET-V4TIANJIN-AS-AP tianjin Mobile Communication Company Limited |
| AS24353 | CN | — | 30071193600.0000 | 0.027569% | CNGI-XA-IX-AS-AP CERNET2 IX at Xian Jiaotong University |
| AS138371 | CN | — | 30066606080.0000 | 0.027565% | CNGI-QDA-IX-AS-AP CERNET2 regional IX at Ocean University of China |
| AS17638 | CN | chinanet | 25250299904.0000 | 0.023149% | CHINATELECOM-TJ-AS-AP ASN for TIANJIN Provincial Net of CT |
| AS17429 | CN | — | 23102226432.0000 | 0.021180% | BGCTVNET BEIJING GEHUA CATV NETWORK CO.LTD |
| AS23724 | CN | 显式排除 | 22268411904.0000 | 0.020415% | CHINANET-IDC-BJ-AP IDC, China Telecommunications Corporation |
| AS24360 | CN | — | 21488140288.0000 | 0.019700% | CNGI-ZHZ-IX-AS-AP CERNET2 IX at Zhengzhou University |
| AS24361 | CN | — | 21487222784.0000 | 0.019699% | CNGI-NJ-IX-AS-AP CERNET2 IX at Southeast University |
| AS55990 | CN | — | 16240476160.0000 | 0.014889% | HWCSNET Huawei Cloud Service data center |
| AS140726 | CN | unicom | 15431892992.0000 | 0.014148% | UNICOM-HEFEI-MAN UNICOM AnHui province network |
| AS58542 | CN | chinanet | 13455458304.0000 | 0.012336% | CHINATELECOM-TIANJIN Tianjij,300000 |
| AS139791 | CN | — | 13086294016.0000 | 0.011997% | WOPAI-AS-AP Langfang Wopai Communications Co Ltd |
| AS134774 | CN | chinanet | 12976128000.0000 | 0.011896% | CHINANET-GUANGDONG-SHENZHEN-MAN CHINANET Guangdong province Shenzhen MAN network |
| AS140329 | CN | chinanet | 12893487104.0000 | 0.011821% | CHINATELECOM-FUJIAN-FUZHOU-5G-NETWORK CHINATELECOM Fujian province Fuzhou 5G network |
| AS24355 | CN | — | 12892569600.0000 | 0.011820% | CNGI-CD-IX-AS-AP CERNET2 IX at University of Electronic Science and Technology of China |
| AS24352 | CN | — | 12888113152.0000 | 0.011816% | CNGI-TJN-IX-AS-AP CERNET2 IX at Tianjin University |
| AS24348 | CN | — | 12886736896.0000 | 0.011814% | CNGI-BJ-IX2-AS-AP CERNET2 IX at Tsinghua University |
| AS45062 | CN | — | 12884901888.0000 | 0.011813% | NETEASE-NETWORK NetEase Building No.16 Ke Yun Road |
| AS38283 | CN | 显式排除 | 9933684736.0000 | 0.009107% | CHINANET-SCIDC-AS-AP CHINANET SiChuan Telecom Internet Data Center |
| AS4809 | CN | 显式排除 | 9880076288.0000 | 0.009058% | CHINATELECOM-CORE-WAN-CN2 China Telecom Next Generation Carrier Network |

## china6 内明确非 CN Origin

| ASN | 国家/地区 | /64 等价数 | 占比 | 描述 |
| --- | --- | ---: | ---: | --- |
| AS214481 | PL | 7929856.0000 | 0.000007% | WCZAPKOWICZ-AS as214481as-wczapkowicz Chunkserve Backbone Transit |
| AS214773 | US | 1638400.0000 | 0.000002% | NEXT-ALAN-AS |
| AS131314 | HK | 131072.0000 | 0.000000% | CBC-AS-AP CBC Tech International Limited |
| AS8075 | US | 131072.0000 | 0.000000% | MICROSOFT-CORP-MSN-AS-BLOCK |
| AS150706 | HK | 65536.0000 | 0.000000% | HKZTCL-AS-AP Hong Kong Zhengxing Technology Co., Ltd. |
| AS198304 | US | 65536.0000 | 0.000000% | GUGENET |
| AS21859 | US | 65536.0000 | 0.000000% | ZEN-ECN |
| AS61120 | RO | 65536.0000 | 0.000000% | VRANBIT Vranbit SRL |

## IPtoASN 未覆盖样本

- `2001:4510:1480::/41`
- `2001:4511:1480::/41`
- `2400:73e0::/32`
- `2400:8fc0::/38`
- `2400:8fc0:400::/40`
- `2400:8fc0:500::/42`
- `2400:8fc0:540::/43`
- `2400:8fc0:560::/44`
- `2400:8fc0:570::/48`
- `2400:8fc0:572::/47`
- `2400:8fc0:574::/46`
- `2400:8fc0:578::/45`
- `2400:8fc0:580::/41`
- `2400:8fc0:600::/39`
- `2400:8fc0:800::/37`
- `2400:8fc0:1000::/36`
- `2400:8fc0:2000::/35`
- `2400:8fc0:4000::/34`
- `2400:8fc0:8000::/33`
- `2400:b620::/48`
- `2400:e680::/32`
- `2401:7700::/32`
- `2401:ca00::/32`
- `2402:1440:200::/39`
- `2402:1440:400::/38`
- `2402:1440:800::/37`
- `2402:1440:1000::/36`
- `2402:1440:2000::/35`
- `2402:1440:4000::/34`
- `2402:1440:8000::/33`

## 明确非 CN Origin 样本

- `2402:7d80::/48`
- `2402:7d80:8888::/48`
- `2403:9b00:2000::/48`
- `2403:9b00:2400::/48`
- `2403:a100::/48`
- `2404:3700::/48`
- `2a0f:9400:6110::/48`
- `2a14:7581:ffb::/48`
- `2a14:7581:3101::/48`
- `2a14:7583:f220::/43`
- `2a14:7583:f240::/42`
- `2a14:7583:f300::/46`
- `2a14:7583:f304::/47`
- `2a14:7583:f306::/48`
- `2a14:7583:f411::/48`
- `2a14:7583:f4f0::/47`
- `2a14:7583:f4f4::/48`
- `2a14:7583:f4fe::/48`
- `2a14:7583:f500::/48`
- `2a14:7583:f701::/48`
- `2a14:7583:f703::/48`
- `2a14:7583:f704::/47`
- `2a14:7583:f707::/48`
- `2a14:7583:f708::/47`
- `2a14:7583:f70b::/48`
- `2a14:7583:f70c::/48`
- `2a14:7583:f743::/48`
- `2a14:7583:f744::/48`
- `2a14:7583:f764::/48`
- `2a14:7586:6100::/48`

## IPtoASN 国家未知样本

- `2401:2780::/32`
- `2401:d920::/48`
- `2402:1440::/39`
- `2407:4980::/32`
- `240a:a080::/25`
- `240a:a100::/24`
- `240a:a200::/23`
- `240a:a480::/25`
- `240a:a500::/24`
- `240a:a600::/23`
- `240a:a800::/21`
