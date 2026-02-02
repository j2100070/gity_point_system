package repository

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// SessionRepository はセッションのリポジトリインターフェース
type SessionRepository interface {
	// Create は新しいセッションを作成
	Create(session *entities.Session) error

	// ReadByToken はトークンでセッションを検索
	ReadByToken(token string) (*entities.Session, error)

	// Update はセッションを更新
	Update(session *entities.Session) error

	// Delete はセッションを削除
	Delete(id uuid.UUID) error

	// DeleteByUserID はユーザーの全セッションを削除（ログアウト）
	DeleteByUserID(userID uuid.UUID) error

	// DeleteExpired は期限切れセッションを削除
	DeleteExpired() error
}
