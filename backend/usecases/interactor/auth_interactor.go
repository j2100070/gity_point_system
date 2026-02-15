package interactor

import (
	"context"
	"errors"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"github.com/gity/point-system/usecases/service"
)

// AuthInteractor は認証のユースケース実装
type AuthInteractor struct {
	userRepo        repository.UserRepository
	sessionRepo     repository.SessionRepository
	passwordService service.PasswordService
	logger          entities.Logger
}

// NewAuthInteractor は新しいAuthInteractorを作成
func NewAuthInteractor(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	passwordService service.PasswordService,
	logger entities.Logger,
) inputport.AuthInputPort {
	return &AuthInteractor{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		passwordService: passwordService,
		logger:          logger,
	}
}

// Register は新しいユーザーを登録
func (i *AuthInteractor) Register(ctx context.Context, req *inputport.RegisterRequest) (*inputport.RegisterResponse, error) {
	i.logger.Info("Registering new user", entities.NewField("username", req.Username))

	// パスワードハッシュ化
	hashedPassword, err := i.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// ユーザー作成
	user, err := entities.NewUser(req.Username, req.Email, hashedPassword, req.DisplayName, req.FirstName, req.LastName)
	if err != nil {
		return nil, err
	}

	// ユーザー保存
	if err := i.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// セッション作成
	session, err := entities.NewSession(user.ID, "", "")
	if err != nil {
		return nil, err
	}

	if err := i.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &inputport.RegisterResponse{
		User:    user,
		Session: session,
	}, nil
}

// Login はログイン処理
func (i *AuthInteractor) Login(ctx context.Context, req *inputport.LoginRequest) (*inputport.LoginResponse, error) {
	i.logger.Info("User login attempt", entities.NewField("username", req.Username))

	// ユーザー検索
	user, err := i.userRepo.ReadByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// パスワード検証
	if !i.passwordService.VerifyPassword(user.PasswordHash, req.Password) {
		return nil, errors.New("invalid username or password")
	}

	// アクティブチェック
	if !user.IsActive {
		return nil, errors.New("user account is not active")
	}

	// セッション作成
	session, err := entities.NewSession(user.ID, req.IPAddress, req.UserAgent)
	if err != nil {
		return nil, err
	}

	if err := i.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &inputport.LoginResponse{
		User:    user,
		Session: session,
	}, nil
}

// Logout はログアウト処理
func (i *AuthInteractor) Logout(ctx context.Context, req *inputport.LogoutRequest) error {
	i.logger.Info("User logout", entities.NewField("user_id", req.UserID))
	return i.sessionRepo.DeleteByUserID(ctx, req.UserID)
}

// GetCurrentUser は現在のユーザー情報を取得
func (i *AuthInteractor) GetCurrentUser(ctx context.Context, req *inputport.GetCurrentUserRequest) (*inputport.GetCurrentUserResponse, error) {
	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetCurrentUserResponse{
		User: user,
	}, nil
}

// ValidateSession はセッションを検証
func (i *AuthInteractor) ValidateSession(ctx context.Context, sessionToken string) (*entities.Session, error) {
	session, err := i.sessionRepo.ReadByToken(ctx, sessionToken)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	if session.IsExpired() {
		return nil, errors.New("session expired")
	}

	// セッションをリフレッシュ（並行更新エラーは無視）
	// 複数のリクエストが同時に来た場合、いずれかが成功すれば良い
	session.Refresh()
	if err := i.sessionRepo.Update(ctx, session); err != nil {
		// 並行更新エラーやその他のエラーでも、セッション自体は有効なので認証は継続
		i.logger.Debug("Failed to refresh session (ignoring)", entities.NewField("error", err))
	}

	return session, nil
}
