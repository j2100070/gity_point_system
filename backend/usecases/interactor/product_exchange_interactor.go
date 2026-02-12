package interactor

import (
	"errors"
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductExchangeInteractor は商品交換のユースケース実装
type ProductExchangeInteractor struct {
	db                  *gorm.DB
	productRepo         repository.ProductRepository
	exchangeRepo        repository.ProductExchangeRepository
	userRepo            repository.UserRepository
	transactionRepo     repository.TransactionRepository
	dailyBonusPort      inputport.DailyBonusInputPort
	logger              entities.Logger
}

// NewProductExchangeInteractor は新しいProductExchangeInteractorを作成
func NewProductExchangeInteractor(
	db *gorm.DB,
	productRepo repository.ProductRepository,
	exchangeRepo repository.ProductExchangeRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	logger entities.Logger,
) *ProductExchangeInteractor {
	return &ProductExchangeInteractor{
		db:              db,
		productRepo:     productRepo,
		exchangeRepo:    exchangeRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		dailyBonusPort:  nil, // 後からSetDailyBonusPortで設定
		logger:          logger,
	}
}

// SetDailyBonusPort はデイリーボーナスポートを設定（循環依存回避のため）
func (i *ProductExchangeInteractor) SetDailyBonusPort(dailyBonusPort inputport.DailyBonusInputPort) {
	i.dailyBonusPort = dailyBonusPort
}

// ExchangeProduct はポイントで商品を交換
//
// セキュリティと整合性の保証:
// 1. トランザクション: 在庫減算、ポイント減算、交換記録を原子的に実行
// 2. 悲観的ロック: 在庫とユーザー残高をロック
// 3. 残高チェック: 十分なポイントがあるか確認
// 4. 在庫チェック: 十分な在庫があるか確認
func (i *ProductExchangeInteractor) ExchangeProduct(req *inputport.ExchangeProductRequest) (*inputport.ExchangeProductResponse, error) {
	i.logger.Info("Starting product exchange",
		entities.NewField("user_id", req.UserID),
		entities.NewField("product_id", req.ProductID),
		entities.NewField("quantity", req.Quantity))

	// バリデーション
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	var user *entities.User
	var product *entities.Product
	var exchange *entities.ProductExchange
	var transaction *entities.Transaction

	err := i.db.Transaction(func(tx *gorm.DB) error {
		// 1. 商品情報を取得
		var err error
		product, err = i.productRepo.Read(req.ProductID)
		if err != nil {
			return fmt.Errorf("product not found: %w", err)
		}

		// 2. 商品の交換可否をチェック
		if err := product.CanExchange(req.Quantity); err != nil {
			return fmt.Errorf("cannot exchange product: %w", err)
		}

		// 3. 必要なポイント数を計算
		totalPoints := product.Price * int64(req.Quantity)

		// 4. ユーザー情報を取得（残高確認のためロック）
		user, err = i.userRepo.Read(req.UserID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		if !user.IsActive {
			return errors.New("user account is not active")
		}

		// 5. 残高チェック
		if user.Balance < totalPoints {
			return fmt.Errorf("insufficient balance: required %d, have %d", totalPoints, user.Balance)
		}

		// 6. 在庫を減らす（商品テーブルを更新）
		if err := product.DeductStock(req.Quantity); err != nil {
			return fmt.Errorf("failed to deduct stock: %w", err)
		}
		if err := i.productRepo.Update(product); err != nil {
			return fmt.Errorf("failed to update product stock: %w", err)
		}

		// 7. ユーザーの残高を減らす
		updates := []repository.BalanceUpdate{
			{UserID: req.UserID, Amount: totalPoints, IsDeduct: true},
		}
		if err := i.userRepo.UpdateBalancesWithLock(tx, updates); err != nil {
			return fmt.Errorf("failed to deduct balance: %w", err)
		}

		// 8. トランザクション記録を作成（ポイント減算記録）
		// NewAdminDeductは既にCompletedステータスで作成される
		transaction, err = entities.NewAdminDeduct(
			req.UserID,
			totalPoints,
			fmt.Sprintf("商品交換: %s x%d", product.Name, req.Quantity),
			uuid.Nil, // システム処理
		)
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		if err := i.transactionRepo.Create(tx, transaction); err != nil {
			return fmt.Errorf("failed to save transaction: %w", err)
		}

		// 9. 商品交換記録を作成
		exchange, err = entities.NewProductExchange(
			req.UserID,
			req.ProductID,
			req.Quantity,
			totalPoints,
			req.Notes,
		)
		if err != nil {
			return fmt.Errorf("failed to create exchange: %w", err)
		}

		if err := exchange.Complete(transaction.ID); err != nil {
			return fmt.Errorf("failed to complete exchange: %w", err)
		}

		if err := i.exchangeRepo.Create(tx, exchange); err != nil {
			return fmt.Errorf("failed to save exchange: %w", err)
		}

		return nil
	})

	if err != nil {
		i.logger.Error("Product exchange failed", entities.NewField("error", err))
		return nil, err
	}

	// 最新の情報を取得
	user, _ = i.userRepo.Read(req.UserID)
	product, _ = i.productRepo.Read(req.ProductID)

	i.logger.Info("Product exchange completed successfully",
		entities.NewField("exchange_id", exchange.ID),
		entities.NewField("points_used", exchange.PointsUsed))

	// デイリーボーナスチェック（商品交換ボーナス）
	if i.dailyBonusPort != nil {
		_, err := i.dailyBonusPort.CheckExchangeBonus(&inputport.CheckExchangeBonusRequest{
			UserID:     req.UserID,
			ExchangeID: exchange.ID,
			Date:       exchange.CreatedAt,
		})
		if err != nil {
			// ボーナス付与失敗してもメイントランザクションは成功扱い
			i.logger.Error("Failed to check exchange bonus", entities.NewField("error", err))
		}
	}

	return &inputport.ExchangeProductResponse{
		Exchange:    exchange,
		Product:     product,
		User:        user,
		Transaction: transaction,
	}, nil
}

// GetExchangeHistory は交換履歴を取得
func (i *ProductExchangeInteractor) GetExchangeHistory(req *inputport.GetExchangeHistoryRequest) (*inputport.GetExchangeHistoryResponse, error) {
	exchanges, err := i.exchangeRepo.ReadListByUserID(req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	total, err := i.exchangeRepo.CountByUserID(req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetExchangeHistoryResponse{
		Exchanges: exchanges,
		Total:     total,
	}, nil
}

// CancelExchange は交換をキャンセル
func (i *ProductExchangeInteractor) CancelExchange(req *inputport.CancelExchangeRequest) error {
	return i.db.Transaction(func(tx *gorm.DB) error {
		// 交換情報を取得
		exchange, err := i.exchangeRepo.Read(req.ExchangeID)
		if err != nil {
			return fmt.Errorf("exchange not found: %w", err)
		}

		// 権限チェック
		if exchange.UserID != req.UserID {
			return errors.New("unauthorized: not your exchange")
		}

		// キャンセル可能かチェック
		if err := exchange.Cancel(); err != nil {
			return err
		}

		// 在庫を戻す
		product, err := i.productRepo.Read(exchange.ProductID)
		if err != nil {
			return fmt.Errorf("product not found: %w", err)
		}

		if err := product.RestoreStock(exchange.Quantity); err != nil {
			return fmt.Errorf("failed to restore stock: %w", err)
		}

		if err := i.productRepo.Update(product); err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		// ポイントを戻す
		updates := []repository.BalanceUpdate{
			{UserID: req.UserID, Amount: exchange.PointsUsed, IsDeduct: false},
		}
		if err := i.userRepo.UpdateBalancesWithLock(tx, updates); err != nil {
			return fmt.Errorf("failed to restore balance: %w", err)
		}

		// 交換記録を更新
		if err := i.exchangeRepo.Update(tx, exchange); err != nil {
			return fmt.Errorf("failed to update exchange: %w", err)
		}

		return nil
	})
}

// MarkExchangeDelivered は配達完了にする（管理者用）
func (i *ProductExchangeInteractor) MarkExchangeDelivered(req *inputport.MarkExchangeDeliveredRequest) error {
	return i.db.Transaction(func(tx *gorm.DB) error {
		exchange, err := i.exchangeRepo.Read(req.ExchangeID)
		if err != nil {
			return fmt.Errorf("exchange not found: %w", err)
		}

		if err := exchange.MarkAsDelivered(); err != nil {
			return err
		}

		if err := i.exchangeRepo.Update(tx, exchange); err != nil {
			return fmt.Errorf("failed to update exchange: %w", err)
		}

		return nil
	})
}

// GetAllExchanges はすべての交換履歴を取得（管理者用）
func (i *ProductExchangeInteractor) GetAllExchanges(offset, limit int) (*inputport.GetExchangeHistoryResponse, error) {
	exchanges, err := i.exchangeRepo.ReadListAll(offset, limit)
	if err != nil {
		return nil, err
	}

	total, err := i.exchangeRepo.CountAll()
	if err != nil {
		return nil, err
	}

	return &inputport.GetExchangeHistoryResponse{
		Exchanges: exchanges,
		Total:     total,
	}, nil
}
