# VS_0013 00_context

## Ticket
- Queue line: - Ticket VS_0013 (`docs/tickets/TICKET_VS_0013_LLM_ENRICHER.md`): LLM sidecar (API+CLI) to enrich/digest events
- Ticket doc: docs/tickets/TICKET_VS_0013_LLM_ENRICHER.md

## Do / Don't (scope guardrails)
- Do:
  - Add LLM sidecar CLI to enrich/digest events (not in hot-path).
  - Support provider `api` (OpenAI-compatible chat) and `cli` (exec).
  - Keep LLM output strictly JSON and parse it.
- Don't:
  - Let LLM alter trading decisions or tiers automatically.
  - Add auto trading.

## Decision points (0-3)
- Default provider interface:
  - `api` uses chat-completions style endpoint.
  - `cli` reads prompt from stdin and must print JSON to stdout.

## Expected minimal changes
- Files:
  - `cmd/value-sniffer-radar-llm/main.go`
  - `internal/llm/*`
  - `README.md` (how to run)

## Acceptance (3 steps, copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. `go run .\\cmd\\value-sniffer-radar-llm -mode enrich -provider cli ...` works on a sample paper log.
3. Docs show how to run with cloud API + CLI.

## Rollback
- Remove `cmd/value-sniffer-radar-llm` and `internal/llm`.

