package web

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// UserSettingsController はユーザー設定関連のコントローラー
type UserSettingsController struct {
	userSettingsUC inputport.UserSettingsInputPort
	presenter      *presenter.UserSettingsPresenter
}

// NewUserSettingsController は新しいUserSettingsControllerを作成
func NewUserSettingsController(
	userSettingsUC inputport.UserSettingsInputPort,
	presenter *presenter.UserSettingsPresenter,
) *UserSettingsController {
	return &UserSettingsController{
		userSettingsUC: userSettingsUC,
		presenter:      presenter,
	}
}

// UpdateProfileRequest はプロフィール更新リクエスト
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
	Email       string `json:"email" binding:"required,email"`
	FirstName   string `json:"first_name" binding:"required,max=100"`
	LastName    string `json:"last_name" binding:"required,max=100"`
}

// UpdateProfile はプロフィールを更新
// PUT /api/settings/profile
func (c *UserSettingsController) UpdateProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.userSettingsUC.UpdateProfile(ctx, &inputport.UpdateProfileRequest{
		UserID:      userID.(uuid.UUID),
		DisplayName: req.DisplayName,
		Email:       req.Email,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentUpdateProfileResponse(resp)
	ctx.JSON(http.StatusOK, output)
}

// UpdateUsernameRequest はユーザー名変更リクエスト
type UpdateUsernameRequest struct {
	NewUsername string `json:"new_username" binding:"required,min=3,max=50"`
}

// UpdateUsername はユーザー名を変更
// PUT /api/settings/username
func (c *UserSettingsController) UpdateUsername(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req UpdateUsernameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := ctx.ClientIP()
	err := c.userSettingsUC.UpdateUsername(ctx, &inputport.UpdateUsernameRequest{
		UserID:      userID.(uuid.UUID),
		NewUsername: req.NewUsername,
		IPAddress:   &ipAddress,
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentSuccessMessage("username updated successfully")
	ctx.JSON(http.StatusOK, output)
}

// ChangePasswordRequest はパスワード変更リクエスト
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword はパスワードを変更
// PUT /api/settings/password
func (c *UserSettingsController) ChangePassword(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")
	err := c.userSettingsUC.ChangePassword(ctx, &inputport.ChangePasswordRequest{
		UserID:          userID.(uuid.UUID),
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
		IPAddress:       &ipAddress,
		UserAgent:       &userAgent,
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentSuccessMessage("password changed successfully")
	ctx.JSON(http.StatusOK, output)
}

// UploadAvatar はアバターをアップロード
// POST /api/settings/avatar
func (c *UserSettingsController) UploadAvatar(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	file, header, err := ctx.Request.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}
	defer file.Close()

	// ファイルデータを読み込み
	fileData, err := io.ReadAll(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	// Content-Typeを取得
	contentType := header.Header.Get("Content-Type")

	resp, err := c.userSettingsUC.UploadAvatar(ctx, &inputport.UploadAvatarRequest{
		UserID:      userID.(uuid.UUID),
		FileData:    fileData,
		FileName:    header.Filename,
		ContentType: contentType,
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentUploadAvatarResponse(resp)
	ctx.JSON(http.StatusOK, output)
}

// DeleteAvatar はアバターを削除（自動生成に戻す）
// DELETE /api/settings/avatar
func (c *UserSettingsController) DeleteAvatar(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	err := c.userSettingsUC.DeleteAvatar(ctx, &inputport.DeleteAvatarRequest{
		UserID: userID.(uuid.UUID),
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentSuccessMessage("avatar deleted successfully")
	ctx.JSON(http.StatusOK, output)
}

// SendEmailVerification はメール認証メールを送信
// POST /api/settings/email/verify
func (c *UserSettingsController) SendEmailVerification(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// ユーザー情報を取得してメールアドレスを取得
	profile, err := c.userSettingsUC.GetProfile(ctx, &inputport.GetProfileRequest{
		UserID: userID.(uuid.UUID),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}

	uid := userID.(uuid.UUID)
	err = c.userSettingsUC.SendEmailVerification(ctx, &inputport.SendEmailVerificationRequest{
		UserID:    &uid,
		Email:     profile.User.Email,
		TokenType: "email_verification",
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentSuccessMessage("verification email sent successfully")
	ctx.JSON(http.StatusOK, output)
}

// VerifyEmailRequest はメール認証リクエスト
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// VerifyEmail はメールアドレスを認証
// POST /api/settings/email/verify/confirm
func (c *UserSettingsController) VerifyEmail(ctx *gin.Context) {
	var req VerifyEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.userSettingsUC.VerifyEmail(ctx, &inputport.VerifyEmailRequest{
		Token: req.Token,
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentVerifyEmailResponse(resp)
	ctx.JSON(http.StatusOK, output)
}

// ArchiveAccountRequest はアカウント削除リクエスト
type ArchiveAccountRequest struct {
	Password       string  `json:"password" binding:"required"`
	DeletionReason *string `json:"deletion_reason"`
}

// ArchiveAccount はアカウントを削除（アーカイブ）
// DELETE /api/settings/account
func (c *UserSettingsController) ArchiveAccount(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req ArchiveAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.userSettingsUC.ArchiveAccount(ctx, &inputport.ArchiveAccountRequest{
		UserID:         userID.(uuid.UUID),
		Password:       req.Password,
		DeletionReason: req.DeletionReason,
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// セッションをクリア
	ctx.SetCookie(
		"session_token",
		"",
		-1, // 即座に削除
		"/",
		"",
		false,
		true,
	)

	output := c.presenter.PresentSuccessMessage("account deleted successfully")
	ctx.JSON(http.StatusOK, output)
}

// GetProfile はプロフィール情報を取得
// GET /api/settings/profile
func (c *UserSettingsController) GetProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	resp, err := c.userSettingsUC.GetProfile(ctx, &inputport.GetProfileRequest{
		UserID: userID.(uuid.UUID),
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentGetProfileResponse(resp)
	ctx.JSON(http.StatusOK, output)
}
