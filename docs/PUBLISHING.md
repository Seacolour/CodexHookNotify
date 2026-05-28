# Publishing Checklist

## Before First Public Release

- GitHub repository: `Seacolour/CodexHookNotify`.
- Go module path: `github.com/Seacolour/CodexHookNotify`.
- Verify `notify-mail.yaml` is ignored and not committed.
- Run:

  ```powershell
  go test ./...
  .\build.ps1
  .\scripts\install.ps1 -DryRun
  ```

- Confirm screenshots under `docs/images/` contain no private email address, project path, session id, or authorization code.
- Create a GitHub release with a Windows executable artifact.

## Suggested Release Artifact Names

```text
notify-mail-windows-amd64.exe
checksums.txt
```

## Suggested Tags

```text
v0.1.0
```

## README Release Note

Mention that users must fill in SMTP credentials locally and trust the Codex Stop hook after restarting Codex Desktop.
