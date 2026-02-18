package infraakerun_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/infraakerun"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// Mock Repositories
// ========================================

// mockDailyBonusRepo は DailyBonusRepository のモック
type mockDailyBonusRepo struct {
	bonuses      map[string]*entities.DailyBonus // key: "userID-bonusDate"
	lastPolledAt time.Time
	created      []*entities.DailyBonus
}

func newMockDailyBonusRepo() *mockDailyBonusRepo {
	return &mockDailyBonusRepo{
		bonuses:      make(map[string]*entities.DailyBonus),
		lastPolledAt: time.Now().Add(-1 * time.Hour), // 1時間前
		created:      make([]*entities.DailyBonus, 0),
	}
}

func (m *mockDailyBonusRepo) Create(ctx context.Context, bonus *entities.DailyBonus) error {
	key := fmt.Sprintf("%s-%s", bonus.UserID.String(), bonus.BonusDate.Format("2006-01-02"))
	if _, exists := m.bonuses[key]; exists {
		return fmt.Errorf("duplicate bonus for user %s on %s", bonus.UserID, bonus.BonusDate.Format("2006-01-02"))
	}
	m.bonuses[key] = bonus
	m.created = append(m.created, bonus)
	return nil
}

func (m *mockDailyBonusRepo) ReadByUserAndDate(ctx context.Context, userID uuid.UUID, bonusDate time.Time) (*entities.DailyBonus, error) {
	key := fmt.Sprintf("%s-%s", userID.String(), bonusDate.Format("2006-01-02"))
	if bonus, ok := m.bonuses[key]; ok {
		return bonus, nil
	}
	return nil, nil
}

func (m *mockDailyBonusRepo) ReadRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.DailyBonus, error) {
	var result []*entities.DailyBonus
	for _, bonus := range m.bonuses {
		if bonus.UserID == userID {
			result = append(result, bonus)
		}
	}
	return result, nil
}

