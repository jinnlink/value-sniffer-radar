# VS_0011 20_builder

## What changed
- Optimizer now supports `-out-reco` to write a machine-readable quota recommendation file (JSON).
- Added shared reco schema package `internal/reco` for engine + optimizer.
- Engine can load reco once at startup via `engine.reco_path` and override `engine.action_max_events_per_signal_per_day`.

## Files touched
- `cmd/value-sniffer-radar-optimizer/main.go`
- `internal/optimizer/reco.go`
- `internal/reco/reco.go`
- `internal/config/config.go`
- `configs/config.example.yaml`
- `internal/engine/reco_loader.go`
- `internal/engine/engine.go`
- Tests: `internal/optimizer/reco_test.go`, `internal/reco/reco_test.go`, `internal/engine/reco_loader_test.go`

## Acceptance
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. `go run .\\cmd\\value-sniffer-radar-optimizer -in .\\state\\paper.jsonl -labels .\\state\\labels.repo.jsonl -seed 7 -slots 30 -out-reco .\\state\\optimizer.reco.json`
3. Set `engine.reco_path: .\\state\\optimizer.reco.json` and run engine; observe log line `reco loaded ... quotas=...` (or rely on unit test).

## Rollback
- Revert the files above and remove `engine.reco_path`.

