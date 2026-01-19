package usecase

import (
	"errors"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
)

// FriendshipUseCase は友達機能に関するユースケース
type FriendshipUseCase struct {
	friendshipRepo domain.FriendshipRepository
	userRepo       domain.UserRepository
}

// NewFriendshipUseCase は新しいFriendshipUseCaseを作成
func NewFriendshipUseCase(
	friendshipRepo domain.FriendshipRepository,
	userRepo domain.UserRepository,
) *FriendshipUseCase {
	return &FriendshipUseCase{
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
	}
}

// SendFriendRequestRequest は友達申請リクエスト
type SendFriendRequestRequest struct {
	RequesterID uuid.UUID
	AddresseeID uuid.UUID
}

// SendFriendRequestResponse は友達申請レスポンス
type SendFriendRequestResponse struct {
	Friendship *domain.Friendship
}

// SendFriendRequest は友達申請を送信
func (u *FriendshipUseCase) SendFriendRequest(req *SendFriendRequestRequest) (*SendFriendRequestResponse, error) {
	// 受信者の存在確認
	addressee, err := u.userRepo.FindByID(req.AddresseeID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !addressee.IsActive {
		return nil, errors.New("user is not active")
	}

	// 既存の友達関係チェック
	existing, _ := u.friendshipRepo.FindByUsers(req.RequesterID, req.AddresseeID)
	if existing != nil {
		if existing.Status == domain.FriendshipStatusAccepted {
			return nil, errors.New("already friends")
		}
		if existing.Status == domain.FriendshipStatusPending {
			return nil, errors.New("friend request already sent")
		}
		if existing.Status == domain.FriendshipStatusBlocked {
			return nil, errors.New("cannot send friend request")
		}
	}

	// 友達申請作成
	friendship, err := domain.NewFriendship(req.RequesterID, req.AddresseeID)
	if err != nil {
		return nil, err
	}

	if err := u.friendshipRepo.Create(friendship); err != nil {
		return nil, err
	}

	return &SendFriendRequestResponse{Friendship: friendship}, nil
}

// AcceptFriendRequestRequest は友達申請承認リクエスト
type AcceptFriendRequestRequest struct {
	FriendshipID uuid.UUID
	UserID       uuid.UUID // 承認するユーザー（受信者であることを確認）
}

// AcceptFriendRequestResponse は友達申請承認レスポンス
type AcceptFriendRequestResponse struct {
	Friendship *domain.Friendship
}

// AcceptFriendRequest は友達申請を承認
func (u *FriendshipUseCase) AcceptFriendRequest(req *AcceptFriendRequestRequest) (*AcceptFriendRequestResponse, error) {
	friendship, err := u.friendshipRepo.FindByID(req.FriendshipID)
	if err != nil {
		return nil, err
	}

	// 受信者であることを確認
	if friendship.AddresseeID != req.UserID {
		return nil, errors.New("unauthorized to accept this request")
	}

	if err := friendship.Accept(); err != nil {
		return nil, err
	}

	if err := u.friendshipRepo.Update(friendship); err != nil {
		return nil, err
	}

	return &AcceptFriendRequestResponse{Friendship: friendship}, nil
}

// RejectFriendRequestRequest は友達申請拒否リクエスト
type RejectFriendRequestRequest struct {
	FriendshipID uuid.UUID
	UserID       uuid.UUID
}

// RejectFriendRequestResponse は友達申請拒否レスポンス
type RejectFriendRequestResponse struct {
	Friendship *domain.Friendship
}

// RejectFriendRequest は友達申請を拒否
func (u *FriendshipUseCase) RejectFriendRequest(req *RejectFriendRequestRequest) (*RejectFriendRequestResponse, error) {
	friendship, err := u.friendshipRepo.FindByID(req.FriendshipID)
	if err != nil {
		return nil, err
	}

	// 受信者であることを確認
	if friendship.AddresseeID != req.UserID {
		return nil, errors.New("unauthorized to reject this request")
	}

	if err := friendship.Reject(); err != nil {
		return nil, err
	}

	if err := u.friendshipRepo.Update(friendship); err != nil {
		return nil, err
	}

	return &RejectFriendRequestResponse{Friendship: friendship}, nil
}

// GetFriendsRequest は友達一覧取得リクエスト
type GetFriendsRequest struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}

// FriendInfo は友達情報
type FriendInfo struct {
	Friendship *domain.Friendship
	Friend     *domain.User
}

// GetFriendsResponse は友達一覧取得レスポンス
type GetFriendsResponse struct {
	Friends []*FriendInfo
}

// GetFriends は友達一覧を取得
func (u *FriendshipUseCase) GetFriends(req *GetFriendsRequest) (*GetFriendsResponse, error) {
	friendships, err := u.friendshipRepo.ListFriends(req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	friends := make([]*FriendInfo, len(friendships))
	for i, friendship := range friendships {
		// 相手のユーザーIDを取得
		friendID := friendship.AddresseeID
		if friendship.AddresseeID == req.UserID {
			friendID = friendship.RequesterID
		}

		// ユーザー情報取得
		user, err := u.userRepo.FindByID(friendID)
		if err != nil {
			continue
		}

		friends[i] = &FriendInfo{
			Friendship: friendship,
			Friend:     user,
		}
	}

	return &GetFriendsResponse{Friends: friends}, nil
}

// GetPendingRequestsRequest は保留中の友達申請取得リクエスト
type GetPendingRequestsRequest struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}

// PendingRequestInfo は保留中の友達申請情報
type PendingRequestInfo struct {
	Friendship *domain.Friendship
	Requester  *domain.User
}

// GetPendingRequestsResponse は保留中の友達申請取得レスポンス
type GetPendingRequestsResponse struct {
	Requests []*PendingRequestInfo
}

// GetPendingRequests は保留中の友達申請を取得
func (u *FriendshipUseCase) GetPendingRequests(req *GetPendingRequestsRequest) (*GetPendingRequestsResponse, error) {
	friendships, err := u.friendshipRepo.ListPendingRequests(req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	requests := make([]*PendingRequestInfo, len(friendships))
	for i, friendship := range friendships {
		// 申請者のユーザー情報取得
		user, err := u.userRepo.FindByID(friendship.RequesterID)
		if err != nil {
			continue
		}

		requests[i] = &PendingRequestInfo{
			Friendship: friendship,
			Requester:  user,
		}
	}

	return &GetPendingRequestsResponse{Requests: requests}, nil
}

// RemoveFriendRequest は友達削除リクエスト
type RemoveFriendRequest struct {
	UserID       uuid.UUID
	FriendshipID uuid.UUID
}

// RemoveFriendResponse は友達削除レスポンス
type RemoveFriendResponse struct {
	Success bool
}

// RemoveFriend は友達を削除
func (u *FriendshipUseCase) RemoveFriend(req *RemoveFriendRequest) (*RemoveFriendResponse, error) {
	friendship, err := u.friendshipRepo.FindByID(req.FriendshipID)
	if err != nil {
		return nil, err
	}

	// 関係者であることを確認
	if friendship.RequesterID != req.UserID && friendship.AddresseeID != req.UserID {
		return nil, errors.New("unauthorized to remove this friendship")
	}

	if err := u.friendshipRepo.Delete(req.FriendshipID); err != nil {
		return nil, err
	}

	return &RemoveFriendResponse{Success: true}, nil
}
