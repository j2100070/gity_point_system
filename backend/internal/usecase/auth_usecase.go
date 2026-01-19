package usecase

import (
	"errors"

	"github.com/gity/point-system/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase は認証に関するユースケース
type AuthUseCase struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
}

// NewAuthUseCase は新しいAuthUseCaseを作成
func NewAuthUseCase(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// RegisterRequest は登録リクエスト
type RegisterRequest struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
}

// RegisterResponse は登録レスポンス
type RegisterResponse struct {
	User *domain.User
}

// Register はユーザーを登録
// セキュリティ:
//   - パスワードはbcryptでハッシュ化（cost=12）
//   - ユーザー名とメールの重複チェック
func (u *AuthUseCase) Register(req *RegisterRequest) (*RegisterResponse, error) {
	// バリデーション
	if req.Username == "" || req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return nil, errors.New("all fields are required")
	}

	// パスワードの強度チェック
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// ユーザー名の重複チェック
	existingUser, _ := u.userRepo.FindByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// メールアドレスの重複チェック
	existingEmail, _ := u.userRepo.FindByEmail(req.Email)
	if existingEmail != nil {
		return nil, errors.New("email already exists")
	}

	// パスワードのハッシュ化（bcrypt cost=12）
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	// ユーザー作成
	user, err := domain.NewUser(req.Username, req.Email, string(passwordHash), req.DisplayName)
	if err != nil {
		return nil, err
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}

	return &RegisterResponse{User: user}, nil
}

// LoginRequest はログインリクエスト
type LoginRequest struct {
	Username  string
	Password  string
	IPAddress string
	UserAgent string
}

// LoginResponse はログインレスポンス
type LoginResponse struct {
	User         *domain.User
	Session      *domain.Session
	SessionToken string
	CSRFToken    string
}

// Login はユーザーをログイン
// セキュリティ:
//   - bcryptでパスワード検証
//   - セッショントークン生成（ランダム32バイト）
//   - CSRFトークン生成（ランダム32バイト）
func (u *AuthUseCase) Login(req *LoginRequest) (*LoginResponse, error) {
	// バリデーション
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("username and password are required")
	}

	// ユーザー検索
	user, err := u.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// アカウント状態チェック
	if !user.IsActive {
		return nil, errors.New("account is not active")
	}

	// パスワード検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// セッション作成
	session, err := domain.NewSession(user.ID, req.IPAddress, req.UserAgent)
	if err != nil {
		return nil, err
	}

	if err := u.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return &LoginResponse{
		User:         user,
		Session:      session,
		SessionToken: session.SessionToken,
		CSRFToken:    session.CSRFToken,
	}, nil
}

// LogoutRequest はログアウトリクエスト
type LogoutRequest struct {
	SessionToken string
}

// Logout はユーザーをログアウト
func (u *AuthUseCase) Logout(req *LogoutRequest) error {
	session, err := u.sessionRepo.FindByToken(req.SessionToken)
	if err != nil {
		return err
	}

	return u.sessionRepo.Delete(session.ID)
}

// ValidateSessionRequest はセッション検証リクエスト
type ValidateSessionRequest struct {
	SessionToken string
	CSRFToken    string // CSRFトークン（POST/PUT/DELETE時のみ）
}

// ValidateSessionResponse はセッション検証レスポンス
type ValidateSessionResponse struct {
	User    *domain.User
	Session *domain.Session
}

// ValidateSession はセッションを検証
// セキュリティ:
//   - セッションの有効期限チェック
//   - CSRFトークン検証（必要な場合）
func (u *AuthUseCase) ValidateSession(req *ValidateSessionRequest) (*ValidateSessionResponse, error) {
	// セッション検索
	session, err := u.sessionRepo.FindByToken(req.SessionToken)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	// 有効期限チェック
	if session.IsExpired() {
		u.sessionRepo.Delete(session.ID)
		return nil, errors.New("session expired")
	}

	// CSRFトークン検証（指定がある場合）
	if req.CSRFToken != "" {
		if err := session.ValidateCSRF(req.CSRFToken); err != nil {
			return nil, err
		}
	}

	// ユーザー検索
	user, err := u.userRepo.FindByID(session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// アカウント状態チェック
	if !user.IsActive {
		return nil, errors.New("account is not active")
	}

	// セッションの有効期限を延長
	session.Refresh()
	u.sessionRepo.Update(session)

	return &ValidateSessionResponse{
		User:    user,
		Session: session,
	}, nil
}

// GetCurrentUserRequest は現在のユーザー取得リクエスト
type GetCurrentUserRequest struct {
	SessionToken string
}

// GetCurrentUserResponse は現在のユーザー取得レスポンス
type GetCurrentUserResponse struct {
	User *domain.User
}

// GetCurrentUser は現在のユーザーを取得
func (u *AuthUseCase) GetCurrentUser(req *GetCurrentUserRequest) (*GetCurrentUserResponse, error) {
	resp, err := u.ValidateSession(&ValidateSessionRequest{
		SessionToken: req.SessionToken,
	})
	if err != nil {
		return nil, err
	}

	return &GetCurrentUserResponse{User: resp.User}, nil
}
