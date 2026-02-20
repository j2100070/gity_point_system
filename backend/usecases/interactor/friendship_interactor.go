package interactor

import (
	"context"
	"errors"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
)

// FriendshipInteractor は友達機能のユースケース実装
type FriendshipInteractor struct {
	friendshipRepo repository.FriendshipRepository
	userRepo       repository.UserRepository
	logger         entities.Logger
}

// NewFriendshipInteractor は新しいFriendshipInteractorを作成
func NewFriendshipInteractor(
	friendshipRepo repository.FriendshipRepository,
	userRepo repository.UserRepository,
	logger entities.Logger,
) inputport.FriendshipInputPort {
	return &FriendshipInteractor{
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
		logger:         logger,
	}
}

// SendFriendRequest は友達申請を送信
func (i *FriendshipInteractor) SendFriendRequest(ctx context.Context, req *inputport.SendFriendRequestRequest) (*inputport.SendFriendRequestResponse, error) {
	i.logger.Info("Sending friend request",
		entities.NewField("requester_id", req.RequesterID),
		entities.NewField("addressee_id", req.AddresseeID))

	// 受信者の存在確認
	addressee, err := i.userRepo.Read(ctx, req.AddresseeID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !addressee.IsActive {
		return nil, errors.New("user is not active")
	}

	// 既存の友達関係チェック
	existing, _ := i.friendshipRepo.ReadByUsers(ctx, req.RequesterID, req.AddresseeID)
	if existing != nil {
		if existing.Status == entities.FriendshipStatusAccepted {
			return nil, errors.New("already friends")
		}
		if existing.Status == entities.FriendshipStatusPending {
			return nil, errors.New("friend request already sent")
		}
		if existing.Status == entities.FriendshipStatusBlocked {
			return nil, errors.New("cannot send friend request")
		}
		// rejected の場合は再申請を許可（既存レコードを更新）
		if existing.Status == entities.FriendshipStatusRejected {
			existing.Status = entities.FriendshipStatusPending
			existing.RequesterID = req.RequesterID
			existing.AddresseeID = req.AddresseeID
			if err := i.friendshipRepo.Update(ctx, existing); err != nil {
				return nil, err
			}
			return &inputport.SendFriendRequestResponse{Friendship: existing}, nil
		}
	}

	// 友達申請作成
	friendship, err := entities.NewFriendship(req.RequesterID, req.AddresseeID)
	if err != nil {
		return nil, err
	}

	if err := i.friendshipRepo.Create(ctx, friendship); err != nil {
		return nil, err
	}

	return &inputport.SendFriendRequestResponse{Friendship: friendship}, nil
}

// AcceptFriendRequest は友達申請を承認
func (i *FriendshipInteractor) AcceptFriendRequest(ctx context.Context, req *inputport.AcceptFriendRequestRequest) (*inputport.AcceptFriendRequestResponse, error) {
	friendship, err := i.friendshipRepo.Read(ctx, req.FriendshipID)
	if err != nil {
		return nil, err
	}

	if friendship.AddresseeID != req.UserID {
		return nil, errors.New("unauthorized to accept this request")
	}

	if err := friendship.Accept(); err != nil {
		return nil, err
	}

	if err := i.friendshipRepo.Update(ctx, friendship); err != nil {
		return nil, err
	}

	return &inputport.AcceptFriendRequestResponse{Friendship: friendship}, nil
}

// RejectFriendRequest は友達申請を拒否
func (i *FriendshipInteractor) RejectFriendRequest(ctx context.Context, req *inputport.RejectFriendRequestRequest) (*inputport.RejectFriendRequestResponse, error) {
	friendship, err := i.friendshipRepo.Read(ctx, req.FriendshipID)
	if err != nil {
		return nil, err
	}

	if friendship.AddresseeID != req.UserID {
		return nil, errors.New("unauthorized to reject this request")
	}

	if err := friendship.Reject(); err != nil {
		return nil, err
	}

	if err := i.friendshipRepo.Update(ctx, friendship); err != nil {
		return nil, err
	}

	return &inputport.RejectFriendRequestResponse{Friendship: friendship}, nil
}

// GetFriends は友達一覧を取得
func (i *FriendshipInteractor) GetFriends(ctx context.Context, req *inputport.GetFriendsRequest) (*inputport.GetFriendsResponse, error) {
	results, err := i.friendshipRepo.ReadListFriendsWithUsers(ctx, req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	friends := make([]*inputport.FriendInfo, 0, len(results))
	for _, r := range results {
		friends = append(friends, &inputport.FriendInfo{
			Friendship: r.Friendship,
			Friend:     r.User,
		})
	}

	return &inputport.GetFriendsResponse{Friends: friends}, nil
}

// GetPendingRequests は保留中の友達申請を取得
func (i *FriendshipInteractor) GetPendingRequests(ctx context.Context, req *inputport.GetPendingRequestsRequest) (*inputport.GetPendingRequestsResponse, error) {
	results, err := i.friendshipRepo.ReadListPendingRequestsWithUsers(ctx, req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	requests := make([]*inputport.PendingRequestInfo, 0, len(results))
	for _, r := range results {
		requests = append(requests, &inputport.PendingRequestInfo{
			Friendship: r.Friendship,
			Requester:  r.User,
		})
	}

	return &inputport.GetPendingRequestsResponse{Requests: requests}, nil
}

// RemoveFriend は友達を削除（アーカイブに移動）
func (i *FriendshipInteractor) RemoveFriend(ctx context.Context, req *inputport.RemoveFriendRequest) (*inputport.RemoveFriendResponse, error) {
	friendship, err := i.friendshipRepo.Read(ctx, req.FriendshipID)
	if err != nil {
		return nil, err
	}

	if friendship.RequesterID != req.UserID && friendship.AddresseeID != req.UserID {
		return nil, errors.New("unauthorized to remove this friendship")
	}

	if err := i.friendshipRepo.ArchiveAndDelete(ctx, req.FriendshipID, req.UserID); err != nil {
		return nil, err
	}

	return &inputport.RemoveFriendResponse{Success: true}, nil
}

// GetFriendPendingRequestCount は保留中の友達申請件数を取得
func (i *FriendshipInteractor) GetFriendPendingRequestCount(ctx context.Context, req *inputport.GetFriendPendingRequestCountRequest) (*inputport.GetFriendPendingRequestCountResponse, error) {
	count, err := i.friendshipRepo.CountPendingRequests(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetFriendPendingRequestCountResponse{Count: count}, nil
}
