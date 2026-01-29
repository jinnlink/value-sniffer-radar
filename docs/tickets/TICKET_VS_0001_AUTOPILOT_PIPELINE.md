# Ticket VS_0001 — Bootstrap Industrial Autopilot Workflow

## Goal

Make this repo “industrialized”:
- One in-progress ticket queue
- Per-ticket run folders with file-based handoff (Spec → Build → Review → Run)
- Reusable templates + checklists

## Scope

In scope:
- Add/standardize process docs and templates required by the autopilot scripts
- Generate the first run folder for VS_0001 (prompts + workers)

Out of scope:
- Feature work in signals/data/execution
- Auto-trading

## Acceptance (3 steps)

1) `Test-Path docs\\tickets\\QUEUE.md`
2) `Test-Path SPEC.md; Test-Path IMPL.md; Test-Path docs\\process\\REVIEW_CHECKLIST.md`
3) Run (generates run folder + workers):
   - `powershell -NoProfile -ExecutionPolicy Bypass -File \"$env:USERPROFILE\\.codex\\skills\\ai-autopilot-pipeline\\scripts\\new_ticket_run.ps1\" -RepoRoot .`
   - Then verify `Test-Path docs\\ai\\runs\\VS_0001\\_mwr\\workers_10_spec_keeper.json`

## Rollback

- Delete `docs/tickets`, `docs/ai/templates`, `docs/process` and repo-root `SPEC.md`/`IMPL.md`.

