# VS_0006 20_builder

## What changed

- Added an **optimizer CLI** to support “多尝试、可验证、可迭代”的策略改进流程：
  - Reads `paper_log` JSONL
  - Uses a Beta-Bernoulli bandit with Thompson Sampling to suggest which signals deserve `tier=action` “名额”
  - Produces a Markdown report (and optional file output)
- Added a deterministic polling-plan helper (priority → next interval) as the base for future dynamic polling integration.

## Files added/changed

- `cmd/value-sniffer-radar-optimizer/main.go`
- `internal/optimizer/*`

## Notes / limitations

- This ticket does **not** compute real PnL. For now it consumes a `reward` field inside `event.data.reward` (0/1).
- A later ticket should compute `reward` from price windows after costs and write it back into paper logs or a sidecar labels file.

