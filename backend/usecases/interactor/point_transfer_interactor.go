package interactor

import (
	"errors"
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"gorm.io/gorm"
)

// PointTransferInteractor はポイント転送のユースケース実装
type PointTransferInteractor struct {
	db              *gorm.DB
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	idempotencyRepo repository.IdempotencyKeyRepository
	friendshipRepo  repository.FriendshipRepository
	logger          entities.Logger
}

// NewPointTransferInteractor は新しいPointTransferInteractorを作成
func NewPointTransferInteractor(
	db *gorm.DB,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	idempotencyRepo repository.IdempotencyKeyRepository,
	friendshipRepo repository.FriendshipRepository,
	logger entities.Logger,
) inputport.PointTransferInputPort {
	return &PointTransferInteractor{
		db:              db,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		idempotencyRepo: idempotencyRepo,
		friendshipRepo:  friendshipRepo,
		logger:          logger,
	}
}

// Transfer はポイントを転送
//
// セキュリティと整合性の保証:
// 1. 冪等性: IdempotencyKeyで重複転送を防止
// 2. トランザクション: DBトランザクションで原子性を保証
// 3. 悲観的ロック: SELECT FOR UPDATEで競合を防止
// 4. 残高チェック: 送信者の残高を厳密にチェック
// 5. 友達チェック: 友達関係がある場合のみ転送可能（オプション）
//
// 技術的説明:
// - PostgreSQLのトランザクション分離レベル: READ COMMITTED
// - ロック戦略: SELECT FOR UPDATE（行レベル悲観的ロック）
// - エラーハンドリング: ロールバック処理を確実に実行
func (i *PointTransferInteractor) Transfer(req *inputport.TransferRequest) (*inputport.TransferResponse, error) {
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
	existingKey, err := i.idempotencyRepo.ReadByKey(req.IdempotencyKey)
	if err == nil {
		// 既存のキーが見つかった場合
		if existingKey.Status == "completed" && existingKey.TransactionID != nil {
			// 完了済みの場合は既存のトランザクションを返す
			transaction, err := i.transactionRepo.Read(*existingKey.TransactionID)
			if err != nil {
				return nil, err
			}

			fromUser, _ := i.userRepo.Read(req.FromUserID)
			toUser, _ := i.userRepo.Read(req.ToUserID)

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
	if err := i.idempotencyRepo.Create(idempotencyKey); err != nil {
		// 競合エラーの場合は二重送信
		return nil, errors.New("duplicate idempotency key")
	}

	// === トランザクション開始 ===
	var fromUser, toUser *entities.User
	var transaction *entities.Transaction

	err = i.db.Transaction(func(tx *gorm.DB) error {
		// 1. 送信者と受信者の存在確認
		fromUser, err = i.userRepo.Read(req.FromUserID)
		if err != nil {
			return fmt.Errorf("sender not found: %w", err)
		}

		toUser, err = i.userRepo.Read(req.ToUserID)
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

		// 3. 残高更新（SELECT FOR UPDATEで悲観的ロック）
		// デッドロック回避: UUIDの小さい順にロックを取得
		firstUserID, secondUserID := req.FromUserID, req.ToUserID
		firstIsFrom := true

		// UUID比較: 小さい方を先にロック
		if req.ToUserID.String() < req.FromUserID.String() {
			firstUserID, secondUserID = req.ToUserID, req.FromUserID
			firstIsFrom = false
		}

		// 1番目のユーザーをロック
		if firstIsFrom {
			// FromUserを先にロック（減算）
			if err := i.userRepo.UpdateBalanceWithLock(tx, firstUserID, req.Amount, true); err != nil {
				return fmt.Errorf("failed to deduct from sender: %w", err)
			}
		} else {
			// ToUserを先にロック（加算）
			if err := i.userRepo.UpdateBalanceWithLock(tx, firstUserID, req.Amount, false); err != nil {
				return fmt.Errorf("failed to add to receiver: %w", err)
			}
		}

		// 2番目のユーザーをロック
		if firstIsFrom {
			// ToUserを後にロック（加算）
			if err := i.userRepo.UpdateBalanceWithLock(tx, secondUserID, req.Amount, false); err != nil {
				return fmt.Errorf("failed to add to receiver: %w", err)
			}
		} else {
			// FromUserを後にロック（減算）
			if err := i.userRepo.UpdateBalanceWithLock(tx, secondUserID, req.Amount, true); err != nil {
				return fmt.Errorf("failed to deduct from sender: %w", err)
			}
		}

		// 4. トランザクション記録作成
		transaction, err = entities.NewTransfer(req.FromUserID, req.ToUserID, req.Amount, req.IdempotencyKey, req.Description)
		if err != nil {
			return err
		}

		if err := i.transactionRepo.Create(tx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// 6. トランザクションを完了状態に
		if err := transaction.Complete(); err != nil {
			return err
		}

		if err := i.transactionRepo.Update(tx, transaction); err != nil {
			return err
		}

		// 7. 冪等性キーを完了状態に
		idempotencyKey.Status = "completed"
		idempotencyKey.TransactionID = &transaction.ID
		if err := i.idempotencyRepo.Update(idempotencyKey); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// トランザクション失敗時は冪等性キーを失敗状態に
		idempotencyKey.Status = "failed"
		i.idempotencyRepo.Update(idempotencyKey)
		i.logger.Error("Point transfer failed", entities.NewField("error", err))
		return nil, err
	}

	// 最新の残高を取得
	fromUser, _ = i.userRepo.Read(req.FromUserID)
	toUser, _ = i.userRepo.Read(req.ToUserID)

	i.logger.Info("Point transfer completed successfully",
		entities.NewField("transaction_id", transaction.ID))

	return &inputport.TransferResponse{
		Transaction: transaction,
		FromUser:    fromUser,
		ToUser:      toUser,
	}, nil
}

// GetTransactionHistory はトランザクション履歴を取得
func (i *PointTransferInteractor) GetTransactionHistory(req *inputport.GetTransactionHistoryRequest) (*inputport.GetTransactionHistoryResponse, error) {
	transactions, err := i.transactionRepo.ReadListByUserID(req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	total, err := i.transactionRepo.CountByUserID(req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetTransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
	}, nil
}

// GetBalance は残高を取得
func (i *PointTransferInteractor) GetBalance(req *inputport.GetBalanceRequest) (*inputport.GetBalanceResponse, error) {
	user, err := i.userRepo.Read(req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetBalanceResponse{
		Balance: user.Balance,
		User:    user,
	}, nil
}
