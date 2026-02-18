package infraakerun

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/repository"
	"github.com/gity/point-system/usecases/service"
	"github.com/google/uuid"
)

// AkerunWorker はAkerun入退室ポーリングワーカー
type AkerunWorker struct {
	client             *AkerunClient
	dailyBonusRepo     repository.DailyBonusRepository
	userRepo           repository.UserRepository
	transactionRepo    repository.TransactionRepository
	txManager          repository.TransactionManager
	systemSettingsRepo repository.SystemSettingsRepository
	timeProvider       service.TimeProvider
	logger             entities.Logger
	interval           time.Duration
	recoverySleep      time.Duration
	stopCh             chan struct{}
}

// NewAkerunWorker は新しいAkerunWorkerを作成
func NewAkerunWorker(
	client *AkerunClient,
	dailyBonusRepo repository.DailyBonusRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	txManager repository.TransactionManager,
	systemSettingsRepo repository.SystemSettingsRepository,
	timeProvider service.TimeProvider,
	logger entities.Logger,
) *AkerunWorker {
	return &AkerunWorker{
		client:             client,
		dailyBonusRepo:     dailyBonusRepo,
		userRepo:           userRepo,
		transactionRepo:    transactionRepo,
		txManager:          txManager,
		systemSettingsRepo: systemSettingsRepo,
		timeProvider:       timeProvider,
		logger:             logger,
		interval:           5 * time.Minute,
		recoverySleep:      1 * time.Minute,
		stopCh:             make(chan struct{}),
	}
}

// Start はポーリングを開始（バックグラウンドgoroutine）
func (w *AkerunWorker) Start() {
	if !w.client.IsConfigured() {
		w.logger.Info("Akerun worker: not configured, skipping")
		return
	}

	w.logger.Info("Akerun worker: starting polling", entities.NewField("interval", w.interval.String()))

	go func() {
		// 起動直後に1回実行
		w.poll()

		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.poll()
			case <-w.stopCh:
				w.logger.Info("Akerun worker: stopped")
				return
			}
		}
	}()
}

// Stop はポーリングを停止
func (w *AkerunWorker) Stop() {
	close(w.stopCh)
}

const (
	// 通常モード
	normalLimit    = 300
	normalInterval = 5 * time.Minute

	// リカバリモード
	recoveryLimit        = 720
	recoverySleep        = 1 * time.Minute
	recoveryWindow       = 1 * time.Hour
	recoveryGapThreshold = 10 * time.Minute // この閾値を超えたらリカバリモード
)

// poll は1回のポーリング処理
func (w *AkerunWorker) poll() {
	ctx := context.Background()

	// 前回ポーリング時刻を取得
	lastPolledAt, err := w.dailyBonusRepo.GetLastPolledAt(ctx)
	if err != nil {
		w.logger.Error("Akerun worker: failed to get last polled time", entities.NewField("error", err))
		return
	}

	now := w.timeProvider.Now()
	gap := now.Sub(lastPolledAt)

	if gap > recoveryGapThreshold {
		// リカバリモード: 1時間ウィンドウで分割取得
		w.pollRecovery(ctx, lastPolledAt, now)
	} else {
		// 通常モード: 一括取得
		w.pollNormal(ctx, lastPolledAt, now)
	}
}

// pollNormal は通常モードのポーリング（5分間隔、limit=300）
func (w *AkerunWorker) pollNormal(ctx context.Context, after, before time.Time) {
	accesses, err := w.client.GetAccesses(ctx, after, before, normalLimit)
	if err != nil {
		w.logger.Error("Akerun worker: failed to get accesses", entities.NewField("error", err))
		return
	}

	w.logger.Info("Akerun worker: fetched accesses",
		entities.NewField("count", len(accesses)),
		entities.NewField("from", after.Format(time.RFC3339)),
		entities.NewField("to", before.Format(time.RFC3339)))

	if len(accesses) > 0 {
		w.processAccesses(ctx, accesses)
	}

	if err := w.dailyBonusRepo.UpdateLastPolledAt(ctx, before); err != nil {
		w.logger.Error("Akerun worker: failed to update last polled time", entities.NewField("error", err))
	}
}

