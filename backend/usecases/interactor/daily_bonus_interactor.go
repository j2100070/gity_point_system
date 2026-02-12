package interactor

import (
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DailyBonusInteractor はデイリーボーナスのインタラクター
type DailyBonusInteractor struct {
	dailyBonusRepo repository.DailyBonusRepository
	userRepo       repository.UserRepository
	transactionRepo repository.TransactionRepository
	db             inframysql.DB
	logger         entities.Logger
}

// NewDailyBonusInteractor は新しいDailyBonusInteractorを作成
func NewDailyBonusInteractor(
	dailyBonusRepo repository.DailyBonusRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	db inframysql.DB,
	logger entities.Logger,
) *DailyBonusInteractor {
	return &DailyBonusInteractor{
		dailyBonusRepo:  dailyBonusRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		db:              db,
		logger:          logger,
	}
}

// CheckLoginBonus はログインボーナスをチェックして付与
func (i *DailyBonusInteractor) CheckLoginBonus(req *inputport.CheckLoginBonusRequest) (*inputport.CheckLoginBonusResponse, error) {
	i.logger.Info("Checking login bonus", entities.NewField("user_id", req.UserID))

	// トランザクション開始
	tx := i.db.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 本日のボーナスレコードを取得または作成
	dateOnly := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
	dailyBonus, err := i.dailyBonusRepo.ReadByUserAndDate(req.UserID, dateOnly)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if dailyBonus == nil {
		// 新規作成
		dailyBonus = entities.NewDailyBonus(req.UserID, dateOnly)
	}

	// ログインボーナスを達成
	bonusPoints := dailyBonus.CompleteLogin()

	// ボーナスレコードを保存
	if dailyBonus.CreatedAt.IsZero() {
		err = i.dailyBonusRepo.Create(dailyBonus)
	} else {
		err = i.dailyBonusRepo.Update(dailyBonus)
	}
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ポイントを付与
	var user *entities.User
	if bonusPoints > 0 {
		user, err = i.grantBonusPoints(tx, req.UserID, bonusPoints, "ログインボーナス")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		user, err = i.userRepo.Read(req.UserID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// コミット
	if err := tx.Commit().Error; err != nil {
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
func (i *DailyBonusInteractor) CheckTransferBonus(req *inputport.CheckTransferBonusRequest) (*inputport.CheckTransferBonusResponse, error) {
	i.logger.Info("Checking transfer bonus", entities.NewField("user_id", req.UserID))

	// トランザクション開始
	tx := i.db.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 本日のボーナスレコードを取得または作成
	dateOnly := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
	dailyBonus, err := i.dailyBonusRepo.ReadByUserAndDate(req.UserID, dateOnly)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if dailyBonus == nil {
		// 新規作成
		dailyBonus = entities.NewDailyBonus(req.UserID, dateOnly)
	}

	// 送金ボーナスを達成
	bonusPoints := dailyBonus.CompleteTransfer(req.TransactionID)

	// ボーナスレコードを保存
	if dailyBonus.CreatedAt.IsZero() {
		err = i.dailyBonusRepo.Create(dailyBonus)
	} else {
		err = i.dailyBonusRepo.Update(dailyBonus)
	}
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ポイントを付与
	var user *entities.User
	if bonusPoints > 0 {
		user, err = i.grantBonusPoints(tx, req.UserID, bonusPoints, "送金ボーナス")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		user, err = i.userRepo.Read(req.UserID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// コミット
	if err := tx.Commit().Error; err != nil {
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
func (i *DailyBonusInteractor) CheckExchangeBonus(req *inputport.CheckExchangeBonusRequest) (*inputport.CheckExchangeBonusResponse, error) {
	i.logger.Info("Checking exchange bonus", entities.NewField("user_id", req.UserID))

	// トランザクション開始
	tx := i.db.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 本日のボーナスレコードを取得または作成
	dateOnly := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
	dailyBonus, err := i.dailyBonusRepo.ReadByUserAndDate(req.UserID, dateOnly)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if dailyBonus == nil {
		// 新規作成
		dailyBonus = entities.NewDailyBonus(req.UserID, dateOnly)
	}

	// 交換ボーナスを達成
	bonusPoints := dailyBonus.CompleteExchange(req.ExchangeID)

	// ボーナスレコードを保存
	if dailyBonus.CreatedAt.IsZero() {
		err = i.dailyBonusRepo.Create(dailyBonus)
	} else {
		err = i.dailyBonusRepo.Update(dailyBonus)
	}
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ポイントを付与
	var user *entities.User
	if bonusPoints > 0 {
		user, err = i.grantBonusPoints(tx, req.UserID, bonusPoints, "商品交換ボーナス")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		user, err = i.userRepo.Read(req.UserID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// コミット
	if err := tx.Commit().Error; err != nil {
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
func (i *DailyBonusInteractor) GetTodayBonus(req *inputport.GetTodayBonusRequest) (*inputport.GetTodayBonusResponse, error) {
	now := time.Now()
	dateOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	dailyBonus, err := i.dailyBonusRepo.ReadByUserAndDate(req.UserID, dateOnly)
	if err != nil {
		return nil, err
	}

	// 存在しない場合は新規作成して返す
	if dailyBonus == nil {
		dailyBonus = entities.NewDailyBonus(req.UserID, dateOnly)
	}

	// 全達成回数を取得
	allCompletedCount, err := i.dailyBonusRepo.CountAllCompletedByUser(req.UserID)
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
func (i *DailyBonusInteractor) GetRecentBonuses(req *inputport.GetRecentBonusesRequest) (*inputport.GetRecentBonusesResponse, error) {
	bonuses, err := i.dailyBonusRepo.ReadRecentByUser(req.UserID, req.Limit)
	if err != nil {
		return nil, err
	}

	allCompletedCount, err := i.dailyBonusRepo.CountAllCompletedByUser(req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetRecentBonusesResponse{
		Bonuses:           bonuses,
		AllCompletedCount: allCompletedCount,
	}, nil
}

// grantBonusPoints はボーナスポイントを付与（トランザクション内で実行）
func (i *DailyBonusInteractor) grantBonusPoints(tx *gorm.DB, userID uuid.UUID, points int64, description string) (*entities.User, error) {
	// ユーザーを取得してロック
	user, err := i.userRepo.Read(userID)
	if err != nil {
		return nil, err
	}

	// 残高を増やす
	user.Balance += points
	user.UpdatedAt = time.Now()

	// ユーザーを更新
	if err := tx.Save(user).Error; err != nil {
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

	if err := i.transactionRepo.Create(tx, transaction); err != nil {
		return nil, errors.New("failed to create bonus transaction")
	}

	return user, nil
}

// timePtr はtime.Timeのポインタを返すヘルパー関数
func timePtr(t time.Time) *time.Time {
	return &t
}
