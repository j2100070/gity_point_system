package interactor_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// Mock Repositories for DailyBonusInteractor (ProcessAccesses)
// ========================================

// abMockDailyBonusRepo は DailyBonusRepository のモック（akerun bonus用）
type abMockDailyBonusRepo struct {
	bonuses      map[string]*entities.DailyBonus // key: "userID-bonusDate"
	lastPolledAt time.Time
	created      []*entities.DailyBonus
}

func newABMockDailyBonusRepo() *abMockDailyBonusRepo {
	return &abMockDailyBonusRepo{
		bonuses:      make(map[string]*entities.DailyBonus),
		lastPolledAt: time.Now().Add(-1 * time.Hour),
		created:      make([]*entities.DailyBonus, 0),
	}
}

func (m *abMockDailyBonusRepo) Create(ctx context.Context, bonus *entities.DailyBonus) error {
	key := fmt.Sprintf("%s-%s", bonus.UserID.String(), bonus.BonusDate.Format("2006-01-02"))
	if _, exists := m.bonuses[key]; exists {
		return fmt.Errorf("duplicate bonus for user %s on %s", bonus.UserID, bonus.BonusDate.Format("2006-01-02"))
	}
	m.bonuses[key] = bonus
	m.created = append(m.created, bonus)
	return nil
}

func (m *abMockDailyBonusRepo) ReadByUserAndDate(ctx context.Context, userID uuid.UUID, bonusDate time.Time) (*entities.DailyBonus, error) {
	key := fmt.Sprintf("%s-%s", userID.String(), bonusDate.Format("2006-01-02"))
	if bonus, ok := m.bonuses[key]; ok {
		return bonus, nil
	}
	return nil, nil
}

func (m *abMockDailyBonusRepo) ReadRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.DailyBonus, error) {
	var result []*entities.DailyBonus
	for _, bonus := range m.bonuses {
		if bonus.UserID == userID {
			result = append(result, bonus)
		}
	}
	return result, nil
}

