package dsmysql

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// SessionDataSource はMySQLのセッションデータソースインターフェース
type SessionDataSource interface {
	// Insert は新しいセッションを挿入
	Insert(ctx context.Context, session *entities.Session) error

	// SelectByToken はトークンでセッションを検索
	SelectByToken(ctx context.Context, token string) (*entities.Session, error)

	// Update はセッションを更新
	Update(ctx context.Context, session *entities.Session) error

	// Delete はセッションを削除
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserID はユーザーの全セッションを削除
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired は期限切れセッションを削除
	DeleteExpired(ctx context.Context) error
}
