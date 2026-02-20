package interactor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// Mock Repositories
// ========================================

type mockFriendshipRepo struct {
	friendships  map[uuid.UUID]*entities.Friendship
	byUsers      map[string]*entities.Friendship // key: "userID1-userID2"
	friends      []*entities.Friendship
	pending      []*entities.Friendship
	friendsUsers map[uuid.UUID]*entities.User // friendID -> User (for WithUsers)
	pendingUsers map[uuid.UUID]*entities.User // requesterID -> User (for WithUsers)
	createErr    error
	readErr      error
	updateErr    error
	deleteErr    error
	archiveErr   error
}

func newMockFriendshipRepo() *mockFriendshipRepo {
	return &mockFriendshipRepo{
		friendships:  make(map[uuid.UUID]*entities.Friendship),
		byUsers:      make(map[string]*entities.Friendship),
		friendsUsers: make(map[uuid.UUID]*entities.User),
		pendingUsers: make(map[uuid.UUID]*entities.User),
	}
}

func (m *mockFriendshipRepo) Create(ctx context.Context, f *entities.Friendship) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.friendships[f.ID] = f
	key := f.RequesterID.String() + "-" + f.AddresseeID.String()
	m.byUsers[key] = f
	return nil
}

func (m *mockFriendshipRepo) Read(ctx context.Context, id uuid.UUID) (*entities.Friendship, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	f, ok := m.friendships[id]
	if !ok {
		return nil, errors.New("friendship not found")
	}
	return f, nil
}

func (m *mockFriendshipRepo) ReadByUsers(ctx context.Context, userID1, userID2 uuid.UUID) (*entities.Friendship, error) {
	key1 := userID1.String() + "-" + userID2.String()
	key2 := userID2.String() + "-" + userID1.String()
	if f, ok := m.byUsers[key1]; ok {
		return f, nil
	}
	if f, ok := m.byUsers[key2]; ok {
		return f, nil
	}
	return nil, errors.New("friendship not found")
}

func (m *mockFriendshipRepo) ReadListFriends(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	return m.friends, nil
}

func (m *mockFriendshipRepo) ReadListPendingRequests(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.Friendship, error) {
	return m.pending, nil
}

func (m *mockFriendshipRepo) Update(ctx context.Context, f *entities.Friendship) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.friendships[f.ID] = f
	return nil
}

func (m *mockFriendshipRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.friendships, id)
	return nil
}

func (m *mockFriendshipRepo) ArchiveAndDelete(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error {
	if m.archiveErr != nil {
		return m.archiveErr
	}
	if f, ok := m.friendships[id]; ok {
		key := f.RequesterID.String() + "-" + f.AddresseeID.String()
		delete(m.byUsers, key)
	}
	delete(m.friendships, id)
	return nil
}

func (m *mockFriendshipRepo) CheckAreFriends(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error) {
	key1 := userID1.String() + "-" + userID2.String()
	key2 := userID2.String() + "-" + userID1.String()
	if f, ok := m.byUsers[key1]; ok && f.Status == entities.FriendshipStatusAccepted {
		return true, nil
	}
	if f, ok := m.byUsers[key2]; ok && f.Status == entities.FriendshipStatusAccepted {
		return true, nil
	}
	return false, nil
}

func (m *mockFriendshipRepo) ReadListFriendsWithUsers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.FriendshipWithUser, error) {
	results := make([]*entities.FriendshipWithUser, 0, len(m.friends))
	for _, f := range m.friends {
		friendID := f.AddresseeID
		if f.AddresseeID == userID {
			friendID = f.RequesterID
		}
		user := m.friendsUsers[friendID]
		results = append(results, &entities.FriendshipWithUser{
			Friendship: f,
			User:       user,
		})
	}
	return results, nil
}

func (m *mockFriendshipRepo) ReadListPendingRequestsWithUsers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.FriendshipWithUser, error) {
	results := make([]*entities.FriendshipWithUser, 0, len(m.pending))
	for _, f := range m.pending {
		user := m.pendingUsers[f.RequesterID]
		results = append(results, &entities.FriendshipWithUser{
			Friendship: f,
			User:       user,
		})
	}
	return results, nil
}

