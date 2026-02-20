//go:build integration
// +build integration

package datasource

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
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ========================================
// グローバル変数: TestMainでコンテナ+DB接続を保持
// ========================================

var (
	testGormDB  *gorm.DB
	testCleanup func()
)

// ========================================
// TestMain: テストスイート開始時にコンテナ起動
// ========================================

func TestMain(m *testing.M) {
	fmt.Fprintln(os.Stderr, "[TestMain] ===== STARTING TestMain =====")
	ctx := context.Background()

	// マイグレーションファイルのパスを取得
	migrationDir := findMigrationDir()
	migrationFiles := getMigrationFiles(migrationDir)
	fmt.Printf("[TestMain] migration dir: %s\n", migrationDir)
	fmt.Printf("[TestMain] migration files: %d files found\n", len(migrationFiles))
	for _, f := range migrationFiles {
		fmt.Printf("  - %s\n", filepath.Base(f))
	}

	// PostgreSQLコンテナを起動
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

	// 接続文字列を取得
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		testCleanup()
		log.Fatalf("[TestMain] FATAL: failed to get connection string: %v", err)
	}
	fmt.Printf("[TestMain] connection string: %s\n", connStr)

	// GORM接続
	testGormDB, err = gorm.Open(gormPostgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		testCleanup()
		log.Fatalf("[TestMain] FATAL: failed to connect to test database: %v", err)
	}
	if testGormDB == nil {
		testCleanup()
		log.Fatalf("[TestMain] FATAL: gorm.Open returned nil DB without error")
	}
	fmt.Println("[TestMain] GORM connected successfully")

	// AutoMigrate（main.goと同じ: 追加テーブル用）
	if err := testGormDB.AutoMigrate(&dsmysqlimpl.CategoryModel{}); err != nil {
		testCleanup()
		log.Fatalf("[TestMain] FATAL: failed to auto migrate: %v", err)
	}
	fmt.Println("[TestMain] AutoMigrate completed. Running tests...")

	// テスト実行
	code := m.Run()

	// クリーンアップ
	testCleanup()
	os.Exit(code)
}

// ========================================
// testTxDB: トランザクションを inframysql.DB として扱うラッパー
// ========================================

// testTxDB は inframysql.DB インターフェースを実装するトランザクションラッパー。
// GetDB() がトランザクションを返すため、datasource の操作は全てこの TX 内で実行される。
type testTxDB struct {
	tx *gorm.DB
}

var _ inframysql.DB = (*testTxDB)(nil) // コンパイル時のインターフェース確認

func (t *testTxDB) GetDB() *gorm.DB {
	return t.tx
}

func (t *testTxDB) Close() error {
	return nil // トランザクションなので Close は不要
}

// ========================================
// setupTestTx: 各テストで使うヘルパー
// ========================================

// setupTestTx はテスト用のトランザクションを開始し、テスト終了時に自動ロールバックする。
// 返り値の inframysql.DB を datasource のコンストラクタに渡すことで、
// テストデータが自動的にクリーンアップされる。
func setupTestTx(t *testing.T) inframysql.DB {
	t.Helper()
	if testGormDB == nil {
		t.Fatal("testGormDB is nil — TestMain did not initialize the database. Is Docker running?")
	}
	tx := testGormDB.Begin()
	require.NotNil(t, tx)
	require.NoError(t, tx.Error)

	// テストトランザクション内でFK制約を無効化
	// （テストデータは毎回ロールバックされるため、データ整合性の心配は不要）
	tx.Exec("SET session_replication_role = 'replica'")

	t.Cleanup(func() {
		tx.Rollback()
	})

	return &testTxDB{tx: tx}
}

// ========================================
// テストユーザー作成ヘルパー
// ========================================

// createTestUser はテスト用ユーザーをDBに挿入
func createTestUser(t *testing.T, db inframysql.DB, username string) *entities.User {
	t.Helper()
	userDS := dsmysqlimpl.NewUserDataSource(db)
	user, err := entities.NewUser(username, username+"@example.com", "hash", "User "+username, "Test", "User")
	require.NoError(t, err)
	require.NoError(t, userDS.Insert(context.Background(), user))
	return user
}

// createTestUserWithBalanceDB はテスト用ユーザーを指定残高でDBに挿入
func createTestUserWithBalanceDB(t *testing.T, db inframysql.DB, username string, balance int64) *entities.User {
	t.Helper()
	userDS := dsmysqlimpl.NewUserDataSource(db)
	user, err := entities.NewUser(username, username+"@example.com", "hash", "User "+username, "Test", "User")
	require.NoError(t, err)
	user.Balance = balance
	require.NoError(t, userDS.Insert(context.Background(), user))
	return user
}

// ========================================
// マイグレーション関連ヘルパー
// ========================================

// findMigrationDir はマイグレーションディレクトリのパスをこのファイルからの相対パスで取得
func findMigrationDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("cannot get caller info")
	}
	// このファイル: backend/tests/unit/datasource/test_helpers.go
	// マイグレーション: backend/migrations/
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
}

// getMigrationFiles はマイグレーションディレクトリ内の .sql ファイルをソート済みで返す
func getMigrationFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read migration dir %s: %v", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files) // ファイル名順（001, 002, ...）
	return files
}

// ========================================
// 既存テスト互換ヘルパー
// ========================================

// setupFriendshipTestDB は既存テストとの互換性のため setupTestTx のエイリアス
func setupFriendshipTestDB(t *testing.T) inframysql.DB {
	return setupTestTx(t)
}

// setupTransferRequestTestDB は既存テストとの互換性のため setupTestTx のエイリアス
func setupTransferRequestTestDB(t *testing.T) inframysql.DB {
	return setupTestTx(t)
}

// setupUserSettingsTestDB は既存テストとの互換性のため setupTestTx のエイリアス
func setupUserSettingsTestDB(t *testing.T) inframysql.DB {
	return setupTestTx(t)
}

// createTestUserInDB は既存テストとの互換性のため createTestUser のエイリアス
func createTestUserInDB(t *testing.T, db inframysql.DB, username string) *entities.User {
	return createTestUser(t, db, username)
}

// setupIntegrationDB は既存テストとの互換性のため setupTestTx のエイリアス
func setupIntegrationDB(t *testing.T) inframysql.DB {
	return setupTestTx(t)
}

// setupAnalyticsTestDB は既存テストとの互換性のため setupTestTx のエイリアス
func setupAnalyticsTestDB(t *testing.T) inframysql.DB {
	return setupTestTx(t)
}

// コンテナの接続文字列を表示するデバッグヘルパー
func debugPrintConnectionInfo() {
	sqlDB, err := testGormDB.DB()
	if err != nil {
		fmt.Printf("Debug: cannot get sql.DB: %v\n", err)
		return
	}
	fmt.Printf("Debug: max open connections: %d\n", sqlDB.Stats().MaxOpenConnections)
}
