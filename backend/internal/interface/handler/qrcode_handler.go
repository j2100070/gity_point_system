package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/internal/interface/middleware"
	"github.com/gity/point-system/internal/usecase"
)

// QRCodeHandler はQRコード関連のHTTPハンドラー
type QRCodeHandler struct {
	qrCodeUC *usecase.QRCodeUseCase
}

// NewQRCodeHandler は新しいQRCodeHandlerを作成
func NewQRCodeHandler(qrCodeUC *usecase.QRCodeUseCase) *QRCodeHandler {
	return &QRCodeHandler{
		qrCodeUC: qrCodeUC,
	}
}

// GenerateReceiveQRRequest は受取用QRコード生成リクエスト
type GenerateReceiveQRRequest struct {
	Amount *int64 `json:"amount"` // nilの場合は金額未指定
}

// GenerateReceiveQR は受取用QRコードを生成
// POST /api/qr/generate/receive
func (h *QRCodeHandler) GenerateReceiveQR(c *gin.Context) {
	var req GenerateReceiveQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.qrCodeUC.GenerateReceiveQR(&usecase.GenerateReceiveQRRequest{
		UserID: userID,
		Amount: req.Amount,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"qr_code": gin.H{
			"id":         resp.QRCode.ID,
			"code":       resp.QRCode.Code,
			"qr_data":    resp.QRCodeData, // QRコードに含めるデータ
			"amount":     resp.QRCode.Amount,
			"qr_type":    resp.QRCode.QRType,
			"expires_at": resp.QRCode.ExpiresAt,
		},
	})
}

// GenerateSendQRRequest は送信用QRコード生成リクエスト
type GenerateSendQRRequest struct {
	Amount int64 `json:"amount" binding:"required,min=1"`
}

// GenerateSendQR は送信用QRコードを生成
// POST /api/qr/generate/send
func (h *QRCodeHandler) GenerateSendQR(c *gin.Context) {
	var req GenerateSendQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.qrCodeUC.GenerateSendQR(&usecase.GenerateSendQRRequest{
		UserID: userID,
		Amount: req.Amount,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"qr_code": gin.H{
			"id":         resp.QRCode.ID,
			"code":       resp.QRCode.Code,
			"qr_data":    resp.QRCodeData,
			"amount":     resp.QRCode.Amount,
			"qr_type":    resp.QRCode.QRType,
			"expires_at": resp.QRCode.ExpiresAt,
		},
	})
}

// ScanQRRequest はQRコードスキャンリクエスト
type ScanQRRequest struct {
	QRCode         string `json:"qr_code" binding:"required"`
	Amount         *int64 `json:"amount"` // 金額未指定QRの場合
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

// ScanQR はQRコードをスキャンしてポイント転送
// POST /api/qr/scan
func (h *QRCodeHandler) ScanQR(c *gin.Context) {
	var req ScanQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.qrCodeUC.ScanQR(&usecase.ScanQRRequest{
		ScannerUserID:  userID,
		QRCode:         req.QRCode,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "qr code scanned successfully",
		"transaction": gin.H{
			"id":          resp.Transaction.ID,
			"amount":      resp.Transaction.Amount,
			"status":      resp.Transaction.Status,
			"created_at":  resp.Transaction.CreatedAt,
		},
	})
}

// GetQRCodeHistory はQRコード履歴を取得
// GET /api/qr/history
func (h *QRCodeHandler) GetQRCodeHistory(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
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

	resp, err := h.qrCodeUC.GetQRCodeHistory(&usecase.GetQRCodeHistoryRequest{
		UserID: userID,
		Offset: offset,
		Limit:  limit,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	qrCodes := make([]gin.H, len(resp.QRCodes))
	for i, qr := range resp.QRCodes {
		qrCodes[i] = gin.H{
			"id":             qr.ID,
			"code":           qr.Code,
			"amount":         qr.Amount,
			"qr_type":        qr.QRType,
			"expires_at":     qr.ExpiresAt,
			"used_at":        qr.UsedAt,
			"used_by_user_id": qr.UsedByUserID,
			"created_at":     qr.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"qr_codes": qrCodes,
	})
}
