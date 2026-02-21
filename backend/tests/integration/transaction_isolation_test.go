//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ========================================
// Transaction Isolation Level Tests
// ========================================

func TestTransactionIsolation_RepeatableRead(t *testing.T) {
	db := setupIntegrationDB(t)

	userDS := dsmysqlimpl.NewUserDataSource(db)

	// テストユーザーを作成
	testUser, err := entities.NewUser("isolation_test", "isolation@test.com", "hash", "Isolation Test", "Test", "User")
	testUser.Balance = 10000
	require.NoError(t, err)
	require.NoError(t, userDS.Insert(context.Background(), testUser))

	t.Run("REPEATABLE READ: トランザクション内で一貫したスナップショット", func(t *testing.T) {
		var balance1, balance2 int64

		gormDB := db.GetDB()

		// トランザクション1: 残高を2回読み取る
		err := gormDB.Transaction(func(tx *gorm.DB) error {
			// 1回目の読み取り
			var user1 entities.User
			if err := tx.Where("id = ?", testUser.ID).First(&user1).Error; err != nil {
				return err
			}
			balance1 = user1.Balance

			// 別のゴルーチンで残高を更新（コミット済み）
			go func() {
				gormDB.Model(&entities.User{}).Where("id = ?", testUser.ID).Update("balance", 5000)
			}()
			time.Sleep(100 * time.Millisecond) // 更新完了を待つ

			// 2回目の読み取り（REPEATABLE READなら同じ値が読める）
			var user2 entities.User
			if err := tx.Where("id = ?", testUser.ID).First(&user2).Error; err != nil {
				return err
			}
			balance2 = user2.Balance

			return nil
		})

		require.NoError(t, err)
		// REPEATABLE READの場合、トランザクション内では同じ値が読める
		assert.Equal(t, balance1, balance2, "REPEATABLE READ isolation should maintain consistent snapshot")
		assert.Equal(t, int64(10000), balance1, "First read should see original value")
	})
}

func TestTransactionIsolation_PhantomRead(t *testing.T) {
	db := setupIntegrationDB(t)

	gormDB := db.GetDB()

	t.Run("REPEATABLE READ: Phantom Readの防止", func(t *testing.T) {
		// トランザクション1: 複数回のSELECT
		var count1, count2 int64

		err := gormDB.Transaction(func(tx *gorm.DB) error {
			// 1回目のカウント
			tx.Model(&entities.User{}).Where("balance > ?", 5000).Count(&count1)

			// 別のゴルーチンで新しいユーザーを追加
			go func() {
				newUser, _ := entities.NewUser(
					fmt.Sprintf("phantom_%d", time.Now().UnixNano()),
					fmt.Sprintf("phantom_%d@test.com", time.Now().UnixNano()),
					"hash",
					"Phantom User",
					"Phantom",
					"User",
				)
				newUser.Balance = 10000
				gormDB.Create(newUser)
			}()
			time.Sleep(100 * time.Millisecond)

			// 2回目のカウント（REPEATABLE READならPhantom Readは防止される）
			tx.Model(&entities.User{}).Where("balance > ?", 5000).Count(&count2)

			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, count1, count2, "REPEATABLE READ should prevent phantom reads")
	})
}

// ========================================
// Concurrent Transaction Tests
// ========================================

func TestConcurrentTransactions_DeadlockPrevention(t *testing.T) {
	db := setupIntegrationDB(t)

	userDS := dsmysqlimpl.NewUserDataSource(db)
	gormDB := db.GetDB()

	// 2人のユーザーを作成
	userA, _ := entities.NewUser("userA", "userA@test.com", "hash", "User A", "A", "User")
	userA.Balance = 5000
	userB, _ := entities.NewUser("userB", "userB@test.com", "hash", "User B", "B", "User")
	userB.Balance = 5000

	require.NoError(t, userDS.Insert(context.Background(), userA))
	require.NoError(t, userDS.Insert(context.Background(), userB))

	t.Run("UUID順序ロックによるデッドロック回避", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 2)

		// トランザクション1: A -> B にロック
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := gormDB.Transaction(func(tx *gorm.DB) error {
				// UUID順序でロックを取得
				firstID, secondID := userA.ID, userB.ID
				if userB.ID.String() < userA.ID.String() {
					firstID, secondID = userB.ID, userA.ID
				}

				if err := userDS.UpdateBalanceWithLock(context.Background(), firstID, 100, true); err != nil {
					return err
				}
				time.Sleep(50 * time.Millisecond) // 競合を誘発
				if err := userDS.UpdateBalanceWithLock(context.Background(), secondID, 100, false); err != nil {
					return err
				}
				return nil
			})
			errors <- err
		}()

		// トランザクション2: B -> A にロック（UUID順序なので実際は同じ順序）
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond) // 少し遅らせて競合を誘発
			err := gormDB.Transaction(func(tx *gorm.DB) error {
				firstID, secondID := userB.ID, userA.ID
				if userA.ID.String() < userB.ID.String() {
					firstID, secondID = userA.ID, userB.ID
				}

				if err := userDS.UpdateBalanceWithLock(context.Background(), firstID, 100, true); err != nil {
					return err
				}
				time.Sleep(50 * time.Millisecond)
				if err := userDS.UpdateBalanceWithLock(context.Background(), secondID, 100, false); err != nil {
					return err
				}
				return nil
			})
			errors <- err
		}()

		wg.Wait()
		close(errors)

		// デッドロックが発生しないことを確認
		for err := range errors {
			assert.NoError(t, err, "Deadlock should be prevented by UUID ordering")
		}
	})
}

