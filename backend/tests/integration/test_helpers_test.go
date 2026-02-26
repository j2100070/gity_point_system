//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	infrapostgres "github.com/gity/point-system/gateways/infra/infrapostgres"
	categoryRepo "github.com/gity/point-system/gateways/repository/category"
	dailyBonusRepo "github.com/gity/point-system/gateways/repository/daily_bonus"
	friendshipRepo "github.com/gity/point-system/gateways/repository/friendship"
	lotteryTierRepo "github.com/gity/point-system/gateways/repository/lottery_tier"
	pointBatchRepo "github.com/gity/point-system/gateways/repository/point_batch"
	productRepo "github.com/gity/point-system/gateways/repository/product"
	qrcodeRepo "github.com/gity/point-system/gateways/repository/qrcode"
	sessionRepo "github.com/gity/point-system/gateways/repository/session"
	systemSettingsRepo "github.com/gity/point-system/gateways/repository/system_settings"
	transactionRepo "github.com/gity/point-system/gateways/repository/transaction"
	transferRequestRepo "github.com/gity/point-system/gateways/repository/transfer_request"
	userRepo "github.com/gity/point-system/gateways/repository/user"
	userSettingsRepo "github.com/gity/point-system/gateways/repository/user_settings"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/gity/point-system/usecases/repository"
	"github.com/gity/point-system/usecases/service"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ========================================
// グローバル変数
// ========================================

var (
	testGormDB  *gorm.DB
	testCleanup func()
)

// ========================================
// TestMain: testcontainers で PostgreSQL を起動
// ========================================

func TestMain(m *testing.M) {
	fmt.Fprintln(os.Stderr, "[TestMain] ===== STARTING INTEGRATION TESTS =====")
	ctx := context.Background()

	// マイグレーションファイルの取得
	migrationDir := findMigrationDir()
	migrationFiles := getMigrationFiles(migrationDir)
	fmt.Printf("[TestMain] migration dir: %s\n", migrationDir)
	fmt.Printf("[TestMain] migration files: %d files found\n", len(migrationFiles))

	// PostgreSQL コンテナ起動
	fmt.Println("[TestMain] starting postgres container...")
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(migrationFiles...),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("[TestMain] FATAL: failed to start postgres container: %v", err)
	}
	fmt.Println("[TestMain] postgres container started successfully")

	testCleanup = func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}

	// 接続
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		testCleanup()
		log.Fatalf("[TestMain] FATAL: failed to get connection string: %v", err)
	}
	fmt.Printf("[TestMain] connection string: %s\n", connStr)

	testGormDB, err = gorm.Open(gormPostgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		testCleanup()
		log.Fatalf("[TestMain] FATAL: failed to connect: %v", err)
	}
	if testGormDB == nil {
		testCleanup()
		log.Fatalf("[TestMain] FATAL: gorm.Open returned nil")
	}
	fmt.Println("[TestMain] GORM connected successfully")

	// AutoMigrate
	if err := testGormDB.AutoMigrate(&dspostgresimpl.CategoryModel{}); err != nil {
		testCleanup()
		log.Fatalf("[TestMain] FATAL: auto migrate failed: %v", err)
	}
	fmt.Println("[TestMain] AutoMigrate completed. Running tests...")

	code := m.Run()
	testCleanup()
	os.Exit(code)
}

// ========================================
// testDBWrapper: infrapostgres.DB インターフェースを実装
// ========================================

type testDBWrapper struct {
	db *gorm.DB
}

var _ infrapostgres.DB = (*testDBWrapper)(nil)

func (w *testDBWrapper) GetDB() *gorm.DB {
	return w.db
}

func (w *testDBWrapper) Close() error {
	return nil
}

// ========================================
// setupIntegrationDB: 各テストで使うヘルパー
// ========================================

// truncatedTables は TRUNCATE 対象テーブル一覧（依存順序を考慮）
var truncatedTables = []string{
	"product_exchanges",
	"transfer_requests",
	"transactions",
	"idempotency_keys",
	"friendships",
	"sessions",
	"daily_bonuses",
	"akerun_poll_state",
	"point_batches",
	"qr_codes",
	"username_change_history",
	"password_change_history",
	"email_verification_tokens",
	"archived_users",
	"bonus_lottery_tiers",
	"products",
	"categories",
	"users",
}

// setupIntegrationDB は各テスト前にテーブルを TRUNCATE し、infrapostgres.DB を返す。
func setupIntegrationDB(t *testing.T) infrapostgres.DB {
	t.Helper()
	if testGormDB == nil {
		t.Fatal("testGormDB is nil — TestMain did not initialize the database")
	}

	// 全テーブルを TRUNCATE（CASCADE で FK 依存を解決）
	for _, table := range truncatedTables {
		if err := testGormDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			// テーブルが存在しない場合はスキップ
			t.Logf("TRUNCATE %s: %v (skipping)", table, err)
		}
	}

	// akerun_poll_state を再シード
	testGormDB.Exec("INSERT INTO akerun_poll_state (id, last_polled_at) VALUES (1, NOW()) ON CONFLICT DO NOTHING")
	testGormDB.Exec("SET default_transaction_isolation TO 'repeatable read'")

	return &testDBWrapper{db: testGormDB}
}

