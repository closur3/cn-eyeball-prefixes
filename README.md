# 中国三网终端用户接入网 IP 前缀列表

本仓库自动维护中国电信、中国移动和中国联通终端用户接入网的 IPv4、IPv6 CIDR 候选列表，可用于 ACL、路由策略和流量分析。

这里的“终端用户接入网”主要指固定宽带接入和移动网络接入，不包括 IDC、云计算、CDN、企业专线、机构统一出口和专用骨干。结果是依据当前公开 BGP 与注册数据生成的 **best-effort 候选集**，不是地址实际用途、归属或安全性的保证。使用方应按自己的风险边界叠加认证、限速和审计策略，而不应把 IP 白名单当作唯一信任依据。

本文统一使用“固定宽带接入”和“移动网络接入”这两个术语；配置及审计数据中的机器字段 `fixed` / `fixed_broadband` 和 `mobile` 分别对应这两类接入。

## 获取列表

公开文件统一位于 [`lists/`](lists/)：

| 路径 | 内容 |
| --- | --- |
| [`lists/ipv4/cn.txt`](lists/ipv4/cn.txt) | 三网 IPv4 全国合表 |
| [`lists/ipv4/chinatelecom.txt`](lists/ipv4/chinatelecom.txt) | 中国电信 IPv4 |
| [`lists/ipv4/chinamobile.txt`](lists/ipv4/chinamobile.txt) | 中国移动 IPv4 |
| [`lists/ipv4/chinaunicom.txt`](lists/ipv4/chinaunicom.txt) | 中国联通 IPv4 |
| [`lists/ipv4/provinces/`](lists/ipv4/provinces/)`<pinyin>.txt` | IPv4 省级合表 |
| [`lists/ipv6/cn.txt`](lists/ipv6/cn.txt) | 三网 IPv6 全国合表 |
| [`lists/ipv6/chinatelecom.txt`](lists/ipv6/chinatelecom.txt) | 中国电信 IPv6 |
| [`lists/ipv6/chinamobile.txt`](lists/ipv6/chinamobile.txt) | 中国移动 IPv6 |
| [`lists/ipv6/chinaunicom.txt`](lists/ipv6/chinaunicom.txt) | 中国联通 IPv6 |
| [`lists/ipv6/provinces/`](lists/ipv6/provinces/)`<pinyin>.txt` | IPv6 省级合表 |
| [`lists/manifest.json`](lists/manifest.json) | 公开文件的内容标识、CIDR 数和 SHA-256 |

省级文件覆盖中国大陆 31 个省级行政区，使用拼音文件名，例如 `beijing.txt`、`guangdong.txt`、`shaanxi.txt` 和 `xinjiang.txt`。每个文本文件每行一个规范化 CIDR，按地址排序，文件内部无重叠。

[`lists/manifest.json`](lists/manifest.json) 的 `content_id` 标识整套公开列表的确定内容；下载后可用其中每个文件的 SHA-256 校验完整性。

## 收录范围

| 网络类型 | 策略 |
| --- | --- |
| 三网固定宽带接入和移动网络接入地址 | 满足相应地址族的准入条件后保留 |
| IDC、托管、云计算、CDN、VPS | IPv4 有可靠 ASN、CIDR 或 APNIC 证据时排除 |
| 企业专线、企业固定地址、机构统一出口 | 非目标；IPv4 有可靠证据时排除 |
| 政务、行业专网、VPN/MPLS、IoT/M2M、OA、监控 | IPv4 有明确用途登记时排除 |
| CN2、CUII 等专用精品骨干 | IPv4 按 Origin ASN 排除；IPv6 的已确认接入地址段允许由其承载 |
| 证据不足或用途混合的地址 | 保守保留或拒绝，具体取决于地址族的准入模型 |

IPv4 与 IPv6 的公开目录相同，但两者的数据条件和登记方式不同，不能把一方的判定规则直接套用到另一方。

## 生成方法

### IPv4

IPv4 生成器先用 IPtoASN 识别当前 Origin ASN，并与 `gaoyifan/china-operator-ip` 的中国地址边界取交集。随后依次应用：

