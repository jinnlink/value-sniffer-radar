# Builder Output

## Impact

- Scope: process/bootstrap only (no trading features).
- Adds repo contracts: `SPEC.md`, `IMPL.md`.
- Adds ticket workflow: `docs/tickets/*`, `docs/process/REVIEW_CHECKLIST.md`, `docs/ai/templates/*`.
- Generates autopilot run scaffolding: `docs/ai/runs/VS_0001/*`.

## Files Changed

- `SPEC.md`
- `IMPL.md`
- `docs/process/REVIEW_CHECKLIST.md`
- `docs/ai/templates/SPEC_KEEPER.md`
- `docs/ai/templates/BUILDER.md`
- `docs/ai/templates/REVIEWER.md`
- `docs/ai/templates/RUNNER.md`
- `docs/tickets/QUEUE.md`
- `docs/tickets/TICKET_VS_0001_AUTOPILOT_PIPELINE.md`
- `docs/ai/runs/VS_0001/00_context.md`
- `docs/ai/runs/VS_0001/_mwr/*`

## Acceptance (3 steps)

1. `Test-Path docs\\tickets\\QUEUE.md`
2. `Test-Path SPEC.md; Test-Path IMPL.md; Test-Path docs\\process\\REVIEW_CHECKLIST.md`
3. `powershell -NoProfile -ExecutionPolicy Bypass -File \"$env:USERPROFILE\\.codex\\skills\\ai-autopilot-pipeline\\scripts\\new_ticket_run.ps1\" -RepoRoot .; Test-Path docs\\ai\\runs\\VS_0001\\_mwr\\workers_10_spec_keeper.json`

## Rollback

- `git revert` the VS_0001 commits; or delete the added process/ticket files and run folders.

