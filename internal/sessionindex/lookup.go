package sessionindex

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type entry struct {
	ID         string `json:"id"`
	ThreadName string `json:"thread_name"`
}

func LookupTitle(path, sessionID string) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || strings.TrimSpace(path) == "" {
		return "", nil
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("open session index: %w", err)
	}
	defer file.Close()

	var title string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item entry
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}
		if item.ID == sessionID {
			title = strings.TrimSpace(item.ThreadName)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read session index: %w", err)
	}
	return title, nil
}
