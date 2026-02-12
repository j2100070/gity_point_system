package repository

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// UserSettingsRepository はユーザー設定のリポジトリインターフェース
type UserSettingsRepository interface {
	// UpdateProfile はプロフィール情報を更新（楽観的ロック対応）
	UpdateProfile(ctx context.Context, user *entities.User) (bool, error)

	// UpdateUsername はユーザー名を更新（一意性チェック付き）
	UpdateUsername(ctx context.Context, user *entities.User) (bool, error)

	// UpdatePassword はパスワードを更新
	UpdatePassword(ctx context.Context, user *entities.User) (bool, error)

	// CheckUsernameExists はユーザー名が既に存在するかチェック
	CheckUsernameExists(ctx context.Context, username string, excludeUserID uuid.UUID) (bool, error)

	// CheckEmailExists はメールアドレスが既に存在するかチェック
	CheckEmailExists(ctx context.Context, email string, excludeUserID uuid.UUID) (bool, error)
}

// ArchivedUserRepository はアーカイブユーザーのリポジトリインターフェース
type ArchivedUserRepository interface {
	// Create はアーカイブユーザーを作成
	Create(ctx context.Context, archivedUser *entities.ArchivedUser) error

	// Read はIDでアーカイブユーザーを検索
	Read(ctx context.Context, id uuid.UUID) (*entities.ArchivedUser, error)

	// ReadByUsername はユーザー名でアーカイブユーザーを検索
	ReadByUsername(ctx context.Context, username string) (*entities.ArchivedUser, error)

	// ReadList はアーカイブユーザー一覧を取得
	ReadList(ctx context.Context, offset, limit int) ([]*entities.ArchivedUser, error)

	// Count はアーカイブユーザー総数を取得
	Count(ctx context.Context) (int64, error)

	// Restore はアーカイブユーザーを復元（アーカイブから削除してユーザーに戻す）
	Restore(ctx context.Context, tx interface{}, archivedUser *entities.ArchivedUser, user *entities.User) error
}

// EmailVerificationRepository はメール認証トークンのリポジトリインターフェース
type EmailVerificationRepository interface {
	// Create は新しいメール認証トークンを作成
	Create(ctx context.Context, token *entities.EmailVerificationToken) error

	// ReadByToken はトークンで検索
	ReadByToken(ctx context.Context, token string) (*entities.EmailVerificationToken, error)

	// Update はトークン情報を更新
	Update(ctx context.Context, token *entities.EmailVerificationToken) error

	// DeleteExpired は期限切れのトークンを削除
	DeleteExpired(ctx context.Context) error

	// DeleteByUserID はユーザーIDに紐づくトークンを削除
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// UsernameChangeHistoryRepository はユーザー名変更履歴のリポジトリインターフェース
type UsernameChangeHistoryRepository interface {
	// Create は新しいユーザー名変更履歴を作成
	Create(ctx context.Context, history *entities.UsernameChangeHistory) error

	// ReadListByUserID はユーザーIDで履歴を取得
	ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.UsernameChangeHistory, error)

	// CountByUserID はユーザーIDで履歴数を取得
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

// PasswordChangeHistoryRepository はパスワード変更履歴のリポジトリインターフェース
type PasswordChangeHistoryRepository interface {
	// Create は新しいパスワード変更履歴を作成
	Create(ctx context.Context, history *entities.PasswordChangeHistory) error

	// ReadListByUserID はユーザーIDで履歴を取得
	ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.PasswordChangeHistory, error)

	// CountByUserID はユーザーIDで履歴数を取得
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
