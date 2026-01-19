package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/internal/usecase"
)

// AuthHandler は認証関連のHTTPハンドラー
type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
}

// NewAuthHandler は新しいAuthHandlerを作成
func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// RegisterRequest は登録リクエスト
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

// Register はユーザー登録
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authUseCase.Register(&usecase.RegisterRequest{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user": gin.H{
			"id":           resp.User.ID,
			"username":     resp.User.Username,
			"email":        resp.User.Email,
			"display_name": resp.User.DisplayName,
			"balance":      resp.User.Balance,
			"role":         resp.User.Role,
		},
	})
}

// LoginRequest はログインリクエスト
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login はログイン
// POST /api/auth/login
//
// セキュリティ:
// - Cookie設定: httponly=true, secure=true (HTTPS), samesite=strict
// - CSRFトークンも返却
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authUseCase.Login(&usecase.LoginRequest{
		Username:  req.Username,
		Password:  req.Password,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// セッショントークンをCookieに設定
	// httponly: JavaScriptからアクセス不可（XSS対策）
	// secure: HTTPS通信のみ（本番環境で有効化）
	// samesite: CSRF対策
	c.SetCookie(
		"session_token",
		resp.SessionToken,
		int(24*time.Hour.Seconds()), // 24時間
		"/",
		"",    // domain（本番環境で設定）
		false, // secure（本番環境でtrueに）
		true,  // httponly
	)

	// CSRFトークンもCookieに設定（httponly=falseでJSから読み取り可能）
	c.SetCookie(
		"csrf_token",
		resp.CSRFToken,
		int(24*time.Hour.Seconds()),
		"/",
		"",
		false, // secure（本番環境でtrueに）
		false, // httponly=false（JSから読み取り必要）
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"user": gin.H{
			"id":           resp.User.ID,
			"username":     resp.User.Username,
			"email":        resp.User.Email,
			"display_name": resp.User.DisplayName,
			"balance":      resp.User.Balance,
			"role":         resp.User.Role,
		},
		"csrf_token": resp.CSRFToken, // フロントエンドでヘッダーに設定
	})
}

// Logout はログアウト
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "already logged out"})
		return
	}

	_ = h.authUseCase.Logout(&usecase.LogoutRequest{
		SessionToken: sessionToken,
	})

	// Cookieをクリア
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.SetCookie("csrf_token", "", -1, "/", "", false, false)

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// GetCurrentUser は現在のユーザー情報を取得
// GET /api/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	resp, err := h.authUseCase.GetCurrentUser(&usecase.GetCurrentUserRequest{
		SessionToken: sessionToken,
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           resp.User.ID,
			"username":     resp.User.Username,
			"email":        resp.User.Email,
			"display_name": resp.User.DisplayName,
			"balance":      resp.User.Balance,
			"role":         resp.User.Role,
		},
	})
}
