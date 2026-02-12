package inputport

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// AdminInputPort は管理者機能のユースケースインターフェース
type AdminInputPort interface {
	// GrantPoints はユーザーにポイントを付与
	GrantPoints(ctx context.Context, req *GrantPointsRequest) (*GrantPointsResponse, error)

	// DeductPoints はユーザーからポイントを減算
	DeductPoints(ctx context.Context, req *DeductPointsRequest) (*DeductPointsResponse, error)

	// ListAllUsers はすべてのユーザー一覧を取得
	ListAllUsers(ctx context.Context, req *ListAllUsersRequest) (*ListAllUsersResponse, error)

	// ListAllTransactions はすべての取引履歴を取得
	ListAllTransactions(ctx context.Context, req *ListAllTransactionsRequest) (*ListAllTransactionsResponse, error)

	// UpdateUserRole はユーザーの役割を更新
	UpdateUserRole(ctx context.Context, req *UpdateUserRoleRequest) (*UpdateUserRoleResponse, error)

	// DeactivateUser はユーザーを無効化
	DeactivateUser(ctx context.Context, req *DeactivateUserRequest) (*DeactivateUserResponse, error)
}

// GrantPointsRequest はポイント付与リクエスト
type GrantPointsRequest struct {
	AdminID        uuid.UUID
	UserID         uuid.UUID
	Amount         int64
	Description    string
	IdempotencyKey string
}

// GrantPointsResponse はポイント付与レスポンス
type GrantPointsResponse struct {
	Transaction *entities.Transaction
	User        *entities.User
}

// DeductPointsRequest はポイント減算リクエスト
type DeductPointsRequest struct {
	AdminID        uuid.UUID
	UserID         uuid.UUID
	Amount         int64
	Description    string
	IdempotencyKey string
}

// DeductPointsResponse はポイント減算レスポンス
type DeductPointsResponse struct {
	Transaction *entities.Transaction
	User        *entities.User
}

// ListAllUsersRequest はユーザー一覧取得リクエスト
type ListAllUsersRequest struct {
	Offset int
	Limit  int
}

// ListAllUsersResponse はユーザー一覧取得レスポンス
type ListAllUsersResponse struct {
	Users []*entities.User
	Total int
}

// ListAllTransactionsRequest は取引履歴一覧取得リクエスト
type ListAllTransactionsRequest struct {
	Offset int
	Limit  int
}

// TransactionWithUsers はユーザー情報付きトランザクション
type TransactionWithUsers struct {
	Transaction *entities.Transaction
	FromUser    *entities.User
	ToUser      *entities.User
}

// ListAllTransactionsResponse は取引履歴一覧取得レスポンス
type ListAllTransactionsResponse struct {
	Transactions []*TransactionWithUsers
	Total        int
}

// UpdateUserRoleRequest はユーザー役割更新リクエスト
type UpdateUserRoleRequest struct {
	AdminID uuid.UUID
	UserID  uuid.UUID
	Role    string
}

// UpdateUserRoleResponse はユーザー役割更新レスポンス
type UpdateUserRoleResponse struct {
	User *entities.User
}

// DeactivateUserRequest はユーザー無効化リクエスト
type DeactivateUserRequest struct {
	AdminID uuid.UUID
	UserID  uuid.UUID
}

// DeactivateUserResponse はユーザー無効化レスポンス
type DeactivateUserResponse struct {
	User *entities.User
}
