package middleware

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// maxRequestBodySize はリクエストボディの最大サイズ（1MB）
const maxRequestBodySize = 1 * 1024 * 1024

// htmlTagPattern はHTMLタグを検出する正規表現
var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// SecurityHeadersMiddleware はセキュリティヘッダーを設定
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

// InputSanitizationMiddleware は入力サニタイゼーションを行うミドルウェア
// クエリパラメータ、リクエストボディのサニタイゼーション、
// リクエストサイズ制限、Content-Typeバリデーションを実施する
func InputSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. リクエストボディサイズの制限
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodySize)
		}

		// 2. Content-Typeバリデーション（POST/PUT/PATCHリクエスト）
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !isAllowedContentType(contentType) {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "サポートされていないContent-Typeです",
				})
				return
			}
		}

		// 3. クエリパラメータのサニタイゼーション
		query := c.Request.URL.Query()
		for key, values := range query {
			for i, v := range values {
				query[key][i] = sanitizeString(v)
			}
		}
		c.Request.URL.RawQuery = query.Encode()

		// 4. JSONリクエストボディのサニタイゼーション
		if c.Request.Body != nil && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
					"error": "リクエストボディが大きすぎます",
				})
				return
			}
			c.Request.Body.Close()

			if len(body) > 0 {
				sanitizedBody := sanitizeJSONBody(body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(sanitizedBody))
				c.Request.ContentLength = int64(len(sanitizedBody))
			} else {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}

		c.Next()
	}
}

// sanitizeString は文字列のサニタイゼーションを行う
// - 前後の空白を除去
// - HTMLタグを除去
// - HTMLエンティティにエスケープ
// - NULLバイトを除去
func sanitizeString(s string) string {
	// 前後の空白を除去
	s = strings.TrimSpace(s)
	// NULLバイトを除去
	s = strings.ReplaceAll(s, "\x00", "")
	// HTMLタグを除去
	s = htmlTagPattern.ReplaceAllString(s, "")
	// HTMLエンティティにエスケープ
	s = html.EscapeString(s)
	return s
}

// sanitizeJSONBody はJSONボディ内の文字列値をサニタイゼーションする
func sanitizeJSONBody(body []byte) []byte {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body
	}

	sanitized := sanitizeJSONValue(data)
	result, err := json.Marshal(sanitized)
	if err != nil {
		return body
	}
	return result
}

// sanitizeJSONValue はJSON値を再帰的にサニタイゼーションする
func sanitizeJSONValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return sanitizeString(val)
	case map[string]interface{}:
		for k, v := range val {
			val[k] = sanitizeJSONValue(v)
		}
		return val
	case []interface{}:
		for i, v := range val {
			val[i] = sanitizeJSONValue(v)
		}
		return val
	default:
		return v
	}
}

// isAllowedContentType は許可されたContent-Typeかどうかを判定する
func isAllowedContentType(contentType string) bool {
	allowed := []string{
		"application/json",
		"multipart/form-data",
		"application/x-www-form-urlencoded",
	}
	ct := strings.ToLower(contentType)
	for _, a := range allowed {
		if strings.Contains(ct, a) {
			return true
		}
	}
	return false
}
