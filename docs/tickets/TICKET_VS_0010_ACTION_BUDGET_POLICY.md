# Ticket VS_0010 — Action Budget + Quality Gate (target ~30 action/day)

## Goal

Turn “broad coverage” into “~30 actionable alerts/day” by adding a **policy layer** that:
- enforces action budget (daily + per-run) without relying on “manual discipline”
- applies a consistent **net-edge** (after-cost) filter so “action” means “net advantage likely”
- keeps everything alerts-first (no auto execution)

## Scope

In scope:
- Add a policy module (or engine-stage) that can:
  - down-rank / downgrade events from `tier=action` → `tier=observe` if they don’t meet net-edge conditions
  - enforce action budgets:
    - global `action_max_events_per_day` (already exists) + per-signal quotas (new)
    - per-run caps remain as last-resort
- Define a minimal net-edge schema for events (in `event.data`), e.g.:
  - `expected_edge_pct` (expected advantage / deviation)
  - `spread_pct` (bid/ask spread proxy, if available)
  - `slippage_pct` (config default)
  - `fee_pct` (config default, market-specific)
  - compute `net_edge_pct = expected_edge_pct - spread_pct - slippage_pct - fee_pct`
- Add unit tests for:
  - budget accounting (daily + per-signal)
  - net-edge gating (edge below threshold never stays action)

Out of scope:
- New market data providers
- Auto execution / auto trading
- Full backtester (still alerts-first + paper log)

## Acceptance (3 steps)

1) `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2) With a config that sets `action_max_events_per_day: 30`, a synthetic burst of events is capped to <= 30 action/day.
3) Reported (paper log) events include `net_edge_pct` and the policy downgrades events that fail the threshold.

## Rollback

- Revert policy layer integration and any schema additions.

