package entities

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

// GenerateSecureTokenBase64 はbase64エンコードされた安全なランダムトークンを生成（セッション用）
func GenerateSecureTokenBase64(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateSecureTokenHex は16進数エンコードされた安全なランダムトークンを生成（メール認証用）
func GenerateSecureTokenHex(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