func (m *abMockDailyBonusRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	for _, bonus := range m.bonuses {
		if bonus.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *abMockDailyBonusRepo) GetLastPolledAt(ctx context.Context) (time.Time, error) {
	return m.lastPolledAt, nil
}

func (m *abMockDailyBonusRepo) UpdateLastPolledAt(ctx context.Context, t time.Time) error {
	m.lastPolledAt = t
	return nil
}

func (m *abMockDailyBonusRepo) MarkAsViewed(ctx context.Context, id uuid.UUID) error {
	return nil
}

// abMockLotteryTierRepo は LotteryTierRepository のモック
type abMockLotteryTierRepo struct {
	tiers []*entities.LotteryTier
}

func newABMockLotteryTierRepo() *abMockLotteryTierRepo {
	return &abMockLotteryTierRepo{}
}

func (m *abMockLotteryTierRepo) ReadAll(ctx context.Context) ([]*entities.LotteryTier, error) {
	return m.tiers, nil
}

func (m *abMockLotteryTierRepo) ReadActive(ctx context.Context) ([]*entities.LotteryTier, error) {
	var active []*entities.LotteryTier
	for _, t := range m.tiers {
		if t.IsActive {
			active = append(active, t)
		}
	}
	return active, nil
}

func (m *abMockLotteryTierRepo) Create(ctx context.Context, tier *entities.LotteryTier) error {
	m.tiers = append(m.tiers, tier)
	return nil
}

func (m *abMockLotteryTierRepo) Update(ctx context.Context, tier *entities.LotteryTier) error {
	return nil
}

func (m *abMockLotteryTierRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *abMockLotteryTierRepo) ReplaceAll(ctx context.Context, tiers []*entities.LotteryTier) error {
	m.tiers = tiers
	return nil
}

// abMockUserRepo は UserRepository のモック
type abMockUserRepo struct {
	users          map[uuid.UUID]*entities.User
	balanceUpdates []repository.BalanceUpdate
}

func newABMockUserRepo() *abMockUserRepo {
	return &abMockUserRepo{
		users:          make(map[uuid.UUID]*entities.User),
		balanceUpdates: make([]repository.BalanceUpdate, 0),
	}
}

func (m *abMockUserRepo) Create(ctx context.Context, user *entities.User) error { return nil }
func (m *abMockUserRepo) Read(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}
func (m *abMockUserRepo) ReadByUsername(ctx context.Context, username string) (*entities.User, error) {
	return nil, nil
}
func (m *abMockUserRepo) ReadByEmail(ctx context.Context, email string) (*entities.User, error) {
	return nil, nil
}
func (m *abMockUserRepo) Update(ctx context.Context, user *entities.User) (bool, error) {
	return true, nil
}
func (m *abMockUserRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *abMockUserRepo) ReadList(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	var result []*entities.User
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}
func (m *abMockUserRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}
func (m *abMockUserRepo) ReadListWithSearch(ctx context.Context, search, sortBy, sortOrder string, offset, limit int) ([]*entities.User, error) {
	return nil, nil
}
func (m *abMockUserRepo) CountWithSearch(ctx context.Context, search string) (int64, error) {
	return 0, nil
}
func (m *abMockUserRepo) UpdateBalanceWithLock(ctx context.Context, userID uuid.UUID, amount int64, isDeduct bool) error {
	return nil
}
func (m *abMockUserRepo) UpdateBalancesWithLock(ctx context.Context, updates []repository.BalanceUpdate) error {
	m.balanceUpdates = append(m.balanceUpdates, updates...)
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
func (m *abMockUserRepo) ReadPersonalQRCode(ctx context.Context, userID uuid.UUID) (*entities.QRCode, error) {
	return nil, nil
}

func (m *abMockUserRepo) addUser(user *entities.User) {
	m.users[user.ID] = user
}

// abMockTransactionRepo は TransactionRepository のモック
type abMockTransactionRepo struct {
	transactions []*entities.Transaction
}

func newABMockTransactionRepo() *abMockTransactionRepo {
	return &abMockTransactionRepo{
		transactions: make([]*entities.Transaction, 0),
	}
}

func (m *abMockTransactionRepo) Create(ctx context.Context, tx *entities.Transaction) error {
	m.transactions = append(m.transactions, tx)
	return nil
}
func (m *abMockTransactionRepo) Read(ctx context.Context, id uuid.UUID) (*entities.Transaction, error) {
	return nil, nil
}
func (m *abMockTransactionRepo) ReadByIdempotencyKey(ctx context.Context, key string) (*entities.Transaction, error) {
	return nil, nil
}
func (m *abMockTransactionRepo) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Transaction, error) {
	return nil, nil
}
func (m *abMockTransactionRepo) ReadListAll(ctx context.Context, offset, limit int) ([]*entities.Transaction, error) {
	return nil, nil
}
func (m *abMockTransactionRepo) Update(ctx context.Context, tx *entities.Transaction) error {
	return nil
}
func (m *abMockTransactionRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *abMockTransactionRepo) ReadListAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo, sortBy, sortOrder string, offset, limit int) ([]*entities.Transaction, error) {
	return nil, nil
}
func (m *abMockTransactionRepo) CountAll(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *abMockTransactionRepo) CountAllWithFilter(ctx context.Context, transactionType, dateFrom, dateTo string) (int64, error) {
	return 0, nil
}

// abMockTxManager は TransactionManager のモック（そのまま実行）
type abMockTxManager struct{}

