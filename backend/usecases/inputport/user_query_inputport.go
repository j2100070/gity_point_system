package inputport

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// UserQueryInputPort はユーザー情報検索のユースケースインターフェース
type UserQueryInputPort interface {
	// GetUserByID はユーザーIDでユーザー情報を取得
	GetUserByID(ctx context.Context, req *GetUserByIDRequest) (*GetUserByIDResponse, error)

	// SearchUserByUsername はユーザー名でユーザーを検索
	SearchUserByUsername(ctx context.Context, req *SearchUserByUsernameRequest) (*SearchUserByUsernameResponse, error)
}

// GetUserByIDRequest はユーザーID検索のリクエスト
type GetUserByIDRequest struct {
	UserID uuid.UUID
}

// GetUserByIDResponse はユーザーID検索のレスポンス
type GetUserByIDResponse struct {
	User *entities.User
}

// SearchUserByUsernameRequest はユーザー名検索のリクエスト
type SearchUserByUsernameRequest struct {
	Username string
}

// SearchUserByUsernameResponse はユーザー名検索のレスポンス
type SearchUserByUsernameResponse struct {
	User *entities.User
}
