package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// QRCodeController はQRコード機能のコントローラー
type QRCodeController struct {
	qrCodeUC  inputport.QRCodeInputPort
	presenter *presenter.QRCodePresenter
}

// NewQRCodeController は新しいQRCodeControllerを作成
func NewQRCodeController(
	qrCodeUC inputport.QRCodeInputPort,
	presenter *presenter.QRCodePresenter,
) *QRCodeController {
	return &QRCodeController{
		qrCodeUC:  qrCodeUC,
		presenter: presenter,
	}
}

// GenerateReceiveQR は受取用QRコードを生成
// POST /api/qrcodes/receive
func (c *QRCodeController) GenerateReceiveQR(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// リクエストボディ解析
	var req struct {
		Amount *int64 `json:"amount"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// ユースケース実行
	resp, err := c.qrCodeUC.GenerateReceiveQR(ctx, &inputport.GenerateReceiveQRRequest{
		UserID: userID.(uuid.UUID),
		Amount: req.Amount,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusCreated, c.presenter.PresentGenerateReceiveQR(resp))
}

// GenerateSendQR は送信用QRコードを生成
// POST /api/qrcodes/send
func (c *QRCodeController) GenerateSendQR(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// リクエストボディ解析
	var req struct {
		Amount int64 `json:"amount" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// ユースケース実行
	resp, err := c.qrCodeUC.GenerateSendQR(ctx, &inputport.GenerateSendQRRequest{
		UserID: userID.(uuid.UUID),
		Amount: req.Amount,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusCreated, c.presenter.PresentGenerateSendQR(resp))
}

// ScanQR はQRコードをスキャンしてポイント転送
// POST /api/qrcodes/scan
func (c *QRCodeController) ScanQR(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// リクエストボディ解析
	var req struct {
		Code           string `json:"code" binding:"required"`
		Amount         *int64 `json:"amount"`
		IdempotencyKey string `json:"idempotency_key" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// ユースケース実行
	resp, err := c.qrCodeUC.ScanQR(ctx, &inputport.ScanQRRequest{
		UserID:         userID.(uuid.UUID),
		Code:           req.Code,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentScanQR(resp))
}

// GetQRCodeHistory はQRコード履歴を取得
// GET /api/qrcodes/history
func (c *QRCodeController) GetQRCodeHistory(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// クエリパラメータ取得
	var offset, limit int
	fmt.Sscanf(ctx.Query("offset"), "%d", &offset)
	fmt.Sscanf(ctx.Query("limit"), "%d", &limit)
	if limit == 0 {
		limit = 20
	}

	// ユースケース実行
	resp, err := c.qrCodeUC.GetQRCodeHistory(ctx, &inputport.GetQRCodeHistoryRequest{
		UserID: userID.(uuid.UUID),
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentGetQRCodeHistory(resp))
}
