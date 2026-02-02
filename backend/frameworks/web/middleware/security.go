package middleware

import (
	"github.com/gin-gonic/gin"
)

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

// InputSanitizationMiddleware は入力サニタイゼーション
func InputSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 入力のサニタイゼーション処理
		// 実装は要件に応じて追加
		c.Next()
	}
}
