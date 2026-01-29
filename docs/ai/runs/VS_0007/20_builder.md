# VS_0007 20_builder

## What changed

1) Conditional Tushare token requirement:
- Only require `TUSHARE_TOKEN` when it is actually needed:
  - `engine.trade_date_mode=latest_open` (needs trade_cal), or
  - any enabled signal uses Tushare-backed data (`cb_*`, `fund_premium`, `cn_repo_sniper`)

2) Eastmoney repo scaling default:
- Updated `configs/config.example.yaml` Eastmoney repo provider to `rate_divisor: 1000.0` so `f43=1600` maps to `1.600%`.

## Files changed

- `internal/config/config.go`
- `internal/engine/engine.go`
- `configs/config.example.yaml`

## Acceptance (copy/paste)

1) `go test ./...`
2) Realtime-only startup without token:
   - Ensure `TUSHARE_TOKEN` is unset
   - Use `engine.trade_date_mode: fixed`
   - Enable only `cn_repo_realtime` signals
3) `rg -n \"rate_divisor: 1000\\.0\" configs\\config.example.yaml`

