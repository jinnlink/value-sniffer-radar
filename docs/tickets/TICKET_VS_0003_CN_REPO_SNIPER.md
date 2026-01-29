# Ticket VS_0003 — CN Reverse Repo Monitor (GC001 / R-001)

## Goal

Add a monitor for reverse repo opportunities (cash management baseline), with observe+action tiers.

## Scope (draft)

In scope:
- New signal type for CN reverse repos (e.g. `SH204001`, `SZ131810`) based on yield thresholds and time windows.
- Alerts via existing notifiers (stdout / aival_queue / paper_log).

Out of scope:
- Auto execution.

## Acceptance (draft)

1) Config accepts a new signal entry type (e.g. `cn_repo_sniper`)
2) Running with a fixed trade date produces events when thresholds are forced low
3) Works with `tier=observe` and `tier=action` configs

