package inframysql

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// contextKey はcontext内でトランザクションを保持するためのキー
type contextKey string

const txKey contextKey = "tx"

// maxRetries はシリアライゼーションエラー時の最大リトライ回数
const maxRetries = 3

// GormTransactionManager はGORMを使ったTransactionManagerの実装
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager は新しいGormTransactionManagerを生成します
func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

// Do は関数fnをトランザクション内で実行します。
// fn内でエラーが返ればRollback、nilならCommitされます。
// REPEATABLE READでのシリアライゼーションエラー（SQLSTATE 40001）が発生した場合、
// exponential backoff + ジッターで最大3回リトライします。
func (tm *GormTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// exponential backoff + jitter: 50ms, 100ms, 200ms + ランダムジッター
			backoff := time.Duration(50<<(attempt-1)) * time.Millisecond
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			time.Sleep(backoff + jitter)
		}

		lastErr = tm.doOnce(ctx, fn)
		if lastErr == nil {
			return nil
		}

		// シリアライゼーションエラー（SQLSTATE 40001）の場合のみリトライ
		var pgErr *pgconn.PgError
		if !errors.As(lastErr, &pgErr) || pgErr.Code != "40001" {
			return lastErr
		}
	}

	return fmt.Errorf("transaction failed after %d retries due to serialization conflict: %w", maxRetries, lastErr)
}

// doOnce は1回のトランザクション実行を行います
func (tm *GormTransactionManager) doOnce(ctx context.Context, fn func(ctx context.Context) error) error {
	// トランザクション開始
	tx := tm.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// panic時のRollback対応
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // panicは再throw
		}
	}()

	// トランザクションをcontextに保存
	ctxWithTx := context.WithValue(ctx, txKey, tx)

	// 関数を実行
	if err := fn(ctxWithTx); err != nil {
		// エラー時はRollback
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("failed to rollback (original error: %v): %w", err, rbErr)
		}
		return err
	}

	// 成功時はCommit
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDB はcontextからトランザクションを取得します
// トランザクションが存在しない場合はdefaultDBを返します
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}
