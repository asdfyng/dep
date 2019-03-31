package crypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

func HmacSha1(message, secret string) string {
	hash := hmac.New(sha1.New, []byte(secret))
	hash.Write([]byte(message))

	// to lowercase hexits
	return hex.EncodeToString(hash.Sum(nil))
}
