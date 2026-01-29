I will read the necessary context files, ticket details, specifications, and the template to generate the Spec Keeper report.
I will read `IMPL.md` to check for any language constraints or implementation conventions that might conflict with the ticket's "stdlib-only Python" requirement.
I will generate the `10_spec_keeper.md` file with the assessment of the ticket, noting the minor conflict regarding the programming language and providing clear guardrails and acceptance criteria.
# Spec Keeper: Ticket VS_0002

## Pass/Fail
**PASS.** Ticket is clear, scoped, and actionable; specific override for Python tools is acceptable.

## Conflicts / Ambiguities
1. **Language Convention**: `IMPL.md` specifies Go, but ticket requests Python.
   - *Decision*: **Allow.** Offline analysis tools (`tools/`) may use Python (stdlib only) as requested, while the main engine remains Go.

## Do / Don't (Scope Guardrails)
- **Do**:
  - Use **Python Standard Library only** (no `pip install`).
  - Read JSONL from file/stdin.
  - Output simple text summaries to stdout.
- **Don't**:
  - Add `requirements.txt` or external dependencies (e.g., `pandas`).
  - Fetch network data (prices).
  - Modify any Go code in `cmd/` or `internal/`.

## Acceptance Reminders
1. `Test-Path tools\paper_eval.py`
2. `python tools\paper_eval.py --help`
3. **Smoke Test**:
   ```powershell
   @'{"ts":"2026-01-29T00:00:00+08:00","event":{"source":"cb_double_low_action","trade_date":"20260128","market":"CN-A","symbol":"110000.SH","title":"demo","body":"","tags":{"tier":"action"}}}'@ | Set-Content -Encoding UTF8 .\state\paper.sample.jsonl
   python tools\paper_eval.py --in .\state\paper.sample.jsonl
   ```

## Decision Points (0–3)
- None.
