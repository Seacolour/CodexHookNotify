# CodexHookNotify Agent Guide

Use this guide when a user asks you to install, update, test, or troubleshoot CodexHookNotify from this repository.

## Safety Rules

- Do not ask the user to paste SMTP passwords, SMTP authorization codes, app passwords, or mailbox tokens into chat.
- Do not print the contents of `notify-mail.yaml` unless all sensitive fields are redacted.
- Do not overwrite an existing `~/.codex/hooks.json` by hand. Use `scripts/install.ps1`, which creates a backup before writing.
- Do not remove unrelated Codex hooks. Preserve other hook groups and only add or update the `notify-mail.exe` command.
- Remind the user that Codex Desktop must be fully restarted and the Stop hook must be trusted before it can run.

## Install Workflow

1. Confirm the host is Windows PowerShell.
2. From the repository root, run:

   ```powershell
   .\scripts\install.ps1
   ```

3. Ask the user to edit the config locally:

   ```powershell
   notepad $env:USERPROFILE\.codex\hooks\notify-mail.yaml
   ```

4. Tell the user to fill in SMTP settings, especially `smtp.username`, `smtp.password`, `smtp.from`, and `smtp.to`.
5. Run a dry-run parse test after the config exists:

   ```powershell
   $exe = "$env:USERPROFILE\.codex\hooks\notify-mail.exe"
   $cfg = "$env:USERPROFILE\.codex\hooks\notify-mail.yaml"
   & $exe --config $cfg --dry-run --test-json '{"cwd":"D:\\Code\\Test","model":"gpt-5.5","last_assistant_message":"手动测试邮件"}'
   ```

6. If the user wants to test actual email delivery, run the same command without `--dry-run`.
7. Read logs with UTF-8:

   ```powershell
   Get-Content "$env:USERPROFILE\.codex\hooks\notify-mail.log" -Tail 20 -Encoding UTF8
   ```

## Expected Files

- Executable: `%USERPROFILE%\.codex\hooks\notify-mail.exe`
- User config: `%USERPROFILE%\.codex\hooks\notify-mail.yaml`
- Codex hook registry: `%USERPROFILE%\.codex\hooks.json`
- Runtime log: `%USERPROFILE%\.codex\hooks\notify-mail.log`

## Update Workflow

1. Pull or download the latest repository.
2. Run `.\scripts\install.ps1` again.
3. Preserve the user's existing `notify-mail.yaml`.
4. Verify that `hooks.json` still contains the Stop hook command.

## Troubleshooting Hints

- If no email arrives, check whether Codex Desktop was restarted and the hook was trusted.
- If manual send works but Codex does not trigger it, inspect `~/.codex/hooks.json`.
- If the user expects a Markdown attachment, check `attachment.enabled` and `attachment.mode`; the default only attaches when the preview is truncated.
- If the email has no session title, check whether `~/.codex/session_index.jsonl` contains the hook `session_id`.
- If a notification is skipped unexpectedly, check `session.skipUnindexed`; the default skips Stop events whose `session_id` is absent from `~/.codex/session_index.jsonl`.
- If Chinese text appears as question marks only during manual tests, avoid PowerShell pipes and use `--test-json` or `--test-json-file`.
- If logs look garbled in PowerShell, read them with `-Encoding UTF8`.
