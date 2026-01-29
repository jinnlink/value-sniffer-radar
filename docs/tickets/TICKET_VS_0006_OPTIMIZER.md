# Ticket VS_0006 — Online Optimizer (Bandit + Dynamic Polling + Auto-threshold Suggestions)

## Goal

“多尝试但不乱来”：在不自动交易的前提下，让系统能**持续验证**哪些信号/阈值/数据源组合更有效，并给出可审核的“建议改配置”输出。

本票目标：
- 在线学习（bandit / Bayesian）分配 `tier=action` 的名额
- 动态轮询（在预算内把采样打到“最可能产出 action 的地方”）
- 只输出**建议**（PR/ticket/patch），不自动改配置、不自动下单

## Scope

In scope:
- 基于 `paper_log` 的标签：定义统一的 `eval_windows`（例如 T+5m / T+30m / close）和 `net_return_after_costs` 的近似计算框架（先用价格源占位或后续票接入价格）
- Bandit 策略（建议 Thompson Sampling）：
  - Arms：`(signal_instance, threshold_variant, provider_combo)` 的组合
  - Reward：`net_return_after_costs` 或 “命中率 proxy”
  - Output：每日报告 + 建议阈值/名额分配
- 动态轮询（budgeted sampling）：
  - 输入：provider `score`、时段窗口、最近波动/冲突率
  - 输出：每个 symbol 的下一次 poll 计划（优先级队列）
- 产物：
  - `state/optimizer/*.json`（本地）+ `docs/optimizer/*.md`（可提交的总结）

Out of scope:
- 自动交易
- 大规模 ML（深度学习等）

## Acceptance (3 steps)

1) `python tools\\paper_eval.py --in .\\state\\paper.jsonl` 输出包含新增字段统计（如果本票改了 paper 格式）
2) `go test ./...`（优化器核心逻辑有单测：bandit 更新/采样/建议生成）
3) 运行一次 dry-run（不联网也可）：用 synthetic paper_log 输入，产出一份“建议配置变更”的 Markdown 报告

## Rollback

- 删除 optimizer 相关文件与配置，不影响原有报警链路。

