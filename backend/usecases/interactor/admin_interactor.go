package interactor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// AdminInteractor は管理者機能のユースケース実装
type AdminInteractor struct {
	txManager       repository.TransactionManager
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	idempotencyRepo repository.IdempotencyKeyRepository
	pointBatchRepo  repository.PointBatchRepository
	analyticsDS     dsmysql.AnalyticsDataSource
	logger          entities.Logger
}

// NewAdminInteractor は新しいAdminInteractorを作成
func NewAdminInteractor(
	txManager repository.TransactionManager,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	idempotencyRepo repository.IdempotencyKeyRepository,
	pointBatchRepo repository.PointBatchRepository,
	analyticsDS dsmysql.AnalyticsDataSource,
	logger entities.Logger,
) inputport.AdminInputPort {
	return &AdminInteractor{
		txManager:       txManager,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		idempotencyRepo: idempotencyRepo,
		pointBatchRepo:  pointBatchRepo,
		analyticsDS:     analyticsDS,
		logger:          logger,
	}
}

// GrantPoints はユーザーにポイントを付与
func (i *AdminInteractor) GrantPoints(ctx context.Context, req *inputport.GrantPointsRequest) (*inputport.GrantPointsResponse, error) {
	i.logger.Info("Admin granting points",
		entities.NewField("admin_id", req.AdminID),
		entities.NewField("user_id", req.UserID),
		entities.NewField("amount", req.Amount))

	// 金額検証
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	// 管理者権限チェック
	admin, err := i.userRepo.Read(ctx, req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}
	if admin.Role != "admin" {
		return nil, errors.New("unauthorized: admin role required")
	}

	// 冪等性チェック
	existingKey, err := i.idempotencyRepo.ReadByKey(ctx, req.IdempotencyKey)
	if err == nil && existingKey != nil && existingKey.TransactionID != nil {
		i.logger.Info("Idempotency key already used", entities.NewField("key", req.IdempotencyKey))
		existingTx, _ := i.transactionRepo.Read(ctx, *existingKey.TransactionID)
		user, _ := i.userRepo.Read(ctx, req.UserID)
		return &inputport.GrantPointsResponse{
			Transaction: existingTx,
			User:        user,
		}, nil
	}

	var user *entities.User
	var transaction *entities.Transaction

	// トランザクション実行
	err = i.txManager.Do(ctx, func(txCtx context.Context) error {

		// ユーザー取得
		var err error
		user, err = i.userRepo.Read(txCtx, req.UserID)
		if err != nil {
			return errors.New("user not found")
		}

		if !user.IsActive {
			return errors.New("user is not active")
		}

		// ポイント付与（残高更新はロック付きで実行）
		if err := i.userRepo.UpdateBalanceWithLock(txCtx, req.UserID, req.Amount, false); err != nil {
			return err
		}

		// ユーザーの Balance を更新
		user.Balance += req.Amount

		// 取引記録作成（システムから付与）
		transaction, err = entities.NewAdminGrant(
			req.UserID,
			req.Amount,
			fmt.Sprintf("Admin grant: %s", req.Description),
			req.AdminID,
		)
		if err != nil {
			return err
		}

		if err := i.transactionRepo.Create(txCtx, transaction); err != nil {
			return err
		}

		// ポイントバッチ作成
		batch := entities.NewPointBatch(req.UserID, req.Amount, entities.PointBatchSourceAdminGrant, &transaction.ID, time.Now())
		if err := i.pointBatchRepo.Create(txCtx, batch); err != nil {
			return fmt.Errorf("failed to create point batch: %w", err)
		}

		// 冪等性キー保存
		idempotencyKey := entities.NewIdempotencyKey(req.IdempotencyKey, req.AdminID)
		idempotencyKey.TransactionID = &transaction.ID
		idempotencyKey.Status = "completed"
		if err := i.idempotencyRepo.Create(ctx, idempotencyKey); err != nil {
			i.logger.Warn("Failed to save idempotency key", entities.NewField("error", err))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	i.logger.Info("Points granted successfully",
		entities.NewField("user_id", req.UserID),
		entities.NewField("amount", req.Amount))

	return &inputport.GrantPointsResponse{
		Transaction: transaction,
		User:        user,
	}, nil
}

// DeductPoints はユーザーからポイントを減算
func (i *AdminInteractor) DeductPoints(ctx context.Context, req *inputport.DeductPointsRequest) (*inputport.DeductPointsResponse, error) {
	i.logger.Info("Admin deducting points",
		entities.NewField("admin_id", req.AdminID),
		entities.NewField("user_id", req.UserID),
		entities.NewField("amount", req.Amount))

	// 金額検証
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	// 管理者権限チェック
	admin, err := i.userRepo.Read(ctx, req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}
	if admin.Role != "admin" {
		return nil, errors.New("unauthorized: admin role required")
	}

	// 冪等性チェック
	existingKey, err := i.idempotencyRepo.ReadByKey(ctx, req.IdempotencyKey)
	if err == nil && existingKey != nil && existingKey.TransactionID != nil {
		i.logger.Info("Idempotency key already used", entities.NewField("key", req.IdempotencyKey))
		existingTx, _ := i.transactionRepo.Read(ctx, *existingKey.TransactionID)
		user, _ := i.userRepo.Read(ctx, req.UserID)
		return &inputport.DeductPointsResponse{
			Transaction: existingTx,
			User:        user,
		}, nil
	}

	var user *entities.User
	var transaction *entities.Transaction

	// トランザクション実行
	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		// ユーザー取得
		var err error
		user, err = i.userRepo.Read(txCtx, req.UserID)
		if err != nil {
			return errors.New("user not found")
		}

		if !user.IsActive {
			return errors.New("user is not active")
		}

		// 残高チェック
		if user.Balance < req.Amount {
			return errors.New("insufficient balance")
		}

		// ポイント減算（残高更新はロック付きで実行）
		if err := i.userRepo.UpdateBalanceWithLock(txCtx, req.UserID, req.Amount, true); err != nil {
			return err
		}

		// ユーザーの Balance を更新
		user.Balance -= req.Amount

		// 取引記録作成（システムへの減算）
		transaction, err = entities.NewAdminDeduct(
			req.UserID,
			req.Amount,
			fmt.Sprintf("Admin deduct: %s", req.Description),
			req.AdminID,
		)
		if err != nil {
			return err
		}

		if err := i.transactionRepo.Create(txCtx, transaction); err != nil {
			return err
		}

		// 冪等性キー保存
		idempotencyKey := entities.NewIdempotencyKey(req.IdempotencyKey, req.AdminID)
		idempotencyKey.TransactionID = &transaction.ID
		idempotencyKey.Status = "completed"
		if err := i.idempotencyRepo.Create(ctx, idempotencyKey); err != nil {
			i.logger.Warn("Failed to save idempotency key", entities.NewField("error", err))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	i.logger.Info("Points deducted successfully",
		entities.NewField("user_id", req.UserID),
		entities.NewField("amount", req.Amount))

	return &inputport.DeductPointsResponse{
		Transaction: transaction,
		User:        user,
	}, nil
}

// ListAllUsers はすべてのユーザー一覧を取得
func (i *AdminInteractor) ListAllUsers(ctx context.Context, req *inputport.ListAllUsersRequest) (*inputport.ListAllUsersResponse, error) {
	users, err := i.userRepo.ReadList(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	total, err := i.userRepo.Count(ctx)
	if err != nil {
		total = int64(len(users))
	}

	return &inputport.ListAllUsersResponse{
		Users: users,
		Total: int(total),
	}, nil
}

// ListAllTransactions はすべての取引履歴を取得
func (i *AdminInteractor) ListAllTransactions(ctx context.Context, req *inputport.ListAllTransactionsRequest) (*inputport.ListAllTransactionsResponse, error) {
	transactions, err := i.transactionRepo.ReadListAll(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	// 各トランザクションにユーザー情報を付与
	transactionsWithUsers := make([]*inputport.TransactionWithUsers, 0, len(transactions))
	for _, tx := range transactions {
		txWithUsers := &inputport.TransactionWithUsers{
			Transaction: tx,
		}

		// 送信者情報を取得
		if tx.FromUserID != nil {
			fromUser, err := i.userRepo.Read(ctx, *tx.FromUserID)
			if err == nil {
				txWithUsers.FromUser = fromUser
			}
		}

		// 受信者情報を取得
		if tx.ToUserID != nil {
			toUser, err := i.userRepo.Read(ctx, *tx.ToUserID)
			if err == nil {
				txWithUsers.ToUser = toUser
			}
		}

		transactionsWithUsers = append(transactionsWithUsers, txWithUsers)
	}

	return &inputport.ListAllTransactionsResponse{
		Transactions: transactionsWithUsers,
		Total:        len(transactions),
	}, nil
}

// UpdateUserRole はユーザーの役割を更新
func (i *AdminInteractor) UpdateUserRole(ctx context.Context, req *inputport.UpdateUserRoleRequest) (*inputport.UpdateUserRoleResponse, error) {
	i.logger.Info("Admin updating user role",
		entities.NewField("admin_id", req.AdminID),
		entities.NewField("user_id", req.UserID),
		entities.NewField("role", req.Role))

	// 管理者権限チェック
	admin, err := i.userRepo.Read(ctx, req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}
	if admin.Role != "admin" {
		return nil, errors.New("unauthorized: admin role required")
	}

	// 役割検証
	if req.Role != "user" && req.Role != "admin" {
		return nil, errors.New("invalid role: must be 'user' or 'admin'")
	}

	// 楽観ロック競合時リトライ（最大3回）
	const maxRetries = 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		user, err := i.userRepo.Read(ctx, req.UserID)
		if err != nil {
			return nil, errors.New("user not found")
		}

		if err := user.UpdateRole(entities.UserRole(req.Role)); err != nil {
			return nil, err
		}

		updated, err := i.userRepo.Update(ctx, user)
		if err != nil {
			return nil, err
		}
		if updated {
			i.logger.Info("User role updated successfully",
				entities.NewField("user_id", req.UserID),
				entities.NewField("role", req.Role))
			return &inputport.UpdateUserRoleResponse{User: user}, nil
		}

		i.logger.Info("Optimistic lock conflict, retrying",
			entities.NewField("attempt", attempt+1))
	}

	return nil, errors.New("update conflict: please retry later")
}

// DeactivateUser はユーザーを無効化
func (i *AdminInteractor) DeactivateUser(ctx context.Context, req *inputport.DeactivateUserRequest) (*inputport.DeactivateUserResponse, error) {
	i.logger.Info("Admin deactivating user",
		entities.NewField("admin_id", req.AdminID),
		entities.NewField("user_id", req.UserID))

	// 管理者権限チェック
	admin, err := i.userRepo.Read(ctx, req.AdminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}
	if admin.Role != "admin" {
		return nil, errors.New("unauthorized: admin role required")
	}

	// 自分自身を無効化しようとしていないかチェック
	if req.AdminID == req.UserID {
		return nil, errors.New("cannot deactivate yourself")
	}

	// 楽観ロック競合時リトライ（最大3回）
	const maxRetries = 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		user, err := i.userRepo.Read(ctx, req.UserID)
		if err != nil {
			return nil, errors.New("user not found")
		}

		user.Deactivate()

		updated, err := i.userRepo.Update(ctx, user)
		if err != nil {
			return nil, err
		}
		if updated {
			i.logger.Info("User deactivated successfully", entities.NewField("user_id", req.UserID))
			return &inputport.DeactivateUserResponse{User: user}, nil
		}

		i.logger.Info("Optimistic lock conflict, retrying",
			entities.NewField("attempt", attempt+1))
	}

	return nil, errors.New("update conflict: please retry later")
}

// GetAnalytics は分析データを取得
func (i *AdminInteractor) GetAnalytics(ctx context.Context, req *inputport.GetAnalyticsRequest) (*inputport.GetAnalyticsResponse, error) {
	i.logger.Info("Getting analytics data", entities.NewField("days", req.Days))

	// 日数のバリデーション
	days := req.Days
	if days != 7 && days != 30 && days != 90 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	summary, err := i.analyticsDS.GetUserBalanceSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance summary: %w", err)
	}

	monthlyIssued, err := i.analyticsDS.GetMonthlyIssuedPoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly issued points: %w", err)
	}

	monthlyTxCount, err := i.analyticsDS.GetMonthlyTransactionCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly transaction count: %w", err)
	}

	topHoldersResult, err := i.analyticsDS.GetTopHolders(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top holders: %w", err)
	}

	dailyStatsResult, err := i.analyticsDS.GetDailyStats(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}

	typeBreakdownResult, err := i.analyticsDS.GetTransactionTypeBreakdown(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction type breakdown: %w", err)
	}

	// レスポンス組み立て
	analyticsSummary := &inputport.AnalyticsSummary{
		TotalPointsInCirculation: summary.TotalBalance,
		AverageBalance:           summary.AverageBalance,
		PointsIssuedThisMonth:    monthlyIssued,
		TransactionsThisMonth:    monthlyTxCount,
		ActiveUsers:              summary.ActiveUsers,
	}

	topHolders := make([]*inputport.TopHolder, 0, len(topHoldersResult))
	for _, h := range topHoldersResult {
		userID, _ := uuid.Parse(h.ID)
		pct := float64(0)
		if summary.TotalBalance > 0 {
			pct = float64(h.Balance) / float64(summary.TotalBalance) * 100
		}
		topHolders = append(topHolders, &inputport.TopHolder{
			UserID:      userID,
			Username:    h.Username,
			DisplayName: h.DisplayName,
			Balance:     h.Balance,
			Percentage:  pct,
		})
	}

	dailyStats := make([]*inputport.DailyStat, 0, len(dailyStatsResult))
	for _, d := range dailyStatsResult {
		dailyStats = append(dailyStats, &inputport.DailyStat{
			Date:        d.Date,
			Issued:      d.Issued,
			Consumed:    d.Consumed,
			Transferred: d.Transferred,
		})
	}

	typeBreakdown := make([]*inputport.TransactionTypeBreakdown, 0, len(typeBreakdownResult))
	for _, t := range typeBreakdownResult {
		typeBreakdown = append(typeBreakdown, &inputport.TransactionTypeBreakdown{
			Type:        t.Type,
			Count:       t.Count,
			TotalAmount: t.TotalAmount,
		})
	}

	return &inputport.GetAnalyticsResponse{
		Summary:                  analyticsSummary,
		TopHolders:               topHolders,
		DailyStats:               dailyStats,
		TransactionTypeBreakdown: typeBreakdown,
	}, nil
}
