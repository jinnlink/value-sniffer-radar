# VS_0013 40_runner

## Commands

### 1) Unit tests
`state/_toolchains/go1.25.6/go/bin/go.exe test ./...`

### 2) CLI provider smoke test (no network)
Create a tiny sample `paper.jsonl` (or use your existing one), then run:

`state/_toolchains/go1.25.6/go/bin/go.exe run .\\cmd\\value-sniffer-radar-llm -mode enrich -provider cli -cli-cmd cmd.exe -cli-args "/c echo {\"summary\":\"ok\",\"risks\":[],\"checklist\":[]}" -in .\\state\\paper.jsonl -out .\\state\\llm.enriched.jsonl`

