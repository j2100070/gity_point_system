package dsmysql

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// TransactionDataSource はMySQLのトランザクションデータソースインターフェース
type TransactionDataSource interface {
	// Insert は新しいトランザクションを挿入
	Insert(ctx context.Context, transaction *entities.Transaction) error

	// Select はIDでトランザクションを検索
	Select(ctx context.Context, id uuid.UUID) (*entities.Transaction, error)

	// SelectByIdempotencyKey は冪等性キーでトランザクションを検索
	SelectByIdempotencyKey(ctx context.Context, key string) (*entities.Transaction, error)

	// SelectListByUserID はユーザーに関連するトランザクション一覧を取得
	SelectListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Transaction, error)

	// SelectListAll は全トランザクション一覧を取得（管理者用）
	SelectListAll(ctx context.Context, offset, limit int) ([]*entities.Transaction, error)

	// Update はトランザクションを更新
	Update(ctx context.Context, transaction *entities.Transaction) error

	// CountByUserID はユーザーのトランザクション総数を取得
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

// IdempotencyKeyDataSource はMySQLの冪等性キーデータソースインターフェース
type IdempotencyKeyDataSource interface {
	// Insert は新しい冪等性キーを挿入（既存の場合はエラー）
	Insert(ctx context.Context, key *entities.IdempotencyKey) error

	// SelectByKey はキーで冪等性キーを検索
	SelectByKey(ctx context.Context, key string) (*entities.IdempotencyKey, error)

	// Update は冪等性キーを更新
	Update(ctx context.Context, key *entities.IdempotencyKey) error

	// DeleteExpired は期限切れの冪等性キーを削除
	DeleteExpired(ctx context.Context) error
}
