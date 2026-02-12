package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// CategoryController はカテゴリ管理のコントローラー
type CategoryController struct {
	categoryUseCase inputport.CategoryManagementInputPort
	logger          entities.Logger
}

// NewCategoryController は新しいCategoryControllerを作成
func NewCategoryController(
	categoryUseCase inputport.CategoryManagementInputPort,
	logger entities.Logger,
) *CategoryController {
	return &CategoryController{
		categoryUseCase: categoryUseCase,
		logger:          logger,
	}
}

// CreateCategory は新しいカテゴリを作成（管理者のみ）
// POST /admin/categories
func (c *CategoryController) CreateCategory(ctx *gin.Context) {
	var req inputport.CreateCategoryRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.categoryUseCase.CreateCategory(ctx, &req)
	if err != nil {
		c.logger.Error("Failed to create category", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// UpdateCategory はカテゴリ情報を更新（管理者のみ）
// PUT /admin/categories/:id
func (c *CategoryController) UpdateCategory(ctx *gin.Context) {
	categoryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	var req inputport.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.CategoryID = categoryID

	resp, err := c.categoryUseCase.UpdateCategory(ctx, &req)
	if err != nil {
		c.logger.Error("Failed to update category", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DeleteCategory はカテゴリを削除（管理者のみ）
// DELETE /admin/categories/:id
func (c *CategoryController) DeleteCategory(ctx *gin.Context) {
	categoryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	req := &inputport.DeleteCategoryRequest{
		CategoryID: categoryID,
	}

	if err := c.categoryUseCase.DeleteCategory(ctx, req); err != nil {
		c.logger.Error("Failed to delete category", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "category deleted successfully"})
}

// GetCategoryList はカテゴリ一覧を取得
// GET /categories?active_only=true
func (c *CategoryController) GetCategoryList(ctx *gin.Context) {
	activeOnly := ctx.Query("active_only") == "true"

	req := &inputport.GetCategoryListRequest{
		ActiveOnly: activeOnly,
	}

	resp, err := c.categoryUseCase.GetCategoryList(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get category list", entities.NewField("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
