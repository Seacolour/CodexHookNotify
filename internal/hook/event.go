package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Seacolour/CodexHookNotify/internal/encodingutil"
)

type Event struct {
	HookEventName        string `json:"hook_event_name"`
	SessionID            string `json:"session_id"`
	TurnID               string `json:"turn_id"`
	CWD                  string `json:"cwd"`
	Model                string `json:"model"`
	LastAssistantMessage string `json:"last_assistant_message"`
	StopHookActive       bool   `json:"stop_hook_active"`
}

func ReadStdin(r io.Reader) (Event, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return Event{}, fmt.Errorf("read stdin: %w", err)
	}
	raw = encodingutil.DecodeStdin(raw)
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return Event{}, nil
	}
	var event Event
	if err := json.Unmarshal([]byte(text), &event); err != nil {
		return Event{}, fmt.Errorf("parse hook json: %w", err)
	}
	return event, nil
}

func (e Event) TruncateLast(max int) string {
	msg := strings.TrimSpace(e.LastAssistantMessage)
	if max <= 0 {
		return msg
	}
	runes := []rune(msg)
	if len(runes) <= max {
		return msg
	}
	return string(runes[:max]) + "..."
}
