# VS_0009 00_context

## Ticket
- Queue line: - Ticket VS_0009 (`docs/tickets/TICKET_VS_0009_OPTIMIZER_CONSUME_LABELS.md`): Optimizer consumes labels.repo.jsonl
- Ticket doc: docs/tickets/TICKET_VS_0009_OPTIMIZER_CONSUME_LABELS.md

## Do / Don't (scope guardrails)
- Do:
  - Add optional `-labels` flag to optimizer CLI.
  - Prefer rewards from `labels.repo.jsonl` over `event.data.reward`.
  - Add report sections: labeled coverage + per-signal reward rate by window.
  - Keep changes minimal and testable (`go test ./...`).
- Don't:
  - Add new market data sources.
  - Add auto execution / auto trading.
  - Change unrelated signal logic.

## Decision points (0-3)
- Default label window selection when multiple windows exist:
  - Pick the window with the most labels (highest coverage), unless user passes `-label-window-sec`.

## Expected minimal changes
- Files:
  - `cmd/value-sniffer-radar-optimizer/main.go`
  - `internal/optimizer/*` (labels parsing + report fields)
  - `internal/labeler/labeler.go` (reuse shared EventID helper)

## Acceptance (3 steps, copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. With a small sample `labels.repo.jsonl`, optimizer report changes ranking vs paper-only.
3. Report includes a labeled coverage summary.

## Rollback
- Revert optimizer label ingestion changes (`cmd/value-sniffer-radar-optimizer` + `internal/optimizer/labels.go`).

