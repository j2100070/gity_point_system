package dsmysql

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// BalanceUpdate は残高更新のパラメータ
type BalanceUpdate struct {
	UserID   uuid.UUID
	Amount   int64
	IsDeduct bool // true: 減算, false: 加算
}

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

	// UpdatePartial は指定されたフィールドのみを更新（楽観的ロックなし）
	UpdatePartial(userID uuid.UUID, fields map[string]interface{}) (bool, error)

	// UpdateBalanceWithLock は残高を更新（悲観的ロック）
	UpdateBalanceWithLock(tx interface{}, userID uuid.UUID, amount int64, isDeduct bool) error

	// UpdateBalancesWithLock は複数ユーザーの残高を一括更新（悲観的ロック、デッドロック回避）
	// 内部でID順にロックを取得することでデッドロックを回避します
	UpdateBalancesWithLock(tx interface{}, updates []BalanceUpdate) error

	// SelectList はユーザー一覧を取得
	SelectList(offset, limit int) ([]*entities.User, error)

	// Count はユーザー総数を取得
	Count() (int64, error)

	// Delete はユーザーを論理削除
	Delete(id uuid.UUID) error
}
