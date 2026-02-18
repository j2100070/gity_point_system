package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// AdminController は管理者機能のコントローラー
type AdminController struct {
	adminUC   inputport.AdminInputPort
	presenter *presenter.AdminPresenter
}

// NewAdminController は新しいAdminControllerを作成
func NewAdminController(
	adminUC inputport.AdminInputPort,
	presenter *presenter.AdminPresenter,
) *AdminController {
	return &AdminController{
		adminUC:   adminUC,
		presenter: presenter,
	}
}

// GrantPoints はユーザーにポイントを付与
// POST /api/admin/points/grant
func (c *AdminController) GrantPoints(ctx *gin.Context) {
	// ログインユーザー（管理者）取得
	adminID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// リクエストボディ解析
	var req struct {
		UserID         string `json:"user_id" binding:"required"`
		Amount         int64  `json:"amount" binding:"required"`
		Description    string `json:"description" binding:"required"`
		IdempotencyKey string `json:"idempotency_key" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// UUID変換
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// ユースケース実行
	resp, err := c.adminUC.GrantPoints(ctx, &inputport.GrantPointsRequest{
		AdminID:        adminID.(uuid.UUID),
		UserID:         userID,
		Amount:         req.Amount,
		Description:    req.Description,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentGrantPoints(resp))
}

// DeductPoints はユーザーからポイントを減算
// POST /api/admin/points/deduct
func (c *AdminController) DeductPoints(ctx *gin.Context) {
	// ログインユーザー（管理者）取得
	adminID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// リクエストボディ解析
	var req struct {
		UserID         string `json:"user_id" binding:"required"`
		Amount         int64  `json:"amount" binding:"required"`
		Description    string `json:"description" binding:"required"`
		IdempotencyKey string `json:"idempotency_key" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// UUID変換
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// ユースケース実行
	resp, err := c.adminUC.DeductPoints(ctx, &inputport.DeductPointsRequest{
		AdminID:        adminID.(uuid.UUID),
		UserID:         userID,
		Amount:         req.Amount,
		Description:    req.Description,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentDeductPoints(resp))
}

// ListAllUsers はすべてのユーザー一覧を取得
// GET /api/admin/users
func (c *AdminController) ListAllUsers(ctx *gin.Context) {
	// 管理者権限チェックはInteractor層で行う

	// クエリパラメータ取得
	var offset, limit int
	fmt.Sscanf(ctx.Query("offset"), "%d", &offset)
	fmt.Sscanf(ctx.Query("limit"), "%d", &limit)
	if limit == 0 {
		limit = 50
	}

	search := ctx.Query("search")
	sortBy := ctx.Query("sort_by")
	sortOrder := ctx.Query("sort_order")

	// ユースケース実行
	resp, err := c.adminUC.ListAllUsers(ctx, &inputport.ListAllUsersRequest{
		Offset:    offset,
		Limit:     limit,
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentListAllUsers(resp))
}

// ListAllTransactions はすべての取引履歴を取得
// GET /api/admin/transactions
func (c *AdminController) ListAllTransactions(ctx *gin.Context) {
	// 管理者権限チェックはInteractor層で行う

	// クエリパラメータ取得
	var offset, limit int
	fmt.Sscanf(ctx.Query("offset"), "%d", &offset)
	fmt.Sscanf(ctx.Query("limit"), "%d", &limit)
	if limit == 0 {
		limit = 50
	}

	transactionType := ctx.Query("transaction_type")
	dateFrom := ctx.Query("date_from")
	dateTo := ctx.Query("date_to")
	sortBy := ctx.Query("sort_by")
	sortOrder := ctx.Query("sort_order")

	// ユースケース実行
	resp, err := c.adminUC.ListAllTransactions(ctx, &inputport.ListAllTransactionsRequest{
		Offset:          offset,
		Limit:           limit,
		TransactionType: transactionType,
		DateFrom:        dateFrom,
		DateTo:          dateTo,
		SortBy:          sortBy,
		SortOrder:       sortOrder,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentListAllTransactions(resp))
}

// UpdateUserRole はユーザーの役割を更新
// PUT /api/admin/users/:id/role
func (c *AdminController) UpdateUserRole(ctx *gin.Context) {
	// ログインユーザー（管理者）取得
	adminID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// パスパラメータ取得
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// リクエストボディ解析
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// ユースケース実行
	resp, err := c.adminUC.UpdateUserRole(ctx, &inputport.UpdateUserRoleRequest{
		AdminID: adminID.(uuid.UUID),
		UserID:  userID,
		Role:    req.Role,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentUpdateUserRole(resp))
}

// DeactivateUser はユーザーを無効化
// POST /api/admin/users/:id/deactivate
func (c *AdminController) DeactivateUser(ctx *gin.Context) {
	// ログインユーザー（管理者）取得
	adminID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// パスパラメータ取得
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// ユースケース実行
	resp, err := c.adminUC.DeactivateUser(ctx, &inputport.DeactivateUserRequest{
		AdminID: adminID.(uuid.UUID),
		UserID:  userID,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentDeactivateUser(resp))
}

// GetAnalytics は分析データを取得
// GET /api/admin/analytics
func (c *AdminController) GetAnalytics(ctx *gin.Context) {
	var days int
	fmt.Sscanf(ctx.Query("days"), "%d", &days)
	if days == 0 {
		days = 30
	}

	resp, err := c.adminUC.GetAnalytics(ctx, &inputport.GetAnalyticsRequest{
		Days: days,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, c.presenter.PresentAnalytics(resp))
}
