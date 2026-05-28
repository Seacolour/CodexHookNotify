param(
    [string]$CodexHome = (Join-Path $env:USERPROFILE ".codex"),
    [string]$ConfigPath = "",
    [int]$TimeoutSeconds = 20,
    [switch]$SkipBuild,
    [switch]$ForceConfig,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message)
    Write-Host "[CodexHookNotify] $Message"
}

function Set-JsonProperty {
    param(
        [Parameter(Mandatory = $true)] [object]$Object,
        [Parameter(Mandatory = $true)] [string]$Name,
        [Parameter(Mandatory = $true)] [object]$Value
    )

    if ($Object.PSObject.Properties.Name -contains $Name) {
        $Object.$Name = $Value
    }
    else {
        $Object | Add-Member -MemberType NoteProperty -Name $Name -Value $Value
    }
}

function Update-HooksJson {
    param(
        [Parameter(Mandatory = $true)] [string]$Path,
        [Parameter(Mandatory = $true)] [string]$Command,
        [Parameter(Mandatory = $true)] [int]$TimeoutSeconds,
        [switch]$DryRun
    )

    $hookEntry = [pscustomobject]@{
        type = "command"
        command = $Command
        timeout = $TimeoutSeconds
        statusMessage = "发送完成通知邮件"
    }

    if (Test-Path -LiteralPath $Path) {
        $data = Get-Content -LiteralPath $Path -Encoding UTF8 -Raw | ConvertFrom-Json
    }
    else {
        $data = [pscustomobject]@{}
    }

    if ($null -eq $data.hooks) {
        Set-JsonProperty -Object $data -Name "hooks" -Value ([pscustomobject]@{})
    }

    $stopGroups = @()
    if ($null -ne $data.hooks.Stop) {
        $stopGroups = @($data.hooks.Stop)
    }

    $updated = $false
    foreach ($group in $stopGroups) {
        foreach ($hook in @($group.hooks)) {
            if ($hook.command -and ($hook.command -match "notify-mail(\.exe)?")) {
                Set-JsonProperty -Object $hook -Name "type" -Value "command"
                Set-JsonProperty -Object $hook -Name "command" -Value $Command
                Set-JsonProperty -Object $hook -Name "timeout" -Value $TimeoutSeconds
                Set-JsonProperty -Object $hook -Name "statusMessage" -Value "发送完成通知邮件"
                $updated = $true
            }
        }
    }

    if (-not $updated) {
        $stopGroups += [pscustomobject]@{
            hooks = @($hookEntry)
        }
    }

    Set-JsonProperty -Object $data.hooks -Name "Stop" -Value $stopGroups
    $json = $data | ConvertTo-Json -Depth 20

    if ($DryRun) {
        Write-Step "Dry-run: would write $Path"
        Write-Host $json
        return
    }

    $parent = Split-Path -Parent $Path
    New-Item -ItemType Directory -Force -Path $parent | Out-Null

    if (Test-Path -LiteralPath $Path) {
        $stamp = Get-Date -Format "yyyyMMdd-HHmmss"
        Copy-Item -LiteralPath $Path -Destination "$Path.bak-$stamp" -Force
        Write-Step "Backed up existing hooks.json to $Path.bak-$stamp"
    }

    $json | Set-Content -LiteralPath $Path -Encoding UTF8
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$codexHomeFull = [System.IO.Path]::GetFullPath($CodexHome)
$hooksDir = Join-Path $codexHomeFull "hooks"
$exePath = Join-Path $hooksDir "notify-mail.exe"
if ([string]::IsNullOrWhiteSpace($ConfigPath)) {
    $ConfigPath = Join-Path $hooksDir "notify-mail.yaml"
}
$configPathFull = [System.IO.Path]::GetFullPath($ConfigPath)
$hooksJsonPath = Join-Path $codexHomeFull "hooks.json"

Write-Step "Repository: $repoRoot"
Write-Step "Codex home: $codexHomeFull"
Write-Step "Executable: $exePath"
Write-Step "Config: $configPathFull"

if (-not $DryRun) {
    New-Item -ItemType Directory -Force -Path $hooksDir | Out-Null
}

if ($SkipBuild) {
    Write-Step "Skipping Go build"
}
else {
    if ($DryRun) {
        Write-Step "Dry-run: would run build.ps1"
    }
    else {
        & (Join-Path $repoRoot "build.ps1") -OutputDir $hooksDir
    }
}

$exampleConfig = Join-Path $repoRoot "notify-mail.yaml.example"
if ((Test-Path -LiteralPath $configPathFull) -and -not $ForceConfig) {
    Write-Step "Keeping existing config: $configPathFull"
}
else {
    if ($DryRun) {
        Write-Step "Dry-run: would copy $exampleConfig to $configPathFull"
    }
    else {
        Copy-Item -LiteralPath $exampleConfig -Destination $configPathFull -Force
        Write-Step "Created config template: $configPathFull"
    }
}

$command = "`"$exePath`" --config `"$configPathFull`""
Update-HooksJson -Path $hooksJsonPath -Command $command -TimeoutSeconds $TimeoutSeconds -DryRun:$DryRun

Write-Step "Done."
Write-Step "Next: edit $configPathFull, restart Codex Desktop, then trust the Stop hook."