// pollRecovery はリカバリモードのポーリング（1時間ウィンドウ、limit=720）
func (w *AkerunWorker) pollRecovery(ctx context.Context, lastPolledAt, now time.Time) {
	gap := now.Sub(lastPolledAt)
	totalWindows := int(gap/recoveryWindow) + 1

	w.logger.Info("Akerun worker: recovery mode started",
		entities.NewField("gap", gap.String()),
		entities.NewField("lastPolledAt", lastPolledAt.Format(time.RFC3339)),
		entities.NewField("totalWindows", totalWindows))

	cursor := lastPolledAt

	for windowIdx := 0; cursor.Before(now); windowIdx++ {
		end := cursor.Add(recoveryWindow)
		if end.After(now) {
			end = now
		}

		accesses, err := w.client.GetAccesses(ctx, cursor, end, recoveryLimit)
		if err != nil {
			w.logger.Error("Akerun worker: recovery fetch failed",
				entities.NewField("window", windowIdx+1),
				entities.NewField("error", err))
			return // エラー時は中断、次回pollで再開
		}

		w.logger.Info("Akerun worker: recovery window fetched",
			entities.NewField("window", fmt.Sprintf("%d/%d", windowIdx+1, totalWindows)),
			entities.NewField("count", len(accesses)),
			entities.NewField("from", cursor.Format(time.RFC3339)),
			entities.NewField("to", end.Format(time.RFC3339)))

		if len(accesses) >= recoveryLimit {
			w.logger.Warn("Akerun worker: recovery window hit limit, some records may be missed",
				entities.NewField("window", windowIdx+1),
				entities.NewField("limit", recoveryLimit))
		}

		if len(accesses) > 0 {
			w.processAccesses(ctx, accesses)
		}

		// ウィンドウ完了 → last_polled_at を段階的に更新（途中で落ちても再開可能）
		if err := w.dailyBonusRepo.UpdateLastPolledAt(ctx, end); err != nil {
			w.logger.Error("Akerun worker: failed to update last polled time", entities.NewField("error", err))
			return
		}

		cursor = end

		// レートリミット配慮（1分間隔、最後のウィンドウ以降はsleep不要）
		if cursor.Before(now) {
			time.Sleep(w.recoverySleep)
		}
	}

	w.logger.Info("Akerun worker: recovery completed",
		entities.NewField("gap", gap.String()),
		entities.NewField("windows", totalWindows))
}

