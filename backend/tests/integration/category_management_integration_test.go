//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCategoryManagement_CRUD はカテゴリの作成・取得・更新・削除を検証
func TestCategoryManagement_CRUD(t *testing.T) {
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	catMgmt := interactor.NewCategoryManagementInteractor(repos.Category, lg)
	ctx := context.Background()

	// 作成
	createResp, err := catMgmt.CreateCategory(ctx, &inputport.CreateCategoryRequest{
		Name: "テストカテゴリ",
		Code: "integ_test_cat",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	assert.Equal(t, "テストカテゴリ", createResp.Category.Name)

	catID := createResp.Category.ID

	// 更新
	updateResp, err := catMgmt.UpdateCategory(ctx, &inputport.UpdateCategoryRequest{
		CategoryID:  catID,
		Name:        "更新済みカテゴリ",
		Description: "更新された説明",
	})
	require.NoError(t, err)
	assert.Equal(t, "更新済みカテゴリ", updateResp.Category.Name)

	// 一覧
	listResp, err := catMgmt.GetCategoryList(ctx, &inputport.GetCategoryListRequest{
		ActiveOnly: false,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(listResp.Categories), 1)

	// 削除
	err = catMgmt.DeleteCategory(ctx, &inputport.DeleteCategoryRequest{
		CategoryID: catID,
	})
	assert.NoError(t, err)
}

// TestCategoryManagement_DuplicateCode は重複カテゴリコードでエラーを検証
func TestCategoryManagement_DuplicateCode(t *testing.T) {
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)
	catMgmt := interactor.NewCategoryManagementInteractor(repos.Category, lg)
	ctx := context.Background()

	// 1回目
	_, err := catMgmt.CreateCategory(ctx, &inputport.CreateCategoryRequest{
		Name: "カテゴリA",
		Code: "integ_dup_code",
	})
	require.NoError(t, err)

	// 2回目（同一コード）
	_, err = catMgmt.CreateCategory(ctx, &inputport.CreateCategoryRequest{
		Name: "カテゴリB",
		Code: "integ_dup_code",
	})
	assert.Error(t, err)
}
