package interactor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// DailyBonusInteractor はデイリーボーナスの統合インタラクター
// HTTP API 向けの参照メソッドと、AkerunWorker 向けのボーナス付与メソッドを両方提供する
type DailyBonusInteractor struct {
	dailyBonusRepo     repository.DailyBonusRepository
	userRepo           repository.UserRepository
	transactionRepo    repository.TransactionRepository
	txManager          repository.TransactionManager
	systemSettingsRepo repository.SystemSettingsRepository
	pointBatchRepo     repository.PointBatchRepository
	lotteryTierRepo    repository.LotteryTierRepository
	logger             entities.Logger
}

// NewDailyBonusInteractor は新しいDailyBonusInteractorを作成
func NewDailyBonusInteractor(
	dailyBonusRepo repository.DailyBonusRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	txManager repository.TransactionManager,
	systemSettingsRepo repository.SystemSettingsRepository,
	pointBatchRepo repository.PointBatchRepository,
	lotteryTierRepo repository.LotteryTierRepository,
	logger entities.Logger,
) *DailyBonusInteractor {
	return &DailyBonusInteractor{
		dailyBonusRepo:     dailyBonusRepo,
		userRepo:           userRepo,
		transactionRepo:    transactionRepo,
		txManager:          txManager,
		systemSettingsRepo: systemSettingsRepo,
		pointBatchRepo:     pointBatchRepo,
		lotteryTierRepo:    lotteryTierRepo,
		logger:             logger,
	}
}

// ========================================
// HTTP API 向けメソッド（DailyBonusInputPort）
// ========================================

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

	// 設定ポイントを取得（フォールバック用）
	bonusPoints := i.getBonusPoints(ctx)

	// 未抽選のボーナスがあればルーレットペンディング
	isLotteryPending := false
	if bonus != nil && !bonus.IsDrawn {
		isLotteryPending = true
	}

	return &inputport.GetTodayBonusResponse{
		DailyBonus:       bonus,
		BonusPoints:      bonusPoints,
		TotalDays:        totalDays,
		IsLotteryPending: isLotteryPending,
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

// GetBonusSettings はボーナス設定を取得（抽選ティア含む）
func (i *DailyBonusInteractor) GetBonusSettings(ctx context.Context) (*inputport.BonusSettingsResponse, error) {
	tiers, err := i.lotteryTierRepo.ReadAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get lottery tiers: %w", err)
	}

	return &inputport.BonusSettingsResponse{
		BonusPoints:  i.getBonusPoints(ctx),
		LotteryTiers: tiers,
	}, nil
}

// UpdateLotteryTiers は抽選ティアを一括更新（管理者用）
func (i *DailyBonusInteractor) UpdateLotteryTiers(ctx context.Context, req *inputport.UpdateLotteryTiersRequest) error {
	tiers := make([]*entities.LotteryTier, len(req.Tiers))
	for idx, t := range req.Tiers {
		tiers[idx] = entities.NewLotteryTier(t.Name, t.Points, t.Probability, t.DisplayOrder)
	}

	return i.txManager.Do(ctx, func(ctx context.Context) error {
		return i.lotteryTierRepo.ReplaceAll(ctx, tiers)
	})
}

// MarkBonusViewed はボーナスを閲覧済みにする
func (i *DailyBonusInteractor) MarkBonusViewed(ctx context.Context, req *inputport.MarkBonusViewedRequest) error {
	// ボーナスの所有者チェック
	bonusDate := entities.GetBonusDateJST(time.Now())
	bonus, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, req.UserID, bonusDate)
	if err != nil {
		return err
	}
	if bonus == nil {
		return fmt.Errorf("bonus not found")
	}
	if bonus.ID != req.BonusID {
		return fmt.Errorf("bonus does not belong to user")
	}

	return i.dailyBonusRepo.MarkAsViewed(ctx, req.BonusID)
}

// ========================================
// AkerunWorker 向けメソッド（AkerunBonusInputPort）
// ========================================

