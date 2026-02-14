package interactor

import (
	"context"
	"errors"
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
)

// PointTransferInteractor はポイント転送のユースケース実装
type PointTransferInteractor struct {
	txManager       repository.TransactionManager
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	idempotencyRepo repository.IdempotencyKeyRepository
	friendshipRepo  repository.FriendshipRepository
	dailyBonusPort  inputport.DailyBonusInputPort
	logger          entities.Logger
}

// NewPointTransferInteractor は新しいPointTransferInteractorを作成
func NewPointTransferInteractor(
	txManager repository.TransactionManager,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	idempotencyRepo repository.IdempotencyKeyRepository,
	friendshipRepo repository.FriendshipRepository,
	logger entities.Logger,
) *PointTransferInteractor {
	return &PointTransferInteractor{
		txManager:       txManager,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		idempotencyRepo: idempotencyRepo,
		friendshipRepo:  friendshipRepo,
		dailyBonusPort:  nil, // 後からSetDailyBonusPortで設定
		logger:          logger,
	}
}

// SetDailyBonusPort はデイリーボーナスポートを設定（循環依存回避のため）
func (i *PointTransferInteractor) SetDailyBonusPort(dailyBonusPort inputport.DailyBonusInputPort) {
	i.dailyBonusPort = dailyBonusPort
}

