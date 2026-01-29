# Ticket VS_0005 — CN Realtime Marketdata Adapter（多源融合校验，Repo 优先）

## Goal

把项目从“慢数据报警”升级为“近实时（秒级）嗅探”，但仍坚持：**只做提示/报警，不自动交易，不承诺收益**。

本票交付一个关键底座：引入 `marketdata` 近实时行情适配层，并实现**多源交叉验证（fusion）+ 自动降级**；先落地在逆回购（repo）小品类，后续再扩展到转债/ETF/LOF。

## Non-goals（明确不做）

- 不做自动下单/券商 API / GUI 模拟点击
- 不做全市场 5000+ 扫描（只做小 watchlist）
- 不做“无风险套利”宣传

## Success Metrics（本票可验收的结果）

- 能稳定产出 `Snapshot`（含 `ts/provider/symbol/value`）并可被信号使用
- “多源一致性”不满足时**不发 `tier=action`**（最多 `tier=observe` 并写明原因）
- `paper_log` 中能复盘：触发时刻、各源原始值、融合后值、冲突/降级原因

## Scope

In scope:
- 新增 `internal/marketdata`（或同等目录）定义统一接口：
  - `Provider`：单源拉取快照（带 timeout）
  - `Snapshot`：统一字段（至少 repo 的 `rate_pct` + 原始字段 `raw`）
- 至少实现 **2 个 cheap-first provider** 用于交叉验证（只覆盖 repo watchlist）：
  - 方案候选：Eastmoney / Tencent / Sina / Python sidecar(AkShare)
- 实现 `MultiSourceProvider`：
  - 并发拉取（per-provider timeout）
  - 统一归一（把不同源字段映射到 `rate_pct`）
  - 多源一致性校验（consensus + outlier）
  - provider 可靠性评分 + 熔断（自动降级）
- 先支持 repo watchlist：`204001.SH` / `131810.SZ`
- 在 repo 信号上接入 realtime：
  - 新增信号类型（建议：`cn_repo_realtime`），或增强 `cn_repo_sniper`（需在 spec 阶段拍板）
- 轮询节流/退避：固定轮询间隔 + 抖动(jitter) + 失败退避(backoff) + 单源熔断(circuit breaker)
- `paper_log` 增强：记录 `providers[]` + `consensus` + `confidence` + `reason`

Out of scope:
- 交易执行
- 其它品类的实时化（转债/ETF/LOF）——后续票

## Advanced Algorithm（先进但可落地：多源融合 + 可信度）

### 1) 多源融合（robust consensus）

对同一 `symbol` 在同一轮询周期得到多个候选值 `x_i`（repo 的 `rate_pct`）：

- **质量过滤**：丢弃 stale（`now-ts > staleness_sec`）、缺字段、越界值（例如 rate 不在 `[0, 20]`）
- **共识值**：对剩余值取 `median` 作为 `consensus`
- **一致性判定**：若 `|x_i - consensus| <= max_abs_diff` 的来源数 >= `required_sources` → `confidence=PASS`
- **动作门槛**：`confidence=PASS` 且 `consensus >= min_yield_pct` 才允许发 `tier=action`
- **降级**：
  - `confidence=FAIL`：只发 `tier=observe`（或只记日志），并写明 `conflict`/`stale`/`missing`
  - 只有单源：默认不发 action（除非配置允许 `allow_single_source_action=true`，默认 false）

### 2) 在线可靠性评分（provider reliability score）

维护每个 provider 的 `score∈[0,1]`（指数衰减更新）：
- 命中共识（接近 median）→ `score` 上升
- 超时/无效/经常 outlier → `score` 下降

用途：
- 优先拉取高分源；低分源抽检
- 连续失败达到阈值触发熔断：短时间停用该源

### 3) 防抖/确认（anti-noise confirm）

可选 “连续 k 次确认”：
- 连续 `k` 次（例如 2 次）都满足 `consensus>=min_yield_pct` 才 action
- `k` 由配置控制（repo spike 可能很短，默认 k=1）

## Decision Points（Spec Keeper 必须拍板）

1) Provider 组合（默认建议 2 个）
   - A: Eastmoney + Tencent（全 Go）
   - B: Eastmoney + Sina（全 Go）
   - C: Go 主进程 + Python sidecar(AkShare) 作为第二源（更稳但更复杂）
2) `rate_pct` 的定义（必须可解释、可复盘）
   - 若 provider 直接给利率/收益率：直接用
   - 若只给价格/盘口字段：必须明确映射与单位，并在 `raw` 中保留原字段
3) 合规/ToS 护栏
   - 仅个人使用；保守轮询；不做多用户服务；不在仓库写入任何账号/密钥

## Deliverables（落地清单）

- `configs/config.example.yaml` 新增 `marketdata` 示例配置（providers、timeout、interval、required_sources、max_abs_diff、staleness_sec、circuit_breaker、confirm_k）
- Repo realtime 信号（`cn_repo_realtime` 或增强版 `cn_repo_sniper`）：
  - 使用 `MultiSourceProvider` 的 `consensus` + `confidence`
  - 产出事件 tags：`kind=repo`、`strategy=yield_spike`、`tier=...`、`confidence=...`
- `paper_log` 补充关键字段（用于未来评估“哪个源更稳、延迟多大、冲突多不多”）
- 单测（不依赖真实网络）：
  - 归一/解析
  - 多源融合/一致性判定
  - 可靠性评分与熔断

## Acceptance (3 steps)

1) `Test-Path configs\\config.example.yaml` 且示例包含 `marketdata` + repo realtime signal
2) `go test ./...`（至少覆盖融合/一致性/熔断的单测，无需联网）
3) 可选手工 smoke（受网络/交易时间影响）：对单一 symbol 获取一次 snapshot 并打印（或输出 observe 事件）

## Rollback

- 删除 `internal/marketdata`（或新增目录）并回退信号/配置变更
- 保持现有 Tushare-only 信号不受影响

