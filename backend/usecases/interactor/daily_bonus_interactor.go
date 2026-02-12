package interactor

import (
	"context"
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// DailyBonusInteractor はデイリーボーナスのインタラクター
type DailyBonusInteractor struct {
	dailyBonusRepo  repository.DailyBonusRepository
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	txManager       repository.TransactionManager // TransactionManagerを使用
	logger          entities.Logger
}

// NewDailyBonusInteractor は新しいDailyBonusInteractorを作成
func NewDailyBonusInteractor(
	dailyBonusRepo repository.DailyBonusRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	txManager repository.TransactionManager,
	logger entities.Logger,
) *DailyBonusInteractor {
	return &DailyBonusInteractor{
		dailyBonusRepo:  dailyBonusRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		txManager:       txManager,
		logger:          logger,
	}
}

// CheckLoginBonus はログインボーナスをチェックして付与
func (i *DailyBonusInteractor) CheckLoginBonus(ctx context.Context, req *inputport.CheckLoginBonusRequest) (*inputport.CheckLoginBonusResponse, error) {
	i.logger.Info("Checking login bonus", entities.NewField("user_id", req.UserID))

	var dailyBonus *entities.DailyBonus
	var bonusPoints int64
	var user *entities.User

	// TransactionManagerを使用してトランザクション内で処理を実行
	err := i.txManager.Do(ctx, func(ctx context.Context) error {
		// 本日のボーナスレコードを取得または作成
		dateOnly := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
		db, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, req.UserID, dateOnly)
		if err != nil {
			return err
		}

		if db == nil {
			// 新規作成
			db = entities.NewDailyBonus(req.UserID, dateOnly)
		}
		dailyBonus = db

		// ログインボーナスを達成
		bonusPoints = dailyBonus.CompleteLogin()

		// ボーナスレコードを保存
		if dailyBonus.CreatedAt.IsZero() {
			err = i.dailyBonusRepo.Create(ctx, dailyBonus)
		} else {
			err = i.dailyBonusRepo.Update(ctx, dailyBonus)
		}
		if err != nil {
			return err
		}

		// ポイントを付与
		if bonusPoints > 0 {
			user, err = i.grantBonusPoints(ctx, req.UserID, bonusPoints, "ログインボーナス")
			if err != nil {
				return err
			}
		} else {
			user, err = i.userRepo.Read(ctx, req.UserID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	i.logger.Info("Login bonus checked",
		entities.NewField("user_id", req.UserID),
		entities.NewField("bonus_awarded", bonusPoints))

	return &inputport.CheckLoginBonusResponse{
		DailyBonus:   dailyBonus,
		BonusAwarded: bonusPoints,
		User:         user,
	}, nil
}

// CheckTransferBonus は送金ボーナスをチェックして付与
func (i *DailyBonusInteractor) CheckTransferBonus(ctx context.Context, req *inputport.CheckTransferBonusRequest) (*inputport.CheckTransferBonusResponse, error) {
	i.logger.Info("Checking transfer bonus", entities.NewField("user_id", req.UserID))

	var dailyBonus *entities.DailyBonus
	var bonusPoints int64
	var user *entities.User

	err := i.txManager.Do(ctx, func(ctx context.Context) error {
		// 本日のボーナスレコードを取得または作成
		dateOnly := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
		db, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, req.UserID, dateOnly)
		if err != nil {
			return err
		}

		if db == nil {
			// 新規作成
			db = entities.NewDailyBonus(req.UserID, dateOnly)
		}
		dailyBonus = db

		// 送金ボーナスを達成
		bonusPoints = dailyBonus.CompleteTransfer(req.TransactionID)

		// ボーナスレコードを保存
		if dailyBonus.CreatedAt.IsZero() {
			err = i.dailyBonusRepo.Create(ctx, dailyBonus)
		} else {
			err = i.dailyBonusRepo.Update(ctx, dailyBonus)
		}
		if err != nil {
			return err
		}

		// ポイントを付与
		if bonusPoints > 0 {
			user, err = i.grantBonusPoints(ctx, req.UserID, bonusPoints, "送金ボーナス")
			if err != nil {
				return err
			}
		} else {
			user, err = i.userRepo.Read(ctx, req.UserID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	i.logger.Info("Transfer bonus checked",
		entities.NewField("user_id", req.UserID),
		entities.NewField("bonus_awarded", bonusPoints))

	return &inputport.CheckTransferBonusResponse{
		DailyBonus:   dailyBonus,
		BonusAwarded: bonusPoints,
		User:         user,
	}, nil
}

// CheckExchangeBonus は交換ボーナスをチェックして付与
func (i *DailyBonusInteractor) CheckExchangeBonus(ctx context.Context, req *inputport.CheckExchangeBonusRequest) (*inputport.CheckExchangeBonusResponse, error) {
	i.logger.Info("Checking exchange bonus", entities.NewField("user_id", req.UserID))

	var dailyBonus *entities.DailyBonus
	var bonusPoints int64
	var user *entities.User

	err := i.txManager.Do(ctx, func(ctx context.Context) error {
		// 本日のボーナスレコードを取得または作成
		dateOnly := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
		db, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, req.UserID, dateOnly)
		if err != nil {
			return err
		}

		if db == nil {
			// 新規作成
			db = entities.NewDailyBonus(req.UserID, dateOnly)
		}
		dailyBonus = db

		// 交換ボーナスを達成
		bonusPoints = dailyBonus.CompleteExchange(req.ExchangeID)

		// ボーナスレコードを保存
		if dailyBonus.CreatedAt.IsZero() {
			err = i.dailyBonusRepo.Create(ctx, dailyBonus)
		} else {
			err = i.dailyBonusRepo.Update(ctx, dailyBonus)
		}
		if err != nil {
			return err
		}

		// ポイントを付与
		if bonusPoints > 0 {
			user, err = i.grantBonusPoints(ctx, req.UserID, bonusPoints, "商品交換ボーナス")
			if err != nil {
				return err
			}
		} else {
			user, err = i.userRepo.Read(ctx, req.UserID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	i.logger.Info("Exchange bonus checked",
		entities.NewField("user_id", req.UserID),
		entities.NewField("bonus_awarded", bonusPoints))

	return &inputport.CheckExchangeBonusResponse{
		DailyBonus:   dailyBonus,
		BonusAwarded: bonusPoints,
		User:         user,
	}, nil
}

// GetTodayBonus は本日のボーナス状況を取得
func (i *DailyBonusInteractor) GetTodayBonus(ctx context.Context, req *inputport.GetTodayBonusRequest) (*inputport.GetTodayBonusResponse, error) {
	now := time.Now()
	dateOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	dailyBonus, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, req.UserID, dateOnly)
	if err != nil {
		return nil, err
	}

	// 存在しない場合は新規作成して返す
	if dailyBonus == nil {
		dailyBonus = entities.NewDailyBonus(req.UserID, dateOnly)
	}

	// 全達成回数を取得
	allCompletedCount, err := i.dailyBonusRepo.CountAllCompletedByUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetTodayBonusResponse{
		DailyBonus:            dailyBonus,
		AllCompletedCount:     allCompletedCount,
		CanClaimLoginBonus:    !dailyBonus.LoginCompleted,
		CanClaimTransferBonus: !dailyBonus.TransferCompleted,
		CanClaimExchangeBonus: !dailyBonus.ExchangeCompleted,
	}, nil
}

// GetRecentBonuses は最近のボーナス履歴を取得
func (i *DailyBonusInteractor) GetRecentBonuses(ctx context.Context, req *inputport.GetRecentBonusesRequest) (*inputport.GetRecentBonusesResponse, error) {
	bonuses, err := i.dailyBonusRepo.ReadRecentByUser(ctx, req.UserID, req.Limit)
	if err != nil {
		return nil, err
	}

	allCompletedCount, err := i.dailyBonusRepo.CountAllCompletedByUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetRecentBonusesResponse{
		Bonuses:           bonuses,
		AllCompletedCount: allCompletedCount,
	}, nil
}

// grantBonusPoints はボーナスポイントを付与（トランザクション内で実行）
func (i *DailyBonusInteractor) grantBonusPoints(ctx context.Context, userID uuid.UUID, points int64, description string) (*entities.User, error) {
	// 残高を更新（SELECT FOR UPDATE使用）
	if err := i.userRepo.UpdateBalanceWithLock(ctx, userID, points, false); err != nil {
		return nil, err
	}

	// 更新後のユーザー情報を取得
	user, err := i.userRepo.Read(ctx, userID)
	if err != nil {
		return nil, err
	}

	// トランザクションレコードを作成
	transaction := &entities.Transaction{
		ID:              uuid.New(),
		FromUserID:      nil, // システムから付与
		ToUserID:        &userID,
		Amount:          points,
		TransactionType: "daily_bonus",
		Status:          "completed",
		Description:     description,
		CreatedAt:       time.Now(),
		CompletedAt:     timePtr(time.Now()),
	}

	if err := i.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, errors.New("failed to create bonus transaction")
	}

	return user, nil
}

// timePtr はtime.Timeのポインタを返すヘルパー関数
func timePtr(t time.Time) *time.Time {
	return &t
}
