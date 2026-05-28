package main

import (
	"strings"
	"testing"
	"time"

	"github.com/Seacolour/CodexHookNotify/internal/config"
	"github.com/Seacolour/CodexHookNotify/internal/hook"
)

func TestBuildAttachmentsWhenPreviewTruncated(t *testing.T) {
	cfg := config.Config{}
	cfg.Attachment.Mode = "when_truncated"
	cfg.Attachment.FilenamePrefix = "codex-reply"
	cfg.Attachment.MaxBytes = 2 * 1024 * 1024

	event := hook.Event{
		HookEventName:        "Stop",
		SessionID:            "019e68f4-7843-70d1-9183-70fcb41a8b5e",
		TurnID:               "turn-a",
		CWD:                  "D:\\Code\\CodexHookNotify",
		Model:                "gpt-5.5",
		LastAssistantMessage: strings.Repeat("hello ", 200),
	}

	attachments, note := buildAttachments(event, cfg, "Session title", event.LastAssistantMessage, true, time.Date(2026, 5, 28, 15, 1, 2, 0, time.UTC))
	if len(attachments) != 1 {
		t.Fatalf("attachments = %d, want 1", len(attachments))
	}
	if attachments[0].Filename != "codex-reply-20260528-150102-019e68f4.md" {
		t.Fatalf("filename = %q", attachments[0].Filename)
	}
	if !strings.Contains(string(attachments[0].Content), "## 最后回复") {
		t.Fatalf("attachment does not contain full reply markdown:\n%s", string(attachments[0].Content))
	}
	if !strings.Contains(note, attachments[0].Filename) {
		t.Fatalf("note %q does not mention filename %q", note, attachments[0].Filename)
	}
}

func TestBuildAttachmentsSkippedWhenNotTruncated(t *testing.T) {
	cfg := config.Config{}
	cfg.Attachment.Mode = "when_truncated"
	cfg.Attachment.MaxBytes = 2 * 1024 * 1024

	attachments, note := buildAttachments(hook.Event{LastAssistantMessage: "short"}, cfg, "", "short", false, time.Now())
	if len(attachments) != 0 || note != "" {
		t.Fatalf("attachments = %d note=%q, want none", len(attachments), note)
	}
}

func TestLimitUTF8Bytes(t *testing.T) {
	got, truncated := limitUTF8Bytes(strings.Repeat("你好", 100), 80)
	if !truncated {
		t.Fatal("expected truncation")
	}
	if len([]byte(got)) > 80 {
		t.Fatalf("limited content is %d bytes, want <= 80", len([]byte(got)))
	}
}
