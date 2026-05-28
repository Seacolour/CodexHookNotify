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

type Result struct {
	IndexExists bool
	Found       bool
	Title       string
}

func Lookup(path, sessionID string) (Result, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || strings.TrimSpace(path) == "" {
		return Result{}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Result{}, nil
		}
		return Result{}, fmt.Errorf("open session index: %w", err)
	}
	defer file.Close()

	result := Result{IndexExists: true}
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
			result.Found = true
			result.Title = strings.TrimSpace(item.ThreadName)
		}
	}
	if err := scanner.Err(); err != nil {
		return Result{}, fmt.Errorf("read session index: %w", err)
	}
	return result, nil
}

func LookupTitle(path, sessionID string) (string, error) {
	result, err := Lookup(path, sessionID)
	if err != nil {
		return "", err
	}
	return result.Title, nil
}
