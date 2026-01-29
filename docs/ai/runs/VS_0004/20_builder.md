# VS_0004 20_builder

## What changed

- Made `tools/paper_eval.py` tolerant of UTF-8 BOM in JSONL input.

## Files changed

- `tools/paper_eval.py`

## Implementation notes

- When reading from a file path, use `encoding="utf-8-sig"` so Python transparently strips UTF-8 BOM.
- As an extra guard, strip `\ufeff` from each line before `json.loads()` (handles cases where BOM appears at line start).

## Acceptance (copy/paste)

1) `python tools\\paper_eval.py --help`
2) `@'{\"ts\":\"2026-01-29T00:00:00+08:00\",\"event\":{\"source\":\"demo\",\"trade_date\":\"20260129\",\"market\":\"CN-A\",\"symbol\":\"110000.SH\",\"title\":\"demo\",\"body\":\"\",\"tags\":{\"tier\":\"action\"}}}'@ | Set-Content -Encoding UTF8 .\\state\\paper.sample.jsonl`
3) `python tools\\paper_eval.py --in .\\state\\paper.sample.jsonl`

## Rollback

- Revert `tools/paper_eval.py`.

