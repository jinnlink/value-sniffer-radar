# VS_0006 10_spec_keeper

## Pass/Fail gate

Pass if:
- Adds an offline optimizer CLI that reads `paper_log` JSONL and outputs a Markdown report with suggested action allocation.
- Implements a minimal bandit core (Beta-Bernoulli Thompson Sampling) and a deterministic scheduler helper.
- Includes unit tests for bandit, JSONL parsing, and scheduler.
- No auto-trading, no automatic config mutation.

Fail if:
- Introduces auto execution / broker integrations.
- Adds heavyweight ML deps.

## Acceptance (copy/paste)

1) `go test ./...`
2) Create sample and run:
   - Use the JSONL snippet in `docs/ai/runs/VS_0006/00_context.md`
   - `go run .\\cmd\\value-sniffer-radar-optimizer -- -in .\\state\\optimizer.sample.jsonl -out-md .\\state\\optimizer.report.md -seed 7 -slots 2`
3) `Test-Path .\\state\\optimizer.report.md`

