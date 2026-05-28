param(
    [string]$OutputDir = (Join-Path $env:USERPROFILE ".codex\hooks"),
    [string]$Version = "",
    [switch]$SkipTidy
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$outDir = [System.IO.Path]::GetFullPath($OutputDir)
$exePath = Join-Path $outDir "notify-mail.exe"

New-Item -ItemType Directory -Force -Path $outDir | Out-Null

Push-Location $root
try {
    if ([string]::IsNullOrWhiteSpace($Version)) {
        $Version = (git describe --tags --always --dirty 2>$null)
        if ([string]::IsNullOrWhiteSpace($Version)) {
            $Version = "dev"
        }
    }
    if (-not $SkipTidy) {
        go mod tidy
    }
    go build -ldflags "-s -w -X main.version=$Version" -o $exePath ./cmd/notify-mail
}
finally {
    Pop-Location
}

Write-Host "Built: $exePath"
Write-Host "Version: $Version"
