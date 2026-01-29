# Ticket VS_0011 — Optimizer → Quotas + Runtime Scheduling

## Goal

Close the loop from `labels.repo.jsonl` to “daily ~30 action signals” by turning optimizer outputs into:
- per-signal daily quotas (and/or weights)
- runtime scheduling priorities (polling plan / evaluation cadence)

So the system learns which signals deserve scarce `action` slots.

## Scope

In scope:
- Extend optimizer output to emit a machine-readable recommendation file, e.g.:
  - `.\\state\\optimizer.reco.json` containing:
    - `primary_window_sec`
    - per-signal `mean_reward`, `n`, `suggested_daily_quota`
    - suggested `action_net_edge_min_pct` adjustments (optional)
- Add a small runtime loader that can:
  - read reco file
  - apply per-signal quotas automatically (override config map)
  - optionally adjust per-signal `min_interval_seconds` (or scheduling priority)
- Add tests for:
  - deterministic quota recommendations given fixed seed + labels
  - runtime quota enforcement using reco file

Out of scope:
- Auto trading execution
- New market data sources
- Full portfolio optimizer

## Acceptance (3 steps)

1) `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2) Given a small `labels.repo.jsonl`, optimizer produces a stable `optimizer.reco.json`.
3) Engine run uses reco quotas and produces ~30 action/day with fewer false positives (paper_eval improves).

## Rollback

- Stop loading reco file and revert optimizer output additions.

