[CmdletBinding()]
param(
  [string]$Config = ".\\config.yaml",
  [string]$Paper = ".\\state\\paper.jsonl",
  [string]$Labels = ".\\state\\labels.repo.jsonl",
  [string]$ReportMD = ".\\state\\optimizer.report.md",
  [string]$Reco = ".\\state\\optimizer.reco.json",
  [string]$Windows = "10s,30s,5m",
  [string]$Grace = "30s",
  [int]$Slots = 30,
  [int]$Seed = 7,
  [int]$LabelWindowSec = 0,
  [int]$MaxLabelsPerRun = 200
)

$ErrorActionPreference = "Stop"

function Resolve-RepoRoot {
  $p = (Resolve-Path -LiteralPath ".").Path
  while ($true) {
    if (Test-Path -LiteralPath (Join-Path $p "go.mod")) { return $p }
    $parent = Split-Path -Parent $p
    if (-not $parent -or $parent -eq $p) { throw "Cannot locate repo root (go.mod) from: $pwd" }
    $p = $parent
  }
}

$repoRoot = Resolve-RepoRoot
$go = Join-Path $repoRoot "state\\_toolchains\\go1.25.6\\go\\bin\\go.exe"
if (-not (Test-Path -LiteralPath $go)) { $go = "go" }

Write-Host "[daily_loop] repo_root=$repoRoot"
Write-Host "[daily_loop] go=$go"
Write-Host "[daily_loop] config=$Config"
Write-Host "[daily_loop] paper=$Paper"

if (-not (Test-Path -LiteralPath $Paper)) {
  throw "paper_log not found: $Paper"
}

Write-Host "[daily_loop] step=labeler out=$Labels windows=$Windows grace=$Grace"
& $go run .\\cmd\\value-sniffer-radar-labeler `
  -config $Config `
  -in $Paper `
  -out $Labels `
  -windows $Windows `
  -grace $Grace `
  -max $MaxLabelsPerRun | Write-Host

Write-Host "[daily_loop] step=optimizer out_md=$ReportMD out_reco=$Reco slots=$Slots seed=$Seed"
$args = @(
  "run", ".\\cmd\\value-sniffer-radar-optimizer",
  "-in", $Paper,
  "-labels", $Labels,
  "-slots", "$Slots",
  "-seed", "$Seed",
  "-out-md", $ReportMD,
  "-out-reco", $Reco
)
if ($LabelWindowSec -gt 0) {
  $args += @("-label-window-sec", "$LabelWindowSec")
}
& $go @args | Write-Host

Write-Host "[daily_loop] done"
Write-Host "[daily_loop] labels=$Labels"
Write-Host "[daily_loop] report_md=$ReportMD"
Write-Host "[daily_loop] reco=$Reco"
