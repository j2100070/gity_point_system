package repository

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// TransactionRepository はトランザクションのリポジトリインターフェース
type TransactionRepository interface {
	// Create は新しいトランザクションを作成
	Create(tx interface{}, transaction *entities.Transaction) error

	// Read はIDでトランザクションを検索
	Read(id uuid.UUID) (*entities.Transaction, error)

	// ReadByIdempotencyKey は冪等性キーでトランザクションを検索
	ReadByIdempotencyKey(key string) (*entities.Transaction, error)

	// ReadListByUserID はユーザーに関連するトランザクション一覧を取得
	ReadListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.Transaction, error)

	// ReadListAll は全トランザクション一覧を取得（管理者用）
	ReadListAll(offset, limit int) ([]*entities.Transaction, error)

	// Update はトランザクションを更新
	Update(tx interface{}, transaction *entities.Transaction) error

	// CountByUserID はユーザーのトランザクション総数を取得
	CountByUserID(userID uuid.UUID) (int64, error)
}

// IdempotencyKeyRepository は冪等性キーのリポジトリインターフェース
type IdempotencyKeyRepository interface {
	// Create は新しい冪等性キーを作成（既存の場合はエラー）
	Create(key *entities.IdempotencyKey) error

	// ReadByKey はキーで冪等性キーを検索
	ReadByKey(key string) (*entities.IdempotencyKey, error)

	// Update は冪等性キーを更新
	Update(key *entities.IdempotencyKey) error

	// DeleteExpired は期限切れの冪等性キーを削除
	DeleteExpired() error
}
