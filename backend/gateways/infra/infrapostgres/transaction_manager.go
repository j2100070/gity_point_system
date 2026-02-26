package infrapostgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// contextKey はcontext内でトランザクションを保持するためのキー
type contextKey string

const txKey contextKey = "tx"

// GormTransactionManager はGORMを使ったTransactionManagerの実装
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager は新しいGormTransactionManagerを生成します
func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

// Do は関数fnをトランザクション内で実行します
// fn内でエラーが返ればRollback、nilならCommitされます
func (tm *GormTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
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
