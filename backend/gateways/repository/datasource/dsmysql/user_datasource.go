package dsmysql

import (
	"context"

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
	Insert(ctx context.Context, user *entities.User) error

	// Select はIDでユーザーを検索
	Select(ctx context.Context, id uuid.UUID) (*entities.User, error)

	// SelectByUsername はユーザー名でユーザーを検索
	SelectByUsername(ctx context.Context, username string) (*entities.User, error)

	// SelectByEmail はメールアドレスでユーザーを検索
	SelectByEmail(ctx context.Context, email string) (*entities.User, error)

	// Update はユーザー情報を更新（楽観的ロック対応）
	Update(ctx context.Context, user *entities.User) (bool, error)

	// UpdatePartial は指定されたフィールドのみを更新（楽観的ロックなし）
	UpdatePartial(ctx context.Context, userID uuid.UUID, fields map[string]interface{}) (bool, error)

	// UpdateBalanceWithLock は残高を更新（悲観的ロック）
	UpdateBalanceWithLock(ctx context.Context, userID uuid.UUID, amount int64, isDeduct bool) error

	// UpdateBalancesWithLock は複数ユーザーの残高を一括更新（悲観的ロック、デッドロック回避）
	// 内部でID順にロックを取得することでデッドロックを回避します
	UpdateBalancesWithLock(ctx context.Context, updates []BalanceUpdate) error

	// SelectList はユーザー一覧を取得
	SelectList(ctx context.Context, offset, limit int) ([]*entities.User, error)

	// SelectListWithSearch は検索・ソート付きでユーザー一覧を取得
	SelectListWithSearch(ctx context.Context, search string, sortBy string, sortOrder string, offset, limit int) ([]*entities.User, error)

	// Count はユーザー総数を取得
	Count(ctx context.Context) (int64, error)

	// CountWithSearch は検索条件付きでユーザー総数を取得
	CountWithSearch(ctx context.Context, search string) (int64, error)

	// Delete はユーザーを論理削除
	Delete(ctx context.Context, id uuid.UUID) error
}
