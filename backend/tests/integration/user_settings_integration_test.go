//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	infrapostgres "github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserSettings(t *testing.T) (inputport.UserSettingsInputPort, infrapostgres.DB) {
	t.Helper()
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	txManager := infrapostgres.NewGormTransactionManager(db.GetDB())
	pwdSvc := &mockPasswordService{}
	emailSvc := &mockEmailService{}
	fileSvc := &mockFileStorageService{}

	us := interactor.NewUserSettingsInteractor(
		txManager,
		repos.User,
		repos.UserSettings,
		repos.ArchivedUser,
		repos.EmailVerification,
		repos.UsernameChangeHistory,
		repos.PasswordChangeHistory,
		fileSvc,
		pwdSvc,
		emailSvc,
		lg,
	)
	return us, db
}

// TestUserSettings_UpdateProfile はプロフィール更新を検証
func TestUserSettings_UpdateProfile(t *testing.T) {
	us, db := setupUserSettings(t)
	ctx := context.Background()

	user := createTestUser(t, db, "settings_user1")

	resp, err := us.UpdateProfile(ctx, &inputport.UpdateProfileRequest{
		UserID:      user.ID,
		DisplayName: "Updated Name",
		Email:       user.Email,
		FirstName:   "Updated",
		LastName:    "Last",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", resp.User.DisplayName)
}

// TestUserSettings_UpdateUsername はユーザー名変更を検証
func TestUserSettings_UpdateUsername(t *testing.T) {
	us, db := setupUserSettings(t)
	ctx := context.Background()

	user := createTestUser(t, db, "settings_username")

	err := us.UpdateUsername(ctx, &inputport.UpdateUsernameRequest{
		UserID:      user.ID,
		NewUsername: "new_username_integ",
	})
	require.NoError(t, err)

	// 変更後のプロフィールを確認
	profile, err := us.GetProfile(ctx, &inputport.GetProfileRequest{UserID: user.ID})
	require.NoError(t, err)
	assert.Equal(t, "new_username_integ", profile.User.Username)
}

// TestUserSettings_ChangePassword はパスワード変更を検証
func TestUserSettings_ChangePassword(t *testing.T) {
	us, db := setupUserSettings(t)
	ctx := context.Background()

	// mockPasswordService で hashed password を生成
	user := createTestUser(t, db, "settings_pwd")
	// パスワードを設定（mockPasswordService の形式に合わせる）
	db.GetDB().Exec("UPDATE users SET password_hash = ? WHERE id = ?", "$2a$10$mock_hashed_current_pass", user.ID)

	err := us.ChangePassword(ctx, &inputport.ChangePasswordRequest{
		UserID:          user.ID,
		CurrentPassword: "current_pass",
		NewPassword:     "new_pass",
	})
	require.NoError(t, err)
}

// TestUserSettings_GetProfile はプロフィール取得を検証
func TestUserSettings_GetProfile(t *testing.T) {
	us, db := setupUserSettings(t)
	ctx := context.Background()

	user := createTestUser(t, db, "settings_profile")

	resp, err := us.GetProfile(ctx, &inputport.GetProfileRequest{UserID: user.ID})
	require.NoError(t, err)
	assert.Equal(t, "settings_profile", resp.User.Username)
	assert.Equal(t, "Display settings_profile", resp.User.DisplayName)
}

// TestUserSettings_ArchiveAccount はアカウント削除（アーカイブ）を検証
func TestUserSettings_ArchiveAccount(t *testing.T) {
	us, db := setupUserSettings(t)
	ctx := context.Background()

	user := createTestUser(t, db, "settings_archive")
	db.GetDB().Exec("UPDATE users SET password_hash = ? WHERE id = ?", "$2a$10$mock_hashed_delete_pass", user.ID)

	err := us.ArchiveAccount(ctx, &inputport.ArchiveAccountRequest{
		UserID:   user.ID,
		Password: "delete_pass",
	})
	require.NoError(t, err)
}
