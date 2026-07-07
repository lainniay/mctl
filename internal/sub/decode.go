package sub

import (
	"encoding/base64"
	"strings"
)

func decodeBase64(body string) (string, bool) {
	data, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(body)
	}
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(data)), true
}

func DecodeBody(body string) string {
	body = strings.TrimSpace(body)
	if decoded, ok := decodeBase64(body); ok && strings.Contains(decoded, "://") {
		return decoded
	}
	return body
}
