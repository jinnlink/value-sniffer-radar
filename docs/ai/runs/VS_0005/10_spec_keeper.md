# VS_0005 10_spec_keeper

## Pass/Fail gate (this ticket)

Pass if:
- A realtime `marketdata` interface exists and is usable by a repo realtime signal.
- Multi-source fusion enforces `tier=action` confidence gating (no “single-source action” by default).
- Throttling/backoff/circuit-breaker exist and are unit-tested.
- Example config includes the new `marketdata` section.

Fail if:
- Scope creep into broker execution or full-market scanning.
- No deterministic unit tests for fusion/throttling/circuit-breaker.

## Required decisions (chair)

1) Provider combo for v1 (choose one): A / B / C
2) `rate_pct` definition + raw fields to persist
3) Default thresholds:
   - `required_sources`
   - `max_abs_diff`
   - `staleness_sec`

