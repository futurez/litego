package util

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func GetUUID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}
