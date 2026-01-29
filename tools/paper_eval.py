#!/usr/bin/env python3
"""
VS_0002: Paper Evaluation Tool (stdlib-only)

Input:
- JSONL produced by the `paper_log` notifier:
  {"ts":"...","event":{...}}

Output:
- Human-readable summary to stdout
- Optional Markdown report
"""

from __future__ import annotations

import argparse
import json
import sys
from collections import Counter
from datetime import datetime
from pathlib import Path
from typing import Any, Iterable, TextIO


def iter_jsonl(fp: TextIO) -> Iterable[dict[str, Any]]:
    for line_no, raw in enumerate(fp, 1):
        s = raw.strip()
        if not s:
            continue
        try:
            obj = json.loads(s)
        except json.JSONDecodeError:
            print(f"[warn] invalid json line={line_no}", file=sys.stderr)
            continue
        if isinstance(obj, dict):
            yield obj
        else:
            print(f"[warn] non-object json line={line_no}", file=sys.stderr)


def coerce_event(row: dict[str, Any]) -> dict[str, Any] | None:
    # paper_log format: {"ts": "...", "event": {...}}
    ev = row.get("event")
    if isinstance(ev, dict):
        return ev
    # tolerate raw events (future-proof)
    if "source" in row and "title" in row and "trade_date" in row:
        return row
    return None


def render_markdown(
    *,
    generated_at: str,
    input_path: str,
    total: int,
    tier_counts: Counter[str],
    source_counts: Counter[str],
    symbol_counts: Counter[str],
    trade_date_counts: Counter[str],
) -> str:
    def table(counter: Counter[str], top: int = 20) -> str:
        rows = counter.most_common(top)
        if not rows:
            return "_(none)_\n"
        lines = ["| key | count |", "|---|---:|"]
        for k, v in rows:
            lines.append(f"| {k} | {v} |")
        return "\n".join(lines) + "\n"

    parts = [
        "# Paper Eval Report",
        "",
        f"- generated_at: `{generated_at}`",
        f"- input: `{input_path}`",
        f"- total_events: `{total}`",
        "",
        "## By Tier",
        table(tier_counts, top=20),
        "## By Signal (source)",
        table(source_counts, top=30),
        "## Top Symbols",
        table(symbol_counts, top=30),
        "## By Trade Date",
        table(trade_date_counts, top=30),
        "",
        "## Notes",
        "- This ticket does **not** fetch prices; PnL/returns require a separate data-source ticket.",
        "",
    ]
    return "\n".join(parts)


def main() -> int:
    ap = argparse.ArgumentParser(description="Value Sniffer Radar - paper_log evaluator (stdlib-only)")
    ap.add_argument("--in", dest="in_path", required=True, help="Input JSONL path (paper_log). Use '-' for stdin.")
    ap.add_argument("--out-md", dest="out_md", default="", help="Optional Markdown report output path.")
    args = ap.parse_args()

    in_path = args.in_path.strip()
    if in_path == "-":
        fp = sys.stdin
        input_label = "<stdin>"
        rows = list(iter_jsonl(fp))
    else:
        p = Path(in_path)
        if not p.exists():
            print(f"[error] not found: {p}", file=sys.stderr)
            return 2
        input_label = str(p)
        with p.open("r", encoding="utf-8") as fp:
            rows = list(iter_jsonl(fp))

    tier_counts: Counter[str] = Counter()
    source_counts: Counter[str] = Counter()
    symbol_counts: Counter[str] = Counter()
    trade_date_counts: Counter[str] = Counter()

    total = 0
    for row in rows:
        ev = coerce_event(row)
        if not ev:
            continue
        total += 1

        src = str(ev.get("source", "") or "").strip() or "unknown"
        source_counts[src] += 1

        sym = str(ev.get("symbol", "") or "").strip() or "unknown"
        symbol_counts[sym] += 1

        td = str(ev.get("trade_date", "") or "").strip() or "unknown"
        trade_date_counts[td] += 1

        tags = ev.get("tags")
        tier = "action"
        if isinstance(tags, dict):
            t = str(tags.get("tier", "") or "").strip().lower()
            if t in {"action", "observe"}:
                tier = t
        tier_counts[tier] += 1

    generated_at = datetime.now().astimezone().isoformat(timespec="seconds")
    md = render_markdown(
        generated_at=generated_at,
        input_path=input_label,
        total=total,
        tier_counts=tier_counts,
        source_counts=source_counts,
        symbol_counts=symbol_counts,
        trade_date_counts=trade_date_counts,
    )

    # stdout summary
    print(md)

    out_md = args.out_md.strip()
    if out_md:
        out_path = Path(out_md)
        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(md, encoding="utf-8")
        print(f"[ok] wrote: {out_path}", file=sys.stderr)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())

