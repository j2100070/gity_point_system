package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/internal/interface/middleware"
	"github.com/gity/point-system/internal/usecase"
	"github.com/google/uuid"
)

// PointHandler はポイント関連のHTTPハンドラー
type PointHandler struct {
	pointTransferUC *usecase.PointTransferUseCase
}

// NewPointHandler は新しいPointHandlerを作成
func NewPointHandler(pointTransferUC *usecase.PointTransferUseCase) *PointHandler {
	return &PointHandler{
		pointTransferUC: pointTransferUC,
	}
}

// TransferRequest はポイント転送リクエスト
type TransferRequest struct {
	ToUserID       string `json:"to_user_id" binding:"required,uuid"`
	Amount         int64  `json:"amount" binding:"required,min=1"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
	Description    string `json:"description"`
}

// Transfer はポイント転送
// POST /api/points/transfer
//
// セキュリティ:
// - 認証ミドルウェアで認証済み
// - CSRF保護有効
// - 冪等性キーで重複防止
func (h *PointHandler) Transfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 認証ユーザーIDを取得
	fromUserID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_user_id"})
		return
	}

	resp, err := h.pointTransferUC.Transfer(&usecase.TransferRequest{
		FromUserID:     fromUserID,
		ToUserID:       toUserID,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
		Description:    req.Description,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "transfer successful",
		"transaction": gin.H{
			"id":          resp.Transaction.ID,
			"amount":      resp.Transaction.Amount,
			"status":      resp.Transaction.Status,
			"created_at":  resp.Transaction.CreatedAt,
		},
		"new_balance": resp.FromUser.Balance,
	})
}

// GetBalance は残高を取得
// GET /api/points/balance
func (h *PointHandler) GetBalance(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.pointTransferUC.GetBalance(&usecase.GetBalanceRequest{
		UserID: userID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": resp.Balance,
		"user": gin.H{
			"id":           resp.User.ID,
			"username":     resp.User.Username,
			"display_name": resp.User.DisplayName,
		},
	})
}

// GetTransactionHistory はトランザクション履歴を取得
// GET /api/points/history
func (h *PointHandler) GetTransactionHistory(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// ページネーション
	offset := 0
	limit := 20
	if c.Query("offset") != "" {
		fmt.Sscanf(c.Query("offset"), "%d", &offset)
	}
	if c.Query("limit") != "" {
		fmt.Sscanf(c.Query("limit"), "%d", &limit)
	}

	resp, err := h.pointTransferUC.GetTransactionHistory(&usecase.GetTransactionHistoryRequest{
		UserID: userID,
		Offset: offset,
		Limit:  limit,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		"total":        resp.Total,
		"offset":       offset,
		"limit":        limit,
	})
}
