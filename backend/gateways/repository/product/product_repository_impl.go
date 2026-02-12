package product

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// ProductRepositoryImpl はProductRepositoryの実装
type ProductRepositoryImpl struct {
	productDS dsmysql.ProductDataSource
	logger    entities.Logger
}

// NewProductRepository は新しいProductRepositoryを作成
func NewProductRepository(productDS dsmysql.ProductDataSource, logger entities.Logger) repository.ProductRepository {
	return &ProductRepositoryImpl{
		productDS: productDS,
		logger:    logger,
	}
}

// Create は新しい商品を作成
func (r *ProductRepositoryImpl) Create(ctx context.Context, product *entities.Product) error {
	r.logger.Debug("Creating product", entities.NewField("name", product.Name))
	return r.productDS.Insert(ctx, product)
}

// Read はIDで商品を検索
func (r *ProductRepositoryImpl) Read(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	return r.productDS.Select(ctx, id)
}

// Update は商品情報を更新
func (r *ProductRepositoryImpl) Update(ctx context.Context, product *entities.Product) error {
	r.logger.Debug("Updating product", entities.NewField("product_id", product.ID))
	return r.productDS.Update(ctx, product)
}

// Delete は商品を論理削除
func (r *ProductRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting product", entities.NewField("product_id", id))
	return r.productDS.Delete(ctx, id)
}

// ReadList は商品一覧を取得
func (r *ProductRepositoryImpl) ReadList(ctx context.Context, offset, limit int) ([]*entities.Product, error) {
	return r.productDS.SelectList(ctx, offset, limit)
}

// ReadListByCategory はカテゴリ別の商品一覧を取得
func (r *ProductRepositoryImpl) ReadListByCategory(ctx context.Context, categoryCode string, offset, limit int) ([]*entities.Product, error) {
	return r.productDS.SelectListByCategory(ctx, categoryCode, offset, limit)
}

// ReadAvailableList は交換可能な商品一覧を取得
func (r *ProductRepositoryImpl) ReadAvailableList(ctx context.Context, offset, limit int) ([]*entities.Product, error) {
	return r.productDS.SelectAvailableList(ctx, offset, limit)
}

// Count は商品総数を取得
func (r *ProductRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.productDS.Count(ctx)
}

// UpdateStock は在庫を更新
func (r *ProductRepositoryImpl) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return r.productDS.UpdateStock(ctx, productID, quantity)
}
