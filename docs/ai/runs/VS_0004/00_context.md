# VS_0004 00_context

## Ticket
- Queue line: - Ticket VS_0004 (`docs/tickets/TICKET_VS_0004_PAPER_EVAL_BOM.md`): Fix paper_eval JSONL BOM compatibility (PowerShell UTF8)
- Ticket doc: docs/tickets/TICKET_VS_0004_PAPER_EVAL_BOM.md

## Do / Don't (scope guardrails)
- Do:
  - Make `tools/paper_eval.py` tolerate UTF-8 BOM in JSONL (PowerShell `Set-Content -Encoding UTF8`).
  - Keep tool stdlib-only; minimal code change.
  - Update VS_0004 run artifacts (spec/build/review/run/conclusion).
- Don't:
  - Don't add new metrics/features to paper_eval.
  - Don't change other repo components/signals.
  - Don't introduce new dependencies.

## Decision points (0-3)
- Decide BOM handling approach: `utf-8-sig` and/or strip leading `\ufeff` per line.

## Expected minimal changes
- Files:
  - `tools/paper_eval.py` (BOM tolerant)
  - `docs/ai/runs/VS_0004/*` (run artifacts)

## Acceptance (3 steps, copy/paste)
1) `python tools\\paper_eval.py --help`
2) `@'{\"ts\":\"2026-01-29T00:00:00+08:00\",\"event\":{\"source\":\"demo\",\"trade_date\":\"20260129\",\"market\":\"CN-A\",\"symbol\":\"110000.SH\",\"title\":\"demo\",\"body\":\"\",\"tags\":{\"tier\":\"action\"}}}'@ | Set-Content -Encoding UTF8 .\\state\\paper.sample.jsonl`
3) `python tools\\paper_eval.py --in .\\state\\paper.sample.jsonl`

## Rollback
- Revert `tools/paper_eval.py` changes.