func (m *mockFriendshipRepo) CountPendingRequests(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	for _, f := range m.pending {
		if f.AddresseeID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockFriendshipRepo) setExistingFriendship(f *entities.Friendship) {
	m.friendships[f.ID] = f
	key := f.RequesterID.String() + "-" + f.AddresseeID.String()
	m.byUsers[key] = f
}

type mockUserRepo struct {
	users   map[uuid.UUID]*entities.User
	readErr error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[uuid.UUID]*entities.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) error { return nil }
func (m *mockUserRepo) Read(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}
func (m *mockUserRepo) ReadByUsername(ctx context.Context, username string) (*entities.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ReadByEmail(ctx context.Context, email string) (*entities.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *entities.User) (bool, error) {
	return true, nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockUserRepo) ReadListAll(ctx context.Context, offset, limit int) ([]*entities.User, int, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) ReadList(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ReadPersonalQRCode(ctx context.Context, userID uuid.UUID) (string, error) {
	return "", nil
}
func (m *mockUserRepo) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockUserRepo) ReadListWithSearch(ctx context.Context, search, sortBy, sortOrder string, offset, limit int) ([]*entities.User, error) {
	return nil, nil
}
func (m *mockUserRepo) CountWithSearch(ctx context.Context, search string) (int64, error) {
	return 0, nil
}
func (m *mockUserRepo) UpdateBalanceWithLock(ctx context.Context, userID uuid.UUID, amount int64, isDeduct bool) error {
	return nil
}
func (m *mockUserRepo) UpdateBalancesWithLock(ctx context.Context, updates []repository.BalanceUpdate) error {
	return nil
}

func (m *mockUserRepo) addUser(user *entities.User) {
	m.users[user.ID] = user
}

type mockFriendshipLogger struct{}

func (m *mockFriendshipLogger) Debug(msg string, fields ...entities.Field) {}
func (m *mockFriendshipLogger) Info(msg string, fields ...entities.Field)  {}
func (m *mockFriendshipLogger) Warn(msg string, fields ...entities.Field)  {}
func (m *mockFriendshipLogger) Error(msg string, fields ...entities.Field) {}
func (m *mockFriendshipLogger) Fatal(msg string, fields ...entities.Field) {}

// ========================================
// Helper functions
// ========================================

func createActiveUser(id uuid.UUID) *entities.User {
	return &entities.User{
		ID:          id,
		Username:    "user_" + id.String()[:8],
		DisplayName: "User " + id.String()[:8],
		IsActive:    true,
		Role:        entities.RoleUser,
	}
}

func createInactiveUser(id uuid.UUID) *entities.User {
	u := createActiveUser(id)
	u.IsActive = false
	return u
}

// ========================================
// SendFriendRequest Tests
// ========================================

func TestSendFriendRequest(t *testing.T) {
	t.Run("正常にフレンド申請を送信", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()
		userRepo.addUser(createActiveUser(requesterID))
		userRepo.addUser(createActiveUser(addresseeID))

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: requesterID,
			AddresseeID: addresseeID,
		})

		require.NoError(t, err)
		assert.NotNil(t, resp.Friendship)
		assert.Equal(t, entities.FriendshipStatusPending, resp.Friendship.Status)
		assert.Equal(t, requesterID, resp.Friendship.RequesterID)
		assert.Equal(t, addresseeID, resp.Friendship.AddresseeID)
	})

	t.Run("存在しないユーザーへの申請はエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()
		userRepo.addUser(createActiveUser(requesterID))
		// addresseeを追加しない

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: requesterID,
			AddresseeID: addresseeID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("非アクティブユーザーへの申請はエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()
		userRepo.addUser(createActiveUser(requesterID))
		userRepo.addUser(createInactiveUser(addresseeID))

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: requesterID,
			AddresseeID: addresseeID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user is not active")
	})

	t.Run("既にフレンドのユーザーへの申請はエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()
		userRepo.addUser(createActiveUser(requesterID))
		userRepo.addUser(createActiveUser(addresseeID))

		existing, _ := entities.NewFriendship(requesterID, addresseeID)
		existing.Accept()
		friendshipRepo.setExistingFriendship(existing)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: requesterID,
			AddresseeID: addresseeID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already friends")
	})

	t.Run("保留中の申請が既にある場合はエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()
		userRepo.addUser(createActiveUser(requesterID))
		userRepo.addUser(createActiveUser(addresseeID))

		existing, _ := entities.NewFriendship(requesterID, addresseeID)
		friendshipRepo.setExistingFriendship(existing)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: requesterID,
			AddresseeID: addresseeID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "friend request already sent")
	})

	t.Run("ブロック中のユーザーへの申請はエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()
		userRepo.addUser(createActiveUser(requesterID))
		userRepo.addUser(createActiveUser(addresseeID))

		existing, _ := entities.NewFriendship(requesterID, addresseeID)
		existing.Block()
		friendshipRepo.setExistingFriendship(existing)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: requesterID,
			AddresseeID: addresseeID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot send friend request")
	})

	t.Run("拒否済みの申請後に再申請が可能", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()
		userRepo.addUser(createActiveUser(requesterID))
		userRepo.addUser(createActiveUser(addresseeID))

		existing, _ := entities.NewFriendship(requesterID, addresseeID)
		existing.Reject()
		friendshipRepo.setExistingFriendship(existing)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: requesterID,
			AddresseeID: addresseeID,
		})

		require.NoError(t, err)
		assert.NotNil(t, resp.Friendship)
		assert.Equal(t, entities.FriendshipStatusPending, resp.Friendship.Status)
		// 既存レコードが再利用される（IDが同じ）
		assert.Equal(t, existing.ID, resp.Friendship.ID)
	})
}

