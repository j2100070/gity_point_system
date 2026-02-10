package dsmysql

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// ArchivedUserDataSource はアーカイブユーザーのデータソースインターフェース
type ArchivedUserDataSource interface {
	// Insert は新しいアーカイブユーザーを挿入
	Insert(archivedUser *entities.ArchivedUser) error

	// Select はIDでアーカイブユーザーを検索
	Select(id uuid.UUID) (*entities.ArchivedUser, error)

	// SelectByUsername はユーザー名でアーカイブユーザーを検索
	SelectByUsername(username string) (*entities.ArchivedUser, error)

	// SelectList はアーカイブユーザー一覧を取得
	SelectList(offset, limit int) ([]*entities.ArchivedUser, error)

	// Count はアーカイブユーザー総数を取得
	Count() (int64, error)

	// Restore はアーカイブユーザーを復元
	Restore(tx interface{}, archivedUser *entities.ArchivedUser, user *entities.User) error
}

// EmailVerificationDataSource はメール認証トークンのデータソースインターフェース
type EmailVerificationDataSource interface {
	// Insert は新しいメール認証トークンを挿入
	Insert(token *entities.EmailVerificationToken) error

	// SelectByToken はトークンで検索
	SelectByToken(token string) (*entities.EmailVerificationToken, error)

	// Update はトークン情報を更新
	Update(token *entities.EmailVerificationToken) error

	// DeleteExpired は期限切れのトークンを削除
	DeleteExpired() error

	// DeleteByUserID はユーザーIDに紐づくトークンを削除
	DeleteByUserID(userID uuid.UUID) error
}

// UsernameChangeHistoryDataSource はユーザー名変更履歴のデータソースインターフェース
type UsernameChangeHistoryDataSource interface {
	// Insert は新しいユーザー名変更履歴を挿入
	Insert(history *entities.UsernameChangeHistory) error

	// SelectListByUserID はユーザーIDで履歴を取得
	SelectListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.UsernameChangeHistory, error)

	// CountByUserID はユーザーIDで履歴数を取得
	CountByUserID(userID uuid.UUID) (int64, error)
}

// PasswordChangeHistoryDataSource はパスワード変更履歴のデータソースインターフェース
type PasswordChangeHistoryDataSource interface {
	// Insert は新しいパスワード変更履歴を挿入
	Insert(history *entities.PasswordChangeHistory) error

	// SelectListByUserID はユーザーIDで履歴を取得
	SelectListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.PasswordChangeHistory, error)

	// CountByUserID はユーザーIDで履歴数を取得
	CountByUserID(userID uuid.UUID) (int64, error)
}
