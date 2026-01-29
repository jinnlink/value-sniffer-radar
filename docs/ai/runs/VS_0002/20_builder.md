# Builder Output

## Impact

- Adds a stdlib-only Python evaluator to summarize `paper_log` JSONL.
- No changes to Go engine runtime, signals, or notifiers.

## Files Changed

- `tools/paper_eval.py`
- `IMPL.md` (documented `tools/` Python allowance)
- `docs/ai/runs/VS_0002/00_context.md` (filled placeholders)

## Acceptance (3 steps)

1. `Test-Path tools\\paper_eval.py`
2. `python tools\\paper_eval.py --help`
3. `@'{\"ts\":\"2026-01-29T00:00:00+08:00\",\"event\":{\"source\":\"cb_double_low_action\",\"trade_date\":\"20260128\",\"market\":\"CN-A\",\"symbol\":\"110000.SH\",\"title\":\"demo\",\"body\":\"\",\"tags\":{\"tier\":\"action\"}}}'@ | Set-Content -Encoding UTF8 .\\state\\paper.sample.jsonl; python tools\\paper_eval.py --in .\\state\\paper.sample.jsonl`

## Rollback

- Delete `tools/paper_eval.py` and revert `IMPL.md` if needed.

