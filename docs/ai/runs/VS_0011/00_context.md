# VS_0011 00_context

## Ticket
- Queue line: - Ticket VS_0011 (`docs/tickets/TICKET_VS_0011_OPTIMIZER_TO_QUOTAS.md`): Use optimizer results to drive per-signal quotas + runtime scheduling
- Ticket doc: docs/tickets/TICKET_VS_0011_OPTIMIZER_TO_QUOTAS.md

## Do / Don't (scope guardrails)
- Do:
  - Add a machine-readable optimizer output file (quota recommendations).
  - Add runtime loader to apply recommended quotas (override per-signal map).
  - Add unit tests proving deterministic recommendations + quota application.
- Don't:
  - Add new market data providers.
  - Add auto trading.

## Decision points (0-3)
- Quota suggestion algorithm:
  - Thompson-sampling per-slot allocation (deterministic given seed) to produce integer daily quotas.
- Runtime behavior:
  - Load `engine.reco_path` once at startup; if load fails, fall back to config quotas.

## Expected minimal changes
- Files:
  - `cmd/value-sniffer-radar-optimizer/main.go`
  - `internal/optimizer/reco.go`
  - `internal/reco/reco.go`
  - `internal/config/config.go`
  - `internal/engine/reco_loader.go`
  - `internal/engine/engine.go`

## Acceptance (3 steps, copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. Given a small `labels.repo.jsonl`, optimizer writes stable `.\\state\\optimizer.reco.json` (fixed seed).
3. Engine loads `engine.reco_path` and uses it to override per-signal daily action quotas.

## Rollback
- Remove `engine.reco_path` loading and revert optimizer reco output.

