package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// PointController はポイント関連のコントローラー
// 外界からの入力を、達成するユースケースが求めるインターフェースに変換する責務
type PointController struct {
	pointTransferUC inputport.PointTransferInputPort
	presenter       *presenter.PointPresenter
}

// NewPointController は新しいPointControllerを作成
func NewPointController(
	pointTransferUC inputport.PointTransferInputPort,
	presenter *presenter.PointPresenter,
) *PointController {
	return &PointController{
		pointTransferUC: pointTransferUC,
		presenter:       presenter,
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
func (c *PointController) Transfer(ctx *gin.Context, currentTime time.Time) {
	var req TransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 認証ユーザーIDを取得（ミドルウェアでセット済み）
	fromUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_user_id"})
		return
	}

	// ユースケースを実行
	resp, err := c.pointTransferUC.Transfer(ctx, &inputport.TransferRequest{
		FromUserID:     fromUserID.(uuid.UUID),
		ToUserID:       toUserID,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
		Description:    req.Description,
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Presenterで変換して出力
	output := c.presenter.PresentTransferResponse(resp)
	ctx.JSON(http.StatusOK, output)
}

// GetBalance は残高を取得
// GET /api/points/balance
func (c *PointController) GetBalance(ctx *gin.Context, currentTime time.Time) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := c.pointTransferUC.GetBalance(ctx, &inputport.GetBalanceRequest{
		UserID: userID.(uuid.UUID),
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentBalanceResponse(resp)
	ctx.JSON(http.StatusOK, output)
}

// GetTransactionHistory はトランザクション履歴を取得
// GET /api/points/history
func (c *PointController) GetTransactionHistory(ctx *gin.Context, currentTime time.Time) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// ページネーション
	offset := 0
	limit := 20
	if ctx.Query("offset") != "" {
		fmt.Sscanf(ctx.Query("offset"), "%d", &offset)
	}
	if ctx.Query("limit") != "" {
		fmt.Sscanf(ctx.Query("limit"), "%d", &limit)
	}

	resp, err := c.pointTransferUC.GetTransactionHistory(ctx, &inputport.GetTransactionHistoryRequest{
		UserID: userID.(uuid.UUID),
		Offset: offset,
		Limit:  limit,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := c.presenter.PresentTransactionHistoryResponse(resp)
	ctx.JSON(http.StatusOK, output)
}
