package dsmysql

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// UserDataSource はMySQLのユーザーデータソースインターフェース
// Repositoryが期待するDataSourceのインターフェース定義
type UserDataSource interface {
	// Insert は新しいユーザーを挿入
	Insert(user *entities.User) error

	// Select はIDでユーザーを検索
	Select(id uuid.UUID) (*entities.User, error)

	// SelectByUsername はユーザー名でユーザーを検索
	SelectByUsername(username string) (*entities.User, error)

	// SelectByEmail はメールアドレスでユーザーを検索
	SelectByEmail(email string) (*entities.User, error)

	// Update はユーザー情報を更新（楽観的ロック対応）
	Update(user *entities.User) (bool, error)

	// UpdateBalanceWithLock は残高を更新（悲観的ロック）
	UpdateBalanceWithLock(tx interface{}, userID uuid.UUID, amount int64, isDeduct bool) error

	// SelectList はユーザー一覧を取得
	SelectList(offset, limit int) ([]*entities.User, error)

	// Count はユーザー総数を取得
	Count() (int64, error)

	// Delete はユーザーを論理削除
	Delete(id uuid.UUID) error
}