// ========================================
// リポジトリ一括生成ヘルパー
// ========================================

// Repos は全リポジトリをまとめた構造体
type Repos struct {
	User                  repository.UserRepository
	Session               repository.SessionRepository
	Transaction           repository.TransactionRepository
	IdempotencyKey        repository.IdempotencyKeyRepository
	Friendship            repository.FriendshipRepository
	TransferRequest       repository.TransferRequestRepository
	Product               repository.ProductRepository
	ProductExchange       repository.ProductExchangeRepository
	Category              repository.CategoryRepository
	QRCode                repository.QRCodeRepository
	DailyBonus            repository.DailyBonusRepository
	PointBatch            repository.PointBatchRepository
	SystemSettings        repository.SystemSettingsRepository
	LotteryTier           repository.LotteryTierRepository
	Analytics             repository.AnalyticsRepository
	UserSettings          repository.UserSettingsRepository
	ArchivedUser          repository.ArchivedUserRepository
	EmailVerification     repository.EmailVerificationRepository
	UsernameChangeHistory repository.UsernameChangeHistoryRepository
	PasswordChangeHistory repository.PasswordChangeHistoryRepository
}

func setupAllRepos(db infrapostgres.DB, lg entities.Logger) *Repos {
	// DataSources
	userDS := dspostgresimpl.NewUserDataSource(db)
	sessionDS := dspostgresimpl.NewSessionDataSource(db)
	transactionDS := dspostgresimpl.NewTransactionDataSource(db)
	idempotencyDS := dspostgresimpl.NewIdempotencyKeyDataSource(db)
	friendshipDS := dspostgresimpl.NewFriendshipDataSource(db)
	transferRequestDS := dspostgresimpl.NewTransferRequestDataSource(db)
	productDS := dspostgresimpl.NewProductDataSource(db)
	productExchangeDS := dspostgresimpl.NewProductExchangeDataSource(db)
	categoryDS := dspostgresimpl.NewCategoryDataSource(db)
	qrcodeDS := dspostgresimpl.NewQRCodeDataSource(db)
	dailyBonusDS := dspostgresimpl.NewDailyBonusDataSource(db)
	pointBatchDS := dspostgresimpl.NewPointBatchDataSource(db)
	systemSettingsDS := dspostgresimpl.NewSystemSettingsDataSource(db)
	lotteryTierDS := dspostgresimpl.NewLotteryTierDataSource(db)
	analyticsDS := dspostgresimpl.NewAnalyticsDataSource(db)
	archivedUserDS := dspostgresimpl.NewArchivedUserDataSource(db)
	emailVerificationDS := dspostgresimpl.NewEmailVerificationDataSource(db)
	usernameChangeHistoryDS := dspostgresimpl.NewUsernameChangeHistoryDataSource(db)
	passwordChangeHistoryDS := dspostgresimpl.NewPasswordChangeHistoryDataSource(db)

	// Repositories
	return &Repos{
		User:                  userRepo.NewUserRepository(userDS, lg),
		Session:               sessionRepo.NewSessionRepository(sessionDS, lg),
		Transaction:           transactionRepo.NewTransactionRepository(transactionDS, lg),
		IdempotencyKey:        transactionRepo.NewIdempotencyKeyRepository(idempotencyDS, lg),
		Friendship:            friendshipRepo.NewFriendshipRepository(friendshipDS, lg),
		TransferRequest:       transferRequestRepo.NewTransferRequestRepository(transferRequestDS, lg),
		Product:               productRepo.NewProductRepository(productDS, lg),
		ProductExchange:       productRepo.NewProductExchangeRepository(productExchangeDS, lg),
		Category:              categoryRepo.NewCategoryRepository(categoryDS, lg),
		QRCode:                qrcodeRepo.NewQRCodeRepository(qrcodeDS, lg),
		DailyBonus:            dailyBonusRepo.NewDailyBonusRepository(dailyBonusDS),
		PointBatch:            pointBatchRepo.NewPointBatchRepository(pointBatchDS),
		SystemSettings:        systemSettingsRepo.NewSystemSettingsRepository(systemSettingsDS),
		LotteryTier:           lotteryTierRepo.NewLotteryTierRepository(lotteryTierDS),
		Analytics:             analyticsDS,
		UserSettings:          userSettingsRepo.NewUserSettingsRepository(userDS, lg),
		ArchivedUser:          userSettingsRepo.NewArchivedUserRepository(archivedUserDS, lg),
		EmailVerification:     userSettingsRepo.NewEmailVerificationRepository(emailVerificationDS, lg),
		UsernameChangeHistory: userSettingsRepo.NewUsernameChangeHistoryRepository(usernameChangeHistoryDS, lg),
		PasswordChangeHistory: userSettingsRepo.NewPasswordChangeHistoryRepository(passwordChangeHistoryDS, lg),
	}
}

