package updatecheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Seacolour/CodexHookNotify/internal/config"
)

const githubAPIBase = "https://api.github.com"

type Notice struct {
	Available     bool
	LatestVersion string
	URL           string
}

type state struct {
	LastCheckedUnix int64  `json:"last_checked_unix"`
	LatestVersion   string `json:"latest_version,omitempty"`
	LatestURL       string `json:"latest_url,omitempty"`
}

type release struct {
	TagName    string `json:"tag_name"`
	HTMLURL    string `json:"html_url"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

func Check(cfg config.UpdateConfig, statePath, currentVersion string, now time.Time) (Notice, error) {
	currentVersion = strings.TrimSpace(currentVersion)
	if currentVersion == "" || currentVersion == "dev" || currentVersion == "ci" {
		return Notice{}, nil
	}

	loaded, _ := readState(statePath)
	if !due(loaded.LastCheckedUnix, cfg.IntervalHours, now) {
		return noticeFromRelease(loaded.LatestVersion, loaded.LatestURL, currentVersion, cfg.SkippedVersions), nil
	}

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	latest, err := fetchLatest(ctx, http.DefaultClient, cfg)
	if err != nil {
		return noticeFromRelease(loaded.LatestVersion, loaded.LatestURL, currentVersion, cfg.SkippedVersions), err
	}

	next := state{
		LastCheckedUnix: now.Unix(),
		LatestVersion:   latest.TagName,
		LatestURL:       latest.HTMLURL,
	}
	if err := writeState(statePath, next); err != nil {
		return noticeFromRelease(latest.TagName, latest.HTMLURL, currentVersion, cfg.SkippedVersions), err
	}
	return noticeFromRelease(latest.TagName, latest.HTMLURL, currentVersion, cfg.SkippedVersions), nil
}

func fetchLatest(ctx context.Context, client *http.Client, cfg config.UpdateConfig) (release, error) {
	repo := strings.Trim(strings.TrimSpace(cfg.Repository), "/")
	if repo == "" {
		return release{}, errors.New("update.repository is empty")
	}
	if cfg.IncludePrerelease {
		return fetchLatestIncludingPrerelease(ctx, client, repo)
	}
	return fetchLatestStable(ctx, client, repo)
}

func fetchLatestStable(ctx context.Context, client *http.Client, repo string) (release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIBase+"/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "CodexHookNotify")

	resp, err := client.Do(req)
	if err != nil {
		return release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return release{}, fmt.Errorf("github latest release returned %s", resp.Status)
	}
	var latest release
	if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return release{}, err
	}
	return latest, nil
}

func fetchLatestIncludingPrerelease(ctx context.Context, client *http.Client, repo string) (release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIBase+"/repos/"+repo+"/releases?per_page=10", nil)
	if err != nil {
		return release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "CodexHookNotify")

	resp, err := client.Do(req)
	if err != nil {
		return release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return release{}, fmt.Errorf("github releases returned %s", resp.Status)
	}
	var releases []release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return release{}, err
	}
	for _, item := range releases {
		if !item.Draft {
			return item, nil
		}
	}
	return release{}, errors.New("no non-draft release found")
}

func noticeFromRelease(latestVersion, latestURL, currentVersion string, skipped []string) Notice {
	if latestVersion == "" || isSkipped(latestVersion, skipped) || !IsNewer(latestVersion, currentVersion) {
		return Notice{}
	}
	return Notice{Available: true, LatestVersion: latestVersion, URL: latestURL}
}

func IsNewer(latest, current string) bool {
	latestParts, latestOK := parseVersion(latest)
	currentParts, currentOK := parseVersion(current)
	if !latestOK || !currentOK {
		return false
	}
	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

func parseVersion(value string) ([3]int, bool) {
	var parts [3]int
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(strings.TrimPrefix(value, "v"), "V")
	if idx := strings.IndexAny(value, "-+"); idx >= 0 {
		value = value[:idx]
	}
	raw := strings.Split(value, ".")
	if len(raw) != 3 {
		return parts, false
	}
	for i, part := range raw {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return parts, false
		}
		parts[i] = n
	}
	return parts, true
}

func isSkipped(version string, skipped []string) bool {
	version = normalizeVersion(version)
	for _, item := range skipped {
		if normalizeVersion(item) == version {
			return true
		}
	}
	return false
}

func normalizeVersion(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.TrimPrefix(value, "v")
}

func due(lastCheckedUnix int64, intervalHours int, now time.Time) bool {
	if lastCheckedUnix <= 0 {
		return true
	}
	if intervalHours <= 0 {
		return true
	}
	last := time.Unix(lastCheckedUnix, 0)
	return now.Sub(last) >= time.Duration(intervalHours)*time.Hour
}

func readState(path string) (state, error) {
	if strings.TrimSpace(path) == "" {
		return state{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state{}, nil
		}
		return state{}, err
	}
	var value state
	if err := json.Unmarshal(data, &value); err != nil {
		return state{}, err
	}
	return value, nil
}

func writeState(path string, value state) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
