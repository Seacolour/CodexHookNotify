package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Seacolour/CodexHookNotify/internal/config"
	"github.com/Seacolour/CodexHookNotify/internal/dedup"
	"github.com/Seacolour/CodexHookNotify/internal/hook"
	"github.com/Seacolour/CodexHookNotify/internal/mail"
	"github.com/Seacolour/CodexHookNotify/internal/sessionindex"
)

func main() {
	os.Exit(run())
}

func run() int {
	configPath := flag.String("config", "", "path to notify-mail.yaml")
	dryRun := flag.Bool("dry-run", false, "build email body but do not send")
	testJSON := flag.String("test-json", "", "hook JSON string for manual testing (UTF-8, avoids PowerShell pipe encoding)")
	testJSONFile := flag.String("test-json-file", "", "path to UTF-8 JSON file for manual testing")
	flag.Parse()

	cfgPath := *configPath
	if cfgPath == "" {
		cfgPath = defaultConfigPath()
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "notify-mail: %v\n", err)
		return 1
	}

	logPath := resolveLogPath(cfgPath, cfg.Log.Path)
	logger := newLogger(logPath)

	var event hook.Event
	switch {
	case *testJSONFile != "":
		data, readErr := os.ReadFile(*testJSONFile)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "notify-mail: read test json file: %v\n", readErr)
			return 1
		}
		event, err = hook.ReadStdin(strings.NewReader(string(data)))
	case *testJSON != "":
		event, err = hook.ReadStdin(strings.NewReader(*testJSON))
	default:
		event, err = hook.ReadStdin(os.Stdin)
	}
	if err != nil {
		logger.error("parse event: %v", err)
		return 0
	}

	if cfg.Dedup.Enabled {
		statePath := filepath.Join(filepath.Dir(cfgPath), "notify-mail.state.json")
		window := time.Duration(cfg.Dedup.WindowSeconds) * time.Second
		skip, err := dedup.ShouldSkip(statePath, window)
		if err != nil {
			logger.error("dedup check: %v", err)
		} else if skip {
			logger.info("skipped duplicate notification within %s", window)
			return 0
		}
	}

	sessionTitle := ""
	if cfg.Session.TitleLookupEnabled() && event.SessionID != "" {
		sessionTitle, err = sessionindex.LookupTitle(cfg.Session.IndexPath, event.SessionID)
		if err != nil {
			logger.error("session title lookup: %v", err)
		}
	}

	body := buildBody(event, cfg, sessionTitle)
	if strings.TrimSpace(body) == "" && !cfg.Mail.IncludeEmptyReply {
		logger.info("empty body, skip send")
		return 0
	}

	if *dryRun {
		logger.info("dry-run subject=%q body_len=%d", cfg.Mail.Subject, len(body))
		fmt.Println(body)
		return 0
	}

	if err := mail.Send(cfg, cfg.Mail.Subject, body); err != nil {
		logger.error("send mail: %v", err)
		return 0
	}

	if cfg.Dedup.Enabled {
		statePath := filepath.Join(filepath.Dir(cfgPath), "notify-mail.state.json")
		if err := dedup.MarkSent(statePath); err != nil {
			logger.error("dedup mark: %v", err)
		}
	}

	logger.info("mail sent to %s", cfg.SMTP.To)
	return 0
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(home, ".codex", "hooks", "notify-mail.yaml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	exe, err := os.Executable()
	if err == nil {
		return filepath.Join(filepath.Dir(exe), "notify-mail.yaml")
	}
	return "notify-mail.yaml"
}

func resolveLogPath(cfgPath, logPath string) string {
	if filepath.IsAbs(logPath) {
		return logPath
	}
	return filepath.Join(filepath.Dir(cfgPath), logPath)
}

func buildBody(event hook.Event, cfg config.Config, sessionTitle string) string {
	last := event.TruncateLast(cfg.Mail.MaxMessageLength)
	if last == "" && !cfg.Mail.IncludeEmptyReply {
		last = "(无文本回复)"
	}
	sessionTitle = truncateText(strings.TrimSpace(sessionTitle), cfg.Session.MaxTitleLength)

	lines := []string{
		"Codex 任务已完成。",
		"",
		fmt.Sprintf("事件：%s", fallback(event.HookEventName, "Stop")),
		fmt.Sprintf("目录：%s", event.CWD),
		fmt.Sprintf("模型：%s", event.Model),
	}
	if sessionTitle != "" {
		lines = append(lines, fmt.Sprintf("会话标题：%s", sessionTitle))
	}
	if event.SessionID != "" {
		lines = append(lines, fmt.Sprintf("会话：%s", event.SessionID))
	}
	if event.TurnID != "" {
		lines = append(lines, fmt.Sprintf("轮次：%s", event.TurnID))
	}
	lines = append(lines, "", "最后回复：", last)
	return strings.Join(lines, "\r\n")
}

func truncateText(value string, max int) string {
	if max <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max]) + "..."
}

func fallback(value, def string) string {
	if strings.TrimSpace(value) == "" {
		return def
	}
	return value
}

type logger struct {
	path string
}

func newLogger(path string) *logger {
	return &logger{path: path}
}

func (l *logger) info(format string, args ...any) {
	l.write("INFO", format, args...)
}

func (l *logger) error(format string, args ...any) {
	l.write("ERROR", format, args...)
}

func (l *logger) write(level, format string, args ...any) {
	line := fmt.Sprintf("%s %s %s\n", time.Now().Format(time.RFC3339), level, fmt.Sprintf(format, args...))
	_ = os.MkdirAll(filepath.Dir(l.path), 0o755)
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}
