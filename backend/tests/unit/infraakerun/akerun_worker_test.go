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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// Mock: AkerunAccessGateway
// ========================================

type mockAkerunGateway struct {
	accesses     []entities.AccessRecord
	fetchCount   int
	isConfigured bool
	fetchErr     error
	// 各リクエストの引数を記録
	fetchCalls []fetchCall
	// コールバック（n回目で動作を変える場合）
	fetchFn func(callIdx int) ([]entities.AccessRecord, error)
}

type fetchCall struct {
	after  time.Time
	before time.Time
	limit  int
}

func newMockGateway() *mockAkerunGateway {
	return &mockAkerunGateway{
		isConfigured: true,
		fetchCalls:   make([]fetchCall, 0),
	}
}

func (m *mockAkerunGateway) FetchAccesses(ctx context.Context, after, before time.Time, limit int) ([]entities.AccessRecord, error) {
	m.fetchCalls = append(m.fetchCalls, fetchCall{after: after, before: before, limit: limit})
	m.fetchCount++
	if m.fetchFn != nil {
		return m.fetchFn(m.fetchCount - 1)
	}
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	return m.accesses, nil
}

func (m *mockAkerunGateway) IsConfigured() bool {
	return m.isConfigured
}

// ========================================
// Mock: AkerunBonusInputPort (Interactor)
// ========================================

type mockBonusInteractor struct {
	lastPolledAt     time.Time
	processedBatches [][]entities.AccessRecord
	processErr       error
}

func newMockBonusInteractor(lastPolledAt time.Time) *mockBonusInteractor {
	return &mockBonusInteractor{
		lastPolledAt:     lastPolledAt,
		processedBatches: make([][]entities.AccessRecord, 0),
	}
}

func (m *mockBonusInteractor) ProcessAccesses(ctx context.Context, accesses []entities.AccessRecord) error {
	m.processedBatches = append(m.processedBatches, accesses)
	if m.processErr != nil {
		return m.processErr
	}
	return nil
}

func (m *mockBonusInteractor) GetLastPolledAt(ctx context.Context) (time.Time, error) {
	return m.lastPolledAt, nil
}

func (m *mockBonusInteractor) UpdateLastPolledAt(ctx context.Context, t time.Time) error {
	m.lastPolledAt = t
	return nil
}

// ========================================
// Mock: Logger / TimeProvider
// ========================================

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

type mockTimeProvider struct {
	now time.Time
}

func newMockTimeProvider(t time.Time) *mockTimeProvider {
	return &mockTimeProvider{now: t}
}

func (m *mockTimeProvider) Now() time.Time { return m.now }

// ========================================
// ポーリング制御テスト
// ========================================

