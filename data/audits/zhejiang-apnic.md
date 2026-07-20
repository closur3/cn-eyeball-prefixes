# 浙江 IPv4 APNIC 登记事实审计

本报告以浙江输出为样本，复核全国分层准入规则：当前 BGP Origin 必须属于三网，并由同一家运营商的 APNIC 上层登记覆盖；最具体登记用于审计及强证据排除，不再单独决定准入。完整逐地址事实保存在 [`zhejiang-apnic.json.gz`](./zhejiang-apnic.json.gz)。

## 总览

| 指标 | 数值 |
|---|---:|
| 最大聚合 ACL CIDR | 2,561 |
| IPv4 地址 | 15,952,407 |
| 最具体 APNIC 事实片段 | 116,707 |
| APNIC 登记覆盖 | 15,952,407（100.0000%） |
| 构建规则仍识别出的强非公众信号 | 0 |

## 前缀清洗前后对照

这里的“准入前候选”指已满足三网 Origin 与中国边界、但尚未执行云 CIDR、APNIC、route、MOAS 排除及全国同运营商上层 APNIC 登记准入的浙江地址。分类地址数按证据行累加，多个上游命中同一地址时可能重复；总未准入地址数按地址并集计算。

| 阶段 | 地址 |
|---|---:|
| 准入前候选 | 16,049,454 |
| 未准入（并集） | 97,047 |
| 最终保留 | 15,952,407 |

| 排除类别 | 证据范围 | 地址（可重复） |
|---|---:|---:|
| `apnic_delegated_holder` | 1 | 4 |
| `apnic_independent_legal_entity_holder` | 13 | 5,120 |
| `apnic_inetnum` | 399 | 46,907 |
| `apnic_operator_admission_independent_legal_entity` | 31 | 23,296 |
| `apnic_operator_admission_operator_registration_conflict` | 20 | 216 |
| `apnic_operator_admission_other_registration` | 4 | 5,120 |
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

其余 290 条排除证据未在 Markdown 展开；完整内容保存在 gzip JSON 与 manifest。

## 登记分类

| 分类 | 事实片段 | 地址 | 占全部地址 | 含义 |
|---|---:|---:|---:|---|
| `operator_registration` | 13,212 | 15,302,562 | 95.9264% | 登记文本可归属于三网运营商 |
| `independent_legal_entity` | 42,148 | 255,544 | 1.6019% | 最具体登记为独立法定主体；仅在第二项强证据成立时排除 |
| `other_registration` | 61,347 | 394,301 | 2.4717% | 最具体登记无法归属对应三网；在上层准入成立时保留审计 |

## 怎样阅读

- ACL 文件采用最大 CIDR 聚合；表中的“保留范围”才是与 APNIC 登记边界对齐后的精确地址范围。
- 全国输出统一采用上层运营商登记准入；最具体登记中的运营商冲突会被拒绝，其他独立主体或归属不明记录只有命中既有强证据规则才会删除。
- 排名按覆盖地址量排列，用来优先投入人工审查，不代表风险评分。
- 下方索引只负责让主要事实可读；完整证据、全部小片段和全部字段仍以 gzip JSON 为准。

## 当前规则仍识别出的强非公众信号

这些条目已处于最终 ACL 的登记事实中，应优先检查生成边界为何仍保留它们。

| 保留范围 | 所属 ACL CIDR | 运营商 | APNIC 登记主体 | APNIC 登记范围 | 命中原因 |
|---|---|---|---|---|---|
| — | — | — | 当前没有残留强信号 | — | — |

## 独立法定主体登记：地址量前 100 项

共 37,887 个登记主体标签；下表展示前 100 项。标签优先取 APNIC organisation name，其次取 description、netname 或 organisation handle。

