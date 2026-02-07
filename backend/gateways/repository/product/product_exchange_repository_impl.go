package product

import (
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
func (r *ProductExchangeRepositoryImpl) Create(tx interface{}, exchange *entities.ProductExchange) error {
	r.logger.Debug("Creating product exchange",
		entities.NewField("user_id", exchange.UserID),
		entities.NewField("product_id", exchange.ProductID))
	return r.exchangeDS.Insert(tx, exchange)
}

// Read はIDで交換を検索
func (r *ProductExchangeRepositoryImpl) Read(id uuid.UUID) (*entities.ProductExchange, error) {
	return r.exchangeDS.Select(id)
}

// Update は交換情報を更新
func (r *ProductExchangeRepositoryImpl) Update(tx interface{}, exchange *entities.ProductExchange) error {
	r.logger.Debug("Updating product exchange", entities.NewField("exchange_id", exchange.ID))
	return r.exchangeDS.Update(tx, exchange)
}

// ReadListByUserID はユーザーの交換履歴を取得
func (r *ProductExchangeRepositoryImpl) ReadListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.ProductExchange, error) {
	return r.exchangeDS.SelectListByUserID(userID, offset, limit)
}

// ReadListAll はすべての交換履歴を取得
func (r *ProductExchangeRepositoryImpl) ReadListAll(offset, limit int) ([]*entities.ProductExchange, error) {
	return r.exchangeDS.SelectListAll(offset, limit)
}

// CountByUserID はユーザーの交換総数を取得
func (r *ProductExchangeRepositoryImpl) CountByUserID(userID uuid.UUID) (int64, error) {
	return r.exchangeDS.CountByUserID(userID)
}

// CountAll は全体の交換総数を取得
func (r *ProductExchangeRepositoryImpl) CountAll() (int64, error) {
	return r.exchangeDS.CountAll()
}
