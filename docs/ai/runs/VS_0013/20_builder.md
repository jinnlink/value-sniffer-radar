# VS_0013 20_builder

## What changed
- Added LLM sidecar CLI: `cmd/value-sniffer-radar-llm`
  - `-mode enrich|digest`
  - `-provider api|cli`
- Added `internal/llm` package:
  - prompt builder + strict JSON parser
  - `APIClient` (OpenAI-compatible chat completions)
  - `CLIClient` (stdin→stdout)
- README updated with usage examples.

## Files touched
- `cmd/value-sniffer-radar-llm/main.go`
- `internal/llm/*`
- `README.md`
- `docs/tickets/TICKET_VS_0013_LLM_ENRICHER.md`

## Acceptance
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. CLI provider sample produces `.\\state\\llm.enriched.jsonl`
3. README explains API + CLI execution.

