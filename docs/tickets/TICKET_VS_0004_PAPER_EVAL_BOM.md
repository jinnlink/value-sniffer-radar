# Ticket VS_0004 — Paper Eval JSONL BOM Compatibility

## Goal

Fix `tools/paper_eval.py` so it can parse JSONL files created by Windows PowerShell:
- `Set-Content -Encoding UTF8` (often writes UTF-8 BOM)

This avoids “invalid json line=1” when the first line begins with BOM.

## Scope

In scope:
- Make JSONL reader tolerant of UTF-8 BOM on the first line.
- Keep the tool stdlib-only.
- Update ticket run artifacts (spec/build/review/run) and docs if needed.

Out of scope:
- Any new evaluation metrics
- Any trading logic changes

## Acceptance (3 steps)

1) `python tools\\paper_eval.py --help`
2) Create a sample JSONL with PowerShell `Set-Content -Encoding UTF8` then run:
   - `@'{\"ts\":\"2026-01-29T00:00:00+08:00\",\"event\":{\"source\":\"demo\",\"trade_date\":\"20260129\",\"market\":\"CN-A\",\"symbol\":\"110000.SH\",\"title\":\"demo\",\"body\":\"\",\"tags\":{\"tier\":\"action\"}}}'@ | Set-Content -Encoding UTF8 .\\state\\paper.sample.jsonl`
   - `python tools\\paper_eval.py --in .\\state\\paper.sample.jsonl`
3) Output includes `rows=` and does NOT print `invalid json` errors.

## Rollback

- Revert changes to `tools/paper_eval.py`.

