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

	// SelectListAllWithFilter はフィルタ・ソート付きで全トランザクション一覧を取得
	SelectListAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo, sortBy, sortOrder string, offset, limit int) ([]*entities.Transaction, error)

	// CountAll は全トランザクション総数を取得
	CountAll(ctx context.Context) (int64, error)

	// CountAllWithFilter はフィルタ付きで全トランザクション総数を取得
	CountAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo string) (int64, error)

	// Update はトランザクションを更新
	Update(ctx context.Context, transaction *entities.Transaction) error

	// CountByUserID はユーザーのトランザクション総数を取得
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// SelectListByUserIDWithUsers はユーザーに関連するトランザクション一覧をユーザー情報付きで取得（JOIN）
	SelectListByUserIDWithUsers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.TransactionWithUsers, error)

	// SelectListAllWithFilterAndUsers はフィルタ・ソート付きで全トランザクション一覧をユーザー情報付きで取得（JOIN）
	SelectListAllWithFilterAndUsers(ctx context.Context, transactionType, dateFrom, dateTo, sortBy, sortOrder string, offset, limit int) ([]*entities.TransactionWithUsers, error)
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
