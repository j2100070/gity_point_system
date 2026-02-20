//go:build integration
// +build integration

package datasource

import (
	"context"
	"testing"

	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// SystemSettingsDataSource Tests
// ========================================

func TestSystemSettingsDataSource_SetAndGet(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewSystemSettingsDataSource(db)

	t.Run("設定値を保存して取得", func(t *testing.T) {
		err := ds.SetSetting(context.Background(), "test_key", "test_value", "テスト用設定")
		require.NoError(t, err)

		value, err := ds.GetSetting(context.Background(), "test_key")
		require.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("存在しないキーは空文字を返す", func(t *testing.T) {
		value, err := ds.GetSetting(context.Background(), "nonexistent_key")
		require.NoError(t, err)
		assert.Equal(t, "", value)
	})

	t.Run("既存キーをupsertで上書き", func(t *testing.T) {
		err := ds.SetSetting(context.Background(), "upsert_key", "original", "初期値")
		require.NoError(t, err)

		err = ds.SetSetting(context.Background(), "upsert_key", "updated", "更新後")
		require.NoError(t, err)

		value, err := ds.GetSetting(context.Background(), "upsert_key")
		require.NoError(t, err)
		assert.Equal(t, "updated", value)
	})

	t.Run("複数の設定値を独立して管理", func(t *testing.T) {
		require.NoError(t, ds.SetSetting(context.Background(), "key_a", "value_a", ""))
		require.NoError(t, ds.SetSetting(context.Background(), "key_b", "value_b", ""))

		valA, err := ds.GetSetting(context.Background(), "key_a")
		require.NoError(t, err)
		assert.Equal(t, "value_a", valA)

		valB, err := ds.GetSetting(context.Background(), "key_b")
		require.NoError(t, err)
		assert.Equal(t, "value_b", valB)
	})

	t.Run("空のdescriptionでも保存可能", func(t *testing.T) {
		err := ds.SetSetting(context.Background(), "no_desc_key", "some_value", "")
		require.NoError(t, err)

		value, err := ds.GetSetting(context.Background(), "no_desc_key")
		require.NoError(t, err)
		assert.Equal(t, "some_value", value)
	})
}
