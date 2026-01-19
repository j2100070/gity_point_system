package usecase

import (
	"errors"
	"fmt"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PointTransferUseCase はポイント転送に関するユースケース
type PointTransferUseCase struct {
	db               *gorm.DB
	userRepo         domain.UserRepository
	transactionRepo  domain.TransactionRepository
	idempotencyRepo  domain.IdempotencyKeyRepository
	friendshipRepo   domain.FriendshipRepository
}

// NewPointTransferUseCase は新しいPointTransferUseCaseを作成
func NewPointTransferUseCase(
	db *gorm.DB,
	userRepo domain.UserRepository,
	transactionRepo domain.TransactionRepository,
	idempotencyRepo domain.IdempotencyKeyRepository,
	friendshipRepo domain.FriendshipRepository,
) *PointTransferUseCase {
	return &PointTransferUseCase{
		db:              db,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		idempotencyRepo: idempotencyRepo,
		friendshipRepo:  friendshipRepo,
	}
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
	Transaction *domain.Transaction
	FromUser    *domain.User
	ToUser      *domain.User
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
func (u *PointTransferUseCase) Transfer(req *TransferRequest) (*TransferResponse, error) {
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
	existingKey, err := u.idempotencyRepo.FindByKey(req.IdempotencyKey)
	if err == nil {
		// 既存のキーが見つかった場合
		if existingKey.Status == "completed" && existingKey.TransactionID != nil {
			// 完了済みの場合は既存のトランザクションを返す
			transaction, err := u.transactionRepo.FindByID(*existingKey.TransactionID)
			if err != nil {
				return nil, err
			}

			fromUser, _ := u.userRepo.FindByID(req.FromUserID)
			toUser, _ := u.userRepo.FindByID(req.ToUserID)

			return &TransferResponse{
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
	idempotencyKey := domain.NewIdempotencyKey(req.IdempotencyKey, req.FromUserID)
	if err := u.idempotencyRepo.Create(idempotencyKey); err != nil {
		// 競合エラーの場合は二重送信
		return nil, errors.New("duplicate idempotency key")
	}

	// === トランザクション開始 ===
	var fromUser, toUser *domain.User
	var transaction *domain.Transaction

	err = u.db.Transaction(func(tx *gorm.DB) error {
		// 1. 送信者と受信者の存在確認
		fromUser, err = u.userRepo.FindByID(req.FromUserID)
		if err != nil {
			return fmt.Errorf("sender not found: %w", err)
		}

		toUser, err = u.userRepo.FindByID(req.ToUserID)
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

		// 3. 友達関係チェック（オプション: 要件に応じて有効化）
		// areFriends, err := u.friendshipRepo.AreFriends(req.FromUserID, req.ToUserID)
		// if err != nil {
		// 	return err
		// }
		// if !areFriends {
		// 	return errors.New("can only transfer to friends")
		// }

		// 4. 残高更新（SELECT FOR UPDATEで悲観的ロック）
		// 送信者から減算
		if err := u.userRepo.UpdateBalanceWithLock(tx, req.FromUserID, req.Amount, true); err != nil {
			return fmt.Errorf("failed to deduct from sender: %w", err)
		}

		// 受信者に加算
		if err := u.userRepo.UpdateBalanceWithLock(tx, req.ToUserID, req.Amount, false); err != nil {
			return fmt.Errorf("failed to add to receiver: %w", err)
		}

		// 5. トランザクション記録作成
		transaction, err = domain.NewTransfer(req.FromUserID, req.ToUserID, req.Amount, req.IdempotencyKey, req.Description)
		if err != nil {
			return err
		}

		if err := u.transactionRepo.Create(tx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// 6. トランザクションを完了状態に
		if err := transaction.Complete(); err != nil {
			return err
		}

		if err := u.transactionRepo.Update(tx, transaction); err != nil {
			return err
		}

		// 7. 冪等性キーを完了状態に
		idempotencyKey.Status = "completed"
		idempotencyKey.TransactionID = &transaction.ID
		if err := u.idempotencyRepo.Update(idempotencyKey); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// トランザクション失敗時は冪等性キーを失敗状態に
		idempotencyKey.Status = "failed"
		u.idempotencyRepo.Update(idempotencyKey)
		return nil, err
	}

	// 最新の残高を取得
	fromUser, _ = u.userRepo.FindByID(req.FromUserID)
	toUser, _ = u.userRepo.FindByID(req.ToUserID)

	return &TransferResponse{
		Transaction: transaction,
		FromUser:    fromUser,
		ToUser:      toUser,
	}, nil
}

// GetTransactionHistoryRequest はトランザクション履歴取得リクエスト
type GetTransactionHistoryRequest struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}

// GetTransactionHistoryResponse はトランザクション履歴取得レスポンス
type GetTransactionHistoryResponse struct {
	Transactions []*domain.Transaction
	Total        int64
}

// GetTransactionHistory はトランザクション履歴を取得
func (u *PointTransferUseCase) GetTransactionHistory(req *GetTransactionHistoryRequest) (*GetTransactionHistoryResponse, error) {
	transactions, err := u.transactionRepo.ListByUserID(req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	total, err := u.transactionRepo.CountByUserID(req.UserID)
	if err != nil {
		return nil, err
	}

	return &GetTransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
	}, nil
}

// GetBalanceRequest は残高取得リクエスト
type GetBalanceRequest struct {
	UserID uuid.UUID
}

// GetBalanceResponse は残高取得レスポンス
type GetBalanceResponse struct {
	Balance int64
	User    *domain.User
}

// GetBalance は残高を取得
func (u *PointTransferUseCase) GetBalance(req *GetBalanceRequest) (*GetBalanceResponse, error) {
	user, err := u.userRepo.FindByID(req.UserID)
	if err != nil {
		return nil, err
	}

	return &GetBalanceResponse{
		Balance: user.Balance,
		User:    user,
	}, nil
}
