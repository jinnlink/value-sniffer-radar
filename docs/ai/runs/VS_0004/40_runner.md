# VS_0004 40_runner

## Commands + Output

### Help

- `python tools\\paper_eval.py --help`

### BOM smoke test (PowerShell UTF8)

- Create sample (PowerShell `Set-Content -Encoding UTF8`):
  - `$s = '{\"ts\":\"2026-01-29T00:00:00+08:00\",\"event\":{\"source\":\"demo\",\"trade_date\":\"20260129\",\"market\":\"CN-A\",\"symbol\":\"110000.SH\",\"title\":\"demo\",\"body\":\"\",\"tags\":{\"tier\":\"action\"}}}'`
  - `$s | Set-Content -Encoding UTF8 .\\state\\paper.sample.jsonl`
- Run:
  - `python tools\\paper_eval.py --in .\\state\\paper.sample.jsonl`
- Result:
  - `total_events: 1` and no `invalid json line=1` warnings.

