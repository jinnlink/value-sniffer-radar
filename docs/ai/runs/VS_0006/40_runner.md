# VS_0006 40_runner

## Commands + Output

### Unit tests

- `go test ./...`
- PASS

### Sample optimizer run

1) Create sample JSONL:
   - Use the snippet in `docs/ai/runs/VS_0006/00_context.md`
2) Run (most robust on Windows is build+run):
   - `go build -o .\\state\\optimizer.exe .\\cmd\\value-sniffer-radar-optimizer`
   - `.\\state\\optimizer.exe -in .\\state\\optimizer.sample.jsonl -out-md .\\state\\optimizer.report.md -seed 7 -slots 2`
3) Verify:
   - `Test-Path .\\state\\optimizer.report.md`

