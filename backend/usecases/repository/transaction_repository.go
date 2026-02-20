package repository

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// TransactionRepository はトランザクションのリポジトリインターフェース
type TransactionRepository interface {
	// Create は新しいトランザクションを作成
	Create(ctx context.Context, transaction *entities.Transaction) error

	// Read はIDでトランザクションを検索
	Read(ctx context.Context, id uuid.UUID) (*entities.Transaction, error)

	// ReadByIdempotencyKey は冪等性キーでトランザクションを検索
	ReadByIdempotencyKey(ctx context.Context, key string) (*entities.Transaction, error)

	// ReadListByUserID はユーザーに関連するトランザクション一覧を取得
	ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Transaction, error)

	// ReadListAll は全トランザクション一覧を取得（管理者用）
	ReadListAll(ctx context.Context, offset, limit int) ([]*entities.Transaction, error)

	// ReadListAllWithFilter はフィルタ・ソート付きで全トランザクション一覧を取得
	ReadListAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo, sortBy, sortOrder string, offset, limit int) ([]*entities.Transaction, error)

	// CountAll は全トランザクション総数を取得
	CountAll(ctx context.Context) (int64, error)

	// CountAllWithFilter はフィルタ付きで全トランザクション総数を取得
	CountAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo string) (int64, error)

	// Update はトランザクションを更新
	Update(ctx context.Context, transaction *entities.Transaction) error

	// CountByUserID はユーザーのトランザクション総数を取得
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// ReadListByUserIDWithUsers はユーザーに関連するトランザクション一覧をユーザー情報付きで取得（JOIN）
	ReadListByUserIDWithUsers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.TransactionWithUsers, error)

	// ReadListAllWithFilterAndUsers はフィルタ・ソート付きで全トランザクション一覧をユーザー情報付きで取得（JOIN）
	ReadListAllWithFilterAndUsers(ctx context.Context, transactionType, dateFrom, dateTo, sortBy, sortOrder string, offset, limit int) ([]*entities.TransactionWithUsers, error)
}

// IdempotencyKeyRepository は冪等性キーのリポジトリインターフェース
type IdempotencyKeyRepository interface {
	// Create は新しい冪等性キーを作成（既存の場合はエラー）
	Create(ctx context.Context, key *entities.IdempotencyKey) error

	// ReadByKey はキーで冪等性キーを検索
	ReadByKey(ctx context.Context, key string) (*entities.IdempotencyKey, error)

	// Update は冪等性キーを更新
	Update(ctx context.Context, key *entities.IdempotencyKey) error

	// DeleteExpired は期限切れの冪等性キーを削除
	DeleteExpired(ctx context.Context) error
}