1. `generator/config/operators.json` 中的三网 ASN 识别与显式例外；
2. 引用自 IP-Data 的七家云厂商独立 CIDR 强排除；
3. APNIC `inetnum`、`organisation`、`aut-num` 和 `route` 登记证据；
4. RIPE RISWhois 的当前多 Origin、可见度和登记冲突证据；
5. 同运营商 APNIC 父级准入及受安全阈值约束的 BGP 登记冲突孔洞修复；
6. GeoCN 的省级归属切分。

云、IDC、专线、专网及独立主体的强排除先于冲突修复执行，冲突修复不能把这些地址重新放回。全国表是三家运营商表的去重并集；IPv4 省级数据仅用于切分，不决定全国表是否收录，因而无法定位到省份的地址仍可留在全国表中。

### IPv6

IPv6 生成器从 RIPE RISWhois 读取当前 BGP 前缀和全部 Origin，并用 IPtoASN 的国家与 ASN 描述进行三网分类。一个 BGP 前缀只有在完整落入接入地址准入范围、且所有 Origin 都匹配同一家运营商时才会收录。

- 中国电信准入范围在每次构建时从完整 APNIC `inet6num` 数据库发现，匹配固定宽带接入和移动网络接入的精确登记描述；更具体的 APNIC 登记会覆盖较宽的登记。
- 中国移动和中国联通使用代码中审计过的全国接入地址边界。
- `generator/config/ipv6-province-prefixes.json` 是 31 个省级 IPv6 分配的仓库内事实表，同时区分运营商、固定宽带接入与移动网络接入；公开列表不再拆分接入类型。
- 每个运营商列表必须完全落入省级分配表。三网互不重叠，运营商并集和省级并集都必须严格等于 `lists/ipv6/cn.txt`。

IPv6 的详细准入来源、Origin、拒绝原因和输入摘要作为 CI 审计产物上传并保留 30 天，不写入精简的公开 manifest。

## 外部上游

下面列出生成流程下载和引用的全部外部数据上游。仓库不提交原始上游文件；每次运行在隔离的临时目录重新下载。

