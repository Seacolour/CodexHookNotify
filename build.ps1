param(
    [string]$OutputDir = (Join-Path $env:USERPROFILE ".codex\hooks"),
    [switch]$SkipTidy
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$outDir = [System.IO.Path]::GetFullPath($OutputDir)
$exePath = Join-Path $outDir "notify-mail.exe"

New-Item -ItemType Directory -Force -Path $outDir | Out-Null

Push-Location $root
try {
    if (-not $SkipTidy) {
        go mod tidy
    }
    go build -o $exePath ./cmd/notify-mail
}
finally {
    Pop-Location
}

Write-Host "Built: $exePath"
