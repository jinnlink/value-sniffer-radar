# VS_0007 30_reviewer

## Summary

Good, minimal usability fix:
- Makes realtime-only runs possible without a Tushare token.
- Fixes a common Eastmoney repo scaling pitfall in the example config.

## Issues

### Blocking

- None.

### Major

- Ensure `RequiresTushare()` stays accurate as new signals are added. If a future signal uses Tushare, add it to the switch in `internal/config/config.go`.

### Minor

- `engine.New()` now allows `client=nil`. Any signal that needs Tushare should fail fast (currently enforced by config validation + the signal’s runtime behavior).

