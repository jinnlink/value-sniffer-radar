# VS_0009 20_builder

## What changed
- Optimizer CLI now supports `-labels` + `-label-window-sec`, and prefers label rewards over `event.data.reward`.
- Optimizer report now includes:
  - labeled coverage by window
  - reward rate by signal/window
- Shared `optimizer.EventID()` introduced and reused by labeler to keep join logic identical.

## Files touched
- `cmd/value-sniffer-radar-optimizer/main.go`
- `internal/optimizer/labels.go`
- `internal/optimizer/paper.go`
- `internal/optimizer/render.go`
- `internal/optimizer/labels_test.go`
- `internal/labeler/labeler.go`

## Acceptance
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. Run optimizer once without labels, then with labels; verify ranking flips in report.
3. Confirm coverage table exists when `-labels` is set.

## Rollback
- Revert the files above.

