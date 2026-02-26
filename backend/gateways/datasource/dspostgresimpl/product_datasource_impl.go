package dspostgresimpl

import (
	"context"
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	infrapostgres "github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductModel はGORM用の商品モデル
type ProductModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string     `gorm:"type:varchar(255);not null"`
	Description string     `gorm:"type:text"`
	Category    string     `gorm:"type:varchar(100);not null"`
	Price       int64      `gorm:"not null;check:price > 0"`
	Stock       int        `gorm:"not null;check:stock >= -1"`
	ImageURL    string     `gorm:"type:text"`
	IsAvailable bool       `gorm:"not null;default:true"`
	CreatedAt   time.Time  `gorm:"not null;default:now()"`
	UpdatedAt   time.Time  `gorm:"not null;default:now()"`
	DeletedAt   *time.Time `gorm:"index"`
}

// TableName はテーブル名を指定
func (ProductModel) TableName() string {
	return "products"
}

// ToDomain はドメインモデルに変換
func (p *ProductModel) ToDomain() *entities.Product {
	return &entities.Product{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		CategoryCode: p.Category,
		Price:        p.Price,
		Stock:        p.Stock,
		ImageURL:     p.ImageURL,
		IsAvailable:  p.IsAvailable,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		DeletedAt:    p.DeletedAt,
	}
}

// FromDomain はドメインモデルから変換
func (p *ProductModel) FromDomain(product *entities.Product) {
	p.ID = product.ID
	p.Name = product.Name
	p.Description = product.Description
	p.Category = product.CategoryCode
	p.Price = product.Price
	p.Stock = product.Stock
	p.ImageURL = product.ImageURL
	p.IsAvailable = product.IsAvailable
	p.CreatedAt = product.CreatedAt
	p.UpdatedAt = product.UpdatedAt
	p.DeletedAt = product.DeletedAt
}

// ProductDataSourceImpl はProductDataSourceの実装
type ProductDataSourceImpl struct {
	db infrapostgres.DB
}

// NewProductDataSource は新しいProductDataSourceを作成
func NewProductDataSource(db infrapostgres.DB) dsmysql.ProductDataSource {
	return &ProductDataSourceImpl{db: db}
}

// Insert は新しい商品を挿入
func (ds *ProductDataSourceImpl) Insert(ctx context.Context, product *entities.Product) error {
	model := &ProductModel{}
	model.FromDomain(product)

	if err := infrapostgres.GetDB(ctx, ds.db.GetDB()).Create(model).Error; err != nil {
		return err
	}

	*product = *model.ToDomain()
	return nil
}

// Select はIDで商品を検索
func (ds *ProductDataSourceImpl) Select(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	var model ProductModel

	err := infrapostgres.GetDB(ctx, ds.db.GetDB()).Where("id = ? AND deleted_at IS NULL", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update は商品情報を更新
func (ds *ProductDataSourceImpl) Update(ctx context.Context, product *entities.Product) error {
	model := &ProductModel{}
	model.FromDomain(product)
	model.UpdatedAt = time.Now()

	return infrapostgres.GetDB(ctx, ds.db.GetDB()).Where("id = ? AND deleted_at IS NULL", product.ID).
		Updates(model).Error
}

// Delete は商品を論理削除
func (ds *ProductDataSourceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return infrapostgres.GetDB(ctx, ds.db.GetDB()).Model(&ProductModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
}

// SelectList は商品一覧を取得
func (ds *ProductDataSourceImpl) SelectList(ctx context.Context, offset, limit int) ([]*entities.Product, error) {
	var models []ProductModel

	err := infrapostgres.GetDB(ctx, ds.db.GetDB()).Where("deleted_at IS NULL").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	products := make([]*entities.Product, len(models))
	for i, model := range models {
		products[i] = model.ToDomain()
	}

	return products, nil
}

// SelectListByCategory はカテゴリ別の商品一覧を取得
func (ds *ProductDataSourceImpl) SelectListByCategory(ctx context.Context, categoryCode string, offset, limit int) ([]*entities.Product, error) {
	var models []ProductModel

	err := infrapostgres.GetDB(ctx, ds.db.GetDB()).Where("category = ? AND deleted_at IS NULL", categoryCode).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	products := make([]*entities.Product, len(models))
	for i, model := range models {
		products[i] = model.ToDomain()
	}

	return products, nil
}

// SelectAvailableList は交換可能な商品一覧を取得
func (ds *ProductDataSourceImpl) SelectAvailableList(ctx context.Context, offset, limit int) ([]*entities.Product, error) {
	var models []ProductModel

	err := infrapostgres.GetDB(ctx, ds.db.GetDB()).Where("is_available = ? AND deleted_at IS NULL", true).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	products := make([]*entities.Product, len(models))
	for i, model := range models {
		products[i] = model.ToDomain()
	}

	return products, nil
}

// Count は商品総数を取得
func (ds *ProductDataSourceImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := infrapostgres.GetDB(ctx, ds.db.GetDB()).Model(&ProductModel{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

// UpdateStock は在庫を更新
func (ds *ProductDataSourceImpl) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	db := infrapostgres.GetDB(ctx, ds.db.GetDB())

	return db.Model(&ProductModel{}).
		Where("id = ? AND deleted_at IS NULL", productID).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity)).Error
}
