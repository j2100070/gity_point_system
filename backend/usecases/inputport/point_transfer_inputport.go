package inputport

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// PointTransferInputPort はポイント転送のユースケースインターフェース
type PointTransferInputPort interface {
	// Transfer はポイントを転送
	Transfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error)

	// GetTransactionHistory はトランザクション履歴を取得
	GetTransactionHistory(ctx context.Context, req *GetTransactionHistoryRequest) (*GetTransactionHistoryResponse, error)

	// GetBalance は残高を取得
	GetBalance(ctx context.Context, req *GetBalanceRequest) (*GetBalanceResponse, error)
}

// TransferRequest はポイント転送リクエスト
type TransferRequest struct {
	FromUserID     uuid.UUID
	ToUserID       uuid.UUID
	Amount         int64
	IdempotencyKey string // 冪等性キー（クライアントが生成）
	Description    string
}

// TransferResponse はポイント転送レスポンス
type TransferResponse struct {
	Transaction *entities.Transaction
	FromUser    *entities.User
	ToUser      *entities.User
}

// GetTransactionHistoryRequest はトランザクション履歴取得リクエスト
type GetTransactionHistoryRequest struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}

// TransactionWithUsersForHistory はユーザー情報付きトランザクション（履歴用）
type TransactionWithUsersForHistory struct {
	Transaction *entities.Transaction
	FromUser    *entities.User
	ToUser      *entities.User
}

// GetTransactionHistoryResponse はトランザクション履歴取得レスポンス
type GetTransactionHistoryResponse struct {
	Transactions []*TransactionWithUsersForHistory
	Total        int64
}

// GetBalanceRequest は残高取得リクエスト
type GetBalanceRequest struct {
	UserID uuid.UUID
}

// GetBalanceResponse は残高取得レスポンス
type GetBalanceResponse struct {
	Balance int64
	User    *entities.User
}
