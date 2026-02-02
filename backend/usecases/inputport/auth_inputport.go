package inputport

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// AuthInputPort は認証のユースケースインターフェース
type AuthInputPort interface {
	// Register は新しいユーザーを登録
	Register(req *RegisterRequest) (*RegisterResponse, error)

	// Login はログイン処理
	Login(req *LoginRequest) (*LoginResponse, error)

	// Logout はログアウト処理
	Logout(req *LogoutRequest) error

	// GetCurrentUser は現在のユーザー情報を取得
	GetCurrentUser(req *GetCurrentUserRequest) (*GetCurrentUserResponse, error)

	// ValidateSession はセッションを検証
	ValidateSession(sessionToken string) (*entities.Session, error)
}

// RegisterRequest は登録リクエスト
type RegisterRequest struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
}

// RegisterResponse は登録レスポンス
type RegisterResponse struct {
	User    *entities.User
	Session *entities.Session
}

// LoginRequest はログインリクエスト
type LoginRequest struct {
	Username  string
	Password  string
	IPAddress string
	UserAgent string
}

// LoginResponse はログインレスポンス
type LoginResponse struct {
	User    *entities.User
	Session *entities.Session
}

// LogoutRequest はログアウトリクエスト
type LogoutRequest struct {
	UserID uuid.UUID
}

// GetCurrentUserRequest は現在のユーザー情報取得リクエスト
type GetCurrentUserRequest struct {
	UserID uuid.UUID
}

// GetCurrentUserResponse は現在のユーザー情報取得レスポンス
type GetCurrentUserResponse struct {
	User *entities.User
}
