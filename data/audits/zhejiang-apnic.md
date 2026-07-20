# 浙江 IPv4 APNIC 登记事实审计

本报告用于复核浙江正向准入 ACL：只有当前 BGP Origin 属于三网、且 APNIC 最具体登记可明确归属同一家运营商的地址才会保留。完整逐地址事实保存在 [`zhejiang-apnic.json.gz`](./zhejiang-apnic.json.gz)。

## 总览

| 指标 | 数值 |
|---|---:|
| 最大聚合 ACL CIDR | 14,577 |
| IPv4 地址 | 15,325,043 |
| 最具体 APNIC 事实片段 | 16,947 |
| APNIC 登记覆盖 | 15,325,043（100.0000%） |
| 构建规则仍识别出的强非公众信号 | 0 |

## 前缀清洗前后对照

这里的“准入前候选”指已满足三网 Origin 与中国边界、但尚未执行云 CIDR、APNIC、route、MOAS 排除及浙江同运营商 APNIC 登记准入的地址。分类地址数按证据行累加，多个上游命中同一地址时可能重复；总未准入地址数按地址并集计算。

| 阶段 | 地址 |
|---|---:|
| 准入前候选 | 16,049,454 |
| 未准入（并集） | 724,411 |
| 最终保留 | 15,325,043 |

| 排除类别 | 证据范围 | 地址（可重复） |
|---|---:|---:|
| `apnic_delegated_holder` | 1 | 4 |
| `apnic_independent_legal_entity_holder` | 13 | 5,120 |
| `apnic_inetnum` | 399 | 46,907 |
| `apnic_operator_admission_independent_legal_entity` | 41,693 | 267,807 |
| `apnic_operator_admission_operator_registration_conflict` | 23 | 364 |
| `apnic_operator_admission_other_registration` | 60,498 | 387,825 |
| `apnic_portable_holder` | 20 | 14,336 |
| `cloud_provider_cidr` | 2 | 2,048 |

### 排除证据样本

