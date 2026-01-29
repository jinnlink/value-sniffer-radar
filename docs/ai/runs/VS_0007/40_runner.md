# VS_0007 40_runner

## Commands + Output

### Unit tests

- `go test ./...`
- PASS

### Realtime-only startup without `TUSHARE_TOKEN`

- Preconditions:
  - `Remove-Item Env:TUSHARE_TOKEN -ErrorAction SilentlyContinue`
  - `engine.trade_date_mode: fixed`
  - only realtime signals enabled (e.g. `cn_repo_realtime`)
- Result:
  - Process starts and runs without token errors.

