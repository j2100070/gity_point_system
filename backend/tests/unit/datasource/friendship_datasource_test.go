//go:build integration
// +build integration

package datasource

import (
	"context"
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/datasource/dsmysqlimpl"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// FriendshipDataSource Insert / Select Tests
// ========================================

func TestFriendshipDataSource_InsertAndSelect(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")

	t.Run("友達申請を作成して取得", func(t *testing.T) {
		friendship, _ := entities.NewFriendship(userA.ID, userB.ID)
		err := ds.Insert(ctx, friendship)
		require.NoError(t, err)

		retrieved, err := ds.Select(ctx, friendship.ID)
		require.NoError(t, err)
		assert.Equal(t, friendship.ID, retrieved.ID)
		assert.Equal(t, userA.ID, retrieved.RequesterID)
		assert.Equal(t, userB.ID, retrieved.AddresseeID)
		assert.Equal(t, entities.FriendshipStatusPending, retrieved.Status)
	})
}

func TestFriendshipDataSource_SelectByUsers(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")

	t.Run("ユーザーペアで友達関係を検索", func(t *testing.T) {
		friendship, _ := entities.NewFriendship(userA.ID, userB.ID)
		require.NoError(t, ds.Insert(ctx, friendship))

		// A→B方向で検索
		found, err := ds.SelectByUsers(ctx, userA.ID, userB.ID)
		require.NoError(t, err)
		assert.Equal(t, friendship.ID, found.ID)

		// B→A方向でも検索可能
		found2, err := ds.SelectByUsers(ctx, userB.ID, userA.ID)
		require.NoError(t, err)
		assert.Equal(t, friendship.ID, found2.ID)
	})

	t.Run("存在しないペアはエラー", func(t *testing.T) {
		_, err := ds.SelectByUsers(ctx, uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

// ========================================
// FriendshipDataSource Update Tests
// ========================================

func TestFriendshipDataSource_Update(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")

	t.Run("ステータスをacceptedに更新", func(t *testing.T) {
		friendship, _ := entities.NewFriendship(userA.ID, userB.ID)
		require.NoError(t, ds.Insert(ctx, friendship))

		friendship.Accept()
		err := ds.Update(ctx, friendship)
		require.NoError(t, err)

		retrieved, err := ds.Select(ctx, friendship.ID)
		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusAccepted, retrieved.Status)
	})
}

// ========================================
// FriendshipDataSource List Tests
// ========================================

func TestFriendshipDataSource_SelectListFriends(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")
	userC := createTestUserInDB(t, db, "user_c")

	t.Run("承認済みの友達のみ返す", func(t *testing.T) {
		// A-B: accepted
		f1, _ := entities.NewFriendship(userA.ID, userB.ID)
		require.NoError(t, ds.Insert(ctx, f1))
		f1.Accept()
		require.NoError(t, ds.Update(ctx, f1))

		// A-C: pending (未承認)
		f2, _ := entities.NewFriendship(userA.ID, userC.ID)
		require.NoError(t, ds.Insert(ctx, f2))

		friends, err := ds.SelectListFriends(ctx, userA.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, friends, 1)
		assert.Equal(t, entities.FriendshipStatusAccepted, friends[0].Status)
	})
}

func TestFriendshipDataSource_SelectListPendingRequests(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")
	userC := createTestUserInDB(t, db, "user_c")

	t.Run("受信者の保留中申請を返す", func(t *testing.T) {
		// B→A: pending
		f1, _ := entities.NewFriendship(userB.ID, userA.ID)
		require.NoError(t, ds.Insert(ctx, f1))

		// C→A: pending
		f2, _ := entities.NewFriendship(userC.ID, userA.ID)
		require.NoError(t, ds.Insert(ctx, f2))

		pending, err := ds.SelectListPendingRequests(ctx, userA.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, pending, 2)

		// 申請者の保留中申請は返さない
		pendingForB, err := ds.SelectListPendingRequests(ctx, userB.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, pendingForB, 0)
	})
}

// ========================================
// FriendshipDataSource Delete Tests
// ========================================

func TestFriendshipDataSource_Delete(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")

	t.Run("友達関係を物理削除", func(t *testing.T) {
		friendship, _ := entities.NewFriendship(userA.ID, userB.ID)
		require.NoError(t, ds.Insert(ctx, friendship))

		err := ds.Delete(ctx, friendship.ID)
		require.NoError(t, err)

		_, err = ds.Select(ctx, friendship.ID)
		assert.Error(t, err)
	})
}

// ========================================
// FriendshipDataSource ArchiveAndDelete Tests
// ========================================

func TestFriendshipDataSource_ArchiveAndDelete(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")

	t.Run("友達関係をアーカイブしてから削除", func(t *testing.T) {
		friendship, _ := entities.NewFriendship(userA.ID, userB.ID)
		friendship.Accept()
		require.NoError(t, ds.Insert(ctx, friendship))

		err := ds.ArchiveAndDelete(ctx, friendship.ID, userA.ID)
		require.NoError(t, err)

		// 元テーブルから削除されている
		_, err = ds.Select(ctx, friendship.ID)
		assert.Error(t, err)

		// アーカイブテーブルにレコードが存在する
		var count int64
		db.GetDB().Table("friendships_archive").
			Where("id = ?", friendship.ID).Count(&count)
		assert.Equal(t, int64(1), count)

		// アーカイブのフィールドを検証（archived_byが正しいか）
		var archivedByCount int64
		db.GetDB().Table("friendships_archive").
			Where("id = ? AND archived_by = ?", friendship.ID, userA.ID).
			Count(&archivedByCount)
		assert.Equal(t, int64(1), archivedByCount, "archived_byがuserA.IDと一致する")
	})

	t.Run("存在しないIDのアーカイブはエラー", func(t *testing.T) {
		err := ds.ArchiveAndDelete(ctx, uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

// ========================================
// FriendshipDataSource CheckAreFriends Tests
// ========================================

func TestFriendshipDataSource_CheckAreFriends(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")
	userC := createTestUserInDB(t, db, "user_c")

	t.Run("承認済みならtrue", func(t *testing.T) {
		friendship, _ := entities.NewFriendship(userA.ID, userB.ID)
		require.NoError(t, ds.Insert(ctx, friendship))
		friendship.Accept()
		require.NoError(t, ds.Update(ctx, friendship))

		areFriends, err := ds.CheckAreFriends(ctx, userA.ID, userB.ID)
		require.NoError(t, err)
		assert.True(t, areFriends)

		// 逆方向でもtrue
		areFriends2, err := ds.CheckAreFriends(ctx, userB.ID, userA.ID)
		require.NoError(t, err)
		assert.True(t, areFriends2)
	})

	t.Run("pending状態ならfalse", func(t *testing.T) {
		friendship, _ := entities.NewFriendship(userA.ID, userC.ID)
		require.NoError(t, ds.Insert(ctx, friendship))

		areFriends, err := ds.CheckAreFriends(ctx, userA.ID, userC.ID)
		require.NoError(t, err)
		assert.False(t, areFriends)
	})

	t.Run("関係が存在しなければfalse", func(t *testing.T) {
		areFriends, err := ds.CheckAreFriends(ctx, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.False(t, areFriends)
	})
}

// ========================================
// Full Flow Integration Test
// ========================================

func TestFriendshipDataSource_FullFlow(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")

	t.Run("申請→承認→解散→アーカイブ確認→再申請のフルフロー", func(t *testing.T) {
		// 1. 友達申請
		friendship, _ := entities.NewFriendship(userA.ID, userB.ID)
		require.NoError(t, ds.Insert(ctx, friendship))
		assert.Equal(t, entities.FriendshipStatusPending, friendship.Status)

		// 2. 保留中の申請一覧に表示される
		pending, err := ds.SelectListPendingRequests(ctx, userB.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, pending, 1)

		// 3. 承認
		friendship.Accept()
		require.NoError(t, ds.Update(ctx, friendship))

		// 4. 友達一覧に表示される
		friends, err := ds.SelectListFriends(ctx, userA.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, friends, 1)

		// 5. CheckAreFriendsがtrue
		areFriends, err := ds.CheckAreFriends(ctx, userA.ID, userB.ID)
		require.NoError(t, err)
		assert.True(t, areFriends)

		// 6. フレンド解散（アーカイブ）
		err = ds.ArchiveAndDelete(ctx, friendship.ID, userA.ID)
		require.NoError(t, err)

		// 7. 友達一覧から消える
		friends, err = ds.SelectListFriends(ctx, userA.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, friends, 0)

		// 8. CheckAreFriendsがfalse
		areFriends, err = ds.CheckAreFriends(ctx, userA.ID, userB.ID)
		require.NoError(t, err)
		assert.False(t, areFriends)

		// 9. アーカイブテーブルにレコードが存在
		var archiveCount int64
		db.GetDB().Table("friendships_archive").
			Where("requester_id = ? AND addressee_id = ?", userA.ID, userB.ID).
			Count(&archiveCount)
		assert.Equal(t, int64(1), archiveCount)

		// 10. 再申請が可能
		newFriendship, _ := entities.NewFriendship(userA.ID, userB.ID)
		err = ds.Insert(ctx, newFriendship)
		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusPending, newFriendship.Status)
	})
}

// ========================================
// Rejected → Re-request Flow
// ========================================

func TestFriendshipDataSource_RejectAndReRequest(t *testing.T) {
	db := setupFriendshipTestDB(t)
	ctx := context.Background()

	ds := dsmysqlimpl.NewFriendshipDataSource(db)
	userA := createTestUserInDB(t, db, "user_a")
	userB := createTestUserInDB(t, db, "user_b")

	t.Run("申請→拒否→ステータス更新で再申請", func(t *testing.T) {
		// 1. 友達申請
		friendship, _ := entities.NewFriendship(userA.ID, userB.ID)
		require.NoError(t, ds.Insert(ctx, friendship))

		// 2. 拒否
		friendship.Reject()
		require.NoError(t, ds.Update(ctx, friendship))

		retrieved, err := ds.Select(ctx, friendship.ID)
		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusRejected, retrieved.Status)

		// 3. 既存レコードのステータスをpendingに更新して再申請
		retrieved.Status = entities.FriendshipStatusPending
		require.NoError(t, ds.Update(ctx, retrieved))

		updated, err := ds.Select(ctx, friendship.ID)
		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusPending, updated.Status)
	})
}
