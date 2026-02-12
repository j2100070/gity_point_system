package repository

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// SessionRepository はセッションのリポジトリインターフェース
type SessionRepository interface {
	// Create は新しいセッションを作成
	Create(ctx context.Context, session *entities.Session) error

	// ReadByToken はトークンでセッションを検索
	ReadByToken(ctx context.Context, token string) (*entities.Session, error)

	// Update はセッションを更新
	Update(ctx context.Context, session *entities.Session) error

	// Delete はセッションを削除
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserID はユーザーの全セッションを削除（ログアウト）
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired は期限切れセッションを削除
	DeleteExpired(ctx context.Context) error
}
