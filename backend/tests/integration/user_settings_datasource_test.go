// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// Test Setup
// ========================================

func setupUserSettingsTestDB(t *testing.T) inframysql.DB {
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

	// テスト用のテーブルをクリーンアップ
	db.GetDB().Exec("TRUNCATE TABLE users CASCADE")
	db.GetDB().Exec("TRUNCATE TABLE archived_users CASCADE")
	db.GetDB().Exec("TRUNCATE TABLE email_verification_tokens CASCADE")
	db.GetDB().Exec("TRUNCATE TABLE username_change_history CASCADE")
	db.GetDB().Exec("TRUNCATE TABLE password_change_history CASCADE")

	return db
}

// ========================================
// User DataSource Tests (新フィールド)
// ========================================

func TestUserDataSource_InsertWithNewFields(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewUserDataSource(db)

	t.Run("新フィールド付きでユーザーを作成", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User")
		avatarURL := "https://example.com/avatar.jpg"
		user.AvatarURL = &avatarURL
		user.AvatarType = entities.AvatarTypeUploaded
		user.VerifyEmail()

		err := ds.Insert(user)
		require.NoError(t, err)

		// 取得して検証
		retrieved, err := ds.Select(user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Username, retrieved.Username)
		assert.NotNil(t, retrieved.AvatarURL)
		assert.Equal(t, avatarURL, *retrieved.AvatarURL)
		assert.Equal(t, entities.AvatarTypeUploaded, retrieved.AvatarType)
		assert.True(t, retrieved.EmailVerified)
		assert.NotNil(t, retrieved.EmailVerifiedAt)
	})
}

func TestUserDataSource_UpdateWithNewFields(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewUserDataSource(db)

	t.Run("新フィールドを更新", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User")
		require.NoError(t, ds.Insert(user))

		// プロフィール更新
		user.UpdateProfile("New Name", "")
		success, err := ds.Update(user)
		require.NoError(t, err)
		assert.True(t, success)

		// アバター更新（DBから再取得してから更新）
		user, err = ds.Select(user.ID)
		require.NoError(t, err)
		avatarURL := "https://example.com/new-avatar.jpg"
		user.UpdateAvatar(avatarURL, entities.AvatarTypeUploaded)
		success, err = ds.Update(user)
		require.NoError(t, err)
		assert.True(t, success)

		// メール認証（DBから再取得してから更新）
		user, err = ds.Select(user.ID)
		require.NoError(t, err)
		user.VerifyEmail()
		success, err = ds.Update(user)
		require.NoError(t, err)
		assert.True(t, success)

		// 最終的な状態を検証
		retrieved, err := ds.Select(user.ID)
		require.NoError(t, err)
		assert.Equal(t, "New Name", retrieved.DisplayName)
		assert.NotNil(t, retrieved.AvatarURL)
		assert.Equal(t, avatarURL, *retrieved.AvatarURL)
		assert.True(t, retrieved.EmailVerified)
	})
}

// ========================================
// ArchivedUser DataSource Tests
// ========================================

func TestArchivedUserDataSource_Insert(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewArchivedUserDataSource(db)

	t.Run("アーカイブユーザーを作成", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User")
		user.Balance = 1000
		archivedBy := uuid.New()
		reason := "User requested deletion"

		archived := user.ToArchivedUser(&archivedBy, &reason)

		err := ds.Insert(archived)
		require.NoError(t, err)

		// 取得して検証
		retrieved, err := ds.Select(archived.ID)
		require.NoError(t, err)
		assert.Equal(t, archived.Username, retrieved.Username)
		assert.Equal(t, archived.Balance, retrieved.Balance)
		assert.Equal(t, archivedBy, *retrieved.ArchivedBy)
		assert.Equal(t, reason, *retrieved.DeletionReason)
	})
}

