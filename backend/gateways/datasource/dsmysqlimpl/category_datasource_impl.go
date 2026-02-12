package dsmysqlimpl

import (
	"context"
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryModel はGORM用のカテゴリモデル
type CategoryModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         string     `gorm:"type:varchar(100);not null"`
	Code         string     `gorm:"type:varchar(50);not null;uniqueIndex"`
	Description  string     `gorm:"type:text"`
	DisplayOrder int        `gorm:"not null;default:0"`
	IsActive     bool       `gorm:"not null;default:true"`
	CreatedAt    time.Time  `gorm:"not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"not null;default:now()"`
	DeletedAt    *time.Time `gorm:"index"`
}

// TableName はテーブル名を指定
func (CategoryModel) TableName() string {
	return "categories"
}

// ToDomain はドメインモデルに変換
func (c *CategoryModel) ToDomain() *entities.Category {
	return &entities.Category{
		ID:           c.ID,
		Name:         c.Name,
		Code:         c.Code,
		Description:  c.Description,
		DisplayOrder: c.DisplayOrder,
		IsActive:     c.IsActive,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		DeletedAt:    c.DeletedAt,
	}
}

// FromDomain はドメインモデルから変換
func (c *CategoryModel) FromDomain(category *entities.Category) {
	c.ID = category.ID
	c.Name = category.Name
	c.Code = category.Code
	c.Description = category.Description
	c.DisplayOrder = category.DisplayOrder
	c.IsActive = category.IsActive
	c.CreatedAt = category.CreatedAt
	c.UpdatedAt = category.UpdatedAt
	c.DeletedAt = category.DeletedAt
}

// CategoryDataSourceImpl はCategoryDataSourceの実装
type CategoryDataSourceImpl struct {
	db inframysql.DB
}

// NewCategoryDataSource は新しいCategoryDataSourceを作成
func NewCategoryDataSource(db inframysql.DB) dsmysql.CategoryDataSource {
	return &CategoryDataSourceImpl{db: db}
}

// Insert は新しいカテゴリを挿入
func (ds *CategoryDataSourceImpl) Insert(ctx context.Context, category *entities.Category) error {
	model := &CategoryModel{}
	model.FromDomain(category)

	if err := inframysql.GetDB(ctx, ds.db.GetDB()).Create(model).Error; err != nil {
		return err
	}

	*category = *model.ToDomain()
	return nil
}

// Select はIDでカテゴリを検索
func (ds *CategoryDataSourceImpl) Select(ctx context.Context, id uuid.UUID) (*entities.Category, error) {
	var model CategoryModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Where("id = ? AND deleted_at IS NULL", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectByCode はコードでカテゴリを検索
func (ds *CategoryDataSourceImpl) SelectByCode(ctx context.Context, code string) (*entities.Category, error) {
	var model CategoryModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Where("code = ? AND deleted_at IS NULL", code).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update はカテゴリ情報を更新
func (ds *CategoryDataSourceImpl) Update(ctx context.Context, category *entities.Category) error {
	model := &CategoryModel{}
	model.FromDomain(category)
	model.UpdatedAt = time.Now()

	return inframysql.GetDB(ctx, ds.db.GetDB()).Where("id = ? AND deleted_at IS NULL", category.ID).
		Updates(model).Error
}

// Delete はカテゴリを論理削除
func (ds *CategoryDataSourceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return inframysql.GetDB(ctx, ds.db.GetDB()).Model(&CategoryModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
}

// SelectList はカテゴリ一覧を取得
func (ds *CategoryDataSourceImpl) SelectList(ctx context.Context, activeOnly bool) ([]*entities.Category, error) {
	var models []CategoryModel

	query := inframysql.GetDB(ctx, ds.db.GetDB()).Where("deleted_at IS NULL")
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("display_order ASC, name ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}

	categories := make([]*entities.Category, len(models))
	for i, model := range models {
		categories[i] = model.ToDomain()
	}

	return categories, nil
}

// Count はカテゴリ総数を取得
func (ds *CategoryDataSourceImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := inframysql.GetDB(ctx, ds.db.GetDB()).Model(&CategoryModel{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

// ExistsCode はコードの存在確認
func (ds *CategoryDataSourceImpl) ExistsCode(ctx context.Context, code string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := inframysql.GetDB(ctx, ds.db.GetDB()).Model(&CategoryModel{}).Where("code = ? AND deleted_at IS NULL", code)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
