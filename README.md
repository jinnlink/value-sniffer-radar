# Value Sniffer Radar（只报警的机会嗅探雷达）

目标：合规市场里，把“极端偏离 / 错价 / 异常状态”尽快报警出来；**不自动交易**，不承诺收益。

## 特点

- **双层信号桶**：`tier=observe`（广覆盖）+ `tier=action`（更可执行）。
- **cheap-first 数据策略**：Tushare（慢/历史/静态）+ 免费实时源（腾讯/东财等，HTTP 轮询）。
- **闭环进化**：`paper_log` → `labeler` → `optimizer` → 反哺阈值/配额（持续改进，不靠感觉）。
- **工业化推进**：严格一票一推进（见 `SPEC.md` + `docs/tickets/QUEUE.md`）。

## 内置信号（当前）

- `cb_premium`：可转债价格 vs 转股价值（溢价率）极端报警
- `cb_double_low`：可转债“双低”（价格 + 溢价率）报警
- `fund_premium`：场内基金（ETF/LOF）价格 vs NAV（溢价率）极端报警
- `cn_repo_sniper`：逆回购利率（Tushare repo_daily 加权价）阈值报警（现金管理/利率雷达）
- `cn_repo_realtime`：逆回购实时利率（多源一致性融合）阈值报警（需要开启 `marketdata`）

同一 `type` 可以配置多次，用 `signals[].name` 区分实例（示例见 `configs/config.example.yaml`）。

## 快速开始

1) 安装 Go（建议 1.22+）
2) 配置 Tushare Token（只放环境变量，不要写进仓库）：

```powershell
$env:TUSHARE_TOKEN="你的token"
```

3) 运行：

```powershell
Copy-Item .\configs\config.example.yaml .\config.yaml
go run .\cmd\value-sniffer-radar -config .\config.yaml
```

## 通知（推荐：AstrBot / QQ）

你的机器上已有 AstrBot 体系（`ai-value` / `ai-value-core`），它用“文件队列”推送到 QQ。

在 `config.yaml` 启用 `aival_queue`，并把 `queue_dir` 指向：
`E:\Program Files (x86)\bot\share\ai-value-core\queue`

## 去重/限流/配额（关键）

基础限流：
- `engine.dedupe_seconds`：内容级去重 TTL
- `engine.action_symbol_cooldown_seconds` / `engine.observe_symbol_cooldown_seconds`
- `engine.action_max_events_per_run` / `engine.observe_max_events_per_run`
- `engine.action_max_events_per_day` / `engine.observe_max_events_per_day`
- `signals[].min_interval_seconds`：单信号最小计算间隔

VS_0010 新增（把“每天 30 条 action”做成制度）：
- `engine.action_max_events_per_signal_per_day`：按信号分配 action 配额（超额会降级到 observe）

## Action 质量闸门：净优势（net edge）

VS_0010 增加统一公式（写入 `event.data`）：

`net_edge_pct = expected_edge_pct - spread_pct - slippage_pct - fee_pct`

- 当 `engine.action_net_edge_min_pct > 0` 时：不达标的 `tier=action` 会被**自动降级**为 `tier=observe`。
- 默认是关闭的（`action_net_edge_min_pct: 0.0`），保证兼容老配置。

## 闭环（paper → labeler → optimizer）

1) 开 `paper_log`（配置里启用 notifier `paper_log`）
2) 运行 labeler 产出 `labels.repo.jsonl`：

```powershell
go run .\cmd\value-sniffer-radar-labeler -config .\config.yaml -in .\state\paper.jsonl -out .\state\labels.repo.jsonl
```

3) 运行 optimizer 做配额/优先级建议：

```powershell
go run .\cmd\value-sniffer-radar-optimizer -in .\state\paper.jsonl -labels .\state\labels.repo.jsonl -label-window-sec 30 -slots 30
```

下一步（VS_0011 Backlog）：把 optimizer 的结果自动写回“按信号配额/调度策略”，减少手工调参。

## 一键日常闭环（推荐）

VS_0012 提供 `tools/daily_loop.ps1`，把日常操作收敛成一个命令：label → report → reco。

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\tools\daily_loop.ps1 -Config .\config.yaml
```

将 reco 接入运行时（VS_0011）：
- 在 `config.yaml` 配置：`engine.reco_path: .\state\optimizer.reco.json`
- 重启 radar 进程后生效（目前是启动时加载一次）。

## LLM 事件增强（可选，不在热路径）

VS_0013 提供 `cmd/value-sniffer-radar-llm`：只做“解释/摘要/清单”，不改 `tier`，不做任何交易决策。

### 用 CLI 执行（推荐先用这个跑通）

要求：你的 CLI 必须从 stdin 读 prompt，并在 stdout 输出严格 JSON（enrich 模式）。

推荐（最稳）：用 `codex exec --output-schema` 包装成“stdin→stdout 纯 JSON”：

```powershell
go run .\cmd\value-sniffer-radar-llm `
  -mode enrich `
  -provider cli `
  -cli-cmd powershell.exe `
  -cli-args "-NoProfile -ExecutionPolicy Bypass -File tools\\llm_wrappers\\codex_enrich.ps1" `
  -in .\state\paper.jsonl `
  -out .\state\llm.enriched.jsonl
```

你也可以用 `gemini` / `claude` / `opencode`，但它们不一定能稳定做到“stdout 只输出 JSON”。如果输出里混了额外文本，`value-sniffer-radar-llm` 会尝试从中提取第一个 JSON 对象解析。

```powershell
go run .\cmd\value-sniffer-radar-llm `
  -mode enrich `
  -provider cli `
  -cli-cmd gemini.cmd `
  -cli-args "" `
  -in .\state\paper.jsonl `
  -out .\state\llm.enriched.jsonl
```

### 用云端 API（OpenAI-compatible Chat Completions）

```powershell
$env:LLM_API_KEY="YOUR_KEY"
go run .\cmd\value-sniffer-radar-llm `
  -mode enrich `
  -provider api `
  -api-base-url "https://api.openai.com/v1" `
  -api-model "gpt-4o-mini" `
  -in .\state\paper.jsonl `
  -out .\state\llm.enriched.jsonl
```
