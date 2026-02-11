package presenter

import (
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// TransferRequestPresenter は送金リクエスト機能のプレゼンター
type TransferRequestPresenter struct{}

// NewTransferRequestPresenter は新しいTransferRequestPresenterを作成
func NewTransferRequestPresenter() *TransferRequestPresenter {
	return &TransferRequestPresenter{}
}

// TransferRequestResponse は送金リクエストのレスポンス
type TransferRequestResponse struct {
	ID             uuid.UUID  `json:"id"`
	FromUserID     uuid.UUID  `json:"from_user_id"`
	ToUserID       uuid.UUID  `json:"to_user_id"`
	Amount         int64      `json:"amount"`
	Message        string     `json:"message"`
	Status         string     `json:"status"`
	ExpiresAt      time.Time  `json:"expires_at"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectedAt     *time.Time `json:"rejected_at,omitempty"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`
	TransactionID  *uuid.UUID `json:"transaction_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TransferRequestInfoResponse は送金リクエスト情報のレスポンス
type TransferRequestInfoResponse struct {
	TransferRequest TransferRequestResponse `json:"transfer_request"`
	FromUser        UserResponse            `json:"from_user"`
	ToUser          UserResponse            `json:"to_user"`
}

// PresentCreateTransferRequest は送金リクエスト作成レスポンスを生成
func (p *TransferRequestPresenter) PresentCreateTransferRequest(resp *inputport.CreateTransferRequestResponse) map[string]interface{} {
	return map[string]interface{}{
		"transfer_request": p.toTransferRequestResponse(resp.TransferRequest),
		"from_user":        p.toUserResponse(resp.FromUser),
		"to_user":          p.toUserResponse(resp.ToUser),
	}
}

// PresentApproveTransferRequest は送金リクエスト承認レスポンスを生成
func (p *TransferRequestPresenter) PresentApproveTransferRequest(resp *inputport.ApproveTransferRequestResponse) map[string]interface{} {
	return map[string]interface{}{
		"transfer_request": p.toTransferRequestResponse(resp.TransferRequest),
		"transaction": TransactionResponse{
			ID:              resp.Transaction.ID,
			FromUserID:      resp.Transaction.FromUserID,
			ToUserID:        resp.Transaction.ToUserID,
			Amount:          resp.Transaction.Amount,
			TransactionType: string(resp.Transaction.TransactionType),
			Status:          string(resp.Transaction.Status),
			Description:     resp.Transaction.Description,
			CreatedAt:       resp.Transaction.CreatedAt,
		},
		"from_user": p.toUserResponse(resp.FromUser),
		"to_user":   p.toUserResponse(resp.ToUser),
	}
}

// PresentRejectTransferRequest は送金リクエスト拒否レスポンスを生成
func (p *TransferRequestPresenter) PresentRejectTransferRequest(resp *inputport.RejectTransferRequestResponse) map[string]interface{} {
	return map[string]interface{}{
		"transfer_request": p.toTransferRequestResponse(resp.TransferRequest),
	}
}

// PresentCancelTransferRequest は送金リクエストキャンセルレスポンスを生成
func (p *TransferRequestPresenter) PresentCancelTransferRequest(resp *inputport.CancelTransferRequestResponse) map[string]interface{} {
	return map[string]interface{}{
		"transfer_request": p.toTransferRequestResponse(resp.TransferRequest),
	}
}

// PresentGetPendingRequests は承認待ちリクエスト一覧レスポンスを生成
func (p *TransferRequestPresenter) PresentGetPendingRequests(resp *inputport.GetPendingTransferRequestsResponse) map[string]interface{} {
	requests := make([]TransferRequestInfoResponse, 0, len(resp.Requests))
	for _, info := range resp.Requests {
		requests = append(requests, TransferRequestInfoResponse{
			TransferRequest: p.toTransferRequestResponse(info.TransferRequest),
			FromUser:        p.toUserResponse(info.FromUser),
			ToUser:          p.toUserResponse(info.ToUser),
		})
	}

	return map[string]interface{}{
		"requests": requests,
	}
}

// PresentGetSentRequests は送信済みリクエスト一覧レスポンスを生成
func (p *TransferRequestPresenter) PresentGetSentRequests(resp *inputport.GetSentTransferRequestsResponse) map[string]interface{} {
	requests := make([]TransferRequestInfoResponse, 0, len(resp.Requests))
	for _, info := range resp.Requests {
		requests = append(requests, TransferRequestInfoResponse{
			TransferRequest: p.toTransferRequestResponse(info.TransferRequest),
			FromUser:        p.toUserResponse(info.FromUser),
			ToUser:          p.toUserResponse(info.ToUser),
		})
	}

	return map[string]interface{}{
		"requests": requests,
	}
}

// PresentGetRequestDetail は送金リクエスト詳細レスポンスを生成
func (p *TransferRequestPresenter) PresentGetRequestDetail(resp *inputport.GetTransferRequestDetailResponse) map[string]interface{} {
	return map[string]interface{}{
		"transfer_request": p.toTransferRequestResponse(resp.TransferRequest),
		"from_user":        p.toUserResponse(resp.FromUser),
		"to_user":          p.toUserResponse(resp.ToUser),
	}
}

// toTransferRequestResponse はTransferRequestエンティティをレスポンスに変換
func (p *TransferRequestPresenter) toTransferRequestResponse(tr *entities.TransferRequest) TransferRequestResponse {
	return TransferRequestResponse{
		ID:            tr.ID,
		FromUserID:    tr.FromUserID,
		ToUserID:      tr.ToUserID,
		Amount:        tr.Amount,
		Message:       tr.Message,
		Status:        string(tr.Status),
		ExpiresAt:     tr.ExpiresAt,
		ApprovedAt:    tr.ApprovedAt,
		RejectedAt:    tr.RejectedAt,
		CancelledAt:   tr.CancelledAt,
		TransactionID: tr.TransactionID,
		CreatedAt:     tr.CreatedAt,
		UpdatedAt:     tr.UpdatedAt,
	}
}

// toUserResponse はUserエンティティをレスポンスに変換
func (p *TransferRequestPresenter) toUserResponse(user *entities.User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Balance:     user.Balance,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
