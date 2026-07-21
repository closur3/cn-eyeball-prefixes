# gaoyifan `china6.txt` 事实审计

生成时间：`2026-07-21T18:07:37.4703219Z`

本报告只验证 `china6.txt` 的当前 BGP 可见性、国家登记边界和三网 IPv6 Origin 覆盖，不参与正式地址准入或排除。空间占比按精确 IPv6 地址数量计算，`/64 等价数`用于提高可读性。

## 总览

| 项目 | CIDR | /64 等价数 | 占 china6 空间 |
| --- | ---: | ---: | ---: |
| 规范化 china6 | 1594 | 109076809908224.0000 | 100.000000% |
| IPtoASN 当前可见 | 1546 | 108982769876992.0000 | 99.913785% |
| IPtoASN 未覆盖 | 73 | 94040031232.0000 | 0.086215% |
| RIS 当前可见 | 0 | 0.0000 | 0.000000% |
| RIS 未观测 | 1594 | 109076809908224.0000 | 100.000000% |
| IPtoASN 非 CN/未知 | 48 | 16501308063744.0000 | 15.128154% |

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

## 三网当前 Origin 对 china6 的覆盖

| 运营商 | 当前 Origin CIDR | 已在 china6 | china6 外 | china6 外 /64 等价数 |
| --- | ---: | ---: | ---: | ---: |
| chinanet | 3133 | 3090 | 43 | 4329177088.0000 |

chinanet 的 china6 外样本：`2001:df0:2c40::/48`, `2001:df5:900::/48`, `2001:df5:cd00::/48`, `2001:df6:4f00::/48`, `2400:9380:80c0::/44`, `2400:9380:81c0::/44`, `2400:9380:82c0::/44`, `2400:9380:83c0::/44`, `2400:9380:84c0::/44`, `2400:9380:85c0::/44`, `2400:9380:86c0::/44`, `2400:9380:87c0::/44`, `2400:9380:88c0::/44`, `2400:9380:89c0::/44`, `2400:9380:8ac0::/44`, `2400:9380:8bc0::/44`, `2400:9380:8cc0::/44`, `2400:9380:8dc0::/44`, `2400:9380:8ec0::/44`, `2400:9380:8fc0::/44`, `2400:9380:90c0::/44`, `2400:9380:91c0::/44`, `2400:9380:9265::/48`, `2400:9380:92c0::/44`, `2400:9380:93c0::/44`, `2400:9380:94c0::/44`, `2400:9380:95c0::/44`, `2400:9380:96c0::/44`, `2400:9380:97c0::/44`, `2400:9380:98c0::/44`

| cmcc | 358 | 352 | 6 | 8859549696.0000 |

cmcc 的 china6 外样本：`2001:67c:994::/48`, `2401:3000:a000::/36`, `2402:4f00::/32`, `2804:5bac::/48`, `2a03:ab80::/32`, `2a0c:9a40:8a10::/44`

| unicom | 4201 | 4161 | 40 | 13421903872.0000 |

unicom 的 china6 外样本：`2001:4b58::/32`, `2401:8a00::/44`, `2401:8a00:10::/48`, `2401:8a00:13::/48`, `2401:8a00:14::/46`, `2401:8a00:18::/45`, `2401:8a00:20::/43`, `2401:8a00:40::/42`, `2401:8a00:80::/41`, `2401:8a00:100::/40`, `2401:8a00:200::/39`, `2401:8a00:400::/38`, `2401:8a00:800::/37`, `2401:8a00:1000::/36`, `2401:8a00:2000::/35`, `2401:8a00:4000::/34`, `2401:8a00:8000::/33`, `2401:b0c0:1000::/36`, `2401:b0c0:a000::/48`, `2401:b0c0:b000::/48`, `2401:b0c0:c000::/48`, `2401:b0c0:d000::/48`, `2401:b0c0:e000::/48`, `2602:fe76::/36`, `2a02:6300::/43`, `2a02:6300:20::/44`, `2a02:6300:31::/48`, `2a02:6300:32::/47`, `2a02:6300:34::/46`, `2a02:6300:38::/45`


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

## RIS 未观测样本

- `2001:250::/30`
- `2001:254::/33`
- `2001:255::/32`
- `2001:256:100::/48`
- `2001:678:120::/48`
- `2001:678:53c::/48`
- `2001:678:10d0::/48`
- `2001:67c:c28::/48`
- `2001:c68::/32`
- `2001:cc0::/32`
- `2001:da8::/32`
- `2001:daa:1::/48`
- `2001:daa:2::/47`
- `2001:daa:4::/47`
- `2001:daa:6::/48`
- `2001:daa:9::/48`
- `2001:dc7::/32`
- `2001:dd8:1::/48`
- `2001:dd9::/48`
- `2001:df6:40::/48`
- `2001:4510:400::/40`
- `2001:4510:1480::/41`
- `2001:4511:1480::/41`
- `2400:1160::/32`
- `2400:3200::/32`
- `2400:5280:f803::/48`
- `2400:5a00::/32`
- `2400:6000::/32`
- `2400:6600::/32`
- `2400:6e60:1301::/48`

## 非 CN/未知样本

- `2401:2780::/32`
- `2401:d920::/48`
- `2402:1440::/39`
- `2402:7d80::/48`
- `2402:7d80:8888::/48`
- `2403:9b00:2000::/48`
- `2403:9b00:2400::/48`
- `2403:a100::/48`
- `2404:3700::/48`
- `2407:4980::/32`
- `240a:a080::/25`
- `240a:a100::/24`
- `240a:a200::/23`
- `240a:a480::/25`
- `240a:a500::/24`
- `240a:a600::/23`
- `240a:a800::/21`
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
