# Ticket VS_0009 — Optimizer Consumes labels.repo.jsonl

## Goal

Make the optimizer use real labels produced by the labeler:
- Input: `labels.repo.jsonl` (append-only)
- Output: allocator report that reflects labeled outcomes

## Scope

In scope:
- Add a `-labels` input flag to optimizer CLI (optional).
- If `-labels` provided, prefer rewards from labels file rather than `event.data.reward`.
- Produce report sections:
  - coverage: % events labeled per window
  - per-signal reward rate by window

Out of scope:
- New price sources
- Auto execution

## Acceptance (3 steps)

1) `go test ./...`
2) Given a small sample `labels.repo.jsonl`, optimizer report changes ranking.
3) Report includes a labeled coverage summary.

## Rollback

- Revert optimizer input changes.

