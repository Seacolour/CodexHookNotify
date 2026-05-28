package mail

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Seacolour/CodexHookNotify/internal/config"
)

type Attachment struct {
	Filename    string
	ContentType string
	Content     []byte
}

func Send(cfg config.Config, subject, body string, attachments ...Attachment) error {
	from := cfg.SMTP.From
	to := cfg.SMTP.To
	msg := buildMessage(from, to, subject, body, attachments)

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

func buildMessage(from, to, subject, body string, attachments []Attachment) []byte {
	if len(attachments) == 0 {
		return buildTextMessage(from, to, subject, body)
	}

	boundary := fmt.Sprintf("codex-hook-notify-%d", time.Now().UnixNano())
	var msg bytes.Buffer
	writeHeaders(&msg, []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + encodeSubject(subject),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: multipart/mixed; boundary=%q", boundary),
	})
	fmt.Fprintf(&msg, "--%s\r\n", boundary)
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	msg.WriteString(body)
	msg.WriteString("\r\n")

	for _, attachment := range attachments {
		filename := sanitizeAttachmentFilename(attachment.Filename)
		contentType := strings.TrimSpace(attachment.ContentType)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		fmt.Fprintf(&msg, "--%s\r\n", boundary)
		fmt.Fprintf(&msg, "Content-Type: %s; charset=UTF-8; name=%q\r\n", contentType, filename)
		msg.WriteString("Content-Transfer-Encoding: base64\r\n")
		fmt.Fprintf(&msg, "Content-Disposition: attachment; filename=%q\r\n\r\n", filename)
		writeBase64Lines(&msg, attachment.Content)
		msg.WriteString("\r\n")
	}
	fmt.Fprintf(&msg, "--%s--\r\n", boundary)
	return msg.Bytes()
}

func buildTextMessage(from, to, subject, body string) []byte {
	var msg bytes.Buffer
	writeHeaders(&msg, []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + encodeSubject(subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	})
	msg.WriteString(body)
	return msg.Bytes()
}

func writeHeaders(msg *bytes.Buffer, headers []string) {
	msg.WriteString(strings.Join(headers, "\r\n"))
	msg.WriteString("\r\n\r\n")
}

func writeBase64Lines(msg *bytes.Buffer, content []byte) {
	encoded := base64.StdEncoding.EncodeToString(content)
	for len(encoded) > 76 {
		msg.WriteString(encoded[:76])
		msg.WriteString("\r\n")
		encoded = encoded[76:]
	}
	if encoded != "" {
		msg.WriteString(encoded)
		msg.WriteString("\r\n")
	}
}

func sanitizeAttachmentFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "attachment.md"
	}
	filename = strings.ReplaceAll(filename, "\r", "-")
	filename = strings.ReplaceAll(filename, "\n", "-")
	filename = strings.ReplaceAll(filename, `"`, "'")
	return filename
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
