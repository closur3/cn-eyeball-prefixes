# cn-operator-allowlist

面向 ACL 的中国大陆基础电信运营商 IPv4 CIDR 列表，仅收录中国电信、中国移动和中国联通的地址，并显式排除阿里云、腾讯云、华为云、UCloud、金山云、百度智能云和京东云上游 CIDR。仓库提供全国合表及 31 个省级行政区合表，可用作允许名单、路由规则、流量统计和地域分流等需要 CIDR 输入的场景。

列表按运营商归属筛选，并以云厂商 BGP CIDR 清单作为显式排除项；但不尝试识别或剔除运营商地址空间内的 IDC、企业专线或住宅宽带。

## 数据文件

| 文件 | 内容 |
| --- | --- |
| `data/cn.txt` | 中国电信、中国移动和中国联通的 IPv4 CIDR 去重合表 |
| `data/provinces/<pinyin>.txt` | 相应省级行政区内上述运营商的 IPv4 CIDR 去重合表 |
| `data/manifest.json` | 本次生成时间、上游 URL、输入文件 SHA-256，以及每个列表文件的路径和统计信息 |

省级文件以拼音命名，例如 `beijing.txt`、`guangdong.txt`、`shaanxi.txt`、`xinjiang.txt`。每个文本文件一行一个 CIDR，按地址排序，且文件内部不存在重叠网段。

## 生成规则

- 运营商候选采用 [gaoyifan/china-operator-ip](https://github.com/gaoyifan/china-operator-ip/tree/ip-lists) `ip-lists` 分支的 `chinanet.txt`、`cmcc.txt` 与 `unicom.txt`；仅保留同时出现在该上游 `china.txt`（起源 ASN 为中国 ASN）的 CIDR，以排除仅在路由中途经过运营商 ASN 的误收项。
- 云厂商排除项同时采用 [rezmoss/cloud-provider-ip-addresses](https://github.com/rezmoss/cloud-provider-ip-addresses) 的阿里、腾讯、华为、百度 IPv4 CIDR 文件，以及 [axpwx/IP-Data](https://github.com/axpwx/IP-Data) 的阿里云、腾讯云、华为云、UCloud、金山云、百度智能云和京东云独立 IPv4 CIDR 文件。任一来源显式列出即排除；不使用 IP-Data 的 `all-cidr` 集合，也不以 IP2region 或 IDC/托管标签进行推断。
- 省级归属采用 [lionsoul2014/ip2region](https://github.com/lionsoul2014/ip2region) 的 IPv4 源数据；它只用于地域切分，不参与运营商判定。
- 仅处理 IPv4 和中国大陆 31 个省级行政区；非中国大陆地址及无法归入省级行政区的网段不进入省级文件。
- 相邻或重叠网段会合并为最小 CIDR 集合。全国合表会去除三个上游运营商列表间可能存在的重复地址。

## 自动更新

[GitHub Actions](.github/workflows/update.yml) 每天 UTC 08:08 执行，也可从 Actions 页面手动运行。

工作流在 runner 临时目录下载上游数据、生成 `data/`、执行 CIDR 校验，并且仅当列表内容、上游来源或统计信息变化时提交更新。上游源文件不会被写入仓库或保留在本地工作区。
