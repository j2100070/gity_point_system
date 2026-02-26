//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	infrapostgres "github.com/gity/point-system/gateways/infra/infrapostgres"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFriendship(t *testing.T) (inputport.FriendshipInputPort, infrapostgres.DB) {
	t.Helper()
	db := setupIntegrationDB(t)
	lg := newTestLogger(t)
	repos := setupAllRepos(db, lg)

	friendship := interactor.NewFriendshipInteractor(repos.Friendship, repos.User, lg)
	return friendship, db
}

// TestFriendship_SendAndAccept は友達申請→承認の統合フローを検証
func TestFriendship_SendAndAccept(t *testing.T) {
	friendship, db := setupFriendship(t)
	ctx := context.Background()

	alice := createTestUser(t, db, "alice_friend")
	bob := createTestUser(t, db, "bob_friend")

	// 申請送信
	sendResp, err := friendship.SendFriendRequest(ctx, &inputport.SendFriendRequestRequest{
		RequesterID: alice.ID,
		AddresseeID: bob.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "pending", string(sendResp.Friendship.Status))

	// 保留リクエスト確認
	pendingResp, err := friendship.GetPendingRequests(ctx, &inputport.GetPendingRequestsRequest{
		UserID: bob.ID,
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Len(t, pendingResp.Requests, 1)

	// 承認
	acceptResp, err := friendship.AcceptFriendRequest(ctx, &inputport.AcceptFriendRequestRequest{
		FriendshipID: sendResp.Friendship.ID,
		UserID:       bob.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "accepted", string(acceptResp.Friendship.Status))

	// フレンド一覧確認
	friendsResp, err := friendship.GetFriends(ctx, &inputport.GetFriendsRequest{
		UserID: alice.ID,
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Len(t, friendsResp.Friends, 1)
}

// TestFriendship_RejectRequest は友達申請拒否を検証
func TestFriendship_RejectRequest(t *testing.T) {
	friendship, db := setupFriendship(t)
	ctx := context.Background()

	alice := createTestUser(t, db, "alice_reject")
	bob := createTestUser(t, db, "bob_reject")

	// 申請
	sendResp, err := friendship.SendFriendRequest(ctx, &inputport.SendFriendRequestRequest{
		RequesterID: alice.ID,
		AddresseeID: bob.ID,
	})
	require.NoError(t, err)

	// 拒否
	rejectResp, err := friendship.RejectFriendRequest(ctx, &inputport.RejectFriendRequestRequest{
		FriendshipID: sendResp.Friendship.ID,
		UserID:       bob.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "rejected", string(rejectResp.Friendship.Status))

	// フレンド一覧は空
	friendsResp, err := friendship.GetFriends(ctx, &inputport.GetFriendsRequest{
		UserID: alice.ID,
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Empty(t, friendsResp.Friends)
}

// TestFriendship_RemoveFriend はフレンド削除を検証
func TestFriendship_RemoveFriend(t *testing.T) {
	friendship, db := setupFriendship(t)
	ctx := context.Background()

	alice := createTestUser(t, db, "alice_remove")
	bob := createTestUser(t, db, "bob_remove")

	// 申請→承認
	sendResp, err := friendship.SendFriendRequest(ctx, &inputport.SendFriendRequestRequest{
		RequesterID: alice.ID,
		AddresseeID: bob.ID,
	})
	require.NoError(t, err)
	_, err = friendship.AcceptFriendRequest(ctx, &inputport.AcceptFriendRequestRequest{
		FriendshipID: sendResp.Friendship.ID,
		UserID:       bob.ID,
	})
	require.NoError(t, err)

	// 削除
	removeResp, err := friendship.RemoveFriend(ctx, &inputport.RemoveFriendRequest{
		UserID:       alice.ID,
		FriendshipID: sendResp.Friendship.ID,
	})
	require.NoError(t, err)
	assert.True(t, removeResp.Success)

	// フレンド一覧は空
	friendsResp, err := friendship.GetFriends(ctx, &inputport.GetFriendsRequest{
		UserID: alice.ID,
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Empty(t, friendsResp.Friends)
}

// TestFriendship_PendingRequestCount は保留リクエスト件数を検証
func TestFriendship_PendingRequestCount(t *testing.T) {
	friendship, db := setupFriendship(t)
	ctx := context.Background()

	bob := createTestUser(t, db, "bob_count")

	// 3人から申請
	for i := 0; i < 3; i++ {
		sender := createTestUser(t, db, fmt.Sprintf("sender_%d", i))
		_, err := friendship.SendFriendRequest(ctx, &inputport.SendFriendRequestRequest{
			RequesterID: sender.ID,
			AddresseeID: bob.ID,
		})
		require.NoError(t, err)
	}

	countResp, err := friendship.GetFriendPendingRequestCount(ctx, &inputport.GetFriendPendingRequestCountRequest{
		UserID: bob.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(3), countResp.Count)
}
