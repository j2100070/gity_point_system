package interactor_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// テスト用共通インフラ
// （同一パッケージ内で point_transfer_interactor_test.go からも使用）
// ========================================

// --- Context-Tracking TransactionManager ---

// txCtxMarkerKey はテスト用のコンテキスト識別キー
type txCtxMarkerKey string

const testTxMarker txCtxMarkerKey = "test_tx_marker"

// ctxTrackingTxManager はトランザクションコンテキストを追跡するモック。
// 実際の GormTransactionManager と同様に、新しいコンテキストを作成して fn に渡す。
type ctxTrackingTxManager struct {
	TxCtx context.Context
}

func (m *ctxTrackingTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	m.TxCtx = context.WithValue(ctx, testTxMarker, true)
	return fn(m.TxCtx)
}

// isTxContext はコンテキストがトランザクションコンテキストかどうかを判定
func isTxContext(ctx context.Context) bool {
	val, ok := ctx.Value(testTxMarker).(bool)
	return ok && val
}

// --- Context-Tracking UserRepository ---

type ctxTrackingUserRepo struct {
	ctxRecords map[string]context.Context
	users      map[uuid.UUID]*entities.User
	readErr    error
	updateOK   bool
}

func newCtxTrackingUserRepo() *ctxTrackingUserRepo {
	return &ctxTrackingUserRepo{
		ctxRecords: make(map[string]context.Context),
		users:      make(map[uuid.UUID]*entities.User),
		updateOK:   true,
	}
}

func (m *ctxTrackingUserRepo) setUser(u *entities.User) { m.users[u.ID] = u }

func (m *ctxTrackingUserRepo) Create(ctx context.Context, user *entities.User) error { return nil }
func (m *ctxTrackingUserRepo) Read(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	m.ctxRecords["Read_"+id.String()] = ctx
	if m.readErr != nil {
		return nil, m.readErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	// コピーを返して副作用を防止
	copy := *u
	return &copy, nil
}
func (m *ctxTrackingUserRepo) ReadByUsername(ctx context.Context, username string) (*entities.User, error) {
	return nil, nil
}
func (m *ctxTrackingUserRepo) ReadByEmail(ctx context.Context, email string) (*entities.User, error) {
	return nil, nil
}
func (m *ctxTrackingUserRepo) Update(ctx context.Context, user *entities.User) (bool, error) {
	m.ctxRecords["Update"] = ctx
	return m.updateOK, nil
}
func (m *ctxTrackingUserRepo) UpdateBalanceWithLock(ctx context.Context, userID uuid.UUID, amount int64, isDeduct bool) error {
	m.ctxRecords["UpdateBalanceWithLock"] = ctx
	return nil
}
func (m *ctxTrackingUserRepo) UpdateBalancesWithLock(ctx context.Context, updates []repository.BalanceUpdate) error {
	m.ctxRecords["UpdateBalancesWithLock"] = ctx
	return nil
}
func (m *ctxTrackingUserRepo) ReadList(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	m.ctxRecords["ReadList"] = ctx
	result := make([]*entities.User, 0)
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}
func (m *ctxTrackingUserRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}
func (m *ctxTrackingUserRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *ctxTrackingUserRepo) ReadListWithSearch(ctx context.Context, search, sortBy, sortOrder string, offset, limit int) ([]*entities.User, error) {
	m.ctxRecords["ReadListWithSearch"] = ctx
	return []*entities.User{}, nil
}
func (m *ctxTrackingUserRepo) CountWithSearch(ctx context.Context, search string) (int64, error) {
	return 0, nil
}
func (m *ctxTrackingUserRepo) ReadPersonalQRCode(ctx context.Context, userID uuid.UUID) (string, error) {
	return "", nil
}

// --- Context-Tracking TransactionRepository ---

type ctxTrackingTransactionRepo struct {
	ctxRecords   map[string]context.Context
	transactions []*entities.Transaction
}

func newCtxTrackingTransactionRepo() *ctxTrackingTransactionRepo {
	return &ctxTrackingTransactionRepo{
		ctxRecords: make(map[string]context.Context),
	}
}

func (m *ctxTrackingTransactionRepo) Create(ctx context.Context, tx *entities.Transaction) error {
	m.ctxRecords["Create"] = ctx
	m.transactions = append(m.transactions, tx)
	return nil
}
func (m *ctxTrackingTransactionRepo) Read(ctx context.Context, id uuid.UUID) (*entities.Transaction, error) {
	m.ctxRecords["Read"] = ctx
	for _, tx := range m.transactions {
		if tx.ID == id {
			return tx, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *ctxTrackingTransactionRepo) ReadByIdempotencyKey(ctx context.Context, key string) (*entities.Transaction, error) {
	return nil, errors.New("not found")
}
func (m *ctxTrackingTransactionRepo) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Transaction, error) {
	m.ctxRecords["ReadListByUserID"] = ctx
	return m.transactions, nil
}
func (m *ctxTrackingTransactionRepo) ReadListAll(ctx context.Context, offset, limit int) ([]*entities.Transaction, error) {
	m.ctxRecords["ReadListAll"] = ctx
	return m.transactions, nil
}
func (m *ctxTrackingTransactionRepo) Update(ctx context.Context, tx *entities.Transaction) error {
	m.ctxRecords["Update"] = ctx
	return nil
}
func (m *ctxTrackingTransactionRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return int64(len(m.transactions)), nil
}
func (m *ctxTrackingTransactionRepo) ReadListAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo, sortBy, sortOrder string, offset, limit int) ([]*entities.Transaction, error) {
	m.ctxRecords["ReadListAllWithFilter"] = ctx
	return m.transactions, nil
}
func (m *ctxTrackingTransactionRepo) CountAll(ctx context.Context) (int64, error) {
	return int64(len(m.transactions)), nil
}
func (m *ctxTrackingTransactionRepo) CountAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo string) (int64, error) {
	return int64(len(m.transactions)), nil
}
func (m *ctxTrackingTransactionRepo) ReadListByUserIDWithUsers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.TransactionWithUsers, error) {
	return nil, nil
}
func (m *ctxTrackingTransactionRepo) ReadListAllWithFilterAndUsers(ctx context.Context, transactionType, dateFrom, dateTo, sortBy, sortOrder string, offset, limit int) ([]*entities.TransactionWithUsers, error) {
	return nil, nil
}

