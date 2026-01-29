# VS_0009 40_runner

## Commands

### 1) Unit tests
`state/_toolchains/go1.25.6/go/bin/go.exe test ./...`

Output:
```
?   	value-sniffer-radar/cmd/value-sniffer-radar	[no test files]
?   	value-sniffer-radar/cmd/value-sniffer-radar-labeler	[no test files]
?   	value-sniffer-radar/cmd/value-sniffer-radar-optimizer	[no test files]
?   	value-sniffer-radar/internal/config	[no test files]
?   	value-sniffer-radar/internal/engine	[no test files]
ok  	value-sniffer-radar/internal/labeler	1.224s
ok  	value-sniffer-radar/internal/marketdata	(cached)
?   	value-sniffer-radar/internal/notifier	[no test files]
ok  	value-sniffer-radar/internal/optimizer	1.166s
ok  	value-sniffer-radar/internal/signals	(cached)
?   	value-sniffer-radar/internal/tushare	[no test files]
```

### 2) Sample: ranking flip with labels

Paper-only:
`state/_toolchains/go1.25.6/go/bin/go.exe run .\\cmd\\value-sniffer-radar-optimizer -in state\\tmp\\paper.jsonl -slots 2`

With labels:
`state/_toolchains/go1.25.6/go/bin/go.exe run .\\cmd\\value-sniffer-radar-optimizer -in state\\tmp\\paper.jsonl -labels state\\tmp\\labels.repo.jsonl -label-window-sec 30 -slots 2`

Observed:
- Paper-only ranks `sig_A` above `sig_B`
- With labels (window=30), ranks `sig_B` above `sig_A`
- Report includes coverage + per-signal/window reward tables

