// +build integration

package persistence

import (
	"testing"
	"time"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=gity_point_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate
	err = db.AutoMigrate(
		&UserModel{},
		&TransactionModel{},
		&IdempotencyKeyModel{},
		&SessionModel{},
		&QRCodeModel{},
		&FriendshipModel{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

// cleanupTestDB cleans up test data
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE users, transactions, idempotency_keys, sessions, qr_codes, friendships CASCADE")
}

func TestUserRepository_Create_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Test User",
		Balance:      1000,
		Role:         domain.RoleUser,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(user)

	if err != nil {
		t.Fatalf("Create() unexpected error = %v", err)
	}

	// Verify actual values in database
	found, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID() unexpected error = %v", err)
	}

	if found.Username != user.Username {
		t.Errorf("Username = %v, want %v", found.Username, user.Username)
	}
	if found.Balance != 1000 {
		t.Errorf("Balance = %v, want 1000", found.Balance)
	}
	if found.Version != 1 {
		t.Errorf("Version = %v, want 1", found.Version)
	}
}

func TestUserRepository_UpdateBalanceWithLock_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create test user with initial balance
	initialBalance := int64(5000)
	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Test User",
		Balance:      initialBalance,
		Role:         domain.RoleUser,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create() unexpected error = %v", err)
	}

	t.Run("deduct balance with actual value verification", func(t *testing.T) {
		deductAmount := int64(1000)

		err := db.Transaction(func(tx *gorm.DB) error {
			return repo.UpdateBalanceWithLock(tx, user.ID, deductAmount, true)
		})

		if err != nil {
			t.Fatalf("UpdateBalanceWithLock() unexpected error = %v", err)
		}

		// Verify actual balance after deduction
		updated, err := repo.FindByID(user.ID)
		if err != nil {
			t.Fatalf("FindByID() unexpected error = %v", err)
		}

		expectedBalance := initialBalance - deductAmount
		if updated.Balance != expectedBalance {
			t.Errorf("Balance after deduction = %v, want exactly %v (initial %v - deducted %v)",
				updated.Balance, expectedBalance, initialBalance, deductAmount)
		}

		// Verify version incremented
		if updated.Version != 2 {
			t.Errorf("Version = %v, want 2 (incremented)", updated.Version)
		}
	})

	t.Run("add balance with actual value verification", func(t *testing.T) {
		// Get current balance
		current, _ := repo.FindByID(user.ID)
		currentBalance := current.Balance
		addAmount := int64(500)

		err := db.Transaction(func(tx *gorm.DB) error {
			return repo.UpdateBalanceWithLock(tx, user.ID, addAmount, false)
		})

		if err != nil {
			t.Fatalf("UpdateBalanceWithLock() unexpected error = %v", err)
		}

		// Verify actual balance after addition
		updated, err := repo.FindByID(user.ID)
		if err != nil {
			t.Fatalf("FindByID() unexpected error = %v", err)
		}

		expectedBalance := currentBalance + addAmount
		if updated.Balance != expectedBalance {
			t.Errorf("Balance after addition = %v, want exactly %v (current %v + added %v)",
				updated.Balance, expectedBalance, currentBalance, addAmount)
		}
	})

	t.Run("deduct more than balance - should fail", func(t *testing.T) {
		current, _ := repo.FindByID(user.ID)
		excessiveAmount := current.Balance + 1000

		err := db.Transaction(func(tx *gorm.DB) error {
			return repo.UpdateBalanceWithLock(tx, user.ID, excessiveAmount, true)
		})

		if err == nil {
			t.Errorf("UpdateBalanceWithLock() expected error for insufficient balance, got nil")
		}

		// Verify balance unchanged
		unchanged, _ := repo.FindByID(user.ID)
		if unchanged.Balance != current.Balance {
			t.Errorf("Balance changed on error: got %v, want %v", unchanged.Balance, current.Balance)
		}
	})
}

func TestUserRepository_Update_OptimisticLock_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Test User",
		Balance:      1000,
		Role:         domain.RoleUser,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create() unexpected error = %v", err)
	}

	t.Run("successful update with correct version", func(t *testing.T) {
		// Get current state
		current, _ := repo.FindByID(user.ID)
		current.DisplayName = "Updated Name"
		current.Version++ // Increment version for update

		success, err := repo.Update(current)

		if err != nil {
			t.Fatalf("Update() unexpected error = %v", err)
		}
		if !success {
			t.Errorf("Update() success = false, want true")
		}

		// Verify actual updated values
		updated, _ := repo.FindByID(user.ID)
		if updated.DisplayName != "Updated Name" {
			t.Errorf("DisplayName = %v, want 'Updated Name'", updated.DisplayName)
		}
		if updated.Version != 2 {
			t.Errorf("Version = %v, want 2 (incremented)", updated.Version)
		}
	})

	t.Run("failed update with stale version", func(t *testing.T) {
		// Use stale version (version 1, but DB has version 2 now)
		staleUser := &domain.User{
			ID:          user.ID,
			DisplayName: "Another Update",
			Version:     2, // Stale: pretending current version is 1 (Version-1 in Update)
		}

		success, err := repo.Update(staleUser)

		if err != nil {
			t.Fatalf("Update() unexpected error = %v", err)
		}
		if success {
			t.Errorf("Update() success = true, want false (stale version)")
		}

		// Verify data unchanged
		unchanged, _ := repo.FindByID(user.ID)
		if unchanged.DisplayName != "Updated Name" {
			t.Errorf("DisplayName changed on stale update: got %v, want 'Updated Name'", unchanged.DisplayName)
		}
	})
}

func TestUserRepository_FindByUsername_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "findme",
		Email:        "findme@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Find Me",
		Balance:      2000,
		Role:         domain.RoleUser,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create() unexpected error = %v", err)
	}

	found, err := repo.FindByUsername("findme")

	if err != nil {
		t.Fatalf("FindByUsername() unexpected error = %v", err)
	}

	// Verify actual values
	if found.ID != user.ID {
		t.Errorf("ID = %v, want %v", found.ID, user.ID)
	}
	if found.Balance != 2000 {
		t.Errorf("Balance = %v, want 2000", found.Balance)
	}
	if found.Email != "findme@example.com" {
		t.Errorf("Email = %v, want 'findme@example.com'", found.Email)
	}
}

func TestUserRepository_List_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create multiple users
	for i := 0; i < 5; i++ {
		user := &domain.User{
			ID:           uuid.New(),
			Username:     "user" + string(rune('1'+i)),
			Email:        "user" + string(rune('1'+i)) + "@example.com",
			PasswordHash: "hashedpassword",
			DisplayName:  "User " + string(rune('1'+i)),
			Balance:      int64(1000 * (i + 1)),
			Role:         domain.RoleUser,
			Version:      1,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		repo.Create(user)
	}

	t.Run("list with pagination", func(t *testing.T) {
		users, err := repo.List(0, 3)

		if err != nil {
			t.Fatalf("List() unexpected error = %v", err)
		}

		if len(users) != 3 {
			t.Errorf("List count = %v, want 3", len(users))
		}

		// Verify each user has valid balance
		for _, u := range users {
			if u.Balance < 1000 || u.Balance > 5000 {
				t.Errorf("User %s has unexpected balance: %v", u.Username, u.Balance)
			}
		}
	})

	t.Run("count total users", func(t *testing.T) {
		count, err := repo.Count()

		if err != nil {
			t.Fatalf("Count() unexpected error = %v", err)
		}

		if count != 5 {
			t.Errorf("Count = %v, want 5", count)
		}
	})
}
