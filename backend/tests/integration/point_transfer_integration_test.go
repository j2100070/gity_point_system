//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/gity/point-system/usecases/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPointTransfer(t *testing.T) (*interactor.PointTransferInteractor, inframysql.DB) {
	t.Helper()
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	txManager := inframysql.NewGormTransactionManager(db.GetDB())

	pt := interactor.NewPointTransferInteractor(
		txManager, repos.User, repos.Transaction, repos.IdempotencyKey, repos.Friendship, repos.PointBatch, lg,
	)
	return pt, db
}

// TestPointTransfer_Success は送金成功フローを検証
func TestPointTransfer_Success(t *testing.T) {
	pt, db := setupPointTransfer(t)
	ctx := context.Background()

	alice := createTestUserWithBalance(t, db, "alice_transfer", 1000)
	bob := createTestUser(t, db, "bob_transfer")

	// フレンド関係を直接DBに挿入（送金にはフレンド必須の場合）
	db.GetDB().Exec(
		`INSERT INTO friendships (id, requester_id, addressee_id, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), ?, ?, 'accepted', NOW(), NOW())`,
		alice.ID, bob.ID,
	)

	resp, err := pt.Transfer(ctx, &inputport.TransferRequest{
		FromUserID:     alice.ID,
		ToUserID:       bob.ID,
		Amount:         300,
		IdempotencyKey: "integ-transfer-001",
		Description:    "integration test transfer",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, int64(700), resp.FromUser.Balance)
	assert.Equal(t, int64(300), resp.ToUser.Balance)
	assert.Equal(t, int64(300), resp.Transaction.Amount)
}

// TestPointTransfer_InsufficientBalance は残高不足でエラーになることを検証
func TestPointTransfer_InsufficientBalance(t *testing.T) {
	pt, db := setupPointTransfer(t)
	ctx := context.Background()

	alice := createTestUserWithBalance(t, db, "alice_insuf", 100)
	bob := createTestUser(t, db, "bob_insuf")

	db.GetDB().Exec(
		`INSERT INTO friendships (id, requester_id, addressee_id, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), ?, ?, 'accepted', NOW(), NOW())`,
		alice.ID, bob.ID,
	)

	_, err := pt.Transfer(ctx, &inputport.TransferRequest{
		FromUserID:     alice.ID,
		ToUserID:       bob.ID,
		Amount:         500,
		IdempotencyKey: "integ-transfer-insuf",
	})
	assert.Error(t, err)
}

// TestPointTransfer_Idempotency は冪等性キーの重複処理を検証
func TestPointTransfer_Idempotency(t *testing.T) {
	pt, db := setupPointTransfer(t)
	ctx := context.Background()

	alice := createTestUserWithBalance(t, db, "alice_idemp", 1000)
	bob := createTestUser(t, db, "bob_idemp")

	db.GetDB().Exec(
		`INSERT INTO friendships (id, requester_id, addressee_id, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), ?, ?, 'accepted', NOW(), NOW())`,
		alice.ID, bob.ID,
	)

	req := &inputport.TransferRequest{
		FromUserID:     alice.ID,
		ToUserID:       bob.ID,
		Amount:         200,
		IdempotencyKey: "integ-transfer-idemp",
	}

	// 1回目
	resp1, err := pt.Transfer(ctx, req)
	require.NoError(t, err)

	// 2回目（同一キー → 同一結果を返す or エラー）
	resp2, err := pt.Transfer(ctx, req)
	if err == nil {
		// 冪等レスポンスの場合、同一トランザクションを返す
		assert.Equal(t, resp1.Transaction.ID, resp2.Transaction.ID)
	}
	// エラーが返る実装でも OK

	// 残高は1回分のみ減算されていること
	balResp, err := pt.GetBalance(ctx, &inputport.GetBalanceRequest{UserID: alice.ID})
	require.NoError(t, err)
	assert.Equal(t, int64(800), balResp.Balance)
}

// TestPointTransfer_GetBalance は残高取得を検証
func TestPointTransfer_GetBalance(t *testing.T) {
	pt, db := setupPointTransfer(t)
	ctx := context.Background()

	user := createTestUserWithBalance(t, db, "balance_user", 5000)

	resp, err := pt.GetBalance(ctx, &inputport.GetBalanceRequest{UserID: user.ID})
	require.NoError(t, err)
	assert.Equal(t, int64(5000), resp.Balance)
}

// TestPointTransfer_GetHistory は送金後の履歴取得を検証
func TestPointTransfer_GetHistory(t *testing.T) {
	pt, db := setupPointTransfer(t)
	ctx := context.Background()

	alice := createTestUserWithBalance(t, db, "alice_hist", 1000)
	bob := createTestUser(t, db, "bob_hist")

	db.GetDB().Exec(
		`INSERT INTO friendships (id, requester_id, addressee_id, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), ?, ?, 'accepted', NOW(), NOW())`,
		alice.ID, bob.ID,
	)

	// 送金
	_, err := pt.Transfer(ctx, &inputport.TransferRequest{
		FromUserID:     alice.ID,
		ToUserID:       bob.ID,
		Amount:         100,
		IdempotencyKey: "integ-hist-001",
	})
	require.NoError(t, err)

	// 履歴取得
	histResp, err := pt.GetTransactionHistory(ctx, &inputport.GetTransactionHistoryRequest{
		UserID: alice.ID,
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(histResp.Transactions), 1)
}

// TestPointTransfer_GetExpiringPoints は失効予定ポイント取得を検証
func TestPointTransfer_GetExpiringPoints(t *testing.T) {
	pt, db := setupPointTransfer(t)
	ctx := context.Background()

	user := createTestUser(t, db, "expiring_user")

	resp, err := pt.GetExpiringPoints(ctx, &inputport.GetExpiringPointsRequest{UserID: user.ID})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	// 新規ユーザーは失効予定ポイントなし
	assert.Equal(t, int64(0), resp.TotalExpiring)
}

// setupPointTransferWithTxManager はトランザクションマネージャーも返す（他テストで再利用）
func setupPointTransferWithTxManager(t *testing.T) (*interactor.PointTransferInteractor, *Repos, repository.TransactionManager, inframysql.DB) {
	t.Helper()
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	txManager := inframysql.NewGormTransactionManager(db.GetDB())

	pt := interactor.NewPointTransferInteractor(
		txManager, repos.User, repos.Transaction, repos.IdempotencyKey, repos.Friendship, repos.PointBatch, lg,
	)
	return pt, repos, txManager, db
}
