package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/entities"
)

// CSRFMiddleware はCSRF保護ミドルウェア
type CSRFMiddleware struct{}

// NewCSRFMiddleware は新しいCSRFMiddlewareを作成
func NewCSRFMiddleware() *CSRFMiddleware {
	return &CSRFMiddleware{}
}

// Protect はCSRF保護を行う
func (m *CSRFMiddleware) Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET, HEAD, OPTIONS以外はCSRFトークンをチェック
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
			csrfToken := c.GetHeader("X-CSRF-Token")
			if csrfToken == "" {
				c.JSON(http.StatusForbidden, gin.H{"error": "csrf token required"})
				c.Abort()
				return
			}

			// セッションから取得
			sessionInterface, exists := c.Get("session")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				c.Abort()
				return
			}

			session := sessionInterface.(*entities.Session)

			// CSRFトークン検証
			if err := session.ValidateCSRF(csrfToken); err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "invalid csrf token"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
