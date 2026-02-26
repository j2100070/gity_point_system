//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	infrapostgres "github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTransferRequest(t *testing.T) (inputport.TransferRequestInputPort, infrapostgres.DB) {
	t.Helper()
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	txManager := infrapostgres.NewGormTransactionManager(db.GetDB())

	pt := interactor.NewPointTransferInteractor(
		txManager, repos.User, repos.Transaction, repos.IdempotencyKey, repos.Friendship, repos.PointBatch, lg,
	)
	tr := interactor.NewTransferRequestInteractor(repos.TransferRequest, repos.User, pt, lg)
	return tr, db
}

// TestTransferRequest_CreateAndApprove は送金リクエスト作成→承認フローを検証
func TestTransferRequest_CreateAndApprove(t *testing.T) {
	tr, db := setupTransferRequest(t)
	ctx := context.Background()

	alice := createTestUserWithBalance(t, db, "alice_treq", 1000)
	bob := createTestUser(t, db, "bob_treq")

	// フレンド関係
	db.GetDB().Exec(
		`INSERT INTO friendships (id, requester_id, addressee_id, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), ?, ?, 'accepted', NOW(), NOW())`,
		alice.ID, bob.ID,
	)

	// リクエスト作成
	createResp, err := tr.CreateTransferRequest(ctx, &inputport.CreateTransferRequestRequest{
		FromUserID:     alice.ID,
		ToUserID:       bob.ID,
		Amount:         300,
		Message:        "test transfer request",
		IdempotencyKey: "integ-treq-001",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)

	// 承認
	approveResp, err := tr.ApproveTransferRequest(ctx, &inputport.ApproveTransferRequestRequest{
		RequestID: createResp.TransferRequest.ID,
		UserID:    bob.ID,
	})
	require.NoError(t, err)
	assert.NotNil(t, approveResp.Transaction)
}

// TestTransferRequest_Reject は送金リクエスト拒否を検証
func TestTransferRequest_Reject(t *testing.T) {
	tr, db := setupTransferRequest(t)
	ctx := context.Background()

	alice := createTestUserWithBalance(t, db, "alice_treq_rej", 1000)
	bob := createTestUser(t, db, "bob_treq_rej")

	// リクエスト作成
	createResp, err := tr.CreateTransferRequest(ctx, &inputport.CreateTransferRequestRequest{
		FromUserID:     alice.ID,
		ToUserID:       bob.ID,
		Amount:         200,
		Message:        "reject test",
		IdempotencyKey: "integ-treq-reject",
	})
	require.NoError(t, err)

	// 拒否
	rejectResp, err := tr.RejectTransferRequest(ctx, &inputport.RejectTransferRequestRequest{
		RequestID: createResp.TransferRequest.ID,
		UserID:    bob.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "rejected", string(rejectResp.TransferRequest.Status))
}

// TestTransferRequest_Cancel は送信者によるキャンセルを検証
func TestTransferRequest_Cancel(t *testing.T) {
	tr, db := setupTransferRequest(t)
	ctx := context.Background()

	alice := createTestUserWithBalance(t, db, "alice_treq_cancel", 1000)
	bob := createTestUser(t, db, "bob_treq_cancel")

	// リクエスト作成
	createResp, err := tr.CreateTransferRequest(ctx, &inputport.CreateTransferRequestRequest{
		FromUserID:     alice.ID,
		ToUserID:       bob.ID,
		Amount:         100,
		Message:        "cancel test",
		IdempotencyKey: "integ-treq-cancel",
	})
	require.NoError(t, err)

	// キャンセル
	cancelResp, err := tr.CancelTransferRequest(ctx, &inputport.CancelTransferRequestRequest{
		RequestID: createResp.TransferRequest.ID,
		UserID:    alice.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "cancelled", string(cancelResp.TransferRequest.Status))
}

// TestTransferRequest_ListPending は保留中リクエスト一覧を検証
func TestTransferRequest_ListPending(t *testing.T) {
	tr, db := setupTransferRequest(t)
	ctx := context.Background()

	alice := createTestUserWithBalance(t, db, "alice_treq_list", 1000)
	bob := createTestUser(t, db, "bob_treq_list")

	// 2件のリクエスト作成
	for i := 0; i < 2; i++ {
		_, err := tr.CreateTransferRequest(ctx, &inputport.CreateTransferRequestRequest{
			FromUserID:     alice.ID,
			ToUserID:       bob.ID,
			Amount:         int64(100 + i*50),
			Message:        "pending test",
			IdempotencyKey: fmt.Sprintf("integ-treq-list-%d", i),
		})
		require.NoError(t, err)
	}

	// 保留リクエスト一覧
	pendingResp, err := tr.GetPendingRequests(ctx, &inputport.GetPendingTransferRequestsRequest{
		ToUserID: bob.ID,
		Offset:   0,
		Limit:    10,
	})
	require.NoError(t, err)
	assert.Len(t, pendingResp.Requests, 2)
}
