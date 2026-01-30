[CmdletBinding()]
param(
  [string]$Model = ""
)

$ErrorActionPreference = "Stop"
$prompt = [Console]::In.ReadToEnd()
if (-not $prompt -or $prompt.Trim() -eq "") { throw "empty stdin prompt" }

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
$schema = Join-Path $repoRoot "tools\\llm_schema\\enrich.schema.json"
if (-not (Test-Path -LiteralPath $schema)) { throw "schema not found: $schema" }

$tmp = Join-Path $env:TEMP ("vsr-codex-last-" + [Guid]::NewGuid().ToString("n") + ".txt")
try {
  $args = @("exec", "--skip-git-repo-check", "--output-schema", $schema, "--output-last-message", $tmp)
  if ($Model -and $Model.Trim() -ne "") {
    $args += @("-m", $Model)
  }

  $prompt | codex.cmd @args | Out-Null

  if (-not (Test-Path -LiteralPath $tmp)) { throw "codex did not write last message" }
  Get-Content -LiteralPath $tmp -Raw -Encoding UTF8 | Write-Output
} finally {
  Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
}

