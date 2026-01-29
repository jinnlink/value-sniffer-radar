# VS_0010 10_spec_keeper

## Contract Check
- Ticket: `docs/tickets/TICKET_VS_0010_ACTION_BUDGET_POLICY.md`
- Goal: enforce action budgets + add net-edge quality gate (alerts-first).
- In scope:
  - per-signal daily action quotas
  - net-edge computation in `event.data`
  - downgrade action→observe when failing
  - unit tests
- Out of scope: new data providers, auto trading, full backtester.

## Guardrails
- Backward compatible defaults (net-edge gate disabled unless configured).
- When caps exceeded, prefer downgrade→observe rather than silent drop.
- Ensure paper log always carries net-edge fields when `expected_edge_pct` exists.

## Acceptance (copy/paste)
1. `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2. Tests prove action/day <= 30 and per-signal cap works.
3. Tests prove net-edge below threshold never remains `tier=action`.

## Rollback
- Revert changes in `internal/engine`, `internal/config`, and affected signals.

