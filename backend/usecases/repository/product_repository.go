package repository

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// ProductRepository は商品のリポジトリインターフェース
type ProductRepository interface {
	// Create は新しい商品を作成
	Create(ctx context.Context, product *entities.Product) error

	// Read はIDで商品を検索
	Read(ctx context.Context, id uuid.UUID) (*entities.Product, error)

	// Update は商品情報を更新
	Update(ctx context.Context, product *entities.Product) error

	// Delete は商品を論理削除
	Delete(ctx context.Context, id uuid.UUID) error

	// ReadList は商品一覧を取得（ページネーション対応）
	ReadList(ctx context.Context, offset, limit int) ([]*entities.Product, error)

	// ReadListByCategory はカテゴリ別の商品一覧を取得
	ReadListByCategory(ctx context.Context, categoryCode string, offset, limit int) ([]*entities.Product, error)

	// ReadAvailableList は交換可能な商品一覧を取得
	ReadAvailableList(ctx context.Context, offset, limit int) ([]*entities.Product, error)

	// Count は商品総数を取得
	Count(ctx context.Context) (int64, error)

	// UpdateStock は在庫を更新（トランザクション対応）
	UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error
}

// ProductExchangeRepository は商品交換のリポジトリインターフェース
type ProductExchangeRepository interface {
	// Create は新しい交換を作成
	Create(ctx context.Context, exchange *entities.ProductExchange) error

	// Read はIDで交換を検索
	Read(ctx context.Context, id uuid.UUID) (*entities.ProductExchange, error)

	// Update は交換情報を更新
	Update(ctx context.Context, exchange *entities.ProductExchange) error

	// ReadListByUserID はユーザーの交換履歴を取得
	ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.ProductExchange, error)

	// ReadListAll はすべての交換履歴を取得（管理者用）
	ReadListAll(ctx context.Context, offset, limit int) ([]*entities.ProductExchange, error)

	// CountByUserID はユーザーの交換総数を取得
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// CountAll は全体の交換総数を取得
	CountAll(ctx context.Context) (int64, error)
}
