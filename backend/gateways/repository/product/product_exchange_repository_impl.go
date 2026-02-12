package product

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// ProductExchangeRepositoryImpl はProductExchangeRepositoryの実装
type ProductExchangeRepositoryImpl struct {
	exchangeDS dsmysql.ProductExchangeDataSource
	logger     entities.Logger
}

// NewProductExchangeRepository は新しいProductExchangeRepositoryを作成
func NewProductExchangeRepository(exchangeDS dsmysql.ProductExchangeDataSource, logger entities.Logger) repository.ProductExchangeRepository {
	return &ProductExchangeRepositoryImpl{
		exchangeDS: exchangeDS,
		logger:     logger,
	}
}

// Create は新しい交換を作成
func (r *ProductExchangeRepositoryImpl) Create(ctx context.Context, exchange *entities.ProductExchange) error {
	r.logger.Debug("Creating product exchange",
		entities.NewField("user_id", exchange.UserID),
		entities.NewField("product_id", exchange.ProductID))
	return r.exchangeDS.Insert(ctx, exchange)
}

// Read はIDで交換を検索
func (r *ProductExchangeRepositoryImpl) Read(ctx context.Context, id uuid.UUID) (*entities.ProductExchange, error) {
	return r.exchangeDS.Select(ctx, id)
}

// Update は交換情報を更新
func (r *ProductExchangeRepositoryImpl) Update(ctx context.Context, exchange *entities.ProductExchange) error {
	r.logger.Debug("Updating product exchange", entities.NewField("exchange_id", exchange.ID))
	return r.exchangeDS.Update(ctx, exchange)
}

// ReadListByUserID はユーザーの交換履歴を取得
func (r *ProductExchangeRepositoryImpl) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.ProductExchange, error) {
	return r.exchangeDS.SelectListByUserID(ctx, userID, offset, limit)
}

// ReadListAll はすべての交換履歴を取得
func (r *ProductExchangeRepositoryImpl) ReadListAll(ctx context.Context, offset, limit int) ([]*entities.ProductExchange, error) {
	return r.exchangeDS.SelectListAll(ctx, offset, limit)
}

// CountByUserID はユーザーの交換総数を取得
func (r *ProductExchangeRepositoryImpl) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.exchangeDS.CountByUserID(ctx, userID)
}

// CountAll は全体の交換総数を取得
func (r *ProductExchangeRepositoryImpl) CountAll(ctx context.Context) (int64, error) {
	return r.exchangeDS.CountAll(ctx)
}
