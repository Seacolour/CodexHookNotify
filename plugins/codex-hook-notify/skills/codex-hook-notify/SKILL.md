---
name: codex-hook-notify
description: Install, update, test, or troubleshoot CodexHookNotify, a Windows-first SMTP email notifier for Codex Desktop Stop hooks. Use when a user asks Codex to install this repository, configure Codex email completion notifications, inspect notify-mail.exe, update hooks.json, or debug why Codex task-complete email notifications are not arriving.
---

# Codex Hook Notify

Use this skill to install and maintain CodexHookNotify without exposing SMTP secrets.

## Safety

- Never ask the user to paste SMTP passwords, app passwords, or authorization codes into chat.
- Never print `notify-mail.yaml` unless sensitive fields are redacted.
- Preserve unrelated hooks in `~/.codex/hooks.json`.
- Remind the user to restart Codex Desktop and trust the Stop hook after install.

## Install

From the repository root, run:

```powershell
.\scripts\install.ps1
```

Then ask the user to edit:

```powershell
notepad $env:USERPROFILE\.codex\hooks\notify-mail.yaml
```

They must fill in SMTP settings locally.

## Verify

Check expected files:

```powershell
Test-Path "$env:USERPROFILE\.codex\hooks\notify-mail.exe"
Test-Path "$env:USERPROFILE\.codex\hooks\notify-mail.yaml"
Test-Path "$env:USERPROFILE\.codex\hooks.json"
```

Run a dry-run parse test:

```powershell
.\scripts\test-mail.ps1 -DryRun
```

If session titles are missing, check whether `~/.codex/session_index.jsonl` contains the hook `session_id`.

If Markdown attachments are missing, check `attachment.enabled` and `attachment.mode`; the default only attaches when the email preview is truncated.

Send a real test email only if the user explicitly asks:

```powershell
.\scripts\test-mail.ps1
```

Read logs with UTF-8:

```powershell
Get-Content "$env:USERPROFILE\.codex\hooks\notify-mail.log" -Tail 20 -Encoding UTF8
```
