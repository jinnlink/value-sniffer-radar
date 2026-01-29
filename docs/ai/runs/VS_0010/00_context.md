# VS_0010 00_context

## Ticket
- Queue line: - Ticket VS_0010 (`docs/tickets/TICKET_VS_0010_ACTION_BUDGET_POLICY.md`): Action budget + quality gate (target ~30 action/day)
- Ticket doc: docs/tickets/TICKET_VS_0010_ACTION_BUDGET_POLICY.md

## Do / Don't (scope guardrails)
- Do:
  - Add a policy layer that enforces "action" budget and downgrades to "observe" when needed.
  - Compute and attach `net_edge_pct` into `event.data` for paper logging.
  - Add unit tests for budget and net-edge gating.
- Don't:
  - Add new market data providers.
  - Add auto execution / auto trading.
  - Build a full backtester.

## Decision points (0-3)
- Default behavior when budgets exceeded:
  - Downgrade to `tier=observe` (keep coverage) instead of dropping.
- Net-edge gating default:
  - Disabled by default (`action_net_edge_min_pct: 0.0`) for backward compatibility.

## Expected minimal changes
- Files:
  - `internal/engine/*` (policy stage)
  - `internal/config/config.go` + `configs/config.example.yaml` (policy config)
  - `internal/signals/*` (write `expected_edge_pct` for core signals)
  - `internal/engine/policy_test.go`

## Acceptance (3 steps, copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. With `action_max_events_per_day: 30`, synthetic burst is capped to <= 30 action/day (tests).
3. Paper-logged events contain `net_edge_pct` and failing events are downgraded to observe (tests).

## Rollback
- Revert policy stage integration and config additions (VS_0010).

