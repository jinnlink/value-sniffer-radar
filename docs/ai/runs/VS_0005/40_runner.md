# VS_0005 40_runner

## Commands + Output

### Unit tests

- Command:
  - `go test ./...`
- Result:
  - PASS (no failures)

### Optional manual smoke

To try realtime repo in live mode (network + market time dependent):

1) Copy example:
   - `Copy-Item .\\configs\\config.example.yaml .\\config.yaml`
2) Edit:
   - `marketdata.enabled: true`
   - enable `signals` entries `cn_repo_realtime_action` / `cn_repo_realtime_observe`
   - set `engine.interval_seconds: 3`
3) Run:
   - `go run .\\cmd\\value-sniffer-radar -config .\\config.yaml`