func TestAkerunWorker_Polling(t *testing.T) {
	t.Run("通常モード: gapが10分以内の場合はFetchAccesses1回でInteractorに委譲", func(t *testing.T) {
		nowTime := time.Date(2026, 2, 17, 17, 5, 0, 0, time.UTC)

		gateway := newMockGateway()
		gateway.accesses = []entities.AccessRecord{
			{UserName: "テスト太郎", AccessedAt: nowTime.Add(-3 * time.Minute)},
		}

		interactorMock := newMockBonusInteractor(nowTime.Add(-5 * time.Minute)) // 5分前

		worker := infraakerun.NewAkerunWorker(gateway, interactorMock, newMockTimeProvider(nowTime), newMockLogger())
		worker.SetRecoverySleepForTest(0)

		worker.PollForTest()

		// FetchAccesses は1回呼ばれる（通常モード）
		assert.Equal(t, 1, gateway.fetchCount, "通常モードではAPI1回のみ")
		// Interactor.ProcessAccesses が1回呼ばれる
		assert.Len(t, interactorMock.processedBatches, 1, "InteractorにProcessAccessesが委譲される")
		assert.Len(t, interactorMock.processedBatches[0], 1)
		// lastPolledAt が更新される
		assert.Equal(t, nowTime, interactorMock.lastPolledAt)
	})

	t.Run("リカバリモード: 2.5時間のgapは3ウィンドウで取得", func(t *testing.T) {
		nowTime := time.Date(2026, 2, 17, 17, 30, 0, 0, time.UTC)
		startTime := nowTime.Add(-2*time.Hour - 30*time.Minute)

		gateway := newMockGateway()
		interactorMock := newMockBonusInteractor(startTime)

		worker := infraakerun.NewAkerunWorker(gateway, interactorMock, newMockTimeProvider(nowTime), newMockLogger())
		worker.SetRecoverySleepForTest(0)

		worker.PollForTest()

		// 2.5時間 = 3ウィンドウ
		assert.Equal(t, 3, gateway.fetchCount, "リカバリモードで3回API呼び出し")
		assert.Equal(t, nowTime, interactorMock.lastPolledAt)
	})

	t.Run("リカバリモード: 40分のgapは1ウィンドウ（nowで切る）", func(t *testing.T) {
		nowTime := time.Date(2026, 2, 17, 17, 40, 0, 0, time.UTC)
		startTime := nowTime.Add(-40 * time.Minute)

		gateway := newMockGateway()
		interactorMock := newMockBonusInteractor(startTime)

		worker := infraakerun.NewAkerunWorker(gateway, interactorMock, newMockTimeProvider(nowTime), newMockLogger())
		worker.SetRecoverySleepForTest(0)

		worker.PollForTest()

		assert.Equal(t, 1, gateway.fetchCount, "40分のgapは1回API呼び出し")
		assert.Equal(t, nowTime, interactorMock.lastPolledAt)
	})

	t.Run("リカバリモード: APIエラー時は中断し途中までのlastPolledAtが保存される", func(t *testing.T) {
		nowTime := time.Date(2026, 2, 17, 17, 30, 0, 0, time.UTC)
		startTime := nowTime.Add(-2*time.Hour - 30*time.Minute)

		gateway := newMockGateway()
		gateway.fetchFn = func(callIdx int) ([]entities.AccessRecord, error) {
			if callIdx == 1 { // 2回目でエラー
				return nil, fmt.Errorf("API error")
			}
			return []entities.AccessRecord{}, nil
		}

		interactorMock := newMockBonusInteractor(startTime)

		worker := infraakerun.NewAkerunWorker(gateway, interactorMock, newMockTimeProvider(nowTime), newMockLogger())
		worker.SetRecoverySleepForTest(0)

		worker.PollForTest()

		// 2回目のリクエストでエラー → 中断
		assert.Equal(t, 2, gateway.fetchCount)
		// 1ウィンドウ分だけlastPolledAtが更新される
		expectedPolledAt := startTime.Add(1 * time.Hour)
		assert.Equal(t, expectedPolledAt, interactorMock.lastPolledAt)
	})

	t.Run("リカバリモード: 1時間おきにウィンドウが分割される", func(t *testing.T) {
		nowTime := time.Date(2026, 2, 17, 18, 0, 0, 0, time.UTC)
		startTime := time.Date(2026, 2, 17, 15, 0, 0, 0, time.UTC) // 3時間前

		gateway := newMockGateway()
		interactorMock := newMockBonusInteractor(startTime)

		worker := infraakerun.NewAkerunWorker(gateway, interactorMock, newMockTimeProvider(nowTime), newMockLogger())
		worker.SetRecoverySleepForTest(0)

		worker.PollForTest()

		require.Equal(t, 3, gateway.fetchCount, "3ウィンドウのリクエスト")

		// ウィンドウ境界を検証
		assert.Equal(t, startTime, gateway.fetchCalls[0].after, "ウィンドウ1 after")
		assert.Equal(t, startTime.Add(1*time.Hour), gateway.fetchCalls[0].before, "ウィンドウ1 before")

		assert.Equal(t, startTime.Add(1*time.Hour), gateway.fetchCalls[1].after, "ウィンドウ2 after")
		assert.Equal(t, startTime.Add(2*time.Hour), gateway.fetchCalls[1].before, "ウィンドウ2 before")

		assert.Equal(t, startTime.Add(2*time.Hour), gateway.fetchCalls[2].after, "ウィンドウ3 after")
		assert.Equal(t, nowTime, gateway.fetchCalls[2].before, "ウィンドウ3 before")

		// 各ウィンドウの before == 次のウィンドウの after
		for i := 0; i < len(gateway.fetchCalls)-1; i++ {
			assert.Equal(t, gateway.fetchCalls[i].before, gateway.fetchCalls[i+1].after,
				fmt.Sprintf("ウィンドウ%dのbefore == ウィンドウ%dのafter", i+1, i+2))
		}
	})

	t.Run("アクセスが0件の場合はInteractor.ProcessAccessesは呼ばれない", func(t *testing.T) {
		nowTime := time.Date(2026, 2, 17, 17, 5, 0, 0, time.UTC)

		gateway := newMockGateway()
		gateway.accesses = []entities.AccessRecord{} // 0件

		interactorMock := newMockBonusInteractor(nowTime.Add(-5 * time.Minute))

		worker := infraakerun.NewAkerunWorker(gateway, interactorMock, newMockTimeProvider(nowTime), newMockLogger())
		worker.SetRecoverySleepForTest(0)

		worker.PollForTest()

		assert.Equal(t, 1, gateway.fetchCount)
		assert.Len(t, interactorMock.processedBatches, 0, "0件の場合はProcessAccesses呼ばれない")
	})
}

