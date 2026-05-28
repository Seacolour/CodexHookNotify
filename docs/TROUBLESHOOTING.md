# Troubleshooting

## Manual test sends mail, but Codex does not

- Fully quit and restart Codex Desktop after editing `~/.codex/hooks.json`.
- Open the Hooks page and trust the Stop hook.
- Confirm the command points to existing files:

  ```powershell
  Test-Path "$env:USERPROFILE\.codex\hooks\notify-mail.exe"
  Test-Path "$env:USERPROFILE\.codex\hooks\notify-mail.yaml"
  ```

## No email arrives

- Check the log:

  ```powershell
  Get-Content "$env:USERPROFILE\.codex\hooks\notify-mail.log" -Tail 20 -Encoding UTF8
  ```

- For QQ Mail, use the SMTP authorization code, not the account password.
- Try `587 + starttls`; if your provider requires implicit TLS, use `465 + tls`.
- Check whether the same notification was skipped by the dedup window.

## Chinese text becomes question marks

Avoid PowerShell pipes for manual JSON tests:

```powershell
& $exe --config $cfg --test-json '{"cwd":"D:\\Code\\Test","model":"gpt-5.5","last_assistant_message":"手动测试邮件"}'
```

Codex Desktop sends UTF-8 JSON directly to stdin, so this usually affects only manual PowerShell tests.

## Logs look garbled

Read logs as UTF-8:

```powershell
Get-Content "$env:USERPROFILE\.codex\hooks\notify-mail.log" -Tail 20 -Encoding UTF8
```

## Duplicate notifications

The default dedup window is 30 seconds:

```yaml
dedup:
  enabled: true
  windowSeconds: 30
```

Increase or disable it in `notify-mail.yaml` if needed.

## Markdown attachment is missing

By default, Markdown attachments are only added when the email preview is truncated:

```yaml
attachment:
  enabled: true
  mode: when_truncated
```

Use `mode: always` if every notification should include a Markdown file. Empty assistant replies are not attached.

## Update notice is missing

Check the installed binary version:

```powershell
& "$env:USERPROFILE\.codex\hooks\notify-mail.exe" --version
```

Update notices are disabled for development builds such as `dev` or `ci`. They are also cached by `update.intervalHours`, so a new check may not run on every email. To skip one specific release:

```yaml
update:
  skippedVersions:
    - v0.1.4
```

## Session title is missing

CodexHookNotify looks up titles in:

```text
%USERPROFILE%\.codex\session_index.jsonl
```

The lookup is best-effort. If the hook `session_id` is not present in that file, the email falls back to the raw session id. You can also disable the lookup:

```yaml
session:
  titleLookup: false
```

## Expected notification was skipped

By default, CodexHookNotify skips sessions that are not listed in Codex Desktop's local `session_index.jsonl`. This filters internal context-summary or memory-maintenance turns. Disable it if you want every Stop hook event:

```yaml
session:
  skipUnindexed: false
```
