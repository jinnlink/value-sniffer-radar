# VS_0010 20_builder

## What changed
- Added a policy stage to compute `net_edge_pct` and (optionally) downgrade `tier=action` to `tier=observe` when net edge is below `engine.action_net_edge_min_pct`.
- Enhanced daily cap logic to:
  - downgrade when daily action cap is exceeded (instead of dropping)
  - enforce per-signal daily action quotas (`engine.action_max_events_per_signal_per_day`)
- Updated core signals to emit `expected_edge_pct` so net-edge can be computed:
  - repo realtime/sniper: `expected_edge_pct = rate - threshold_yield_pct`
  - cb/fund premium: `expected_edge_pct = abs(premium - threshold_premium_pct)`
  - cb double-low: `expected_edge_pct = threshold_double_low - double_low`
- Added unit tests covering:
  - net-edge gating
  - global + per-signal daily budgets

## Files touched
- `internal/config/config.go`
- `configs/config.example.yaml`
- `internal/engine/engine.go`
- `internal/engine/policy.go`
- `internal/engine/policy_test.go`
- `internal/signals/cn_repo_realtime.go`
- `internal/signals/cn_repo_sniper.go`
- `internal/signals/cb_premium.go`
- `internal/signals/fund_premium.go`
- `internal/signals/cb_double_low.go`

## Acceptance
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. Verify caps enforce <= 30 action/day in tests.
3. Verify `net_edge_pct` exists and failing action downgrades in tests.

## Rollback
- Revert the files above.

