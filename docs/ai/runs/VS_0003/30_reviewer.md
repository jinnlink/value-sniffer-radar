# VS_0003 30_reviewer

## Summary

The `cn_repo_sniper` signal is implemented with minimal scope and fits the existing engine/notifier model. It is configurable per-tier and has unit tests for normalization/window parsing.

## Issues

### Blocking

- None.

### Major

- Data freshness: `repo_daily` is “日行情” and may not reflect intraday rate spikes in real time. This ticket satisfies the plumbing; true intraday monitoring will require a later ticket with a realtime source (and explicit ToS/compliance review).
- Access cost: Tushare `repo_daily` requires points; users without access will see API errors. Mitigation: keep signal optional; document requirements in README/config.

### Minor / Suggestions

- Time window uses machine local time. If running on a non-CN timezone server, window behavior may surprise; consider adding a `window_tz_offset` config in a future ticket.
- `withinWindow()` currently “fails open” on malformed `window_start/window_end`. This avoids accidental silent disable, but could also cause extra alerts if misconfigured; current caps/dedupe reduce blast radius.

## Review checklist quick pass

- Scope: ✅ no auto-trading, no new data sources
- Config: ✅ new fields are additive; example config updated
- Tests: ✅ added for parsing/normalization