| # | APNIC 登记主体 | 地址 | 占全部地址 | 事实片段 | 保留范围样本 / 所属 ACL CIDR |
|---:|---|---:|---:|---:|---|
| 1 | Ningbo Telecom Co.ltd | 1,800 | 0.0113% | 36 | `61.130.113.164–61.130.113.167` in `61.130.112.0/20`<br>`61.153.144.196–61.153.144.199` in `61.153.144.0/21`<br>`61.175.132.64–61.175.132.79` in `61.175.128.0/18` |
| 2 | Giant Network Technology Co., Ltd. Hangzhou Chong | 1,408 | 0.0088% | 7 | `115.236.22.128–115.236.22.159` in `115.236.22.0/24`<br>`115.236.22.160–115.236.22.191` in `115.236.22.0/24`<br>`115.236.22.192–115.236.22.223` in `115.236.22.0/24` |
| 3 | WENZHOU GAOJIE TECHNOLOGY CO.LTD | 1,224 | 0.0077% | 30 | `60.190.168.128–60.190.168.191` in `60.190.160.0/19`<br>`60.190.172.0–60.190.172.31` in `60.190.160.0/19`<br>`60.190.172.80–60.190.172.95` in `60.190.160.0/19` |
| 4 | Hangzhou international Cci Capital Ltd | 772 | 0.0048% | 92 | `60.190.239.136–60.190.239.143` in `60.190.236.0/22`<br>`60.191.21.112–60.191.21.127` in `60.191.21.64/26`<br>`60.191.23.64–60.191.23.67` in `60.191.22.0/23` |
| 5 | Zhejiang Hairui Network Technology Co., Ltd. | 768 | 0.0048% | 3 | `60.191.140.0–60.191.140.255` in `60.191.140.0/23`<br>`60.191.141.0–60.191.141.255` in `60.191.140.0/23`<br>`115.238.228.0–115.238.228.255` in `115.238.228.0/24` |
| 6 | Hangzhou Network Technology Co., Ltd. Bank of Internet | 656 | 0.0041% | 6 | `115.236.73.128–115.236.73.255` in `115.236.72.0/22`<br>`122.224.103.128–122.224.103.255` in `122.224.100.0/22`<br>`122.224.128.112–122.224.128.127` in `122.224.128.0/22` |
| 7 | HangZhou Woodnn Technology Co., Ltd. | 640 | 0.0040% | 2 | `60.190.249.0–60.190.249.127` in `60.190.248.0/21`<br>`183.131.24.0–183.131.25.255` in `183.131.24.0/22` |
| 8 | WENZHOU TELECOM CO.,LTD | 636 | 0.0040% | 9 | `60.190.112.64–60.190.112.95` in `60.190.112.64/26`<br>`60.190.116.0–60.190.116.255` in `60.190.116.0/23`<br>`60.190.117.0–60.190.117.255` in `60.190.116.0/23` |
| 9 | Hangzhou Matrix Technology Co., Ltd. | 528 | 0.0033% | 29 | `122.225.96.0–122.225.96.15` in `122.225.96.0/23`<br>`122.225.96.16–122.225.96.31` in `122.225.96.0/23`<br>`122.225.96.32–122.225.96.47` in `122.225.96.0/23` |
| 10 | HangZhouSuLianXinXiKeJiYouXianGongSi Co.,ltd | 512 | 0.0032% | 2 | `183.131.68.0–183.131.68.255` in `183.131.68.0/22`<br>`183.131.69.0–183.131.69.255` in `183.131.68.0/22` |
| 11 | JinHuaXunChengXinXiChanYeFaZhanGongSi Co.,ltd | 512 | 0.0032% | 2 | `115.231.87.0–115.231.87.255` in `115.231.87.0/24`<br>`115.231.88.0–115.231.88.255` in `115.231.88.0/24` |
| 12 | Massive hangzhou network technology co., LTD | 512 | 0.0032% | 2 | `115.231.111.0–115.231.111.255` in `115.231.111.0/24`<br>`183.136.238.0–183.136.238.255` in `183.136.238.0/24` |
| 13 | Shaoxing Dingqi Network Technology Co., Ltd. | 512 | 0.0032% | 6 | `122.225.97.128–122.225.97.255` in `122.225.96.0/23`<br>`122.225.103.0–122.225.103.63` in `122.225.102.0/23`<br>`122.225.103.128–122.225.103.255` in `122.225.102.0/23` |
| 14 | Zhejiang province Telecom CO.,LTD Shaoxing Office | 512 | 0.0032% | 2 | `60.190.184.0–60.190.184.255` in `60.190.160.0/19`<br>`60.190.191.0–60.190.191.255` in `60.190.160.0/19` |
| 15 | ZHENGHAI TELECOM. CO.,LTD | 472 | 0.0030% | 11 | `61.153.148.144–61.153.148.159` in `61.153.144.0/21`<br>`61.153.148.160–61.153.148.191` in `61.153.144.0/21`<br>`115.238.130.208–115.238.130.223` in `115.238.130.0/23` |
| 16 | Jilin province high technology co., LTD | 412 | 0.0026% | 7 | `61.164.33.96–61.164.33.99` in `61.164.32.0/20`<br>`115.231.96.0–115.231.96.255` in `115.231.96.0/24`<br>`115.231.98.0–115.231.98.127` in `115.231.98.0/25` |
| 17 | Wenzhou Telecom Co.,ltd | 340 | 0.0021% | 10 | `60.190.80.156–60.190.80.159` in `60.190.80.0/21`<br>`60.190.97.192–60.190.97.207` in `60.190.96.0/21`<br>`60.190.112.24–60.190.112.31` in `60.190.112.16/28` |
| 18 | Jiaxing Telecom Co.,LTD | 332 | 0.0021% | 13 | `60.190.129.196–60.190.129.199` in `60.190.128.0/22`<br>`60.190.131.64–60.190.131.79` in `60.190.128.0/22`<br>`60.190.136.32–60.190.136.39` in `60.190.136.0/23` |
| 19 | HANGZHOU SRT TECHNOLOGY CO., LTD | 320 | 0.0020% | 4 | `61.174.51.0–61.174.51.63` in `61.174.50.0/23`<br>`61.174.51.192–61.174.51.255` in `61.174.50.0/23`<br>`115.238.227.128–115.238.227.255` in `115.238.224.0/22` |
| 20 | Hangzhou Sulian Mdt InfoTech Ltd | 320 | 0.0020% | 2 | `61.153.106.0–61.153.106.63` in `61.153.106.0/24`<br>`115.231.23.0–115.231.23.255` in `115.231.23.0/24` |
| 21 | ZheJiang Province Telecom Co.,Ltd. | 320 | 0.0020% | 5 | `60.190.224.16–60.190.224.31` in `60.190.224.0/22`<br>`60.190.247.0–60.190.247.255` in `60.190.247.0/24`<br>`60.190.252.16–60.190.252.31` in `60.190.248.0/21` |
| 22 | Citic-kington Securities Co.,Ltd. | 296 | 0.0019% | 19 | `115.236.34.160–115.236.34.191` in `115.236.32.0/20`<br>`115.238.32.176–115.238.32.191` in `115.238.32.0/20`<br>`115.238.41.104–115.238.41.111` in `115.238.32.0/20` |
| 23 | Hangzhou Shell Star Tracker Network Technology Co., Ltd. | 288 | 0.0018% | 6 | `115.236.4.192–115.236.4.207` in `115.236.4.0/24`<br>`122.224.114.128–122.224.114.191` in `122.224.112.0/21`<br>`122.224.167.224–122.224.167.255` in `122.224.164.0/22` |
| 24 | Internet Banking Internet Technology Co., Ltd. Hangzhou | 288 | 0.0018% | 3 | `115.236.4.0–115.236.4.127` in `115.236.4.0/24`<br>`115.236.5.64–115.236.5.95` in `115.236.5.0/25`<br>`122.225.219.0–122.225.219.127` in `122.225.219.0/25` |
| 25 | ZheJiang Province Telecom Co.,LTD HangZhou City Filiale | 288 | 0.0018% | 3 | `60.190.226.240–60.190.226.255` in `60.190.224.0/22`<br>`61.164.45.208–61.164.45.223` in `61.164.32.0/20`<br>`61.164.55.0–61.164.55.255` in `61.164.48.0/21` |
| 26 | Zhejiang Telecom Company Limited Information Center | 288 | 0.0018% | 5 | `60.191.62.0–60.191.62.63` in `60.191.60.0/22`<br>`60.191.125.192–60.191.125.255` in `60.191.124.0/22`<br>`61.164.35.128–61.164.35.159` in `61.164.32.0/20` |
| 27 | HUACHEN NET LTD | 272 | 0.0017% | 3 | `122.227.45.16–122.227.45.23` in `122.227.44.0/22`<br>`122.227.45.32–122.227.45.39` in `122.227.44.0/22`<br>`183.131.67.0–183.131.67.255` in `183.131.67.0/24` |
| 28 | Jinhua City Meidiya Network Ltd. | 272 | 0.0017% | 2 | `115.239.195.0–115.239.195.255` in `115.239.192.0/22`<br>`122.226.222.48–122.226.222.63` in `122.226.222.0/24` |
| 29 | Ningbo Zhenhai Telecom Co.LTD | 264 | 0.0017% | 2 | `61.130.114.144–61.130.114.151` in `61.130.112.0/20`<br>`183.136.155.0–183.136.155.255` in `183.136.144.0/20` |
| 30 | Zhejiang Public Communication System Co.,Ltd. | 264 | 0.0017% | 2 | `60.191.115.0–60.191.115.255` in `60.191.112.0/21`<br>`122.225.221.24–122.225.221.31` in `122.225.220.0/22` |
| 31 | Fenghua YiSui Network Technology Co.,Ltd. | 256 | 0.0016% | 1 | `61.174.18.0–61.174.18.255` in `61.174.16.0/20` |
| 32 | Haining telecom Co.,ltd | 256 | 0.0016% | 1 | `60.190.155.0–60.190.155.255` in `60.190.155.0/24` |
| 33 | Haiou Network Technology Co., Ltd. | 256 | 0.0016% | 8 | `61.174.53.0–61.174.53.31` in `61.174.52.0/23`<br>`61.174.53.32–61.174.53.63` in `61.174.52.0/23`<br>`61.174.53.64–61.174.53.95` in `61.174.52.0/23` |
| 34 | Hangzhou Taobao Netwoks Co.,Ltd. | 256 | 0.0016% | 1 | `61.164.54.0–61.164.54.255` in `61.164.48.0/21` |
| 35 | Hangzhou Xiaoshan Information Co.,Ltd( Feilan Alliance of Network) | 256 | 0.0016% | 1 | `218.75.122.0–218.75.122.255` in `218.75.112.0/20` |
| 36 | Hangzhou century etang information technologies CO.,LTD | 256 | 0.0016% | 6 | `60.191.0.160–60.191.0.191` in `60.191.0.0/20`<br>`60.191.3.128–60.191.3.159` in `60.191.0.0/20`<br>`60.191.14.0–60.191.14.63` in `60.191.0.0/20` |
| 37 | Hangzhou treasure network co., LTD | 256 | 0.0016% | 1 | `183.131.13.0–183.131.13.255` in `183.131.13.0/24` |
| 38 | Huzhou Mobile Communications Co.,Ltd.(HUMCC) | 256 | 0.0016% | 1 | `211.140.116.0–211.140.116.255` in `211.140.0.0/17` |
| 39 | Jiaxingshi Xinda Dianzi Keji Co.,Ltd | 256 | 0.0016% | 1 | `183.134.49.0–183.134.49.255` in `183.134.49.0/24` |
| 40 | Jilin province high technology Co., Ltd. | 256 | 0.0016% | 1 | `183.131.18.0–183.131.18.255` in `183.131.18.0/24` |
| 41 | JinHua MeiDiYa Co.,ltd | 256 | 0.0016% | 1 | `61.153.99.0–61.153.99.255` in `61.153.99.0/24` |
| 42 | JinHuaShiMeiDiYaWangLuoKeJiYouXianGongSi Co.,ltd | 256 | 0.0016% | 1 | `183.146.212.0–183.146.212.255` in `183.146.212.0/24` |
| 43 | Linan Tianjian Computer Network Co., Ltd. | 256 | 0.0016% | 1 | `60.191.138.0–60.191.138.255` in `60.191.138.0/23` |
| 44 | QuZhou Mobile Communications Co.,Ltd.(QZMCC) | 256 | 0.0016% | 1 | `211.140.158.0–211.140.158.255` in `211.140.128.0/18` |
| 45 | ShaoXing mobile communication, Ltd. | 256 | 0.0016% | 1 | `211.140.133.0–211.140.133.255` in `211.140.128.0/18` |
| 46 | Transfer highway port logistics co., LTD | 256 | 0.0016% | 1 | `115.233.215.0–115.233.215.255` in `115.233.212.0/22` |
| 47 | Wenzhou Zhongbo Co.,ltd | 256 | 0.0016% | 1 | `61.164.121.0–61.164.121.255` in `61.164.120.0/23` |
| 48 | Zhejiang Public Communication System  Co.,Ltd. | 256 | 0.0016% | 1 | `60.191.114.0–60.191.114.255` in `60.191.112.0/21` |
| 49 | Zhejiang Radio And TV New Media Co.,Ltd | 256 | 0.0016% | 1 | `115.233.200.0–115.233.200.255` in `115.233.200.0/23` |
| 50 | Zhejiang rock Information Technology Co., Ltd. | 256 | 0.0016% | 1 | `122.224.234.0–122.224.234.255` in `122.224.232.0/21` |
| 51 | Zhejiang seized Road Network Technology Co., Ltd. | 256 | 0.0016% | 1 | `122.224.70.0–122.224.70.255` in `122.224.70.0/23` |
| 52 | ZheJiang Province Telecom Co.,Ltd. HangZhou City Filiale | 248 | 0.0016% | 7 | `60.191.54.216–60.191.54.223` in `60.191.48.0/21`<br>`122.224.108.128–122.224.108.255` in `122.224.108.0/22`<br>`122.224.109.0–122.224.109.31` in `122.224.108.0/22` |
| 53 | Xiaoshan Info Co.,Ltd. | 228 | 0.0014% | 4 | `60.191.41.32–60.191.41.63` in `60.191.40.0/23`<br>`60.191.41.64–60.191.41.127` in `60.191.40.0/23`<br>`61.175.193.220–61.175.193.223` in `61.175.192.0/20` |
| 54 | Fenghua Tuosu Co.,ltd | 224 | 0.0014% | 7 | `122.227.135.32–122.227.135.63` in `122.227.128.0/20`<br>`122.227.135.64–122.227.135.95` in `122.227.128.0/20`<br>`122.227.135.96–122.227.135.127` in `122.227.128.0/20` |
| 55 | Jinhua Meidiya Netware Science Co.,ltd | 216 | 0.0014% | 7 | `122.226.62.0–122.226.62.63` in `122.226.60.0/22`<br>`122.226.75.0–122.226.75.63` in `122.226.75.0/24`<br>`122.226.246.176–122.226.246.191` in `122.226.244.0/22` |
| 56 | Hangzhou eastern communications limited company,Hangzhou, Zhejiang Province | 208 | 0.0013% | 3 | `61.175.197.32–61.175.197.47` in `61.175.192.0/20`<br>`115.238.37.0–115.238.37.127` in `115.238.32.0/20`<br>`115.238.37.128–115.238.37.191` in `115.238.32.0/20` |
| 57 | Hempel (China) Co., Ltd. | 200 | 0.0013% | 12 | `115.236.5.224–115.236.5.239` in `115.236.5.192/26`<br>`115.236.17.144–115.236.17.159` in `115.236.17.0/24`<br>`115.236.17.160–115.236.17.175` in `115.236.17.0/24` |
| 58 | Jinhua Telecom Co.,ltd | 196 | 0.0012% | 44 | `60.191.198.64–60.191.198.67` in `60.191.198.0/23`<br>`60.191.210.16–60.191.210.19` in `60.191.208.0/21`<br>`60.191.210.228–60.191.210.231` in `60.191.208.0/21` |
| 59 | Ningbo Bitian science and Technology Co., Ltd. | 192 | 0.0012% | 2 | `183.136.146.0–183.136.146.63` in `183.136.144.0/20`<br>`183.136.150.128–183.136.150.255` in `183.136.144.0/20` |
| 60 | Ningbo Gaoxin Qianxin Network Information Co.,Ltd. | 192 | 0.0012% | 4 | `61.130.111.128–61.130.111.191` in `61.130.108.0/22`<br>`61.130.111.192–61.130.111.255` in `61.130.108.0/22`<br>`115.238.139.128–115.238.139.159` in `115.238.139.0/24` |
| 61 | Wenzhou Gaojie Co.,ltd | 192 | 0.0012% | 3 | `60.190.81.128–60.190.81.159` in `60.190.80.0/21`<br>`60.190.81.160–60.190.81.191` in `60.190.80.0/21`<br>`60.190.113.0–60.190.113.127` in `60.190.113.0/24` |
| 62 | Ningbo hi-tech park bi tian technology Co., LTD | 160 | 0.0010% | 2 | `183.136.156.128–183.136.156.255` in `183.136.144.0/20`<br>`183.136.192.0–183.136.192.31` in `183.136.192.0/21` |
| 63 | Ningbo Public Information Industry Co., Ltd. | 156 | 0.0010% | 5 | `61.130.106.168–61.130.106.171` in `61.130.106.0/23`<br>`61.130.106.172–61.130.106.175` in `61.130.106.0/23`<br>`61.130.106.248–61.130.106.251` in `61.130.106.0/23` |
| 64 | Hangzhou Yuesheng Electronic Dommunications Co.,Ltd | 152 | 0.0010% | 3 | `122.227.145.200–122.227.145.207` in `122.227.144.0/23`<br>`122.227.173.80–122.227.173.95` in `122.227.168.0/21`<br>`122.227.184.0–122.227.184.127` in `122.227.184.0/24` |
| 65 | Ningbo Shengshi Network Technology Co., Ltd | 144 | 0.0009% | 2 | `122.227.182.128–122.227.182.143` in `122.227.182.128/25`<br>`183.136.156.0–183.136.156.127` in `183.136.144.0/20` |
| 66 | Ningbo obang Information Technology Co., Ltd.TWO | 144 | 0.0009% | 5 | `122.226.168.0–122.226.168.31` in `122.226.168.0/21`<br>`122.226.168.32–122.226.168.63` in `122.226.168.0/21`<br>`122.226.171.128–122.226.171.159` in `122.226.168.0/21` |
| 67 | Shenzhen Century Triumph Technology Co., Ltd. | 144 | 0.0009% | 3 | `122.227.178.144–122.227.178.151` in `122.227.176.0/22`<br>`122.227.250.40–122.227.250.47` in `122.227.250.0/24`<br>`183.136.158.128–183.136.158.255` in `183.136.144.0/20` |
| 68 | ZheJiang taobao Co.,ltd | 144 | 0.0009% | 5 | `60.190.115.128–60.190.115.159` in `60.190.115.0/24`<br>`60.190.115.160–60.190.115.191` in `60.190.115.0/24`<br>`60.190.115.192–60.190.115.223` in `60.190.115.0/24` |
| 69 | Beijing Lanxun  Communications Technology Co.,Ltd | 140 | 0.0009% | 4 | `60.190.236.128–60.190.236.255` in `60.190.236.0/22`<br>`122.224.97.0–122.224.97.3` in `122.224.97.0/24`<br>`122.224.97.4–122.224.97.7` in `122.224.97.0/24` |
| 70 | Zhejiang Provice Telecom Limited Company Hangzhou Branch | 140 | 0.0009% | 13 | `60.190.240.240–60.190.240.243` in `60.190.240.0/24`<br>`60.191.74.252–60.191.74.255` in `60.191.72.0/21`<br>`60.191.88.96–60.191.88.127` in `60.191.88.0/22` |
| 71 | Zhejiang Telecom Co.,Ltd Hangzhou Branch | 140 | 0.0009% | 14 | `60.190.240.248–60.190.240.255` in `60.190.240.0/24`<br>`60.191.65.96–60.191.65.103` in `60.191.64.0/22`<br>`60.191.69.120–60.191.69.127` in `60.191.69.0/24` |
| 72 | Zhejiang Telecom Co.Ltd Hangzhou Limited | 136 | 0.0009% | 18 | `60.190.231.88–60.190.231.91` in `60.190.231.0/25`<br>`60.191.61.124–60.191.61.127` in `60.191.60.0/22`<br>`60.191.62.224–60.191.62.227` in `60.191.60.0/22` |
| 73 | Zhejiang Telecom Company Limited | 136 | 0.0009% | 4 | `60.190.236.120–60.190.236.127` in `60.190.236.0/22`<br>`60.191.43.192–60.191.43.255` in `60.191.43.128/25`<br>`60.191.78.0–60.191.78.31` in `60.191.72.0/21` |
| 74 | Ningbo Dianxin Co.,Ltd | 132 | 0.0008% | 32 | `60.190.19.12–60.190.19.15` in `60.190.16.0/21`<br>`60.190.23.148–60.190.23.151` in `60.190.16.0/21`<br>`60.190.30.240–60.190.30.243` in `60.190.30.0/24` |
| 75 | ZHEJIANG MOBILE COMMUNICATION CO.LTD | 132 | 0.0008% | 2 | `60.191.74.96–60.191.74.99` in `60.191.72.0/21`<br>`60.191.77.0–60.191.77.127` in `60.191.72.0/21` |
| 76 | Beijing Sohu Internet Message Server Co.,ltd | 128 | 0.0008% | 1 | `122.226.47.0–122.226.47.127` in `122.226.47.0/24` |
| 77 | Blue Communications Technology Co., Ltd. | 128 | 0.0008% | 2 | `61.153.151.128–61.153.151.191` in `61.153.144.0/21`<br>`122.227.245.64–122.227.245.127` in `122.227.244.0/22` |
| 78 | Ding Kai Shaoxing Network Technology Co., Ltd. | 128 | 0.0008% | 4 | `60.190.238.0–60.190.238.31` in `60.190.236.0/22`<br>`60.191.42.96–60.191.42.127` in `60.191.42.0/24`<br>`60.191.43.0–60.191.43.31` in `60.191.43.0/26` |
| 79 | FUYANG TELECOM CO.,LTD | 128 | 0.0008% | 1 | `202.107.193.0–202.107.193.127` in `202.107.193.0/24` |
| 80 | HANGZHOU DIFO TELECOMMUNICATION CO.LTD,Hangzhou,Zhejiang Province | 128 | 0.0008% | 1 | `202.101.184.128–202.101.184.255` in `202.101.176.0/20` |
| 81 | Hangzhou CLUDE Electric Information Technology Company Limited | 128 | 0.0008% | 4 | `60.190.144.0–60.190.144.31` in `60.190.144.0/24`<br>`60.190.144.32–60.190.144.63` in `60.190.144.0/24`<br>`60.190.144.64–60.190.144.95` in `60.190.144.0/24` |
| 82 | Hangzhou Woaiwojia real estate Co., Ltd. | 128 | 0.0008% | 32 | `60.190.226.136–60.190.226.139` in `60.190.224.0/22`<br>`60.190.239.188–60.190.239.191` in `60.190.236.0/22`<br>`60.191.54.128–60.191.54.131` in `60.191.48.0/21` |
| 83 | Hangzhou Yunshao Technology Co., Ltd. | 128 | 0.0008% | 1 | `183.131.16.0–183.131.16.127` in `183.131.16.0/23` |
| 84 | Jiaxing telecom Co.,LTD (IPTV Project) | 128 | 0.0008% | 1 | `122.225.0.0–122.225.0.127` in `122.225.0.0/21` |
| 85 | Shanghai Chen Yi-contact network Technology Co., Ltd. | 128 | 0.0008% | 1 | `122.224.115.0–122.224.115.127` in `122.224.112.0/21` |
| 86 | Shanghai Tuchu Computer Co.,Ltd. | 128 | 0.0008% | 1 | `115.238.166.128–115.238.166.255` in `115.238.166.0/23` |
| 87 | UTStarcom Telecom Co.,Ltd. | 128 | 0.0008% | 1 | `122.224.216.0–122.224.216.127` in `122.224.216.0/22` |
| 88 | ZHENGHAI OIL REFINING CHEMICAL CO.,LTD | 128 | 0.0008% | 1 | `218.0.4.0–218.0.4.127` in `218.0.0.0/19` |
| 89 | ZheJiang Province Telecom LTD LinAn City Branch | 128 | 0.0008% | 1 | `60.191.72.0–60.191.72.127` in `60.191.72.0/21` |
| 90 | ZheJiang Province Telecom LTD Xiaoshan Branch | 128 | 0.0008% | 1 | `60.191.121.128–60.191.121.255` in `60.191.120.0/23` |
| 91 | Zhejiang Electronic Port Co.,Ltd | 128 | 0.0008% | 7 | `60.191.76.80–60.191.76.95` in `60.191.72.0/21`<br>`60.191.126.168–60.191.126.175` in `60.191.124.0/22`<br>`115.236.37.128–115.236.37.159` in `115.236.32.0/20` |
| 92 | Zhejiang Securities Co., Ltd. | 128 | 0.0008% | 6 | `115.236.14.96–115.236.14.127` in `115.236.8.0/21`<br>`115.236.39.80–115.236.39.87` in `115.236.32.0/20`<br>`115.236.66.16–115.236.66.31` in `115.236.64.0/22` |
| 93 | Zhejiang Telecommunication Shaoxing Ltd | 128 | 0.0008% | 1 | `60.190.198.0–60.190.198.127` in `60.190.192.0/21` |
| 94 | Zhejiang commercial bank co. LTD | 128 | 0.0008% | 2 | `183.134.213.0–183.134.213.63` in `183.134.208.0/21`<br>`183.134.213.64–183.134.213.127` in `183.134.208.0/21` |
| 95 | cernrt network ltd. | 128 | 0.0008% | 1 | `122.224.127.128–122.224.127.255` in `122.224.124.0/22` |
| 96 | Zhejiang Telecom Company Limited Safety Dep. | 116 | 0.0007% | 15 | `60.190.231.0–60.190.231.7` in `60.190.231.0/25`<br>`60.190.231.8–60.190.231.15` in `60.190.231.0/25`<br>`60.190.231.16–60.190.231.23` in `60.190.231.0/25` |
| 97 | HANGZHOU UTSTAR Telecommunication Co.Ltd,Hangzhou,Zhejiang Province | 108 | 0.0007% | 4 | `60.191.4.64–60.191.4.127` in `60.191.0.0/20`<br>`61.175.193.128–61.175.193.159` in `61.175.192.0/20`<br>`61.175.199.224–61.175.199.227` in `61.175.192.0/20` |
| 98 | NINGBO TELECOM CO.,LTD | 104 | 0.0007% | 6 | `60.190.62.80–60.190.62.95` in `60.190.62.0/24`<br>`61.175.133.64–61.175.133.79` in `61.175.128.0/18`<br>`61.175.204.128–61.175.204.159` in `61.175.192.0/20` |
| 99 | Hangzou Tianjifeng Communications Ltd | 100 | 0.0006% | 3 | `60.191.72.128–60.191.72.159` in `60.191.72.0/21`<br>`122.224.97.104–122.224.97.107` in `122.224.97.0/24`<br>`122.224.176.128–122.224.176.191` in `122.224.176.0/24` |
| 100 | A few meters in Hangzhou Network Technology Co., Ltd. | 96 | 0.0006% | 3 | `115.236.72.0–115.236.72.31` in `115.236.72.0/22`<br>`115.236.75.224–115.236.75.255` in `115.236.72.0/22`<br>`122.224.94.32–122.224.94.63` in `122.224.94.0/24` |

