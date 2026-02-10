package presenter

import (
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// FriendPresenter は友達機能のプレゼンター
type FriendPresenter struct{}

// NewFriendPresenter は新しいFriendPresenterを作成
func NewFriendPresenter() *FriendPresenter {
	return &FriendPresenter{}
}

// FriendshipResponse は友達関係のレスポンス
type FriendshipResponse struct {
	ID          uuid.UUID `json:"id"`
	RequesterID uuid.UUID `json:"requester_id"`
	AddresseeID uuid.UUID `json:"addressee_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FriendInfoResponse は友達情報のレスポンス
type FriendInfoResponse struct {
	Friendship FriendshipResponse `json:"friendship"`
	Friend     UserResponse       `json:"friend"`
}

// PendingRequestInfoResponse は保留中申請情報のレスポンス
type PendingRequestInfoResponse struct {
	Friendship FriendshipResponse `json:"friendship"`
	Requester  UserResponse       `json:"requester"`
}

// PresentSendFriendRequest は友達申請送信レスポンスを生成
func (p *FriendPresenter) PresentSendFriendRequest(resp *inputport.SendFriendRequestResponse) map[string]interface{} {
	return map[string]interface{}{
		"friendship": p.toFriendshipResponse(resp.Friendship),
	}
}

// PresentAcceptFriendRequest は友達申請承認レスポンスを生成
func (p *FriendPresenter) PresentAcceptFriendRequest(resp *inputport.AcceptFriendRequestResponse) map[string]interface{} {
	return map[string]interface{}{
		"friendship": p.toFriendshipResponse(resp.Friendship),
	}
}

// PresentRejectFriendRequest は友達申請拒否レスポンスを生成
func (p *FriendPresenter) PresentRejectFriendRequest(resp *inputport.RejectFriendRequestResponse) map[string]interface{} {
	return map[string]interface{}{
		"friendship": p.toFriendshipResponse(resp.Friendship),
	}
}

// PresentGetFriends は友達一覧レスポンスを生成
func (p *FriendPresenter) PresentGetFriends(resp *inputport.GetFriendsResponse) map[string]interface{} {
	friends := make([]FriendInfoResponse, 0, len(resp.Friends))
	for _, f := range resp.Friends {
		friends = append(friends, FriendInfoResponse{
			Friendship: p.toFriendshipResponse(f.Friendship),
			Friend:     p.toUserResponse(f.Friend),
		})
	}

	return map[string]interface{}{
		"friends": friends,
	}
}

// PresentGetPendingRequests は保留中申請一覧レスポンスを生成
func (p *FriendPresenter) PresentGetPendingRequests(resp *inputport.GetPendingRequestsResponse) map[string]interface{} {
	requests := make([]PendingRequestInfoResponse, 0, len(resp.Requests))
	for _, r := range resp.Requests {
		requests = append(requests, PendingRequestInfoResponse{
			Friendship: p.toFriendshipResponse(r.Friendship),
			Requester:  p.toUserResponse(r.Requester),
		})
	}

	return map[string]interface{}{
		"requests": requests,
	}
}

// PresentRemoveFriend は友達削除レスポンスを生成
func (p *FriendPresenter) PresentRemoveFriend(resp *inputport.RemoveFriendResponse) map[string]interface{} {
	return map[string]interface{}{
		"success": resp.Success,
	}
}

// toFriendshipResponse はFriendshipエンティティをレスポンスに変換
func (p *FriendPresenter) toFriendshipResponse(friendship *entities.Friendship) FriendshipResponse {
	return FriendshipResponse{
		ID:          friendship.ID,
		RequesterID: friendship.RequesterID,
		AddresseeID: friendship.AddresseeID,
		Status:      string(friendship.Status),
		CreatedAt:   friendship.CreatedAt,
		UpdatedAt:   friendship.UpdatedAt,
	}
}

// toUserResponse はUserエンティティをレスポンスに変換
func (p *FriendPresenter) toUserResponse(user *entities.User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Balance:     user.Balance,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
