package interactor

import (
	"errors"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthInteractor は認証のユースケース実装
type AuthInteractor struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	logger      entities.Logger
}

// NewAuthInteractor は新しいAuthInteractorを作成
func NewAuthInteractor(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	logger entities.Logger,
) inputport.AuthInputPort {
	return &AuthInteractor{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

// Register は新しいユーザーを登録
func (i *AuthInteractor) Register(req *inputport.RegisterRequest) (*inputport.RegisterResponse, error) {
	i.logger.Info("Registering new user", entities.NewField("username", req.Username))

	// パスワードハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// ユーザー作成
	user, err := entities.NewUser(req.Username, req.Email, string(hashedPassword), req.DisplayName)
	if err != nil {
		return nil, err
	}

	// ユーザー保存
	if err := i.userRepo.Create(user); err != nil {
		return nil, err
	}

	// セッション作成
	session, err := entities.NewSession(user.ID, "", "")
	if err != nil {
		return nil, err
	}

	if err := i.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return &inputport.RegisterResponse{
		User:    user,
		Session: session,
	}, nil
}

// Login はログイン処理
func (i *AuthInteractor) Login(req *inputport.LoginRequest) (*inputport.LoginResponse, error) {
	i.logger.Info("User login attempt", entities.NewField("username", req.Username))

	// ユーザー検索
	user, err := i.userRepo.ReadByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// パスワード検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
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

	if err := i.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return &inputport.LoginResponse{
		User:    user,
		Session: session,
	}, nil
}

// Logout はログアウト処理
func (i *AuthInteractor) Logout(req *inputport.LogoutRequest) error {
	i.logger.Info("User logout", entities.NewField("user_id", req.UserID))
	return i.sessionRepo.DeleteByUserID(req.UserID)
}

// GetCurrentUser は現在のユーザー情報を取得
func (i *AuthInteractor) GetCurrentUser(req *inputport.GetCurrentUserRequest) (*inputport.GetCurrentUserResponse, error) {
	user, err := i.userRepo.Read(req.UserID)
	if err != nil {
		return nil, err
	}

	return &inputport.GetCurrentUserResponse{
		User: user,
	}, nil
}

// ValidateSession はセッションを検証
func (i *AuthInteractor) ValidateSession(sessionToken string) (*entities.Session, error) {
	session, err := i.sessionRepo.ReadByToken(sessionToken)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	if session.IsExpired() {
		return nil, errors.New("session expired")
	}

	// セッションをリフレッシュ
	session.Refresh()
	if err := i.sessionRepo.Update(session); err != nil {
		i.logger.Warn("Failed to refresh session", entities.NewField("error", err))
	}

	return session, nil
}
