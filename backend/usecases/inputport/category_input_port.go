package inputport

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// ==================== カテゴリ管理（管理者用） ====================

// CreateCategoryRequest はカテゴリ作成リクエスト
type CreateCategoryRequest struct {
	Name         string `json:"name" binding:"required"`
	Code         string `json:"code" binding:"required"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
}

// CreateCategoryResponse はカテゴリ作成レスポンス
type CreateCategoryResponse struct {
	Category *entities.Category `json:"category"`
}

// UpdateCategoryRequest はカテゴリ更新リクエスト
type UpdateCategoryRequest struct {
	CategoryID   uuid.UUID
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
}

// UpdateCategoryResponse はカテゴリ更新レスポンス
type UpdateCategoryResponse struct {
	Category *entities.Category `json:"category"`
}

// DeleteCategoryRequest はカテゴリ削除リクエスト
type DeleteCategoryRequest struct {
	CategoryID uuid.UUID
}

// GetCategoryListRequest はカテゴリ一覧取得リクエスト
type GetCategoryListRequest struct {
	ActiveOnly bool // trueの場合は有効なカテゴリのみ
}

// GetCategoryListResponse はカテゴリ一覧取得レスポンス
type GetCategoryListResponse struct {
	Categories []*entities.Category `json:"categories"`
	Total      int64                `json:"total"`
}

// CategoryManagementInputPort はカテゴリ管理のユースケースインターフェース
type CategoryManagementInputPort interface {
	// CreateCategory は新しいカテゴリを作成（管理者のみ）
	CreateCategory(req *CreateCategoryRequest) (*CreateCategoryResponse, error)

	// UpdateCategory はカテゴリ情報を更新（管理者のみ）
	UpdateCategory(req *UpdateCategoryRequest) (*UpdateCategoryResponse, error)

	// DeleteCategory はカテゴリを削除（管理者のみ）
	DeleteCategory(req *DeleteCategoryRequest) error

	// GetCategoryList はカテゴリ一覧を取得
	GetCategoryList(req *GetCategoryListRequest) (*GetCategoryListResponse, error)
}
