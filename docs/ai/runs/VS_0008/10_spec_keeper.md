# VS_0008 10_spec_keeper

## Pass/Fail gate

Pass if:
- Adds a repo-focused labeler CLI that converts `paper_log` events into `labels.jsonl` with `reward` labels.
- Default reward definition is an actionability proxy: at T+window, fused rate is still >= threshold (and confidence PASS).
- Supports a deterministic offline mode (mock fusion) so acceptance doesn't require network.
- Includes unit tests for the core labeling flow.

Fail if:
- Attempts to auto trade or to label ETFs/CBs in this ticket.
- Requires network access for acceptance.