// ========================================
// インタラクター一括生成ヘルパー
// ========================================

// Interactors は全インタラクターをまとめた構造体
type Interactors struct {
	Auth               *interactor.AuthInteractor
	PointTransfer      *interactor.PointTransferInteractor
	TransferRequest    *interactor.TransferRequestInteractor
	ProductExchange    *interactor.ProductExchangeInteractor
	ProductManagement  *interactor.ProductManagementInteractor
	CategoryManagement *interactor.CategoryManagementInteractor
	Friendship         *interactor.FriendshipInteractor
	UserSettings       *interactor.UserSettingsInteractor
	UserQuery          *interactor.UserQueryInteractor
	DailyBonus         *interactor.DailyBonusInteractor
	Admin              *interactor.AdminInteractor
	QRCode             *interactor.QRCodeInteractor
}

// Services は外部サービスのモック
type Services struct {
	Password    service.PasswordService
	Email       service.EmailService
	FileStorage service.FileStorageService
}

func setupAllInteractors(repos *Repos, svcs *Services, txManager repository.TransactionManager, lg entities.Logger) *Interactors {
	// PointTransfer は他のインタラクターの依存でもある
	pointTransfer := interactor.NewPointTransferInteractor(
		txManager, repos.User, repos.Transaction, repos.IdempotencyKey, repos.Friendship, repos.PointBatch, lg,
	)

	return &Interactors{
		PointTransfer: pointTransfer,
		ProductExchange: interactor.NewProductExchangeInteractor(
			txManager, repos.Product, repos.ProductExchange, repos.User, repos.Transaction, repos.PointBatch, lg,
		),
		DailyBonus: interactor.NewDailyBonusInteractor(
			repos.DailyBonus, repos.User, repos.Transaction, txManager, repos.SystemSettings, repos.PointBatch, repos.LotteryTier, lg,
		),
	}
}

// ========================================
// テストユーザー作成ヘルパー
// ========================================

func createTestUser(t *testing.T, db infrapostgres.DB, username string) *entities.User {
	t.Helper()
	user, err := entities.NewUser(
		username,
		username+"@test.com",
		"$2a$10$hashedpassword",
		"Display "+username,
		"First",
		"Last",
	)
	require.NoError(t, err)

	gormDB := db.GetDB()
	err = gormDB.Exec(
		`INSERT INTO users (id, username, email, password_hash, display_name, first_name, last_name, balance, is_active, role, personal_qr_code, version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 0, true, 'user', 'user:' || ?::text, 1, NOW(), NOW())`,
		user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName, user.FirstName, user.LastName, user.ID,
	).Error
	require.NoError(t, err)
	return user
}

func createTestUserWithBalance(t *testing.T, db infrapostgres.DB, username string, balance int64) *entities.User {
	t.Helper()
	user := createTestUser(t, db, username)
	err := db.GetDB().Exec("UPDATE users SET balance = ? WHERE id = ?", balance, user.ID).Error
	require.NoError(t, err)
	user.Balance = balance
	return user
}

func createTestAdminUser(t *testing.T, db infrapostgres.DB, username string) *entities.User {
	t.Helper()
	user := createTestUser(t, db, username)
	err := db.GetDB().Exec("UPDATE users SET role = 'admin' WHERE id = ?", user.ID).Error
	require.NoError(t, err)
	user.Role = entities.RoleAdmin
	return user
}

// ========================================
// マイグレーションファイル取得ヘルパー
// ========================================

func findMigrationDir() string {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	return filepath.Join(testDir, "..", "..", "migrations")
}

func getMigrationFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read migration dir: %v", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files
}

// ========================================
// テスト用ロガー
// ========================================

type testLogger struct {
	t *testing.T
}

func newTestLogger(t *testing.T) entities.Logger {
	return &testLogger{t: t}
}

func (l *testLogger) Debug(msg string, fields ...entities.Field) {
	l.t.Logf("[DEBUG] %s %v", msg, fields)
}
func (l *testLogger) Info(msg string, fields ...entities.Field) {
	l.t.Logf("[INFO] %s %v", msg, fields)
}
func (l *testLogger) Warn(msg string, fields ...entities.Field) {
	l.t.Logf("[WARN] %s %v", msg, fields)
}
func (l *testLogger) Error(msg string, fields ...entities.Field) {
	l.t.Logf("[ERROR] %s %v", msg, fields)
}
func (l *testLogger) Fatal(msg string, fields ...entities.Field) {
	l.t.Fatalf("[FATAL] %s %v", msg, fields)
}
