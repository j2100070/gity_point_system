//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/stretchr/testify/require"
)

// setupIntegrationDB は統合テスト用のDB接続を初期化
// Integration テストでは TRUNCATE でクリーンアップする（並行TX テストには TX ロールバック不可のため）
func setupIntegrationDB(t *testing.T) inframysql.DB {
	db, err := inframysql.NewPostgresDB(&inframysql.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "point_system_test",
		SSLMode:  "disable",
		Env:      "test",
	})
	require.NoError(t, err)

	// テストデータをクリーンアップ
	gormDB := db.GetDB()
	gormDB.Exec("TRUNCATE TABLE product_exchanges CASCADE")
	gormDB.Exec("TRUNCATE TABLE transactions CASCADE")
	gormDB.Exec("TRUNCATE TABLE point_batches CASCADE")
	gormDB.Exec("TRUNCATE TABLE qr_codes CASCADE")
	gormDB.Exec("TRUNCATE TABLE sessions CASCADE")
	gormDB.Exec("TRUNCATE TABLE friendships CASCADE")
	gormDB.Exec("TRUNCATE TABLE friendships_archive CASCADE")
	gormDB.Exec("TRUNCATE TABLE transfer_requests CASCADE")
	gormDB.Exec("TRUNCATE TABLE daily_bonuses CASCADE")
	gormDB.Exec("TRUNCATE TABLE products CASCADE")
	gormDB.Exec("TRUNCATE TABLE categories CASCADE")
	gormDB.Exec("TRUNCATE TABLE users CASCADE")

	return db
}

// createTestUser はテスト用ユーザーをDBに挿入
func createTestUser(t *testing.T, db inframysql.DB, username string) *entities.User {
	userDS := dsmysqlimpl.NewUserDataSource(db)
	user, err := entities.NewUser(username, username+"@example.com", "hash", "User "+username, "", "")
	require.NoError(t, err)
	require.NoError(t, userDS.Insert(context.Background(), user))
	return user
}