func TestConcurrentTransactions_RaceCondition(t *testing.T) {
	db := setupIntegrationDB(t)

	userDS := dsmysqlimpl.NewUserDataSource(db)

	// テストユーザーを作成
	testUser, _ := entities.NewUser("concurrent_test", "concurrent@test.com", "hash", "Concurrent Test", "Test", "User")
	testUser.Balance = 100000
	require.NoError(t, userDS.Insert(context.Background(), testUser))

	t.Run("100並行トランザクション: 残高の整合性確認", func(t *testing.T) {
		var wg sync.WaitGroup
		concurrency := 100
		deductAmount := int64(100)

		initialBalance := testUser.Balance
		gormDB := db.GetDB()

		// 100個の並行トランザクション
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				gormDB.Transaction(func(tx *gorm.DB) error {
					return userDS.UpdateBalanceWithLock(context.Background(), testUser.ID, deductAmount, true)
				})
			}(i)
		}

		wg.Wait()

		// 最終残高を確認
		finalUser, err := userDS.Select(context.Background(), testUser.ID)
		require.NoError(t, err)

		expectedBalance := initialBalance - (int64(concurrency) * deductAmount)
		assert.Equal(t, expectedBalance, finalUser.Balance,
			"Final balance should be exactly initial - (concurrency * amount)")

		t.Logf("Initial: %d, Final: %d, Expected: %d", initialBalance, finalUser.Balance, expectedBalance)
	})
}

func TestConcurrentTransactions_LostUpdate(t *testing.T) {
	db := setupIntegrationDB(t)

	userDS := dsmysqlimpl.NewUserDataSource(db)
	gormDB := db.GetDB()

	testUser, _ := entities.NewUser("lost_update_test", "lost@test.com", "hash", "Lost Update Test", "Test", "User")
	testUser.Balance = 10000
	require.NoError(t, userDS.Insert(context.Background(), testUser))

	t.Run("SELECT FOR UPDATE による Lost Update 防止", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make([]int64, 2)

		// トランザクション1
		wg.Add(1)
		go func() {
			defer wg.Done()
			gormDB.Transaction(func(tx *gorm.DB) error {
				var user entities.User
				tx.Clauses().Where("id = ?", testUser.ID).First(&user)
				time.Sleep(100 * time.Millisecond)
				user.Balance -= 1000
				tx.Save(&user)
				results[0] = user.Balance
				return nil
			})
		}()

		// トランザクション2（少し遅れて開始）
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
			gormDB.Transaction(func(tx *gorm.DB) error {
				// SELECT FOR UPDATE を使用
				if err := userDS.UpdateBalanceWithLock(context.Background(), testUser.ID, 2000, true); err != nil {
					return err
				}
				return nil
			})
		}()

		wg.Wait()

		// 最終残高を確認
		finalUser, _ := userDS.Select(context.Background(), testUser.ID)

		// Lost Updateが発生していないことを確認
		// 正しくは 10000 - 1000 - 2000 = 7000
		assert.Equal(t, int64(7000), finalUser.Balance,
			"Lost update should be prevented by pessimistic locking")
	})
}

