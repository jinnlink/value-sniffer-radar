# Runner Output

## Environment
- time: 2026-01-29T11:04:11
- pwd: F:\文档修复\new
- powershell: 5.1.19041.3803
- python: Python 3.13.9

## Commands + Output
### Step 1

```powershell
Test-Path tools\\paper_eval.py
```

```
True
```

### Step 2

```powershell
python tools\\paper_eval.py --help
```

```
usage: paper_eval.py [-h] --in IN_PATH [--out-md OUT_MD]

Value Sniffer Radar - paper_log evaluator (stdlib-only)

options:
  -h, --help       show this help message and exit
  --in IN_PATH     Input JSONL path (paper_log). Use '-' for stdin.
  --out-md OUT_MD  Optional Markdown report output path.
```

### Step 3

```powershell
$sample = @'
{"ts":"2026-01-29T00:00:00+08:00","event":{"source":"cb_double_low_action","trade_date":"20260128","market":"CN-A","symbol":"110000.SH","title":"demo","body":"","tags":{"tier":"action"}}}
'@
$sample | Set-Content -Encoding UTF8 .\\state\\paper.sample.jsonl
python tools\\paper_eval.py --in .\\state\\paper.sample.jsonl
```

```
python : [warn] invalid json line=1
At line:45 char:3
+   python 'tools\paper_eval.py' --in '.\state\paper.sample.jsonl'
+   ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
    + CategoryInfo          : NotSpecified: ([warn] invalid json line=1:String) [], RemoteException
    + FullyQualifiedErrorId : NativeCommandError
 
# Paper Eval Report

- generated_at: `2026-01-29T11:04:12+08:00`
- input: `state\paper.sample.jsonl`
- total_events: `0`

## By Tier
_(none)_

## By Signal (source)
_(none)_

## Top Symbols
_(none)_

## By Trade Date
_(none)_


## Notes
- This ticket does **not** fetch prices; PnL/returns require a separate data-source ticket.
```

