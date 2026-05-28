package dedup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type state struct {
	LastSentUnix int64 `json:"last_sent_unix"`
}

func ShouldSkip(statePath string, window time.Duration) (bool, error) {
	if window <= 0 {
		return false, nil
	}
	last, err := readLast(statePath)
	if err != nil {
		return false, err
	}
	if last.IsZero() {
		return false, nil
	}
	if time.Since(last) < window {
		return true, nil
	}
	return false, nil
}

func MarkSent(statePath string) error {
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return err
	}
	payload := state{LastSentUnix: time.Now().Unix()}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return os.WriteFile(statePath, data, 0o644)
}

func readLast(statePath string) (time.Time, error) {
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	var s state
	if err := json.Unmarshal(data, &s); err != nil {
		return time.Time{}, err
	}
	if s.LastSentUnix == 0 {
		return time.Time{}, nil
	}
	return time.Unix(s.LastSentUnix, 0), nil
}
