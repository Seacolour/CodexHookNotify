# Contributing

Thanks for helping improve CodexHookNotify.

## Development

```powershell
go test ./...
.\build.ps1
```

## Local Install Test

```powershell
.\scripts\install.ps1 -DryRun
```

Run without `-DryRun` only when you want to update your local Codex hook configuration.

## Security

Do not commit real `notify-mail.yaml`, SMTP passwords, app passwords, logs, or session ids.

## Pull Requests

Keep changes focused. For hook behavior changes, include the expected hook JSON shape and a manual test command.
