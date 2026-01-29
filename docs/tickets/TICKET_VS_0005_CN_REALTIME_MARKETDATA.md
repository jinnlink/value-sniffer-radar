# Ticket VS_0005 — CN Realtime Marketdata Adapter (Repo First)

## Goal

We want “实时嗅探”能力（支撑每天约 30 条 `tier=action` 报警），而当前 Tushare 主要是日频/慢数据。

This ticket adds a **realtime marketdata adapter** so signals can choose:
- **Slow/reference**: Tushare (lists, EOD, fundamentals)
- **Realtime/snapshot**: a cheap-first provider for a small watchlist

Start with **reverse repo** (204001/131810) because:
- signal logic清晰
- 执行窗口集中（临近收盘/节假日前后）

## Scope

In scope:
- Define a small `marketdata` interface (realtime snapshot) usable by signals.
- Implement one cheap-first realtime provider (to be decided in spec step), for:
  - repo quotes/rates for `204001.SH` / `131810.SZ`
- Add a new signal instance type (or upgrade `cn_repo_sniper`) to use realtime snapshots when enabled.
- Add strict throttling / polling interval + backoff to reduce ban risk.
- Add paper_log events with enough fields to evaluate latency/freshness (timestamp + provider + raw fields).

Out of scope:
- Auto execution / broker APIs
- Full-market scanning of 5000+ symbols
- “无风险套利”承诺

## Decision points (must be resolved in spec)

1) Realtime provider choice (cheap-first):
   - Option A: Eastmoney HTTP snapshot (Go implementation)
   - Option B: Tencent/Sina snapshot (Go implementation)
   - Option C: Local Python sidecar (AkShare) + HTTP to Go
2) Rate meaning:
   - If provider returns repo rate/yield directly: use it
   - If only returns price fields: define mapping clearly
3) Compliance/ToS guardrails:
   - personal use; conservative polling; no redistribution; no multi-user SaaS

## Acceptance (3 steps)

1) New config supports enabling realtime provider + poll interval, and a repo-realtime signal:
   - `Test-Path configs\\config.example.yaml`
2) Unit test(s) cover:
   - URL building/parsing and basic throttling behavior (no live network required)
3) Manual smoke (optional, depends on market time/network):
   - Run for a single symbol watchlist and print a snapshot within configured timeout.

## Rollback

- Remove new `marketdata` package, revert signal/config changes.

