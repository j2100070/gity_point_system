package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/gity/point-system/internal/domain"
	"github.com/gity/point-system/internal/domain/mock"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSessionRepo := mock.NewMockSessionRepository(ctrl)
	authUseCase := NewAuthUseCase(mockUserRepo, mockSessionRepo)

	t.Run("successful registration", func(t *testing.T) {
		req := &RegisterRequest{
			Username:    "testuser",
			Email:       "test@example.com",
			Password:    "SecurePass123",
			DisplayName: "Test User",
		}

		// Expectations: no existing user with same username or email
		mockUserRepo.EXPECT().
			FindByUsername(req.Username).
			Return(nil, errors.New("not found"))

		mockUserRepo.EXPECT().
			FindByEmail(req.Email).
			Return(nil, errors.New("not found"))

		mockUserRepo.EXPECT().
			Create(gomock.Any()).
			DoAndReturn(func(user *domain.User) error {
				// Verify actual values
				if user.Username != req.Username {
					t.Errorf("Username = %v, want %v", user.Username, req.Username)
				}
				if user.Email != req.Email {
					t.Errorf("Email = %v, want %v", user.Email, req.Email)
				}
				if user.DisplayName != req.DisplayName {
					t.Errorf("DisplayName = %v, want %v", user.DisplayName, req.DisplayName)
				}
				if user.Balance != 0 {
					t.Errorf("Initial balance = %v, want 0", user.Balance)
				}
				if user.Role != domain.RoleUser {
					t.Errorf("Role = %v, want %v", user.Role, domain.RoleUser)
				}
				// Verify password is hashed
				err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
				if err != nil {
					t.Errorf("Password hash verification failed: %v", err)
				}
				return nil
			})

		resp, err := authUseCase.Register(req)

		if err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		if resp.User == nil {
			t.Errorf("User is nil, want user object")
		}
	})

	t.Run("registration with duplicate username", func(t *testing.T) {
		req := &RegisterRequest{
			Username:    "existinguser",
			Email:       "new@example.com",
			Password:    "SecurePass123",
			DisplayName: "Test User",
		}

		existingUser := &domain.User{
			ID:       uuid.New(),
			Username: "existinguser",
		}

		mockUserRepo.EXPECT().
			FindByUsername(req.Username).
			Return(existingUser, nil)

		_, err := authUseCase.Register(req)

		if err == nil {
			t.Errorf("Register() expected error for duplicate username, got nil")
		}
		if err.Error() != "username already exists" {
			t.Errorf("Register() error = %v, want 'username already exists'", err.Error())
		}
	})

	t.Run("registration with duplicate email", func(t *testing.T) {
		req := &RegisterRequest{
			Username:    "newuser",
			Email:       "existing@example.com",
			Password:    "SecurePass123",
			DisplayName: "Test User",
		}

		existingUser := &domain.User{
			ID:    uuid.New(),
			Email: "existing@example.com",
		}

		mockUserRepo.EXPECT().
			FindByUsername(req.Username).
			Return(nil, errors.New("not found"))

		mockUserRepo.EXPECT().
			FindByEmail(req.Email).
			Return(existingUser, nil)

		_, err := authUseCase.Register(req)

		if err == nil {
			t.Errorf("Register() expected error for duplicate email, got nil")
		}
		if err.Error() != "email already exists" {
			t.Errorf("Register() error = %v, want 'email already exists'", err.Error())
		}
	})

	t.Run("registration with weak password", func(t *testing.T) {
		req := &RegisterRequest{
			Username:    "testuser",
			Email:       "test@example.com",
			Password:    "short", // Less than 8 characters
			DisplayName: "Test User",
		}

		_, err := authUseCase.Register(req)

		if err == nil {
			t.Errorf("Register() expected error for weak password, got nil")
		}
		if err.Error() != "password must be at least 8 characters" {
			t.Errorf("Register() error = %v, want 'password must be at least 8 characters'", err.Error())
		}
	})

	t.Run("registration with empty fields", func(t *testing.T) {
		tests := []struct {
			name string
			req  *RegisterRequest
		}{
			{
				name: "empty username",
				req: &RegisterRequest{
					Username:    "",
					Email:       "test@example.com",
					Password:    "SecurePass123",
					DisplayName: "Test User",
				},
			},
			{
				name: "empty email",
				req: &RegisterRequest{
					Username:    "testuser",
					Email:       "",
					Password:    "SecurePass123",
					DisplayName: "Test User",
				},
			},
			{
				name: "empty password",
				req: &RegisterRequest{
					Username:    "testuser",
					Email:       "test@example.com",
					Password:    "",
					DisplayName: "Test User",
				},
			},
			{
				name: "empty display name",
				req: &RegisterRequest{
					Username:    "testuser",
					Email:       "test@example.com",
					Password:    "SecurePass123",
					DisplayName: "",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := authUseCase.Register(tt.req)

				if err == nil {
					t.Errorf("Register() expected error for %s, got nil", tt.name)
				}
				if err.Error() != "all fields are required" {
					t.Errorf("Register() error = %v, want 'all fields are required'", err.Error())
				}
			})
		}
	})
}