// ========================================
// Point Transfer Integration Tests
// ========================================

func TestPointTransfer_Integration_ConcurrentTransfers(t *testing.T) {
	db := setupIntegrationDB(t)

	userDS := dsmysqlimpl.NewUserDataSource(db)
	transactionDS := dsmysqlimpl.NewTransactionDataSource(db)
	gormDB := db.GetDB()

	// 3人のユーザーを作成
	users := make([]*entities.User, 3)
	for i := 0; i < 3; i++ {
		u, _ := entities.NewUser(
			fmt.Sprintf("user%d", i),
			fmt.Sprintf("user%d@test.com", i),
			"hash",
			fmt.Sprintf("User %d", i),
			"Test",
			"User",
		)
		u.Balance = 10000
		require.NoError(t, userDS.Insert(context.Background(), u))
		users[i] = u
	}

	t.Run("複数ユーザー間の並行送金: ポイント保存則", func(t *testing.T) {
		// 初期総ポイント数
		initialTotal := int64(30000)

		var wg sync.WaitGroup

		// 10個の並行送金
		transfers := []struct {
			from   int
			to     int
			amount int64
		}{
			{0, 1, 500},
			{1, 2, 300},
			{2, 0, 400},
			{0, 2, 600},
			{1, 0, 200},
			{2, 1, 350},
			{0, 1, 450},
			{1, 2, 250},
			{2, 0, 550},
			{0, 2, 300},
		}

		for i, transfer := range transfers {
			wg.Add(1)
			go func(idx int, tf struct {
				from   int
				to     int
				amount int64
			}) {
				defer wg.Done()

				err := gormDB.Transaction(func(tx *gorm.DB) error {
					fromUser := users[tf.from]
					toUser := users[tf.to]

					// UUID順序ロック
					firstID, secondID := fromUser.ID, toUser.ID
					firstIsFrom := true
					if toUser.ID.String() < fromUser.ID.String() {
						firstID, secondID = toUser.ID, fromUser.ID
						firstIsFrom = false
					}

					if firstIsFrom {
						if err := userDS.UpdateBalanceWithLock(context.Background(), firstID, tf.amount, true); err != nil {
							return err
						}
						if err := userDS.UpdateBalanceWithLock(context.Background(), secondID, tf.amount, false); err != nil {
							return err
						}
					} else {
						if err := userDS.UpdateBalanceWithLock(context.Background(), firstID, tf.amount, false); err != nil {
							return err
						}
						if err := userDS.UpdateBalanceWithLock(context.Background(), secondID, tf.amount, true); err != nil {
							return err
						}
					}

					// トランザクション記録作成
					txRecord, _ := entities.NewTransfer(
						fromUser.ID,
						toUser.ID,
						tf.amount,
						fmt.Sprintf("concurrent-key-%d-%d", idx, time.Now().UnixNano()),
						"Concurrent transfer test",
					)
					txRecord.Complete()
					return transactionDS.Insert(context.Background(), txRecord)
				})

				if err != nil {
					t.Logf("Transfer %d failed: %v", idx, err)
				}
			}(i, transfer)

			time.Sleep(10 * time.Millisecond) // 少し間隔を空ける
		}

		wg.Wait()

		// 最終的なポイント総数を確認
		finalTotal := int64(0)
		for i := 0; i < 3; i++ {
			u, _ := userDS.Select(context.Background(), users[i].ID)
			finalTotal += u.Balance
			t.Logf("User %d final balance: %d", i, u.Balance)
		}

		// ポイント保存則: 総ポイント数は変わらない
		assert.Equal(t, initialTotal, finalTotal,
			"Total points must be conserved across all concurrent transfers")
	})
}

