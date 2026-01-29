# Ticket VS_0007 — Runtime Usability Tweaks (Realtime-only mode)

## Goal

Two fixes to make realtime-only running smooth:

1) Allow running without `TUSHARE_TOKEN` when **no Tushare-backed signals** are enabled and `engine.trade_date_mode=fixed`.
2) Make Eastmoney repo default mapping correct by default (Eastmoney `f43` for repo often returns e.g. `1600` meaning `1.600%`).

## Scope

In scope:
- Conditional Tushare token requirement (only when needed).
- Update example config defaults for repo realtime (`rate_divisor: 1000.0`).
- Update docs/run artifacts for this ticket.

Out of scope:
- Any optimizer work (VS_0006)
- Trading execution

## Acceptance (3 steps)

1) With `marketdata.enabled=true`, only `cn_repo_realtime` enabled, and `engine.trade_date_mode=fixed`, startup does **not** require `TUSHARE_TOKEN`.
2) `go test ./...`
3) `configs/config.example.yaml` includes `rate_divisor: 1000.0` for the Eastmoney repo provider.

## Rollback

- Revert conditional token logic and config default change.

