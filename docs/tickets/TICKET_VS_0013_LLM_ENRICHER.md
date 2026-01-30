# Ticket VS_0013 — LLM Enricher Sidecar (API + CLI)

## Goal

Use an LLM to make alerts “smarter for humans” without putting LLM in the trading/decision hot-path:
- Enrich events with short explanations + risks + checklist
- Produce a digest (reduce spam) for QQ/email
- Support both:
  - Cloud API (OpenAI-compatible chat completions)
  - CLI execution (call `codex`/`gemini`/etc. via stdin/stdout)

## Scope

In scope:
- Add a new CLI: `cmd/value-sniffer-radar-llm`
  - `-mode enrich|digest`
  - Input: `paper_log` JSONL (same as optimizer input)
  - Output:
    - `enriched.jsonl` (append or overwrite)
    - optional `digest.md`
- Add provider abstraction:
  - `provider=api` (HTTP)
  - `provider=cli` (exec)
- Strict JSON output contract from LLM (parseable). If parse fails, skip and log warning.
- Unit tests (no network):
  - JSON parsing
  - CLI provider invocation (using a dummy command)

Out of scope:
- Auto trading / execution
- Letting LLM change `tier` or trade decisions

## Acceptance (3 steps)

1) `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2) `cmd/value-sniffer-radar-llm -mode enrich` produces enriched output for a sample `paper.jsonl` using `provider=cli`.
3) README (or a short doc) shows how to run it with cloud API and CLI.

## Rollback

- Remove `cmd/value-sniffer-radar-llm` and `internal/llm` package.

