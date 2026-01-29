# VS_0009 10_spec_keeper

## Contract Check
- Ticket: `docs/tickets/TICKET_VS_0009_OPTIMIZER_CONSUME_LABELS.md`
- Goal: Optimizer uses `labels.repo.jsonl` produced by labeler.
- In scope: `-labels` flag, prefer label rewards, coverage + per-signal/window reward rate report.
- Out of scope: new price sources, auto execution.

## Guardrails
- Keep optimizer usable without labels.
- If `-labels` path is missing, do not crash; warn and proceed with paper-only behavior.
- Avoid import cycles between `internal/labeler` and `internal/optimizer`.

## Acceptance (copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. Provide sample `labels.repo.jsonl` that flips optimizer ranking vs paper-only.
3. Report contains labeled coverage summary (% labeled per window).

## Rollback
- Revert changes in `cmd/value-sniffer-radar-optimizer` + `internal/optimizer` label ingestion.

