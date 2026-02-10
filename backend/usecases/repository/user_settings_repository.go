package repository

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// UserSettingsRepository はユーザー設定のリポジトリインターフェース
type UserSettingsRepository interface {
	// UpdateProfile はプロフィール情報を更新（楽観的ロック対応）
	UpdateProfile(user *entities.User) (bool, error)

	// UpdateUsername はユーザー名を更新（一意性チェック付き）
	UpdateUsername(user *entities.User) (bool, error)

	// UpdatePassword はパスワードを更新
	UpdatePassword(user *entities.User) (bool, error)

	// CheckUsernameExists はユーザー名が既に存在するかチェック
	CheckUsernameExists(username string, excludeUserID uuid.UUID) (bool, error)

	// CheckEmailExists はメールアドレスが既に存在するかチェック
	CheckEmailExists(email string, excludeUserID uuid.UUID) (bool, error)
}

// ArchivedUserRepository はアーカイブユーザーのリポジトリインターフェース
type ArchivedUserRepository interface {
	// Create はアーカイブユーザーを作成
	Create(archivedUser *entities.ArchivedUser) error

	// Read はIDでアーカイブユーザーを検索
	Read(id uuid.UUID) (*entities.ArchivedUser, error)

	// ReadByUsername はユーザー名でアーカイブユーザーを検索
	ReadByUsername(username string) (*entities.ArchivedUser, error)

	// ReadList はアーカイブユーザー一覧を取得
	ReadList(offset, limit int) ([]*entities.ArchivedUser, error)

	// Count はアーカイブユーザー総数を取得
	Count() (int64, error)

	// Restore はアーカイブユーザーを復元（アーカイブから削除してユーザーに戻す）
	Restore(tx interface{}, archivedUser *entities.ArchivedUser, user *entities.User) error
}

// EmailVerificationRepository はメール認証トークンのリポジトリインターフェース
type EmailVerificationRepository interface {
	// Create は新しいメール認証トークンを作成
	Create(token *entities.EmailVerificationToken) error

	// ReadByToken はトークンで検索
	ReadByToken(token string) (*entities.EmailVerificationToken, error)

	// Update はトークン情報を更新
	Update(token *entities.EmailVerificationToken) error

	// DeleteExpired は期限切れのトークンを削除
	DeleteExpired() error

	// DeleteByUserID はユーザーIDに紐づくトークンを削除
	DeleteByUserID(userID uuid.UUID) error
}

// UsernameChangeHistoryRepository はユーザー名変更履歴のリポジトリインターフェース
type UsernameChangeHistoryRepository interface {
	// Create は新しいユーザー名変更履歴を作成
	Create(history *entities.UsernameChangeHistory) error

	// ReadListByUserID はユーザーIDで履歴を取得
	ReadListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.UsernameChangeHistory, error)

	// CountByUserID はユーザーIDで履歴数を取得
	CountByUserID(userID uuid.UUID) (int64, error)
}

// PasswordChangeHistoryRepository はパスワード変更履歴のリポジトリインターフェース
type PasswordChangeHistoryRepository interface {
	// Create は新しいパスワード変更履歴を作成
	Create(history *entities.PasswordChangeHistory) error

	// ReadListByUserID はユーザーIDで履歴を取得
	ReadListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.PasswordChangeHistory, error)

	// CountByUserID はユーザーIDで履歴数を取得
	CountByUserID(userID uuid.UUID) (int64, error)
}
