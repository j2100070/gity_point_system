package interactor

import (
	"context"
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
)

// CategoryManagementInteractor はカテゴリ管理のユースケース実装（管理者用）
type CategoryManagementInteractor struct {
	categoryRepo repository.CategoryRepository
	logger       entities.Logger
}

// NewCategoryManagementInteractor は新しいCategoryManagementInteractorを作成
func NewCategoryManagementInteractor(
	categoryRepo repository.CategoryRepository,
	logger entities.Logger,
) inputport.CategoryManagementInputPort {
	return &CategoryManagementInteractor{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

// CreateCategory は新しいカテゴリを作成（管理者のみ）
func (i *CategoryManagementInteractor) CreateCategory(ctx context.Context, req *inputport.CreateCategoryRequest) (*inputport.CreateCategoryResponse, error) {
	i.logger.Info("Creating new category", entities.NewField("name", req.Name), entities.NewField("code", req.Code))

	// コードの重複チェック
	exists, err := i.categoryRepo.ExistsCode(ctx, req.Code, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check code existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("category code '%s' already exists", req.Code)
	}

	category, err := entities.NewCategory(req.Name, req.Code, req.Description, req.DisplayOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	if err := i.categoryRepo.Create(ctx, category); err != nil {
		i.logger.Error("Failed to create category", entities.NewField("error", err))
		return nil, fmt.Errorf("failed to save category: %w", err)
	}

	i.logger.Info("Category created successfully", entities.NewField("category_id", category.ID))

	return &inputport.CreateCategoryResponse{
		Category: category,
	}, nil
}

// UpdateCategory はカテゴリ情報を更新（管理者のみ）
func (i *CategoryManagementInteractor) UpdateCategory(ctx context.Context, req *inputport.UpdateCategoryRequest) (*inputport.UpdateCategoryResponse, error) {
	i.logger.Info("Updating category", entities.NewField("category_id", req.CategoryID))

	category, err := i.categoryRepo.Read(ctx, req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// カテゴリ情報を更新（コードは変更不可）
	category.Update(req.Name, req.Description, req.DisplayOrder, req.IsActive)

	if err := i.categoryRepo.Update(ctx, category); err != nil {
		i.logger.Error("Failed to update category", entities.NewField("error", err))
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	i.logger.Info("Category updated successfully", entities.NewField("category_id", category.ID))

	return &inputport.UpdateCategoryResponse{
		Category: category,
	}, nil
}

// DeleteCategory はカテゴリを削除（管理者のみ）
func (i *CategoryManagementInteractor) DeleteCategory(ctx context.Context, req *inputport.DeleteCategoryRequest) error {
	i.logger.Info("Deleting category", entities.NewField("category_id", req.CategoryID))

	// カテゴリの存在確認
	_, err := i.categoryRepo.Read(ctx, req.CategoryID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// TODO: このカテゴリを使用している商品がないかチェックする

	if err := i.categoryRepo.Delete(ctx, req.CategoryID); err != nil {
		i.logger.Error("Failed to delete category", entities.NewField("error", err))
		return fmt.Errorf("failed to delete category: %w", err)
	}

	i.logger.Info("Category deleted successfully", entities.NewField("category_id", req.CategoryID))
	return nil
}

// GetCategoryList はカテゴリ一覧を取得
func (i *CategoryManagementInteractor) GetCategoryList(ctx context.Context, req *inputport.GetCategoryListRequest) (*inputport.GetCategoryListResponse, error) {
	categories, err := i.categoryRepo.ReadList(ctx, req.ActiveOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to get category list: %w", err)
	}

	total, err := i.categoryRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count categories: %w", err)
	}

	return &inputport.GetCategoryListResponse{
		Categories: categories,
		Total:      total,
	}, nil
}