func TestArchivedUserDataSource_SelectByUsername(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewArchivedUserDataSource(db)

	t.Run("ユーザー名でアーカイブユーザーを検索", func(t *testing.T) {
		user, _ := entities.NewUser("archiveduser", "archived@example.com", "hash", "Archived User")
		archived := user.ToArchivedUser(nil, nil)
		require.NoError(t, ds.Insert(archived))

		retrieved, err := ds.SelectByUsername("archiveduser")
		require.NoError(t, err)
		assert.Equal(t, archived.ID, retrieved.ID)
	})
}

func TestArchivedUserDataSource_SelectList(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewArchivedUserDataSource(db)

	t.Run("アーカイブユーザー一覧を取得", func(t *testing.T) {
		// 3つのアーカイブユーザーを作成
		for i := 0; i < 3; i++ {
			user, _ := entities.NewUser(
				"user"+string(rune('0'+i)),
				"user"+string(rune('0'+i))+"@example.com",
				"hash",
				"User "+string(rune('0'+i)),
			)
			archived := user.ToArchivedUser(nil, nil)
			require.NoError(t, ds.Insert(archived))
			time.Sleep(10 * time.Millisecond) // archived_atを少しずらす
		}

		list, err := ds.SelectList(0, 10)
		require.NoError(t, err)
		assert.Len(t, list, 3)

		// 最新のものが最初に来る（archived_at DESC）
		assert.True(t, list[0].ArchivedAt.After(list[1].ArchivedAt))
	})
}

