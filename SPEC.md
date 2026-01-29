# Value Sniffer Radar — Engineering Spec (Repo Contract)

This repository is run as an **industrial, ticket-driven pipeline**:

`QUEUE.md` → **Spec Keeper** → **Builder** → **Reviewer** → **Runner** → **Docs** → next ticket.

## Source of Truth (contract priority)

1) `SPEC.md` (this file) — process + quality gates
2) `IMPL.md` — implementation conventions
3) Ticket doc referenced from `docs/tickets/QUEUE.md`
4) Everything else

## Ticket workflow (single in-progress rule)

- The queue is `docs/tickets/QUEUE.md`.
- There must be **exactly one** ticket under `## In Progress`.
- Each ticket line must include:
  - `Ticket <ID>`
  - a ticket doc path in backticks, e.g. ``(`docs/tickets/TICKET_<ID>_*.md`)``

## Autopilot run folders (file handoff, not chat)

For each ticket `<ID>`, the run folder is:
- `docs/ai/runs/<ID>/`

Stages (fixed filenames):
- `00_context.md` — scope guardrails + acceptance + rollback (short)
- `10_spec_keeper.md` — pass/fail + guardrails + acceptance reminders
- `20_builder.md` — implementation report (files + acceptance commands + rollback)
- `25_diff.patch` — optional but recommended for review
- `30_reviewer.md` — issues by severity + fix order
- `40_runner.md` — commands + raw output + failures
- `50_conclusion.md` — final decision + what shipped

## Quality gates (must pass)

- **No scope creep** beyond the current ticket.
- **No secrets** in repo or logs (tokens/keys only via env).
- **Actionable acceptance**: every ticket defines 3 copy/paste checks.
- **Review is blocking**: any “blocking” issue must be fixed before closing a ticket.
- **Docs updated** when behavior/config changes.

## Safety

- No advice or automation that violates laws/regulations or broker/exchange terms.
- No auto-trading unless a ticket explicitly scopes it and includes risk controls + paper validation.

