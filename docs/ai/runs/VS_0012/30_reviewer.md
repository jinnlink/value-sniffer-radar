# VS_0012 30_reviewer

## Review Summary
- Script is minimal and safe: it fails fast when `paper.jsonl` is missing; does not embed secrets.
- Uses repo-local Go toolchain if present; otherwise uses system `go`.

## Risks
- The labeler may require realtime marketdata when not using `-mock-rate`; ensure `marketdata.enabled=true` if you rely on realtime labels.
- Engine reads reco at startup; you must restart radar to apply new quotas (documented).

