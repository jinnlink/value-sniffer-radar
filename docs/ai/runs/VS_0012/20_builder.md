# VS_0012 20_builder

## What changed
- Added `tools/daily_loop.ps1` to run:
  - `value-sniffer-radar-labeler` → `labels.repo.jsonl`
  - `value-sniffer-radar-optimizer` → `optimizer.report.md` + `optimizer.reco.json`
- Updated `README.md` with the daily loop command and `engine.reco_path` wiring.

## Files touched
- `tools/daily_loop.ps1`
- `README.md`

## Acceptance
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. `powershell -NoProfile -ExecutionPolicy Bypass -File .\\tools\\daily_loop.ps1 -Config .\\config.yaml`
3. Confirm outputs exist: `.\\state\\labels.repo.jsonl`, `.\\state\\optimizer.report.md`, `.\\state\\optimizer.reco.json`

