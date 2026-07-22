# IP2Location IPv6 省级审计

本报告仅用于人工核对，不参与三网 IPv6 正式准入。输入为当前已准入的活跃 BGP 前缀。

- 数据：IP2Location LITE DB11 IPv6 `2026.7.1`（CC BY-SA 4.0）
- 发布：`2026.07.04`
- 规则：每条原始 BGP 前缀独立匹配，不拆分、不聚合；首尾地址必须落在同一省级行政区，否则进入冲突项。

| 运营商 | 业务 | 省级代码 | 省份 | 前缀数 | /32 等价值 | 文件 |
|---|---|---|---|---:|---:|---|
| chinamobile | fixed_broadband | CN-AH | 安徽 | 229 | 1.886734009 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-AH.txt` |
| chinamobile | fixed_broadband | CN-BJ | 北京 | 450 | 5.124557495 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-BJ.txt` |
| chinamobile | fixed_broadband | CN-CQ | 重庆 | 1 | 1.000000000 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-CQ.txt` |
| chinamobile | fixed_broadband | CN-FJ | 福建 | 37 | 0.601669312 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-FJ.txt` |
| chinamobile | fixed_broadband | CN-GD | 广东 | 266 | 7.008331299 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-GD.txt` |
| chinamobile | fixed_broadband | CN-GS | 甘肃 | 57 | 3.019531250 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-GS.txt` |
| chinamobile | fixed_broadband | CN-GX | 广西 | 132 | 2.507812500 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-GX.txt` |
| chinamobile | fixed_broadband | CN-GZ | 贵州 | 75 | 1.574218750 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-GZ.txt` |
| chinamobile | fixed_broadband | CN-HA | 河南 | 195 | 8.730468750 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-HA.txt` |
| chinamobile | fixed_broadband | CN-HB | 湖北 | 43 | 3.781250000 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-HB.txt` |
| chinamobile | fixed_broadband | CN-HE | 河北 | 181 | 1.847656250 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-HE.txt` |
| chinamobile | fixed_broadband | CN-HL | 黑龙江 | 97 | 0.398437500 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-HL.txt` |
| chinamobile | fixed_broadband | CN-HN | 湖南 | 274 | 5.054687500 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-HN.txt` |
| chinamobile | fixed_broadband | CN-JL | 吉林 | 49 | 1.187500000 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-JL.txt` |
| chinamobile | fixed_broadband | CN-JS | 江苏 | 418 | 2.839843750 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-JS.txt` |
| chinamobile | fixed_broadband | CN-JX | 江西 | 88 | 0.679687500 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-JX.txt` |
| chinamobile | fixed_broadband | CN-LN | 辽宁 | 98 | 3.734375000 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-LN.txt` |
| chinamobile | fixed_broadband | CN-SC | 四川 | 73 | 1.644531250 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-SC.txt` |
| chinamobile | fixed_broadband | CN-SD | 山东 | 136 | 1.527343750 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-SD.txt` |
| chinamobile | fixed_broadband | CN-SH | 上海 | 520 | 1.082534790 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-SH.txt` |
| chinamobile | fixed_broadband | CN-SN | 陕西 | 4 | 1.011718750 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-SN.txt` |
| chinamobile | fixed_broadband | CN-SX | 山西 | 79 | 3.691406250 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-SX.txt` |
| chinamobile | fixed_broadband | CN-TJ | 天津 | 3 | 0.011718750 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-TJ.txt` |
| chinamobile | fixed_broadband | CN-YN | 云南 | 133 | 3.507812500 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-YN.txt` |
| chinamobile | fixed_broadband | CN-ZJ | 浙江 | 730 | 5.839843750 | `data/ipv6/provinces/chinamobile/fixed_broadband/CN-ZJ.txt` |
| chinamobile | mobile | CN-BJ | 北京 | 2883 | 46.060150146 | `data/ipv6/provinces/chinamobile/mobile/CN-BJ.txt` |
| chinamobile | mobile | CN-CQ | 重庆 | 5 | 1.000366211 | `data/ipv6/provinces/chinamobile/mobile/CN-CQ.txt` |
| chinamobile | mobile | CN-FJ | 福建 | 2 | 0.000030518 | `data/ipv6/provinces/chinamobile/mobile/CN-FJ.txt` |
| chinamobile | mobile | CN-GD | 广东 | 3 | 3.000000000 | `data/ipv6/provinces/chinamobile/mobile/CN-GD.txt` |
| chinamobile | mobile | CN-GX | 广西 | 8 | 2.000106812 | `data/ipv6/provinces/chinamobile/mobile/CN-GX.txt` |
| chinamobile | mobile | CN-HA | 河南 | 130 | 3.234375000 | `data/ipv6/provinces/chinamobile/mobile/CN-HA.txt` |
| chinamobile | mobile | CN-HE | 河北 | 196 | 2.047363281 | `data/ipv6/provinces/chinamobile/mobile/CN-HE.txt` |
| chinamobile | mobile | CN-HL | 黑龙江 | 7 | 3.000061035 | `data/ipv6/provinces/chinamobile/mobile/CN-HL.txt` |
| chinamobile | mobile | CN-JL | 吉林 | 2 | 1.000122070 | `data/ipv6/provinces/chinamobile/mobile/CN-JL.txt` |
| chinamobile | mobile | CN-JX | 江西 | 34 | 0.590118408 | `data/ipv6/provinces/chinamobile/mobile/CN-JX.txt` |
| chinamobile | mobile | CN-LN | 辽宁 | 848 | 3.897399902 | `data/ipv6/provinces/chinamobile/mobile/CN-LN.txt` |
| chinamobile | mobile | CN-SC | 四川 | 778 | 4.392974854 | `data/ipv6/provinces/chinamobile/mobile/CN-SC.txt` |
| chinamobile | mobile | CN-SD | 山东 | 30 | 0.664062500 | `data/ipv6/provinces/chinamobile/mobile/CN-SD.txt` |
| chinamobile | mobile | CN-SH | 上海 | 295 | 6.125030518 | `data/ipv6/provinces/chinamobile/mobile/CN-SH.txt` |
| chinamobile | mobile | CN-TJ | 天津 | 5 | 1.000518799 | `data/ipv6/provinces/chinamobile/mobile/CN-TJ.txt` |
| chinamobile | mobile | CN-ZJ | 浙江 | 26 | 0.101562500 | `data/ipv6/provinces/chinamobile/mobile/CN-ZJ.txt` |
| chinatelecom | fixed_broadband | CN-BJ | 北京 | 6 | 4.078125000 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-BJ.txt` |
| chinatelecom | fixed_broadband | CN-CQ | 重庆 | 1 | 4.000000000 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-CQ.txt` |
| chinatelecom | fixed_broadband | CN-FJ | 福建 | 23 | 0.406250000 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-FJ.txt` |
| chinatelecom | fixed_broadband | CN-GD | 广东 | 7 | 0.117202759 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-GD.txt` |
| chinatelecom | fixed_broadband | CN-GX | 广西 | 3 | 0.031265259 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-GX.txt` |
| chinatelecom | fixed_broadband | CN-HI | 海南 | 263 | 192.635375977 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-HI.txt` |
| chinatelecom | fixed_broadband | CN-JS | 江苏 | 568 | 0.085342407 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-JS.txt` |
| chinatelecom | fixed_broadband | CN-LN | 辽宁 | 16 | 4.234375000 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-LN.txt` |
| chinatelecom | fixed_broadband | CN-SD | 山东 | 13 | 4.125000000 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-SD.txt` |
| chinatelecom | fixed_broadband | CN-SH | 上海 | 7 | 6.000015259 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-SH.txt` |
| chinatelecom | fixed_broadband | CN-SX | 山西 | 5 | 8.000000000 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-SX.txt` |
| chinatelecom | fixed_broadband | CN-ZJ | 浙江 | 12 | 8.002685547 | `data/ipv6/provinces/chinatelecom/fixed_broadband/CN-ZJ.txt` |
| chinatelecom | mobile | CN-HI | 海南 | 1255 | 92.523208618 | `data/ipv6/provinces/chinatelecom/mobile/CN-HI.txt` |
| chinatelecom | mobile | CN-SH | 上海 | 3 | 3.000000000 | `data/ipv6/provinces/chinatelecom/mobile/CN-SH.txt` |
| chinatelecom | mobile | CN-TJ | 天津 | 1 | 4.000000000 | `data/ipv6/provinces/chinatelecom/mobile/CN-TJ.txt` |
| chinaunicom | fixed_broadband | CN-AH | 安徽 | 66 | 0.257812500 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-AH.txt` |
| chinaunicom | fixed_broadband | CN-BJ | 北京 | 52 | 4.758056641 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-BJ.txt` |
| chinaunicom | fixed_broadband | CN-CQ | 重庆 | 3 | 3.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-CQ.txt` |
| chinaunicom | fixed_broadband | CN-GD | 广东 | 49 | 5.169952393 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-GD.txt` |
| chinaunicom | fixed_broadband | CN-GX | 广西 | 3 | 3.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-GX.txt` |
| chinaunicom | fixed_broadband | CN-HA | 河南 | 1 | 1.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-HA.txt` |
| chinaunicom | fixed_broadband | CN-HB | 湖北 | 9 | 6.023437500 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-HB.txt` |
| chinaunicom | fixed_broadband | CN-HE | 河北 | 1 | 1.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-HE.txt` |
| chinaunicom | fixed_broadband | CN-HL | 黑龙江 | 24 | 21.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-HL.txt` |
| chinaunicom | fixed_broadband | CN-HN | 湖南 | 8 | 8.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-HN.txt` |
| chinaunicom | fixed_broadband | CN-JL | 吉林 | 4 | 4.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-JL.txt` |
| chinaunicom | fixed_broadband | CN-LN | 辽宁 | 93 | 24.281250000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-LN.txt` |
| chinaunicom | fixed_broadband | CN-SC | 四川 | 5 | 5.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-SC.txt` |
| chinaunicom | fixed_broadband | CN-SD | 山东 | 1 | 1.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-SD.txt` |
| chinaunicom | fixed_broadband | CN-SH | 上海 | 79 | 3.017822266 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-SH.txt` |
| chinaunicom | fixed_broadband | CN-SN | 陕西 | 37 | 18.000015259 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-SN.txt` |
| chinaunicom | fixed_broadband | CN-TJ | 天津 | 3 | 3.000000000 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-TJ.txt` |
| chinaunicom | fixed_broadband | CN-ZJ | 浙江 | 5 | 0.019531250 | `data/ipv6/provinces/chinaunicom/fixed_broadband/CN-ZJ.txt` |
| chinaunicom | mobile | CN-BJ | 北京 | 542 | 0.146972656 | `data/ipv6/provinces/chinaunicom/mobile/CN-BJ.txt` |
| chinaunicom | mobile | CN-CQ | 重庆 | 9 | 3.001464844 | `data/ipv6/provinces/chinaunicom/mobile/CN-CQ.txt` |
| chinaunicom | mobile | CN-GD | 广东 | 1 | 0.000244141 | `data/ipv6/provinces/chinaunicom/mobile/CN-GD.txt` |
| chinaunicom | mobile | CN-HL | 黑龙江 | 1623 | 66.366195679 | `data/ipv6/provinces/chinaunicom/mobile/CN-HL.txt` |
| chinaunicom | mobile | CN-SH | 上海 | 249 | 31.956237793 | `data/ipv6/provinces/chinaunicom/mobile/CN-SH.txt` |

