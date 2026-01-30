# VS_0012 10_spec_keeper

## Contract Check
- Ticket: `docs/tickets/TICKET_VS_0012_DAILY_LOOP_SCRIPT.md`
- Goal: one-command daily loop that produces labels + optimizer report + reco json.
- Out of scope: service install/restart automation, auto trading.

## Acceptance (copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. `powershell -NoProfile -ExecutionPolicy Bypass -File .\\tools\\daily_loop.ps1 -Config .\\config.yaml`
3. Verify README mentions `tools/daily_loop.ps1` and `engine.reco_path`.

