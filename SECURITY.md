# Security Policy

CodexHookNotify sends email through a local SMTP configuration file. Treat that file as secret material.

## Supported Versions

Security fixes target the latest released version.

## Reporting a Vulnerability

Please open a private security advisory on GitHub if the repository supports it. If not, contact the maintainer privately before opening a public issue.

## Local Secret Handling

- Do not commit `notify-mail.yaml`.
- Use mailbox-specific SMTP authorization codes where possible.
- Rotate the SMTP authorization code if it was pasted into chat, logs, screenshots, or issues.
- Review and trust Codex hooks only after confirming the command path.
