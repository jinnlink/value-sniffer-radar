# Ticket VS_0003 — CN Reverse Repo Monitor (GC001 / R-001)

## Goal

Add a monitor for reverse repo opportunities (cash management baseline), with observe+action tiers.

## Scope (draft)

In scope:
- New signal type `cn_repo_sniper` for CN reverse repos (e.g. `204001.SH` / `131810.SZ`, also accepts `SH204001` / `SZ131810`) based on:
  - weighted rate threshold (`min_yield_pct`, %)
  - optional active window (`window_start` / `window_end`)
- Alerts via existing notifiers (stdout / aival_queue / paper_log).

Out of scope:
- Auto execution.
- New realtime data sources (AkShare / websockets). This ticket is plumbing on Tushare `repo_daily` (日行情).

## Acceptance (draft)

1) Config accepts a new signal entry type: `signals[].type: "cn_repo_sniper"`
2) `go test ./...` passes (includes unit tests for parsing/normalization)
3) Running with a fixed trade date produces events when thresholds are forced low (requires:
   - env `TUSHARE_TOKEN`
   - Tushare access to `repo_daily`, which requires points)
