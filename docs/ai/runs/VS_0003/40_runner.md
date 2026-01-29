# VS_0003 40_runner

## Environment

- Go was not installed on PATH.
- Installed a local Go toolchain (not committed) under `state/_toolchains/go1.25.6/` for this validation.

## Commands + Output

### Go version

- Command:
  - `state\_toolchains\go1.25.6\go\bin\go.exe version`
- Output:
  - `go version go1.25.6 windows/amd64`

### Dependency lockfile

- Command:
  - `state\_toolchains\go1.25.6\go\bin\go.exe mod tidy`
- Result:
  - Generated `go.sum` (new tracked file).

### Unit tests

- Command:
  - `state\_toolchains\go1.25.6\go\bin\go.exe test ./...`
- Output (summary):
  - `ok  	value-sniffer-radar/internal/signals	...`
  - others: `[no test files]`

### Live smoke (skipped)

- Reason:
  - `TUSHARE_TOKEN` is missing in the current environment.
- Next:
  - Set `TUSHARE_TOKEN`, then run the Acceptance #3 smoke command in `docs/ai/runs/VS_0003/00_context.md`.

