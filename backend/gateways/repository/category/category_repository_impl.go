package category

import (
	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// CategoryRepositoryImpl はCategoryRepositoryの実装
type CategoryRepositoryImpl struct {
	categoryDS dsmysql.CategoryDataSource
	logger     entities.Logger
}

// NewCategoryRepository は新しいCategoryRepositoryを作成
func NewCategoryRepository(categoryDS dsmysql.CategoryDataSource, logger entities.Logger) repository.CategoryRepository {
	return &CategoryRepositoryImpl{
		categoryDS: categoryDS,
		logger:     logger,
	}
}

// Create は新しいカテゴリを作成
func (r *CategoryRepositoryImpl) Create(category *entities.Category) error {
	r.logger.Debug("Creating category", entities.NewField("name", category.Name))
	return r.categoryDS.Insert(category)
}

// Read はIDでカテゴリを検索
func (r *CategoryRepositoryImpl) Read(id uuid.UUID) (*entities.Category, error) {
	return r.categoryDS.Select(id)
}

// ReadByCode はコードでカテゴリを検索
func (r *CategoryRepositoryImpl) ReadByCode(code string) (*entities.Category, error) {
	return r.categoryDS.SelectByCode(code)
}

// Update はカテゴリ情報を更新
func (r *CategoryRepositoryImpl) Update(category *entities.Category) error {
	r.logger.Debug("Updating category", entities.NewField("category_id", category.ID))
	return r.categoryDS.Update(category)
}

// Delete はカテゴリを論理削除
func (r *CategoryRepositoryImpl) Delete(id uuid.UUID) error {
	r.logger.Debug("Deleting category", entities.NewField("category_id", id))
	return r.categoryDS.Delete(id)
}

// ReadList はカテゴリ一覧を取得
func (r *CategoryRepositoryImpl) ReadList(activeOnly bool) ([]*entities.Category, error) {
	return r.categoryDS.SelectList(activeOnly)
}

// Count はカテゴリ総数を取得
func (r *CategoryRepositoryImpl) Count() (int64, error) {
	return r.categoryDS.Count()
}

// ExistsCode はコードが存在するか確認
func (r *CategoryRepositoryImpl) ExistsCode(code string, excludeID *uuid.UUID) (bool, error) {
	return r.categoryDS.ExistsCode(code, excludeID)
}