// --- Context-Tracking IdempotencyKeyRepository ---

type ctxTrackingIdempotencyRepo struct {
	ctxRecords map[string]context.Context
	keys       map[string]*entities.IdempotencyKey
}

func newCtxTrackingIdempotencyRepo() *ctxTrackingIdempotencyRepo {
	return &ctxTrackingIdempotencyRepo{
		ctxRecords: make(map[string]context.Context),
		keys:       make(map[string]*entities.IdempotencyKey),
	}
}

func (m *ctxTrackingIdempotencyRepo) Create(ctx context.Context, key *entities.IdempotencyKey) error {
	m.ctxRecords["Create"] = ctx
	m.keys[key.Key] = key
	return nil
}
func (m *ctxTrackingIdempotencyRepo) ReadByKey(ctx context.Context, key string) (*entities.IdempotencyKey, error) {
	m.ctxRecords["ReadByKey"] = ctx
	k, ok := m.keys[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return k, nil
}
func (m *ctxTrackingIdempotencyRepo) Update(ctx context.Context, key *entities.IdempotencyKey) error {
	m.ctxRecords["Update"] = ctx
	m.keys[key.Key] = key
	return nil
}
func (m *ctxTrackingIdempotencyRepo) DeleteExpired(ctx context.Context) error { return nil }

// --- Context-Tracking PointBatchRepository ---

type ctxTrackingPointBatchRepo struct {
	ctxRecords map[string]context.Context
}

func newCtxTrackingPointBatchRepo() *ctxTrackingPointBatchRepo {
	return &ctxTrackingPointBatchRepo{
		ctxRecords: make(map[string]context.Context),
	}
}

func (m *ctxTrackingPointBatchRepo) Create(ctx context.Context, batch *entities.PointBatch) error {
	m.ctxRecords["Create"] = ctx
	return nil
}
func (m *ctxTrackingPointBatchRepo) ConsumePointsFIFO(ctx context.Context, userID uuid.UUID, amount int64) error {
	m.ctxRecords["ConsumePointsFIFO"] = ctx
	return nil
}
func (m *ctxTrackingPointBatchRepo) FindExpiredBatches(ctx context.Context, before time.Time, limit int) ([]*entities.PointBatch, error) {
	return nil, nil
}
func (m *ctxTrackingPointBatchRepo) MarkExpired(ctx context.Context, batchID uuid.UUID) error {
	return nil
}

// --- Context-Tracking FriendshipRepository ---

type ctxTrackingFriendshipRepo struct {
	areFriends bool
}

func newCtxTrackingFriendshipRepo() *ctxTrackingFriendshipRepo {
	return &ctxTrackingFriendshipRepo{areFriends: true}
}

func (m *ctxTrackingFriendshipRepo) Create(ctx context.Context, f *entities.Friendship) error {
	return nil
}
func (m *ctxTrackingFriendshipRepo) Read(ctx context.Context, id uuid.UUID) (*entities.Friendship, error) {
	return nil, nil
}
func (m *ctxTrackingFriendshipRepo) ReadByUsers(ctx context.Context, u1, u2 uuid.UUID) (*entities.Friendship, error) {
	return nil, nil
}
func (m *ctxTrackingFriendshipRepo) ReadListFriends(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	return nil, nil
}
func (m *ctxTrackingFriendshipRepo) ReadListPendingRequests(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	return nil, nil
}
func (m *ctxTrackingFriendshipRepo) Update(ctx context.Context, f *entities.Friendship) error {
	return nil
}
func (m *ctxTrackingFriendshipRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *ctxTrackingFriendshipRepo) ArchiveAndDelete(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error {
	return nil
}
func (m *ctxTrackingFriendshipRepo) CheckAreFriends(ctx context.Context, u1, u2 uuid.UUID) (bool, error) {
	return m.areFriends, nil
}
func (m *ctxTrackingFriendshipRepo) ReadListFriendsWithUsers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.FriendshipWithUser, error) {
	return nil, nil
}
func (m *ctxTrackingFriendshipRepo) ReadListPendingRequestsWithUsers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.FriendshipWithUser, error) {
	return nil, nil
}
func (m *ctxTrackingFriendshipRepo) CountPendingRequests(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}

