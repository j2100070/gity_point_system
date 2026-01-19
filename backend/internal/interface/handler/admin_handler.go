package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/internal/domain"
	"github.com/gity/point-system/internal/interface/middleware"
	"github.com/gity/point-system/internal/usecase"
	"github.com/google/uuid"
)

// AdminHandler は管理者機能関連のHTTPハンドラー
type AdminHandler struct {
	adminUC *usecase.AdminUseCase
}

// NewAdminHandler は新しいAdminHandlerを作成
func NewAdminHandler(adminUC *usecase.AdminUseCase) *AdminHandler {
	return &AdminHandler{
		adminUC: adminUC,
	}
}

// GrantPointsRequest はポイント付与リクエスト
type GrantPointsRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required,uuid"`
	Amount       int64  `json:"amount" binding:"required,min=1"`
	Description  string `json:"description" binding:"required"`
}

// GrantPoints はユーザーにポイントを付与（管理者のみ）
// POST /api/admin/points/grant
func (h *AdminHandler) GrantPoints(c *gin.Context) {
	var req GrantPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetUserID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_user_id"})
		return
	}

	resp, err := h.adminUC.GrantPoints(&usecase.GrantPointsRequest{
		AdminID:      adminID,
		TargetUserID: targetUserID,
		Amount:       req.Amount,
		Description:  req.Description,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "points granted successfully",
		"transaction": gin.H{
			"id":     resp.Transaction.ID,
			"amount": resp.Transaction.Amount,
		},
		"user": gin.H{
			"id":      resp.User.ID,
			"balance": resp.User.Balance,
		},
	})
}

// DeductPointsRequest はポイント減算リクエスト
type DeductPointsRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required,uuid"`
	Amount       int64  `json:"amount" binding:"required,min=1"`
	Description  string `json:"description" binding:"required"`
}

// DeductPoints はユーザーからポイントを減算（管理者のみ）
// POST /api/admin/points/deduct
func (h *AdminHandler) DeductPoints(c *gin.Context) {
	var req DeductPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetUserID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_user_id"})
		return
	}

	resp, err := h.adminUC.DeductPoints(&usecase.DeductPointsRequest{
		AdminID:      adminID,
		TargetUserID: targetUserID,
		Amount:       req.Amount,
		Description:  req.Description,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "points deducted successfully",
		"transaction": gin.H{
			"id":     resp.Transaction.ID,
			"amount": resp.Transaction.Amount,
		},
		"user": gin.H{
			"id":      resp.User.ID,
			"balance": resp.User.Balance,
		},
	})
}

// ListAllUsers は全ユーザー一覧を取得（管理者のみ）
// GET /api/admin/users
func (h *AdminHandler) ListAllUsers(c *gin.Context) {
	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	offset := 0
	limit := 20
	if c.Query("offset") != "" {
		fmt.Sscanf(c.Query("offset"), "%d", &offset)
	}
	if c.Query("limit") != "" {
		fmt.Sscanf(c.Query("limit"), "%d", &limit)
	}

	resp, err := h.adminUC.ListAllUsers(&usecase.ListAllUsersRequest{
		AdminID: adminID,
		Offset:  offset,
		Limit:   limit,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users := make([]gin.H, len(resp.Users))
	for i, u := range resp.Users {
		users[i] = gin.H{
			"id":           u.ID,
			"username":     u.Username,
			"email":        u.Email,
			"display_name": u.DisplayName,
			"balance":      u.Balance,
			"role":         u.Role,
			"is_active":    u.IsActive,
			"created_at":   u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  resp.Total,
		"offset": offset,
		"limit":  limit,
	})
}

// ListAllTransactions は全トランザクション一覧を取得（管理者のみ）
// GET /api/admin/transactions
func (h *AdminHandler) ListAllTransactions(c *gin.Context) {
	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	offset := 0
	limit := 20
	if c.Query("offset") != "" {
		fmt.Sscanf(c.Query("offset"), "%d", &offset)
	}
	if c.Query("limit") != "" {
		fmt.Sscanf(c.Query("limit"), "%d", &limit)
	}

	resp, err := h.adminUC.ListAllTransactions(&usecase.ListAllTransactionsRequest{
		AdminID: adminID,
		Offset:  offset,
		Limit:   limit,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactions := make([]gin.H, len(resp.Transactions))
	for i, tx := range resp.Transactions {
		transactions[i] = gin.H{
			"id":               tx.ID,
			"from_user_id":     tx.FromUserID,
			"to_user_id":       tx.ToUserID,
			"amount":           tx.Amount,
			"transaction_type": tx.TransactionType,
			"status":           tx.Status,
			"description":      tx.Description,
			"created_at":       tx.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
	})
}

// UpdateUserRoleRequest はユーザー役割変更リクエスト
type UpdateUserRoleRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required,uuid"`
	NewRole      string `json:"new_role" binding:"required"`
}

// UpdateUserRole はユーザーの役割を変更（管理者のみ）
// POST /api/admin/users/role
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetUserID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_user_id"})
		return
	}

	resp, err := h.adminUC.UpdateUserRole(&usecase.UpdateUserRoleRequest{
		AdminID:      adminID,
		TargetUserID: targetUserID,
		NewRole:      domain.UserRole(req.NewRole),
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user role updated successfully",
		"user": gin.H{
			"id":   resp.User.ID,
			"role": resp.User.Role,
		},
	})
}

// DeactivateUserRequest はユーザー無効化リクエスト
type DeactivateUserRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required,uuid"`
}

// DeactivateUser はユーザーを無効化（管理者のみ）
// POST /api/admin/users/deactivate
func (h *AdminHandler) DeactivateUser(c *gin.Context) {
	var req DeactivateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetUserID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_user_id"})
		return
	}

	resp, err := h.adminUC.DeactivateUser(&usecase.DeactivateUserRequest{
		AdminID:      adminID,
		TargetUserID: targetUserID,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user deactivated successfully",
		"user": gin.H{
			"id":        resp.User.ID,
			"is_active": resp.User.IsActive,
		},
	})
}