func (m *mockDailyBonusRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	for _, bonus := range m.bonuses {
		if bonus.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockDailyBonusRepo) GetLastPolledAt(ctx context.Context) (time.Time, error) {
	return m.lastPolledAt, nil
}

func (m *mockDailyBonusRepo) UpdateLastPolledAt(ctx context.Context, t time.Time) error {
	m.lastPolledAt = t
	return nil
}

// mockUserRepo は UserRepository のモック
type mockUserRepo struct {
	users          map[uuid.UUID]*entities.User
	balanceUpdates []repository.BalanceUpdate
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:          make(map[uuid.UUID]*entities.User),
		balanceUpdates: make([]repository.BalanceUpdate, 0),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) error { return nil }
func (m *mockUserRepo) Read(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}
func (m *mockUserRepo) ReadByUsername(ctx context.Context, username string) (*entities.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ReadByEmail(ctx context.Context, email string) (*entities.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *entities.User) (bool, error) {
	return true, nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockUserRepo) ReadList(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	var result []*entities.User
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}
func (m *mockUserRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}
func (m *mockUserRepo) UpdateBalanceWithLock(ctx context.Context, userID uuid.UUID, amount int64, isDeduct bool) error {
	return nil
}
func (m *mockUserRepo) UpdateBalancesWithLock(ctx context.Context, updates []repository.BalanceUpdate) error {
	m.balanceUpdates = append(m.balanceUpdates, updates...)
	// 実際に残高も更新
	for _, update := range updates {
		if u, ok := m.users[update.UserID]; ok {
			if update.IsDeduct {
				u.Balance -= update.Amount
			} else {
				u.Balance += update.Amount
			}
		}
	}
	return nil
}

func (m *mockUserRepo) addUser(user *entities.User) {
	m.users[user.ID] = user
}

// mockTransactionRepo は TransactionRepository のモック
type mockTransactionRepo struct {
	transactions []*entities.Transaction
}

func newMockTransactionRepo() *mockTransactionRepo {
	return &mockTransactionRepo{
		transactions: make([]*entities.Transaction, 0),
	}
}

func (m *mockTransactionRepo) Create(ctx context.Context, tx *entities.Transaction) error {
	m.transactions = append(m.transactions, tx)
	return nil
}
func (m *mockTransactionRepo) Read(ctx context.Context, id uuid.UUID) (*entities.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionRepo) ReadByIdempotencyKey(ctx context.Context, key string) (*entities.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionRepo) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionRepo) ReadListAll(ctx context.Context, offset, limit int) ([]*entities.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionRepo) Update(ctx context.Context, tx *entities.Transaction) error {
	return nil
}
func (m *mockTransactionRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}

// mockTxManager は TransactionManager のモック（そのまま実行）
type mockTxManager struct{}

func (m *mockTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// mockSystemSettingsRepo は SystemSettingsRepository のモック
type mockSystemSettingsRepo struct {
	settings map[string]string
}

func newMockSystemSettingsRepo() *mockSystemSettingsRepo {
	return &mockSystemSettingsRepo{
		settings: make(map[string]string),
	}
}

func (m *mockSystemSettingsRepo) GetSetting(ctx context.Context, key string) (string, error) {
	v, ok := m.settings[key]
	if !ok {
		return "", fmt.Errorf("setting not found: %s", key)
	}
	return v, nil
}

func (m *mockSystemSettingsRepo) SetSetting(ctx context.Context, key, value, description string) error {
	m.settings[key] = value
	return nil
}

// mockPointBatchRepo は PointBatchRepository のモック
type mockPointBatchRepo struct{}

func (m *mockPointBatchRepo) Create(ctx context.Context, batch *entities.PointBatch) error {
	return nil
}
func (m *mockPointBatchRepo) ConsumePointsFIFO(ctx context.Context, userID uuid.UUID, amount int64) error {
	return nil
}
func (m *mockPointBatchRepo) FindExpiredBatches(ctx context.Context, before time.Time, limit int) ([]*entities.PointBatch, error) {
	return nil, nil
}
func (m *mockPointBatchRepo) MarkExpired(ctx context.Context, batchID uuid.UUID) error {
	return nil
}

// mockLogger はテスト用ログ
type mockLogger struct {
	infos  []string
	errors []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{}
}

func (m *mockLogger) Debug(msg string, fields ...entities.Field) {}
func (m *mockLogger) Info(msg string, fields ...entities.Field)  { m.infos = append(m.infos, msg) }
func (m *mockLogger) Warn(msg string, fields ...entities.Field)  {}
func (m *mockLogger) Error(msg string, fields ...entities.Field) { m.errors = append(m.errors, msg) }
func (m *mockLogger) Fatal(msg string, fields ...entities.Field) {}

// mockTimeProvider はテスト用の時刻プロバイダー
type mockTimeProvider struct {
	now time.Time
}

func newMockTimeProvider(t time.Time) *mockTimeProvider {
	return &mockTimeProvider{now: t}
}

func (m *mockTimeProvider) Now() time.Time { return m.now }

// ========================================
// Helper: Akerun APIモックサーバー
// ========================================

// akerunAPIResponse はAkerun APIの成功レスポンスを表す構造体
// https://developers.akerun.com/#list-access の例に準拠
type akerunAPIResponse struct {
	Accesses []akerunAccessJSON `json:"accesses"`
}

type akerunAccessJSON struct {
	ID         json.Number     `json:"id"`
	Action     string          `json:"action"`
	DeviceType string          `json:"device_type"`
	DeviceName string          `json:"device_name"`
	AccessedAt string          `json:"accessed_at"`
	Akerun     *akerunInfoJSON `json:"akerun"`
	User       *akerunUserJSON `json:"user"`
}

type akerunInfoJSON struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

type akerunUserJSON struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

// createMockAkerunServer はAkerun APIのモックHTTPサーバーを作成する
func createMockAkerunServer(response akerunAPIResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization ヘッダーの確認
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// ========================================
// Akerun API公式ドキュメントのExampleレスポンス
// https://developers.akerun.com/#list-access
// ========================================

func createAkerunExampleResponse() akerunAPIResponse {
	return akerunAPIResponse{
		Accesses: []akerunAccessJSON{
			{
				ID:         json.Number("1234567890921123456789"),
				Action:     "unlock",
				DeviceType: "akerun_app",
				DeviceName: "iOS iPhone",
				AccessedAt: "2017-07-24T06:37:19Z",
				Akerun: &akerunInfoJSON{
					ID:       "A1030001",
					Name:     "執務室表口",
					ImageURL: "https://akerun.com/akerun_example1.jpg",
				},
				User: &akerunUserJSON{
					ID:       "U-ab345-678ij",
					Name:     "Photosynth太郎",
					ImageURL: "https://akerun.com/user_example1.jpg",
				},
			},
			{
				ID:         json.Number("1234567890921123456790"),
				Action:     "unlock",
				DeviceType: "akerun_app",
				DeviceName: "iOS iPhone",
				AccessedAt: "2017-07-24T06:40:19Z",
				Akerun: &akerunInfoJSON{
					ID:       "A1030002",
					Name:     "執務室裏口",
					ImageURL: "https://akerun.com/akerun_example2.jpg",
				},
				User: &akerunUserJSON{
					ID:       "U-ab345-678ij",
					Name:     "Photosynth太郎",
					ImageURL: "https://akerun.com/user_example2.jpg",
				},
			},
		},
	}
}

// ========================================
// テストケース
// ========================================

func TestAkerunWorkerProcessAccesses_ExampleResponse(t *testing.T) {
	t.Run("Akerun APIドキュメントの例レスポンスでポイントが付与される", func(t *testing.T) {
		// === Setup ===
		// Akerun APIモックサーバー（公式ドキュメントの例レスポンス）
		mockResponse := createAkerunExampleResponse()
		server := createMockAkerunServer(mockResponse)
		defer server.Close()

		// リポジトリ準備
		dailyBonusRepo := newMockDailyBonusRepo()
		dailyBonusRepo.lastPolledAt = time.Date(2017, 7, 24, 0, 0, 0, 0, time.UTC)

		userRepo := newMockUserRepo()
		transactionRepo := newMockTransactionRepo()
		txManager := &mockTxManager{}
		systemSettingsRepo := newMockSystemSettingsRepo()
		logger := newMockLogger()

		// テスト用ユーザー: "Photosynth太郎" とマッチするように LastName+FirstName を設定
		userID := uuid.New()
		user := &entities.User{
			ID:        userID,
			Username:  "photosynth_taro",
			LastName:  "Photosynth",
			FirstName: "太郎",
			Balance:   100,
			IsActive:  true,
			Role:      entities.RoleUser,
		}
		userRepo.addUser(user)

		// Akerunクライアント（モックサーバーに接続）
		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "test-token",
			OrganizationID: "O-ab345-678ij",
			BaseURL:        server.URL,
		})

		// Worker作成
		worker := infraakerun.NewAkerunWorker(
			client,
			dailyBonusRepo,
			userRepo,
			transactionRepo,
			txManager,
			systemSettingsRepo,
			&mockPointBatchRepo{},
			&mockTimeProvider{now: time.Now()},
			logger,
		)

		// === Execute ===
		// Exportedではないpollメソッドをテストするため、Startを使わず
		// AkerunClientを直接呼んでprocessAccessesを間接的にテスト
		// → poll() を呼びたいがexportedではない。poll()相当の動きをテスト。
		ctx := context.Background()
		accesses, err := client.GetAccesses(ctx,
			time.Date(2017, 7, 23, 10, 0, 0, 0, time.UTC),
			time.Date(2017, 7, 29, 19, 0, 0, 0, time.UTC),
			300,
		)
		require.NoError(t, err)
		require.Len(t, accesses, 2, "Akerun APIから2件のアクセス記録を取得")

		// Worker.ProcessAccessesForTest を使う（テスト用エクスポート）
		worker.ProcessAccessesForTest(ctx, accesses)

		// === Assert ===
		// 1. ボーナスレコードが1件作成される（同じユーザー・同じ日は1回のみ）
		assert.Len(t, dailyBonusRepo.created, 1, "同一ユーザー・同一日のアクセスは1件のみボーナス付与")

		bonus := dailyBonusRepo.created[0]
		assert.Equal(t, userID, bonus.UserID, "正しいユーザーにボーナスが付与される")
		assert.Equal(t, int64(5), bonus.BonusPoints, "デフォルト5ポイントが付与される")
		assert.Equal(t, "Photosynth太郎", bonus.AkerunUserName, "Akerunユーザー名が記録される")
		assert.NotNil(t, bonus.AccessedAt, "アクセス時刻が記録される")

		// 2. トランザクションが1件作成される
		assert.Len(t, transactionRepo.transactions, 1, "トランザクションが1件記録される")
		tx := transactionRepo.transactions[0]
		assert.Equal(t, int64(5), tx.Amount, "トランザクション金額が5ポイント")

		// 3. ユーザー残高が更新される（100 + 5 = 105）
		assert.Len(t, userRepo.balanceUpdates, 1, "残高更新が1回実行される")
		assert.Equal(t, userID, userRepo.balanceUpdates[0].UserID, "正しいユーザーの残高が更新される")
		assert.Equal(t, int64(5), userRepo.balanceUpdates[0].Amount, "5ポイント加算")
		assert.False(t, userRepo.balanceUpdates[0].IsDeduct, "加算（減算ではない）")
		assert.Equal(t, int64(105), userRepo.users[userID].Balance, "残高が100→105に更新される")
	})

	t.Run("管理者設定でボーナスポイントを変更した場合", func(t *testing.T) {
		// === Setup ===
		mockResponse := createAkerunExampleResponse()
		server := createMockAkerunServer(mockResponse)
		defer server.Close()

		dailyBonusRepo := newMockDailyBonusRepo()
		dailyBonusRepo.lastPolledAt = time.Date(2017, 7, 24, 0, 0, 0, 0, time.UTC)

		userRepo := newMockUserRepo()
		transactionRepo := newMockTransactionRepo()
		txManager := &mockTxManager{}
		systemSettingsRepo := newMockSystemSettingsRepo()
		logger := newMockLogger()

		// ボーナスポイントを10に設定
		systemSettingsRepo.SetSetting(context.Background(), "akerun_bonus_points", "10", "テスト")

		userID := uuid.New()
		userRepo.addUser(&entities.User{
			ID:        userID,
			Username:  "photosynth_taro",
			LastName:  "Photosynth",
			FirstName: "太郎",
			Balance:   200,
			IsActive:  true,
			Role:      entities.RoleUser,
		})

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "test-token",
			OrganizationID: "O-ab345-678ij",
			BaseURL:        server.URL,
		})

		worker := infraakerun.NewAkerunWorker(
			client, dailyBonusRepo, userRepo, transactionRepo,
			txManager, systemSettingsRepo, &mockPointBatchRepo{}, &mockTimeProvider{}, logger,
		)

		// === Execute ===
		ctx := context.Background()
		accesses, err := client.GetAccesses(ctx,
			time.Date(2017, 7, 23, 10, 0, 0, 0, time.UTC),
			time.Date(2017, 7, 29, 19, 0, 0, 0, time.UTC),
			300,
		)
		require.NoError(t, err)
		worker.ProcessAccessesForTest(ctx, accesses)

		// === Assert ===
		require.Len(t, dailyBonusRepo.created, 1)
		assert.Equal(t, int64(10), dailyBonusRepo.created[0].BonusPoints, "管理者設定の10ポイントが適用される")
		assert.Equal(t, int64(210), userRepo.users[userID].Balance, "残高が200→210に更新される")
	})

	t.Run("マッチしないユーザーのアクセスはスキップされる", func(t *testing.T) {
		// === Setup ===
		mockResponse := createAkerunExampleResponse()
		server := createMockAkerunServer(mockResponse)
		defer server.Close()

		dailyBonusRepo := newMockDailyBonusRepo()
		userRepo := newMockUserRepo()
		transactionRepo := newMockTransactionRepo()
		txManager := &mockTxManager{}
		systemSettingsRepo := newMockSystemSettingsRepo()
		logger := newMockLogger()

		// Akerun上の "Photosynth太郎" とマッチしないユーザーのみ登録
		userRepo.addUser(&entities.User{
			ID:        uuid.New(),
			Username:  "yamada",
			LastName:  "山田",
			FirstName: "花子",
			Balance:   100,
			IsActive:  true,
			Role:      entities.RoleUser,
		})

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "test-token",
			OrganizationID: "O-ab345-678ij",
			BaseURL:        server.URL,
		})

		worker := infraakerun.NewAkerunWorker(
			client, dailyBonusRepo, userRepo, transactionRepo,
			txManager, systemSettingsRepo, &mockPointBatchRepo{}, &mockTimeProvider{}, logger,
		)

		// === Execute ===
		ctx := context.Background()
		accesses, err := client.GetAccesses(ctx,
			time.Date(2017, 7, 23, 10, 0, 0, 0, time.UTC),
			time.Date(2017, 7, 29, 19, 0, 0, 0, time.UTC),
			300,
		)
		require.NoError(t, err)
		worker.ProcessAccessesForTest(ctx, accesses)

		// === Assert ===
		assert.Len(t, dailyBonusRepo.created, 0, "マッチしないユーザーにはボーナスが付与されない")
		assert.Len(t, transactionRepo.transactions, 0, "トランザクションも作成されない")
		assert.Len(t, userRepo.balanceUpdates, 0, "残高更新もされない")
	})

	t.Run("既にボーナス付与済みの場合は重複付与されない", func(t *testing.T) {
		// === Setup ===
		mockResponse := createAkerunExampleResponse()
		server := createMockAkerunServer(mockResponse)
		defer server.Close()

		dailyBonusRepo := newMockDailyBonusRepo()
		userRepo := newMockUserRepo()
		transactionRepo := newMockTransactionRepo()
		txManager := &mockTxManager{}
		systemSettingsRepo := newMockSystemSettingsRepo()
		logger := newMockLogger()

		userID := uuid.New()
		userRepo.addUser(&entities.User{
			ID:        userID,
			Username:  "photosynth_taro",
			LastName:  "Photosynth",
			FirstName: "太郎",
			Balance:   100,
			IsActive:  true,
			Role:      entities.RoleUser,
		})

		// 既にボーナスを付与済み
		jst := time.FixedZone("JST", 9*60*60)
		bonusDate := time.Date(2017, 7, 24, 0, 0, 0, 0, jst)
		existingBonus := entities.NewDailyBonus(userID, bonusDate, 5, "existing", "Photosynth太郎", nil)
		dailyBonusRepo.bonuses[fmt.Sprintf("%s-%s", userID.String(), bonusDate.Format("2006-01-02"))] = existingBonus

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "test-token",
			OrganizationID: "O-ab345-678ij",
			BaseURL:        server.URL,
		})

		worker := infraakerun.NewAkerunWorker(
			client, dailyBonusRepo, userRepo, transactionRepo,
			txManager, systemSettingsRepo, &mockPointBatchRepo{}, &mockTimeProvider{}, logger,
		)

		// === Execute ===
		ctx := context.Background()
		accesses, err := client.GetAccesses(ctx,
			time.Date(2017, 7, 23, 10, 0, 0, 0, time.UTC),
			time.Date(2017, 7, 29, 19, 0, 0, 0, time.UTC),
			300,
		)
		require.NoError(t, err)
		worker.ProcessAccessesForTest(ctx, accesses)

		// === Assert ===
		assert.Len(t, dailyBonusRepo.created, 0, "既に付与済みなので新規ボーナスは作成されない")
		assert.Len(t, transactionRepo.transactions, 0, "トランザクションも作成されない")
		assert.Equal(t, int64(100), userRepo.users[userID].Balance, "残高は変わらない")
	})

	t.Run("userフィールドがnullのアクセスレコードはスキップされる", func(t *testing.T) {
		// === Setup ===
		mockResponse := akerunAPIResponse{
			Accesses: []akerunAccessJSON{
				{
					ID:         json.Number("9999999999"),
					Action:     "unlock",
					DeviceType: "nfc",
					DeviceName: "ICカード",
					AccessedAt: "2017-07-24T10:00:00Z",
					Akerun: &akerunInfoJSON{
						ID:   "A1030001",
						Name: "執務室",
					},
					User: nil, // ユーザー情報がない
				},
			},
		}
		server := createMockAkerunServer(mockResponse)
		defer server.Close()

		dailyBonusRepo := newMockDailyBonusRepo()
		userRepo := newMockUserRepo()
		transactionRepo := newMockTransactionRepo()
		txManager := &mockTxManager{}
		systemSettingsRepo := newMockSystemSettingsRepo()
		logger := newMockLogger()

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "test-token",
			OrganizationID: "O-ab345-678ij",
			BaseURL:        server.URL,
		})

		worker := infraakerun.NewAkerunWorker(
			client, dailyBonusRepo, userRepo, transactionRepo,
			txManager, systemSettingsRepo, &mockPointBatchRepo{}, &mockTimeProvider{}, logger,
		)

		// === Execute ===
		ctx := context.Background()
		accesses, err := client.GetAccesses(ctx,
			time.Date(2017, 7, 23, 0, 0, 0, 0, time.UTC),
			time.Date(2017, 7, 29, 0, 0, 0, 0, time.UTC),
			300,
		)
		require.NoError(t, err)
		worker.ProcessAccessesForTest(ctx, accesses)

		// === Assert ===
		assert.Len(t, dailyBonusRepo.created, 0, "ユーザー情報なしのレコードはスキップ")
	})
}

