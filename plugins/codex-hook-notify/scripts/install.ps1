param(
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"
$pluginRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$repoRoot = Resolve-Path (Join-Path $pluginRoot "..\..")
$installer = Join-Path $repoRoot "scripts\install.ps1"

if (-not (Test-Path -LiteralPath $installer)) {
    throw "Root installer not found. Clone the full CodexHookNotify repository, then run scripts\\install.ps1 from the repository root."
}

& $installer -DryRun:$DryRun
