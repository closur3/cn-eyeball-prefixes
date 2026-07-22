# 中国电信 IPv6 准入流程

本目录由 `ipv6` 开发分支生成；在人工审核并明确合并前，不作为正式 IPv6 发布结果。

## 固定政策与动态事实

生成器将固定政策与每次构建获取的动态事实分开。

固定政策：

- 允许的 APNIC `inet6num` 用途描述：
  - `Chinatelecom IPv6 address for fixed broadband`；
  - `Chinatelecom IPv6 address for mobile`。

每次构建更新的动态事实：

- APNIC `apnic.db.inet6num.gz` 登记层级；
- RIPE RISWhois 当前 IPv6 BGP 前缀及 Origin；
- IPtoASN IPv6 Origin 国家和描述元数据。

`manifest.json` 中 `registry_admission.matched_inet6num_prefixes` 展示的是本次构建从完整 APNIC 数据库动态发现的登记范围，不是生成器中写死的准入前缀。中国电信母段也不是准入输入。

## 准入算法

1. 读取完整的 APNIC `inet6num` 数据库。
2. 对重叠登记采用最具体前缀优先。
3. 根据两个精确用途描述生成本次构建的固网、移网有效准入范围。
4. 读取与动态准入范围相交的当前 IPv6 BGP 前缀。
5. 要求整个 BGP 前缀均由同一种获准 APNIC 用途完整覆盖。
6. 要求所有观测到的 Origin 均为中国境内的中国电信 ASN。
7. 原样输出 BGP 前缀，不聚合，也不拆分。

如果更具体的 APNIC 登记使用其他描述，它会覆盖父级登记并在准入空间中形成缺口。任何跨越该缺口的 BGP 前缀都会整段拒绝。

## 审核依据

`manifest.json` 记录：

- 所有动态上游的文件大小和 SHA-256；
- 作为固定政策的 APNIC 用途描述；
- 本次构建实际命中的 `inet6num` 前缀；
- 命中登记数和最终有效范围数；
- 按用途统计的获准 BGP 前缀数；
- 按原因统计的拒绝 BGP 前缀数。

供使用者读取的结果为 `operators/chinatelecom.txt`。
