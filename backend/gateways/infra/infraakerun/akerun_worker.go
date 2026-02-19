package infraakerun

import (
	"context"
	"fmt"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/service"
)

// AkerunWorker はAkerun入退室ポーリングワーカー
// ポーリング制御のみを担当し、ビジネスロジックはAkerunBonusInputPortに委譲する
type AkerunWorker struct {
	gateway       service.AkerunAccessGateway
	interactor    inputport.AkerunBonusInputPort
	timeProvider  service.TimeProvider
	logger        entities.Logger
	interval      time.Duration
	recoverySleep time.Duration
	stopCh        chan struct{}
}

// NewAkerunWorker は新しいAkerunWorkerを作成
func NewAkerunWorker(
	gateway service.AkerunAccessGateway,
	interactor inputport.AkerunBonusInputPort,
	timeProvider service.TimeProvider,
	logger entities.Logger,
) *AkerunWorker {
	return &AkerunWorker{
		gateway:       gateway,
		interactor:    interactor,
		timeProvider:  timeProvider,
		logger:        logger,
		interval:      5 * time.Minute,
		recoverySleep: 1 * time.Minute,
		stopCh:        make(chan struct{}),
	}
}

// Start はポーリングを開始（バックグラウンドgoroutine）
func (w *AkerunWorker) Start() {
	if !w.gateway.IsConfigured() {
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
	recoveryLimit        = 720 // 一時間で最高720件までの取得(12回/分 * 60分)
	recoveryWindow       = 1 * time.Hour
	recoveryGapThreshold = 10 * time.Minute // この閾値を超えたらリカバリモード
)

// poll は1回のポーリング処理
func (w *AkerunWorker) poll() {
	ctx := context.Background()

	// 前回ポーリング時刻を取得
	lastPolledAt, err := w.interactor.GetLastPolledAt(ctx)
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
	accesses, err := w.gateway.FetchAccesses(ctx, after, before, normalLimit)
	if err != nil {
		w.logger.Error("Akerun worker: failed to get accesses", entities.NewField("error", err))
		return
	}

	w.logger.Info("Akerun worker: fetched accesses",
		entities.NewField("count", len(accesses)),
		entities.NewField("from", after.Format(time.RFC3339)),
		entities.NewField("to", before.Format(time.RFC3339)))

	if len(accesses) > 0 {
		if err := w.interactor.ProcessAccesses(ctx, accesses); err != nil {
			w.logger.Error("Akerun worker: failed to process accesses", entities.NewField("error", err))
		}
	}

	if err := w.interactor.UpdateLastPolledAt(ctx, before); err != nil {
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

		accesses, err := w.gateway.FetchAccesses(ctx, cursor, end, recoveryLimit)
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
			if err := w.interactor.ProcessAccesses(ctx, accesses); err != nil {
				w.logger.Error("Akerun worker: failed to process accesses in recovery",
					entities.NewField("error", err))
			}
		}

		// ウィンドウ完了 → last_polled_at を段階的に更新（途中で落ちても再開可能）
		if err := w.interactor.UpdateLastPolledAt(ctx, end); err != nil {
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

// PollForTest はテスト用にpollをエクスポート
func (w *AkerunWorker) PollForTest() {
	w.poll()
}

// SetRecoverySleepForTest はテスト用にrecoverySleepをオーバーライド
func (w *AkerunWorker) SetRecoverySleepForTest(d time.Duration) {
	w.recoverySleep = d
}
