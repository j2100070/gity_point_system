package dsmysql

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// ProductDataSource は商品のデータソースインターフェース
type ProductDataSource interface {
	// Insert は新しい商品を挿入
	Insert(product *entities.Product) error

	// Select はIDで商品を検索
	Select(id uuid.UUID) (*entities.Product, error)

	// Update は商品情報を更新
	Update(product *entities.Product) error

	// Delete は商品を論理削除
	Delete(id uuid.UUID) error

	// SelectList は商品一覧を取得
	SelectList(offset, limit int) ([]*entities.Product, error)

	// SelectListByCategory はカテゴリ別の商品一覧を取得
	SelectListByCategory(categoryCode string, offset, limit int) ([]*entities.Product, error)

	// SelectAvailableList は交換可能な商品一覧を取得
	SelectAvailableList(offset, limit int) ([]*entities.Product, error)

	// Count は商品総数を取得
	Count() (int64, error)

	// UpdateStock は在庫を更新
	UpdateStock(tx interface{}, productID uuid.UUID, quantity int) error
}

// ProductExchangeDataSource は商品交換のデータソースインターフェース
type ProductExchangeDataSource interface {
	// Insert は新しい交換を挿入
	Insert(tx interface{}, exchange *entities.ProductExchange) error

	// Select はIDで交換を検索
	Select(id uuid.UUID) (*entities.ProductExchange, error)

	// Update は交換情報を更新
	Update(tx interface{}, exchange *entities.ProductExchange) error

	// SelectListByUserID はユーザーの交換履歴を取得
	SelectListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.ProductExchange, error)

	// SelectListAll はすべての交換履歴を取得
	SelectListAll(offset, limit int) ([]*entities.ProductExchange, error)

	// CountByUserID はユーザーの交換総数を取得
	CountByUserID(userID uuid.UUID) (int64, error)

	// CountAll は全体の交換総数を取得
	CountAll() (int64, error)
}
