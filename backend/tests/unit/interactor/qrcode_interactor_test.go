package interactor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// QRCodeInteractor テスト
// ========================================

// --- Mock QRCodeRepository ---

type mockQRCodeRepo struct {
	qrCodes map[uuid.UUID]*entities.QRCode
	codeMap map[string]*entities.QRCode
}

func newMockQRCodeRepo() *mockQRCodeRepo {
	return &mockQRCodeRepo{
		qrCodes: make(map[uuid.UUID]*entities.QRCode),
		codeMap: make(map[string]*entities.QRCode),
	}
}

func (m *mockQRCodeRepo) Create(ctx context.Context, qrCode *entities.QRCode) error {
	m.qrCodes[qrCode.ID] = qrCode
	m.codeMap[qrCode.Code] = qrCode
	return nil
}
func (m *mockQRCodeRepo) ReadByCode(ctx context.Context, code string) (*entities.QRCode, error) {
	qr, ok := m.codeMap[code]
	if !ok {
		return nil, errors.New("qr code not found")
	}
	return qr, nil
}
func (m *mockQRCodeRepo) Read(ctx context.Context, id uuid.UUID) (*entities.QRCode, error) {
	qr, ok := m.qrCodes[id]
	if !ok {
		return nil, errors.New("qr code not found")
	}
	return qr, nil
}
func (m *mockQRCodeRepo) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.QRCode, error) {
	result := make([]*entities.QRCode, 0)
	for _, qr := range m.qrCodes {
		if qr.UserID == userID {
			result = append(result, qr)
		}
	}
	return result, nil
}
func (m *mockQRCodeRepo) Update(ctx context.Context, qrCode *entities.QRCode) error {
	m.qrCodes[qrCode.ID] = qrCode
	m.codeMap[qrCode.Code] = qrCode
	return nil
}
func (m *mockQRCodeRepo) DeleteExpired(ctx context.Context) error { return nil }

// --- Mock PointTransferInputPort (for QRCode) ---

type mockPointTransferUC struct {
	transferResp *inputport.TransferResponse
	transferErr  error
}

func (m *mockPointTransferUC) Transfer(ctx context.Context, req *inputport.TransferRequest) (*inputport.TransferResponse, error) {
	if m.transferErr != nil {
		return nil, m.transferErr
	}
	if m.transferResp != nil {
		return m.transferResp, nil
	}
	// デフォルト: 正常なレスポンスを生成
	tx, _ := entities.NewAdminDeduct(req.FromUserID, req.Amount, "QR transfer", uuid.Nil)
	fromUser, _ := entities.NewUser("from", "from@example.com", "hash", "From", "太", "田")
	toUser, _ := entities.NewUser("to", "to@example.com", "hash", "To", "次", "山")
	return &inputport.TransferResponse{
		Transaction: tx,
		FromUser:    fromUser,
		ToUser:      toUser,
	}, nil
}
func (m *mockPointTransferUC) GetBalance(ctx context.Context, req *inputport.GetBalanceRequest) (*inputport.GetBalanceResponse, error) {
	return nil, nil
}
func (m *mockPointTransferUC) GetTransactionHistory(ctx context.Context, req *inputport.GetTransactionHistoryRequest) (*inputport.GetTransactionHistoryResponse, error) {
	return nil, nil
}
func (m *mockPointTransferUC) GetExpiringPoints(ctx context.Context, req *inputport.GetExpiringPointsRequest) (*inputport.GetExpiringPointsResponse, error) {
	return nil, nil
}

// --- GenerateReceiveQR ---

