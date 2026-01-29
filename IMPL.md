# Value Sniffer Radar — Implementation Notes

## Code conventions

- Language: Go (module name: `value-sniffer-radar`)
- Entry: `cmd/value-sniffer-radar/main.go`
- Config: `internal/config` (YAML)
- Engine loop: `internal/engine`
- Signals: `internal/signals`
- Notifiers: `internal/notifier`
- Data: `internal/tushare` (history/slow data)

## Non-negotiables

- Never commit secrets. Use env vars (e.g. `TUSHARE_TOKEN`).
- Prefer **small, checkable diffs** per ticket.
- Keep features behind config flags.

## “Observe vs Action” tiers

- `tier=observe`: broad coverage, lower frequency, looser thresholds.
- `tier=action`: fewer alerts, stricter filters, higher frequency.

## Testing philosophy

- Start with the smallest check that proves the ticket.
- If no tests exist yet, prefer adding minimal smoke checks or log-based validation.

