package dsmysql

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// ArchivedUserDataSource はアーカイブユーザーのデータソースインターフェース
type ArchivedUserDataSource interface {
	// Insert は新しいアーカイブユーザーを挿入
	Insert(ctx context.Context, archivedUser *entities.ArchivedUser) error

	// Select はIDでアーカイブユーザーを検索
	Select(ctx context.Context, id uuid.UUID) (*entities.ArchivedUser, error)

	// SelectByUsername はユーザー名でアーカイブユーザーを検索
	SelectByUsername(ctx context.Context, username string) (*entities.ArchivedUser, error)

	// SelectList はアーカイブユーザー一覧を取得
	SelectList(ctx context.Context, offset, limit int) ([]*entities.ArchivedUser, error)

	// Count はアーカイブユーザー総数を取得
	Count(ctx context.Context) (int64, error)

	// Restore はアーカイブユーザーを復元
	Restore(ctx context.Context, tx interface{}, archivedUser *entities.ArchivedUser, user *entities.User) error
}

// EmailVerificationDataSource はメール認証トークンのデータソースインターフェース
type EmailVerificationDataSource interface {
	// Insert は新しいメール認証トークンを挿入
	Insert(ctx context.Context, token *entities.EmailVerificationToken) error

	// SelectByToken はトークンで検索
	SelectByToken(ctx context.Context, token string) (*entities.EmailVerificationToken, error)

	// Update はトークン情報を更新
	Update(ctx context.Context, token *entities.EmailVerificationToken) error

	// DeleteExpired は期限切れのトークンを削除
	DeleteExpired(ctx context.Context) error

	// DeleteByUserID はユーザーIDに紐づくトークンを削除
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// UsernameChangeHistoryDataSource はユーザー名変更履歴のデータソースインターフェース
type UsernameChangeHistoryDataSource interface {
	// Insert は新しいユーザー名変更履歴を挿入
	Insert(ctx context.Context, history *entities.UsernameChangeHistory) error

	// SelectListByUserID はユーザーIDで履歴を取得
	SelectListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.UsernameChangeHistory, error)

	// CountByUserID はユーザーIDで履歴数を取得
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

// PasswordChangeHistoryDataSource はパスワード変更履歴のデータソースインターフェース
type PasswordChangeHistoryDataSource interface {
	// Insert は新しいパスワード変更履歴を挿入
	Insert(ctx context.Context, history *entities.PasswordChangeHistory) error

	// SelectListByUserID はユーザーIDで履歴を取得
	SelectListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.PasswordChangeHistory, error)

	// CountByUserID はユーザーIDで履歴数を取得
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