| 上游来源 | 下载项 | 地址族 | 用途 |
| --- | --- | --- | --- |
| [gaoyifan/china-operator-ip](https://github.com/gaoyifan/china-operator-ip/tree/ip-lists) | [`china.txt`](https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china.txt) | IPv4 | 限制中国地址候选边界 |
| [IPtoASN](https://iptoasn.com/) | [`ip2asn-v4.tsv.gz`](https://iptoasn.com/data/ip2asn-v4.tsv.gz) | IPv4 | 当前前缀 Origin ASN、国家和 ASN 描述 |
| [IPtoASN](https://iptoasn.com/) | [`ip2asn-v6.tsv.gz`](https://iptoasn.com/data/ip2asn-v6.tsv.gz) | IPv6 | Origin ASN 的国家和描述元数据 |
| [APNIC Whois Database](https://ftp.apnic.net/apnic/whois/) | [`apnic.db.inetnum.gz`](https://ftp.apnic.net/apnic/whois/apnic.db.inetnum.gz) | IPv4 | 地址登记层级、用途与父级准入 |
| APNIC Whois Database | [`apnic.db.inet6num.gz`](https://ftp.apnic.net/apnic/whois/apnic.db.inet6num.gz) | IPv6 | 中国电信接入地址准入范围 |
| APNIC Whois Database | [`apnic.db.aut-num.gz`](https://ftp.apnic.net/apnic/whois/apnic.db.aut-num.gz) | IPv4 | ASN 登记与独立主体关联 |
| APNIC Whois Database | [`apnic.db.organisation.gz`](https://ftp.apnic.net/apnic/whois/apnic.db.organisation.gz) | IPv4 | 组织句柄和结构化组织名称 |
| APNIC Whois Database | [`apnic.db.route.gz`](https://ftp.apnic.net/apnic/whois/apnic.db.route.gz) | IPv4 | route Origin 与用途登记证据 |
| [RIPE RISWhois](https://ris.ripe.net/docs/ris-whois/) | [`riswhoisdump.IPv4.gz`](https://www.ris.ripe.net/dumps/riswhoisdump.IPv4.gz) | IPv4 | 当前 BGP 宣告、MOAS 与可见度 |
| RIPE RISWhois | [`riswhoisdump.IPv6.gz`](https://www.ris.ripe.net/dumps/riswhoisdump.IPv6.gz) | IPv6 | 当前 BGP 前缀及全部 Origin |
| [ljxi/GeoCN](https://github.com/ljxi/GeoCN) | [`GeoCN.mmdb`](https://github.com/ljxi/GeoCN/releases/latest/download/GeoCN.mmdb) | IPv4 | 最终地址池的省级切分 |
| [axpwx/IP-Data](https://github.com/axpwx/IP-Data/tree/master/provider) | [`aliyun-cidr-ipv4.txt`](https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/aliyun-cidr-ipv4.txt) | IPv4 | 阿里云 CIDR 强排除 |
| axpwx/IP-Data | [`tencent-cidr-ipv4.txt`](https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/tencent-cidr-ipv4.txt) | IPv4 | 腾讯云 CIDR 强排除 |
| axpwx/IP-Data | [`huawei-cidr-ipv4.txt`](https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/huawei-cidr-ipv4.txt) | IPv4 | 华为云 CIDR 强排除 |
| axpwx/IP-Data | [`ucloud-cidr-ipv4.txt`](https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/ucloud-cidr-ipv4.txt) | IPv4 | UCloud CIDR 强排除 |
| axpwx/IP-Data | [`ksyun-cidr-ipv4.txt`](https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/ksyun-cidr-ipv4.txt) | IPv4 | 金山云 CIDR 强排除 |
| axpwx/IP-Data | [`baidu-cidr-ipv4.txt`](https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/baidu-cidr-ipv4.txt) | IPv4 | 百度智能云 CIDR 强排除 |
| axpwx/IP-Data | [`jdcloud-cidr-ipv4.txt`](https://raw.githubusercontent.com/axpwx/IP-Data/master/provider/jdcloud-cidr-ipv4.txt) | IPv4 | 京东云 CIDR 强排除 |

IP-Data 的七个文件均来自其 [`provider/`](https://github.com/axpwx/IP-Data/tree/master/provider) 目录；生成器不使用覆盖范围更宽的 `all-cidr`。精确下载 URL 也保存在 [更新工作流](.github/workflows/update.yml) 和生成器的来源元数据中，二者应同步维护。

## 自动更新与验证

仓库以 `main` 为正式主线、`dev` 为开发线：代码、规则、文档和工作流变更先在 `dev` 验证，再由维护者合入 `main` 发布。

[更新工作流](.github/workflows/update.yml) 每天 UTC 08:08 在默认分支 `main` 定时运行，也支持在所选分支手动触发。向 `dev` 推送 `lists/` 之外的变更时也会触发完整流程；工作流生成的列表提交回其运行分支，不会自动把 `dev` 合入 `main`。

流水线会：

1. 下载全部上游并拒绝空文件或异常小的数据；
2. 运行全部 Go 测试和静态检查；
3. 在临时目录分别生成 IPv4、IPv6；
4. 使用独立校验器检查规范化、重叠、并集、省级映射和安全阈值；
5. 生成公开 manifest，并上传包含来源哈希和详细证据的审计产物；
6. 仅在公开列表确有变化时提交更新。

本地代码检查：

```bash
go -C generator test ./...
go -C generator vet ./...
```

完整再生成需要工作流中列出的全部上游文件；命令和参数以 [`.github/workflows/update.yml`](.github/workflows/update.yml) 为准。

## 许可证与第三方数据

本仓库自有的代码、配置、文档和汇编成果按 [MIT License](LICENSE) 发布。

外部上游数据不因被本项目下载、分析或引用而改变其权利归属和使用条款。再分发、商用或基于本仓库结果提供服务前，请同时检查各上游的许可证、数据库条款和署名要求。本项目不对上游数据的准确性、持续可用性或特定用途适用性作保证。
