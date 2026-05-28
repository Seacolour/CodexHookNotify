package mail

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestBuildTextMessage(t *testing.T) {
	msg := string(buildMessage("from@example.com", "to@example.com", "Codex 任务完成", "body", nil))

	if !strings.Contains(msg, "Content-Type: text/plain; charset=UTF-8") {
		t.Fatalf("text message missing plain text content type:\n%s", msg)
	}
	if strings.Contains(msg, "multipart/mixed") {
		t.Fatalf("text message should not be multipart:\n%s", msg)
	}
	if !strings.Contains(msg, "Subject: =?utf-8?") {
		t.Fatalf("non-ASCII subject should be MIME encoded:\n%s", msg)
	}
}

func TestBuildMessageWithMarkdownAttachment(t *testing.T) {
	content := []byte("# Full reply\n\nhello")
	msg := string(buildMessage(
		"from@example.com",
		"to@example.com",
		"Done",
		"summary",
		[]Attachment{
			{
				Filename:    "codex-reply.md",
				ContentType: "text/markdown",
				Content:     content,
			},
		},
	))

	for _, want := range []string{
		"Content-Type: multipart/mixed;",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Type: text/markdown; charset=UTF-8; name=\"codex-reply.md\"",
		"Content-Disposition: attachment; filename=\"codex-reply.md\"",
		"Content-Transfer-Encoding: base64",
		base64.StdEncoding.EncodeToString(content),
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("multipart message missing %q:\n%s", want, msg)
		}
	}
}
