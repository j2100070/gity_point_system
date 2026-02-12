package inputport

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// ==================== 商品管理（管理者用） ====================

// CreateProductRequest は商品作成リクエスト
type CreateProductRequest struct {
	Name        string
	Description string
	Category    string // カテゴリコード
	Price       int64
	Stock       int
	ImageURL    string
}

// CreateProductResponse は商品作成レスポンス
type CreateProductResponse struct {
	Product *entities.Product
}

// UpdateProductRequest は商品更新リクエスト
type UpdateProductRequest struct {
	ProductID   uuid.UUID
	Name        string
	Description string
	Category    string // カテゴリコード
	Price       int64
	Stock       int
	ImageURL    string
	IsAvailable bool
}

// UpdateProductResponse は商品更新レスポンス
type UpdateProductResponse struct {
	Product *entities.Product
}

// DeleteProductRequest は商品削除リクエスト
type DeleteProductRequest struct {
	ProductID uuid.UUID
}

// GetProductListRequest は商品一覧取得リクエスト
type GetProductListRequest struct {
	Category      string // 空文字列の場合はすべて
	AvailableOnly bool   // trueの場合は交換可能な商品のみ
	Offset        int
	Limit         int
}

// GetProductListResponse は商品一覧取得レスポンス
type GetProductListResponse struct {
	Products []*entities.Product
	Total    int64
}

// ProductManagementInputPort は商品管理のユースケースインターフェース
type ProductManagementInputPort interface {
	// CreateProduct は新しい商品を作成（管理者のみ）
	CreateProduct(ctx context.Context, req *CreateProductRequest) (*CreateProductResponse, error)

	// UpdateProduct は商品情報を更新（管理者のみ）
	UpdateProduct(ctx context.Context, req *UpdateProductRequest) (*UpdateProductResponse, error)

	// DeleteProduct は商品を削除（管理者のみ）
	DeleteProduct(ctx context.Context, req *DeleteProductRequest) error

	// GetProductList は商品一覧を取得
	GetProductList(ctx context.Context, req *GetProductListRequest) (*GetProductListResponse, error)
}

// ==================== ポイント交換（ユーザー用） ====================

// ExchangeProductRequest は商品交換リクエスト
type ExchangeProductRequest struct {
	UserID    uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	Notes     string // 受取場所、希望時間など
}

// ExchangeProductResponse は商品交換レスポンス
type ExchangeProductResponse struct {
	Exchange    *entities.ProductExchange
	Product     *entities.Product
	User        *entities.User
	Transaction *entities.Transaction
}

// GetExchangeHistoryRequest は交換履歴取得リクエスト
type GetExchangeHistoryRequest struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}

// GetExchangeHistoryResponse は交換履歴取得レスポンス
type GetExchangeHistoryResponse struct {
	Exchanges []*entities.ProductExchange
	Total     int64
}

// CancelExchangeRequest は交換キャンセルリクエスト
type CancelExchangeRequest struct {
	UserID     uuid.UUID
	ExchangeID uuid.UUID
}

// MarkExchangeDeliveredRequest は配達完了リクエスト（管理者用）
type MarkExchangeDeliveredRequest struct {
	ExchangeID uuid.UUID
}

// ProductExchangeInputPort は商品交換のユースケースインターフェース
type ProductExchangeInputPort interface {
	// ExchangeProduct はポイントで商品を交換
	ExchangeProduct(ctx context.Context, req *ExchangeProductRequest) (*ExchangeProductResponse, error)

	// GetExchangeHistory は交換履歴を取得
	GetExchangeHistory(ctx context.Context, req *GetExchangeHistoryRequest) (*GetExchangeHistoryResponse, error)

	// CancelExchange は交換をキャンセル（ペンディング状態のみ）
	CancelExchange(ctx context.Context, req *CancelExchangeRequest) error

	// MarkExchangeDelivered は配達完了にする（管理者用）
	MarkExchangeDelivered(ctx context.Context, req *MarkExchangeDeliveredRequest) error

	// GetAllExchanges はすべての交換履歴を取得（管理者用）
	GetAllExchanges(ctx context.Context, offset, limit int) (*GetExchangeHistoryResponse, error)
}
