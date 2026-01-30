# VS_0011 10_spec_keeper

## Contract Check
- Ticket: `docs/tickets/TICKET_VS_0011_OPTIMIZER_TO_QUOTAS.md`
- Goal: optimizer emits reco JSON; engine loads reco and overrides per-signal daily quotas.
- Out of scope: auto trading, new data providers.

## Guardrails
- Deterministic recommendations given fixed seed.
- Runtime should be resilient: if reco file missing/invalid, fall back to config quotas.
- Keep schema versioned (`reco.v1`).

## Acceptance (copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. Run optimizer with `-out-reco` produces stable `.\\state\\optimizer.reco.json` (fixed `-seed`).
3. Engine with `engine.reco_path` loads reco and applies quotas (unit test proves override).

## Rollback
- Stop loading `engine.reco_path` and remove optimizer reco output.

