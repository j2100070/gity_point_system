package interactor_test

import (
	"errors"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// Mock Repositories
// ========================================

type mockTransferRequestRepo struct {
	requests       map[uuid.UUID]*entities.TransferRequest
	byIdempotency  map[string]*entities.TransferRequest
	pendingByTo    []*entities.TransferRequest
	sentByFrom     []*entities.TransferRequest
	pendingCount   int64
	createErr      error
	readErr        error
	updateErr      error
	countErr       error
}

func newMockTransferRequestRepo() *mockTransferRequestRepo {
	return &mockTransferRequestRepo{
		requests:      make(map[uuid.UUID]*entities.TransferRequest),
		byIdempotency: make(map[string]*entities.TransferRequest),
	}
}

func (m *mockTransferRequestRepo) Create(tr *entities.TransferRequest) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.requests[tr.ID] = tr
	m.byIdempotency[tr.IdempotencyKey] = tr
	return nil
}

func (m *mockTransferRequestRepo) Read(id uuid.UUID) (*entities.TransferRequest, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	tr, ok := m.requests[id]
	if !ok {
		return nil, errors.New("transfer request not found")
	}
	return tr, nil
}

func (m *mockTransferRequestRepo) ReadByIdempotencyKey(key string) (*entities.TransferRequest, error) {
	tr, ok := m.byIdempotency[key]
	if !ok {
		return nil, nil
	}
	return tr, nil
}

func (m *mockTransferRequestRepo) Update(tr *entities.TransferRequest) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.requests[tr.ID] = tr
	return nil
}

func (m *mockTransferRequestRepo) ReadPendingByToUser(toUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error) {
	return m.pendingByTo, nil
}

func (m *mockTransferRequestRepo) ReadSentByFromUser(fromUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error) {
	return m.sentByFrom, nil
}

func (m *mockTransferRequestRepo) CountPendingByToUser(toUserID uuid.UUID) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return m.pendingCount, nil
}

func (m *mockTransferRequestRepo) UpdateExpiredRequests() (int64, error) {
	return 0, nil
}

type mockUserRepoForTR struct {
	users   map[uuid.UUID]*entities.User
	readErr error
}

func newMockUserRepoForTR() *mockUserRepoForTR {
	return &mockUserRepoForTR{
		users: make(map[uuid.UUID]*entities.User),
	}
}

func (m *mockUserRepoForTR) Create(user *entities.User) error { return nil }
func (m *mockUserRepoForTR) Read(id uuid.UUID) (*entities.User, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}
func (m *mockUserRepoForTR) ReadByUsername(username string) (*entities.User, error) { return nil, nil }
func (m *mockUserRepoForTR) ReadByEmail(email string) (*entities.User, error)       { return nil, nil }
func (m *mockUserRepoForTR) Update(user *entities.User) (bool, error)               { return true, nil }
func (m *mockUserRepoForTR) ReadPersonalQRCode(userID uuid.UUID) (string, error)    { return "", nil }
func (m *mockUserRepoForTR) Count() (int64, error)                                  { return 0, nil }
func (m *mockUserRepoForTR) Delete(id uuid.UUID) error                              { return nil }

func (m *mockUserRepoForTR) setUser(user *entities.User) {
	m.users[user.ID] = user
}

type mockPointTransferPort struct {
	transferResp *inputport.TransferResponse
	transferErr  error
}

func newMockPointTransferPort() *mockPointTransferPort {
	return &mockPointTransferPort{}
}

func (m *mockPointTransferPort) Transfer(req *inputport.TransferRequest) (*inputport.TransferResponse, error) {
	if m.transferErr != nil {
		return nil, m.transferErr
	}
	return m.transferResp, nil
}

func (m *mockPointTransferPort) GetBalance(req *inputport.GetBalanceRequest) (*inputport.GetBalanceResponse, error) {
	return &inputport.GetBalanceResponse{Balance: 0}, nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...entities.Field) {}
func (m *mockLogger) Info(msg string, fields ...entities.Field)  {}
func (m *mockLogger) Warn(msg string, fields ...entities.Field)  {}
func (m *mockLogger) Error(msg string, fields ...entities.Field) {}
func (m *mockLogger) Fatal(msg string, fields ...entities.Field) {}

// ========================================
// TransferRequest Interactor Tests
// ========================================

