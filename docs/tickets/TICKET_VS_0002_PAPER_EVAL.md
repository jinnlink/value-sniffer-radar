# Ticket VS_0002 — Paper Evaluation Tool (JSONL → Metrics)

## Goal

Turn `paper_log` output (JSONL) into a **repeatable evaluation report** so we can verify which signals are actually profitable **after costs** and stay within the 5% drawdown target.

## Scope

In scope:
- Add a small **stdlib-only Python** tool to parse `paper_log` JSONL and produce summary metrics.
- Output formats:
  - Console summary (stdout)
  - Optional Markdown report file (daily)
- Metrics (minimum viable):
  - count by `tier` and by `signal`
  - top symbols by frequency
  - basic “hit-rate proxy” placeholders (no price fetch yet; just structure)

Out of scope:
- Pulling market prices to compute real PnL (that will be a separate ticket once data source is decided)
- Auto-trading / broker integration

## Acceptance (3 steps)

1) `Test-Path tools\\paper_eval.py`
2) `python tools\\paper_eval.py --help`
3) Create a tiny sample file then run:
   - `@'{\"ts\":\"2026-01-29T00:00:00+08:00\",\"event\":{\"source\":\"cb_double_low_action\",\"trade_date\":\"20260128\",\"market\":\"CN-A\",\"symbol\":\"110000.SH\",\"title\":\"demo\",\"body\":\"\",\"tags\":{\"tier\":\"action\"}}}'@ | Set-Content -Encoding UTF8 .\\state\\paper.sample.jsonl`
   - `python tools\\paper_eval.py --in .\\state\\paper.sample.jsonl`

## Rollback

- Delete `tools/paper_eval.py` and any docs it adds/updates.

