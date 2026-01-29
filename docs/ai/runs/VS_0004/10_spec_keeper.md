# VS_0004 10_spec_keeper

## Pass/Fail gate (this ticket)

Pass if:
- `tools/paper_eval.py` parses JSONL created via PowerShell `Set-Content -Encoding UTF8` without JSON errors.
- Tool remains stdlib-only, and the change is minimal.

Fail if:
- Adds new features/metrics unrelated to BOM tolerance.
- Introduces new dependencies.

## Acceptance reminders

1) `python tools\\paper_eval.py --help`
2) `@'{\"ts\":\"2026-01-29T00:00:00+08:00\",\"event\":{\"source\":\"demo\",\"trade_date\":\"20260129\",\"market\":\"CN-A\",\"symbol\":\"110000.SH\",\"title\":\"demo\",\"body\":\"\",\"tags\":{\"tier\":\"action\"}}}'@ | Set-Content -Encoding UTF8 .\\state\\paper.sample.jsonl`
3) `python tools\\paper_eval.py --in .\\state\\paper.sample.jsonl`

