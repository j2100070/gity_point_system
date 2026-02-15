package presenter

import (
	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/usecases/inputport"
)

// AuthPresenter は認証関連のPresenter
type AuthPresenter struct{}

// NewAuthPresenter は新しいAuthPresenterを作成
func NewAuthPresenter() *AuthPresenter {
	return &AuthPresenter{}
}

// PresentRegisterResponse はRegisterResponseをJSON形式に変換
func (p *AuthPresenter) PresentRegisterResponse(resp *inputport.RegisterResponse) gin.H {
	return gin.H{
		"message": "registration successful",
		"user": gin.H{
			"id":           resp.User.ID,
			"username":     resp.User.Username,
			"email":        resp.User.Email,
			"display_name": resp.User.DisplayName,
			"first_name":   resp.User.FirstName,
			"last_name":    resp.User.LastName,
			"avatar_url":   resp.User.AvatarURL,
			"balance":      resp.User.Balance,
			"role":         resp.User.Role,
		},
		"csrf_token": resp.Session.CSRFToken,
	}
}

// PresentLoginResponse はLoginResponseをJSON形式に変換
func (p *AuthPresenter) PresentLoginResponse(resp *inputport.LoginResponse) gin.H {
	return gin.H{
		"message": "login successful",
		"user": gin.H{
			"id":           resp.User.ID,
			"username":     resp.User.Username,
			"display_name": resp.User.DisplayName,
			"first_name":   resp.User.FirstName,
			"last_name":    resp.User.LastName,
			"avatar_url":   resp.User.AvatarURL,
			"balance":      resp.User.Balance,
			"role":         resp.User.Role,
		},
		"csrf_token": resp.Session.CSRFToken,
	}
}

// PresentCurrentUserResponse はCurrentUserResponseをJSON形式に変換
func (p *AuthPresenter) PresentCurrentUserResponse(resp *inputport.GetCurrentUserResponse) gin.H {
	return gin.H{
		"user": gin.H{
			"id":           resp.User.ID,
			"username":     resp.User.Username,
			"email":        resp.User.Email,
			"display_name": resp.User.DisplayName,
			"first_name":   resp.User.FirstName,
			"last_name":    resp.User.LastName,
			"avatar_url":   resp.User.AvatarURL,
			"balance":      resp.User.Balance,
			"role":         resp.User.Role,
			"is_active":    resp.User.IsActive,
			"created_at":   resp.User.CreatedAt,
		},
	}
}
