package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// PointExpiryWorker はポイント期限切れ処理ワーカー
// 毎時実行し、期限切れのポイントバッチを検出・失効処理する
type PointExpiryWorker struct {
	pointBatchRepo  repository.PointBatchRepository
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	txManager       repository.TransactionManager
	logger          entities.Logger
	interval        time.Duration
	batchSize       int
	stopCh          chan struct{}
}

// NewPointExpiryWorker は新しいPointExpiryWorkerを作成
func NewPointExpiryWorker(
	pointBatchRepo repository.PointBatchRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	txManager repository.TransactionManager,
	logger entities.Logger,
) *PointExpiryWorker {
	return &PointExpiryWorker{
		pointBatchRepo:  pointBatchRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		txManager:       txManager,
		logger:          logger,
		interval:        1 * time.Hour,
		batchSize:       100,
		stopCh:          make(chan struct{}),
	}
}

// Start はワーカーを開始
func (w *PointExpiryWorker) Start() {
	w.logger.Info("PointExpiryWorker started", entities.NewField("interval", w.interval.String()))

	go func() {
		// 初回実行
		w.processExpiredBatches()

		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.processExpiredBatches()
			case <-w.stopCh:
				w.logger.Info("PointExpiryWorker stopped")
				return
			}
		}
	}()
}

// Stop はワーカーを停止
func (w *PointExpiryWorker) Stop() {
	close(w.stopCh)
}

// processExpiredBatches は期限切れバッチを処理
func (w *PointExpiryWorker) processExpiredBatches() {
	ctx := context.Background()
	now := time.Now()

	totalExpired := 0
	totalPoints := int64(0)

	for {
		batches, err := w.pointBatchRepo.FindExpiredBatches(ctx, now, w.batchSize)
		if err != nil {
			w.logger.Error("PointExpiryWorker: failed to find expired batches",
				entities.NewField("error", err))
			return
		}

		if len(batches) == 0 {
			break
		}

		for _, batch := range batches {
			if err := w.expireBatch(ctx, batch); err != nil {
				w.logger.Error("PointExpiryWorker: failed to expire batch",
					entities.NewField("batch_id", batch.ID),
					entities.NewField("user_id", batch.UserID),
					entities.NewField("error", err))
				continue
			}
			totalExpired++
			totalPoints += batch.RemainingAmount
		}

		// バッチサイズ未満 = もうデータなし
		if len(batches) < w.batchSize {
			break
		}
	}

	if totalExpired > 0 {
		w.logger.Info("PointExpiryWorker: completed",
			entities.NewField("expired_batches", totalExpired),
			entities.NewField("expired_points", totalPoints))
	}
}

// expireBatch は1つのバッチを失効処理
func (w *PointExpiryWorker) expireBatch(ctx context.Context, batch *entities.PointBatch) error {
	return w.txManager.Do(ctx, func(txCtx context.Context) error {
		// 1. ユーザー残高から減算
		if err := w.userRepo.UpdateBalanceWithLock(txCtx, batch.UserID, batch.RemainingAmount, true); err != nil {
			return fmt.Errorf("failed to deduct expired points: %w", err)
		}

		// 2. system_expire トランザクション記録
		fromUserID := batch.UserID
		tx := &entities.Transaction{
			ID:              uuid.New(),
			FromUserID:      &fromUserID,
			ToUserID:        nil, // システムへの返却
			Amount:          batch.RemainingAmount,
			TransactionType: entities.TransactionTypeSystemExpire,
			Status:          entities.TransactionStatusCompleted,
			Description:     fmt.Sprintf("ポイント期限切れ（バッチ: %s）", batch.ID),
			Metadata:        map[string]interface{}{"batch_id": batch.ID.String()},
			CreatedAt:       time.Now(),
			CompletedAt:     ptrTime(time.Now()),
		}

		if err := w.transactionRepo.Create(txCtx, tx); err != nil {
			return fmt.Errorf("failed to create expire transaction: %w", err)
		}

		// 3. バッチの remaining_amount を 0 に
		if err := w.pointBatchRepo.MarkExpired(txCtx, batch.ID); err != nil {
			return fmt.Errorf("failed to mark batch expired: %w", err)
		}

		return nil
	})
}

// ptrTime はtime.Timeのポインタを返すヘルパー
func ptrTime(t time.Time) *time.Time {
	return &t
}

// ProcessExpiredBatchesForTest はテスト用にprocessExpiredBatchesをエクスポート
func (w *PointExpiryWorker) ProcessExpiredBatchesForTest() {
	w.processExpiredBatches()
}
