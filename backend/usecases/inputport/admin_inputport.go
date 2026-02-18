package inputport

import (
	"context"
	"time"

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

	// GetAnalytics は分析データを取得
	GetAnalytics(ctx context.Context, req *GetAnalyticsRequest) (*GetAnalyticsResponse, error)
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
	Offset    int
	Limit     int
	Search    string // 名前・ユーザー名・IDで検索
	SortBy    string // created_at, balance, role, username, display_name
	SortOrder string // asc, desc
}

// ListAllUsersResponse はユーザー一覧取得レスポンス
type ListAllUsersResponse struct {
	Users []*entities.User
	Total int
}

// ListAllTransactionsRequest は取引履歴一覧取得リクエスト
type ListAllTransactionsRequest struct {
	Offset          int
	Limit           int
	TransactionType string // フィルタ: transfer, admin_grant, admin_deduct, system_grant, daily_bonus, etc.
	DateFrom        string // フィルタ: 開始日（YYYY-MM-DD）
	DateTo          string // フィルタ: 終了日（YYYY-MM-DD）
	SortBy          string // ソート列: created_at, amount
	SortOrder       string // ソート順: asc, desc
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

// GetAnalyticsRequest は分析データ取得リクエスト
type GetAnalyticsRequest struct {
	Days int // 日別統計の日数（7, 30, 90）
}

// GetAnalyticsResponse は分析データ取得レスポンス
type GetAnalyticsResponse struct {
	Summary                  *AnalyticsSummary
	TopHolders               []*TopHolder
	DailyStats               []*DailyStat
	TransactionTypeBreakdown []*TransactionTypeBreakdown
}

// AnalyticsSummary はKPIサマリー
type AnalyticsSummary struct {
	TotalPointsInCirculation int64
	AverageBalance           float64
	PointsIssuedThisMonth    int64
	TransactionsThisMonth    int64
	ActiveUsers              int64
}

// TopHolder はポイント保有上位ユーザー
type TopHolder struct {
	UserID      uuid.UUID
	Username    string
	DisplayName string
	Balance     int64
	Percentage  float64
}

// DailyStat は日別統計
type DailyStat struct {
	Date        time.Time
	Issued      int64
	Consumed    int64
	Transferred int64
}

// TransactionTypeBreakdown はトランザクション種別構成
type TransactionTypeBreakdown struct {
	Type        string
	Count       int64
	TotalAmount int64
}
