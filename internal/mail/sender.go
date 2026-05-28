package mail

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
	"unicode/utf8"

	"github.com/Seacolour/CodexHookNotify/internal/config"
)

func Send(cfg config.Config, subject, body string) error {
	from := cfg.SMTP.From
	to := cfg.SMTP.To
	msg := buildMessage(from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port)
	auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)

	switch strings.ToLower(cfg.SMTP.Mode) {
	case "tls":
		return sendImplicitTLS(addr, cfg.SMTP.Host, auth, from, []string{to}, msg)
	case "starttls":
		return smtp.SendMail(addr, auth, from, []string{to}, msg)
	default:
		return fmt.Errorf("unsupported smtp mode %q", cfg.SMTP.Mode)
	}
}

func buildMessage(from, to, subject, body string) []byte {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + encodeSubject(subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}
	return []byte(strings.Join(headers, "\r\n"))
}

func encodeSubject(subject string) string {
	if subject == "" || isASCII(subject) {
		return subject
	}
	return mime.QEncoding.Encode("utf-8", subject)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func sendImplicitTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
