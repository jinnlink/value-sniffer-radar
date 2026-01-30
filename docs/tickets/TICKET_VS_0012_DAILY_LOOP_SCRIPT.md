# Ticket VS_0012 — Daily Loop Script (labeler → optimizer → reco) + Report

## Goal

Make the closed-loop workflow runnable by a single command, so you can operate the system daily:
- label yesterday/today’s `paper.jsonl` into `labels.repo.jsonl`
- run optimizer and write:
  - `optimizer.report.md` (human)
  - `optimizer.reco.json` (machine)

## Scope

In scope:
- Add `tools/daily_loop.ps1` (PowerShell) that:
  - runs `value-sniffer-radar-labeler`
  - runs `value-sniffer-radar-optimizer` with `-out-md` and `-out-reco`
  - prints the output paths
- Update docs/README for how to use `engine.reco_path` with this daily loop.

Out of scope:
- Windows service installation / auto restart (ticket separately)
- Auto trading

## Acceptance (3 steps)

1) `state/_toolchains/go1.25.6/go/bin/go.exe test ./...`
2) Running `tools/daily_loop.ps1` with a sample `paper.jsonl` produces `labels.repo.jsonl`, `optimizer.report.md`, and `optimizer.reco.json`.
3) README documents the daily loop and how to point `engine.reco_path` to the reco file.

## Rollback

- Remove `tools/daily_loop.ps1` and README changes.

