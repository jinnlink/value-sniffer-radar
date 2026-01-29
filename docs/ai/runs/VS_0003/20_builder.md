# VS_0003 20_builder

## What changed

- Added `cn_repo_sniper` signal (Tushare `repo_daily`) to alert when reverse repo weighted rate (%) exceeds a threshold, optionally only within a time window.
- Wired the new signal type into the signal factory.
- Updated config schema + example config and README signal list.

## Files changed

- `internal/signals/cn_repo_sniper.go` (new)
- `internal/signals/signal.go`
- `internal/config/config.go`
- `configs/config.example.yaml`
- `README.md`

## How it works (minimal)

- Data source: Tushare `repo_daily` with `ts_code` + `trade_date`.
- Rate field: prefers `weight` (加权价), falls back to `close` if `weight` is missing.
- Filters:
  - `min_yield_pct`: rate threshold
  - optional `window_start` / `window_end` (HH:MM or HHMM)
  - optional `min_amount` (成交额) when provided by Tushare
- Emits standard `notifier.Event` with tags: `kind=repo`, `strategy=yield_spike`, `tier=...`.

## Acceptance (copy/paste)

1) `rg -n "cn_repo_sniper" internal configs README.md`
2) `go test ./...`
3) (Live smoke) run with fixed date + low threshold to force an event:
   - `Copy-Item .\\configs\\config.example.yaml .\\config.yaml`
   - edit `config.yaml`:
     - `engine.trade_date_mode: "fixed"`
     - `engine.fixed_trade_date: "20200804"`
     - `signals[] cn_repo_sniper_action.min_yield_pct: 0.1`
   - `go run .\\cmd\\value-sniffer-radar -config .\\config.yaml`

## Rollback

- Delete `internal/signals/cn_repo_sniper.go`
- Remove `cn_repo_sniper` case from `internal/signals/signal.go`
- Remove new config fields from `internal/config/config.go`
- Revert `configs/config.example.yaml` + `README.md`

