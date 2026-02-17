package interactor

import (
	"context"
	"strconv"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
)

// DailyBonusInteractor はデイリーボーナスのインタラクター
type DailyBonusInteractor struct {
	dailyBonusRepo     repository.DailyBonusRepository
	systemSettingsRepo repository.SystemSettingsRepository
	logger             entities.Logger
}

// NewDailyBonusInteractor は新しいDailyBonusInteractorを作成
func NewDailyBonusInteractor(
	dailyBonusRepo repository.DailyBonusRepository,
	systemSettingsRepo repository.SystemSettingsRepository,
	logger entities.Logger,
) *DailyBonusInteractor {
	return &DailyBonusInteractor{
		dailyBonusRepo:     dailyBonusRepo,
		systemSettingsRepo: systemSettingsRepo,
		logger:             logger,
	}
}

// GetTodayBonus は本日のボーナス状況を取得
func (i *DailyBonusInteractor) GetTodayBonus(ctx context.Context, req *inputport.GetTodayBonusRequest) (*inputport.GetTodayBonusResponse, error) {
	// 今日のボーナス日付を計算（JST AM6:00区切り）
	bonusDate := entities.GetBonusDateJST(time.Now())

	// 今日のボーナスを取得
	bonus, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, req.UserID, bonusDate)
	if err != nil {
		return nil, err
	}

	// 獲得日数を取得
	totalDays, err := i.dailyBonusRepo.CountByUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// 設定ポイントを取得
	bonusPoints := i.getBonusPoints(ctx)

	return &inputport.GetTodayBonusResponse{
		DailyBonus:  bonus,
		BonusPoints: bonusPoints,
		TotalDays:   totalDays,
	}, nil
}

// GetRecentBonuses は最近のボーナス履歴を取得
func (i *DailyBonusInteractor) GetRecentBonuses(ctx context.Context, req *inputport.GetRecentBonusesRequest) (*inputport.GetRecentBonusesResponse, error) {
	bonuses, err := i.dailyBonusRepo.ReadRecentByUser(ctx, req.UserID, req.Limit)
	if err != nil {
		return nil, err
	}

	totalDays, err := i.dailyBonusRepo.CountByUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetRecentBonusesResponse{
		Bonuses:   bonuses,
		TotalDays: totalDays,
	}, nil
}

// GetBonusSettings はボーナス設定を取得
func (i *DailyBonusInteractor) GetBonusSettings(ctx context.Context) (*inputport.BonusSettingsResponse, error) {
	return &inputport.BonusSettingsResponse{
		BonusPoints: i.getBonusPoints(ctx),
	}, nil
}

// UpdateBonusSettings はボーナス設定を更新（管理者用）
func (i *DailyBonusInteractor) UpdateBonusSettings(ctx context.Context, req *inputport.UpdateBonusSettingsRequest) error {
	return i.systemSettingsRepo.SetSetting(ctx,
		"akerun_bonus_points",
		strconv.FormatInt(req.BonusPoints, 10),
		"Akerun入退室ボーナスのポイント数",
	)
}

// getBonusPoints は現在のボーナスポイント設定を取得
func (i *DailyBonusInteractor) getBonusPoints(ctx context.Context) int64 {
	pointsStr, err := i.systemSettingsRepo.GetSetting(ctx, "akerun_bonus_points")
	if err != nil || pointsStr == "" {
		return entities.DefaultAkerunBonusPoints
	}
	pts, err := strconv.ParseInt(pointsStr, 10, 64)
	if err != nil || pts <= 0 {
		return entities.DefaultAkerunBonusPoints
	}
	return pts
}
