package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name         string
		username     string
		email        string
		passwordHash string
		displayName  string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid user",
			username:     "testuser",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			displayName:  "Test User",
			wantErr:      false,
		},
		{
			name:         "empty username",
			username:     "",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			displayName:  "Test User",
			wantErr:      true,
			errMsg:       "username is required",
		},
		{
			name:         "empty email",
			username:     "testuser",
			email:        "",
			passwordHash: "hashedpassword",
			displayName:  "Test User",
			wantErr:      true,
			errMsg:       "email is required",
		},
		{
			name:         "empty password hash",
			username:     "testuser",
			email:        "test@example.com",
			passwordHash: "",
			displayName:  "Test User",
			wantErr:      true,
			errMsg:       "password hash is required",
		},
		{
			name:         "empty display name",
			username:     "testuser",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			displayName:  "",
			wantErr:      true,
			errMsg:       "display name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.email, tt.passwordHash, tt.displayName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUser() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewUser() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewUser() unexpected error = %v", err)
				return
			}

			// Verify actual values
			if user.Username != tt.username {
				t.Errorf("Username = %v, want %v", user.Username, tt.username)
			}
			if user.Email != tt.email {
				t.Errorf("Email = %v, want %v", user.Email, tt.email)
			}
			if user.PasswordHash != tt.passwordHash {
				t.Errorf("PasswordHash = %v, want %v", user.PasswordHash, tt.passwordHash)
			}
			if user.DisplayName != tt.displayName {
				t.Errorf("DisplayName = %v, want %v", user.DisplayName, tt.displayName)
			}
			if user.Balance != 0 {
				t.Errorf("Balance = %v, want 0", user.Balance)
			}
			if user.Role != RoleUser {
				t.Errorf("Role = %v, want %v", user.Role, RoleUser)
			}
			if user.Version != 1 {
				t.Errorf("Version = %v, want 1", user.Version)
			}
			if !user.IsActive {
				t.Errorf("IsActive = %v, want true", user.IsActive)
			}
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{
			name: "admin role",
			role: RoleAdmin,
			want: true,
		},
		{
			name: "user role",
			role: RoleUser,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			if got := user.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanTransfer(t *testing.T) {
	tests := []struct {
		name     string
		balance  int64
		isActive bool
		amount   int64
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "sufficient balance",
			balance:  1000,
			isActive: true,
			amount:   500,
			wantErr:  false,
		},
		{
			name:     "exact balance",
			balance:  1000,
			isActive: true,
			amount:   1000,
			wantErr:  false,
		},
		{
			name:     "insufficient balance",
			balance:  1000,
			isActive: true,
			amount:   1001,
			wantErr:  true,
			errMsg:   "insufficient balance",
		},
		{
			name:     "inactive user",
			balance:  1000,
			isActive: false,
			amount:   500,
			wantErr:  true,
			errMsg:   "user is not active",
		},
		{
			name:     "zero amount",
			balance:  1000,
			isActive: true,
			amount:   0,
			wantErr:  true,
			errMsg:   "amount must be positive",
		},
		{
			name:     "negative amount",
			balance:  1000,
			isActive: true,
			amount:   -100,
			wantErr:  true,
			errMsg:   "amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Balance:  tt.balance,
				IsActive: tt.isActive,
			}

			err := user.CanTransfer(tt.amount)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CanTransfer() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("CanTransfer() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("CanTransfer() unexpected error = %v", err)
			}
		})
	}
}

func TestUser_Deduct(t *testing.T) {
	tests := []struct {
		name            string
		initialBalance  int64
		initialVersion  int
		isActive        bool
		amount          int64
		wantErr         bool
		errMsg          string
		expectedBalance int64
		expectedVersion int
	}{
		{
			name:            "successful deduction",
			initialBalance:  1000,
			initialVersion:  1,
			isActive:        true,
			amount:          300,
			wantErr:         false,
			expectedBalance: 700,
			expectedVersion: 2,
		},
		{
			name:            "deduct all balance",
			initialBalance:  1000,
			initialVersion:  1,
			isActive:        true,
			amount:          1000,
			wantErr:         false,
			expectedBalance: 0,
			expectedVersion: 2,
		},
		{
			name:           "insufficient balance",
			initialBalance: 1000,
			initialVersion: 1,
			isActive:       true,
			amount:         1001,
			wantErr:        true,
			errMsg:         "insufficient balance",
		},
		{
			name:           "inactive user",
			initialBalance: 1000,
			initialVersion: 1,
			isActive:       false,
			amount:         300,
			wantErr:        true,
			errMsg:         "user is not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Balance:  tt.initialBalance,
				Version:  tt.initialVersion,
				IsActive: tt.isActive,
			}

			err := user.Deduct(tt.amount)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Deduct() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Deduct() error = %v, want %v", err.Error(), tt.errMsg)
				}
				// Verify balance unchanged on error
				if user.Balance != tt.initialBalance {
					t.Errorf("Balance changed on error: got %v, want %v", user.Balance, tt.initialBalance)
				}
				return
			}

			if err != nil {
				t.Errorf("Deduct() unexpected error = %v", err)
				return
			}

			// Verify actual balance after deduction
			if user.Balance != tt.expectedBalance {
				t.Errorf("Balance = %v, want %v", user.Balance, tt.expectedBalance)
			}

			// Verify version incremented
			if user.Version != tt.expectedVersion {
				t.Errorf("Version = %v, want %v", user.Version, tt.expectedVersion)
			}
		})
	}
}

