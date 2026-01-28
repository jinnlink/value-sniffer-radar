# Value Sniffer Radar（只报警的机会嗅探雷达）

目标：在你能参与的合规市场里，把“极端偏离 / 错价 / 异常状态”尽快报警出来；不自动交易，不承诺收益。

当前内置信号（Tushare Pro）：
- `cb_premium`：可转债价格 vs 转股价值（溢价率）极端报警
- `cb_double_low`：可转债“双低”（价格 + 溢价率）报警
- `fund_premium`：场内基金（ETF/LOF）价格 vs 基金净值（溢价率）极端报警

## 广覆盖 + 高质量（同时跑）

推荐用两级漏斗同时跑：
- `tier=observe`：广覆盖、低频、阈值更宽（主要用于“别漏掉”）
- `tier=action`：高质量、可执行、阈值更严（主要用于“尽量少而准”）

同一个 `type` 可以配置多次，用 `signals[].name` 区分实例（示例见 `configs/config.example.yaml`）。

## 准备

1) 安装 Go（建议 1.22+）
2) 注册 Tushare Pro，获取 token（放到环境变量，不要写进仓库）

PowerShell 示例：
```powershell
$env:TUSHARE_TOKEN="你的token"
```

## 运行

```powershell
Copy-Item .\configs\config.example.yaml .\config.yaml
go run .\cmd\value-sniffer-radar -config .\config.yaml
```

## 通知（推荐：AstrBot / QQ）

你这个目录里已经有现成的 AstrBot 插件体系（`ai-value` / `ai-value-core`）：它用“文件队列”做模块解耦。

Value Sniffer Radar 这边只要把通知通道改成 `aival_queue`，把 `queue_dir` 指向：
`E:\Program Files (x86)\bot\share\ai-value-core\queue`

AstrBot 插件 `ai_value` 会轮询该目录并推送到你配置的 QQ targets。

## 去重/限流

- `engine.dedupe_seconds`：内容级去重 TTL（默认 3600 秒）
- `engine.action_symbol_cooldown_seconds` / `engine.observe_symbol_cooldown_seconds`
- `engine.action_max_events_per_run` / `engine.observe_max_events_per_run`
- `signals[].min_interval_seconds`：单信号最小计算间隔（observe 慢扫、action 快扫）

## 评估（Paper）

可选启用 `paper_log`，把每条事件追加到 JSONL，方便后续做收益/回撤统计与策略淘汰。
