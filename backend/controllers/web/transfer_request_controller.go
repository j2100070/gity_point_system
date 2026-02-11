package web

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// TransferRequestController は送金リクエスト機能のコントローラー
type TransferRequestController struct {
	transferRequestUC inputport.TransferRequestInputPort
	userRepo          repository.UserRepository
	presenter         *presenter.TransferRequestPresenter
}

// NewTransferRequestController は新しいTransferRequestControllerを作成
func NewTransferRequestController(
	transferRequestUC inputport.TransferRequestInputPort,
	userRepo repository.UserRepository,
	presenter *presenter.TransferRequestPresenter,
) *TransferRequestController {
	return &TransferRequestController{
		transferRequestUC: transferRequestUC,
		userRepo:          userRepo,
		presenter:         presenter,
	}
}

// GetPersonalQRCode は個人固定QRコードを取得
// GET /api/transfer-requests/personal-qr
func (c *TransferRequestController) GetPersonalQRCode(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// ユーザー情報取得
	user, err := c.userRepo.Read(userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// レスポンス
	ctx.JSON(http.StatusOK, gin.H{
		"qr_code":      user.PersonalQRCode,
		"display_name": user.DisplayName,
		"username":     user.Username,
	})
}

// CreateTransferRequest は送金リクエストを作成
// POST /api/transfer-requests
func (c *TransferRequestController) CreateTransferRequest(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// リクエストボディ解析
	var req struct {
		ToUserID       string `json:"to_user_id" binding:"required"`
		Amount         int64  `json:"amount" binding:"required,gt=0"`
		Message        string `json:"message"`
		IdempotencyKey string `json:"idempotency_key" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// UUID変換
	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_user_id"})
		return
	}

	// ユースケース実行
	resp, err := c.transferRequestUC.CreateTransferRequest(&inputport.CreateTransferRequestRequest{
		FromUserID:     userID.(uuid.UUID),
		ToUserID:       toUserID,
		Amount:         req.Amount,
		Message:        req.Message,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusCreated, c.presenter.PresentCreateTransferRequest(resp))
}

// ApproveTransferRequest は送金リクエストを承認
// POST /api/transfer-requests/:id/approve
func (c *TransferRequestController) ApproveTransferRequest(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// パスパラメータ取得
	requestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request_id"})
		return
	}

	// ユースケース実行
	resp, err := c.transferRequestUC.ApproveTransferRequest(&inputport.ApproveTransferRequestRequest{
		RequestID: requestID,
		UserID:    userID.(uuid.UUID),
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentApproveTransferRequest(resp))
}

// RejectTransferRequest は送金リクエストを拒否
// POST /api/transfer-requests/:id/reject
func (c *TransferRequestController) RejectTransferRequest(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// パスパラメータ取得
	requestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request_id"})
		return
	}

	// ユースケース実行
	resp, err := c.transferRequestUC.RejectTransferRequest(&inputport.RejectTransferRequestRequest{
		RequestID: requestID,
		UserID:    userID.(uuid.UUID),
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentRejectTransferRequest(resp))
}

// CancelTransferRequest は送金リクエストをキャンセル
// DELETE /api/transfer-requests/:id
func (c *TransferRequestController) CancelTransferRequest(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// パスパラメータ取得
	requestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request_id"})
		return
	}

	// ユースケース実行
	resp, err := c.transferRequestUC.CancelTransferRequest(&inputport.CancelTransferRequestRequest{
		RequestID: requestID,
		UserID:    userID.(uuid.UUID),
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentCancelTransferRequest(resp))
}

// GetPendingRequests は受取人宛の承認待ちリクエスト一覧を取得
// GET /api/transfer-requests/pending
func (c *TransferRequestController) GetPendingRequests(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// クエリパラメータ
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	// ユースケース実行
	resp, err := c.transferRequestUC.GetPendingRequests(&inputport.GetPendingTransferRequestsRequest{
		ToUserID: userID.(uuid.UUID),
		Offset:   offset,
		Limit:    limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentGetPendingRequests(resp))
}

// GetSentRequests は送信者が送った送金リクエスト一覧を取得
// GET /api/transfer-requests/sent
func (c *TransferRequestController) GetSentRequests(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// クエリパラメータ
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	// ユースケース実行
	resp, err := c.transferRequestUC.GetSentRequests(&inputport.GetSentTransferRequestsRequest{
		FromUserID: userID.(uuid.UUID),
		Offset:     offset,
		Limit:      limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentGetSentRequests(resp))
}

// GetRequestDetail は送金リクエスト詳細を取得
// GET /api/transfer-requests/:id
func (c *TransferRequestController) GetRequestDetail(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// パスパラメータ取得
	requestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request_id"})
		return
	}

	// ユースケース実行
	resp, err := c.transferRequestUC.GetRequestDetail(&inputport.GetTransferRequestDetailRequest{
		RequestID: requestID,
		UserID:    userID.(uuid.UUID),
	})
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentGetRequestDetail(resp))
}

// GetPendingRequestCount は受取人宛の承認待ちリクエスト数を取得
// GET /api/transfer-requests/pending/count
func (c *TransferRequestController) GetPendingRequestCount(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// ユースケース実行
	resp, err := c.transferRequestUC.GetPendingRequestCount(&inputport.GetPendingRequestCountRequest{
		ToUserID: userID.(uuid.UUID),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, gin.H{
		"count": resp.Count,
	})
}
