package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/usecases/inputport"
)

// AuthMiddleware は認証ミドルウェア
type AuthMiddleware struct {
	authUC inputport.AuthInputPort
}

// NewAuthMiddleware は新しいAuthMiddlewareを作成
func NewAuthMiddleware(authUC inputport.AuthInputPort) *AuthMiddleware {
	return &AuthMiddleware{authUC: authUC}
}

// Authenticate は認証を行う
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// セッショントークンを取得
		sessionToken := c.GetHeader("Authorization")
		if sessionToken == "" {
			// Cookieからも取得を試みる
			sessionToken, _ = c.Cookie("session_token")
		}

		if sessionToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// セッション検証
		session, err := m.authUC.ValidateSession(c.Request.Context(), sessionToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		// ユーザーIDをコンテキストにセット
		c.Set("user_id", session.UserID)
		c.Set("session", session)

		c.Next()
	}
}
