package repository

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// ProductRepository は商品のリポジトリインターフェース
type ProductRepository interface {
	// Create は新しい商品を作成
	Create(product *entities.Product) error

	// Read はIDで商品を検索
	Read(id uuid.UUID) (*entities.Product, error)

	// Update は商品情報を更新
	Update(product *entities.Product) error

	// Delete は商品を論理削除
	Delete(id uuid.UUID) error

	// ReadList は商品一覧を取得（ページネーション対応）
	ReadList(offset, limit int) ([]*entities.Product, error)

	// ReadListByCategory はカテゴリ別の商品一覧を取得
	ReadListByCategory(category entities.ProductCategory, offset, limit int) ([]*entities.Product, error)

	// ReadAvailableList は交換可能な商品一覧を取得
	ReadAvailableList(offset, limit int) ([]*entities.Product, error)

	// Count は商品総数を取得
	Count() (int64, error)

	// UpdateStock は在庫を更新（トランザクション対応）
	UpdateStock(tx interface{}, productID uuid.UUID, quantity int) error
}

// ProductExchangeRepository は商品交換のリポジトリインターフェース
type ProductExchangeRepository interface {
	// Create は新しい交換を作成
	Create(tx interface{}, exchange *entities.ProductExchange) error

	// Read はIDで交換を検索
	Read(id uuid.UUID) (*entities.ProductExchange, error)

	// Update は交換情報を更新
	Update(tx interface{}, exchange *entities.ProductExchange) error

	// ReadListByUserID はユーザーの交換履歴を取得
	ReadListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.ProductExchange, error)

	// ReadListAll はすべての交換履歴を取得（管理者用）
	ReadListAll(offset, limit int) ([]*entities.ProductExchange, error)

	// CountByUserID はユーザーの交換総数を取得
	CountByUserID(userID uuid.UUID) (int64, error)

	// CountAll は全体の交換総数を取得
	CountAll() (int64, error)
}
