//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAdmin(t *testing.T) (inputport.AdminInputPort, inframysql.DB) {
	t.Helper()
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	txManager := inframysql.NewGormTransactionManager(db.GetDB())

	admin := interactor.NewAdminInteractor(
		txManager, repos.User, repos.Transaction, repos.IdempotencyKey, repos.PointBatch, repos.Analytics, lg,
	)
	return admin, db
}

// TestAdmin_GrantPoints はポイント付与と残高反映を検証
func TestAdmin_GrantPoints(t *testing.T) {
	admin, db := setupAdmin(t)
	ctx := context.Background()

	adminUser := createTestAdminUser(t, db, "admin_grant")
	targetUser := createTestUser(t, db, "target_grant")

	resp, err := admin.GrantPoints(ctx, &inputport.GrantPointsRequest{
		AdminID:        adminUser.ID,
		UserID:         targetUser.ID,
		Amount:         500,
		Description:    "integration test grant",
		IdempotencyKey: "integ-admin-grant-001",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(500), resp.User.Balance)
	assert.Equal(t, int64(500), resp.Transaction.Amount)
}

// TestAdmin_DeductPoints はポイント減算を検証
func TestAdmin_DeductPoints(t *testing.T) {
	admin, db := setupAdmin(t)
	ctx := context.Background()

	adminUser := createTestAdminUser(t, db, "admin_deduct")
	targetUser := createTestUserWithBalance(t, db, "target_deduct", 1000)

	resp, err := admin.DeductPoints(ctx, &inputport.DeductPointsRequest{
		AdminID:        adminUser.ID,
		UserID:         targetUser.ID,
		Amount:         300,
		Description:    "integration test deduct",
		IdempotencyKey: "integ-admin-deduct-001",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(700), resp.User.Balance)
}

// TestAdmin_ListAllUsers はユーザー一覧取得を検証
func TestAdmin_ListAllUsers(t *testing.T) {
	admin, db := setupAdmin(t)
	ctx := context.Background()

	// テストユーザーを3人作成
	for i := 0; i < 3; i++ {
		createTestUser(t, db, fmt.Sprintf("admin_list_%d", i))
	}

	resp, err := admin.ListAllUsers(ctx, &inputport.ListAllUsersRequest{
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(resp.Users), 3)
}

// TestAdmin_GetAnalytics は分析データ取得を検証
func TestAdmin_GetAnalytics(t *testing.T) {
	admin, db := setupAdmin(t)
	ctx := context.Background()

	// テストユーザー作成（分析元データ）
	createTestUserWithBalance(t, db, "analytics_user", 5000)

	resp, err := admin.GetAnalytics(ctx, &inputport.GetAnalyticsRequest{
		Days: 7,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, resp.Summary)
	assert.GreaterOrEqual(t, resp.Summary.ActiveUsers, int64(1))
}

// TestAdmin_DeactivateUser はユーザー無効化を検証
func TestAdmin_DeactivateUser(t *testing.T) {
	admin, db := setupAdmin(t)
	ctx := context.Background()

	adminUser := createTestAdminUser(t, db, "admin_deact")
	targetUser := createTestUser(t, db, "target_deact")

	resp, err := admin.DeactivateUser(ctx, &inputport.DeactivateUserRequest{
		AdminID: adminUser.ID,
		UserID:  targetUser.ID,
	})
	require.NoError(t, err)
	assert.False(t, resp.User.IsActive)
}