// ========================================
// AkerunClient テスト
// ========================================

func TestAkerunClient_GetAccesses(t *testing.T) {
	t.Run("正常にアクセス履歴を取得できる", func(t *testing.T) {
		mockResponse := createAkerunExampleResponse()
		server := createMockAkerunServer(mockResponse)
		defer server.Close()

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "test-token",
			OrganizationID: "O-ab345-678ij",
			BaseURL:        server.URL,
		})

		accesses, err := client.GetAccesses(context.Background(),
			time.Date(2017, 7, 23, 10, 0, 0, 0, time.UTC),
			time.Date(2017, 7, 29, 19, 0, 0, 0, time.UTC),
			300,
		)

		require.NoError(t, err)
		require.Len(t, accesses, 2)

		// 1件目のアクセスレコード検証
		assert.Equal(t, "unlock", accesses[0].Action)
		assert.Equal(t, "akerun_app", accesses[0].DeviceType)
		assert.Equal(t, "iOS iPhone", accesses[0].DeviceName)
		assert.Equal(t, "2017-07-24T06:37:19Z", accesses[0].AccessedAt)
		assert.Equal(t, "Photosynth太郎", accesses[0].User.Name)
		assert.Equal(t, "U-ab345-678ij", accesses[0].User.ID)
		assert.Equal(t, "A1030001", accesses[0].Akerun.ID)
		assert.Equal(t, "執務室表口", accesses[0].Akerun.Name)

		// 2件目のアクセスレコード検証
		assert.Equal(t, "2017-07-24T06:40:19Z", accesses[1].AccessedAt)
		assert.Equal(t, "A1030002", accesses[1].Akerun.ID)
		assert.Equal(t, "執務室裏口", accesses[1].Akerun.Name)
	})

	t.Run("認証エラーの場合はエラーを返す", func(t *testing.T) {
		mockResponse := createAkerunExampleResponse()
		server := createMockAkerunServer(mockResponse)
		defer server.Close()

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "invalid-token",
			OrganizationID: "O-ab345-678ij",
			BaseURL:        server.URL,
		})

		_, err := client.GetAccesses(context.Background(),
			time.Date(2017, 7, 23, 10, 0, 0, 0, time.UTC),
			time.Date(2017, 7, 29, 19, 0, 0, 0, time.UTC),
			300,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "401")
	})

	t.Run("IsConfiguredの動作確認", func(t *testing.T) {
		client1 := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "token",
			OrganizationID: "org",
		})
		assert.True(t, client1.IsConfigured())

		client2 := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "",
			OrganizationID: "org",
		})
		assert.False(t, client2.IsConfigured())

		client3 := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "token",
			OrganizationID: "",
		})
		assert.False(t, client3.IsConfigured())
	})
}