func (m *abMockTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// abMockSystemSettingsRepo は SystemSettingsRepository のモック
type abMockSystemSettingsRepo struct {
	settings map[string]string
}

func newABMockSystemSettingsRepo() *abMockSystemSettingsRepo {
	return &abMockSystemSettingsRepo{
		settings: make(map[string]string),
	}
}

func (m *abMockSystemSettingsRepo) GetSetting(ctx context.Context, key string) (string, error) {
	v, ok := m.settings[key]
	if !ok {
		return "", fmt.Errorf("setting not found: %s", key)
	}
	return v, nil
}

func (m *abMockSystemSettingsRepo) SetSetting(ctx context.Context, key, value, description string) error {
	m.settings[key] = value
	return nil
}

// abMockPointBatchRepo は PointBatchRepository のモック
type abMockPointBatchRepo struct{}

func (m *abMockPointBatchRepo) Create(ctx context.Context, batch *entities.PointBatch) error {
	return nil
}
func (m *abMockPointBatchRepo) ConsumePointsFIFO(ctx context.Context, userID uuid.UUID, amount int64) error {
	return nil
}
func (m *abMockPointBatchRepo) FindExpiredBatches(ctx context.Context, before time.Time, limit int) ([]*entities.PointBatch, error) {
	return nil, nil
}
func (m *abMockPointBatchRepo) MarkExpired(ctx context.Context, batchID uuid.UUID) error {
	return nil
}

// abMockLogger はテスト用ログ
type abMockLogger struct {
	infos  []string
	errors []string
}

func newABMockLogger() *abMockLogger {
	return &abMockLogger{}
}

func (m *abMockLogger) Debug(msg string, fields ...entities.Field) {}
func (m *abMockLogger) Info(msg string, fields ...entities.Field)  { m.infos = append(m.infos, msg) }
func (m *abMockLogger) Warn(msg string, fields ...entities.Field)  {}
func (m *abMockLogger) Error(msg string, fields ...entities.Field) { m.errors = append(m.errors, msg) }
func (m *abMockLogger) Fatal(msg string, fields ...entities.Field) {}

// ========================================
// ヘルパー: DailyBonusInteractor作成（ProcessAccesses テスト用）
// ========================================

type dailyBonusProcessTestDeps struct {
	dailyBonusRepo     *abMockDailyBonusRepo
	userRepo           *abMockUserRepo
	transactionRepo    *abMockTransactionRepo
	systemSettingsRepo *abMockSystemSettingsRepo
	lotteryTierRepo    *abMockLotteryTierRepo
	logger             *abMockLogger
}

func createDailyBonusInteractorForProcess() (*interactor.DailyBonusInteractor, *dailyBonusProcessTestDeps) {
	deps := &dailyBonusProcessTestDeps{
		dailyBonusRepo:     newABMockDailyBonusRepo(),
		userRepo:           newABMockUserRepo(),
		transactionRepo:    newABMockTransactionRepo(),
		systemSettingsRepo: newABMockSystemSettingsRepo(),
		lotteryTierRepo:    newABMockLotteryTierRepo(),
		logger:             newABMockLogger(),
	}

	i := interactor.NewDailyBonusInteractor(
		deps.dailyBonusRepo,
		deps.userRepo,
		deps.transactionRepo,
		&abMockTxManager{},
		deps.systemSettingsRepo,
		&abMockPointBatchRepo{},
		deps.lotteryTierRepo,
		deps.logger,
	)

	return i, deps
}

// ========================================
// テストケース: ProcessAccesses ビジネスロジック
// ========================================

func TestDailyBonusInteractor_ProcessAccesses(t *testing.T) {
	t.Run("デフォルトポイントでボーナスが付与される", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		userID := uuid.New()
		deps.userRepo.addUser(&entities.User{
			ID: userID, Username: "photosynth_taro",
			LastName: "Photosynth", FirstName: "太郎",
			Balance: 100, IsActive: true, Role: entities.RoleUser,
		})

		accesses := []entities.AccessRecord{
			{
				ID:         uuid.New(),
				UserName:   "Photosynth太郎",
				AccessedAt: time.Date(2017, 7, 24, 6, 37, 19, 0, time.UTC),
			},
		}

		err := i.ProcessAccesses(context.Background(), accesses)
		require.NoError(t, err)

		// ボーナスレコードが1件作成される
		require.Len(t, deps.dailyBonusRepo.created, 1)
		bonus := deps.dailyBonusRepo.created[0]
		assert.Equal(t, userID, bonus.UserID)
		assert.Equal(t, int64(5), bonus.BonusPoints, "デフォルト5ポイント")
		assert.Equal(t, "Photosynth太郎", bonus.AkerunUserName)
		assert.Equal(t, "通常", bonus.LotteryTierName)

		// トランザクションが1件作成される
		require.Len(t, deps.transactionRepo.transactions, 1)
		assert.Equal(t, int64(5), deps.transactionRepo.transactions[0].Amount)

		// ユーザー残高が更新される
		require.Len(t, deps.userRepo.balanceUpdates, 1)
		assert.Equal(t, int64(105), deps.userRepo.users[userID].Balance)
	})

	t.Run("管理者設定でボーナスポイントを変更した場合", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		deps.systemSettingsRepo.SetSetting(context.Background(), "akerun_bonus_points", "10", "テスト")

		userID := uuid.New()
		deps.userRepo.addUser(&entities.User{
			ID: userID, Username: "photosynth_taro",
			LastName: "Photosynth", FirstName: "太郎",
			Balance: 200, IsActive: true, Role: entities.RoleUser,
		})

		accesses := []entities.AccessRecord{
			{
				ID:         uuid.New(),
				UserName:   "Photosynth太郎",
				AccessedAt: time.Date(2017, 7, 24, 6, 37, 19, 0, time.UTC),
			},
		}

		err := i.ProcessAccesses(context.Background(), accesses)
		require.NoError(t, err)

		require.Len(t, deps.dailyBonusRepo.created, 1)
		assert.Equal(t, int64(10), deps.dailyBonusRepo.created[0].BonusPoints, "管理者設定の10ポイント")
		assert.Equal(t, int64(210), deps.userRepo.users[userID].Balance)
	})

	t.Run("マッチしないユーザーのアクセスはスキップされる", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		deps.userRepo.addUser(&entities.User{
			ID: uuid.New(), Username: "yamada",
			LastName: "山田", FirstName: "花子",
			Balance: 100, IsActive: true, Role: entities.RoleUser,
		})

		accesses := []entities.AccessRecord{
			{
				ID:         uuid.New(),
				UserName:   "Photosynth太郎",
				AccessedAt: time.Date(2017, 7, 24, 6, 37, 19, 0, time.UTC),
			},
		}

		err := i.ProcessAccesses(context.Background(), accesses)
		require.NoError(t, err)

		assert.Len(t, deps.dailyBonusRepo.created, 0, "マッチしないユーザーにはボーナスなし")
		assert.Len(t, deps.transactionRepo.transactions, 0)
	})

	t.Run("既にボーナス付与済みの場合は重複付与されない", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		userID := uuid.New()
		deps.userRepo.addUser(&entities.User{
			ID: userID, Username: "photosynth_taro",
			LastName: "Photosynth", FirstName: "太郎",
			Balance: 100, IsActive: true, Role: entities.RoleUser,
		})

		// 既にボーナスを付与済み
		jst := time.FixedZone("JST", 9*60*60)
		bonusDate := time.Date(2017, 7, 24, 0, 0, 0, 0, jst)
		existingBonus := entities.NewDailyBonus(userID, bonusDate, 5, "existing", "Photosynth太郎", nil, nil, "通常")
		deps.dailyBonusRepo.bonuses[fmt.Sprintf("%s-%s", userID.String(), bonusDate.Format("2006-01-02"))] = existingBonus

		accesses := []entities.AccessRecord{
			{
				ID:         uuid.New(),
				UserName:   "Photosynth太郎",
				AccessedAt: time.Date(2017, 7, 24, 6, 37, 19, 0, time.UTC),
			},
		}

		err := i.ProcessAccesses(context.Background(), accesses)
		require.NoError(t, err)

		assert.Len(t, deps.dailyBonusRepo.created, 0, "既に付与済みなので新規ボーナスなし")
		assert.Equal(t, int64(100), deps.userRepo.users[userID].Balance, "残高変わらず")
	})

	t.Run("UserNameが空のアクセスレコードはスキップされる", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		deps.userRepo.addUser(&entities.User{
			ID: uuid.New(), Username: "test",
			LastName: "テスト", FirstName: "太郎",
			Balance: 100, IsActive: true, Role: entities.RoleUser,
		})

		accesses := []entities.AccessRecord{
			{
				ID:         uuid.New(),
				UserName:   "", // 空のユーザー名
				AccessedAt: time.Date(2017, 7, 24, 6, 37, 19, 0, time.UTC),
			},
		}

		err := i.ProcessAccesses(context.Background(), accesses)
		require.NoError(t, err)
		assert.Len(t, deps.dailyBonusRepo.created, 0)
	})

	t.Run("同一ユーザーの同日アクセス2件でもボーナスは1回のみ", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		userID := uuid.New()
		deps.userRepo.addUser(&entities.User{
			ID: userID, Username: "photosynth_taro",
			LastName: "Photosynth", FirstName: "太郎",
			Balance: 100, IsActive: true, Role: entities.RoleUser,
		})

		accesses := []entities.AccessRecord{
			{
				ID:         uuid.New(),
				UserName:   "Photosynth太郎",
				AccessedAt: time.Date(2017, 7, 24, 6, 37, 19, 0, time.UTC),
			},
			{
				ID:         uuid.New(),
				UserName:   "Photosynth太郎",
				AccessedAt: time.Date(2017, 7, 24, 6, 40, 19, 0, time.UTC),
			},
		}

		err := i.ProcessAccesses(context.Background(), accesses)
		require.NoError(t, err)

		assert.Len(t, deps.dailyBonusRepo.created, 1, "同一ユーザー・同一日は1件のみ")
		assert.Equal(t, int64(105), deps.userRepo.users[userID].Balance)
	})
}

