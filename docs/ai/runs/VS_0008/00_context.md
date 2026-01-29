# VS_0008 00_context

## Ticket
- Queue line: - Ticket VS_0008 (`docs/tickets/TICKET_VS_0008_REWARD_LABELING.md`): Add reward labeling (price windows + costs) for optimizer
- Ticket doc: docs/tickets/TICKET_VS_0008_REWARD_LABELING.md

## Do / Don't (scope guardrails)
- Do:
  - Implement a repo-focused reward labeling tool that derives `reward` from post-alert persistence (T+window) using realtime marketdata fusion.
  - Keep tool offline-friendly: it reads `paper_log` JSONL and writes `labels.jsonl` (append-only).
  - Include cost model placeholders only (no heavy dependencies).
- Don't:
  - Don't do auto execution.
  - Don't expand to ETFs/CBs in this ticket.
  - Don't mutate `paper_log` in place by default; prefer a sidecar labels file.

## Decision points (0-3)
- Windows: start with `10s,30s,300s` and a grace period for late sampling.
- Reward: for repo, reward=1 if fused rate stays >= threshold at T+window (actionability proxy).

## Expected minimal changes
- Files:
  - `cmd/value-sniffer-radar-labeler/main.go` (new)
  - `internal/labeler/*` (new)
  - `docs/ai/runs/VS_0008/*` (run artifacts)

## Acceptance (3 steps, copy/paste)
1) `go test ./...`
2) Create a small paper log sample and run labeler in mock mode (no network):
   - `@'{"ts":"2026-01-29T00:00:00+08:00","event":{"source":"cn_repo_realtime_action","trade_date":"20260129","market":"CN-A","symbol":"204001.SH","title":"demo","body":"","tags":{"tier":"action","kind":"repo"},"data":{"consensus_rate_pct":5.0}}}'@ | Set-Content -Encoding UTF8 .\\state\\paper.sample.repo.jsonl`
   - `go run .\\cmd\\value-sniffer-radar-labeler -- -config .\\config.yaml -in .\\state\\paper.sample.repo.jsonl -out .\\state\\labels.repo.jsonl -windows 10s -grace 24h -mock-rate 1.6 -mock-confidence PASS`
3) `Test-Path .\\state\\labels.repo.jsonl`

## Rollback
- Delete `cmd/value-sniffer-radar-labeler` and `internal/labeler` and revert docs changes.