func TestPointTransfer_Integration_Idempotency(t *testing.T) {
	db := setupIntegrationDB(t)

	userDS := dsmysqlimpl.NewUserDataSource(db)
	gormDB := db.GetDB()

	sender, _ := entities.NewUser("sender", "sender@test.com", "hash", "Sender", "Test", "User")
	sender.Balance = 10000
	receiver, _ := entities.NewUser("receiver", "receiver@test.com", "hash", "Receiver", "Test", "User")
	receiver.Balance = 5000

	require.NoError(t, userDS.Insert(context.Background(), sender))
	require.NoError(t, userDS.Insert(context.Background(), receiver))

	t.Run("同じIdempotencyKeyでの並行リクエスト: 1回のみ実行", func(t *testing.T) {
		idempotencyKey := fmt.Sprintf("idempotency-test-%d", time.Now().UnixNano())
		amount := int64(1000)

		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		// 10個の並行リクエスト（同じIdempotencyKey）
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				err := gormDB.Transaction(func(tx *gorm.DB) error {
					// Idempotency keyチェック（簡易実装）
					var existing entities.Transaction
					if err := tx.Where("idempotency_key = ?", idempotencyKey).First(&existing).Error; err == nil {
						return fmt.Errorf("duplicate idempotency key")
					}

					// 残高更新
					if err := userDS.UpdateBalanceWithLock(context.Background(), sender.ID, amount, true); err != nil {
						return err
					}
					if err := userDS.UpdateBalanceWithLock(context.Background(), receiver.ID, amount, false); err != nil {
						return err
					}

					// トランザクション記録
					txRecord, _ := entities.NewTransfer(sender.ID, receiver.ID, amount, idempotencyKey, "Idempotency test")
					txRecord.Complete()

					// DB保存（簡易実装のため直接GORMを使用）
					return tx.Create(txRecord).Error
				})

				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		// 1回のみ成功することを確認
		assert.Equal(t, 1, successCount, "Only one transaction should succeed with the same idempotency key")

		// 残高確認
		finalSender, _ := userDS.Select(context.Background(), sender.ID)
		finalReceiver, _ := userDS.Select(context.Background(), receiver.ID)

		assert.Equal(t, int64(9000), finalSender.Balance, "Sender balance should be deducted once")
		assert.Equal(t, int64(6000), finalReceiver.Balance, "Receiver balance should be credited once")
	})
}

// ========================================
// Transaction Rollback Tests
// ========================================

func TestTransactionRollback_OnError(t *testing.T) {
	db := setupIntegrationDB(t)

	userDS := dsmysqlimpl.NewUserDataSource(db)
	gormDB := db.GetDB()

	testUser, _ := entities.NewUser("rollback_test", "rollback@test.com", "hash", "Rollback Test", "Test", "User")
	testUser.Balance = 5000
	require.NoError(t, userDS.Insert(context.Background(), testUser))

	t.Run("エラー時のロールバック確認", func(t *testing.T) {
		initialBalance := testUser.Balance

		// トランザクション内でエラーを発生させる
		err := gormDB.Transaction(func(tx *gorm.DB) error {
			// 残高を減算
			if err := userDS.UpdateBalanceWithLock(context.Background(), testUser.ID, 2000, true); err != nil {
				return err
			}

			// 意図的にエラーを発生
			return fmt.Errorf("intentional error for rollback test")
		})

		require.Error(t, err)

		// 残高が変わっていないことを確認（ロールバック成功）
		finalUser, _ := userDS.Select(context.Background(), testUser.ID)
		assert.Equal(t, initialBalance, finalUser.Balance,
			"Balance should be rolled back on transaction error")
	})
}

func TestTransactionTimeout(t *testing.T) {
	db := setupIntegrationDB(t)

	gormDB := db.GetDB()

	t.Run("長時間トランザクションのタイムアウト", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := gormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 2秒待機（タイムアウトより長い）
			time.Sleep(2 * time.Second)
			return nil
		})

		assert.Error(t, err, "Transaction should timeout")
	})
}
