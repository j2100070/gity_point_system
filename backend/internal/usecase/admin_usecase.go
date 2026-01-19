package usecase

import (
	"errors"
	"fmt"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminUseCase は管理者機能に関するユースケース
type AdminUseCase struct {
	db              *gorm.DB
	userRepo        domain.UserRepository
	transactionRepo domain.TransactionRepository
}

// NewAdminUseCase は新しいAdminUseCaseを作成
func NewAdminUseCase(
	db *gorm.DB,
	userRepo domain.UserRepository,
	transactionRepo domain.TransactionRepository,
) *AdminUseCase {
	return &AdminUseCase{
		db:              db,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
	}
}

// GrantPointsRequest はポイント付与リクエスト
type GrantPointsRequest struct {
	AdminID     uuid.UUID
	TargetUserID uuid.UUID
	Amount      int64
	Description string
}

// GrantPointsResponse はポイント付与レスポンス
type GrantPointsResponse struct {
	Transaction *domain.Transaction
	User        *domain.User
}

// GrantPoints は管理者がユーザーにポイントを付与
//
// セキュリティ:
// - 管理者権限チェック必須
// - トランザクションで原子性を保証
// - 監査ログに記録（実装推奨）
func (u *AdminUseCase) GrantPoints(req *GrantPointsRequest) (*GrantPointsResponse, error) {
	// 管理者確認
	admin, err := u.userRepo.FindByID(req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	if !admin.IsAdmin() {
		return nil, errors.New("unauthorized: admin role required")
	}

	// 対象ユーザー確認
	targetUser, err := u.userRepo.FindByID(req.TargetUserID)
	if err != nil {
		return nil, errors.New("target user not found")
	}

	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	var transaction *domain.Transaction

	err = u.db.Transaction(func(tx *gorm.DB) error {
		// ポイント加算
		if err := u.userRepo.UpdateBalanceWithLock(tx, req.TargetUserID, req.Amount, false); err != nil {
			return fmt.Errorf("failed to add points: %w", err)
		}

		// トランザクション記録作成
		transaction, err = domain.NewAdminGrant(req.TargetUserID, req.Amount, req.Description, req.AdminID)
		if err != nil {
			return err
		}

		if err := u.transactionRepo.Create(tx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// TODO: 監査ログに記録
		// auditLog := &domain.AuditLog{
		// 	AdminUserID:  req.AdminID,
		// 	TargetUserID: &req.TargetUserID,
		// 	Action:       "grant_points",
		// 	Details:      map[string]interface{}{"amount": req.Amount, "description": req.Description},
		// }
		// u.auditLogRepo.Create(tx, auditLog)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 最新のユーザー情報取得
	targetUser, _ = u.userRepo.FindByID(req.TargetUserID)

	return &GrantPointsResponse{
		Transaction: transaction,
		User:        targetUser,
	}, nil
}

// DeductPointsRequest はポイント減算リクエスト
type DeductPointsRequest struct {
	AdminID      uuid.UUID
	TargetUserID uuid.UUID
	Amount       int64
	Description  string
}

// DeductPointsResponse はポイント減算レスポンス
type DeductPointsResponse struct {
	Transaction *domain.Transaction
	User        *domain.User
}

// DeductPoints は管理者がユーザーからポイントを減算
func (u *AdminUseCase) DeductPoints(req *DeductPointsRequest) (*DeductPointsResponse, error) {
	// 管理者確認
	admin, err := u.userRepo.FindByID(req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	if !admin.IsAdmin() {
		return nil, errors.New("unauthorized: admin role required")
	}

	// 対象ユーザー確認
	targetUser, err := u.userRepo.FindByID(req.TargetUserID)
	if err != nil {
		return nil, errors.New("target user not found")
	}

	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	if targetUser.Balance < req.Amount {
		return nil, errors.New("insufficient balance")
	}

	var transaction *domain.Transaction

	err = u.db.Transaction(func(tx *gorm.DB) error {
		// ポイント減算
		if err := u.userRepo.UpdateBalanceWithLock(tx, req.TargetUserID, req.Amount, true); err != nil {
			return fmt.Errorf("failed to deduct points: %w", err)
		}

		// トランザクション記録作成
		transaction, err = domain.NewAdminDeduct(req.TargetUserID, req.Amount, req.Description, req.AdminID)
		if err != nil {
			return err
		}

		if err := u.transactionRepo.Create(tx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// TODO: 監査ログに記録

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 最新のユーザー情報取得
	targetUser, _ = u.userRepo.FindByID(req.TargetUserID)

	return &DeductPointsResponse{
		Transaction: transaction,
		User:        targetUser,
	}, nil
}

// ListAllUsersRequest は全ユーザー一覧取得リクエスト
type ListAllUsersRequest struct {
	AdminID uuid.UUID
	Offset  int
	Limit   int
}

// ListAllUsersResponse は全ユーザー一覧取得レスポンス
type ListAllUsersResponse struct {
	Users []*domain.User
	Total int64
}

// ListAllUsers は全ユーザー一覧を取得（管理者用）
func (u *AdminUseCase) ListAllUsers(req *ListAllUsersRequest) (*ListAllUsersResponse, error) {
	// 管理者確認
	admin, err := u.userRepo.FindByID(req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	if !admin.IsAdmin() {
		return nil, errors.New("unauthorized: admin role required")
	}

	users, err := u.userRepo.List(req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	total, err := u.userRepo.Count()
	if err != nil {
		return nil, err
	}

	return &ListAllUsersResponse{
		Users: users,
		Total: total,
	}, nil
}

// ListAllTransactionsRequest は全トランザクション一覧取得リクエスト
type ListAllTransactionsRequest struct {
	AdminID uuid.UUID
	Offset  int
	Limit   int
}

// ListAllTransactionsResponse は全トランザクション一覧取得レスポンス
type ListAllTransactionsResponse struct {
	Transactions []*domain.Transaction
}

// ListAllTransactions は全トランザクション一覧を取得（管理者用）
func (u *AdminUseCase) ListAllTransactions(req *ListAllTransactionsRequest) (*ListAllTransactionsResponse, error) {
	// 管理者確認
	admin, err := u.userRepo.FindByID(req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	if !admin.IsAdmin() {
		return nil, errors.New("unauthorized: admin role required")
	}

	transactions, err := u.transactionRepo.ListAll(req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	return &ListAllTransactionsResponse{Transactions: transactions}, nil
}

// UpdateUserRoleRequest はユーザー役割変更リクエスト
type UpdateUserRoleRequest struct {
	AdminID      uuid.UUID
	TargetUserID uuid.UUID
	NewRole      domain.UserRole
}

// UpdateUserRoleResponse はユーザー役割変更レスポンス
type UpdateUserRoleResponse struct {
	User *domain.User
}

// UpdateUserRole はユーザーの役割を変更
func (u *AdminUseCase) UpdateUserRole(req *UpdateUserRoleRequest) (*UpdateUserRoleResponse, error) {
	// 管理者確認
	admin, err := u.userRepo.FindByID(req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	if !admin.IsAdmin() {
		return nil, errors.New("unauthorized: admin role required")
	}

	// 対象ユーザー取得
	targetUser, err := u.userRepo.FindByID(req.TargetUserID)
	if err != nil {
		return nil, errors.New("target user not found")
	}

	// 役割更新
	if err := targetUser.UpdateRole(req.NewRole); err != nil {
		return nil, err
	}

	// 楽観的ロックで更新
	success, err := u.userRepo.Update(targetUser)
	if err != nil {
		return nil, err
	}

	if !success {
		return nil, errors.New("update failed due to version conflict")
	}

	// TODO: 監査ログに記録

	return &UpdateUserRoleResponse{User: targetUser}, nil
}

// DeactivateUserRequest はユーザー無効化リクエスト
type DeactivateUserRequest struct {
	AdminID      uuid.UUID
	TargetUserID uuid.UUID
}

// DeactivateUserResponse はユーザー無効化レスポンス
type DeactivateUserResponse struct {
	User *domain.User
}

// DeactivateUser はユーザーを無効化
func (u *AdminUseCase) DeactivateUser(req *DeactivateUserRequest) (*DeactivateUserResponse, error) {
	// 管理者確認
	admin, err := u.userRepo.FindByID(req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	if !admin.IsAdmin() {
		return nil, errors.New("unauthorized: admin role required")
	}

	// 対象ユーザー取得
	targetUser, err := u.userRepo.FindByID(req.TargetUserID)
	if err != nil {
		return nil, errors.New("target user not found")
	}

	// 無効化
	targetUser.Deactivate()

	// 更新
	success, err := u.userRepo.Update(targetUser)
	if err != nil {
		return nil, err
	}

	if !success {
		return nil, errors.New("update failed due to version conflict")
	}

	// TODO: 監査ログに記録

	return &DeactivateUserResponse{User: targetUser}, nil
}
