package entities_test

import (
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// User Entity Tests (新機能)
// ========================================

func TestUser_UpdateProfile(t *testing.T) {
	t.Run("表示名のみ更新", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Old Name", "", "")
		originalEmail := user.Email

		err := user.UpdateProfile("New Name", "", "", "")
		require.NoError(t, err)

		assert.Equal(t, "New Name", user.DisplayName)
		assert.Equal(t, originalEmail, user.Email)
		assert.False(t, user.EmailVerified) // 新規作成時はfalse（メール未変更なので変わらない）
	})

	t.Run("メールアドレス変更時は認証リセット", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "old@example.com", "hash", "Test User", "", "")
		user.VerifyEmail() // 認証済みにする

		err := user.UpdateProfile("", "new@example.com", "", "")
		require.NoError(t, err)

		assert.Equal(t, "new@example.com", user.Email)
		assert.False(t, user.EmailVerified, "メール変更時は認証状態がリセットされる")
		assert.Nil(t, user.EmailVerifiedAt)
	})

	t.Run("表示名とメールの両方を更新", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "old@example.com", "hash", "Old Name", "", "")

		err := user.UpdateProfile("New Name", "new@example.com", "", "")
		require.NoError(t, err)

		assert.Equal(t, "New Name", user.DisplayName)
		assert.Equal(t, "new@example.com", user.Email)
		assert.False(t, user.EmailVerified)
	})

	t.Run("苗字・名前を更新", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")

		err := user.UpdateProfile("", "", "太郎", "山田")
		require.NoError(t, err)

		assert.Equal(t, "太郎", user.FirstName)
		assert.Equal(t, "山田", user.LastName)
		assert.Equal(t, "Test User", user.DisplayName) // 表示名は変更なし
	})

	t.Run("表示名・メール・苗字・名前の全てを更新", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "old@example.com", "hash", "Old Name", "", "")

		err := user.UpdateProfile("New Name", "new@example.com", "花子", "佐藤")
		require.NoError(t, err)

		assert.Equal(t, "New Name", user.DisplayName)
		assert.Equal(t, "new@example.com", user.Email)
		assert.Equal(t, "花子", user.FirstName)
		assert.Equal(t, "佐藤", user.LastName)
	})
}

func TestUser_NewUser_WithRealName(t *testing.T) {
	t.Run("苗字・名前付きでユーザー作成", func(t *testing.T) {
		user, err := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "太郎", "山田")
		require.NoError(t, err)

		assert.Equal(t, "太郎", user.FirstName)
		assert.Equal(t, "山田", user.LastName)
		assert.Equal(t, "Test User", user.DisplayName)
	})

	t.Run("苗字・名前なしでユーザー作成", func(t *testing.T) {
		user, err := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")
		require.NoError(t, err)

		assert.Equal(t, "", user.FirstName)
		assert.Equal(t, "", user.LastName)
	})
}

func TestUser_UpdateUsername(t *testing.T) {
	t.Run("ユーザー名を正常に変更", func(t *testing.T) {
		user, _ := entities.NewUser("oldusername", "test@example.com", "hash", "Test User", "", "")

		err := user.UpdateUsername("newusername")
		require.NoError(t, err)

		assert.Equal(t, "newusername", user.Username)
	})

	t.Run("空のユーザー名はエラー", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")

		err := user.UpdateUsername("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username is required")
	})

	t.Run("同じユーザー名はエラー", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")

		err := user.UpdateUsername("testuser")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "same as current")
	})
}

func TestUser_UpdateAvatar(t *testing.T) {
	t.Run("アバターURLを正常に更新", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")

		err := user.UpdateAvatar("https://example.com/avatar.jpg", entities.AvatarTypeUploaded)
		require.NoError(t, err)

		assert.NotNil(t, user.AvatarURL)
		assert.Equal(t, "https://example.com/avatar.jpg", *user.AvatarURL)
		assert.Equal(t, entities.AvatarTypeUploaded, user.AvatarType)
	})

	t.Run("無効なアバタータイプはエラー", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")

		err := user.UpdateAvatar("https://example.com/avatar.jpg", "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid avatar type")
	})
}

func TestUser_DeleteAvatar(t *testing.T) {
	t.Run("アバターを削除して自動生成に戻す", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")
		avatarURL := "https://example.com/avatar.jpg"
		user.AvatarURL = &avatarURL
		user.AvatarType = entities.AvatarTypeUploaded

		user.DeleteAvatar()

		assert.Nil(t, user.AvatarURL)
		assert.Equal(t, entities.AvatarTypeGenerated, user.AvatarType)
	})
}

func TestUser_VerifyEmail(t *testing.T) {
	t.Run("メールアドレスを認証", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")

		assert.False(t, user.EmailVerified)
		assert.Nil(t, user.EmailVerifiedAt)

		user.VerifyEmail()

		assert.True(t, user.EmailVerified)
		assert.NotNil(t, user.EmailVerifiedAt)
	})
}

func TestUser_UpdatePassword(t *testing.T) {
	t.Run("パスワードを正常に更新", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "oldhash", "Test User", "", "")

		err := user.UpdatePassword("newhash")
		require.NoError(t, err)

		assert.Equal(t, "newhash", user.PasswordHash)
	})

	t.Run("空のパスワードハッシュはエラー", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")

		err := user.UpdatePassword("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password hash is required")
	})
}

