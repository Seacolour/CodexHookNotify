package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAttachmentDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notify-mail.yaml")
	data := []byte(`smtp:
  username: user@example.com
  password: secret
  from: user@example.com
  to: user@example.com
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Attachment.EnabledDefault() {
		t.Fatal("attachment should be enabled by default")
	}
	if cfg.Attachment.Mode != "when_truncated" {
		t.Fatalf("mode = %q, want when_truncated", cfg.Attachment.Mode)
	}
	if cfg.Attachment.FilenamePrefix != "codex-reply" {
		t.Fatalf("filenamePrefix = %q", cfg.Attachment.FilenamePrefix)
	}
	if cfg.Attachment.MaxBytes != 2*1024*1024 {
		t.Fatalf("maxBytes = %d", cfg.Attachment.MaxBytes)
	}
	if !cfg.Session.SkipUnindexedEnabled() {
		t.Fatal("skipUnindexed should be enabled by default")
	}
	if !cfg.Update.EnabledDefault() {
		t.Fatal("update checks should be enabled by default")
	}
	if cfg.Update.Repository != "Seacolour/CodexHookNotify" {
		t.Fatalf("repository = %q", cfg.Update.Repository)
	}
	if cfg.Update.IntervalHours != 24 {
		t.Fatalf("intervalHours = %d", cfg.Update.IntervalHours)
	}
	if cfg.Update.TimeoutSeconds != 5 {
		t.Fatalf("timeoutSeconds = %d", cfg.Update.TimeoutSeconds)
	}
}

func TestLoadRejectsInvalidAttachmentMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notify-mail.yaml")
	data := []byte(`smtp:
  username: user@example.com
  password: secret
  from: user@example.com
  to: user@example.com
attachment:
  mode: sometimes
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected invalid attachment mode error")
	}
}