// ========================================
// テストケース: 抽選ティア付きProcessAccesses
// ========================================

func TestDailyBonusInteractor_LotteryMode(t *testing.T) {
	t.Run("抽選ティアがある場合は抽選結果でポイントが決定される", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		// 確率100%のティア1つだけ（必ず当たる）
		deps.lotteryTierRepo.tiers = []*entities.LotteryTier{
			{
				ID:           uuid.New(),
				Name:         "大当たり",
				Points:       100,
				Probability:  100.0,
				DisplayOrder: 1,
				IsActive:     true,
			},
		}

		userID := uuid.New()
		deps.userRepo.addUser(&entities.User{
			ID: userID, Username: "taro",
			LastName: "Photosynth", FirstName: "太郎",
			Balance: 0, IsActive: true, Role: entities.RoleUser,
		})

		accesses := []entities.AccessRecord{
			{
				ID:         uuid.New(),
				UserName:   "Photosynth太郎",
				AccessedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			},
		}

		err := i.ProcessAccesses(context.Background(), accesses)
		require.NoError(t, err)

		require.Len(t, deps.dailyBonusRepo.created, 1)
		bonus := deps.dailyBonusRepo.created[0]
		assert.Equal(t, int64(100), bonus.BonusPoints, "抽選ティアの100ポイント")
		assert.Equal(t, "大当たり", bonus.LotteryTierName)
		assert.NotNil(t, bonus.LotteryTierID, "ティアIDが設定される")

		// 残高も更新される
		assert.Equal(t, int64(100), deps.userRepo.users[userID].Balance)
	})

	t.Run("確率0%のティアのみ（ハズレ）は0ptで記録される", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		// 確率0%のティアだけ → 全てハズレ
		deps.lotteryTierRepo.tiers = []*entities.LotteryTier{
			{
				ID:           uuid.New(),
				Name:         "レアティア",
				Points:       1000,
				Probability:  0.0,
				DisplayOrder: 1,
				IsActive:     true,
			},
		}

		userID := uuid.New()
		deps.userRepo.addUser(&entities.User{
			ID: userID, Username: "taro",
			LastName: "Photosynth", FirstName: "太郎",
			Balance: 50, IsActive: true, Role: entities.RoleUser,
		})

		accesses := []entities.AccessRecord{
			{
				ID:         uuid.New(),
				UserName:   "Photosynth太郎",
				AccessedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			},
		}

		err := i.ProcessAccesses(context.Background(), accesses)
		require.NoError(t, err)

		require.Len(t, deps.dailyBonusRepo.created, 1)
		bonus := deps.dailyBonusRepo.created[0]
		assert.Equal(t, int64(0), bonus.BonusPoints, "ハズレは0pt")
		assert.Equal(t, "ハズレ", bonus.LotteryTierName)
		assert.Nil(t, bonus.LotteryTierID, "ハズレはティアIDなし")

		// 0ptなのでトランザクション・残高更新なし
		assert.Len(t, deps.transactionRepo.transactions, 0, "0ptならトランザクションなし")
		assert.Len(t, deps.userRepo.balanceUpdates, 0, "0ptなら残高更新なし")
		assert.Equal(t, int64(50), deps.userRepo.users[userID].Balance, "残高変わらず")
	})
}

// ========================================
// テストケース: GetLastPolledAt / UpdateLastPolledAt
// ========================================

func TestDailyBonusInteractor_PollingState(t *testing.T) {
	t.Run("GetLastPolledAtはリポジトリに委譲される", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		expected := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		deps.dailyBonusRepo.lastPolledAt = expected

		result, err := i.GetLastPolledAt(context.Background())
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("UpdateLastPolledAtはリポジトリに委譲される", func(t *testing.T) {
		i, deps := createDailyBonusInteractorForProcess()

		newTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
		err := i.UpdateLastPolledAt(context.Background(), newTime)
		require.NoError(t, err)
		assert.Equal(t, newTime, deps.dailyBonusRepo.lastPolledAt)
	})
}
