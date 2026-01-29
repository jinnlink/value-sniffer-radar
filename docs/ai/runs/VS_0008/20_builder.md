# VS_0008 20_builder

## What changed

- Added `value-sniffer-radar-labeler` CLI to label repo events with `reward` for the optimizer:
  - input: `paper_log` JSONL
  - output: `labels.repo.jsonl` (append-only)
  - default windows: 10s, 30s, 5m
  - reward logic (repo-only):
    - `reward=1` if at T+window the **fused** repo rate is still `>= threshold` and `confidence=PASS`
    - otherwise `reward=0` with a reason
- Added offline deterministic mode:
  - `-mock-rate` and `-mock-confidence` so tests/acceptance do not need network.

## Files added

- `cmd/value-sniffer-radar-labeler/main.go`
- `internal/labeler/*`

