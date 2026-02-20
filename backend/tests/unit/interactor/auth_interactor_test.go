package interactor_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// AuthInteractor テスト
// ========================================

// --- Mock SessionRepository ---

type mockSessionRepo struct {
	sessions  map[string]*entities.Session
	createErr error
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{sessions: make(map[string]*entities.Session)}
}

func (m *mockSessionRepo) Create(ctx context.Context, session *entities.Session) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.sessions[session.SessionToken] = session
	return nil
}
func (m *mockSessionRepo) ReadByToken(ctx context.Context, token string) (*entities.Session, error) {
	s, ok := m.sessions[token]
	if !ok {
		return nil, errors.New("session not found")
	}
	return s, nil
}
func (m *mockSessionRepo) Update(ctx context.Context, session *entities.Session) error {
	m.sessions[session.SessionToken] = session
	return nil
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockSessionRepo) DeleteExpired(ctx context.Context) error { return nil }

// --- Mock PasswordService ---

type mockPasswordService struct {
	hashResult string
	hashErr    error
	verifyOK   bool
}

func (m *mockPasswordService) HashPassword(password string) (string, error) {
	if m.hashErr != nil {
		return "", m.hashErr
	}
	if m.hashResult != "" {
		return m.hashResult, nil
	}
	return "hashed_" + password, nil
}

func (m *mockPasswordService) VerifyPassword(hashedPassword, password string) bool {
	return m.verifyOK
}

// --- Register ---

