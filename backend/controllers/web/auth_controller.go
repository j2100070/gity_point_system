package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// AuthController は認証関連のコントローラー
type AuthController struct {
	authUC    inputport.AuthInputPort
	presenter *presenter.AuthPresenter
}

// NewAuthController は新しいAuthControllerを作成
func NewAuthController(
	authUC inputport.AuthInputPort,
	presenter *presenter.AuthPresenter,
) *AuthController {
	return &AuthController{
		authUC:    authUC,
		presenter: presenter,
	}
}

// RegisterRequest は登録リクエスト
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

// Register は新しいユーザーを登録
// POST /api/auth/register
func (c *AuthController) Register(ctx *gin.Context, currentTime time.Time) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authUC.Register(&inputport.RegisterRequest{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// セッショントークンをCookieに設定
	ctx.SetCookie(
		"session_token",
		resp.Session.SessionToken,
		24*60*60, // 24時間
		"/",
		"",
		false, // HTTPS only in production
		true,  // HttpOnly
	)

	output := c.presenter.PresentRegisterResponse(resp)
	ctx.JSON(http.StatusCreated, output)
}

// LoginRequest はログインリクエスト
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login はログイン処理
// POST /api/auth/login
func (c *AuthController) Login(ctx *gin.Context, currentTime time.Time) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authUC.Login(&inputport.LoginRequest{
		Username:  req.Username,
		Password:  req.Password,
		IPAddress: ctx.ClientIP(),
		UserAgent: ctx.GetHeader("User-Agent"),
	})

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// セッショントークンをCookieに設定
	ctx.SetCookie(
		"session_token",
		resp.Session.SessionToken,
		24*60*60,
		"/",
		"",
		false,
		true,
	)

	output := c.presenter.PresentLoginResponse(resp)
	ctx.JSON(http.StatusOK, output)
}

// Logout はログアウト処理
// POST /api/auth/logout
func (c *AuthController) Logout(ctx *gin.Context, currentTime time.Time) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := c.authUC.Logout(&inputport.LogoutRequest{
		UserID: userID.(uuid.UUID),
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Cookieをクリア
	ctx.SetCookie("session_token", "", -1, "/", "", false, true)

	ctx.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// GetCurrentUser は現在のユーザー情報を取得
// GET /api/auth/me
func (c *AuthController) GetCurrentUser(ctx *gin.Context, currentTime time.Time) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := c.authUC.GetCurrentUser(&inputport.GetCurrentUserRequest{
		UserID: userID.(uuid.UUID),
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentCurrentUserResponse(resp)
	ctx.JSON(http.StatusOK, output)
}