func TestArchivedUserDataSource_Count(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewArchivedUserDataSource(db)

	t.Run("アーカイブユーザー総数を取得", func(t *testing.T) {
		// 2つのアーカイブユーザーを作成
		for i := 0; i < 2; i++ {
			user, _ := entities.NewUser(
				"countuser"+string(rune('0'+i)),
				"countuser"+string(rune('0'+i))+"@example.com",
				"hash",
				"Count User",
			)
			archived := user.ToArchivedUser(nil, nil)
			require.NoError(t, ds.Insert(archived))
		}

		count, err := ds.Count()
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}

// ========================================
// EmailVerificationToken DataSource Tests
// ========================================

func TestEmailVerificationDataSource_Insert(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewEmailVerificationDataSource(db)

	t.Run("メール認証トークンを作成", func(t *testing.T) {
		token, _ := entities.NewEmailVerificationToken(nil, "test@example.com", entities.TokenTypeRegistration)

		err := ds.Insert(token)
		require.NoError(t, err)

		// トークンで検索
		retrieved, err := ds.SelectByToken(token.Token)
		require.NoError(t, err)
		assert.Equal(t, token.Email, retrieved.Email)
		assert.Equal(t, token.TokenType, retrieved.TokenType)
	})
}

func TestEmailVerificationDataSource_Update(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewEmailVerificationDataSource(db)

	t.Run("トークンを検証済みに更新", func(t *testing.T) {
		token, _ := entities.NewEmailVerificationToken(nil, "test@example.com", entities.TokenTypeRegistration)
		require.NoError(t, ds.Insert(token))

		token.Verify()
		err := ds.Update(token)
		require.NoError(t, err)

		// 取得して検証
		retrieved, err := ds.SelectByToken(token.Token)
		require.NoError(t, err)
		assert.True(t, retrieved.IsVerified())
	})
}

func TestEmailVerificationDataSource_DeleteExpired(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewEmailVerificationDataSource(db)

	t.Run("期限切れトークンを削除", func(t *testing.T) {
		// 期限切れトークンを作成
		token, _ := entities.NewEmailVerificationToken(nil, "test@example.com", entities.TokenTypeRegistration)
		token.ExpiresAt = time.Now().Add(-1 * time.Hour) // 1時間前に期限切れ
		require.NoError(t, ds.Insert(token))

		// 有効なトークンも作成
		validToken, _ := entities.NewEmailVerificationToken(nil, "valid@example.com", entities.TokenTypeRegistration)
		require.NoError(t, ds.Insert(validToken))

		// 期限切れを削除
		err := ds.DeleteExpired()
		require.NoError(t, err)

		// 期限切れトークンは削除されている
		_, err = ds.SelectByToken(token.Token)
		assert.Error(t, err)

		// 有効なトークンは残っている
		_, err = ds.SelectByToken(validToken.Token)
		require.NoError(t, err)
	})
}

func TestEmailVerificationDataSource_DeleteByUserID(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	userDS := dsmysqlimpl.NewUserDataSource(db)
	tokenDS := dsmysqlimpl.NewEmailVerificationDataSource(db)

	t.Run("ユーザーIDに紐づくトークンを削除", func(t *testing.T) {
		// ユーザーを作成
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User")
		require.NoError(t, userDS.Insert(user))

		// ユーザーに紐づくトークンを作成
		token, _ := entities.NewEmailVerificationToken(&user.ID, "test@example.com", entities.TokenTypeEmailChange)
		require.NoError(t, tokenDS.Insert(token))

		// 削除
		err := tokenDS.DeleteByUserID(user.ID)
		require.NoError(t, err)

		// トークンは削除されている
		_, err = tokenDS.SelectByToken(token.Token)
		assert.Error(t, err)
	})
}

// ========================================
// UsernameChangeHistory DataSource Tests
// ========================================

func TestUsernameChangeHistoryDataSource_Insert(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	userDS := dsmysqlimpl.NewUserDataSource(db)
	historyDS := dsmysqlimpl.NewUsernameChangeHistoryDataSource(db)

	t.Run("ユーザー名変更履歴を作成", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User")
		require.NoError(t, userDS.Insert(user))

		ipAddress := "192.168.1.1"
		history := entities.NewUsernameChangeHistory(user.ID, "testuser", "newusername", &user.ID, &ipAddress)

		err := historyDS.Insert(history)
		require.NoError(t, err)

		// 履歴を取得
		list, err := historyDS.SelectListByUserID(user.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "testuser", list[0].OldUsername)
		assert.Equal(t, "newusername", list[0].NewUsername)
	})
}

func TestUsernameChangeHistoryDataSource_CountByUserID(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	userDS := dsmysqlimpl.NewUserDataSource(db)
	historyDS := dsmysqlimpl.NewUsernameChangeHistoryDataSource(db)

	t.Run("ユーザーの変更履歴数を取得", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User")
		require.NoError(t, userDS.Insert(user))

		// 3つの履歴を作成
		for i := 0; i < 3; i++ {
			history := entities.NewUsernameChangeHistory(user.ID, "old"+string(rune('0'+i)), "new"+string(rune('0'+i)), nil, nil)
			require.NoError(t, historyDS.Insert(history))
		}

		count, err := historyDS.CountByUserID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}

// ========================================
// PasswordChangeHistory DataSource Tests
// ========================================

func TestPasswordChangeHistoryDataSource_Insert(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	userDS := dsmysqlimpl.NewUserDataSource(db)
	historyDS := dsmysqlimpl.NewPasswordChangeHistoryDataSource(db)

	t.Run("パスワード変更履歴を作成", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User")
		require.NoError(t, userDS.Insert(user))

		ipAddress := "192.168.1.1"
		userAgent := "Mozilla/5.0"
		history := entities.NewPasswordChangeHistory(user.ID, &ipAddress, &userAgent)

		err := historyDS.Insert(history)
		require.NoError(t, err)

		// 履歴を取得
		list, err := historyDS.SelectListByUserID(user.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, ipAddress, *list[0].IPAddress)
		assert.Equal(t, userAgent, *list[0].UserAgent)
	})
}

func TestPasswordChangeHistoryDataSource_CountByUserID(t *testing.T) {
	db := setupUserSettingsTestDB(t)
	defer db.Close()

	userDS := dsmysqlimpl.NewUserDataSource(db)
	historyDS := dsmysqlimpl.NewPasswordChangeHistoryDataSource(db)

	t.Run("ユーザーのパスワード変更履歴数を取得", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User")
		require.NoError(t, userDS.Insert(user))

		// 2つの履歴を作成
		for i := 0; i < 2; i++ {
			history := entities.NewPasswordChangeHistory(user.ID, nil, nil)
			require.NoError(t, historyDS.Insert(history))
			time.Sleep(10 * time.Millisecond)
		}

		count, err := historyDS.CountByUserID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}
