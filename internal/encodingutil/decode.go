package encodingutil

import (
	"bytes"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// DecodeStdin normalizes hook stdin bytes to UTF-8.
// Codex sends UTF-8; PowerShell pipe on Chinese Windows often uses GBK (CP936).
func DecodeStdin(raw []byte) []byte {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return raw
	}
	if utf8.Valid(raw) {
		return raw
	}
	if decoded, err := gbkToUTF8(raw); err == nil && utf8.Valid(decoded) {
		return decoded
	}
	return raw
}

func gbkToUTF8(raw []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(raw), simplifiedchinese.GBK.NewDecoder())
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