## 未判定与冲突

共 `609` 个分类单元。完整事实见 JSON；以下最多列出 100 项。

| 运营商 | 业务 | 前缀 | 原因 | 首地址结果 | 末地址结果 |
|---|---|---|---|---|---|
| chinamobile | fixed_broadband | `2409:8a02::/32` | province_conflict_within_unit | CN/Beijing/Beijing | CN/Tianjin/Tianjin |
| chinamobile | fixed_broadband | `2409:8a04:400::/40` | province_conflict_within_unit | CN/Tianjin/Tianjin | CN/Hebei/Shijiazhuang |
| chinamobile | fixed_broadband | `2409:8a04::/32` | province_conflict_within_unit | CN/Tianjin/Tianjin | CN/Hebei/Cangzhou |
| chinamobile | fixed_broadband | `2409:8a0c:1000::/38` | province_conflict_within_unit | CN/Hebei/Cangzhou | CN/Shanxi/Taiyuan |
| chinamobile | fixed_broadband | `2409:8a0c::/32` | province_conflict_within_unit | CN/Hebei/Cangzhou | CN/Shanxi/Linfen |
| chinamobile | fixed_broadband | `2409:8a10:1000::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1100::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1200::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1300::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1400::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1500::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1600::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1700::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1800::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1900::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1a00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1b00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1c00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1d00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1e00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:1f00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2000::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2100::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2200::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2300::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2400::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2500::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2600::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2700::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2800::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:2900::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4200::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4300::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4400::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4500::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4600::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4700::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4800::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4900::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4a00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4b00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4c00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4d00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4e00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:4f00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5000::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5100::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5200::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5300::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5400::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5500::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5600::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5700::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5800::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5900::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5a00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5b00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5c00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5d00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5e00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:5f00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6000::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:600::/40` | unknown_province_name | CN/Shanxi/Linfen | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6100::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6200::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6300::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6400::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6500::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6600::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6700::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6800::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6900::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6a00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6b00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6c00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6d00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6e00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:6f00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7000::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:700::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7100::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7200::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7300::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7400::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7500::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7600::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7700::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7800::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7900::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7a00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7b00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7c00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7d00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7e00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:7f00::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:8000::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:800::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:8100::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:8200::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |
| chinamobile | fixed_broadband | `2409:8a10:8300::/40` | unknown_province_name | CN/Nei Mongol/Hohhot | CN/Nei Mongol/Hohhot |

## 许可与边界

省级结果由 IP2Location LITE 数据派生，遵循 CC BY-SA 4.0。该数据仅是待人工复核的第三方地理定位结论，不是三网省级编码的权威事实。
