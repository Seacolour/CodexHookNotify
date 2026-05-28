package sessionindex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLookupTitle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session_index.jsonl")
	data := []byte(`{"id":"session-a","thread_name":"First title","updated_at":"2026/5/28 10:00:00"}` + "\n" +
		`{"id":"session-b","thread_name":"Second title","updated_at":"2026/5/28 10:01:00"}` + "\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LookupTitle(path, "session-b")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Second title" {
		t.Fatalf("title = %q, want %q", got, "Second title")
	}
}

func TestLookupTitleUsesLastMatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session_index.jsonl")
	data := []byte(`{"id":"session-a","thread_name":"Old title"}` + "\n" +
		`{"id":"session-a","thread_name":"New title"}` + "\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LookupTitle(path, "session-a")
	if err != nil {
		t.Fatal(err)
	}
	if got != "New title" {
		t.Fatalf("title = %q, want %q", got, "New title")
	}
}

func TestLookupTitleMissingFile(t *testing.T) {
	got, err := LookupTitle(filepath.Join(t.TempDir(), "missing.jsonl"), "session-a")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("title = %q, want empty", got)
	}
}