func TestTransferRequestInteractor_CreateTransferRequest(t *testing.T) {
	t.Run("正常に送金リクエストを作成", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender, _ := entities.NewUser("sender", "sender@example.com", "hash", "Sender")
		sender.Balance = 10000
		sender.IsActive = true
		receiver, _ := entities.NewUser("receiver", "receiver@example.com", "hash", "Receiver")
		receiver.IsActive = true

		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.CreateTransferRequestRequest{
			FromUserID:     sender.ID,
			ToUserID:       receiver.ID,
			Amount:         1000,
			Message:        "Test transfer",
			IdempotencyKey: "key-123",
		}

		resp, err := interactor.CreateTransferRequest(req)
		require.NoError(t, err)
		assert.Equal(t, sender.ID, resp.TransferRequest.FromUserID)
		assert.Equal(t, receiver.ID, resp.TransferRequest.ToUserID)
		assert.Equal(t, int64(1000), resp.TransferRequest.Amount)
		assert.Equal(t, "Test transfer", resp.TransferRequest.Message)
		assert.Equal(t, entities.TransferRequestStatusPending, resp.TransferRequest.Status)
	})

	t.Run("冪等性キーで既存リクエストを返す", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender, _ := entities.NewUser("sender", "sender@example.com", "hash", "Sender")
		sender.Balance = 10000
		sender.IsActive = true
		receiver, _ := entities.NewUser("receiver", "receiver@example.com", "hash", "Receiver")
		receiver.IsActive = true

		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		// 既存リクエストを作成
		existingTR, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Existing", "key-existing")
		trRepo.Create(existingTR)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.CreateTransferRequestRequest{
			FromUserID:     sender.ID,
			ToUserID:       receiver.ID,
			Amount:         1000,
			Message:        "New",
			IdempotencyKey: "key-existing", // 同じキー
		}

		resp, err := interactor.CreateTransferRequest(req)
		require.NoError(t, err)
		assert.Equal(t, existingTR.ID, resp.TransferRequest.ID)
	})

	t.Run("送信者が存在しない場合エラー", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		receiver, _ := entities.NewUser("receiver", "receiver@example.com", "hash", "Receiver")
		receiver.IsActive = true
		userRepo.setUser(receiver)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.CreateTransferRequestRequest{
			FromUserID:     uuid.New(), // 存在しないユーザー
			ToUserID:       receiver.ID,
			Amount:         1000,
			Message:        "Test",
			IdempotencyKey: "key-nosender",
		}

		_, err := interactor.CreateTransferRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sender not found")
	})

	t.Run("残高不足の場合エラー", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender, _ := entities.NewUser("sender", "sender@example.com", "hash", "Sender")
		sender.Balance = 100 // 残高不足
		sender.IsActive = true
		receiver, _ := entities.NewUser("receiver", "receiver@example.com", "hash", "Receiver")
		receiver.IsActive = true

		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.CreateTransferRequestRequest{
			FromUserID:     sender.ID,
			ToUserID:       receiver.ID,
			Amount:         1000,
			Message:        "Test",
			IdempotencyKey: "key-insufficient",
		}

		_, err := interactor.CreateTransferRequest(req)
		assert.Error(t, err)
	})
}

func TestTransferRequestInteractor_ApproveTransferRequest(t *testing.T) {
	t.Run("正常に送金リクエストを承認", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender, _ := entities.NewUser("sender", "sender@example.com", "hash", "Sender")
		sender.Balance = 10000
		receiver, _ := entities.NewUser("receiver", "receiver@example.com", "hash", "Receiver")
		receiver.Balance = 5000

		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		// リクエスト作成
		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-approve")
		trRepo.Create(tr)

		// ポイント転送のモックレスポンス
		transaction := &entities.Transaction{
			ID:          uuid.New(),
			FromUserID:  sender.ID,
			ToUserID:    receiver.ID,
			Amount:      1000,
			Description: "approval",
		}
		ptPort.transferResp = &inputport.TransferResponse{
			Transaction: transaction,
			FromUser:    sender,
			ToUser:      receiver,
		}

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.ApproveTransferRequestRequest{
			RequestID: tr.ID,
			UserID:    receiver.ID, // 受取人が承認
		}

		resp, err := interactor.ApproveTransferRequest(req)
		require.NoError(t, err)
		assert.Equal(t, entities.TransferRequestStatusApproved, resp.TransferRequest.Status)
		assert.NotNil(t, resp.TransferRequest.ApprovedAt)
		assert.NotNil(t, resp.TransferRequest.TransactionID)
		assert.Equal(t, transaction.ID, *resp.TransferRequest.TransactionID)
	})

	t.Run("送信者が承認しようとするとエラー", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender := &entities.User{ID: uuid.New()}
		receiver := &entities.User{ID: uuid.New()}

		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-wronguser")
		trRepo.Create(tr)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.ApproveTransferRequestRequest{
			RequestID: tr.ID,
			UserID:    sender.ID, // 送信者が承認（エラー）
		}

		_, err := interactor.ApproveTransferRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("期限切れリクエストは承認できない", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender := &entities.User{ID: uuid.New()}
		receiver := &entities.User{ID: uuid.New()}

		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-expired")
		tr.ExpiresAt = time.Now().Add(-1 * time.Hour) // 期限切れ
		trRepo.Create(tr)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.ApproveTransferRequestRequest{
			RequestID: tr.ID,
			UserID:    receiver.ID,
		}

		_, err := interactor.ApproveTransferRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("ポイント転送が失敗した場合エラー", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender := &entities.User{ID: uuid.New()}
		receiver := &entities.User{ID: uuid.New()}

		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-transfer-fail")
		trRepo.Create(tr)

		// ポイント転送を失敗させる
		ptPort.transferErr = errors.New("insufficient balance")

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.ApproveTransferRequestRequest{
			RequestID: tr.ID,
			UserID:    receiver.ID,
		}

		_, err := interactor.ApproveTransferRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute transfer")
	})
}