// DrawLotteryAndGrant はルーレットを実行しポイントを付与する（Phase 2: ユーザーがルーレットを回した時）
func (i *DailyBonusInteractor) DrawLotteryAndGrant(ctx context.Context, req *inputport.DrawLotteryRequest) (*inputport.DrawLotteryResponse, error) {
	// 今日のボーナス日付を計算
	bonusDate := entities.GetBonusDateJST(time.Now())

	// 未抽選のボーナスを取得
	bonus, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, req.UserID, bonusDate)
	if err != nil {
		return nil, fmt.Errorf("failed to read bonus: %w", err)
	}
	if bonus == nil {
		return nil, fmt.Errorf("no pending bonus found")
	}
	if bonus.IsDrawn {
		// 既に抽選済みの場合は結果を返す
		return &inputport.DrawLotteryResponse{
			BonusPoints:     bonus.BonusPoints,
			LotteryTierName: bonus.LotteryTierName,
			BonusID:         bonus.ID,
		}, nil
	}

	// アクティブな抽選ティアを取得
	lotteryTiers, err := i.lotteryTierRepo.ReadActive(ctx)
	if err != nil {
		i.logger.Error("DrawLotteryAndGrant: failed to get lottery tiers", entities.NewField("error", err))
		lotteryTiers = nil
	}

	// くじ引き実行
	fallbackPoints := i.getFallbackPoints(lotteryTiers, ctx)
	bonusPoints, lotteryTierID, lotteryTierName := i.drawLottery(lotteryTiers, fallbackPoints, req.UserID, bonus.AkerunUserName)

	// トランザクション内でポイント付与 + ボーナスレコード更新
	err = i.txManager.Do(ctx, func(ctx context.Context) error {
		// 抽選結果を更新
		if err := i.dailyBonusRepo.UpdateDrawnResult(ctx, bonus.ID, bonusPoints, lotteryTierID, lotteryTierName); err != nil {
			return fmt.Errorf("failed to update drawn result: %w", err)
		}

		// 0ptの場合はポイント付与スキップ
		if bonusPoints > 0 {
			// ポイント付与トランザクション
			desc := fmt.Sprintf("Akerun入退室ボーナス（%s）", lotteryTierName)
			tx, err := entities.NewAdminGrant(
				req.UserID,
				bonusPoints,
				desc,
				uuid.Nil, // システム処理
			)
			if err != nil {
				return fmt.Errorf("failed to create transaction: %w", err)
			}
			if err := i.transactionRepo.Create(ctx, tx); err != nil {
				return fmt.Errorf("failed to save transaction: %w", err)
			}

			// ユーザー残高更新
			updates := []repository.BalanceUpdate{
				{UserID: req.UserID, Amount: bonusPoints, IsDeduct: false},
			}
			if err := i.userRepo.UpdateBalancesWithLock(ctx, updates); err != nil {
				return fmt.Errorf("failed to update balance: %w", err)
			}

			// ポイントバッチ作成
			batch := entities.NewPointBatch(req.UserID, bonusPoints, entities.PointBatchSourceDailyBonus, &tx.ID, time.Now())
			if err := i.pointBatchRepo.Create(ctx, batch); err != nil {
				return fmt.Errorf("failed to create point batch: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	i.logger.Info("DrawLotteryAndGrant: lottery completed",
		entities.NewField("user_id", req.UserID),
		entities.NewField("points", bonusPoints),
		entities.NewField("tier", lotteryTierName))

	return &inputport.DrawLotteryResponse{
		BonusPoints:     bonusPoints,
		LotteryTierName: lotteryTierName,
		BonusID:         bonus.ID,
	}, nil
}

// ProcessAccesses はアクセス記録を処理して未抽選ボーナスを作成する（Phase 1: アクセス記録のみ）
func (i *DailyBonusInteractor) ProcessAccesses(ctx context.Context, accesses []entities.AccessRecord) error {
	// 全ユーザーを取得してマッチング用マップを構築
	nameToUser := i.buildUserNameMap(ctx)
	if nameToUser == nil {
		return fmt.Errorf("failed to build user name map")
	}

	for _, access := range accesses {
		if access.UserName == "" {
			continue
		}

		// Akerunユーザー名を正規化
		akerunName := entities.NormalizeName(access.UserName)

		// アプリユーザーとマッチング
		userID, matched := nameToUser[akerunName]
		if !matched {
			continue
		}

		// ボーナス日付を計算（JST AM6:00区切り）
		bonusDate := entities.GetBonusDateJST(access.AccessedAt)

		// 既にボーナス付与済みかチェック
		existing, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, userID, bonusDate)
		if err != nil {
			i.logger.Error("DailyBonusInteractor: failed to check existing bonus",
				entities.NewField("user_id", userID),
				entities.NewField("error", err))
			continue
		}

		if existing != nil {
			continue
		}

		// 未抽選のボーナスレコードを作成（ポイント未確定）
		accessedAt := access.AccessedAt
		accessIDStr := access.ID.String()
		bonus := entities.NewPendingDailyBonus(userID, bonusDate, accessIDStr, access.UserName, &accessedAt)
		if err := i.dailyBonusRepo.Create(ctx, bonus); err != nil {
			i.logger.Error("DailyBonusInteractor: failed to create pending bonus",
				entities.NewField("user_id", userID),
				entities.NewField("akerun_user", access.UserName),
				entities.NewField("error", err))
		} else {
			i.logger.Info("DailyBonusInteractor: pending bonus created",
				entities.NewField("user_id", userID),
				entities.NewField("akerun_user", access.UserName),
				entities.NewField("date", bonusDate.Format("2006-01-02")))
		}
	}

	return nil
}

// GetLastPolledAt は前回ポーリング時刻を取得する
func (i *DailyBonusInteractor) GetLastPolledAt(ctx context.Context) (time.Time, error) {
	return i.dailyBonusRepo.GetLastPolledAt(ctx)
}

// UpdateLastPolledAt はポーリング時刻を更新する
func (i *DailyBonusInteractor) UpdateLastPolledAt(ctx context.Context, t time.Time) error {
	return i.dailyBonusRepo.UpdateLastPolledAt(ctx, t)
}

// ========================================
// プライベートヘルパー
// ========================================

// getBonusPoints は現在のボーナスポイント設定を取得（フォールバック用）
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

// getFallbackPoints はフォールバック用固定ポイントを取得する（ティアがある場合は不要）
func (i *DailyBonusInteractor) getFallbackPoints(lotteryTiers []*entities.LotteryTier, ctx context.Context) int64 {
	if len(lotteryTiers) > 0 {
		return 0 // ティアがある場合はフォールバック不要
	}
	return i.getBonusPoints(ctx)
}

// buildUserNameMap は全ユーザーを取得し正規化名→UserIDのマップを構築する
func (i *DailyBonusInteractor) buildUserNameMap(ctx context.Context) map[string]uuid.UUID {
	users, err := i.userRepo.ReadList(ctx, 0, 10000)
	if err != nil {
		i.logger.Error("DailyBonusInteractor: failed to get users", entities.NewField("error", err))
		return nil
	}

	nameToUser := make(map[string]uuid.UUID)
	for _, user := range users {
		if user.LastName != "" && user.FirstName != "" {
			// "田中太郎" 形式
			fullName := entities.NormalizeName(user.LastName + user.FirstName)
			nameToUser[fullName] = user.ID

			// "田中 太郎" 形式（スペース区切り）もカバー
			fullNameWithSpace := entities.NormalizeName(user.LastName + " " + user.FirstName)
			nameToUser[fullNameWithSpace] = user.ID
		}
	}

	return nameToUser
}

// drawLottery はくじ引きを実行し、ボーナスポイント・ティアID・ティア名を返す
func (i *DailyBonusInteractor) drawLottery(lotteryTiers []*entities.LotteryTier, fallbackPoints int64, userID uuid.UUID, akerunUserName string) (int64, *uuid.UUID, string) {
	if len(lotteryTiers) > 0 {
		result := entities.DrawLottery(lotteryTiers)
		if result == nil {
			i.logger.Info("DailyBonusInteractor: lottery miss (no bonus)",
				entities.NewField("user_id", userID),
				entities.NewField("akerun_user", akerunUserName))
			return 0, nil, "ハズレ"
		}
		return result.Points, &result.ID, result.Name
	}

	return fallbackPoints, nil, "通常"
}

// grantBonus はボーナスをトランザクション内で付与する
func (i *DailyBonusInteractor) grantBonus(ctx context.Context, userID uuid.UUID, bonusDate time.Time, bonusPoints int64, access entities.AccessRecord, lotteryTierID *uuid.UUID, lotteryTierName string) error {
	return i.txManager.Do(ctx, func(txCtx context.Context) error {
		// ボーナスレコード作成
		accessIDStr := access.ID.String()
		accessedAt := access.AccessedAt
		bonus := entities.NewDailyBonus(userID, bonusDate, bonusPoints, accessIDStr, access.UserName, &accessedAt, lotteryTierID, lotteryTierName)
		if err := i.dailyBonusRepo.Create(txCtx, bonus); err != nil {
			return fmt.Errorf("failed to create daily bonus: %w", err)
		}

		// 0ptの場合はポイント付与スキップ
		if bonusPoints > 0 {
			// ポイント付与トランザクション
			desc := fmt.Sprintf("Akerun入退室ボーナス（%s）", lotteryTierName)
			tx, err := entities.NewAdminGrant(
				userID,
				bonusPoints,
				desc,
				uuid.Nil, // システム処理
			)
			if err != nil {
				return fmt.Errorf("failed to create transaction: %w", err)
			}
			if err := i.transactionRepo.Create(txCtx, tx); err != nil {
				return fmt.Errorf("failed to save transaction: %w", err)
			}

			// ユーザー残高更新
			updates := []repository.BalanceUpdate{
				{UserID: userID, Amount: bonusPoints, IsDeduct: false},
			}
			if err := i.userRepo.UpdateBalancesWithLock(txCtx, updates); err != nil {
				return fmt.Errorf("failed to update balance: %w", err)
			}

			// ポイントバッチ作成
			batch := entities.NewPointBatch(userID, bonusPoints, entities.PointBatchSourceDailyBonus, &tx.ID, time.Now())
			if err := i.pointBatchRepo.Create(txCtx, batch); err != nil {
				return fmt.Errorf("failed to create point batch: %w", err)
			}
		}

		return nil
	})
}
