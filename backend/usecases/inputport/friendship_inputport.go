package inputport

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// FriendshipInputPort は友達機能のユースケースインターフェース
type FriendshipInputPort interface {
	// SendFriendRequest は友達申請を送信
	SendFriendRequest(ctx context.Context, req *SendFriendRequestRequest) (*SendFriendRequestResponse, error)

	// AcceptFriendRequest は友達申請を承認
	AcceptFriendRequest(ctx context.Context, req *AcceptFriendRequestRequest) (*AcceptFriendRequestResponse, error)

	// RejectFriendRequest は友達申請を拒否
	RejectFriendRequest(ctx context.Context, req *RejectFriendRequestRequest) (*RejectFriendRequestResponse, error)

	// GetFriends は友達一覧を取得
	GetFriends(ctx context.Context, req *GetFriendsRequest) (*GetFriendsResponse, error)

	// GetPendingRequests は保留中の友達申請を取得
	GetPendingRequests(ctx context.Context, req *GetPendingRequestsRequest) (*GetPendingRequestsResponse, error)

	// RemoveFriend は友達を削除
	RemoveFriend(ctx context.Context, req *RemoveFriendRequest) (*RemoveFriendResponse, error)
}

// SendFriendRequestRequest は友達申請リクエスト
type SendFriendRequestRequest struct {
	RequesterID uuid.UUID
	AddresseeID uuid.UUID
}

// SendFriendRequestResponse は友達申請レスポンス
type SendFriendRequestResponse struct {
	Friendship *entities.Friendship
}

// AcceptFriendRequestRequest は友達申請承認リクエスト
type AcceptFriendRequestRequest struct {
	FriendshipID uuid.UUID
	UserID       uuid.UUID
}

// AcceptFriendRequestResponse は友達申請承認レスポンス
type AcceptFriendRequestResponse struct {
	Friendship *entities.Friendship
}

// RejectFriendRequestRequest は友達申請拒否リクエスト
type RejectFriendRequestRequest struct {
	FriendshipID uuid.UUID
	UserID       uuid.UUID
}

// RejectFriendRequestResponse は友達申請拒否レスポンス
type RejectFriendRequestResponse struct {
	Friendship *entities.Friendship
}

// GetFriendsRequest は友達一覧取得リクエスト
type GetFriendsRequest struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}

// FriendInfo は友達情報
type FriendInfo struct {
	Friendship *entities.Friendship
	Friend     *entities.User
}

// GetFriendsResponse は友達一覧取得レスポンス
type GetFriendsResponse struct {
	Friends []*FriendInfo
}

// GetPendingRequestsRequest は保留中の友達申請取得リクエスト
type GetPendingRequestsRequest struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}

// PendingRequestInfo は保留中の友達申請情報
type PendingRequestInfo struct {
	Friendship *entities.Friendship
	Requester  *entities.User
}

// GetPendingRequestsResponse は保留中の友達申請取得レスポンス
type GetPendingRequestsResponse struct {
	Requests []*PendingRequestInfo
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