func TestTransferRequestInteractor_RejectTransferRequest(t *testing.T) {
	t.Run("正常に送金リクエストを拒否", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender := &entities.User{ID: uuid.New()}
		receiver := &entities.User{ID: uuid.New()}

		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-reject")
		trRepo.Create(tr)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.RejectTransferRequestRequest{
			RequestID: tr.ID,
			UserID:    receiver.ID,
		}

		resp, err := interactor.RejectTransferRequest(req)
		require.NoError(t, err)
		assert.Equal(t, entities.TransferRequestStatusRejected, resp.TransferRequest.Status)
		assert.NotNil(t, resp.TransferRequest.RejectedAt)
	})

	t.Run("送信者が拒否しようとするとエラー", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender := &entities.User{ID: uuid.New()}
		receiver := &entities.User{ID: uuid.New()}

		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-reject-wrong")
		trRepo.Create(tr)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.RejectTransferRequestRequest{
			RequestID: tr.ID,
			UserID:    sender.ID, // 送信者が拒否（エラー）
		}

		_, err := interactor.RejectTransferRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

func TestTransferRequestInteractor_CancelTransferRequest(t *testing.T) {
	t.Run("正常に送金リクエストをキャンセル", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender := &entities.User{ID: uuid.New()}
		receiver := &entities.User{ID: uuid.New()}

		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-cancel")
		trRepo.Create(tr)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.CancelTransferRequestRequest{
			RequestID: tr.ID,
			UserID:    sender.ID,
		}

		resp, err := interactor.CancelTransferRequest(req)
		require.NoError(t, err)
		assert.Equal(t, entities.TransferRequestStatusCancelled, resp.TransferRequest.Status)
		assert.NotNil(t, resp.TransferRequest.CancelledAt)
	})

	t.Run("受取人がキャンセルしようとするとエラー", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender := &entities.User{ID: uuid.New()}
		receiver := &entities.User{ID: uuid.New()}

		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-cancel-wrong")
		trRepo.Create(tr)

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.CancelTransferRequestRequest{
			RequestID: tr.ID,
			UserID:    receiver.ID, // 受取人がキャンセル（エラー）
		}

		_, err := interactor.CancelTransferRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

func TestTransferRequestInteractor_GetPendingRequests(t *testing.T) {
	t.Run("承認待ちリクエスト一覧を取得", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender, _ := entities.NewUser("sender", "sender@example.com", "hash", "Sender")
		receiver, _ := entities.NewUser("receiver", "receiver@example.com", "hash", "Receiver")

		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-pending")
		trRepo.pendingByTo = []*entities.TransferRequest{tr}

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.GetPendingTransferRequestsRequest{
			ToUserID: receiver.ID,
			Offset:   0,
			Limit:    10,
		}

		resp, err := interactor.GetPendingRequests(req)
		require.NoError(t, err)
		assert.Len(t, resp.Requests, 1)
		assert.Equal(t, tr.ID, resp.Requests[0].TransferRequest.ID)
	})

	t.Run("期限切れリクエストは除外される", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		sender, _ := entities.NewUser("sender", "sender@example.com", "hash", "Sender")
		receiver, _ := entities.NewUser("receiver", "receiver@example.com", "hash", "Receiver")

		userRepo.setUser(sender)
		userRepo.setUser(receiver)

		tr, _ := entities.NewTransferRequest(sender.ID, receiver.ID, 1000, "Test", "key-expired-list")
		tr.ExpiresAt = time.Now().Add(-1 * time.Hour) // 期限切れ
		trRepo.pendingByTo = []*entities.TransferRequest{tr}

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.GetPendingTransferRequestsRequest{
			ToUserID: receiver.ID,
			Offset:   0,
			Limit:    10,
		}

		resp, err := interactor.GetPendingRequests(req)
		require.NoError(t, err)
		assert.Len(t, resp.Requests, 0) // 期限切れは除外
	})
}

func TestTransferRequestInteractor_GetPendingRequestCount(t *testing.T) {
	t.Run("承認待ちリクエスト数を取得", func(t *testing.T) {
		trRepo := newMockTransferRequestRepo()
		userRepo := newMockUserRepoForTR()
		ptPort := newMockPointTransferPort()
		logger := &mockLogger{}

		trRepo.pendingCount = 5

		interactor := interactor.NewTransferRequestInteractor(trRepo, userRepo, ptPort, logger)

		req := &inputport.GetPendingRequestCountRequest{
			ToUserID: uuid.New(),
		}

		resp, err := interactor.GetPendingRequestCount(req)
		require.NoError(t, err)
		assert.Equal(t, int64(5), resp.Count)
	})
}