// ========================================
// AkerunClient テスト（インフラ層のテスト）
// ========================================

// Akerun API モック
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

func createMockAkerunServer(response akerunAPIResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

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

		assert.Equal(t, "unlock", accesses[0].Action)
		assert.Equal(t, "Photosynth太郎", accesses[0].User.Name)
		assert.Equal(t, "A1030001", accesses[0].Akerun.ID)
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
// FetchAccesses (Gateway Adapter) テスト
// ========================================

func TestAkerunClient_FetchAccesses(t *testing.T) {
	t.Run("AccessRecordがentities.AccessRecordに変換される", func(t *testing.T) {
		mockResponse := createAkerunExampleResponse()
		server := createMockAkerunServer(mockResponse)
		defer server.Close()

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "test-token",
			OrganizationID: "O-test",
			BaseURL:        server.URL,
		})

		accesses, err := client.FetchAccesses(context.Background(),
			time.Date(2017, 7, 23, 10, 0, 0, 0, time.UTC),
			time.Date(2017, 7, 29, 19, 0, 0, 0, time.UTC),
			300,
		)

		require.NoError(t, err)
		require.Len(t, accesses, 2, "ユーザー情報ありの2件が変換される")

		assert.Equal(t, "Photosynth太郎", accesses[0].UserName)
		assert.Equal(t, time.Date(2017, 7, 24, 6, 37, 19, 0, time.UTC), accesses[0].AccessedAt)
		assert.NotEqual(t, accesses[0].ID, accesses[1].ID, "異なるアクセスは異なるIDを持つ")
	})

	t.Run("userがnullのアクセスレコードはフィルタされる", func(t *testing.T) {
		mockResponse := akerunAPIResponse{
			Accesses: []akerunAccessJSON{
				{
					ID:         json.Number("999"),
					Action:     "unlock",
					AccessedAt: "2017-07-24T10:00:00Z",
					User:       nil, // ユーザーなし
				},
				{
					ID:         json.Number("1000"),
					Action:     "unlock",
					AccessedAt: "2017-07-24T10:00:00Z",
					User: &akerunUserJSON{
						Name: "テスト太郎",
					},
				},
			},
		}
		server := createMockAkerunServer(mockResponse)
		defer server.Close()

		client := infraakerun.NewAkerunClient(&infraakerun.AkerunConfig{
			AccessToken:    "test-token",
			OrganizationID: "O-test",
			BaseURL:        server.URL,
		})

		accesses, err := client.FetchAccesses(context.Background(),
			time.Date(2017, 7, 23, 0, 0, 0, 0, time.UTC),
			time.Date(2017, 7, 29, 0, 0, 0, 0, time.UTC),
			300,
		)

		require.NoError(t, err)
		require.Len(t, accesses, 1, "userがnullのレコードはフィルタされる")
		assert.Equal(t, "テスト太郎", accesses[0].UserName)
	})
}