func TestAuthInteractor_Register(t *testing.T) {
	setup := func() (*ctxTrackingUserRepo, *mockSessionRepo, *mockPasswordService, inputport.AuthInputPort) {
		userRepo := newCtxTrackingUserRepo()
		sessionRepo := newMockSessionRepo()
		pwService := &mockPasswordService{verifyOK: true}
		logger := &mockLogger{}

		sut := interactor.NewAuthInteractor(userRepo, sessionRepo, pwService, logger)
		return userRepo, sessionRepo, pwService, sut
	}

	t.Run("正常にユーザー登録できる", func(t *testing.T) {
		_, _, _, sut := setup()

		resp, err := sut.Register(context.Background(), &inputport.RegisterRequest{
			Username: "testuser", Email: "test@example.com",
			Password: "password123", DisplayName: "Test User",
			FirstName: "太郎", LastName: "田中",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.User)
		assert.NotNil(t, resp.Session)
		assert.Equal(t, "testuser", resp.User.Username)
	})

	t.Run("パスワードハッシュ化に失敗した場合エラー", func(t *testing.T) {
		_, _, pwService, sut := setup()
		pwService.hashErr = errors.New("hash error")

		_, err := sut.Register(context.Background(), &inputport.RegisterRequest{
			Username: "testuser", Email: "test@example.com",
			Password: "password123", DisplayName: "Test User",
			FirstName: "太郎", LastName: "田中",
		})
		assert.Error(t, err)
	})
}

// --- Login ---

func TestAuthInteractor_Login(t *testing.T) {
	setup := func() (*ctxTrackingUserRepo, *mockSessionRepo, *mockPasswordService, inputport.AuthInputPort) {
		userRepo := newCtxTrackingUserRepo()
		sessionRepo := newMockSessionRepo()
		pwService := &mockPasswordService{verifyOK: true}
		logger := &mockLogger{}

		sut := interactor.NewAuthInteractor(userRepo, sessionRepo, pwService, logger)
		return userRepo, sessionRepo, pwService, sut
	}

	t.Run("正常にログインできる", func(t *testing.T) {
		userRepo, _, _, sut := setup()
		user := createTestUserWithBalance(t, "loginuser", 0, "user")
		userRepo.setUser(user)

		resp, err := sut.Login(context.Background(), &inputport.LoginRequest{
			Username: user.Username, Password: "password123",
			IPAddress: "127.0.0.1", UserAgent: "TestAgent",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.User)
		assert.NotNil(t, resp.Session)
	})

	t.Run("ユーザーが存在しない場合エラー", func(t *testing.T) {
		_, _, _, sut := setup()

		_, err := sut.Login(context.Background(), &inputport.LoginRequest{
			Username: "nonexistent", Password: "password123",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid username or password")
	})

	t.Run("パスワードが不正な場合エラー", func(t *testing.T) {
		userRepo, _, pwService, sut := setup()
		pwService.verifyOK = false
		user := createTestUserWithBalance(t, "loginuser", 0, "user")
		userRepo.setUser(user)

		_, err := sut.Login(context.Background(), &inputport.LoginRequest{
			Username: user.Username, Password: "wrongpassword",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid username or password")
	})

	t.Run("非アクティブユーザーはログインできない", func(t *testing.T) {
		userRepo, _, _, sut := setup()
		user := createTestUserWithBalance(t, "inactive", 0, "user")
		user.IsActive = false
		userRepo.setUser(user)

		_, err := sut.Login(context.Background(), &inputport.LoginRequest{
			Username: user.Username, Password: "password123",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})
}

// --- Logout ---

func TestAuthInteractor_Logout(t *testing.T) {
	t.Run("正常にログアウトできる", func(t *testing.T) {
		sut := interactor.NewAuthInteractor(
			newCtxTrackingUserRepo(), newMockSessionRepo(),
			&mockPasswordService{}, &mockLogger{},
		)
		err := sut.Logout(context.Background(), &inputport.LogoutRequest{
			UserID: uuid.New(),
		})
		assert.NoError(t, err)
	})
}

// --- GetCurrentUser ---

func TestAuthInteractor_GetCurrentUser(t *testing.T) {
	t.Run("正常にユーザー情報を取得できる", func(t *testing.T) {
		userRepo := newCtxTrackingUserRepo()
		sut := interactor.NewAuthInteractor(
			userRepo, newMockSessionRepo(),
			&mockPasswordService{}, &mockLogger{},
		)
		user := createTestUserWithBalance(t, "currentuser", 1000, "user")
		userRepo.setUser(user)

		resp, err := sut.GetCurrentUser(context.Background(), &inputport.GetCurrentUserRequest{
			UserID: user.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, user.ID, resp.User.ID)
	})

	t.Run("ユーザーが存在しない場合エラー", func(t *testing.T) {
		sut := interactor.NewAuthInteractor(
			newCtxTrackingUserRepo(), newMockSessionRepo(),
			&mockPasswordService{}, &mockLogger{},
		)
		_, err := sut.GetCurrentUser(context.Background(), &inputport.GetCurrentUserRequest{
			UserID: uuid.New(),
		})
		assert.Error(t, err)
	})
}

// --- ValidateSession ---

func TestAuthInteractor_ValidateSession(t *testing.T) {
	t.Run("正常にセッションを検証できる", func(t *testing.T) {
		sessionRepo := newMockSessionRepo()
		sut := interactor.NewAuthInteractor(
			newCtxTrackingUserRepo(), sessionRepo,
			&mockPasswordService{}, &mockLogger{},
		)

		session, err := entities.NewSession(uuid.New(), "127.0.0.1", "TestAgent")
		require.NoError(t, err)
		sessionRepo.sessions[session.SessionToken] = session

		result, err := sut.ValidateSession(context.Background(), session.SessionToken)
		require.NoError(t, err)
		assert.Equal(t, session.ID, result.ID)
	})

	t.Run("存在しないセッションの場合エラー", func(t *testing.T) {
		sut := interactor.NewAuthInteractor(
			newCtxTrackingUserRepo(), newMockSessionRepo(),
			&mockPasswordService{}, &mockLogger{},
		)

		_, err := sut.ValidateSession(context.Background(), "invalid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid session")
	})

	t.Run("期限切れセッションの場合エラー", func(t *testing.T) {
		sessionRepo := newMockSessionRepo()
		sut := interactor.NewAuthInteractor(
			newCtxTrackingUserRepo(), sessionRepo,
			&mockPasswordService{}, &mockLogger{},
		)

		session, err := entities.NewSession(uuid.New(), "127.0.0.1", "TestAgent")
		require.NoError(t, err)
		session.ExpiresAt = time.Now().Add(-1 * time.Hour)
		sessionRepo.sessions[session.SessionToken] = session

		_, err = sut.ValidateSession(context.Background(), session.SessionToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session expired")
	})
}
