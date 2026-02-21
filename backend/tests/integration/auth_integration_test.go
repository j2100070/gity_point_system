//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuth(t *testing.T) (inputport.AuthInputPort, inframysql.DB) {
	t.Helper()
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	pwdSvc := &mockPasswordService{}

	auth := interactor.NewAuthInteractor(repos.User, repos.Session, pwdSvc, lg)
	return auth, db
}

// TestAuth_RegisterAndLogin はユーザー登録→ログインの統合フローを検証
func TestAuth_RegisterAndLogin(t *testing.T) {
	auth, _ := setupAuth(t)
	ctx := context.Background()

	// 1. ユーザー登録
	regResp, err := auth.Register(ctx, &inputport.RegisterRequest{
		Username:    "integ_user1",
		Email:       "integ_user1@test.com",
		Password:    "password123",
		DisplayName: "Integration User 1",
		FirstName:   "Test",
		LastName:    "User",
	})
	require.NoError(t, err)
	require.NotNil(t, regResp)
	assert.Equal(t, "integ_user1", regResp.User.Username)
	assert.NotEmpty(t, regResp.Session.SessionToken)

	// 2. 同じユーザー名でログイン
	loginResp, err := auth.Login(ctx, &inputport.LoginRequest{
		Username:  "integ_user1",
		Password:  "password123",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	})
	require.NoError(t, err)
	require.NotNil(t, loginResp)
	assert.Equal(t, regResp.User.ID, loginResp.User.ID)
	assert.NotEmpty(t, loginResp.Session.SessionToken)
}

// TestAuth_LoginInvalidPassword はパスワード不一致でエラーになることを検証
func TestAuth_LoginInvalidPassword(t *testing.T) {
	auth, _ := setupAuth(t)
	ctx := context.Background()

	// ユーザー登録
	_, err := auth.Register(ctx, &inputport.RegisterRequest{
		Username:    "integ_user2",
		Email:       "integ_user2@test.com",
		Password:    "correct_password",
		DisplayName: "Integration User 2",
		FirstName:   "Test",
		LastName:    "User",
	})
	require.NoError(t, err)

	// 間違ったパスワードでログイン
	_, err = auth.Login(ctx, &inputport.LoginRequest{
		Username: "integ_user2",
		Password: "wrong_password",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid username or password")
}

// TestAuth_Logout はログアウト→セッション削除を検証
func TestAuth_Logout(t *testing.T) {
	auth, _ := setupAuth(t)
	ctx := context.Background()

	// ユーザー登録
	regResp, err := auth.Register(ctx, &inputport.RegisterRequest{
		Username:    "integ_logout_user",
		Email:       "integ_logout@test.com",
		Password:    "password123",
		DisplayName: "Logout User",
		FirstName:   "Test",
		LastName:    "User",
	})
	require.NoError(t, err)

	// ログアウト
	err = auth.Logout(ctx, &inputport.LogoutRequest{
		UserID: regResp.User.ID,
	})
	assert.NoError(t, err)

	// セッションは無効
	_, err = auth.ValidateSession(ctx, regResp.Session.SessionToken)
	assert.Error(t, err)
}

// TestAuth_ValidateSession はセッション検証を検証
func TestAuth_ValidateSession(t *testing.T) {
	auth, _ := setupAuth(t)
	ctx := context.Background()

	// 登録してセッション取得
	regResp, err := auth.Register(ctx, &inputport.RegisterRequest{
		Username:    "integ_session_user",
		Email:       "integ_session@test.com",
		Password:    "password123",
		DisplayName: "Session User",
		FirstName:   "Test",
		LastName:    "User",
	})
	require.NoError(t, err)

	// セッション検証（成功）
	session, err := auth.ValidateSession(ctx, regResp.Session.SessionToken)
	require.NoError(t, err)
	assert.Equal(t, regResp.User.ID, session.UserID)
}

// TestAuth_GetCurrentUser は登録ユーザー情報の取得を検証
func TestAuth_GetCurrentUser(t *testing.T) {
	auth, _ := setupAuth(t)
	ctx := context.Background()

	// 登録
	regResp, err := auth.Register(ctx, &inputport.RegisterRequest{
		Username:    "integ_current_user",
		Email:       "integ_current@test.com",
		Password:    "password123",
		DisplayName: "Current User",
		FirstName:   "Test",
		LastName:    "User",
	})
	require.NoError(t, err)

	// ユーザー情報取得
	resp, err := auth.GetCurrentUser(ctx, &inputport.GetCurrentUserRequest{
		UserID: regResp.User.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "integ_current_user", resp.User.Username)
	assert.Equal(t, "Current User", resp.User.DisplayName)
}

// TestAuth_DuplicateUsername は重複ユーザー名でエラーになることを検証
func TestAuth_DuplicateUsername(t *testing.T) {
	auth, _ := setupAuth(t)
	ctx := context.Background()

	req := &inputport.RegisterRequest{
		Username:    "integ_dup_user",
		Email:       "integ_dup@test.com",
		Password:    "password123",
		DisplayName: "Dup User",
		FirstName:   "Test",
		LastName:    "User",
	}

	// 1回目: 成功
	_, err := auth.Register(ctx, req)
	require.NoError(t, err)

	// 2回目: 重複エラー
	req.Email = "integ_dup2@test.com"
	_, err = auth.Register(ctx, req)
	assert.Error(t, err)
}