func TestAuthUseCase_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSessionRepo := mock.NewMockSessionRepository(ctrl)
	authUseCase := NewAuthUseCase(mockUserRepo, mockSessionRepo)

	t.Run("successful login", func(t *testing.T) {
		password := "SecurePass123"
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

		user := &domain.User{
			ID:           uuid.New(),
			Username:     "testuser",
			PasswordHash: string(passwordHash),
			IsActive:     true,
			Balance:      5000,
		}

		req := &LoginRequest{
			Username:  "testuser",
			Password:  password,
			IPAddress: "127.0.0.1",
			UserAgent: "Test Agent",
		}

		mockUserRepo.EXPECT().
			FindByUsername(req.Username).
			Return(user, nil)

		mockSessionRepo.EXPECT().
			Create(gomock.Any()).
			DoAndReturn(func(session *domain.Session) error {
				// Verify actual session values
				if session.UserID != user.ID {
					t.Errorf("Session UserID = %v, want %v", session.UserID, user.ID)
				}
				if session.IPAddress != req.IPAddress {
					t.Errorf("Session IPAddress = %v, want %v", session.IPAddress, req.IPAddress)
				}
				if session.UserAgent != req.UserAgent {
					t.Errorf("Session UserAgent = %v, want %v", session.UserAgent, req.UserAgent)
				}
				if session.SessionToken == "" {
					t.Errorf("SessionToken is empty, want generated token")
				}
				if session.CSRFToken == "" {
					t.Errorf("CSRFToken is empty, want generated token")
				}
				return nil
			})

		resp, err := authUseCase.Login(req)

		if err != nil {
			t.Fatalf("Login() unexpected error = %v", err)
		}

		// Verify actual response values
		if resp.User.ID != user.ID {
			t.Errorf("User ID = %v, want %v", resp.User.ID, user.ID)
		}
		if resp.User.Balance != 5000 {
			t.Errorf("User Balance = %v, want 5000", resp.User.Balance)
		}
		if resp.SessionToken == "" {
			t.Errorf("SessionToken is empty, want token")
		}
		if resp.CSRFToken == "" {
			t.Errorf("CSRFToken is empty, want token")
		}
	})

	t.Run("login with invalid username", func(t *testing.T) {
		req := &LoginRequest{
			Username:  "nonexistent",
			Password:  "SecurePass123",
			IPAddress: "127.0.0.1",
			UserAgent: "Test Agent",
		}

		mockUserRepo.EXPECT().
			FindByUsername(req.Username).
			Return(nil, errors.New("not found"))

		_, err := authUseCase.Login(req)

		if err == nil {
			t.Errorf("Login() expected error for invalid username, got nil")
		}
		if err.Error() != "invalid credentials" {
			t.Errorf("Login() error = %v, want 'invalid credentials'", err.Error())
		}
	})

	t.Run("login with invalid password", func(t *testing.T) {
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass123"), 12)

		user := &domain.User{
			ID:           uuid.New(),
			Username:     "testuser",
			PasswordHash: string(passwordHash),
			IsActive:     true,
		}

		req := &LoginRequest{
			Username:  "testuser",
			Password:  "WrongPassword",
			IPAddress: "127.0.0.1",
			UserAgent: "Test Agent",
		}

		mockUserRepo.EXPECT().
			FindByUsername(req.Username).
			Return(user, nil)

		_, err := authUseCase.Login(req)

		if err == nil {
			t.Errorf("Login() expected error for invalid password, got nil")
		}
		if err.Error() != "invalid credentials" {
			t.Errorf("Login() error = %v, want 'invalid credentials'", err.Error())
		}
	})

	t.Run("login with inactive account", func(t *testing.T) {
		password := "SecurePass123"
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

		user := &domain.User{
			ID:           uuid.New(),
			Username:     "testuser",
			PasswordHash: string(passwordHash),
			IsActive:     false, // Account is inactive
		}

		req := &LoginRequest{
			Username:  "testuser",
			Password:  password,
			IPAddress: "127.0.0.1",
			UserAgent: "Test Agent",
		}

		mockUserRepo.EXPECT().
			FindByUsername(req.Username).
			Return(user, nil)

		_, err := authUseCase.Login(req)

		if err == nil {
			t.Errorf("Login() expected error for inactive account, got nil")
		}
		if err.Error() != "account is not active" {
			t.Errorf("Login() error = %v, want 'account is not active'", err.Error())
		}
	})
}

