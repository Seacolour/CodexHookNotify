package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SMTP    SMTPConfig    `yaml:"smtp"`
	Mail    MailConfig    `yaml:"mail"`
	Session SessionConfig `yaml:"session"`
	Dedup   DedupConfig   `yaml:"dedup"`
	Log     LogConfig     `yaml:"log"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Mode     string `yaml:"mode"` // starttls | tls
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
}

type MailConfig struct {
	Subject           string `yaml:"subject"`
	MaxMessageLength  int    `yaml:"maxMessageLength"`
	IncludeEmptyReply bool   `yaml:"includeEmptyReply"`
}

type SessionConfig struct {
	TitleLookup    *bool  `yaml:"titleLookup"`
	IndexPath      string `yaml:"indexPath"`
	MaxTitleLength int    `yaml:"maxTitleLength"`
}

type DedupConfig struct {
	Enabled       bool `yaml:"enabled"`
	WindowSeconds int  `yaml:"windowSeconds"`
}

type LogConfig struct {
	Path string `yaml:"path"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(&cfg)
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.SMTP.Host == "" {
		cfg.SMTP.Host = "smtp.qq.com"
	}
	if cfg.SMTP.Port == 0 {
		cfg.SMTP.Port = 587
	}
	if cfg.SMTP.Mode == "" {
		if cfg.SMTP.Port == 465 {
			cfg.SMTP.Mode = "tls"
		} else {
			cfg.SMTP.Mode = "starttls"
		}
	}
	if cfg.Mail.Subject == "" {
		cfg.Mail.Subject = "Codex 任务完成"
	}
	if cfg.Mail.MaxMessageLength == 0 {
		cfg.Mail.MaxMessageLength = 800
	}
	if cfg.Session.IndexPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cfg.Session.IndexPath = filepath.Join(home, ".codex", "session_index.jsonl")
		}
	}
	if cfg.Session.MaxTitleLength == 0 {
		cfg.Session.MaxTitleLength = 80
	}
	if cfg.Dedup.WindowSeconds == 0 {
		cfg.Dedup.WindowSeconds = 30
	}
	if cfg.Log.Path == "" {
		cfg.Log.Path = "notify-mail.log"
	}
}

func (c Config) validate() error {
	if c.SMTP.Username == "" || c.SMTP.Password == "" {
		return fmt.Errorf("smtp.username and smtp.password are required")
	}
	if c.SMTP.From == "" {
		return fmt.Errorf("smtp.from is required")
	}
	if c.SMTP.To == "" {
		return fmt.Errorf("smtp.to is required")
	}
	mode := c.SMTP.Mode
	if mode != "starttls" && mode != "tls" {
		return fmt.Errorf("smtp.mode must be starttls or tls, got %q", mode)
	}
	return nil
}

func (s SessionConfig) TitleLookupEnabled() bool {
	return s.TitleLookup == nil || *s.TitleLookup
}