// processAccesses はアクセス履歴を処理してボーナスを付与
func (w *AkerunWorker) processAccesses(ctx context.Context, accesses []AccessRecord) {
	// ボーナスポイント数を取得（管理者設定）
	bonusPoints := entities.DefaultAkerunBonusPoints
	if pointsStr, err := w.systemSettingsRepo.GetSetting(ctx, "akerun_bonus_points"); err == nil && pointsStr != "" {
		if pts, err := strconv.ParseInt(pointsStr, 10, 64); err == nil && pts > 0 {
			bonusPoints = pts
		}
	}

	// 全ユーザーを取得してマッチング用マップを構築
	users, err := w.userRepo.ReadList(ctx, 0, 10000)
	if err != nil {
		w.logger.Error("Akerun worker: failed to get users", entities.NewField("error", err))
		return
	}

	// ユーザー名マッチングマップ: 正規化された名前 → userID
	nameToUser := make(map[string]uuid.UUID)
	for _, user := range users {
		if user.LastName != "" && user.FirstName != "" {
			// "田中太郎" 形式
			fullName := normalizeName(user.LastName + user.FirstName)
			nameToUser[fullName] = user.ID

			// "田中 太郎" 形式（スペース区切り）もカバー
			fullNameWithSpace := normalizeName(user.LastName + " " + user.FirstName)
			nameToUser[fullNameWithSpace] = user.ID
		}
	}

	for _, access := range accesses {
		if access.User == nil || access.User.Name == "" {
			continue
		}

		// Akerunユーザー名を正規化
		akerunName := normalizeName(access.User.Name)

		// アプリユーザーとマッチング
		userID, matched := nameToUser[akerunName]
		if !matched {
			continue
		}

		// アクセス時刻パース
		accessedAt, err := time.Parse(time.RFC3339, access.AccessedAt)
		if err != nil {
			w.logger.Error("Akerun worker: failed to parse accessed_at",
				entities.NewField("accessed_at", access.AccessedAt),
				entities.NewField("error", err))
			continue
		}

		// ボーナス日付を計算（JST AM6:00区切り）
		bonusDate := entities.GetBonusDateJST(accessedAt)

		// 既にボーナス付与済みかチェック
		existing, err := w.dailyBonusRepo.ReadByUserAndDate(ctx, userID, bonusDate)
		if err != nil {
			w.logger.Error("Akerun worker: failed to check existing bonus",
				entities.NewField("user_id", userID),
				entities.NewField("error", err))
			continue
		}

		if existing != nil {
			// 既に付与済み → スキップ
			continue
		}

		// ボーナスを付与（トランザクション内で実行）
		err = w.txManager.Do(ctx, func(txCtx context.Context) error {
			// ボーナスレコード作成
			accessIDStr := access.ID.String()
			bonus := entities.NewDailyBonus(userID, bonusDate, bonusPoints, accessIDStr, access.User.Name, &accessedAt)
			if err := w.dailyBonusRepo.Create(txCtx, bonus); err != nil {
				return fmt.Errorf("failed to create daily bonus: %w", err)
			}

			// ポイント付与トランザクション
			tx, err := entities.NewAdminGrant(
				userID,
				bonusPoints,
				"Akerun入退室ボーナス",
				uuid.Nil, // システム処理
			)
			if err != nil {
				return fmt.Errorf("failed to create transaction: %w", err)
			}
			if err := w.transactionRepo.Create(txCtx, tx); err != nil {
				return fmt.Errorf("failed to save transaction: %w", err)
			}

			// ユーザー残高更新
			updates := []repository.BalanceUpdate{
				{UserID: userID, Amount: bonusPoints, IsDeduct: false},
			}
			if err := w.userRepo.UpdateBalancesWithLock(txCtx, updates); err != nil {
				return fmt.Errorf("failed to update balance: %w", err)
			}

			return nil
		})

		if err != nil {
			w.logger.Error("Akerun worker: failed to grant bonus",
				entities.NewField("user_id", userID),
				entities.NewField("akerun_user", access.User.Name),
				entities.NewField("error", err))
		} else {
			w.logger.Info("Akerun worker: bonus granted",
				entities.NewField("user_id", userID),
				entities.NewField("akerun_user", access.User.Name),
				entities.NewField("points", bonusPoints),
				entities.NewField("date", bonusDate.Format("2006-01-02")))
		}
	}
}

// normalizeName は名前を正規化（全角/半角スペース除去、小文字化）
func normalizeName(name string) string {
	// 全角スペースを半角に統一
	name = strings.ReplaceAll(name, "　", " ")
	// スペースを除去
	name = strings.ReplaceAll(name, " ", "")
	// 小文字化（英語名の場合のため）
	name = strings.ToLower(name)
	// 前後の空白を除去
	name = strings.TrimSpace(name)
	return name
}

// PollForTest はテスト用にpollをエクスポート
func (w *AkerunWorker) PollForTest() {
	w.poll()
}

// SetRecoverySleepForTest はテスト用にrecoverySleepをオーバーライド
func (w *AkerunWorker) SetRecoverySleepForTest(d time.Duration) {
	w.recoverySleep = d
}

// ProcessAccessesForTest はテスト用にprocessAccessesをエクスポート
func (w *AkerunWorker) ProcessAccessesForTest(ctx context.Context, accesses []AccessRecord) {
	w.processAccesses(ctx, accesses)
}

// NormalizeName はnormalizeNameのエクスポート版
func NormalizeName(name string) string {
	return normalizeName(name)
}