// ========================================
// AcceptFriendRequest Tests
// ========================================

func TestAcceptFriendRequest(t *testing.T) {
	t.Run("正常にフレンド申請を承認", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		friendshipRepo.setExistingFriendship(f)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.AcceptFriendRequest(context.Background(), &inputport.AcceptFriendRequestRequest{
			FriendshipID: f.ID,
			UserID:       addresseeID,
		})

		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusAccepted, resp.Friendship.Status)
	})

	t.Run("申請者が自分の申請を承認しようとするとエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		friendshipRepo.setExistingFriendship(f)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.AcceptFriendRequest(context.Background(), &inputport.AcceptFriendRequestRequest{
			FriendshipID: f.ID,
			UserID:       requesterID, // 申請者が承認しようとする
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("無関係なユーザーが承認しようとするとエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()
		otherUser := uuid.New()

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		friendshipRepo.setExistingFriendship(f)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.AcceptFriendRequest(context.Background(), &inputport.AcceptFriendRequestRequest{
			FriendshipID: f.ID,
			UserID:       otherUser,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("存在しないフレンドシップIDはエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.AcceptFriendRequest(context.Background(), &inputport.AcceptFriendRequestRequest{
			FriendshipID: uuid.New(),
			UserID:       uuid.New(),
		})

		assert.Error(t, err)
	})
}

// ========================================
// RejectFriendRequest Tests
// ========================================

func TestRejectFriendRequest(t *testing.T) {
	t.Run("正常にフレンド申請を拒否", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		friendshipRepo.setExistingFriendship(f)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.RejectFriendRequest(context.Background(), &inputport.RejectFriendRequestRequest{
			FriendshipID: f.ID,
			UserID:       addresseeID,
		})

		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusRejected, resp.Friendship.Status)
	})

	t.Run("申請者が自分の申請を拒否しようとするとエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		friendshipRepo.setExistingFriendship(f)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.RejectFriendRequest(context.Background(), &inputport.RejectFriendRequestRequest{
			FriendshipID: f.ID,
			UserID:       requesterID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

// ========================================
// RemoveFriend Tests
// ========================================

func TestRemoveFriend(t *testing.T) {
	t.Run("申請者側がフレンド解散", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		f.Accept()
		friendshipRepo.setExistingFriendship(f)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.RemoveFriend(context.Background(), &inputport.RemoveFriendRequest{
			UserID:       requesterID,
			FriendshipID: f.ID,
		})

		require.NoError(t, err)
		assert.True(t, resp.Success)
		// 削除されていることを確認
		_, readErr := friendshipRepo.Read(context.Background(), f.ID)
		assert.Error(t, readErr)
	})

	t.Run("受信者側がフレンド解散", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		f.Accept()
		friendshipRepo.setExistingFriendship(f)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.RemoveFriend(context.Background(), &inputport.RemoveFriendRequest{
			UserID:       addresseeID,
			FriendshipID: f.ID,
		})

		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("無関係なユーザーが解散しようとするとエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()
		otherUser := uuid.New()

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		f.Accept()
		friendshipRepo.setExistingFriendship(f)

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.RemoveFriend(context.Background(), &inputport.RemoveFriendRequest{
			UserID:       otherUser,
			FriendshipID: f.ID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("存在しないフレンドシップの削除はエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.RemoveFriend(context.Background(), &inputport.RemoveFriendRequest{
			UserID:       uuid.New(),
			FriendshipID: uuid.New(),
		})

		assert.Error(t, err)
	})

	t.Run("アーカイブ失敗時はエラー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		requesterID := uuid.New()
		addresseeID := uuid.New()

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		f.Accept()
		friendshipRepo.setExistingFriendship(f)
		friendshipRepo.archiveErr = errors.New("archive failed")

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		_, err := interactorInstance.RemoveFriend(context.Background(), &inputport.RemoveFriendRequest{
			UserID:       requesterID,
			FriendshipID: f.ID,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "archive failed")
	})
}

// ========================================
// GetFriends Tests
// ========================================

func TestGetFriends(t *testing.T) {
	t.Run("友達一覧を正常に取得", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		userID := uuid.New()
		friendID := uuid.New()

		userRepo.addUser(createActiveUser(userID))
		userRepo.addUser(createActiveUser(friendID))

		f, _ := entities.NewFriendship(userID, friendID)
		f.Accept()
		friendshipRepo.friends = []*entities.Friendship{f}
		friendshipRepo.friendsUsers[friendID] = userRepo.users[friendID]

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.GetFriends(context.Background(), &inputport.GetFriendsRequest{
			UserID: userID,
			Offset: 0,
			Limit:  20,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Friends, 1)
		assert.Equal(t, friendID, resp.Friends[0].Friend.ID)
	})

	t.Run("友達がいない場合は空のリスト", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		userID := uuid.New()
		friendshipRepo.friends = []*entities.Friendship{}

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.GetFriends(context.Background(), &inputport.GetFriendsRequest{
			UserID: userID,
			Offset: 0,
			Limit:  20,
		})

		require.NoError(t, err)
		assert.Empty(t, resp.Friends)
	})
}

