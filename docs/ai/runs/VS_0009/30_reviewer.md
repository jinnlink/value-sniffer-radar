# VS_0009 30_reviewer

## Review Summary
- Meets ticket scope: label ingestion is optional, label rewards are preferred, and report adds coverage + per-window stats.
- No import cycles introduced; labeler now calls `optimizer.EventID()` (good).

## Notes / Risks
- Labels file is append-only; if duplicates ever appear, `ReadLabelsJSONL` is “last write wins” (acceptable).
- Default primary window selection is “most labels” (coverage-driven). This is reasonable, but users should pass `-label-window-sec` when they want a specific horizon.
- Realtime labels depend on unofficial providers; optimizer is offline-only and safe.

## Validation
- Unit tests cover:
  - window auto-selection
  - ranking change when labels invert outcomes