| CIDR | 地址 | 类别 | 运营商 / ASN | APNIC 登记主体 | 原因 |
|---|---:|---|---|---|---|
| `43.240.72.0/23` | 512 | `apnic_inetnum` | `cmcc / AS56041` | Zhejiang zhi cloud information technology co., LTD | APNIC inetnum registration explicitly identifies a cloud-company or cloud-service address range |
| `43.240.204.0/22` | 1,024 | `apnic_portable_holder` | `unicom / AS4837` | Hangzhou Sulian Information Technology Co.,ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `45.250.32.0/22` | 1,024 | `apnic_portable_holder` | `unicom / AS4837` | Hangzhou Sulian Information Technology Co.,ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `45.250.36.0/22` | 1,024 | `apnic_portable_holder` | `unicom / AS4837` | Hangzhou Sulian Information Technology Co.,ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `45.250.40.0/22` | 1,024 | `apnic_portable_holder` | `cmcc / AS56041` | Hangzhou Sulian Information Technology Co.,ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `45.252.0.0/22` | 1,024 | `cloud_provider_cidr` | `unicom / AS4837` | ucloud | Prefix is explicitly listed by IP-Data for ucloud |
| `45.254.48.0/23` | 512 | `apnic_inetnum` | `unicom / AS4837` | Guangzhou NetEase Computer System Co., Ltd | APNIC inetnum registration explicitly identifies a NetEase corporate or service network |
| `60.12.7.224/27` | 32 | `apnic_inetnum` | `unicom / AS4837` | IDCDI2PIGONGXIANGJIERUFUWUQIDIZHIDUAN,HANGZHOU,ZHEJIANG | APNIC inetnum registration explicitly identifies a dedicated IDC resource or service range |
| `60.12.9.144/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | WANGYINHULIANKEJIYOUXIANGONGSI,HANGZHOU,ZHEJIANG | APNIC inetnum registration explicitly identifies a Wangyin Hulian or Netbank Interlink corporate network |
| `60.12.17.120/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.36.140/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | TangLiYou_IDC,ZheJiang,Wenzhou | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.72.240/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | TENGYOUIDC,LISHUI,ZHEJIANG | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.148.0/22` | 1,024 | `apnic_inetnum` | `unicom / AS4837` | IDCBEIYONG,JINHUA,ZHEJIANG | APNIC inetnum registration explicitly identifies a dedicated IDC resource or service range |
| `60.12.174.64/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | GAMEIDC,TAIZHOU,ZHEJIANG | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.0/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.16/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.32/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.36/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.40/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.96/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.112/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.128/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.132/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.136/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.140/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.160/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.164/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.168/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.176/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.192/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.196/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.200/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.204/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.208/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.212/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.216/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.220/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.192.224/27` | 32 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.0/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.16/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.20/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.24/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.32/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.48/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.64/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.80/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.84/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.88/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.92/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.96/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.112/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.128/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.144/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.160/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.168/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.176/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.184/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.188/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.193.192/26` | 64 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.0/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.16/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.32/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.48/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.128/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.144/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.152/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.160/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.176/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.188/30` | 4 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.192/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.208/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.216/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.194.224/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.196.0/24` | 256 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.197.0/24` | 256 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.198.0/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.198.16/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.199.0/24` | 256 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.200.0/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.200.16/28` | 16 | `apnic_inetnum` | `unicom / AS4837` | CHINA169NINGBOIDCChinaunicomNingboChina | APNIC inetnum registration explicitly identifies an IDC network |
| `60.12.227.0/24` | 256 | `apnic_inetnum` | `unicom / AS4837` | SHANGHAIWANGYI,HANGZHOU,ZHEJIANG | APNIC inetnum registration explicitly identifies a NetEase corporate or service network |
| `60.12.230.224/29` | 8 | `apnic_inetnum` | `unicom / AS4837` | WANGLUOYINGPANHEZUOIDCZIYUAN,HANGZHOU,ZHEJIANG | APNIC inetnum registration explicitly identifies a dedicated IDC resource or service range |
| `60.190.56.0/24` | 256 | `apnic_inetnum` | `chinanet / AS4134` | Ningbo Municipal People's Government | APNIC inetnum registration explicitly identifies an electronic-government or government data network |
| `60.190.57.0/24` | 256 | `apnic_inetnum` | `chinanet / AS4134` | Ningbo Municipal People's Government | APNIC inetnum registration explicitly identifies an electronic-government or government data network |
| `60.190.64.64/26` | 64 | `apnic_inetnum` | `chinanet / AS4134` | WenZhou Telecommunication Co.,Ltd Data Centre | APNIC inetnum registration explicitly identifies a data-center network |
| `60.190.65.96/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | Wenzhou Telecommunications Co.,ltd Data Centre | APNIC inetnum registration explicitly identifies a data-center network |
| `60.190.65.152/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Wenzhou E-government Centre | APNIC inetnum registration explicitly identifies a government private network |
| `60.190.74.144/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | China Telecom Yongjia branch (video surveillance, cloud platform storage) | APNIC inetnum registration explicitly identifies a cloud-service network |
| `60.190.79.128/25` | 128 | `apnic_inetnum` | `chinanet / AS4134` | WenZhou Telecommunication Co.,Ltd Data Centre | APNIC inetnum registration explicitly identifies a data-center network |
| `60.190.94.192/26` | 64 | `apnic_inetnum` | `chinanet / AS4134` | WenZhou Telecommunication Co.,Ltd Data Centre | APNIC inetnum registration explicitly identifies a data-center network |
| `60.190.104.128/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | pingyang-dianxin-IDC-jifang1 | APNIC inetnum registration explicitly identifies an IDC network |
| `60.190.104.160/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | pingyang-dianxin-IDC-jifang2 | APNIC inetnum registration explicitly identifies an IDC network |
| `60.190.104.192/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | pingyang-dianxin-IDC-jifang | APNIC inetnum registration explicitly identifies an IDC network |
| `60.190.104.224/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | wenzhou-pingyang-dianxin-IDC-jifang | APNIC inetnum registration explicitly identifies an IDC network |
| `60.190.112.0/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | WenZhou Telecommunication Co.,Ltd Data Centre | APNIC inetnum registration explicitly identifies a data-center network |
| `60.190.125.184/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Lishui Telecom Computer Room | APNIC inetnum registration explicitly identifies a hosted-server or server-room network |
| `60.190.141.96/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | Haining Telecom Co.,LTD IDC Computer lab | APNIC inetnum registration explicitly identifies an IDC network |
| `60.190.154.192/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | ZheJiang TongXiang Telecom IDC Machine room CO.,LTD | APNIC inetnum registration explicitly identifies an IDC network |
| `60.190.228.56/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou Fu Tong Cloud Computing Technology Co., Ltd. | APNIC inetnum registration explicitly identifies a cloud-computing network |
| `60.190.231.200/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou Huawei 3COM Technology Co.,Ltd | APNIC inetnum registration explicitly identifies a qualified Huawei, Baidu, or Alibaba corporate network |
| `60.190.232.0/24` | 256 | `apnic_inetnum` | `chinanet / AS4134` | Alibaba Com (China) Technology Co.,ltd. | APNIC inetnum registration explicitly identifies a qualified Huawei, Baidu, or Alibaba corporate network |
| `60.190.241.0/24` | 256 | `apnic_inetnum` | `chinanet / AS4134` | Alibaba Com (China) Technology Co.,ltd. | APNIC inetnum registration explicitly identifies a qualified Huawei, Baidu, or Alibaba corporate network |
| `60.190.246.240/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou Huawei 3COM Technology Co.,Ltd | APNIC inetnum registration explicitly identifies a qualified Huawei, Baidu, or Alibaba corporate network |
| `60.191.21.56/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Study of Hangzhou City, network security | APNIC inetnum registration explicitly identifies a security or DDoS network |
| `60.191.43.96/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou Telecommunication IDC Center | APNIC inetnum registration explicitly identifies an IDC network |
| `60.191.45.16/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Yanjun Zhuang | APNIC inetnum registration explicitly identifies an IDC network |
| `60.191.47.0/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | ZheJiang Province Telecom Co.,Ltd. Linan Filiale | APNIC inetnum registration explicitly identifies an IDC network |
| `60.191.47.32/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | ZheJiang Province Telecom Co.,Ltd. Linan Filiale | APNIC inetnum registration explicitly identifies an IDC network |
| `60.191.68.224/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou Huawei 3COM Technology Co.,Ltd | APNIC inetnum registration explicitly identifies a qualified Huawei, Baidu, or Alibaba corporate network |
| `60.191.71.32/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Zhejiang Provincial Bureau of Data | APNIC inetnum registration explicitly identifies an electronic-government or government data network |
| `60.191.93.16/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Zhejiang huatong cloud data technology co., LTD | APNIC inetnum registration explicitly identifies a cloud-service network |
| `60.191.93.160/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou beat cloud Technology Co. Ltd. | APNIC inetnum registration explicitly identifies a cloud-company or cloud-service address range |
| `60.191.99.128/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou Huawei 3COM Technology Co.,Ltd | APNIC inetnum registration explicitly identifies a qualified Huawei, Baidu, or Alibaba corporate network |
| `60.191.100.96/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou beat cloud Technology Co. Ltd. | APNIC inetnum registration explicitly identifies a cloud-company or cloud-service address range |
| `60.191.100.152/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | IDC | APNIC inetnum registration explicitly identifies an IDC network |
| `60.191.101.160/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou beat cloud Technology Co. Ltd. | APNIC inetnum registration explicitly identifies a cloud-company or cloud-service address range |
| `60.191.108.96/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Huawei Software Technology Co., Ltd. | APNIC inetnum registration explicitly identifies a qualified Huawei, Baidu, or Alibaba corporate network |
| `60.191.123.0/25` | 128 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou Huawei 3COM Technology Co.,Ltd | APNIC inetnum registration explicitly identifies a qualified Huawei, Baidu, or Alibaba corporate network |
| `60.191.128.208/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | New Machine Room KVM System | APNIC inetnum registration explicitly identifies a hosted-server or server-room network |
| `60.191.128.252/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | guoshui VPN | APNIC inetnum registration explicitly identifies an MPLS or VPN network |
| `60.191.133.96/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | taizhou telecom IDC | APNIC inetnum registration explicitly identifies an IDC network |
| `60.191.144.16/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | IDC2-C5 Cabinet retail | APNIC inetnum registration explicitly identifies a dedicated IDC resource or service range |
| `60.191.151.32/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Bureau Party IDC light collecting system | APNIC inetnum registration explicitly identifies an IDC network |
| `60.191.187.64/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Taizhou CDN Network MSC Server | APNIC inetnum registration explicitly identifies a CDN network |
| `60.191.194.12/30` | 4 | `apnic_delegated_holder` | `chinanet / AS4134` | Star Internet | Most-specific APNIC non-portable registration is linked to a currently active independent ASN |
| `61.130.72.128/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | quzhou telecom IDC | APNIC inetnum registration explicitly identifies an IDC network |
| `61.130.72.160/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | quzhou telecom IDC | APNIC inetnum registration explicitly identifies an IDC network |
| `61.130.72.192/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | quzhou telecom IDC | APNIC inetnum registration explicitly identifies an IDC network |
| `61.130.72.224/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | quzhou telecom IDC | APNIC inetnum registration explicitly identifies an IDC network |
| `61.130.97.160/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | HangZhou Netbank Interlink Technolgies CO.,LTD | APNIC inetnum registration explicitly identifies a Wangyin Hulian or Netbank Interlink corporate network |
| `61.130.99.64/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | Shanghai Wangsu Science and Technology Co.,Ltd | APNIC inetnum registration explicitly identifies a known cloud or CDN brand |
| `61.130.101.188/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Yinjiao OA Address | APNIC inetnum registration explicitly identifies an office-automation network |
| `61.130.104.52/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | SAMSUNG ELECTRONICS CO.,LTD NINGBO VENDITION SERVER CENTRE | APNIC inetnum registration explicitly identifies a hosted-server or server-room network |
| `61.130.149.208/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Jinhua Market Develop Server Center | APNIC inetnum registration explicitly identifies a hosted-server or server-room network |
| `61.153.3.0/24` | 256 | `apnic_inetnum` | `chinanet / AS4134` | Hangzhou Telecommunication IDC Center | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.31.96/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | Wenzhou Telecommunications Co.,ltd Data Centre | APNIC inetnum registration explicitly identifies a data-center network |
| `61.153.36.96/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | Zhoushan Telecom Corp Data Center | APNIC inetnum registration explicitly identifies a data-center network |
| `61.153.37.244/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Zhoushan Telecom Corp Data Center | APNIC inetnum registration explicitly identifies a data-center network |
| `61.153.40.136/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | TaiZhou Telecom Data Center | APNIC inetnum registration explicitly identifies a data-center network |
| `61.153.40.200/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | TaiZhou Telecom Data Center | APNIC inetnum registration explicitly identifies a data-center network |
| `61.153.40.224/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | TaiZhou Telecom Data Center | APNIC inetnum registration explicitly identifies a data-center network |
| `61.153.54.160/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | quzhou telecom IDC | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.62.152/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Xin Shi Tong | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.67.128/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Chuang Jia | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.67.180/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Shang Mao | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.68.24/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Hui Yuan Bin Guan | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.71.40/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Hua Shu Shu Zi | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.71.240/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Shi Zheng Fu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.72.208/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Jinyun Xinjian Yingyuan Network Server Center | APNIC inetnum registration explicitly identifies a hosted-server or server-room network |
| `61.153.72.212/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Jinyun Xinjian Lantanxingyue Network Server Center | APNIC inetnum registration explicitly identifies a hosted-server or server-room network |
| `61.153.73.76/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Zhe Jiang Shu Po Mo | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.202.136/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | IDC2-3550(ddos) | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.223.144/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Cai Zheng Ju | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.241.96/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Nong Ye Ju | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.245.188/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Xin Yong Lian She | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.245.252/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Jiao Yu Ju | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.246.64/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Kai En | APNIC inetnum registration explicitly identifies an IDC network |
| `61.153.247.216/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Si Hai | APNIC inetnum registration explicitly identifies an IDC network |
| `61.164.127.160/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Cangnan branch of network video cloud storage server system | APNIC inetnum registration explicitly identifies an IDC network |
| `61.164.134.232/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Ruian Telecom Idc | APNIC inetnum registration explicitly identifies an IDC network |
| `61.164.146.0/26` | 64 | `apnic_inetnum` | `chinanet / AS4134` | pingyang lantian idc jifang | APNIC inetnum registration explicitly identifies an IDC network |
| `61.164.146.64/26` | 64 | `apnic_inetnum` | `chinanet / AS4134` | PingYang LanTian IDC Machine Room | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.59.128/26` | 64 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.60.8/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.60.32/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.60.64/26` | 64 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.60.192/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.60.224/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.60.228/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.60.232/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.60.240/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.61.20/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Xing Xi Gang | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.61.160/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.63.0/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Xing Xi Gang | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.63.8/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | Xing Xi Gang | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.63.128/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.174.63.192/26` | 64 | `apnic_inetnum` | `chinanet / AS4134` | Wang Shu | APNIC inetnum registration explicitly identifies an IDC network |
| `61.175.99.36/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | ZhouShan Telecom ErShuLouIDC | APNIC inetnum registration explicitly identifies an IDC network |
| `61.175.209.128/26` | 64 | `apnic_inetnum` | `chinanet / AS4134` | WenZhou Telecommunication Co.,ltd Data Centre | APNIC inetnum registration explicitly identifies a data-center network |
| `61.175.209.192/26` | 64 | `apnic_inetnum` | `chinanet / AS4134` | WenZhou Telecommunication Co.,Ltd Data Centre | APNIC inetnum registration explicitly identifies a data-center network |
| `61.175.222.168/29` | 8 | `apnic_inetnum` | `chinanet / AS4134` | SanMen LiFang Communication Server Center | APNIC inetnum registration explicitly identifies a hosted-server or server-room network |
| `61.175.223.176/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | JiaoJiang National Tax Bureau VPN | APNIC inetnum registration explicitly identifies an MPLS or VPN network |
| `61.175.226.0/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | The Data Center Of Telecommunication Co.ltd,Jinhua | APNIC inetnum registration explicitly identifies a data-center network |
| `61.175.226.32/27` | 32 | `apnic_inetnum` | `chinanet / AS4134` | The Data Center Of Telecommunication Co.ltd,Jinhua | APNIC inetnum registration explicitly identifies a data-center network |
| `61.175.242.64/28` | 16 | `apnic_inetnum` | `chinanet / AS4134` | Ji Fa Wei | APNIC inetnum registration explicitly identifies an IDC network |
| `61.175.243.220/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Jinyun Xinjian Yingyuan Network Server Center | APNIC inetnum registration explicitly identifies a hosted-server or server-room network |
| `61.175.243.248/30` | 4 | `apnic_inetnum` | `chinanet / AS4134` | Jinyun Xinjian Lantanxingyue Network Server Center | APNIC inetnum registration explicitly identifies a hosted-server or server-room network |
| `101.251.144.0/20` | 4,096 | `apnic_inetnum` | `unicom / AS4837` | Hangzhou Netbank Technologies co.,LTD | APNIC inetnum registration explicitly identifies an IDC network |
| `103.36.208.0/22` | 1,024 | `apnic_portable_holder` | `unicom / AS4837` | Hangzhou Sulian Information Technology Co.,ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `103.219.28.0/22` | 1,024 | `apnic_portable_holder` | `cmcc / AS56041` | Hangzhou Sulian Information Technology Co.,ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `103.219.32.0/22` | 1,024 | `apnic_portable_holder` | `cmcc / AS56041` | Hangzhou Sulian Information Technology Co.,ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `103.219.36.0/22` | 1,024 | `apnic_portable_holder` | `cmcc / AS56041` | Hangzhou Sulian Information Technology Co.,ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `103.220.240.0/22` | 1,024 | `cloud_provider_cidr` | `unicom / AS4837` | ucloud | Prefix is explicitly listed by IP-Data for ucloud |
| `110.42.0.0/23` | 512 | `apnic_portable_holder` | `cmcc / AS56041` | Ningbo Zhuo Zhi Innovation Network Technology Co., Ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `110.42.3.0/24` | 256 | `apnic_portable_holder` | `cmcc / AS56041` | Ningbo Zhuo Zhi Innovation Network Technology Co., Ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `110.42.4.0/23` | 512 | `apnic_portable_holder` | `cmcc / AS56041` | Ningbo Zhuo Zhi Innovation Network Technology Co., Ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `110.42.6.0/24` | 256 | `apnic_portable_holder` | `cmcc / AS56041` | Ningbo Zhuo Zhi Innovation Network Technology Co., Ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `110.42.8.0/22` | 1,024 | `apnic_portable_holder` | `cmcc / AS56041` | Ningbo Zhuo Zhi Innovation Network Technology Co., Ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `110.42.12.0/24` | 256 | `apnic_portable_holder` | `cmcc / AS56041` | Ningbo Zhuo Zhi Innovation Network Technology Co., Ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |
| `110.42.14.0/24` | 256 | `apnic_portable_holder` | `cmcc / AS56041` | Ningbo Zhuo Zhi Innovation Network Technology Co., Ltd | Most-specific APNIC portable registration is linked to a currently active independent ASN |

