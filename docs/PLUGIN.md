# Codex Plugin Plan

This repository includes an experimental Codex plugin scaffold:

```text
.agents/plugins/marketplace.json
plugins/codex-hook-notify/
  .codex-plugin/plugin.json
  skills/codex-hook-notify/SKILL.md
  scripts/install.ps1
```

The plugin currently provides an installer skill rather than a fully bundled runtime hook.

## Why the Hook Still Uses `~/.codex/hooks.json`

Codex hooks execute local commands and must be reviewed and trusted by the user. Keeping the actual command in the normal Codex hook registry makes the security boundary visible:

1. Codex installs or updates the helper executable.
2. Codex writes or updates `~/.codex/hooks.json`.
3. The user restarts Codex Desktop.
4. The user reviews and trusts the Stop hook.

## Repo Marketplace Install Shape

Once the repository is public, users should be able to add this repository as a marketplace and install the plugin from Codex App. The plugin skill can then guide the normal installer.

The repository marketplace lives at:

```text
.agents/plugins/marketplace.json
```

## Future Plugin Improvements

- Publish signed Windows release binaries and let the plugin installer download them.
- Add checksums for release artifacts.
- Add a plugin app card with screenshots.
- Add non-email channels without changing the hook contract.
- Add a managed migration path if Codex exposes a stable public plugin hook bundle flow.
