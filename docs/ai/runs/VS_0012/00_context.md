# VS_0012 00_context

## Ticket
- Queue line: - Ticket VS_0012 (`docs/tickets/TICKET_VS_0012_DAILY_LOOP_SCRIPT.md`): One-command daily loop (labeler→optimizer→reco) + report
- Ticket doc: docs/tickets/TICKET_VS_0012_DAILY_LOOP_SCRIPT.md

## Do / Don't (scope guardrails)
- Do:
  - Add a one-command PowerShell script for daily loop: labeler → optimizer → reco.
  - Update README with the script usage and `engine.reco_path` wiring.
- Don't:
  - Add service installation / auto restart in this ticket.
  - Add auto trading.

## Decision points (0-3)
- Go toolchain selection:
  - Use repo-local toolchain if present, else fall back to system `go`.

## Expected minimal changes
- Files:
  - `tools/daily_loop.ps1`
  - `README.md`

## Acceptance (3 steps, copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. Run `tools/daily_loop.ps1` with sample `paper.jsonl` and confirm it produces labels/report/reco files.
3. README documents daily loop + `engine.reco_path`.

## Rollback
- Remove `tools/daily_loop.ps1` and revert README changes.

