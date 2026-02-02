package dsmysql

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// SessionDataSource はMySQLのセッションデータソースインターフェース
type SessionDataSource interface {
	// Insert は新しいセッションを挿入
	Insert(session *entities.Session) error

	// SelectByToken はトークンでセッションを検索
	SelectByToken(token string) (*entities.Session, error)

	// Update はセッションを更新
	Update(session *entities.Session) error

	// Delete はセッションを削除
	Delete(id uuid.UUID) error

	// DeleteByUserID はユーザーの全セッションを削除
	DeleteByUserID(userID uuid.UUID) error

	// DeleteExpired は期限切れセッションを削除
	DeleteExpired() error
}
