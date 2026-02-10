package entities_test

import (
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// Friendship Entity Tests
// ========================================

func TestNewFriendship(t *testing.T) {
	t.Run("正常に友達申請を作成", func(t *testing.T) {
		requester := uuid.New()
		addressee := uuid.New()

		f, err := entities.NewFriendship(requester, addressee)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, f.ID)
		assert.Equal(t, requester, f.RequesterID)
		assert.Equal(t, addressee, f.AddresseeID)
		assert.Equal(t, entities.FriendshipStatusPending, f.Status)
		assert.False(t, f.CreatedAt.IsZero())
		assert.False(t, f.UpdatedAt.IsZero())
	})

	t.Run("自分自身への友達申請はエラー", func(t *testing.T) {
		userID := uuid.New()

		_, err := entities.NewFriendship(userID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot send friend request to yourself")
	})
}

func TestFriendship_Accept(t *testing.T) {
	t.Run("pending状態の申請を承認", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())

		err := f.Accept()
		require.NoError(t, err)

		assert.Equal(t, entities.FriendshipStatusAccepted, f.Status)
	})

	t.Run("accepted状態からの承認はエラー", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())
		f.Accept()

		err := f.Accept()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in pending status")
	})

	t.Run("rejected状態からの承認はエラー", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())
		f.Reject()

		err := f.Accept()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in pending status")
	})

	t.Run("blocked状態からの承認はエラー", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())
		f.Block()

		err := f.Accept()
		assert.Error(t, err)
	})
}

func TestFriendship_Reject(t *testing.T) {
	t.Run("pending状態の申請を拒否", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())

		err := f.Reject()
		require.NoError(t, err)

		assert.Equal(t, entities.FriendshipStatusRejected, f.Status)
	})

	t.Run("accepted状態からの拒否はエラー", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())
		f.Accept()

		err := f.Reject()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in pending status")
	})
}

func TestFriendship_Block(t *testing.T) {
	t.Run("ユーザーをブロック", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())

		f.Block()

		assert.Equal(t, entities.FriendshipStatusBlocked, f.Status)
	})

	t.Run("accepted状態からブロック", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())
		f.Accept()

		f.Block()

		assert.Equal(t, entities.FriendshipStatusBlocked, f.Status)
	})
}

func TestFriendship_IsAccepted(t *testing.T) {
	t.Run("accepted状態ならtrue", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())
		f.Accept()

		assert.True(t, f.IsAccepted())
	})

	t.Run("pending状態ならfalse", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())

		assert.False(t, f.IsAccepted())
	})

	t.Run("rejected状態ならfalse", func(t *testing.T) {
		f, _ := entities.NewFriendship(uuid.New(), uuid.New())
		f.Reject()

		assert.False(t, f.IsAccepted())
	})
}
