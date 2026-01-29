# VS_0005 20_builder

## What changed

- Added a realtime marketdata layer with multi-source fusion (consensus + confidence gating + circuit breaker).
- Implemented two cheap-first repo providers:
  - Eastmoney (HTTP JSON; `push2.eastmoney.com` quote API)
  - Tencent (plain text; `qt.gtimg.cn` quote API)
- Added `cn_repo_realtime` signal:
  - Uses fused consensus rate
  - Emits `tier=action` only when `confidence=PASS` and `consensus>=min_yield_pct`
  - Optional `confirm_k` consecutive confirmations
- Wired marketdata into the engine and updated existing signals to accept the new dependency (ignored by non-realtime signals).

## Files changed

- `internal/marketdata/*` (new)
- `internal/engine/engine.go`
- `internal/signals/signal.go`
- `internal/signals/cn_repo_realtime.go` (new)
- `internal/signals/*` (signature update)
- `internal/config/config.go` (marketdata config + `confirm_k`)
- `configs/config.example.yaml` (marketdata + cn_repo_realtime examples)

## Acceptance (copy/paste)

1) `go test ./...`
2) Enable realtime repo example in `config.yaml`:
   - `marketdata.enabled: true`
   - enable `signals` entries `cn_repo_realtime_*`
   - set `engine.interval_seconds: 3`
3) Run:
   - `go run .\\cmd\\value-sniffer-radar -config .\\config.yaml`

## Notes

- This ticket does not promise intraday “true arbitrage”; it provides a robust near-realtime snapshot + confidence gating, and logs enough context for paper evaluation.

