# VS_0013 30_reviewer

## Review Summary
- LLM is not in the hot path: it runs as a separate CLI against `paper.jsonl`.
- Enrich mode requires strict JSON and is safely skippable on parse errors.
- CLI provider is testable without network; API provider is provider-agnostic (OpenAI-compatible).

## Risks / Follow-ups
- CLI args parsing is simple (`strings.Fields`); if you need complex quoting, wrap it in a `.cmd`/`.ps1`.
- Digest mode is “LLM markdown output”; if you need strict format, ticket it separately.

