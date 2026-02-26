//go:build integration
// +build integration

package datasource

import (
	"context"
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dspostgresimpl"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// CategoryDataSource Insert / Select Tests
// ========================================

func TestCategoryDataSource_InsertAndSelect(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewCategoryDataSource(db)

	t.Run("カテゴリを作成してIDで取得", func(t *testing.T) {
		category, err := entities.NewCategory("テスト菓子", "test_snack", "テスト用カテゴリ", 10)
		require.NoError(t, err)

		err = ds.Insert(context.Background(), category)
		require.NoError(t, err)

		retrieved, err := ds.Select(context.Background(), category.ID)
		require.NoError(t, err)
		assert.Equal(t, "テスト菓子", retrieved.Name)
		assert.Equal(t, "test_snack", retrieved.Code)
		assert.Equal(t, "テスト用カテゴリ", retrieved.Description)
		assert.Equal(t, 10, retrieved.DisplayOrder)
		assert.True(t, retrieved.IsActive)
	})

	t.Run("存在しないIDはエラー", func(t *testing.T) {
		_, err := ds.Select(context.Background(), uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "category not found")
	})
}

func TestCategoryDataSource_SelectByCode(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewCategoryDataSource(db)

	t.Run("コードでカテゴリを検索", func(t *testing.T) {
		category, _ := entities.NewCategory("テスト飲み物", "test_drink", "テストドリンク", 11)
		require.NoError(t, ds.Insert(context.Background(), category))

		retrieved, err := ds.SelectByCode(context.Background(), "test_drink")
		require.NoError(t, err)
		assert.Equal(t, category.ID, retrieved.ID)
		assert.Equal(t, "テスト飲み物", retrieved.Name)
	})

	t.Run("存在しないコードはエラー", func(t *testing.T) {
		_, err := ds.SelectByCode(context.Background(), "nonexistent")
		assert.Error(t, err)
	})
}

// ========================================
// CategoryDataSource Update Tests
// ========================================

func TestCategoryDataSource_Update(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewCategoryDataSource(db)

	t.Run("カテゴリ情報を更新", func(t *testing.T) {
		category, _ := entities.NewCategory("更新前", "update_test", "更新テスト", 1)
		require.NoError(t, ds.Insert(context.Background(), category))

		category.Name = "更新後"
		category.Description = "更新完了"
		category.DisplayOrder = 5
		err := ds.Update(context.Background(), category)
		require.NoError(t, err)

		retrieved, err := ds.Select(context.Background(), category.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", retrieved.Name)
		assert.Equal(t, "更新完了", retrieved.Description)
		assert.Equal(t, 5, retrieved.DisplayOrder)
	})
}

// ========================================
// CategoryDataSource Delete Tests
// ========================================

func TestCategoryDataSource_Delete(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewCategoryDataSource(db)

	t.Run("カテゴリを論理削除", func(t *testing.T) {
		category, _ := entities.NewCategory("削除対象", "delete_test", "削除テスト", 1)
		require.NoError(t, ds.Insert(context.Background(), category))

		err := ds.Delete(context.Background(), category.ID)
		require.NoError(t, err)

		// 論理削除後はSelectできない
		_, err = ds.Select(context.Background(), category.ID)
		assert.Error(t, err)
	})
}

// ========================================
// CategoryDataSource List Tests
// ========================================

func TestCategoryDataSource_SelectList(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewCategoryDataSource(db)

	t.Run("全カテゴリ一覧を取得", func(t *testing.T) {
		c1, _ := entities.NewCategory("カテゴリA", "cat_a", "", 2)
		c2, _ := entities.NewCategory("カテゴリB", "cat_b", "", 1)
		c3, _ := entities.NewCategory("カテゴリC", "cat_c", "", 3)
		require.NoError(t, ds.Insert(context.Background(), c1))
		require.NoError(t, ds.Insert(context.Background(), c2))
		require.NoError(t, ds.Insert(context.Background(), c3))

		// 非アクティブも含めて全件取得
		list, err := ds.SelectList(context.Background(), false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 3)

		// display_order順であることを確認
		for i := 0; i < len(list)-1; i++ {
			assert.LessOrEqual(t, list[i].DisplayOrder, list[i+1].DisplayOrder)
		}
	})

	t.Run("アクティブカテゴリのみ取得", func(t *testing.T) {
		inactive, _ := entities.NewCategory("非アクティブ", "inactive_cat", "", 99)
		require.NoError(t, ds.Insert(context.Background(), inactive))

		// 非アクティブに設定
		inactive.IsActive = false
		require.NoError(t, ds.Update(context.Background(), inactive))

		list, err := ds.SelectList(context.Background(), true)
		require.NoError(t, err)

		for _, cat := range list {
			assert.True(t, cat.IsActive, "activeOnly=trueではアクティブカテゴリのみ返す")
		}
	})
}

// ========================================
// CategoryDataSource Count / ExistsCode Tests
// ========================================

func TestCategoryDataSource_Count(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewCategoryDataSource(db)

	t.Run("カテゴリ総数を取得", func(t *testing.T) {
		c1, _ := entities.NewCategory("Count1", "count1", "", 1)
		c2, _ := entities.NewCategory("Count2", "count2", "", 2)
		require.NoError(t, ds.Insert(context.Background(), c1))
		require.NoError(t, ds.Insert(context.Background(), c2))

		count, err := ds.Count(context.Background())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})
}

func TestCategoryDataSource_ExistsCode(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dspostgresimpl.NewCategoryDataSource(db)

	t.Run("存在するコードはtrue", func(t *testing.T) {
		category, _ := entities.NewCategory("Exists", "exists_code", "", 1)
		require.NoError(t, ds.Insert(context.Background(), category))

		exists, err := ds.ExistsCode(context.Background(), "exists_code", nil)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("存在しないコードはfalse", func(t *testing.T) {
		exists, err := ds.ExistsCode(context.Background(), "nonexistent_code", nil)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("自ID除外でコード重複チェック", func(t *testing.T) {
		category, _ := entities.NewCategory("ExcludeTest", "exclude_code", "", 1)
		require.NoError(t, ds.Insert(context.Background(), category))

		// 自分自身を除外すれば重複しない
		exists, err := ds.ExistsCode(context.Background(), "exclude_code", &category.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		// 他IDなら重複する
		otherID := uuid.New()
		exists, err = ds.ExistsCode(context.Background(), "exclude_code", &otherID)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}