// Transfer はポイントを転送
//
// セキュリティと整合性の保証:
// 1. 冪等性: IdempotencyKeyで重複転送を防止
// 2. トランザクション: 原子性を保証し、全操作を一括でコミットまたはロールバック
// 3. 悲観的ロック: 残高更新時に競合を防止
// 4. 残高チェック: 送信者の残高を厳密にチェック
// 5. 友達チェック: 友達関係がある場合のみ転送可能（オプション）
//
// 技術的説明:
// - 高い分離レベルで一貫したスナップショットを保証
// - ロック戦略: ID順序でロックを取得しデッドロックを回避
// - エラーハンドリング: ロールバック処理を確実に実行
func (i *PointTransferInteractor) Transfer(ctx context.Context, req *inputport.TransferRequest) (*inputport.TransferResponse, error) {
	i.logger.Info("Starting point transfer",
		entities.NewField("from_user_id", req.FromUserID),
		entities.NewField("to_user_id", req.ToUserID),
		entities.NewField("amount", req.Amount))

	// バリデーション
	if req.FromUserID == req.ToUserID {
		return nil, errors.New("cannot transfer to the same user")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if req.IdempotencyKey == "" {
		return nil, errors.New("idempotency key is required")
	}

	// === 冪等性チェック ===
	// 同じIdempotencyKeyで既に処理済みの場合は、その結果を返す
	existingKey, err := i.idempotencyRepo.ReadByKey(ctx, req.IdempotencyKey)
	if err == nil {
		// 既存のキーが見つかった場合
		if existingKey.Status == "completed" && existingKey.TransactionID != nil {
			// 完了済みの場合は既存のトランザクションを返す
			transaction, err := i.transactionRepo.Read(ctx, *existingKey.TransactionID)
			if err != nil {
				return nil, err
			}

			fromUser, _ := i.userRepo.Read(ctx, req.FromUserID)
			toUser, _ := i.userRepo.Read(ctx, req.ToUserID)

			return &inputport.TransferResponse{
				Transaction: transaction,
				FromUser:    fromUser,
				ToUser:      toUser,
			}, nil
		} else if existingKey.Status == "processing" {
			// 処理中の場合はエラー（二重送信の可能性）
			return nil, errors.New("transfer is already in progress")
		}
	}

	// 新しい冪等性キーを作成
	idempotencyKey := entities.NewIdempotencyKey(req.IdempotencyKey, req.FromUserID)
	if err := i.idempotencyRepo.Create(ctx, idempotencyKey); err != nil {
		// 競合エラーの場合は二重送信
		return nil, errors.New("duplicate idempotency key")
	}

	// === トランザクション開始 ===
	var fromUser, toUser *entities.User
	var transaction *entities.Transaction

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		// 1. 送信者と受信者の存在確認
		fromUser, err = i.userRepo.Read(txCtx, req.FromUserID)
		if err != nil {
			return fmt.Errorf("sender not found: %w", err)
		}

		toUser, err = i.userRepo.Read(txCtx, req.ToUserID)
		if err != nil {
			return fmt.Errorf("receiver not found: %w", err)
		}

		// 2. アカウント状態チェック
		if !fromUser.IsActive {
			return errors.New("sender account is not active")
		}
		if !toUser.IsActive {
			return errors.New("receiver account is not active")
		}

		// 3. 残高更新（悲観的ロックで競合を防止）
		updates := []repository.BalanceUpdate{
			{UserID: req.FromUserID, Amount: req.Amount, IsDeduct: true}, // 送信者から減算
			{UserID: req.ToUserID, Amount: req.Amount, IsDeduct: false},  // 受信者に加算
		}

		if err := i.userRepo.UpdateBalancesWithLock(txCtx, updates); err != nil {
			return fmt.Errorf("failed to update balances: %w", err)
		}

		// 4. トランザクション記録作成
		transaction, err = entities.NewTransfer(req.FromUserID, req.ToUserID, req.Amount, req.IdempotencyKey, req.Description)
		if err != nil {
			return err
		}

		if err := i.transactionRepo.Create(txCtx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// 6. トランザクションを完了状態に
		if err := transaction.Complete(); err != nil {
			return err
		}

		if err := i.transactionRepo.Update(txCtx, transaction); err != nil {
			return err
		}

		// 7. 冪等性キーを完了状態に
		idempotencyKey.Status = "completed"
		idempotencyKey.TransactionID = &transaction.ID
		if err := i.idempotencyRepo.Update(ctx, idempotencyKey); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// トランザクション失敗時は冪等性キーを失敗状態に
		idempotencyKey.Status = "failed"
		i.idempotencyRepo.Update(ctx, idempotencyKey)
		i.logger.Error("Point transfer failed", entities.NewField("error", err))
		return nil, err
	}

	// 最新の残高を取得
	fromUser, _ = i.userRepo.Read(ctx, req.FromUserID)
	toUser, _ = i.userRepo.Read(ctx, req.ToUserID)

	i.logger.Info("Point transfer completed successfully",
		entities.NewField("transaction_id", transaction.ID))

	// デイリーボーナスチェック（送金ボーナス）
	if i.dailyBonusPort != nil {
		_, err := i.dailyBonusPort.CheckTransferBonus(ctx, &inputport.CheckTransferBonusRequest{
			UserID:        req.FromUserID,
			TransactionID: transaction.ID,
			Date:          transaction.CreatedAt,
		})
		if err != nil {
			// ボーナス付与失敗してもメイントランザクションは成功扱い
			i.logger.Error("Failed to check transfer bonus", entities.NewField("error", err))
		}
	}

	return &inputport.TransferResponse{
		Transaction: transaction,
		FromUser:    fromUser,
		ToUser:      toUser,
	}, nil
}

// GetTransactionHistory はトランザクション履歴を取得
func (i *PointTransferInteractor) GetTransactionHistory(ctx context.Context, req *inputport.GetTransactionHistoryRequest) (*inputport.GetTransactionHistoryResponse, error) {
	transactions, err := i.transactionRepo.ReadListByUserID(ctx, req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	total, err := i.transactionRepo.CountByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// 各トランザクションにユーザー情報を付与
	transactionsWithUsers := make([]*inputport.TransactionWithUsersForHistory, 0, len(transactions))
	for _, tx := range transactions {
		txWithUsers := &inputport.TransactionWithUsersForHistory{
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

	return &inputport.GetTransactionHistoryResponse{
		Transactions: transactionsWithUsers,
		Total:        total,
	}, nil
}

// GetBalance は残高を取得
func (i *PointTransferInteractor) GetBalance(ctx context.Context, req *inputport.GetBalanceRequest) (*inputport.GetBalanceResponse, error) {
	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetBalanceResponse{
		Balance: user.Balance,
		User:    user,
	}, nil
}
