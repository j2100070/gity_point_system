package interactor_test

import (
	"context"
	"testing"

	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// PointTransferInteractor テスト
// ========================================

// --- Transfer ---

func TestPointTransferInteractor_Transfer(t *testing.T) {
	setup := func() (*ctxTrackingTxManager, *ctxTrackingUserRepo, *ctxTrackingTransactionRepo, *ctxTrackingIdempotencyRepo, *ctxTrackingPointBatchRepo, *interactor.PointTransferInteractor) {
		txMgr := &ctxTrackingTxManager{}
		userRepo := newCtxTrackingUserRepo()
		txRepo := newCtxTrackingTransactionRepo()
		idempRepo := newCtxTrackingIdempotencyRepo()
		friendRepo := newCtxTrackingFriendshipRepo()
		pbRepo := newCtxTrackingPointBatchRepo()
		logger := &mockLogger{}

		i := interactor.NewPointTransferInteractor(txMgr, userRepo, txRepo, idempRepo, friendRepo, pbRepo, logger)
		return txMgr, userRepo, txRepo, idempRepo, pbRepo, i
	}

	t.Run("正常にポイント転送できる", func(t *testing.T) {
		_, userRepo, _, _, _, sut := setup()
		sender := createTestUserWithBalance(t, "sender", 10000, "user")
		receiver := createTestUserWithBalance(t, "receiver", 1000, "user")
		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		resp, err := sut.Transfer(context.Background(), &inputport.TransferRequest{
			FromUserID: sender.ID, ToUserID: receiver.ID, Amount: 500,
			IdempotencyKey: "transfer-" + uuid.New().String(),
			Description:    "test transfer",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Transaction)
	})

	t.Run("txManager.Do内の全呼び出しがトランザクションコンテキストを使用する", func(t *testing.T) {
		txMgr, userRepo, txRepo, idempRepo, pbRepo, sut := setup()
		sender := createTestUserWithBalance(t, "sender", 10000, "user")
		receiver := createTestUserWithBalance(t, "receiver", 1000, "user")
		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		_, err := sut.Transfer(context.Background(), &inputport.TransferRequest{
			FromUserID: sender.ID, ToUserID: receiver.ID, Amount: 100,
			IdempotencyKey: "ctx-transfer-" + uuid.New().String(),
			Description:    "ctx test",
		})
		require.NoError(t, err)
		require.NotNil(t, txMgr.TxCtx)

		// txManager.Do 内の残高更新・トランザクション記録がトランザクションコンテキストを使用
		assert.True(t, isTxContext(userRepo.ctxRecords["UpdateBalancesWithLock"]),
			"userRepo.UpdateBalancesWithLock はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(txRepo.ctxRecords["Create"]),
			"transactionRepo.Create はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(txRepo.ctxRecords["Update"]),
			"transactionRepo.Update はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(pbRepo.ctxRecords["ConsumePointsFIFO"]),
			"pointBatchRepo.ConsumePointsFIFO はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(pbRepo.ctxRecords["Create"]),
			"pointBatchRepo.Create はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(idempRepo.ctxRecords["Update"]),
			"idempotencyRepo.Update はトランザクションコンテキストを使用すべき")

	})

	t.Run("金額が0以下ならエラー", func(t *testing.T) {
		_, userRepo, _, _, _, sut := setup()
		sender := createTestUserWithBalance(t, "sender", 10000, "user")
		receiver := createTestUserWithBalance(t, "receiver", 1000, "user")
		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		_, err := sut.Transfer(context.Background(), &inputport.TransferRequest{
			FromUserID: sender.ID, ToUserID: receiver.ID, Amount: 0,
			IdempotencyKey: "key", Description: "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("自分自身への転送はエラー", func(t *testing.T) {
		_, userRepo, _, _, _, sut := setup()
		sender := createTestUserWithBalance(t, "sender", 10000, "user")
		userRepo.setUser(sender)

		_, err := sut.Transfer(context.Background(), &inputport.TransferRequest{
			FromUserID: sender.ID, ToUserID: sender.ID, Amount: 100,
			IdempotencyKey: "key", Description: "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot transfer to the same user")
	})

	t.Run("送信者が存在しない場合エラー", func(t *testing.T) {
		_, userRepo, _, _, _, sut := setup()
		receiver := createTestUserWithBalance(t, "receiver", 1000, "user")
		userRepo.setUser(receiver)

		_, err := sut.Transfer(context.Background(), &inputport.TransferRequest{
			FromUserID: uuid.New(), ToUserID: receiver.ID, Amount: 100,
			IdempotencyKey: "key", Description: "test",
		})
		assert.Error(t, err)
	})

	t.Run("送信者が非アクティブならエラー", func(t *testing.T) {
		_, userRepo, _, _, _, sut := setup()
		sender := createTestUserWithBalance(t, "sender", 10000, "user")
		sender.IsActive = false
		receiver := createTestUserWithBalance(t, "receiver", 1000, "user")
		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		_, err := sut.Transfer(context.Background(), &inputport.TransferRequest{
			FromUserID: sender.ID, ToUserID: receiver.ID, Amount: 100,
			IdempotencyKey: "key", Description: "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})

	t.Run("受信者が非アクティブならエラー", func(t *testing.T) {
		_, userRepo, _, _, _, sut := setup()
		sender := createTestUserWithBalance(t, "sender", 10000, "user")
		receiver := createTestUserWithBalance(t, "receiver", 1000, "user")
		receiver.IsActive = false
		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		_, err := sut.Transfer(context.Background(), &inputport.TransferRequest{
			FromUserID: sender.ID, ToUserID: receiver.ID, Amount: 100,
			IdempotencyKey: "key", Description: "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})

	t.Run("冪等性キーが処理済みの場合は既存の結果を返す", func(t *testing.T) {
		_, userRepo, _, _, _, sut := setup()
		sender := createTestUserWithBalance(t, "sender", 10000, "user")
		receiver := createTestUserWithBalance(t, "receiver", 1000, "user")
		userRepo.setUser(sender)
		userRepo.setUser(receiver)
		key := "idempotent-transfer-" + uuid.New().String()

		resp1, err := sut.Transfer(context.Background(), &inputport.TransferRequest{
			FromUserID: sender.ID, ToUserID: receiver.ID, Amount: 100,
			IdempotencyKey: key, Description: "test",
		})
		require.NoError(t, err)

		// 2度目の呼び出し（バランスをリセット）
		sender2 := createTestUserWithBalance(t, "sender", 10000, "user")
		sender2.ID = sender.ID
		userRepo.setUser(sender2)

		resp2, err := sut.Transfer(context.Background(), &inputport.TransferRequest{
			FromUserID: sender.ID, ToUserID: receiver.ID, Amount: 100,
			IdempotencyKey: key, Description: "test",
		})
		require.NoError(t, err)
		assert.Equal(t, resp1.Transaction.ID, resp2.Transaction.ID)
	})
}

// --- GetTransactionHistory ---

func TestPointTransferInteractor_GetTransactionHistory(t *testing.T) {
	t.Run("正常にトランザクション履歴を取得できる", func(t *testing.T) {
		userRepo := newCtxTrackingUserRepo()
		txRepo := newCtxTrackingTransactionRepo()
		sut := interactor.NewPointTransferInteractor(
			&ctxTrackingTxManager{}, userRepo, txRepo,
			newCtxTrackingIdempotencyRepo(), newCtxTrackingFriendshipRepo(),
			newCtxTrackingPointBatchRepo(), &mockLogger{},
		)

		user := createTestUserWithBalance(t, "user", 1000, "user")
		userRepo.setUser(user)

		resp, err := sut.GetTransactionHistory(context.Background(), &inputport.GetTransactionHistoryRequest{
			UserID: user.ID, Offset: 0, Limit: 20,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(0), resp.Total)
	})
}

// --- GetBalance ---

func TestPointTransferInteractor_GetBalance(t *testing.T) {
	t.Run("正常に残高を取得できる", func(t *testing.T) {
		userRepo := newCtxTrackingUserRepo()
		sut := interactor.NewPointTransferInteractor(
			&ctxTrackingTxManager{}, userRepo, newCtxTrackingTransactionRepo(),
			newCtxTrackingIdempotencyRepo(), newCtxTrackingFriendshipRepo(),
			newCtxTrackingPointBatchRepo(), &mockLogger{},
		)

		user := createTestUserWithBalance(t, "user", 5000, "user")
		userRepo.setUser(user)

		resp, err := sut.GetBalance(context.Background(), &inputport.GetBalanceRequest{
			UserID: user.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(5000), resp.Balance)
		assert.Equal(t, user.ID, resp.User.ID)
	})

	t.Run("ユーザーが存在しない場合エラー", func(t *testing.T) {
		sut := interactor.NewPointTransferInteractor(
			&ctxTrackingTxManager{}, newCtxTrackingUserRepo(), newCtxTrackingTransactionRepo(),
			newCtxTrackingIdempotencyRepo(), newCtxTrackingFriendshipRepo(),
			newCtxTrackingPointBatchRepo(), &mockLogger{},
		)

		_, err := sut.GetBalance(context.Background(), &inputport.GetBalanceRequest{
			UserID: uuid.New(),
		})
		assert.Error(t, err)
	})
}
