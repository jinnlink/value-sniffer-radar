# VS_0002 00_context

## Ticket
- Queue line: - Ticket VS_0002 (`docs/tickets/TICKET_VS_0002_PAPER_EVAL.md`): Build paper evaluation tool (JSONL → metrics)
- Ticket doc: docs/tickets/TICKET_VS_0002_PAPER_EVAL.md

## Do / Don't (scope guardrails)
- Do:
  - Add a stdlib-only Python tool under `tools/` to summarize `paper_log` JSONL.
  - Keep outputs as stdout and optional Markdown file.
- Don't:
  - Add any third-party Python deps (`pip install`, `requirements.txt`).
  - Fetch prices / network data.
  - Modify Go engine code (`cmd/`, `internal/`).

## Decision points (0-3)
- none

## Expected minimal changes
- Files:
  - `tools/paper_eval.py`
  - `docs/ai/runs/VS_0002/*` (run artifacts)

## Acceptance (3 steps, copy/paste)
1. `Test-Path tools\\paper_eval.py`
2. `python tools\\paper_eval.py --help`
3. `@'{\"ts\":\"2026-01-29T00:00:00+08:00\",\"event\":{\"source\":\"cb_double_low_action\",\"trade_date\":\"20260128\",\"market\":\"CN-A\",\"symbol\":\"110000.SH\",\"title\":\"demo\",\"body\":\"\",\"tags\":{\"tier\":\"action\"}}}'@ | Set-Content -Encoding UTF8 .\\state\\paper.sample.jsonl; python tools\\paper_eval.py --in .\\state\\paper.sample.jsonl`

## Rollback
- Delete `tools/paper_eval.py`.

