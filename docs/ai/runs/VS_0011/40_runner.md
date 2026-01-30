# VS_0011 40_runner

## Commands

### 1) Unit tests
`state/_toolchains/go1.25.6/go/bin/go.exe test ./...`

Observed: PASS.

### 2) Sample: emit reco JSON
`state/_toolchains/go1.25.6/go/bin/go.exe run .\\cmd\\value-sniffer-radar-optimizer -in .\\state\\paper.jsonl -labels .\\state\\labels.repo.jsonl -seed 7 -slots 30 -out-reco .\\state\\optimizer.reco.json`
