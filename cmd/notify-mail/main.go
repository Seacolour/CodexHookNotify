package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Seacolour/CodexHookNotify/internal/config"
	"github.com/Seacolour/CodexHookNotify/internal/dedup"
	"github.com/Seacolour/CodexHookNotify/internal/hook"
	"github.com/Seacolour/CodexHookNotify/internal/mail"
	"github.com/Seacolour/CodexHookNotify/internal/sessionindex"
	"github.com/Seacolour/CodexHookNotify/internal/updatecheck"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	configPath := flag.String("config", "", "path to notify-mail.yaml")
	dryRun := flag.Bool("dry-run", false, "build email body but do not send")
	showVersion := flag.Bool("version", false, "print notify-mail version")
	testJSON := flag.String("test-json", "", "hook JSON string for manual testing (UTF-8, avoids PowerShell pipe encoding)")
	testJSONFile := flag.String("test-json-file", "", "path to UTF-8 JSON file for manual testing")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return 0
	}

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
	if (cfg.Session.TitleLookupEnabled() || cfg.Session.SkipUnindexedEnabled()) && event.SessionID != "" {
		sessionResult, err := sessionindex.Lookup(cfg.Session.IndexPath, event.SessionID)
		if err != nil {
			logger.error("session title lookup: %v", err)
		} else {
			if cfg.Session.SkipUnindexedEnabled() && sessionResult.IndexExists && !sessionResult.Found {
				logger.info("skipped unindexed session id=%s index=%s", event.SessionID, cfg.Session.IndexPath)
				if *dryRun {
					fmt.Printf("skipped: session %s was not found in %s\n", event.SessionID, cfg.Session.IndexPath)
				}
				return 0
			}
			if cfg.Session.TitleLookupEnabled() && sessionResult.Found {
				sessionTitle = sessionResult.Title
			}
		}
	}

	lastFull := strings.TrimSpace(event.LastAssistantMessage)
	lastPreview, previewTruncated := truncateTextWithState(lastFull, cfg.Mail.MaxMessageLength)
	if lastPreview == "" && !cfg.Mail.IncludeEmptyReply {
		lastPreview = "(无文本回复)"
	}
	attachments, attachmentNote := buildAttachments(event, cfg, sessionTitle, lastFull, previewTruncated, time.Now())
	updateNote := ""
	if cfg.Update.EnabledDefault() {
		notice, updateErr := updatecheck.Check(cfg.Update, resolveUpdateStatePath(cfgPath, cfg.Update.StatePath), version, time.Now())
		if updateErr != nil {
			logger.error("update check: %v", updateErr)
		} else if notice.Available {
			updateNote = fmt.Sprintf("工具更新：%s -> %s\r\n%s", version, notice.LatestVersion, notice.URL)
		}
	}

	body := buildBody(event, cfg, sessionTitle, lastPreview, attachmentNote, updateNote)
	if strings.TrimSpace(body) == "" && !cfg.Mail.IncludeEmptyReply {
		logger.info("empty body, skip send")
		return 0
	}

	if *dryRun {
		logger.info("dry-run subject=%q body_len=%d attachments=%d", cfg.Mail.Subject, len(body), len(attachments))
		fmt.Println(body)
		for _, attachment := range attachments {
			fmt.Printf("\n[attachment] %s (%d bytes)\n", attachment.Filename, len(attachment.Content))
		}
		return 0
	}

	if err := mail.Send(cfg, cfg.Mail.Subject, body, attachments...); err != nil {
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

func resolveUpdateStatePath(cfgPath, statePath string) string {
	if strings.TrimSpace(statePath) == "" {
		return filepath.Join(filepath.Dir(cfgPath), "notify-mail.update.json")
	}
	if filepath.IsAbs(statePath) {
		return statePath
	}
	return filepath.Join(filepath.Dir(cfgPath), statePath)
}

func buildBody(event hook.Event, cfg config.Config, sessionTitle, lastPreview, attachmentNote, updateNote string) string {
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
	if attachmentNote != "" {
		lines = append(lines, fmt.Sprintf("完整回复：%s", attachmentNote))
	}
	if updateNote != "" {
		lines = append(lines, updateNote)
	}
	lines = append(lines, "", "最后回复：", lastPreview)
	return strings.Join(lines, "\r\n")
}

func truncateText(value string, max int) string {
	truncated, _ := truncateTextWithState(value, max)
	return truncated
}

func truncateTextWithState(value string, max int) (string, bool) {
	if max <= 0 {
		return value, false
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value, false
	}
	return string(runes[:max]) + "...", true
}

func buildAttachments(event hook.Event, cfg config.Config, sessionTitle, lastFull string, previewTruncated bool, now time.Time) ([]mail.Attachment, string) {
	if strings.TrimSpace(lastFull) == "" || !cfg.Attachment.ShouldAttach(previewTruncated) {
		return nil, ""
	}

	filename := markdownAttachmentFilename(cfg.Attachment.FilenamePrefix, event, now)
	content := buildMarkdownAttachment(event, sessionTitle, lastFull)
	content, contentTruncated := limitUTF8Bytes(content, cfg.Attachment.MaxBytes)

	note := fmt.Sprintf("已附加 Markdown 文件 %s", filename)
	if contentTruncated {
		note += "（超过附件大小上限，内容已截断）"
	}

	return []mail.Attachment{
		{
			Filename:    filename,
			ContentType: "text/markdown",
			Content:     []byte(content),
		},
	}, note
}

func buildMarkdownAttachment(event hook.Event, sessionTitle, lastFull string) string {
	lines := []string{
		"# Codex 任务完成",
		"",
		"## Metadata",
		"",
		fmt.Sprintf("- 事件：%s", fallback(event.HookEventName, "Stop")),
		fmt.Sprintf("- 目录：`%s`", event.CWD),
		fmt.Sprintf("- 模型：%s", event.Model),
	}
	if strings.TrimSpace(sessionTitle) != "" {
		lines = append(lines, fmt.Sprintf("- 会话标题：%s", strings.TrimSpace(sessionTitle)))
	}
	if event.SessionID != "" {
		lines = append(lines, fmt.Sprintf("- 会话：`%s`", event.SessionID))
	}
	if event.TurnID != "" {
		lines = append(lines, fmt.Sprintf("- 轮次：`%s`", event.TurnID))
	}
	lines = append(lines, "", "## 最后回复", "", lastFull)
	return strings.Join(lines, "\n")
}

func markdownAttachmentFilename(prefix string, event hook.Event, now time.Time) string {
	prefix = sanitizeFilenamePart(prefix)
	if prefix == "" {
		prefix = "codex-reply"
	}
	parts := []string{prefix, now.Format("20060102-150405")}
	if shortID := shortSessionID(event.SessionID); shortID != "" {
		parts = append(parts, shortID)
	}
	return strings.Join(parts, "-") + ".md"
}

func shortSessionID(sessionID string) string {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return ""
	}
	if idx := strings.Index(sessionID, "-"); idx > 0 {
		return sessionID[:idx]
	}
	if len(sessionID) > 8 {
		return sessionID[:8]
	}
	return sessionID
}

func sanitizeFilenamePart(value string) string {
	value = strings.TrimSpace(value)
	var out strings.Builder
	lastDash := false
	for _, r := range value {
		allowed := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.'
		if allowed {
			out.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			out.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(out.String(), "-.")
}

func limitUTF8Bytes(value string, maxBytes int) (string, bool) {
	if maxBytes <= 0 || len([]byte(value)) <= maxBytes {
		return value, false
	}
	suffix := "\n\n---\n\n_附件内容超过配置的 maxBytes=" + strconv.Itoa(maxBytes) + "，已截断。请在 Codex Desktop 中查看完整回复。_\n"
	limit := maxBytes - len([]byte(suffix))
	if limit <= 0 {
		limit = maxBytes
		suffix = ""
	}
	var out strings.Builder
	for _, r := range value {
		next := string(r)
		if out.Len()+len([]byte(next)) > limit {
			break
		}
		out.WriteString(next)
	}
	out.WriteString(suffix)
	return out.String(), true
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