// ========================================
// GetPendingRequests Tests
// ========================================

func TestGetPendingRequests(t *testing.T) {
	t.Run("保留中の申請を正常に取得", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		addresseeID := uuid.New()
		requesterID := uuid.New()

		userRepo.addUser(createActiveUser(requesterID))
		userRepo.addUser(createActiveUser(addresseeID))

		f, _ := entities.NewFriendship(requesterID, addresseeID)
		friendshipRepo.pending = []*entities.Friendship{f}
		friendshipRepo.pendingUsers[requesterID] = userRepo.users[requesterID]

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.GetPendingRequests(context.Background(), &inputport.GetPendingRequestsRequest{
			UserID: addresseeID,
			Offset: 0,
			Limit:  20,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Requests, 1)
		assert.Equal(t, requesterID, resp.Requests[0].Requester.ID)
	})

	t.Run("保留中の申請がない場合は空のリスト", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		friendshipRepo.pending = []*entities.Friendship{}

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		resp, err := interactorInstance.GetPendingRequests(context.Background(), &inputport.GetPendingRequestsRequest{
			UserID: uuid.New(),
			Offset: 0,
			Limit:  20,
		})

		require.NoError(t, err)
		assert.Empty(t, resp.Requests)
	})
}

// ========================================
// Full Flow Integration-style Tests
// ========================================

func TestFriendshipFullFlow(t *testing.T) {
	t.Run("申請→承認→解散→再申請のフルフロー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(createActiveUser(userA))
		userRepo.addUser(createActiveUser(userB))

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		// 1. フレンド申請
		sendResp, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: userA,
			AddresseeID: userB,
		})
		require.NoError(t, err)
		friendshipID := sendResp.Friendship.ID
		assert.Equal(t, entities.FriendshipStatusPending, sendResp.Friendship.Status)

		// 2. 承認
		acceptResp, err := interactorInstance.AcceptFriendRequest(context.Background(), &inputport.AcceptFriendRequestRequest{
			FriendshipID: friendshipID,
			UserID:       userB,
		})
		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusAccepted, acceptResp.Friendship.Status)

		// 3. フレンド解散
		removeResp, err := interactorInstance.RemoveFriend(context.Background(), &inputport.RemoveFriendRequest{
			UserID:       userA,
			FriendshipID: friendshipID,
		})
		require.NoError(t, err)
		assert.True(t, removeResp.Success)

		// 4. 再申請（解散後なのでエラーにならないこと）
		// ReadByUsersがfriendship not foundを返す（削除済み）ので新規作成される
		reSendResp, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: userA,
			AddresseeID: userB,
		})
		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusPending, reSendResp.Friendship.Status)
		assert.NotEqual(t, friendshipID, reSendResp.Friendship.ID, "新しいフレンドシップIDが生成される")
	})

	t.Run("申請→拒否→再申請のフロー", func(t *testing.T) {
		friendshipRepo := newMockFriendshipRepo()
		userRepo := newMockUserRepo()
		userA := uuid.New()
		userB := uuid.New()
		userRepo.addUser(createActiveUser(userA))
		userRepo.addUser(createActiveUser(userB))

		interactorInstance := interactor.NewFriendshipInteractor(friendshipRepo, userRepo, &mockFriendshipLogger{})

		// 1. フレンド申請
		sendResp, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: userA,
			AddresseeID: userB,
		})
		require.NoError(t, err)
		friendshipID := sendResp.Friendship.ID

		// 2. 拒否
		rejectResp, err := interactorInstance.RejectFriendRequest(context.Background(), &inputport.RejectFriendRequestRequest{
			FriendshipID: friendshipID,
			UserID:       userB,
		})
		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusRejected, rejectResp.Friendship.Status)

		// 3. 再申請（拒否後なのでエラーにならないこと）
		reSendResp, err := interactorInstance.SendFriendRequest(context.Background(), &inputport.SendFriendRequestRequest{
			RequesterID: userA,
			AddresseeID: userB,
		})
		require.NoError(t, err)
		assert.Equal(t, entities.FriendshipStatusPending, reSendResp.Friendship.Status)
		// 同じレコードが再利用される
		assert.Equal(t, friendshipID, reSendResp.Friendship.ID, "既存レコードが再利用される")
	})
}
