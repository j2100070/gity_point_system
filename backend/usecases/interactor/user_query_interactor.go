package interactor

import (
	"context"
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
)

// UserQueryInteractor はユーザー情報検索のユースケース実装
type UserQueryInteractor struct {
	userRepo repository.UserRepository
	logger   entities.Logger
}

// NewUserQueryInteractor は新しいUserQueryInteractorを作成
func NewUserQueryInteractor(
	userRepo repository.UserRepository,
	logger entities.Logger,
) inputport.UserQueryInputPort {
	return &UserQueryInteractor{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetUserByID はユーザーIDでユーザー情報を取得
func (i *UserQueryInteractor) GetUserByID(ctx context.Context, req *inputport.GetUserByIDRequest) (*inputport.GetUserByIDResponse, error) {
	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &inputport.GetUserByIDResponse{
		User: user,
	}, nil
}

// SearchUserByUsername はユーザー名でユーザーを検索
func (i *UserQueryInteractor) SearchUserByUsername(ctx context.Context, req *inputport.SearchUserByUsernameRequest) (*inputport.SearchUserByUsernameResponse, error) {
	user, err := i.userRepo.ReadByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &inputport.SearchUserByUsernameResponse{
		User: user,
	}, nil
}
