# VS_0007 10_spec_keeper

## Pass/Fail gate

Pass if:
- Realtime-only mode can start without `TUSHARE_TOKEN` when:
  - only realtime signals are enabled (e.g. `cn_repo_realtime`)
  - `engine.trade_date_mode=fixed`
- `configs/config.example.yaml` uses correct Eastmoney repo scaling by default (`rate_divisor: 1000.0`).
- `go test ./...` passes.

Fail if:
- Token requirement is removed for cases that still need Tushare (e.g. `trade_date_mode=latest_open` or Tushare-backed signals enabled).

