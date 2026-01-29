# VS_0007 00_context

## Ticket
- Queue line: - Ticket VS_0007 (`docs/tickets/TICKET_VS_0007_RUNTIME_TWEAKS.md`): Realtime-only run without Tushare token + repo divisor default
- Ticket doc: docs/tickets/TICKET_VS_0007_RUNTIME_TWEAKS.md

## Do / Don't (scope guardrails)
- Do:
  - Allow running without `TUSHARE_TOKEN` when only realtime signals are enabled and trade date is fixed.
  - Fix Eastmoney repo `rate_divisor` default in example config to `1000.0`.
  - Keep changes minimal and scoped to runtime/config usability.
- Don't:
  - Don't modify strategy logic or add new providers.
  - Don't do any optimizer work (VS_0006).

## Decision points (0-3)
- Decide “needs Tushare” conditions: any Tushare-backed signal enabled OR `engine.trade_date_mode=latest_open`.

## Expected minimal changes
- Files:
  - `internal/config/config.go` (conditional token requirement)
  - `internal/engine/engine.go` (conditional client init)
  - `configs/config.example.yaml` (eastmoney repo divisor default)

## Acceptance (3 steps, copy/paste)
1) `go test ./...`
2) Create a local `config.yaml` enabling only `marketdata` + `cn_repo_realtime` (trade_date fixed), with no `TUSHARE_TOKEN`, then run:
   - `go run .\\cmd\\value-sniffer-radar -- -config .\\config.yaml`
3) `rg -n \"rate_divisor: 1000\\.0\" configs\\config.example.yaml`

## Rollback
- Revert conditional token logic and config default change.