func TestAuthUseCase_ValidateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSessionRepo := mock.NewMockSessionRepository(ctrl)
	authUseCase := NewAuthUseCase(mockUserRepo, mockSessionRepo)

	t.Run("successful session validation", func(t *testing.T) {
		userID := uuid.New()
		user := &domain.User{
			ID:       userID,
			Username: "testuser",
			IsActive: true,
			Balance:  10000,
		}

		session := &domain.Session{
			ID:           uuid.New(),
			UserID:       userID,
			SessionToken: "valid-token",
			CSRFToken:    "valid-csrf",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
		}

		req := &ValidateSessionRequest{
			SessionToken: "valid-token",
		}

		mockSessionRepo.EXPECT().
			FindByToken(req.SessionToken).
			Return(session, nil)

		mockUserRepo.EXPECT().
			FindByID(userID).
			Return(user, nil)

		mockSessionRepo.EXPECT().
			Update(gomock.Any()).
			Return(nil)

		resp, err := authUseCase.ValidateSession(req)

		if err != nil {
			t.Fatalf("ValidateSession() unexpected error = %v", err)
		}

		// Verify actual values
		if resp.User.ID != userID {
			t.Errorf("User ID = %v, want %v", resp.User.ID, userID)
		}
		if resp.User.Balance != 10000 {
			t.Errorf("User Balance = %v, want 10000", resp.User.Balance)
		}
		if resp.Session.UserID != userID {
			t.Errorf("Session UserID = %v, want %v", resp.Session.UserID, userID)
		}
	})

	t.Run("validation with CSRF token", func(t *testing.T) {
		userID := uuid.New()
		user := &domain.User{
			ID:       userID,
			Username: "testuser",
			IsActive: true,
		}

		session := &domain.Session{
			ID:           uuid.New(),
			UserID:       userID,
			SessionToken: "valid-token",
			CSRFToken:    "valid-csrf",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
		}

		req := &ValidateSessionRequest{
			SessionToken: "valid-token",
			CSRFToken:    "valid-csrf",
		}

		mockSessionRepo.EXPECT().
			FindByToken(req.SessionToken).
			Return(session, nil)

		mockUserRepo.EXPECT().
			FindByID(userID).
			Return(user, nil)

		mockSessionRepo.EXPECT().
			Update(gomock.Any()).
			Return(nil)

		_, err := authUseCase.ValidateSession(req)

		if err != nil {
			t.Fatalf("ValidateSession() unexpected error = %v", err)
		}
	})

	t.Run("validation with invalid CSRF token", func(t *testing.T) {
		userID := uuid.New()

		session := &domain.Session{
			ID:           uuid.New(),
			UserID:       userID,
			SessionToken: "valid-token",
			CSRFToken:    "valid-csrf",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
		}

		req := &ValidateSessionRequest{
			SessionToken: "valid-token",
			CSRFToken:    "invalid-csrf",
		}

		mockSessionRepo.EXPECT().
			FindByToken(req.SessionToken).
			Return(session, nil)

		_, err := authUseCase.ValidateSession(req)

		if err == nil {
			t.Errorf("ValidateSession() expected error for invalid CSRF, got nil")
		}
		if err.Error() != "invalid csrf token" {
			t.Errorf("ValidateSession() error = %v, want 'invalid csrf token'", err.Error())
		}
	})

	t.Run("validation with expired session", func(t *testing.T) {
		userID := uuid.New()

		session := &domain.Session{
			ID:           uuid.New(),
			UserID:       userID,
			SessionToken: "expired-token",
			CSRFToken:    "csrf-token",
			ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
			CreatedAt:    time.Now().Add(-25 * time.Hour),
		}

		req := &ValidateSessionRequest{
			SessionToken: "expired-token",
		}

		mockSessionRepo.EXPECT().
			FindByToken(req.SessionToken).
			Return(session, nil)

		mockSessionRepo.EXPECT().
			Delete(session.ID).
			Return(nil)

		_, err := authUseCase.ValidateSession(req)

		if err == nil {
			t.Errorf("ValidateSession() expected error for expired session, got nil")
		}
		if err.Error() != "session expired" {
			t.Errorf("ValidateSession() error = %v, want 'session expired'", err.Error())
		}
	})

	t.Run("validation with inactive user", func(t *testing.T) {
		userID := uuid.New()
		user := &domain.User{
			ID:       userID,
			Username: "testuser",
			IsActive: false, // Inactive
		}

		session := &domain.Session{
			ID:           uuid.New(),
			UserID:       userID,
			SessionToken: "valid-token",
			CSRFToken:    "valid-csrf",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
		}

		req := &ValidateSessionRequest{
			SessionToken: "valid-token",
		}

		mockSessionRepo.EXPECT().
			FindByToken(req.SessionToken).
			Return(session, nil)

		mockUserRepo.EXPECT().
			FindByID(userID).
			Return(user, nil)

		_, err := authUseCase.ValidateSession(req)

		if err == nil {
			t.Errorf("ValidateSession() expected error for inactive user, got nil")
		}
		if err.Error() != "account is not active" {
			t.Errorf("ValidateSession() error = %v, want 'account is not active'", err.Error())
		}
	})
}

func TestAuthUseCase_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockSessionRepo := mock.NewMockSessionRepository(ctrl)
	authUseCase := NewAuthUseCase(mockUserRepo, mockSessionRepo)

	t.Run("successful logout", func(t *testing.T) {
		sessionID := uuid.New()
		session := &domain.Session{
			ID:           sessionID,
			SessionToken: "valid-token",
		}

		req := &LogoutRequest{
			SessionToken: "valid-token",
		}

		mockSessionRepo.EXPECT().
			FindByToken(req.SessionToken).
			Return(session, nil)

		mockSessionRepo.EXPECT().
			Delete(sessionID).
			Return(nil)

		err := authUseCase.Logout(req)

		if err != nil {
			t.Fatalf("Logout() unexpected error = %v", err)
		}
	})

	t.Run("logout with invalid token", func(t *testing.T) {
		req := &LogoutRequest{
			SessionToken: "invalid-token",
		}

		mockSessionRepo.EXPECT().
			FindByToken(req.SessionToken).
			Return(nil, errors.New("session not found"))

		err := authUseCase.Logout(req)

		if err == nil {
			t.Errorf("Logout() expected error for invalid token, got nil")
		}
	})
}
