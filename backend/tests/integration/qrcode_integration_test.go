//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	infrapostgres "github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupQRCode(t *testing.T) (inputport.QRCodeInputPort, infrapostgres.DB) {
	t.Helper()
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	txManager := infrapostgres.NewGormTransactionManager(db.GetDB())

	pt := interactor.NewPointTransferInteractor(
		txManager, repos.User, repos.Transaction, repos.IdempotencyKey, repos.Friendship, repos.PointBatch, lg,
	)
	qr := interactor.NewQRCodeInteractor(repos.QRCode, pt, lg)
	return qr, db
}

// TestQRCode_GenerateReceiveQR は受取用QRコード生成を検証
func TestQRCode_GenerateReceiveQR(t *testing.T) {
	qr, db := setupQRCode(t)
	ctx := context.Background()

	user := createTestUser(t, db, "qr_receive_user")

	resp, err := qr.GenerateReceiveQR(ctx, &inputport.GenerateReceiveQRRequest{
		UserID: user.ID,
		Amount: nil, // 金額指定なし
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, resp.QRCode)
	assert.NotEmpty(t, resp.QRCodeData)
}

// TestQRCode_GenerateReceiveQRWithAmount は金額指定付き受取用QRを検証
func TestQRCode_GenerateReceiveQRWithAmount(t *testing.T) {
	qr, db := setupQRCode(t)
	ctx := context.Background()

	user := createTestUser(t, db, "qr_amount_user")
	amount := int64(500)

	resp, err := qr.GenerateReceiveQR(ctx, &inputport.GenerateReceiveQRRequest{
		UserID: user.ID,
		Amount: &amount,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, resp.QRCode)
}

// TestQRCode_GenerateSendQR は送信用QRコード生成を検証
func TestQRCode_GenerateSendQR(t *testing.T) {
	qr, db := setupQRCode(t)
	ctx := context.Background()

	user := createTestUserWithBalance(t, db, "qr_send_user", 1000)

	resp, err := qr.GenerateSendQR(ctx, &inputport.GenerateSendQRRequest{
		UserID: user.ID,
		Amount: 300,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, resp.QRCode)
	assert.NotEmpty(t, resp.QRCodeData)
}

// TestQRCode_ScanAndTransfer はQRスキャン→即時送金を検証
func TestQRCode_ScanAndTransfer(t *testing.T) {
	qr, db := setupQRCode(t)
	ctx := context.Background()

	sender := createTestUserWithBalance(t, db, "qr_scan_sender", 1000)
	receiver := createTestUser(t, db, "qr_scan_receiver")

	// フレンド関係
	db.GetDB().Exec(
		`INSERT INTO friendships (id, requester_id, addressee_id, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), ?, ?, 'accepted', NOW(), NOW())`,
		sender.ID, receiver.ID,
	)

	// 受取QR生成
	genResp, err := qr.GenerateReceiveQR(ctx, &inputport.GenerateReceiveQRRequest{
		UserID: receiver.ID,
	})
	require.NoError(t, err)

	// スキャン（送金）
	amount := int64(200)
	scanResp, err := qr.ScanQR(ctx, &inputport.ScanQRRequest{
		UserID:         sender.ID,
		Code:           genResp.QRCode.Code,
		Amount:         &amount,
		IdempotencyKey: "integ-qr-scan-001",
	})
	require.NoError(t, err)
	require.NotNil(t, scanResp)
	assert.Equal(t, int64(200), scanResp.Transaction.Amount)
}

// TestQRCode_GetHistory はQRコード履歴取得を検証
func TestQRCode_GetHistory(t *testing.T) {
	qr, db := setupQRCode(t)
	ctx := context.Background()

	user := createTestUser(t, db, "qr_history_user")

	// QRコードを生成
	_, err := qr.GenerateReceiveQR(ctx, &inputport.GenerateReceiveQRRequest{
		UserID: user.ID,
	})
	require.NoError(t, err)

	// 履歴取得
	histResp, err := qr.GetQRCodeHistory(ctx, &inputport.GetQRCodeHistoryRequest{
		UserID: user.ID,
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(histResp.QRCodes), 1)
}
