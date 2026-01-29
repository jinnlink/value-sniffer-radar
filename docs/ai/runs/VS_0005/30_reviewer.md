# VS_0005 30_reviewer

## Summary

The ticket implements a minimal, extensible realtime marketdata layer with multi-source fusion and a new `cn_repo_realtime` signal. It enforces a strict rule: **no `tier=action` unless multi-source consensus passes**.

## Issues

### Blocking

- None.

### Major

- Provider ToS / stability risk: Eastmoney/Tencent endpoints are unofficial and may change or rate-limit. Mitigation: conservative polling + circuit breaker + keep raw fields logged for debugging; provider swapping is isolated in `internal/marketdata`.
- Timezone assumptions: signal window uses local machine time. Running outside CN timezone may misfire; consider future `window_tz_offset`.

### Minor

- `cn_repo_realtime` currently errors when `marketdata` is disabled; this will surface in logs. Consider optionally downgrading to an observe event in future if log noise is an issue.
- `anyToFloat` is shared helper in one provider file; OK, but could be moved to a small util file if more providers are added.

## Checklist

- Scope creep: ✅ no execution, no broad scanning
- Tests: ✅ fusion + provider parsing tests added
- Config: ✅ example includes `marketdata` section and repo realtime signal stubs

