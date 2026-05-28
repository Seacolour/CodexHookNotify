param(
    [string]$CodexHome = (Join-Path $env:USERPROFILE ".codex"),
    [string]$ConfigPath = "",
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

$hooksDir = Join-Path ([System.IO.Path]::GetFullPath($CodexHome)) "hooks"
$exePath = Join-Path $hooksDir "notify-mail.exe"
if ([string]::IsNullOrWhiteSpace($ConfigPath)) {
    $ConfigPath = Join-Path $hooksDir "notify-mail.yaml"
}
$configPathFull = [System.IO.Path]::GetFullPath($ConfigPath)

if (-not (Test-Path -LiteralPath $exePath)) {
    throw "Executable not found: $exePath"
}
if (-not (Test-Path -LiteralPath $configPathFull)) {
    throw "Config not found: $configPathFull"
}

$json = '{"cwd":"D:\\Code\\Test","model":"gpt-5.5","last_assistant_message":"手动测试邮件"}'
$args = @("--config", $configPathFull, "--test-json", $json)
if ($DryRun) {
    $args = @("--config", $configPathFull, "--dry-run", "--test-json", $json)
}

& $exePath @args