其余 102,449 条排除证据未在 Markdown 展开；完整内容保存在 gzip JSON 与 manifest。

## 登记分类

| 分类 | 事实片段 | 地址 | 占全部地址 | 含义 |
|---|---:|---:|---:|---|
| `operator_registration` | 16,947 | 15,325,043 | 100.0000% | 登记文本可归属于三网运营商 |

## 怎样阅读

- ACL 文件采用最大 CIDR 聚合；表中的“保留范围”才是与 APNIC 登记边界对齐后的精确地址范围。
- 浙江试验采用正向准入；独立主体、归属不明、无登记及运营商冲突范围均不进入浙江输出。
- 排名按覆盖地址量排列，用来优先投入人工审查，不代表风险评分。
- 下方索引只负责让主要事实可读；完整证据、全部小片段和全部字段仍以 gzip JSON 为准。

## 当前规则仍识别出的强非公众信号

这些条目已处于最终 ACL 的登记事实中，应优先检查生成边界为何仍保留它们。

| 保留范围 | 所属 ACL CIDR | 运营商 | APNIC 登记主体 | APNIC 登记范围 | 命中原因 |
|---|---|---|---|---|---|
| — | — | — | 当前没有残留强信号 | — | — |

## 独立法定主体登记：地址量前 100 项

共 0 个登记主体标签；下表展示前 0 项。标签优先取 APNIC organisation name，其次取 description、netname 或 organisation handle。

| # | APNIC 登记主体 | 地址 | 占全部地址 | 事实片段 | 保留范围样本 / 所属 ACL CIDR |
|---:|---|---:|---:|---:|---|
| — | 无 | 0 | 0% | 0 | — |

## 其他登记：地址量前 100 项

共 0 个登记主体标签；下表展示前 0 项。标签优先取 APNIC organisation name，其次取 description、netname 或 organisation handle。

| # | APNIC 登记主体 | 地址 | 占全部地址 | 事实片段 | 保留范围样本 / 所属 ACL CIDR |
|---:|---|---:|---:|---:|---|
| — | 无 | 0 | 0% | 0 | — |
