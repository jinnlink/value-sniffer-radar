# VS_0003 00_context

## Ticket
- Queue line: - Ticket VS_0003 (`docs/tickets/TICKET_VS_0003_CN_REPO_SNIPER.md`): Add reverse repo (GC001/R-001) monitor (observe+action)
- Ticket doc: docs/tickets/TICKET_VS_0003_CN_REPO_SNIPER.md

## Do / Don't (scope guardrails)
- Do:
  - Add a new signal type `cn_repo_sniper` using Tushare `repo_daily` (weighted rate) for a small set of repo codes.
  - Support `tier=action|observe` via config instances (thresholds + min_interval + window).
  - Update `configs/config.example.yaml` and README signal list.
- Don't:
  - Don't add auto-execution / broker APIs.
  - Don't introduce new data sources (AkShare/web scraping) in this ticket.
  - Don't change unrelated signals or notifier behaviors.

## Decision points (0-3)
- Data source: use Tushare `repo_daily` weighted price (`weight`) as the repo rate (%).

## Expected minimal changes
- Files:
  - `internal/signals/cn_repo_sniper.go` (new)
  - `internal/signals/signal.go` (register signal)
  - `internal/config/config.go` (signal config fields)
  - `configs/config.example.yaml` (example config)
  - `README.md` (signal list)

## Acceptance (3 steps, copy/paste)
1. `rg -n "cn_repo_sniper" internal configs README.md`
2. (Requires Go toolchain) `go test ./...`
3. (Live smoke, requires Tushare access to repo_daily) `Copy-Item .\\configs\\config.example.yaml .\\config.yaml; notepad .\\config.yaml` then set:
   - `engine.trade_date_mode: "fixed"`
   - `engine.fixed_trade_date: "20200804"`
   - `signals[] cn_repo_sniper_action.min_yield_pct: 0.1`
   Then run: `go run .\\cmd\\value-sniffer-radar -config .\\config.yaml`

## Rollback
- Delete `internal/signals/cn_repo_sniper.go`, remove `cn_repo_sniper` wiring and config fields, revert docs/config changes.

