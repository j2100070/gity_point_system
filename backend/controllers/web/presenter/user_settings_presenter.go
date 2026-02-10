package presenter

import (
	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/usecases/inputport"
)

// UserSettingsPresenter はユーザー設定関連のPresenter
type UserSettingsPresenter struct{}

// NewUserSettingsPresenter は新しいUserSettingsPresenterを作成
func NewUserSettingsPresenter() *UserSettingsPresenter {
	return &UserSettingsPresenter{}
}

// PresentUpdateProfileResponse はUpdateProfileResponseをJSON形式に変換
func (p *UserSettingsPresenter) PresentUpdateProfileResponse(resp *inputport.UpdateProfileResponse) gin.H {
	result := gin.H{
		"message": "profile updated successfully",
		"user": gin.H{
			"id":                  resp.User.ID,
			"username":            resp.User.Username,
			"email":               resp.User.Email,
			"display_name":        resp.User.DisplayName,
			"avatar_url":          resp.User.AvatarURL,
			"email_verified":      resp.User.EmailVerified,
			"email_verified_at":   resp.User.EmailVerifiedAt,
		},
	}

	if resp.EmailVerificationSent {
		result["email_verification_sent"] = true
		result["message"] = "profile updated successfully. verification email sent to new email address"
	}

	return result
}

// PresentUploadAvatarResponse はUploadAvatarResponseをJSON形式に変換
func (p *UserSettingsPresenter) PresentUploadAvatarResponse(resp *inputport.UploadAvatarResponse) gin.H {
	return gin.H{
		"message":    "avatar uploaded successfully",
		"avatar_url": resp.AvatarURL,
	}
}

// PresentVerifyEmailResponse はVerifyEmailResponseをJSON形式に変換
func (p *UserSettingsPresenter) PresentVerifyEmailResponse(resp *inputport.VerifyEmailResponse) gin.H {
	return gin.H{
		"message": "email verified successfully",
		"user": gin.H{
			"id":                resp.User.ID,
			"username":          resp.User.Username,
			"email":             resp.User.Email,
			"display_name":      resp.User.DisplayName,
			"email_verified":    resp.User.EmailVerified,
			"email_verified_at": resp.User.EmailVerifiedAt,
		},
	}
}

// PresentGetProfileResponse はGetProfileResponseをJSON形式に変換
func (p *UserSettingsPresenter) PresentGetProfileResponse(resp *inputport.GetProfileResponse) gin.H {
	return gin.H{
		"user": gin.H{
			"id":                  resp.User.ID,
			"username":            resp.User.Username,
			"email":               resp.User.Email,
			"display_name":        resp.User.DisplayName,
			"avatar_url":          resp.User.AvatarURL,
			"email_verified":      resp.User.EmailVerified,
			"email_verified_at":   resp.User.EmailVerifiedAt,
			"balance":             resp.User.Balance,
			"role":                resp.User.Role,
			"created_at":          resp.User.CreatedAt,
		},
	}
}

// PresentSuccessMessage は成功メッセージをJSON形式に変換
func (p *UserSettingsPresenter) PresentSuccessMessage(message string) gin.H {
	return gin.H{
		"message": message,
	}
}
