//go:build integration
// +build integration

package datasource

import (
	"context"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// SessionDataSource Insert / Select Tests
// ========================================

func TestSessionDataSource_InsertAndSelectByToken(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewSessionDataSource(db)
	user := createTestUser(t, db, "session_user")

	t.Run("セッションを作成してトークンで取得", func(t *testing.T) {
		session, err := entities.NewSession(user.ID, "127.0.0.1", "Go-Test-Agent")
		require.NoError(t, err)

		err = ds.Insert(context.Background(), session)
		require.NoError(t, err)

		retrieved, err := ds.SelectByToken(context.Background(), session.SessionToken)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, user.ID, retrieved.UserID)
		assert.Equal(t, session.SessionToken, retrieved.SessionToken)
		assert.Equal(t, session.CSRFToken, retrieved.CSRFToken)
		assert.Equal(t, "127.0.0.1", retrieved.IPAddress)
		assert.Equal(t, "Go-Test-Agent", retrieved.UserAgent)
	})

	t.Run("存在しないトークンはエラー", func(t *testing.T) {
		_, err := ds.SelectByToken(context.Background(), "nonexistent_token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})
}

// ========================================
// SessionDataSource Update Tests
// ========================================

func TestSessionDataSource_Update(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewSessionDataSource(db)
	user := createTestUser(t, db, "session_update_user")

	t.Run("セッション有効期限を更新", func(t *testing.T) {
		session, err := entities.NewSession(user.ID, "127.0.0.1", "Go-Test-Agent")
		require.NoError(t, err)
		require.NoError(t, ds.Insert(context.Background(), session))

		// 有効期限を変更
		newExpiry := time.Now().Add(48 * time.Hour)
		session.ExpiresAt = newExpiry
		err = ds.Update(context.Background(), session)
		require.NoError(t, err)

		retrieved, err := ds.SelectByToken(context.Background(), session.SessionToken)
		require.NoError(t, err)
		assert.WithinDuration(t, newExpiry, retrieved.ExpiresAt, time.Second)
	})
}

// ========================================
// SessionDataSource Delete Tests
// ========================================

func TestSessionDataSource_Delete(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewSessionDataSource(db)
	user := createTestUser(t, db, "session_del_user")

	t.Run("IDでセッションを削除", func(t *testing.T) {
		session, err := entities.NewSession(user.ID, "127.0.0.1", "Go-Test-Agent")
		require.NoError(t, err)
		require.NoError(t, ds.Insert(context.Background(), session))

		err = ds.Delete(context.Background(), session.ID)
		require.NoError(t, err)

		_, err = ds.SelectByToken(context.Background(), session.SessionToken)
		assert.Error(t, err)
	})
}

func TestSessionDataSource_DeleteByUserID(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewSessionDataSource(db)
	user := createTestUser(t, db, "session_del_all_user")

	t.Run("ユーザーの全セッションを削除", func(t *testing.T) {
		// 複数セッション作成
		s1, _ := entities.NewSession(user.ID, "127.0.0.1", "Agent-1")
		s2, _ := entities.NewSession(user.ID, "127.0.0.2", "Agent-2")
		require.NoError(t, ds.Insert(context.Background(), s1))
		require.NoError(t, ds.Insert(context.Background(), s2))

		err := ds.DeleteByUserID(context.Background(), user.ID)
		require.NoError(t, err)

		_, err = ds.SelectByToken(context.Background(), s1.SessionToken)
		assert.Error(t, err)
		_, err = ds.SelectByToken(context.Background(), s2.SessionToken)
		assert.Error(t, err)
	})
}

func TestSessionDataSource_DeleteExpired(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewSessionDataSource(db)
	user := createTestUser(t, db, "session_expired_user")

	t.Run("期限切れセッションを一括削除", func(t *testing.T) {
		// 期限切れセッション
		expired, _ := entities.NewSession(user.ID, "127.0.0.1", "Expired-Agent")
		expired.ExpiresAt = time.Now().Add(-1 * time.Hour) // 1時間前に期限切れ
		require.NoError(t, ds.Insert(context.Background(), expired))

		// 有効なセッション
		valid, _ := entities.NewSession(user.ID, "127.0.0.2", "Valid-Agent")
		require.NoError(t, ds.Insert(context.Background(), valid))

		err := ds.DeleteExpired(context.Background())
		require.NoError(t, err)

		// 期限切れセッションが削除されている
		_, err = ds.SelectByToken(context.Background(), expired.SessionToken)
		assert.Error(t, err)

		// 有効なセッションは残っている
		retrieved, err := ds.SelectByToken(context.Background(), valid.SessionToken)
		require.NoError(t, err)
		assert.Equal(t, valid.ID, retrieved.ID)
	})
}

// ========================================
// Session Full Flow Test
// ========================================

func TestSessionDataSource_FullFlow(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewSessionDataSource(db)
	user := createTestUser(t, db, "session_flow_user")

	t.Run("セッション作成→更新→削除のフルフロー", func(t *testing.T) {
		// 1. セッション作成
		session, err := entities.NewSession(user.ID, "192.168.1.1", "Mozilla/5.0")
		require.NoError(t, err)
		require.NoError(t, ds.Insert(context.Background(), session))

		// 2. トークンで取得
		retrieved, err := ds.SelectByToken(context.Background(), session.SessionToken)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.UserID)

		// 3. Refresh (有効期限延長)
		retrieved.ExpiresAt = time.Now().Add(72 * time.Hour)
		require.NoError(t, ds.Update(context.Background(), retrieved))

		// 4. 削除
		require.NoError(t, ds.Delete(context.Background(), session.ID))

		// 5. 取得できないことを確認
		_, err = ds.SelectByToken(context.Background(), session.SessionToken)
		assert.Error(t, err)
	})
}
