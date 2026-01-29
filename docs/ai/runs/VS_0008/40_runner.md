# VS_0008 40_runner

## Commands + Output

### Unit tests

- `go test ./...`
- PASS

### Mock labeler run (no network)

1) Create sample:
   - `$s = '{\"ts\":\"2026-01-29T00:00:00+08:00\",\"event\":{\"source\":\"cn_repo_realtime_action\",\"trade_date\":\"20260129\",\"market\":\"CN-A\",\"symbol\":\"204001.SH\",\"title\":\"demo\",\"body\":\"\",\"tags\":{\"tier\":\"action\",\"kind\":\"repo\"},\"data\":{\"consensus_rate_pct\":5.0}}}'`
   - `$s | Set-Content -Encoding UTF8 .\\state\\paper.sample.repo.jsonl`
2) Run:
   - `go run .\\cmd\\value-sniffer-radar-labeler -- -config .\\config.yaml -in .\\state\\paper.sample.repo.jsonl -out .\\state\\labels.repo.jsonl -windows 10s -grace 24h -mock-rate 1.6 -mock-confidence PASS`
3) Verify:
   - `Test-Path .\\state\\labels.repo.jsonl`