// ========================================
// normalizeName テスト
// ========================================

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"半角スペースを除去", "Photosynth 太郎", "photosynth太郎"},
		{"全角スペースを除去", "Photosynth　太郎", "photosynth太郎"},
		{"スペースなし", "Photosynth太郎", "photosynth太郎"},
		{"英語大文字を小文字化", "TARO YAMADA", "taroyamada"},
		{"前後の空白を除去", "  田中太郎  ", "田中太郎"},
		{"空文字", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := infraakerun.NormalizeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ========================================
// リカバリポーリング テスト
// ========================================

// createCountingMockServer はAPIリクエスト回数を記録するモックサーバー
func createCountingMockServer(response akerunAPIResponse) (*httptest.Server, *int) {
	requestCount := new(int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*requestCount++
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	return server, requestCount
}

func TestRecoveryPolling(t *testing.T) {
	t.Run("通常モード: gapが10分以内の場合はAPI1回呼び出し", func(t *testing.T) {
		// === Setup ===
		mockResponse := createAkerunExampleResponse()
		server, requestCount := createCountingMockServer(mockResponse)
		defer server.Close()

		nowTime := time.Date(2026, 2, 17, 17, 5, 0, 0, time.UTC)

		dailyBonusRepo := newMockDailyBonusRepo()
		dailyBonusRepo.lastPolledAt = nowTime.Add(-5 * time.Minute) // 5分前

		userRepo := newMockUserRepo()
		userID := uuid.New()
		userRepo.addUser(&entities.User{
			ID: userID, Username: "photosynth_taro",
			LastName: "Photosynth", FirstName: "太郎",
			Balance: 100, IsActive: true, Role: entities.RoleUser,
		})

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken: "test-token", OrganizationID: "O-test",
			BaseURL: server.URL,
		})

		worker := infraakerun.NewAkerunWorker(
			client, dailyBonusRepo, userRepo, newMockTransactionRepo(),
			&mockTxManager{}, newMockSystemSettingsRepo(),
			&mockPointBatchRepo{}, newMockTimeProvider(nowTime), newMockLogger(),
		)
		worker.SetRecoverySleepForTest(0)

		// === Execute ===
		worker.PollForTest()

		// === Assert ===
		assert.Equal(t, 1, *requestCount, "通常モードではAPI1回のみ")
		assert.Equal(t, nowTime, dailyBonusRepo.lastPolledAt, "lastPolledAtがnowに更新される")
	})

	t.Run("リカバリモード: 2.5時間のgapは3ウィンドウで取得", func(t *testing.T) {
		// === Setup ===
		mockResponse := createAkerunExampleResponse()
		server, requestCount := createCountingMockServer(mockResponse)
		defer server.Close()

		nowTime := time.Date(2026, 2, 17, 17, 30, 0, 0, time.UTC)

		dailyBonusRepo := newMockDailyBonusRepo()
		dailyBonusRepo.lastPolledAt = nowTime.Add(-2*time.Hour - 30*time.Minute) // 2.5時間前

		userRepo := newMockUserRepo()
		userID := uuid.New()
		userRepo.addUser(&entities.User{
			ID: userID, Username: "photosynth_taro",
			LastName: "Photosynth", FirstName: "太郎",
			Balance: 100, IsActive: true, Role: entities.RoleUser,
		})

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken: "test-token", OrganizationID: "O-test",
			BaseURL: server.URL,
		})

		worker := infraakerun.NewAkerunWorker(
			client, dailyBonusRepo, userRepo, newMockTransactionRepo(),
			&mockTxManager{}, newMockSystemSettingsRepo(),
			&mockPointBatchRepo{}, newMockTimeProvider(nowTime), newMockLogger(),
		)
		worker.SetRecoverySleepForTest(0)

		// === Execute ===
		worker.PollForTest()

		// === Assert ===
		// 2.5時間 = [0:00~1:00] + [1:00~2:00] + [2:00~2:30] = 3ウィンドウ
		assert.Equal(t, 3, *requestCount, "リカバリモードで3回API呼び出し")
		assert.Equal(t, nowTime, dailyBonusRepo.lastPolledAt, "lastPolledAtがnowに更新される")
	})

	t.Run("リカバリモード: 40分のgapは1ウィンドウで取得（nowで切る）", func(t *testing.T) {
		// === Setup ===
		mockResponse := createAkerunExampleResponse()
		server, requestCount := createCountingMockServer(mockResponse)
		defer server.Close()

		nowTime := time.Date(2026, 2, 17, 17, 40, 0, 0, time.UTC)

		dailyBonusRepo := newMockDailyBonusRepo()
		dailyBonusRepo.lastPolledAt = nowTime.Add(-40 * time.Minute) // 40分前

		userRepo := newMockUserRepo()
		userID := uuid.New()
		userRepo.addUser(&entities.User{
			ID: userID, Username: "photosynth_taro",
			LastName: "Photosynth", FirstName: "太郎",
			Balance: 100, IsActive: true, Role: entities.RoleUser,
		})

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken: "test-token", OrganizationID: "O-test",
			BaseURL: server.URL,
		})

		worker := infraakerun.NewAkerunWorker(
			client, dailyBonusRepo, userRepo, newMockTransactionRepo(),
			&mockTxManager{}, newMockSystemSettingsRepo(),
			&mockPointBatchRepo{}, newMockTimeProvider(nowTime), newMockLogger(),
		)
		worker.SetRecoverySleepForTest(0)

		// === Execute ===
		worker.PollForTest()

		// === Assert ===
		// 40分 < 1時間 → 1ウィンドウでnowに切られる
		assert.Equal(t, 1, *requestCount, "40分のgapは1回API呼び出し")
		assert.Equal(t, nowTime, dailyBonusRepo.lastPolledAt, "lastPolledAtがnowに更新される")
	})

	t.Run("リカバリモード: APIエラー時は中断し途中までのlastPolledAtが保存される", func(t *testing.T) {
		// === Setup ===
		// 2回目のリクエストでエラーを返すサーバー
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 2 {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(akerunAPIResponse{Accesses: []akerunAccessJSON{}})
		}))
		defer server.Close()

		nowTime := time.Date(2026, 2, 17, 17, 30, 0, 0, time.UTC)
		startTime := nowTime.Add(-2*time.Hour - 30*time.Minute) // 2.5時間前

		dailyBonusRepo := newMockDailyBonusRepo()
		dailyBonusRepo.lastPolledAt = startTime

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken: "test-token", OrganizationID: "O-test",
			BaseURL: server.URL,
		})

		worker := infraakerun.NewAkerunWorker(
			client, dailyBonusRepo, newMockUserRepo(), newMockTransactionRepo(),
			&mockTxManager{}, newMockSystemSettingsRepo(),
			&mockPointBatchRepo{}, newMockTimeProvider(nowTime), newMockLogger(),
		)
		worker.SetRecoverySleepForTest(0)

		// === Execute ===
		worker.PollForTest()

		// === Assert ===
		assert.Equal(t, 2, callCount, "2回目のリクエストでエラーした")
		// 1ウィンドウ目は成功したので、lastPolledAt = startTime + 1h
		expectedPolledAt := startTime.Add(1 * time.Hour)
		assert.Equal(t, expectedPolledAt, dailyBonusRepo.lastPolledAt,
			"中断時は1ウィンドウ分だけlastPolledAtが更新される")
	})

	t.Run("リカバリモード: 1時間おきにウィンドウが分割される", func(t *testing.T) {
		// === Setup ===
		// 各リクエストのdatetime_after/datetime_beforeを記録
		type requestWindow struct {
			after  string
			before string
		}
		var windows []requestWindow

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			windows = append(windows, requestWindow{
				after:  r.URL.Query().Get("datetime_after"),
				before: r.URL.Query().Get("datetime_before"),
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(akerunAPIResponse{Accesses: []akerunAccessJSON{}})
		}))
		defer server.Close()

		// 3時間のgap: [15:00~16:00] + [16:00~17:00] + [17:00~18:00] = 3ウィンドウ
		nowTime := time.Date(2026, 2, 17, 18, 0, 0, 0, time.UTC)
		startTime := time.Date(2026, 2, 17, 15, 0, 0, 0, time.UTC)

		dailyBonusRepo := newMockDailyBonusRepo()
		dailyBonusRepo.lastPolledAt = startTime

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken: "test-token", OrganizationID: "O-test",
			BaseURL: server.URL,
		})

		worker := infraakerun.NewAkerunWorker(
			client, dailyBonusRepo, newMockUserRepo(), newMockTransactionRepo(),
			&mockTxManager{}, newMockSystemSettingsRepo(),
			&mockPointBatchRepo{}, newMockTimeProvider(nowTime), newMockLogger(),
		)
		worker.SetRecoverySleepForTest(0)

		// === Execute ===
		worker.PollForTest()

		// === Assert ===
		require.Len(t, windows, 3, "3ウィンドウのリクエスト")

		// Akerun APIはJST (UTC+9) で送信される
		// UTC 15:00 = JST 00:00, UTC 16:00 = JST 01:00, etc.
		jst := time.FixedZone("JST", 9*60*60)

		// ウィンドウ1: [15:00 UTC ~ 16:00 UTC] = [00:00 JST ~ 01:00 JST]
		expectedAfter1 := startTime.In(jst).Format(time.RFC3339)
		expectedBefore1 := startTime.Add(1 * time.Hour).In(jst).Format(time.RFC3339)
		assert.Equal(t, expectedAfter1, windows[0].after, "ウィンドウ1 after")
		assert.Equal(t, expectedBefore1, windows[0].before, "ウィンドウ1 before")

		// ウィンドウ2: [16:00 UTC ~ 17:00 UTC] = [01:00 JST ~ 02:00 JST]
		expectedAfter2 := startTime.Add(1 * time.Hour).In(jst).Format(time.RFC3339)
		expectedBefore2 := startTime.Add(2 * time.Hour).In(jst).Format(time.RFC3339)
		assert.Equal(t, expectedAfter2, windows[1].after, "ウィンドウ2 after")
		assert.Equal(t, expectedBefore2, windows[1].before, "ウィンドウ2 before")

		// ウィンドウ3: [17:00 UTC ~ 18:00 UTC] = [02:00 JST ~ 03:00 JST]
		expectedAfter3 := startTime.Add(2 * time.Hour).In(jst).Format(time.RFC3339)
		expectedBefore3 := nowTime.In(jst).Format(time.RFC3339)
		assert.Equal(t, expectedAfter3, windows[2].after, "ウィンドウ3 after")
		assert.Equal(t, expectedBefore3, windows[2].before, "ウィンドウ3 before")

		// 各ウィンドウの間隔が正確に1時間であることを検証
		for i := 0; i < len(windows)-1; i++ {
			assert.Equal(t, windows[i].before, windows[i+1].after,
				fmt.Sprintf("ウィンドウ%dのbefore == ウィンドウ%dのafter", i+1, i+2))
		}
	})
}
