# VS_0013 10_spec_keeper

## Contract Check
- Ticket: `docs/tickets/TICKET_VS_0013_LLM_ENRICHER.md`
- Goal: LLM sidecar CLI enriches/digests events (API + CLI providers).
- Non-goal: LLM decides trades or changes tiers.

## Guardrails
- Strict JSON output parsing for enrich mode; if parse fails, skip and do not crash.
- No secrets in prompts or repo.

## Acceptance (copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. CLI provider: `go run .\\cmd\\value-sniffer-radar-llm -mode enrich -provider cli -cli-cmd cmd.exe -cli-args \"/c echo {\\\"summary\\\":\\\"ok\\\",\\\"risks\\\":[],\\\"checklist\\\":[]}\" -in .\\state\\paper.jsonl`
3. README documents API + CLI usage.

