package inputport

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// TransferRequestInputPort は送金リクエスト機能のユースケースインターフェース
type TransferRequestInputPort interface {
	// CreateTransferRequest は送金リクエストを作成（QRスキャン時）
	CreateTransferRequest(ctx context.Context, req *CreateTransferRequestRequest) (*CreateTransferRequestResponse, error)

	// ApproveTransferRequest は送金リクエストを承認（受取人が承認）
	ApproveTransferRequest(ctx context.Context, req *ApproveTransferRequestRequest) (*ApproveTransferRequestResponse, error)

	// RejectTransferRequest は送金リクエストを拒否（受取人が拒否）
	RejectTransferRequest(ctx context.Context, req *RejectTransferRequestRequest) (*RejectTransferRequestResponse, error)

	// CancelTransferRequest は送金リクエストをキャンセル（送信者がキャンセル）
	CancelTransferRequest(ctx context.Context, req *CancelTransferRequestRequest) (*CancelTransferRequestResponse, error)

	// GetPendingRequests は受取人宛の承認待ちリクエスト一覧を取得
	GetPendingRequests(ctx context.Context, req *GetPendingTransferRequestsRequest) (*GetPendingTransferRequestsResponse, error)

	// GetSentRequests は送信者が送った送金リクエスト一覧を取得
	GetSentRequests(ctx context.Context, req *GetSentTransferRequestsRequest) (*GetSentTransferRequestsResponse, error)

	// GetRequestDetail は送金リクエスト詳細を取得
	GetRequestDetail(ctx context.Context, req *GetTransferRequestDetailRequest) (*GetTransferRequestDetailResponse, error)

	// GetPendingRequestCount は受取人宛の承認待ちリクエスト数を取得
	GetPendingRequestCount(ctx context.Context, req *GetPendingRequestCountRequest) (*GetPendingRequestCountResponse, error)
}

// CreateTransferRequestRequest は送金リクエスト作成リクエスト
type CreateTransferRequestRequest struct {
	FromUserID     uuid.UUID
	ToUserID       uuid.UUID
	Amount         int64
	Message        string
	IdempotencyKey string
}

// CreateTransferRequestResponse は送金リクエスト作成レスポンス
type CreateTransferRequestResponse struct {
	TransferRequest *entities.TransferRequest
	FromUser        *entities.User
	ToUser          *entities.User
}

// ApproveTransferRequestRequest は送金リクエスト承認リクエスト
type ApproveTransferRequestRequest struct {
	RequestID uuid.UUID
	UserID    uuid.UUID // 承認者（受取人）
}

// ApproveTransferRequestResponse は送金リクエスト承認レスポンス
type ApproveTransferRequestResponse struct {
	TransferRequest *entities.TransferRequest
	Transaction     *entities.Transaction
	FromUser        *entities.User
	ToUser          *entities.User
}

// RejectTransferRequestRequest は送金リクエスト拒否リクエスト
type RejectTransferRequestRequest struct {
	RequestID uuid.UUID
	UserID    uuid.UUID // 拒否者（受取人）
}

// RejectTransferRequestResponse は送金リクエスト拒否レスポンス
type RejectTransferRequestResponse struct {
	TransferRequest *entities.TransferRequest
}

// CancelTransferRequestRequest は送金リクエストキャンセルリクエスト
type CancelTransferRequestRequest struct {
	RequestID uuid.UUID
	UserID    uuid.UUID // キャンセル者（送信者）
}

// CancelTransferRequestResponse は送金リクエストキャンセルレスポンス
type CancelTransferRequestResponse struct {
	TransferRequest *entities.TransferRequest
}

// GetPendingTransferRequestsRequest は承認待ちリクエスト一覧取得リクエスト
type GetPendingTransferRequestsRequest struct {
	ToUserID uuid.UUID
	Offset   int
	Limit    int
}

// TransferRequestInfo は送金リクエスト情報
type TransferRequestInfo struct {
	TransferRequest *entities.TransferRequest
	FromUser        *entities.User
	ToUser          *entities.User
}

// GetPendingTransferRequestsResponse は承認待ちリクエスト一覧取得レスポンス
type GetPendingTransferRequestsResponse struct {
	Requests []*TransferRequestInfo
}

// GetSentTransferRequestsRequest は送信済みリクエスト一覧取得リクエスト
type GetSentTransferRequestsRequest struct {
	FromUserID uuid.UUID
	Offset     int
	Limit      int
}

// GetSentTransferRequestsResponse は送信済みリクエスト一覧取得レスポンス
type GetSentTransferRequestsResponse struct {
	Requests []*TransferRequestInfo
}

// GetTransferRequestDetailRequest は送金リクエスト詳細取得リクエスト
type GetTransferRequestDetailRequest struct {
	RequestID uuid.UUID
	UserID    uuid.UUID // 要求者（送信者または受取人）
}

// GetTransferRequestDetailResponse は送金リクエスト詳細取得レスポンス
type GetTransferRequestDetailResponse struct {
	TransferRequest *entities.TransferRequest
	FromUser        *entities.User
	ToUser          *entities.User
}

// GetPendingRequestCountRequest は承認待ちリクエスト数取得リクエスト
type GetPendingRequestCountRequest struct {
	ToUserID uuid.UUID
}

// GetPendingRequestCountResponse は承認待ちリクエスト数取得レスポンス
type GetPendingRequestCountResponse struct {
	Count int64
}
