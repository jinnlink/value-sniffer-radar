# VS_0005 00_context

## Ticket
- Queue line: - Ticket VS_0005 (`docs/tickets/TICKET_VS_0005_CN_REALTIME_MARKETDATA.md`): Add CN realtime marketdata adapter (repo first)
- Ticket doc: docs/tickets/TICKET_VS_0005_CN_REALTIME_MARKETDATA.md

## Do / Don't (scope guardrails)
- Do:
  - Add a realtime `marketdata` adapter + multi-source fusion for a small repo watchlist.
  - Enforce confidence gating: do not emit `tier=action` unless multi-source consensus passes thresholds.
  - Keep strict throttling/backoff/circuit-breaker to reduce ban risk.
  - Log enough fields into `paper_log` for later evaluation (providers/raw/consensus/reason).
- Don't:
  - Don't add auto execution / broker APIs.
  - Don't expand to full-market scanning.
  - Don't claim “risk-free arbitrage”.

## Decision points (0-3)
- Provider combo: A Eastmoney+Tencent / B Eastmoney+Sina / C Go+Python(AkShare) sidecar.
- Define `rate_pct` mapping (direct yield vs derived from price fields) and keep `raw` for audit.

## Expected minimal changes
- Files:
  - `internal/marketdata/*` (new)
  - `internal/engine/*` (wire provider into runtime)
  - `internal/signals/*` (repo realtime signal)
  - `internal/config/config.go` + `configs/config.example.yaml` (config)
  - docs updates if config changes

## Acceptance (3 steps, copy/paste)
1) `Test-Path configs\\config.example.yaml`
2) `go test ./...`
3) Optional smoke: fetch one snapshot for `204001.SH` and print within timeout (market time/network dependent).

## Rollback
- Remove `internal/marketdata`, revert signal/config wiring changes.

