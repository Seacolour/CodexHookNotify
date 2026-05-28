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