其余 37,787 个较小登记主体标签未在 Markdown 展开，可在完整 gzip JSON 中查询。

## 其他登记：地址量前 100 项

共 53,004 个登记主体标签；下表展示前 100 项。标签优先取 APNIC organisation name，其次取 description、netname 或 organisation handle。

| # | APNIC 登记主体 | 地址 | 占全部地址 | 事实片段 | 保留范围样本 / 所属 ACL CIDR |
|---:|---|---:|---:|---:|---|
| 1 | China Railcom Zhejiang Branch | 1,792 | 0.0112% | 7 | `61.232.85.0–61.232.85.255` in `61.232.80.0/21`<br>`61.234.186.0–61.234.186.255` in `61.234.176.0/20`<br>`61.234.187.0–61.234.187.255` in `61.234.176.0/20` |
| 2 | YOUHUADIZHIBEIYONG,JINHUA,ZHEJIANG | 1,536 | 0.0096% | 2 | `60.12.154.0–60.12.155.255` in `60.12.152.0/21`<br>`60.12.156.0–60.12.159.255` in `60.12.152.0/21` |
| 3 | Hangzhou Telecom | 1,532 | 0.0096% | 90 | `60.191.93.216–60.191.93.223` in `60.191.93.192/26`<br>`61.130.8.64–61.130.8.95` in `61.130.8.0/24`<br>`61.130.8.128–61.130.8.191` in `61.130.8.0/24` |
| 4 | MEIDIYA,JINHUA,ZHEJIANG | 1,156 | 0.0072% | 27 | `60.12.160.96–60.12.160.127` in `60.12.160.0/24`<br>`60.12.160.144–60.12.160.159` in `60.12.160.0/24`<br>`60.12.160.192–60.12.160.207` in `60.12.160.0/24` |
| 5 | TELECOM | 1,064 | 0.0067% | 225 | `60.190.76.212–60.190.76.215` in `60.190.76.0/23`<br>`60.190.76.216–60.190.76.219` in `60.190.76.0/23`<br>`60.190.89.48–60.190.89.51` in `60.190.88.0/22` |
| 6 | Zhongguo Dianxin Jiaxing Fengongsi | 1,064 | 0.0067% | 6 | `60.190.129.32–60.190.129.63` in `60.190.128.0/22`<br>`183.131.96.0–183.131.96.255` in `183.131.96.0/21`<br>`183.131.97.0–183.131.97.255` in `183.131.96.0/21` |
| 7 | LANYUEKEJI,HANGZHOU,ZHEJIANG | 1,060 | 0.0066% | 12 | `60.12.225.128–60.12.225.191` in `60.12.224.0/23`<br>`60.12.229.0–60.12.229.255` in `60.12.228.0/23`<br>`60.12.230.0–60.12.230.31` in `60.12.230.0/25` |
| 8 | XINGYUNHENGTONG,HANGZHOU,ZHEJIANG | 1,024 | 0.0064% | 4 | `124.160.96.0–124.160.96.255` in `124.160.96.0/21`<br>`124.160.97.0–124.160.97.255` in `124.160.96.0/21`<br>`124.160.98.0–124.160.98.255` in `124.160.96.0/21` |
| 9 | Zhenhai Lianhua Residential Quarters | 1,024 | 0.0064% | 1 | `60.190.8.0–60.190.11.255` in `60.190.8.0/21` |
| 10 | LIANTONG,HANGZHOU,ZHEJIANG | 1,000 | 0.0063% | 40 | `221.12.3.80–221.12.3.95` in `221.12.0.0/21`<br>`221.12.4.204–221.12.4.207` in `221.12.0.0/21`<br>`221.12.5.252–221.12.5.255` in `221.12.0.0/21` |
| 11 | hangzhoujuzheng,huzhou,zhejiang | 976 | 0.0061% | 10 | `60.12.104.0–60.12.104.255` in `60.12.96.0/19`<br>`60.12.105.64–60.12.105.127` in `60.12.96.0/19`<br>`60.12.106.192–60.12.106.255` in `60.12.96.0/19` |
| 12 | ZHEJIANG SCIENCE AND TECHNOLOGY INFORMATION INSTUTITE | 900 | 0.0056% | 4 | `61.153.5.0–61.153.5.255` in `61.153.4.0/22`<br>`202.107.198.248–202.107.198.251` in `202.107.196.0/22`<br>`202.107.204.0–202.107.205.255` in `202.107.200.0/21` |
| 13 | SHIJIYITENG,HANGZHOU,ZHEJIANG | 772 | 0.0048% | 4 | `60.12.225.72–60.12.225.75` in `60.12.224.0/23`<br>`60.12.226.0–60.12.226.255` in `60.12.226.0/24`<br>`60.12.228.0–60.12.228.255` in `60.12.228.0/23` |
| 14 | Jiashan Radio and Television Bureau | 768 | 0.0048% | 2 | `122.225.44.0–122.225.45.255` in `122.225.40.0/21`<br>`122.225.46.0–122.225.46.255` in `122.225.40.0/21` |
| 15 | ZHEJIANGHENGHUAWANGLUOKEJIYOUXIANGONGSI,HANGZHOU,ZHEJIANG | 768 | 0.0048% | 3 | `124.160.114.0–124.160.114.255` in `124.160.114.0/23`<br>`124.160.115.0–124.160.115.255` in `124.160.114.0/23`<br>`124.160.116.0–124.160.116.255` in `124.160.116.0/22` |
| 16 | SANHAOKEJI,HANGZHOU,ZHEJIANG | 766 | 0.0048% | 3 | `124.160.37.0–124.160.37.254` in `124.160.32.0/19`<br>`124.160.38.0–124.160.38.254` in `124.160.32.0/19`<br>`221.12.16.0–221.12.16.255` in `221.12.16.0/20` |
| 17 | Dongyang city education technology and Information Center | 640 | 0.0040% | 2 | `61.175.238.128–61.175.238.255` in `61.175.232.0/21`<br>`183.131.70.0–183.131.71.255` in `183.131.68.0/22` |
| 18 | Yuyao Telecom | 640 | 0.0040% | 71 | `60.190.27.64–60.190.27.67` in `60.190.24.0/22`<br>`60.190.27.188–60.190.27.191` in `60.190.24.0/22`<br>`60.190.34.144–60.190.34.147` in `60.190.32.0/22` |
| 19 | Chinese life insurance company video conference and monitoring project | 576 | 0.0036% | 2 | `183.134.80.0–183.134.81.255` in `183.134.80.0/21`<br>`183.134.83.0–183.134.83.63` in `183.134.80.0/21` |
| 20 | Ningbo Zhenhai Telecom | 564 | 0.0035% | 24 | `60.190.46.0–60.190.46.15` in `60.190.40.0/21`<br>`60.190.46.128–60.190.46.255` in `60.190.40.0/21`<br>`115.238.130.64–115.238.130.67` in `115.238.130.0/23` |
| 21 | ZheJiang TongXiang Education Burean | 536 | 0.0034% | 3 | `122.225.20.224–122.225.20.231` in `122.225.20.0/23`<br>`122.225.49.64–122.225.49.79` in `122.225.48.0/21`<br>`122.225.50.0–122.225.51.255` in `122.225.48.0/21` |
| 22 | Education Berau Of Ruian | 528 | 0.0033% | 2 | `61.164.102.48–61.164.102.63` in `61.164.96.0/21`<br>`122.228.172.0–122.228.173.255` in `122.228.160.0/19` |
| 23 | SHENGKEJIJU,HANGZHOU,ZHEJIANG | 514 | 0.0032% | 3 | `60.12.14.248–60.12.14.251` in `60.12.12.0/22`<br>`124.160.39.0–124.160.39.254` in `124.160.32.0/19`<br>`124.160.40.0–124.160.40.254` in `124.160.32.0/19` |
| 24 | NINGBO-SHENGDONG-LTD | 512 | 0.0032% | 2 | `61.130.108.0–61.130.108.255` in `61.130.108.0/22`<br>`61.130.109.0–61.130.109.255` in `61.130.108.0/22` |
| 25 | Wang Sukeji | 512 | 0.0032% | 16 | `183.131.120.0–183.131.120.31` in `183.131.120.0/22`<br>`183.131.120.32–183.131.120.63` in `183.131.120.0/22`<br>`183.131.120.64–183.131.120.95` in `183.131.120.0/22` |
| 26 | ZHEJIANG NORMAL UNIVERSITY | 512 | 0.0032% | 2 | `61.153.34.0–61.153.34.255` in `61.153.32.0/22`<br>`61.175.228.0–61.175.228.255` in `61.175.228.0/22` |
| 27 | HZWAWJ | 476 | 0.0030% | 116 | `60.191.52.72–60.191.52.75` in `60.191.48.0/21`<br>`60.191.52.76–60.191.52.79` in `60.191.48.0/21`<br>`60.191.52.88–60.191.52.91` in `60.191.48.0/21` |
| 28 | Changxing Education Bureau | 448 | 0.0028% | 3 | `115.238.224.0–115.238.224.255` in `115.238.224.0/22`<br>`218.75.55.64–218.75.55.127` in `218.75.52.0/22`<br>`218.75.55.128–218.75.55.255` in `218.75.52.0/22` |
| 29 | KANGANKEJI,HANGZHOU,ZHEJIANG | 388 | 0.0024% | 3 | `124.160.46.0–124.160.46.255` in `124.160.32.0/19`<br>`124.160.47.0–124.160.47.127` in `124.160.32.0/19`<br>`124.160.59.4–124.160.59.7` in `124.160.32.0/19` |
| 30 | HangZhou | 384 | 0.0024% | 2 | `122.226.116.128–122.226.116.255` in `122.226.116.0/24`<br>`183.131.212.0–183.131.212.255` in `183.131.212.0/23` |
| 31 | Jiande City Board of Education | 384 | 0.0024% | 2 | `122.225.194.0–122.225.194.255` in `122.225.192.0/21`<br>`122.225.195.0–122.225.195.127` in `122.225.192.0/21` |
| 32 | WenZhou Hotline | 384 | 0.0024% | 2 | `202.107.217.0–202.107.217.255` in `202.107.208.0/20`<br>`220.189.240.0–220.189.240.127` in `220.189.240.0/21` |
| 33 | Shaoxing Telecom Bureau | 360 | 0.0023% | 11 | `60.190.186.180–60.190.186.183` in `60.190.160.0/19`<br>`60.190.190.0–60.190.190.255` in `60.190.160.0/19`<br>`60.190.194.248–60.190.194.255` in `60.190.192.0/21` |
| 34 | Zhejiang Telecom | 360 | 0.0023% | 12 | `60.191.21.32–60.191.21.39` in `60.191.21.32/28`<br>`122.224.182.96–122.224.182.127` in `122.224.182.0/24`<br>`122.224.197.168–122.224.197.175` in `122.224.196.0/22` |
| 35 | China Railcom Zhejiang Jinhua Subbranch | 352 | 0.0022% | 3 | `61.232.80.0–61.232.80.255` in `61.232.80.0/21`<br>`61.232.83.32–61.232.83.63` in `61.232.80.0/21`<br>`61.232.83.64–61.232.83.127` in `61.232.80.0/21` |
| 36 | Beijing Baiwu technolegy | 320 | 0.0020% | 10 | `183.131.122.0–183.131.122.31` in `183.131.120.0/22`<br>`183.131.122.32–183.131.122.63` in `183.131.120.0/22`<br>`183.131.122.64–183.131.122.95` in `183.131.120.0/22` |
| 37 | ZHEJIANG PUBLIC INFORMATION CENTER | 320 | 0.0020% | 2 | `202.101.165.128–202.101.165.191` in `202.101.165.128/25`<br>`218.75.107.0–218.75.107.255` in `218.75.104.0/22` |
| 38 | Telecom | 312 | 0.0020% | 40 | `60.191.2.192–60.191.2.195` in `60.191.0.0/20`<br>`60.191.2.196–60.191.2.199` in `60.191.0.0/20`<br>`61.130.77.96–61.130.77.99` in `61.130.76.0/22` |
| 39 | HANGZHONGZHIYEJSHUXUEYUAN | 304 | 0.0019% | 35 | `115.236.31.216–115.236.31.223` in `115.236.24.0/21`<br>`115.236.33.152–115.236.33.155` in `115.236.32.0/20`<br>`115.236.44.104–115.236.44.111` in `115.236.32.0/20` |
| 40 | The thick of River Street | 292 | 0.0018% | 73 | `60.191.199.136–60.191.199.139` in `60.191.198.0/23`<br>`60.191.199.200–60.191.199.203` in `60.191.198.0/23`<br>`60.191.199.208–60.191.199.211` in `60.191.198.0/23` |
| 41 | WANGYUEKEJI,HANGZHOU,ZHEJIANG | 288 | 0.0018% | 2 | `60.12.234.0–60.12.234.255` in `60.12.232.0/21`<br>`124.160.61.96–124.160.61.127` in `124.160.32.0/19` |
| 42 | ZHEJIANGDAXUEWANGLUOYUXINXIZHONGXIN,HANGZHOU,ZHEJIANG | 288 | 0.0018% | 2 | `60.12.143.0–60.12.143.255` in `60.12.142.0/23`<br>`124.160.45.32–124.160.45.63` in `124.160.32.0/19` |
| 43 | telecom | 288 | 0.0018% | 70 | `60.190.109.136–60.190.109.139` in `60.190.108.0/22`<br>`60.190.109.140–60.190.109.143` in `60.190.108.0/22`<br>`60.190.109.148–60.190.109.151` in `60.190.108.0/22` |
| 44 | WENZHOU RENSHOU BAOXIAN GUFENYOUXIANGONGSI | 284 | 0.0018% | 37 | `60.190.67.168–60.190.67.175` in `60.190.66.0/23`<br>`60.190.69.152–60.190.69.159` in `60.190.68.0/22`<br>`60.190.69.160–60.190.69.167` in `60.190.68.0/22` |
| 45 | NGN EXPERIMENT ENGINEERING OF HANGZHOU TELECOM | 272 | 0.0017% | 2 | `60.191.103.48–60.191.103.63` in `60.191.102.0/23`<br>`61.130.1.0–61.130.1.255` in `61.130.0.0/23` |
| 46 | Yiwu Telecom | 272 | 0.0017% | 30 | `60.191.208.28–60.191.208.31` in `60.191.208.0/21`<br>`60.191.215.48–60.191.215.51` in `60.191.208.0/21`<br>`60.191.227.56–60.191.227.59` in `60.191.224.0/22` |
| 47 | ALIMAMA,HANGZHOU,ZHEJIANG | 271 | 0.0017% | 5 | `124.160.16.0–124.160.16.254` in `124.160.16.0/22`<br>`124.160.33.20–124.160.33.23` in `124.160.32.0/19`<br>`124.160.33.24–124.160.33.27` in `124.160.32.0/19` |
| 48 | HangZhou City ShangCheng District Gov. Office | 264 | 0.0017% | 2 | `60.191.19.128–60.191.19.135` in `60.191.16.0/22`<br>`122.224.136.0–122.224.136.255` in `122.224.136.0/22` |
| 49 | zhoushan Educational Bureau | 264 | 0.0017% | 3 | `61.153.36.64–61.153.36.71` in `61.153.36.64/27`<br>`220.189.203.0–220.189.203.127` in `220.189.192.0/20`<br>`220.189.203.128–220.189.203.255` in `220.189.192.0/20` |
| 50 | BEIJINGBIKONGHENGTONGWANGLUOKEJIYOUXIANGONGSI,HANGZHOU,ZHEJIANG | 260 | 0.0016% | 2 | `124.160.36.100–124.160.36.103` in `124.160.32.0/19`<br>`124.160.119.0–124.160.119.255` in `124.160.116.0/22` |
| 51 | HESHUTONGXIN,HANGZHOU,ZHEJIANG | 260 | 0.0016% | 2 | `124.160.84.168–124.160.84.171` in `124.160.80.0/21`<br>`124.160.87.0–124.160.87.255` in `124.160.80.0/21` |
| 52 | RenJiaXinXiKeJi,ZheJiang,Wenzhou | 260 | 0.0016% | 2 | `60.12.35.124–60.12.35.127` in `60.12.32.0/22`<br>`60.12.50.0–60.12.50.255` in `60.12.48.0/20` |
| 53 | Shanghai Jiaotong University | 260 | 0.0016% | 3 | `61.164.36.0–61.164.36.127` in `61.164.32.0/20`<br>`61.164.36.128–61.164.36.255` in `61.164.32.0/20`<br>`61.164.56.16–61.164.56.19` in `61.164.56.0/22` |
| 54 | Zhejiang Province Electric Power Company,Hangzhou,Zhejiang Province | 260 | 0.0016% | 2 | `61.175.193.212–61.175.193.215` in `61.175.192.0/20`<br>`202.107.201.0–202.107.201.255` in `202.107.200.0/21` |
| 55 | BEIJINGBOZHIRUIHAIWANGLUOKEJIYOUXIANDADAI,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.85.0–124.160.85.255` in `124.160.80.0/21` |
| 56 | BEIJINGGUOCHUANGFUSHENGTONGXINGONGSI,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.118.0–124.160.118.255` in `124.160.116.0/22` |
| 57 | BEIJINGXINGYUNHENGTONG,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.117.0–124.160.117.255` in `124.160.116.0/22` |
| 58 | Enviornmental Protection Monitoring System | 256 | 0.0016% | 8 | `61.164.130.64–61.164.130.95` in `61.164.128.0/22`<br>`61.164.130.96–61.164.130.127` in `61.164.128.0/22`<br>`61.164.130.128–61.164.130.159` in `61.164.128.0/22` |
| 59 | GONGSHANGDAXUEJIAOXUEQU,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.64.0–124.160.64.255` in `124.160.64.0/22` |
| 60 | HANGZHOU GONGSHU Information Center | 256 | 0.0016% | 1 | `122.224.138.0–122.224.138.255` in `122.224.136.0/22` |
| 61 | HUANWANGDIANTONG,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.132.0–124.160.132.255` in `124.160.132.0/22` |
| 62 | HZCNC,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 30 | `60.12.0.0–60.12.0.31` in `60.12.0.0/22`<br>`60.12.0.144–60.12.0.159` in `60.12.0.0/22`<br>`60.12.0.160–60.12.0.175` in `60.12.0.0/22` |
| 63 | HangZhouYiWangHuLianKeJi,ZheJiang,Wenzhou | 256 | 0.0016% | 1 | `60.12.51.0–60.12.51.255` in `60.12.48.0/20` |
| 64 | Hangzhou Lian Tian ship computer network Co. | 256 | 0.0016% | 12 | `183.131.3.128–183.131.3.255` in `183.131.2.0/23`<br>`202.96.98.128–202.96.98.159` in `202.96.96.0/20`<br>`202.96.98.160–202.96.98.175` in `202.96.96.0/20` |
| 65 | Hangzhou Xiaoshan Technology Committee | 256 | 0.0016% | 1 | `61.175.195.0–61.175.195.255` in `61.175.192.0/20` |
| 66 | JINZHILUXINXIKEJIYOUXIANGONGSI,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.122.0–124.160.122.255` in `124.160.122.0/24` |
| 67 | JinHuaShiMeiDiYaWangLuoYouXianGongSi | 256 | 0.0016% | 1 | `60.191.203.0–60.191.203.255` in `60.191.203.0/24` |
| 68 | Jinhua Daily Newspaper Office,Jinhua, Zhejiang Province | 256 | 0.0016% | 1 | `61.153.35.0–61.153.35.255` in `61.153.32.0/22` |
| 69 | Jinhua Jiaoyuju | 256 | 0.0016% | 1 | `61.153.103.0–61.153.103.255` in `61.153.103.0/24` |
| 70 | KEJICHUANGYEYUAN,JINHUA,ZHEJIANG | 256 | 0.0016% | 1 | `60.12.147.0–60.12.147.255` in `60.12.144.0/22` |
| 71 | NINGBO EDUCATION SCIENCE CENTER | 256 | 0.0016% | 1 | `202.107.209.0–202.107.209.255` in `202.107.208.0/20` |
| 72 | NINGBO XINNIAO INTERNET BAR | 256 | 0.0016% | 1 | `61.175.201.0–61.175.201.255` in `61.175.192.0/20` |
| 73 | Ningbo renming zhengfu xinxi office | 256 | 0.0016% | 1 | `60.190.2.0–60.190.2.255` in `60.190.2.0/23` |
| 74 | OHJY | 256 | 0.0016% | 1 | `122.228.129.0–122.228.129.255` in `122.228.128.0/21` |
| 75 | RiBaoJiTuan,ZheJiang,Wenzhou | 256 | 0.0016% | 1 | `221.12.75.0–221.12.75.255` in `221.12.72.0/22` |
| 76 | SHANGHAISHUYUANDADAIKUAN,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.133.0–124.160.133.255` in `124.160.132.0/22` |
| 77 | SHANGHAISHUYUANPUTONG,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.134.0–124.160.134.255` in `124.160.132.0/22` |
| 78 | SHANGHAIYONGTIANXINXIJISHUYOUXIANGONGSI,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.101.0–124.160.101.255` in `124.160.96.0/21` |
| 79 | SHENGGONGSIFUWUQI,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `221.12.14.0–221.12.14.255` in `221.12.14.0/24` |
| 80 | Shaoxing Telecom Bureau Data  Center | 256 | 0.0016% | 1 | `61.174.214.0–61.174.214.255` in `61.174.192.0/19` |
| 81 | Taxation Administration Of Zhoushan | 256 | 0.0016% | 1 | `61.153.39.0–61.153.39.255` in `61.153.38.0/23` |
| 82 | Tonglu County Transport Bureau | 256 | 0.0016% | 1 | `183.129.225.0–183.129.225.255` in `183.129.224.0/22` |
| 83 | WenZhou LuCheng Education Bureau | 256 | 0.0016% | 3 | `61.153.24.128–61.153.24.191` in `61.153.24.0/22`<br>`122.228.156.0–122.228.156.127` in `122.228.144.0/20`<br>`122.228.156.128–122.228.156.191` in `122.228.144.0/20` |
| 84 | XINGZHENGZHIFAJU,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `60.12.252.0–60.12.252.255` in `60.12.240.0/20` |
| 85 | Xiaoshan Education Network | 256 | 0.0016% | 1 | `61.153.9.0–61.153.9.255` in `61.153.8.0/22` |
| 86 | YH-JiaoYu-WeiYuanHui | 256 | 0.0016% | 1 | `61.153.194.0–61.153.194.255` in `61.153.192.0/21` |
| 87 | ZHEJIANGCHENQIAOTONGXINJISHUYOUXIANGONGSI,HANGZHOU,ZHEJIANG | 256 | 0.0016% | 1 | `124.160.100.0–124.160.100.255` in `124.160.96.0/21` |
| 88 | Zhejiang Telecom Hulianxinkong | 256 | 0.0016% | 1 | `218.75.79.0–218.75.79.255` in `218.75.64.0/20` |
| 89 | da qingnaqi | 256 | 0.0016% | 1 | `122.226.167.0–122.226.167.255` in `122.226.167.0/24` |
| 90 | hangzhou-juzheng,huzhou,zhejiang | 256 | 0.0016% | 1 | `60.12.108.0–60.12.108.255` in `60.12.96.0/19` |
| 91 | SUOBEIZIXUN,HANGZHOU,ZHEJIANG | 255 | 0.0016% | 1 | `124.160.41.0–124.160.41.254` in `124.160.32.0/19` |
| 92 | Huzhou Department of Education | 240 | 0.0015% | 6 | `122.225.119.0–122.225.119.3` in `122.225.118.0/23`<br>`122.225.119.4–122.225.119.7` in `122.225.118.0/23`<br>`122.225.119.8–122.225.119.15` in `122.225.118.0/23` |
| 93 | JIANGSHANGUANGDIAN,QUZHOU,ZHEJIANG | 224 | 0.0014% | 3 | `123.157.97.192–123.157.97.223` in `123.157.64.0/18`<br>`123.157.100.0–123.157.100.127` in `123.157.64.0/18`<br>`221.12.138.128–221.12.138.191` in `221.12.136.0/21` |
| 94 | Jiande Internet monitoring project | 224 | 0.0014% | 3 | `115.236.85.64–115.236.85.127` in `115.236.84.0/22`<br>`115.236.85.128–115.236.85.255` in `115.236.84.0/22`<br>`183.129.136.0–183.129.136.31` in `183.129.128.0/20` |
| 95 | ete | 208 | 0.0013% | 27 | `61.130.68.192–61.130.68.199` in `61.130.64.0/21`<br>`61.130.70.8–61.130.70.11` in `61.130.64.0/21`<br>`61.130.70.120–61.130.70.123` in `61.130.64.0/21` |
| 96 | Ningbo Wanli College | 196 | 0.0012% | 3 | `61.130.107.112–61.130.107.115` in `61.130.106.0/23`<br>`61.153.150.0–61.153.150.127` in `61.153.144.0/21`<br>`115.231.49.0–115.231.49.63` in `115.231.48.0/23` |
| 97 | COMPUTE CENTER OF ZHEJIANG STATISTICS BUREAU | 192 | 0.0012% | 4 | `115.236.160.128–115.236.160.255` in `115.236.160.0/20`<br>`122.224.144.32–122.224.144.63` in `122.224.144.0/23`<br>`183.129.164.128–183.129.164.143` in `183.129.164.0/22` |
| 98 | Huzhou Education Breau | 192 | 0.0012% | 2 | `220.189.232.0–220.189.232.127` in `220.189.224.0/20`<br>`220.189.232.128–220.189.232.191` in `220.189.224.0/20` |
| 99 | The Daily Newspaper Co. | 192 | 0.0012% | 2 | `60.190.99.0–60.190.99.127` in `60.190.96.0/21`<br>`60.190.99.128–60.190.99.191` in `60.190.96.0/21` |
| 100 | YUANLIANTONG,HANGZHOU,ZHEJIANG | 192 | 0.0012% | 6 | `124.160.86.0–124.160.86.31` in `124.160.80.0/21`<br>`124.160.86.32–124.160.86.63` in `124.160.80.0/21`<br>`124.160.86.64–124.160.86.95` in `124.160.80.0/21` |

其余 52,904 个较小登记主体标签未在 Markdown 展开，可在完整 gzip JSON 中查询。
