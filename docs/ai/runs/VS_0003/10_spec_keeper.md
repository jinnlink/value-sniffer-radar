# VS_0003 10_spec_keeper

## Pass/Fail gate (this ticket)

Pass if:
- `cn_repo_sniper` can be configured as a new `signals[].type`.
- It supports `tier=action|observe` via separate config instances.
- It only uses existing infra (Tushare client + existing notifiers).
- Example config + README are updated.

Fail if:
- Any scope creep (auto execution / new external data sources / unrelated refactors).
- No runnable acceptance steps.

## Scope guardrails

- In scope:
  - `internal/signals`: add `cn_repo_sniper` signal backed by Tushare `repo_daily`.
  - `internal/config`: add minimal config fields for thresholds + time window.
  - Docs: `configs/config.example.yaml`, `README.md`.
- Out of scope:
  - Trading execution, broker APIs, GUI automation.
  - HFT / “true arbitrage” claims.

## Acceptance reminders

1) `rg -n "cn_repo_sniper" internal configs README.md`
2) `go test ./...`
3) Fixed trade_date smoke run (threshold forced low) shows at least 1 event emitted to stdout.

