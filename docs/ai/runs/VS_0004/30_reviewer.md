# VS_0004 30_reviewer

## Summary

Minimal, targeted fix: `paper_eval.py` now tolerates UTF-8 BOM produced by PowerShell `Set-Content -Encoding UTF8`.

## Issues

### Blocking

- None.

### Minor

- For stdin mode (`--in -`), we do not control encoding; the per-line `lstrip("\ufeff")` still helps, but some pipelines may inject other non-UTF8 bytes. This is acceptable for this ticket.

## Checklist

- Scope creep: ✅ none
- Dependencies: ✅ stdlib-only remains
- Acceptance: ✅ still 3 copy/paste steps

