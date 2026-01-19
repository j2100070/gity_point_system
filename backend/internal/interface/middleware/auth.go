package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/internal/domain"
	"github.com/gity/point-system/internal/usecase"
	"github.com/google/uuid"
)

// AuthMiddleware は認証ミドルウェア
type AuthMiddleware struct {
	authUseCase *usecase.AuthUseCase
}

// NewAuthMiddleware は新しいAuthMiddlewareを作成
func NewAuthMiddleware(authUseCase *usecase.AuthUseCase) *AuthMiddleware {
	return &AuthMiddleware{
		authUseCase: authUseCase,
	}
}

// Authenticate は認証が必要なエンドポイント用ミドルウェア
// Cookie: session_token を検証
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cookieからセッショントークンを取得
		sessionToken, err := c.Cookie("session_token")
		if err != nil || sessionToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: no session token"})
			c.Abort()
			return
		}

		// セッション検証
		resp, err := m.authUseCase.ValidateSession(&usecase.ValidateSessionRequest{
			SessionToken: sessionToken,
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: " + err.Error()})
			c.Abort()
			return
		}

		// ユーザー情報をコンテキストに設定
		c.Set("user_id", resp.User.ID)
		c.Set("user", resp.User)
		c.Set("session", resp.Session)

		c.Next()
	}
}

// RequireAdmin は管理者権限が必要なエンドポイント用ミドルウェア
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 認証ミドルウェアが先に実行されている前提
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		user := userInterface.(*domain.User)
		if !user.IsAdmin() {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: admin role required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CSRFMiddleware はCSRF対策ミドルウェア
type CSRFMiddleware struct{}

// NewCSRFMiddleware は新しいCSRFMiddlewareを作成
func NewCSRFMiddleware() *CSRFMiddleware {
	return &CSRFMiddleware{}
}

// Protect はCSRF保護（POST/PUT/DELETE/PATCH）
// Double Submit Cookie パターン
// - Cookie: csrf_token（httponly=false, samesite=strict）
// - Header: X-CSRF-Token
func (m *CSRFMiddleware) Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET, HEAD, OPTIONS はCSRFチェック不要
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Cookieからトークン取得
		cookieToken, err := c.Cookie("csrf_token")
		if err != nil || cookieToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "csrf token missing in cookie"})
			c.Abort()
			return
		}

		// ヘッダーからトークン取得
		headerToken := c.GetHeader("X-CSRF-Token")
		if headerToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "csrf token missing in header"})
			c.Abort()
			return
		}

		// トークン比較
		if cookieToken != headerToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "csrf token mismatch"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware はセキュリティヘッダーを設定
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// XSS対策
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// HTTPS強制（本番環境）
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy（適宜調整）
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}

// RateLimitMiddleware はレート制限ミドルウェア（簡易実装）
// 本番環境ではRedisなどを使用した実装を推奨
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: レート制限実装
		// - IPアドレスベース
		// - ユーザーベース
		// - エンドポイントごとの制限
		c.Next()
	}
}

// InputSanitizationMiddleware は入力サニタイゼーション
// SQLインジェクション、XSS対策
func InputSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// NOTE: GORMのパラメータ化クエリでSQLインジェクションは防止済み
		// ReactのエスケープでXSSは防止済み
		// 追加の検証が必要な場合はここで実装

		// クエリパラメータのサニタイズ（例）
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				// 危険な文字列のチェック（例: SQL関連）
				if containsDangerousPatterns(value) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input detected in " + key})
					c.Abort()
					return
				}
				values[i] = strings.TrimSpace(value)
			}
		}

		c.Next()
	}
}

// containsDangerousPatterns は危険なパターンをチェック
func containsDangerousPatterns(input string) bool {
	// 簡易実装: より厳密なバリデーションは validator パッケージを使用
	dangerousPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// GetUserID はコンテキストからユーザーIDを取得
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}
	return userID.(uuid.UUID), nil
}

// GetUser はコンテキストからユーザーを取得
func GetUser(c *gin.Context) (*domain.User, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, errors.New("user not found in context")
	}
	return user.(*domain.User), nil
}