func TestQRCodeInteractor_GenerateReceiveQR(t *testing.T) {
	setup := func() (*mockQRCodeRepo, inputport.QRCodeInputPort) {
		qrRepo := newMockQRCodeRepo()
		sut := interactor.NewQRCodeInteractor(qrRepo, &mockPointTransferUC{}, &mockLogger{})
		return qrRepo, sut
	}

	t.Run("金額指定なしでQRコードを生成できる", func(t *testing.T) {
		_, sut := setup()

		resp, err := sut.GenerateReceiveQR(context.Background(), &inputport.GenerateReceiveQRRequest{
			UserID: uuid.New(), Amount: nil,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.QRCode)
		assert.Contains(t, resp.QRCodeData, "receive:")
		assert.Nil(t, resp.QRCode.Amount)
	})

	t.Run("金額指定ありでQRコードを生成できる", func(t *testing.T) {
		_, sut := setup()
		amount := int64(500)

		resp, err := sut.GenerateReceiveQR(context.Background(), &inputport.GenerateReceiveQRRequest{
			UserID: uuid.New(), Amount: &amount,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.QRCode)
		assert.Contains(t, resp.QRCodeData, "receive:")
		assert.NotNil(t, resp.QRCode.Amount)
		assert.Equal(t, int64(500), *resp.QRCode.Amount)
	})

	t.Run("金額が0以下の場合エラー", func(t *testing.T) {
		_, sut := setup()
		amount := int64(0)

		_, err := sut.GenerateReceiveQR(context.Background(), &inputport.GenerateReceiveQRRequest{
			UserID: uuid.New(), Amount: &amount,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})
}

// --- GenerateSendQR ---

func TestQRCodeInteractor_GenerateSendQR(t *testing.T) {
	setup := func() (*mockQRCodeRepo, inputport.QRCodeInputPort) {
		qrRepo := newMockQRCodeRepo()
		sut := interactor.NewQRCodeInteractor(qrRepo, &mockPointTransferUC{}, &mockLogger{})
		return qrRepo, sut
	}

	t.Run("正常に送信用QRコードを生成できる", func(t *testing.T) {
		_, sut := setup()

		resp, err := sut.GenerateSendQR(context.Background(), &inputport.GenerateSendQRRequest{
			UserID: uuid.New(), Amount: 1000,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.QRCode)
		assert.Contains(t, resp.QRCodeData, "send:")
		assert.Equal(t, int64(1000), *resp.QRCode.Amount)
	})

	t.Run("金額が0以下の場合エラー", func(t *testing.T) {
		_, sut := setup()

		_, err := sut.GenerateSendQR(context.Background(), &inputport.GenerateSendQRRequest{
			UserID: uuid.New(), Amount: 0,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})
}

// --- ScanQR ---

func TestQRCodeInteractor_ScanQR(t *testing.T) {
	setup := func() (*mockQRCodeRepo, *mockPointTransferUC, inputport.QRCodeInputPort) {
		qrRepo := newMockQRCodeRepo()
		transferUC := &mockPointTransferUC{}
		sut := interactor.NewQRCodeInteractor(qrRepo, transferUC, &mockLogger{})
		return qrRepo, transferUC, sut
	}

	t.Run("受取用QRコードをスキャンしてポイント転送できる", func(t *testing.T) {
		qrRepo, _, sut := setup()
		ownerID := uuid.New()
		scannerID := uuid.New()
		amount := int64(500)

		qrCode, _ := entities.NewReceiveQRCode(ownerID, &amount)
		qrRepo.codeMap[qrCode.Code] = qrCode
		qrRepo.qrCodes[qrCode.ID] = qrCode

		resp, err := sut.ScanQR(context.Background(), &inputport.ScanQRRequest{
			UserID: scannerID, Code: qrCode.Code, IdempotencyKey: "scan-" + uuid.New().String(),
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Transaction)
	})

	t.Run("存在しないQRコードの場合エラー", func(t *testing.T) {
		_, _, sut := setup()

		_, err := sut.ScanQR(context.Background(), &inputport.ScanQRRequest{
			UserID: uuid.New(), Code: "nonexistent", IdempotencyKey: "key",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "qr code not found")
	})

	t.Run("自分のQRコードはスキャンできない", func(t *testing.T) {
		qrRepo, _, sut := setup()
		userID := uuid.New()
		amount := int64(500)

		qrCode, _ := entities.NewReceiveQRCode(userID, &amount)
		qrRepo.codeMap[qrCode.Code] = qrCode
		qrRepo.qrCodes[qrCode.ID] = qrCode

		_, err := sut.ScanQR(context.Background(), &inputport.ScanQRRequest{
			UserID: userID, Code: qrCode.Code, IdempotencyKey: "key",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use your own qr code")
	})

	t.Run("金額未指定のQRコードで金額なしの場合エラー", func(t *testing.T) {
		qrRepo, _, sut := setup()
		ownerID := uuid.New()
		scannerID := uuid.New()

		qrCode, _ := entities.NewReceiveQRCode(ownerID, nil) // 金額なし
		qrRepo.codeMap[qrCode.Code] = qrCode
		qrRepo.qrCodes[qrCode.ID] = qrCode

		_, err := sut.ScanQR(context.Background(), &inputport.ScanQRRequest{
			UserID: scannerID, Code: qrCode.Code, Amount: nil, IdempotencyKey: "key",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount is required")
	})
}

// --- GetQRCodeHistory ---

func TestQRCodeInteractor_GetQRCodeHistory(t *testing.T) {
	t.Run("正常にQRコード履歴を取得できる", func(t *testing.T) {
		qrRepo := newMockQRCodeRepo()
		sut := interactor.NewQRCodeInteractor(qrRepo, &mockPointTransferUC{}, &mockLogger{})

		userID := uuid.New()
		qr1, _ := entities.NewReceiveQRCode(userID, nil)
		qr2, _ := entities.NewSendQRCode(userID, 100)
		qrRepo.qrCodes[qr1.ID] = qr1
		qrRepo.qrCodes[qr2.ID] = qr2

		resp, err := sut.GetQRCodeHistory(context.Background(), &inputport.GetQRCodeHistoryRequest{
			UserID: userID, Offset: 0, Limit: 20,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, len(resp.QRCodes))
	})
}
