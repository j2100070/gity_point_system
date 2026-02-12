package inputport

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// UserSettingsInputPort はユーザー設定のユースケースインターフェース
type UserSettingsInputPort interface {
	// UpdateProfile はプロフィールを更新
	UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*UpdateProfileResponse, error)

	// UpdateUsername はユーザー名を変更
	UpdateUsername(ctx context.Context, req *UpdateUsernameRequest) error

	// ChangePassword はパスワードを変更
	ChangePassword(ctx context.Context, req *ChangePasswordRequest) error

	// UploadAvatar はアバター画像をアップロード
	UploadAvatar(ctx context.Context, req *UploadAvatarRequest) (*UploadAvatarResponse, error)

	// DeleteAvatar はアバターを削除（自動生成に戻す）
	DeleteAvatar(ctx context.Context, req *DeleteAvatarRequest) error

	// SendEmailVerification はメール認証メールを送信
	SendEmailVerification(ctx context.Context, req *SendEmailVerificationRequest) error

	// VerifyEmail はメールアドレスを認証
	VerifyEmail(ctx context.Context, req *VerifyEmailRequest) (*VerifyEmailResponse, error)

	// ArchiveAccount はアカウントを削除（アーカイブ）
	ArchiveAccount(ctx context.Context, req *ArchiveAccountRequest) error

	// GetProfile はプロフィール情報を取得
	GetProfile(ctx context.Context, req *GetProfileRequest) (*GetProfileResponse, error)
}

// UpdateProfileRequest はプロフィール更新リクエスト
type UpdateProfileRequest struct {
	UserID      uuid.UUID
	DisplayName string
	Email       string
}

// UpdateProfileResponse はプロフィール更新レスポンス
type UpdateProfileResponse struct {
	User                  *entities.User
	EmailVerificationSent bool // メール変更時にtrueになる
}

// UpdateUsernameRequest はユーザー名変更リクエスト
type UpdateUsernameRequest struct {
	UserID      uuid.UUID
	NewUsername string
	IPAddress   *string
}

// ChangePasswordRequest はパスワード変更リクエスト
type ChangePasswordRequest struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewPassword     string
	IPAddress       *string
	UserAgent       *string
}

// UploadAvatarRequest はアバターアップロードリクエスト
type UploadAvatarRequest struct {
	UserID      uuid.UUID
	FileData    []byte
	FileName    string
	ContentType string
}

// UploadAvatarResponse はアバターアップロードレスポンス
type UploadAvatarResponse struct {
	AvatarURL string
}

// DeleteAvatarRequest はアバター削除リクエスト
type DeleteAvatarRequest struct {
	UserID uuid.UUID
}

// SendEmailVerificationRequest はメール認証送信リクエスト
type SendEmailVerificationRequest struct {
	UserID    *uuid.UUID           // 登録時はnil、メール変更時はユーザーID
	Email     string               // 認証するメールアドレス
	TokenType entities.TokenType   // "registration" | "email_change"
}

// VerifyEmailRequest はメール認証リクエスト
type VerifyEmailRequest struct {
	Token string
}

// VerifyEmailResponse はメール認証レスポンス
type VerifyEmailResponse struct {
	User  *entities.User
	Email string
}

// ArchiveAccountRequest はアカウント削除（アーカイブ）リクエスト
type ArchiveAccountRequest struct {
	UserID         uuid.UUID
	Password       string
	DeletionReason *string
}

// GetProfileRequest はプロフィール取得リクエスト
type GetProfileRequest struct {
	UserID uuid.UUID
}

// GetProfileResponse はプロフィール取得レスポンス
type GetProfileResponse struct {
	User *entities.User
}