func TestUser_Add(t *testing.T) {
	tests := []struct {
		name            string
		initialBalance  int64
		initialVersion  int
		amount          int64
		wantErr         bool
		errMsg          string
		expectedBalance int64
		expectedVersion int
	}{
		{
			name:            "successful addition",
			initialBalance:  1000,
			initialVersion:  1,
			amount:          500,
			wantErr:         false,
			expectedBalance: 1500,
			expectedVersion: 2,
		},
		{
			name:            "add to zero balance",
			initialBalance:  0,
			initialVersion:  1,
			amount:          1000,
			wantErr:         false,
			expectedBalance: 1000,
			expectedVersion: 2,
		},
		{
			name:           "zero amount",
			initialBalance: 1000,
			initialVersion: 1,
			amount:         0,
			wantErr:        true,
			errMsg:         "amount must be positive",
		},
		{
			name:           "negative amount",
			initialBalance: 1000,
			initialVersion: 1,
			amount:         -100,
			wantErr:        true,
			errMsg:         "amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Balance: tt.initialBalance,
				Version: tt.initialVersion,
			}

			err := user.Add(tt.amount)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Add() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Add() error = %v, want %v", err.Error(), tt.errMsg)
				}
				// Verify balance unchanged on error
				if user.Balance != tt.initialBalance {
					t.Errorf("Balance changed on error: got %v, want %v", user.Balance, tt.initialBalance)
				}
				return
			}

			if err != nil {
				t.Errorf("Add() unexpected error = %v", err)
				return
			}

			// Verify actual balance after addition
			if user.Balance != tt.expectedBalance {
				t.Errorf("Balance = %v, want %v", user.Balance, tt.expectedBalance)
			}

			// Verify version incremented
			if user.Version != tt.expectedVersion {
				t.Errorf("Version = %v, want %v", user.Version, tt.expectedVersion)
			}
		})
	}
}

func TestUser_UpdateRole(t *testing.T) {
	tests := []struct {
		name            string
		initialRole     UserRole
		initialVersion  int
		newRole         UserRole
		wantErr         bool
		errMsg          string
		expectedRole    UserRole
		expectedVersion int
	}{
		{
			name:            "user to admin",
			initialRole:     RoleUser,
			initialVersion:  1,
			newRole:         RoleAdmin,
			wantErr:         false,
			expectedRole:    RoleAdmin,
			expectedVersion: 2,
		},
		{
			name:            "admin to user",
			initialRole:     RoleAdmin,
			initialVersion:  1,
			newRole:         RoleUser,
			wantErr:         false,
			expectedRole:    RoleUser,
			expectedVersion: 2,
		},
		{
			name:           "invalid role",
			initialRole:    RoleUser,
			initialVersion: 1,
			newRole:        UserRole("invalid"),
			wantErr:        true,
			errMsg:         "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Role:    tt.initialRole,
				Version: tt.initialVersion,
			}

			err := user.UpdateRole(tt.newRole)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateRole() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("UpdateRole() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateRole() unexpected error = %v", err)
				return
			}

			// Verify actual role updated
			if user.Role != tt.expectedRole {
				t.Errorf("Role = %v, want %v", user.Role, tt.expectedRole)
			}

			// Verify version incremented
			if user.Version != tt.expectedVersion {
				t.Errorf("Version = %v, want %v", user.Version, tt.expectedVersion)
			}
		})
	}
}

func TestUser_DeactivateActivate(t *testing.T) {
	t.Run("deactivate user", func(t *testing.T) {
		user := &User{
			ID:       uuid.New(),
			IsActive: true,
			Version:  1,
		}

		user.Deactivate()

		if user.IsActive {
			t.Errorf("IsActive = true, want false")
		}
		if user.Version != 2 {
			t.Errorf("Version = %v, want 2", user.Version)
		}
	})

	t.Run("activate user", func(t *testing.T) {
		user := &User{
			ID:       uuid.New(),
			IsActive: false,
			Version:  1,
		}

		user.Activate()

		if !user.IsActive {
			t.Errorf("IsActive = false, want true")
		}
		if user.Version != 2 {
			t.Errorf("Version = %v, want 2", user.Version)
		}
	})

	t.Run("updated_at changes on deactivate", func(t *testing.T) {
		user := &User{
			ID:        uuid.New(),
			IsActive:  true,
			Version:   1,
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}
		beforeUpdate := user.UpdatedAt

		time.Sleep(10 * time.Millisecond)
		user.Deactivate()

		if !user.UpdatedAt.After(beforeUpdate) {
			t.Errorf("UpdatedAt not updated: before=%v, after=%v", beforeUpdate, user.UpdatedAt)
		}
	})
}
