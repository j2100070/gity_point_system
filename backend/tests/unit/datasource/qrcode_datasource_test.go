//go:build integration
// +build integration

package datasource

import (
	"context"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// QRCodeDataSource Insert / Select Tests
// ========================================

func TestQRCodeDataSource_InsertAndSelectByCode(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewQRCodeDataSource(db)
	user := createTestUser(t, db, "qr_user")

	t.Run("QRコードを作成してコードで取得", func(t *testing.T) {
		qrCode, err := entities.NewReceiveQRCode(user.ID, nil) // 金額指定なし
		require.NoError(t, err)

		err = ds.Insert(context.Background(), qrCode)
		require.NoError(t, err)

		retrieved, err := ds.SelectByCode(context.Background(), qrCode.Code)
		require.NoError(t, err)
		assert.Equal(t, qrCode.ID, retrieved.ID)
		assert.Equal(t, user.ID, retrieved.UserID)
		assert.Equal(t, entities.QRCodeTypeReceive, retrieved.QRType)
		assert.Nil(t, retrieved.Amount)
		assert.Nil(t, retrieved.UsedAt)
	})

	t.Run("金額指定ありのQRコードを作成", func(t *testing.T) {
		amount := int64(500)
		qrCode, err := entities.NewReceiveQRCode(user.ID, &amount)
		require.NoError(t, err)

		err = ds.Insert(context.Background(), qrCode)
		require.NoError(t, err)

		retrieved, err := ds.SelectByCode(context.Background(), qrCode.Code)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.Amount)
		assert.Equal(t, int64(500), *retrieved.Amount)
	})

	t.Run("送信用QRコードを作成", func(t *testing.T) {
		amount := int64(1000)
		qrCode, err := entities.NewSendQRCode(user.ID, amount)
		require.NoError(t, err)

		err = ds.Insert(context.Background(), qrCode)
		require.NoError(t, err)

		retrieved, err := ds.SelectByCode(context.Background(), qrCode.Code)
		require.NoError(t, err)
		assert.Equal(t, entities.QRCodeTypeSend, retrieved.QRType)
		assert.Equal(t, int64(1000), *retrieved.Amount)
	})

	t.Run("存在しないコードはエラー", func(t *testing.T) {
		_, err := ds.SelectByCode(context.Background(), "nonexistent_code")
		assert.Error(t, err)
	})
}

func TestQRCodeDataSource_SelectByID(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewQRCodeDataSource(db)
	user := createTestUser(t, db, "qr_id_user")

	t.Run("IDでQRコードを取得", func(t *testing.T) {
		qrCode, err := entities.NewReceiveQRCode(user.ID, nil)
		require.NoError(t, err)
		require.NoError(t, ds.Insert(context.Background(), qrCode))

		retrieved, err := ds.Select(context.Background(), qrCode.ID)
		require.NoError(t, err)
		assert.Equal(t, qrCode.Code, retrieved.Code)
	})

	t.Run("存在しないIDはエラー", func(t *testing.T) {
		_, err := ds.Select(context.Background(), uuid.New())
		assert.Error(t, err)
	})
}

// ========================================
// QRCodeDataSource List Tests
// ========================================

func TestQRCodeDataSource_SelectListByUserID(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewQRCodeDataSource(db)
	user := createTestUser(t, db, "qr_list_user")
	otherUser := createTestUser(t, db, "qr_other_user")

	t.Run("ユーザーのQRコード一覧を取得", func(t *testing.T) {
		// 自分のQRコード3つ
		for i := 0; i < 3; i++ {
			qr, _ := entities.NewReceiveQRCode(user.ID, nil)
			require.NoError(t, ds.Insert(context.Background(), qr))
			time.Sleep(10 * time.Millisecond) // created_at順序保証
		}

		// 他人のQRコード
		otherQR, _ := entities.NewReceiveQRCode(otherUser.ID, nil)
		require.NoError(t, ds.Insert(context.Background(), otherQR))

		list, err := ds.SelectListByUserID(context.Background(), user.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, list, 3)

		// 作成日降順であることを確認
		for i := 0; i < len(list)-1; i++ {
			assert.True(t, list[i].CreatedAt.After(list[i+1].CreatedAt) || list[i].CreatedAt.Equal(list[i+1].CreatedAt))
		}
	})

	t.Run("ページネーションが機能する", func(t *testing.T) {
		list, err := ds.SelectListByUserID(context.Background(), user.ID, 0, 2)
		require.NoError(t, err)
		assert.Len(t, list, 2)

		list2, err := ds.SelectListByUserID(context.Background(), user.ID, 2, 2)
		require.NoError(t, err)
		assert.Len(t, list2, 1)
	})
}

// ========================================
// QRCodeDataSource Update Tests
// ========================================

func TestQRCodeDataSource_Update(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewQRCodeDataSource(db)
	user := createTestUser(t, db, "qr_update_user")
	scanner := createTestUser(t, db, "qr_scanner_user")

	t.Run("QRコード使用状態を更新", func(t *testing.T) {
		qrCode, err := entities.NewReceiveQRCode(user.ID, nil)
		require.NoError(t, err)
		require.NoError(t, ds.Insert(context.Background(), qrCode))

		// 使用済みにマーキング
		now := time.Now()
		qrCode.UsedAt = &now
		qrCode.UsedByUserID = &scanner.ID

		err = ds.Update(context.Background(), qrCode)
		require.NoError(t, err)

		retrieved, err := ds.SelectByCode(context.Background(), qrCode.Code)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.UsedAt)
		assert.NotNil(t, retrieved.UsedByUserID)
		assert.Equal(t, scanner.ID, *retrieved.UsedByUserID)
	})
}

// ========================================
// QRCodeDataSource DeleteExpired Tests
// ========================================

func TestQRCodeDataSource_DeleteExpired(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	ds := dsmysqlimpl.NewQRCodeDataSource(db)
	user := createTestUser(t, db, "qr_expire_user")

	t.Run("期限切れQRコードを削除", func(t *testing.T) {
		// 期限切れのQRコード
		expired, _ := entities.NewReceiveQRCode(user.ID, nil)
		expired.ExpiresAt = time.Now().Add(-1 * time.Hour)
		require.NoError(t, ds.Insert(context.Background(), expired))

		// 有効なQRコード
		valid, _ := entities.NewReceiveQRCode(user.ID, nil)
		require.NoError(t, ds.Insert(context.Background(), valid))

		err := ds.DeleteExpired(context.Background())
		require.NoError(t, err)

		// 期限切れは削除
		_, err = ds.SelectByCode(context.Background(), expired.Code)
		assert.Error(t, err)

		// 有効なものは残っている
		retrieved, err := ds.SelectByCode(context.Background(), valid.Code)
		require.NoError(t, err)
		assert.Equal(t, valid.ID, retrieved.ID)
	})
}
