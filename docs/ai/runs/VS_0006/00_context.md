# VS_0006 00_context

## Ticket
- Queue line: - Ticket VS_0006 (`docs/tickets/TICKET_VS_0006_OPTIMIZER.md`): Add online optimizer (bandit + dynamic polling + auto-threshold suggestions)
- Ticket doc: docs/tickets/TICKET_VS_0006_OPTIMIZER.md

## Do / Don't (scope guardrails)
- Do:
  - Add an **offline optimizer tool** that reads `paper_log` JSONL and produces a Markdown “suggested config changes” report.
  - Implement a minimal bandit core (Thompson Sampling / Beta-Bernoulli) with deterministic dry-run support (seed).
  - Implement a simple “dynamic polling plan” generator (priority scoring) that can be validated without network.
- Don't:
  - Don't add auto-trading or broker integration.
  - Don't auto-edit `config.yaml` in place; only output suggestions.
  - Don't introduce heavy ML dependencies.

## Decision points (0-3)
- Reward definition: start with synthetic labels / placeholder `reward` field; real PnL requires a later data-source ticket.

## Expected minimal changes
- Files:
  - `cmd/value-sniffer-radar-optimizer/main.go` (new CLI)
  - `internal/optimizer/*` (bandit + scheduler + JSONL parsing)
  - `docs/tickets/TICKET_VS_0006_OPTIMIZER.md` (if acceptance needs tightening)

## Acceptance (3 steps, copy/paste)
1) `go test ./...`
2) Create a synthetic JSONL then run optimizer:
   - `@'
{"ts":"2026-01-29T00:00:00+08:00","event":{"source":"a","trade_date":"20260129","market":"CN-A","symbol":"204001.SH","title":"demo","body":"","tags":{"tier":"action"},"data":{"reward":1}}}
{"ts":"2026-01-29T00:00:01+08:00","event":{"source":"b","trade_date":"20260129","market":"CN-A","symbol":"204001.SH","title":"demo","body":"","tags":{"tier":"action"},"data":{"reward":0}}}
{"ts":"2026-01-29T00:00:02+08:00","event":{"source":"a","trade_date":"20260129","market":"CN-A","symbol":"204001.SH","title":"demo","body":"","tags":{"tier":"action"},"data":{"reward":1}}}
'@ | Set-Content -Encoding UTF8 .\\state\\optimizer.sample.jsonl`
   - `go run .\\cmd\\value-sniffer-radar-optimizer -- -in .\\state\\optimizer.sample.jsonl -out-md .\\state\\optimizer.report.md -seed 7 -slots 2`
3) `Test-Path .\\state\\optimizer.report.md`

## Rollback
- Delete `cmd/value-sniffer-radar-optimizer` and `internal/optimizer` and revert docs changes.