// --- Mock AnalyticsDataSource ---

type mockAnalyticsDS struct{}

func (m *mockAnalyticsDS) GetUserBalanceSummary(ctx context.Context) (*entities.AnalyticsSummaryResult, error) {
	return &entities.AnalyticsSummaryResult{TotalBalance: 100000, AverageBalance: 5000, ActiveUsers: 20}, nil
}
func (m *mockAnalyticsDS) GetTopHolders(ctx context.Context, limit int) ([]*entities.TopHolderResult, error) {
	return []*entities.TopHolderResult{
		{ID: uuid.New().String(), Username: "top1", DisplayName: "Top 1", Balance: 50000},
	}, nil
}
func (m *mockAnalyticsDS) GetDailyStats(ctx context.Context, since time.Time) ([]*entities.DailyStatResult, error) {
	return []*entities.DailyStatResult{
		{Date: time.Now(), Issued: 1000, Consumed: 500, Transferred: 300},
	}, nil
}
func (m *mockAnalyticsDS) GetTransactionTypeBreakdown(ctx context.Context) ([]*entities.TypeBreakdownResult, error) {
	return []*entities.TypeBreakdownResult{
		{Type: "transfer", Count: 10, TotalAmount: 5000},
	}, nil
}
func (m *mockAnalyticsDS) GetMonthlyIssuedPoints(ctx context.Context) (int64, error) {
	return 10000, nil
}
func (m *mockAnalyticsDS) GetMonthlyTransactionCount(ctx context.Context) (int64, error) {
	return 50, nil
}

// --- Mock Logger ---

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...entities.Field) {}
func (m *mockLogger) Info(msg string, fields ...entities.Field)  {}
func (m *mockLogger) Warn(msg string, fields ...entities.Field)  {}
func (m *mockLogger) Error(msg string, fields ...entities.Field) {}
func (m *mockLogger) Fatal(msg string, fields ...entities.Field) {}

// --- ヘルパー ---

func createTestUserWithBalance(t *testing.T, name string, balance int64, role entities.UserRole) *entities.User {
	t.Helper()
	user, err := entities.NewUser(name, name+"@example.com", "hash", name, "太郎", "田中")
	require.NoError(t, err)
	user.Balance = balance
	user.IsActive = true
	user.Role = role
	return user
}

