# Ticket VS_0008 — Reward Labeling (Price Windows + Costs) for Optimizer

## Goal

Make the optimizer meaningful by generating `reward` labels from market prices **after costs**.

This ticket adds a labeling pipeline that:
- takes `paper_log` events (JSONL)
- fetches (or replays) prices for defined windows (e.g. T+5m / T+30m / close)
- computes net return after costs and emits:
  - `reward` (0/1) for bandit
  - optional continuous metrics for future upgrades

## Scope

In scope:
- Define evaluation windows + cost model (commission/slippage/spread proxy).
- Pick a compliant/cheap-first price source for the labeled instruments (repo/ETF/CB to start).
- Output a `labels.jsonl` (or enriched paper log) with deterministic schema.
- Unit tests for cost math + window handling.

Out of scope:
- Auto execution
- Full backtester

## Acceptance (3 steps)

1) `go test ./...` (or `python -m ...` if implemented in Python) passes.
2) Given a small paper_log sample, labeling produces `reward` fields with no crashes.
3) Feeding labeled output into `value-sniffer-radar-optimizer` changes the allocation ranking (non-trivial).

## Rollback

- Revert the labeling tool and schema changes.