// ========================================
// ArchivedUser Tests
// ========================================

func TestUser_ToArchivedUser(t *testing.T) {
	t.Run("ユーザーをアーカイブユーザーに変換", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")
		user.Balance = 1000
		archivedBy := uuid.New()
		reason := "User requested deletion"

		archived := user.ToArchivedUser(&archivedBy, &reason)

		assert.Equal(t, user.ID, archived.ID)
		assert.Equal(t, user.Username, archived.Username)
		assert.Equal(t, user.Email, archived.Email)
		assert.Equal(t, user.Balance, archived.Balance)
		assert.Equal(t, &archivedBy, archived.ArchivedBy)
		assert.Equal(t, &reason, archived.DeletionReason)
		assert.NotZero(t, archived.ArchivedAt)
	})
}

func TestArchivedUser_RestoreToUser(t *testing.T) {
	t.Run("アーカイブユーザーをユーザーに復元", func(t *testing.T) {
		user, _ := entities.NewUser("testuser", "test@example.com", "hash", "Test User", "", "")
		user.Balance = 1000
		archivedBy := uuid.New()
		archived := user.ToArchivedUser(&archivedBy, nil)

		restoredUser := archived.RestoreToUser()

		assert.Equal(t, archived.ID, restoredUser.ID)
		assert.Equal(t, archived.Username, restoredUser.Username)
		assert.Equal(t, archived.Email, restoredUser.Email)
		assert.Equal(t, archived.Balance, restoredUser.Balance)
		assert.Equal(t, 1, restoredUser.Version, "復元時はバージョンがリセットされる")
		assert.True(t, restoredUser.IsActive)
	})
}

// ========================================
// EmailVerificationToken Tests
// ========================================

func TestNewEmailVerificationToken(t *testing.T) {
	t.Run("登録用トークンを正常に作成", func(t *testing.T) {
		email := "test@example.com"

		token, err := entities.NewEmailVerificationToken(nil, email, entities.TokenTypeRegistration)
		require.NoError(t, err)

		assert.NotZero(t, token.ID)
		assert.Nil(t, token.UserID)
		assert.Equal(t, email, token.Email)
		assert.NotEmpty(t, token.Token)
		assert.Equal(t, entities.TokenTypeRegistration, token.TokenType)
		assert.False(t, token.IsExpired())
		assert.False(t, token.IsVerified())
	})

	t.Run("メール変更用トークンを正常に作成", func(t *testing.T) {
		userID := uuid.New()
		email := "new@example.com"

		token, err := entities.NewEmailVerificationToken(&userID, email, entities.TokenTypeEmailChange)
		require.NoError(t, err)

		assert.Equal(t, &userID, token.UserID)
		assert.Equal(t, entities.TokenTypeEmailChange, token.TokenType)
	})

	t.Run("空のメールアドレスはエラー", func(t *testing.T) {
		_, err := entities.NewEmailVerificationToken(nil, "", entities.TokenTypeRegistration)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("無効なトークンタイプはエラー", func(t *testing.T) {
		_, err := entities.NewEmailVerificationToken(nil, "test@example.com", "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token type")
	})
}

func TestEmailVerificationToken_Verify(t *testing.T) {
	t.Run("トークンを正常に検証", func(t *testing.T) {
		token, _ := entities.NewEmailVerificationToken(nil, "test@example.com", entities.TokenTypeRegistration)

		err := token.Verify()
		require.NoError(t, err)

		assert.True(t, token.IsVerified())
		assert.NotNil(t, token.VerifiedAt)
	})

	t.Run("既に検証済みのトークンはエラー", func(t *testing.T) {
		token, _ := entities.NewEmailVerificationToken(nil, "test@example.com", entities.TokenTypeRegistration)
		token.Verify()

		err := token.Verify()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already verified")
	})
}

// ========================================
// UsernameChangeHistory Tests
// ========================================

func TestNewUsernameChangeHistory(t *testing.T) {
	t.Run("ユーザー名変更履歴を正常に作成", func(t *testing.T) {
		userID := uuid.New()
		changedBy := uuid.New()
		ipAddress := "192.168.1.1"

		history := entities.NewUsernameChangeHistory(userID, "oldname", "newname", &changedBy, &ipAddress)

		assert.NotZero(t, history.ID)
		assert.Equal(t, userID, history.UserID)
		assert.Equal(t, "oldname", history.OldUsername)
		assert.Equal(t, "newname", history.NewUsername)
		assert.Equal(t, &changedBy, history.ChangedBy)
		assert.Equal(t, &ipAddress, history.IPAddress)
		assert.NotZero(t, history.ChangedAt)
	})
}

// ========================================
// PasswordChangeHistory Tests
// ========================================

func TestNewPasswordChangeHistory(t *testing.T) {
	t.Run("パスワード変更履歴を正常に作成", func(t *testing.T) {
		userID := uuid.New()
		ipAddress := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		history := entities.NewPasswordChangeHistory(userID, &ipAddress, &userAgent)

		assert.NotZero(t, history.ID)
		assert.Equal(t, userID, history.UserID)
		assert.Equal(t, &ipAddress, history.IPAddress)
		assert.Equal(t, &userAgent, history.UserAgent)
		assert.NotZero(t, history.ChangedAt)
	})
}