// ========================================
// AdminInteractor テスト
// ========================================

// --- GrantPoints ---

func TestAdminInteractor_GrantPoints(t *testing.T) {
	setup := func() (*ctxTrackingTxManager, *ctxTrackingUserRepo, *ctxTrackingTransactionRepo, *ctxTrackingIdempotencyRepo, *ctxTrackingPointBatchRepo, inputport.AdminInputPort, *entities.User, *entities.User) {
		txMgr := &ctxTrackingTxManager{}
		userRepo := newCtxTrackingUserRepo()
		txRepo := newCtxTrackingTransactionRepo()
		idempRepo := newCtxTrackingIdempotencyRepo()
		pbRepo := newCtxTrackingPointBatchRepo()
		logger := &mockLogger{}
		analyticsDS := &mockAnalyticsDS{}

		admin := createTestUserWithBalance(t, "admin", 0, "admin")
		target := createTestUserWithBalance(t, "target", 1000, "user")
		userRepo.setUser(admin)
		userRepo.setUser(target)

		i := interactor.NewAdminInteractor(txMgr, userRepo, txRepo, idempRepo, pbRepo, analyticsDS, logger)
		return txMgr, userRepo, txRepo, idempRepo, pbRepo, i, admin, target
	}

	t.Run("正常にポイント付与できる", func(t *testing.T) {
		_, _, _, _, _, sut, admin, target := setup()
		resp, err := sut.GrantPoints(context.Background(), &inputport.GrantPointsRequest{
			AdminID: admin.ID, UserID: target.ID, Amount: 500,
			Description: "test", IdempotencyKey: "grant-" + uuid.New().String(),
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.Transaction)
		assert.NotNil(t, resp.User)
		assert.Equal(t, int64(1500), resp.User.Balance)
	})

	t.Run("txManager.Do内の全呼び出しがトランザクションコンテキストを使用する", func(t *testing.T) {
		txMgr, userRepo, txRepo, idempRepo, pbRepo, sut, admin, target := setup()
		_, err := sut.GrantPoints(context.Background(), &inputport.GrantPointsRequest{
			AdminID: admin.ID, UserID: target.ID, Amount: 100,
			Description: "ctx test", IdempotencyKey: "ctx-grant-" + uuid.New().String(),
		})
		require.NoError(t, err)
		require.NotNil(t, txMgr.TxCtx)

		// txManager.Do 内で呼ばれるメソッドのコンテキスト検証
		assert.True(t, isTxContext(userRepo.ctxRecords["Read_"+target.ID.String()]),
			"userRepo.Read(target) はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(userRepo.ctxRecords["UpdateBalanceWithLock"]),
			"userRepo.UpdateBalanceWithLock はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(txRepo.ctxRecords["Create"]),
			"transactionRepo.Create はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(pbRepo.ctxRecords["Create"]),
			"pointBatchRepo.Create はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(idempRepo.ctxRecords["Create"]),
			"idempotencyRepo.Create はトランザクションコンテキストを使用すべき")
	})

	t.Run("金額が0以下ならエラー", func(t *testing.T) {
		_, _, _, _, _, sut, admin, target := setup()
		_, err := sut.GrantPoints(context.Background(), &inputport.GrantPointsRequest{
			AdminID: admin.ID, UserID: target.ID, Amount: 0,
			Description: "test", IdempotencyKey: "key1",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("管理者権限がないとエラー", func(t *testing.T) {
		_, _, _, _, _, sut, _, target := setup()
		nonAdmin := createTestUserWithBalance(t, "nonadmin", 0, "user")
		_, err := sut.GrantPoints(context.Background(), &inputport.GrantPointsRequest{
			AdminID: nonAdmin.ID, UserID: target.ID, Amount: 100,
			Description: "test", IdempotencyKey: "key2",
		})
		assert.Error(t, err)
	})

	t.Run("対象ユーザーが存在しないとエラー", func(t *testing.T) {
		_, _, _, _, _, sut, admin, _ := setup()
		_, err := sut.GrantPoints(context.Background(), &inputport.GrantPointsRequest{
			AdminID: admin.ID, UserID: uuid.New(), Amount: 100,
			Description: "test", IdempotencyKey: "key3",
		})
		assert.Error(t, err)
	})

	t.Run("非アクティブユーザーにはポイント付与できない", func(t *testing.T) {
		_, userRepo, _, _, _, sut, admin, _ := setup()
		inactive := createTestUserWithBalance(t, "inactive", 0, "user")
		inactive.IsActive = false
		userRepo.setUser(inactive)

		_, err := sut.GrantPoints(context.Background(), &inputport.GrantPointsRequest{
			AdminID: admin.ID, UserID: inactive.ID, Amount: 100,
			Description: "test", IdempotencyKey: "key4",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})

	t.Run("冪等性キーが既に使用済みの場合は既存の結果を返す", func(t *testing.T) {
		_, _, _, _, _, sut, admin, target := setup()
		key := "idempotent-grant-" + uuid.New().String()

		resp1, err := sut.GrantPoints(context.Background(), &inputport.GrantPointsRequest{
			AdminID: admin.ID, UserID: target.ID, Amount: 100,
			Description: "test", IdempotencyKey: key,
		})
		require.NoError(t, err)

		// 同じキーで再度呼ぶ
		resp2, err := sut.GrantPoints(context.Background(), &inputport.GrantPointsRequest{
			AdminID: admin.ID, UserID: target.ID, Amount: 100,
			Description: "test", IdempotencyKey: key,
		})
		require.NoError(t, err)
		assert.Equal(t, resp1.Transaction.ID, resp2.Transaction.ID)
	})
}

// --- DeductPoints ---

func TestAdminInteractor_DeductPoints(t *testing.T) {
	setup := func() (*ctxTrackingTxManager, *ctxTrackingUserRepo, *ctxTrackingTransactionRepo, *ctxTrackingIdempotencyRepo, inputport.AdminInputPort, *entities.User, *entities.User) {
		txMgr := &ctxTrackingTxManager{}
		userRepo := newCtxTrackingUserRepo()
		txRepo := newCtxTrackingTransactionRepo()
		idempRepo := newCtxTrackingIdempotencyRepo()
		pbRepo := newCtxTrackingPointBatchRepo()
		logger := &mockLogger{}
		analyticsDS := &mockAnalyticsDS{}

		admin := createTestUserWithBalance(t, "admin", 0, "admin")
		target := createTestUserWithBalance(t, "target", 10000, "user")
		userRepo.setUser(admin)
		userRepo.setUser(target)

		i := interactor.NewAdminInteractor(txMgr, userRepo, txRepo, idempRepo, pbRepo, analyticsDS, logger)
		return txMgr, userRepo, txRepo, idempRepo, i, admin, target
	}

	t.Run("正常にポイント減算できる", func(t *testing.T) {
		_, _, _, _, sut, admin, target := setup()
		resp, err := sut.DeductPoints(context.Background(), &inputport.DeductPointsRequest{
			AdminID: admin.ID, UserID: target.ID, Amount: 500,
			Description: "test", IdempotencyKey: "deduct-" + uuid.New().String(),
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.Transaction)
		assert.Equal(t, int64(9500), resp.User.Balance)
	})

	t.Run("txManager.Do内の全呼び出しがトランザクションコンテキストを使用する", func(t *testing.T) {
		txMgr, userRepo, txRepo, idempRepo, sut, admin, target := setup()
		_, err := sut.DeductPoints(context.Background(), &inputport.DeductPointsRequest{
			AdminID: admin.ID, UserID: target.ID, Amount: 100,
			Description: "ctx test", IdempotencyKey: "ctx-deduct-" + uuid.New().String(),
		})
		require.NoError(t, err)
		require.NotNil(t, txMgr.TxCtx)

		assert.True(t, isTxContext(userRepo.ctxRecords["Read_"+target.ID.String()]),
			"userRepo.Read(target) はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(userRepo.ctxRecords["UpdateBalanceWithLock"]),
			"userRepo.UpdateBalanceWithLock はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(txRepo.ctxRecords["Create"]),
			"transactionRepo.Create はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(idempRepo.ctxRecords["Create"]),
			"idempotencyRepo.Create はトランザクションコンテキストを使用すべき")
	})

	t.Run("残高不足ならエラー", func(t *testing.T) {
		_, _, _, _, sut, admin, target := setup()
		_, err := sut.DeductPoints(context.Background(), &inputport.DeductPointsRequest{
			AdminID: admin.ID, UserID: target.ID, Amount: 99999,
			Description: "too much", IdempotencyKey: "deduct-fail-" + uuid.New().String(),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
	})

	t.Run("金額が0以下ならエラー", func(t *testing.T) {
		_, _, _, _, sut, admin, target := setup()
		_, err := sut.DeductPoints(context.Background(), &inputport.DeductPointsRequest{
			AdminID: admin.ID, UserID: target.ID, Amount: -1,
			Description: "test", IdempotencyKey: "key",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("管理者権限がないとエラー", func(t *testing.T) {
		_, _, _, _, sut, _, target := setup()
		_, err := sut.DeductPoints(context.Background(), &inputport.DeductPointsRequest{
			AdminID: uuid.New(), UserID: target.ID, Amount: 100,
			Description: "test", IdempotencyKey: "key",
		})
		assert.Error(t, err)
	})
}

// --- ListAllUsers ---

func TestAdminInteractor_ListAllUsers(t *testing.T) {
	setup := func() (inputport.AdminInputPort, *ctxTrackingUserRepo) {
		userRepo := newCtxTrackingUserRepo()
		u1 := createTestUserWithBalance(t, "user1", 1000, "user")
		u2 := createTestUserWithBalance(t, "user2", 2000, "user")
		userRepo.setUser(u1)
		userRepo.setUser(u2)

		i := interactor.NewAdminInteractor(
			&ctxTrackingTxManager{}, userRepo, newCtxTrackingTransactionRepo(),
			newCtxTrackingIdempotencyRepo(), newCtxTrackingPointBatchRepo(),
			&mockAnalyticsDS{}, &mockLogger{},
		)
		return i, userRepo
	}

	t.Run("検索なしでユーザー一覧を取得できる", func(t *testing.T) {
		sut, _ := setup()
		resp, err := sut.ListAllUsers(context.Background(), &inputport.ListAllUsersRequest{
			Offset: 0, Limit: 20,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, len(resp.Users))
		assert.Equal(t, 2, resp.Total)
	})

	t.Run("検索ありでユーザー一覧を取得できる", func(t *testing.T) {
		sut, userRepo := setup()
		resp, err := sut.ListAllUsers(context.Background(), &inputport.ListAllUsersRequest{
			Offset: 0, Limit: 20, Search: "user1",
		})
		require.NoError(t, err)
		// ReadListWithSearch が呼ばれたことを確認
		_, ok := userRepo.ctxRecords["ReadListWithSearch"]
		assert.True(t, ok, "検索時は ReadListWithSearch が呼ばれるべき")
		assert.NotNil(t, resp)
	})
}

// --- ListAllTransactions ---

func TestAdminInteractor_ListAllTransactions(t *testing.T) {
	setup := func() inputport.AdminInputPort {
		userRepo := newCtxTrackingUserRepo()
		txRepo := newCtxTrackingTransactionRepo()

		i := interactor.NewAdminInteractor(
			&ctxTrackingTxManager{}, userRepo, txRepo,
			newCtxTrackingIdempotencyRepo(), newCtxTrackingPointBatchRepo(),
			&mockAnalyticsDS{}, &mockLogger{},
		)
		return i
	}

	t.Run("フィルタなしで取引履歴を取得できる", func(t *testing.T) {
		sut := setup()
		resp, err := sut.ListAllTransactions(context.Background(), &inputport.ListAllTransactionsRequest{
			Offset: 0, Limit: 20,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("フィルタありで取引履歴を取得できる", func(t *testing.T) {
		sut := setup()
		resp, err := sut.ListAllTransactions(context.Background(), &inputport.ListAllTransactionsRequest{
			Offset: 0, Limit: 20, TransactionType: "daily_bonus",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

// --- UpdateUserRole ---

func TestAdminInteractor_UpdateUserRole(t *testing.T) {
	setup := func() (inputport.AdminInputPort, *entities.User, *entities.User) {
		userRepo := newCtxTrackingUserRepo()
		admin := createTestUserWithBalance(t, "admin", 0, "admin")
		target := createTestUserWithBalance(t, "target", 0, "user")
		userRepo.setUser(admin)
		userRepo.setUser(target)

		i := interactor.NewAdminInteractor(
			&ctxTrackingTxManager{}, userRepo, newCtxTrackingTransactionRepo(),
			newCtxTrackingIdempotencyRepo(), newCtxTrackingPointBatchRepo(),
			&mockAnalyticsDS{}, &mockLogger{},
		)
		return i, admin, target
	}

	t.Run("正常にロールを更新できる", func(t *testing.T) {
		sut, admin, target := setup()
		resp, err := sut.UpdateUserRole(context.Background(), &inputport.UpdateUserRoleRequest{
			AdminID: admin.ID, UserID: target.ID, Role: "admin",
		})
		require.NoError(t, err)
		assert.Equal(t, entities.UserRole("admin"), resp.User.Role)
	})

	t.Run("無効なロールならエラー", func(t *testing.T) {
		sut, admin, target := setup()
		_, err := sut.UpdateUserRole(context.Background(), &inputport.UpdateUserRoleRequest{
			AdminID: admin.ID, UserID: target.ID, Role: "superadmin",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})

	t.Run("管理者権限がないとエラー", func(t *testing.T) {
		sut, _, target := setup()
		_, err := sut.UpdateUserRole(context.Background(), &inputport.UpdateUserRoleRequest{
			AdminID: uuid.New(), UserID: target.ID, Role: "admin",
		})
		assert.Error(t, err)
	})
}

// --- DeactivateUser ---

func TestAdminInteractor_DeactivateUser(t *testing.T) {
	setup := func() (inputport.AdminInputPort, *entities.User, *entities.User) {
		userRepo := newCtxTrackingUserRepo()
		admin := createTestUserWithBalance(t, "admin", 0, "admin")
		target := createTestUserWithBalance(t, "target", 0, "user")
		userRepo.setUser(admin)
		userRepo.setUser(target)

		i := interactor.NewAdminInteractor(
			&ctxTrackingTxManager{}, userRepo, newCtxTrackingTransactionRepo(),
			newCtxTrackingIdempotencyRepo(), newCtxTrackingPointBatchRepo(),
			&mockAnalyticsDS{}, &mockLogger{},
		)
		return i, admin, target
	}

	t.Run("正常にユーザーを無効化できる", func(t *testing.T) {
		sut, admin, target := setup()
		resp, err := sut.DeactivateUser(context.Background(), &inputport.DeactivateUserRequest{
			AdminID: admin.ID, UserID: target.ID,
		})
		require.NoError(t, err)
		assert.False(t, resp.User.IsActive)
	})

	t.Run("自分自身を無効化しようとするとエラー", func(t *testing.T) {
		sut, admin, _ := setup()
		_, err := sut.DeactivateUser(context.Background(), &inputport.DeactivateUserRequest{
			AdminID: admin.ID, UserID: admin.ID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot deactivate yourself")
	})

	t.Run("管理者権限がないとエラー", func(t *testing.T) {
		sut, _, target := setup()
		_, err := sut.DeactivateUser(context.Background(), &inputport.DeactivateUserRequest{
			AdminID: uuid.New(), UserID: target.ID,
		})
		assert.Error(t, err)
	})
}

// --- GetAnalytics ---

func TestAdminInteractor_GetAnalytics(t *testing.T) {
	t.Run("正常に分析データを取得できる", func(t *testing.T) {
		sut := interactor.NewAdminInteractor(
			&ctxTrackingTxManager{}, newCtxTrackingUserRepo(), newCtxTrackingTransactionRepo(),
			newCtxTrackingIdempotencyRepo(), newCtxTrackingPointBatchRepo(),
			&mockAnalyticsDS{}, &mockLogger{},
		)

		resp, err := sut.GetAnalytics(context.Background(), &inputport.GetAnalyticsRequest{
			Days: 7,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.Summary)
		assert.Equal(t, int64(100000), resp.Summary.TotalPointsInCirculation)
		assert.Equal(t, int64(20), resp.Summary.ActiveUsers)
		assert.NotEmpty(t, resp.TopHolders)
		assert.NotEmpty(t, resp.DailyStats)
		assert.NotEmpty(t, resp.TransactionTypeBreakdown)
	})
}
