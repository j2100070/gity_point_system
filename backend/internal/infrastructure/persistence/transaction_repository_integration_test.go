//go:build integration
// +build integration

package persistence

import (
	"testing"
	"time"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestTransactionRepository_Create_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTransactionRepository(db)

	fromUserID := uuid.New()
	toUserID := uuid.New()
	amount := int64(1500)
	idempotencyKey := "test-key-123"

	transaction, err := domain.NewTransfer(fromUserID, toUserID, amount, idempotencyKey, "Test transfer")
	if err != nil {
		t.Fatalf("NewTransfer() unexpected error = %v", err)
	}

	err = repo.Create(nil, transaction)

	if err != nil {
		t.Fatalf("Create() unexpected error = %v", err)
	}

	// Verify actual values in database
	found, err := repo.FindByID(transaction.ID)
	if err != nil {
		t.Fatalf("FindByID() unexpected error = %v", err)
	}

	if found.Amount != amount {
		t.Errorf("Amount = %v, want exactly %v", found.Amount, amount)
	}
	if *found.FromUserID != fromUserID {
		t.Errorf("FromUserID = %v, want %v", *found.FromUserID, fromUserID)
	}
	if *found.ToUserID != toUserID {
		t.Errorf("ToUserID = %v, want %v", *found.ToUserID, toUserID)
	}
	if *found.IdempotencyKey != idempotencyKey {
		t.Errorf("IdempotencyKey = %v, want %v", *found.IdempotencyKey, idempotencyKey)
	}

}

func TestTransactionRepository_FindByIdempotencyKey_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTransactionRepository(db)

	fromUserID := uuid.New()
	toUserID := uuid.New()
	idempotencyKey := "unique-key-456"

	transaction, _ := domain.NewTransfer(fromUserID, toUserID, 2000, idempotencyKey, "Test")
	err := repo.Create(nil, transaction)
	if err != nil {
		t.Fatalf("Create() unexpected error = %v", err)
	}

	found, err := repo.FindByIdempotencyKey(idempotencyKey)

	if err != nil {
		t.Fatalf("FindByIdempotencyKey() unexpected error = %v", err)
	}

	if found.ID != transaction.ID {
		t.Errorf("ID = %v, want %v", found.ID, transaction.ID)
	}
	if found.Amount != 2000 {
		t.Errorf("Amount = %v, want exactly 2000", found.Amount)
	}
}

func TestTransactionRepository_ListByUserID_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTransactionRepository(db)
	userID := uuid.New()
	otherUserID := uuid.New()

	// Create transactions for the user
	amounts := []int64{100, 250, 500, 1000, 1500}
	for i, amount := range amounts {
		var tx *domain.Transaction
		var err error

		if i%2 == 0 {
			// User is sender
			tx, err = domain.NewTransfer(userID, otherUserID, amount, "key-"+string(rune('1'+i)), "Test")
		} else {
			// User is receiver
			tx, err = domain.NewTransfer(otherUserID, userID, amount, "key-"+string(rune('1'+i)), "Test")
		}

		if err != nil {
			t.Fatalf("NewTransfer() unexpected error = %v", err)
		}

		err = repo.Create(nil, tx)
		if err != nil {
			t.Fatalf("Create() unexpected error = %v", err)
		}
	}

	transactions, err := repo.ListByUserID(userID, 0, 10)

	if err != nil {
		t.Fatalf("ListByUserID() unexpected error = %v", err)
	}

	if len(transactions) != 5 {
		t.Errorf("Transaction count = %v, want 5", len(transactions))
	}

	// Verify actual amounts
	foundAmounts := make(map[int64]bool)
	for _, tx := range transactions {
		foundAmounts[tx.Amount] = true
	}

	for _, expectedAmount := range amounts {
		if !foundAmounts[expectedAmount] {
			t.Errorf("Expected transaction with amount %v not found", expectedAmount)
		}
	}
}

func TestTransactionRepository_CountByUserID_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTransactionRepository(db)
	userID := uuid.New()
	otherUserID := uuid.New()

	// Create 7 transactions
	for i := 0; i < 7; i++ {
		tx, _ := domain.NewTransfer(userID, otherUserID, int64(100*(i+1)), "key-"+string(rune('1'+i)), "Test")
		repo.Create(nil, tx)
	}

	count, err := repo.CountByUserID(userID)

	if err != nil {
		t.Fatalf("CountByUserID() unexpected error = %v", err)
	}

	if count != 7 {
		t.Errorf("Count = %v, want exactly 7", count)
	}
}

func TestTransactionRepository_Update_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTransactionRepository(db)

	transaction, _ := domain.NewTransfer(uuid.New(), uuid.New(), 3000, "key-789", "Test")
	err := repo.Create(nil, transaction)
	if err != nil {
		t.Fatalf("Create() unexpected error = %v", err)
	}

	// Complete the transaction
	err = transaction.Complete()
	if err != nil {
		t.Fatalf("Complete() unexpected error = %v", err)
	}

	err = repo.Update(nil, transaction)
	if err != nil {
		t.Fatalf("Update() unexpected error = %v", err)
	}

	// Verify actual status updated
	updated, err := repo.FindByID(transaction.ID)
	if err != nil {
		t.Fatalf("FindByID() unexpected error = %v", err)
	}

	if updated.Status != domain.TransactionStatusCompleted {
		t.Errorf("Status = %v, want %v", updated.Status, domain.TransactionStatusCompleted)
	}
	if updated.CompletedAt == nil {
		t.Errorf("CompletedAt is nil, want timestamp")
	}
}

func TestFullPointTransfer_Integration(t *testing.T) {
	// This test simulates a complete point transfer with actual balance changes
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := NewUserRepository(db)
	transactionRepo := NewTransactionRepository(db)
	idempotencyRepo := NewIdempotencyKeyRepository(db)

	// Create sender with initial balance
	sender := &domain.User{
		ID:           uuid.New(),
		Username:     "sender",
		Email:        "sender@example.com",
		PasswordHash: "hash",
		DisplayName:  "Sender",
		Balance:      10000, // Initial balance
		Role:         domain.RoleUser,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.Create(sender)

	// Create receiver with initial balance
	receiver := &domain.User{
		ID:           uuid.New(),
		Username:     "receiver",
		Email:        "receiver@example.com",
		PasswordHash: "hash",
		DisplayName:  "Receiver",
		Balance:      5000, // Initial balance
		Role:         domain.RoleUser,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.Create(receiver)

	// Transfer amount
	transferAmount := int64(3000)
	idempotencyKey := "transfer-key-unique"

	// Create idempotency key
	idemKey := domain.NewIdempotencyKey(idempotencyKey, sender.ID)
	err := idempotencyRepo.Create(idemKey)
	if err != nil {
		t.Fatalf("Create idempotency key failed: %v", err)
	}

	// Execute transfer in transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		// Deduct from sender
		if err := userRepo.UpdateBalanceWithLock(tx, sender.ID, transferAmount, true); err != nil {
			return err
		}

		// Add to receiver
		if err := userRepo.UpdateBalanceWithLock(tx, receiver.ID, transferAmount, false); err != nil {
			return err
		}

		// Create transaction record
		transaction, err := domain.NewTransfer(sender.ID, receiver.ID, transferAmount, idempotencyKey, "Integration test transfer")
		if err != nil {
			return err
		}

		if err := transactionRepo.Create(tx, transaction); err != nil {
			return err
		}

		// Complete transaction
		if err := transaction.Complete(); err != nil {
			return err
		}

		if err := transactionRepo.Update(tx, transaction); err != nil {
			return err
		}

		// Update idempotency key
		idemKey.Status = "completed"
		idemKey.TransactionID = &transaction.ID
		return idempotencyRepo.Update(idemKey)
	})

	if err != nil {
		t.Fatalf("Transfer transaction failed: %v", err)
	}

	// Verify actual balances after transfer
	senderAfter, err := userRepo.FindByID(sender.ID)
	if err != nil {
		t.Fatalf("FindByID sender failed: %v", err)
	}

	receiverAfter, err := userRepo.FindByID(receiver.ID)
	if err != nil {
		t.Fatalf("FindByID receiver failed: %v", err)
	}

	expectedSenderBalance := int64(10000 - 3000)  // 7000
	expectedReceiverBalance := int64(5000 + 3000) // 8000

	if senderAfter.Balance != expectedSenderBalance {
		t.Errorf("Sender balance after transfer = %v, want exactly %v (initial 10000 - transferred 3000)",
			senderAfter.Balance, expectedSenderBalance)
	}

	if receiverAfter.Balance != expectedReceiverBalance {
		t.Errorf("Receiver balance after transfer = %v, want exactly %v (initial 5000 + received 3000)",
			receiverAfter.Balance, expectedReceiverBalance)
	}

	// Verify transaction record
	transactions, err := transactionRepo.ListByUserID(sender.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("Transaction count = %v, want 1", len(transactions))
	}

	if transactions[0].Amount != transferAmount {
		t.Errorf("Transaction amount = %v, want exactly %v", transactions[0].Amount, transferAmount)
	}

	if transactions[0].Status != domain.TransactionStatusCompleted {
		t.Errorf("Transaction status = %v, want %v", transactions[0].Status, domain.TransactionStatusCompleted)
	}
}

func TestConcurrentTransfers_Integration(t *testing.T) {
	// Test pessimistic locking with concurrent transfers
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := NewUserRepository(db)

	// Create user with balance
	user := &domain.User{
		ID:           uuid.New(),
		Username:     "concurrent_test",
		Email:        "concurrent@example.com",
		PasswordHash: "hash",
		DisplayName:  "Concurrent Test",
		Balance:      10000,
		Role:         domain.RoleUser,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.Create(user)

	// Perform 5 sequential deductions
	deductAmount := int64(1000)
	for i := 0; i < 5; i++ {
		err := db.Transaction(func(tx *gorm.DB) error {
			return userRepo.UpdateBalanceWithLock(tx, user.ID, deductAmount, true)
		})

		if err != nil {
			t.Fatalf("Transfer %d failed: %v", i+1, err)
		}
	}

	// Verify final balance
	final, err := userRepo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	expectedBalance := int64(10000 - 5*1000) // 5000
	if final.Balance != expectedBalance {
		t.Errorf("Final balance = %v, want exactly %v (initial 10000 - 5x1000)", final.Balance, expectedBalance)
	}

	// Verify version incremented correctly
	if final.Version != 6 {
		t.Errorf("Version = %v, want 6 (initial 1 + 5 updates)", final.Version)
	}
}

func TestAdminGrantAndDeduct_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := NewUserRepository(db)
	transactionRepo := NewTransactionRepository(db)

	adminID := uuid.New()
	userID := uuid.New()

	// Create user
	user := &domain.User{
		ID:           userID,
		Username:     "testuser",
		Email:        "testuser@example.com",
		PasswordHash: "hash",
		DisplayName:  "Test User",
		Balance:      1000,
		Role:         domain.RoleUser,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.Create(user)

	t.Run("admin grant points", func(t *testing.T) {
		grantAmount := int64(5000)

		// Create admin grant transaction
		transaction, err := domain.NewAdminGrant(userID, grantAmount, "Admin grant", adminID)
		if err != nil {
			t.Fatalf("NewAdminGrant() failed: %v", err)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			// Add to user balance
			if err := userRepo.UpdateBalanceWithLock(tx, userID, grantAmount, false); err != nil {
				return err
			}

			// Create transaction record
			return transactionRepo.Create(tx, transaction)
		})

		if err != nil {
			t.Fatalf("Grant transaction failed: %v", err)
		}

		// Verify actual balance
		updated, err := userRepo.FindByID(userID)
		if err != nil {
			t.Fatalf("FindByID() failed: %v", err)
		}

		expectedBalance := int64(1000 + 5000) // 6000
		if updated.Balance != expectedBalance {
			t.Errorf("Balance after grant = %v, want exactly %v (initial 1000 + granted 5000)",
				updated.Balance, expectedBalance)
		}

		// Verify transaction metadata
		txRecord, _ := transactionRepo.FindByID(transaction.ID)
		if txRecord.Metadata["admin_id"] != adminID.String() {
			t.Errorf("Admin ID in metadata = %v, want %v", txRecord.Metadata["admin_id"], adminID.String())
		}
	})

	t.Run("admin deduct points", func(t *testing.T) {
		// Get current balance
		current, _ := userRepo.FindByID(userID)
		currentBalance := current.Balance
		deductAmount := int64(2000)

		// Create admin deduct transaction
		transaction, err := domain.NewAdminDeduct(userID, deductAmount, "Admin deduct", adminID)
		if err != nil {
			t.Fatalf("NewAdminDeduct() failed: %v", err)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			// Deduct from user balance
			if err := userRepo.UpdateBalanceWithLock(tx, userID, deductAmount, true); err != nil {
				return err
			}

			// Create transaction record
			return transactionRepo.Create(tx, transaction)
		})

		if err != nil {
			t.Fatalf("Deduct transaction failed: %v", err)
		}

		// Verify actual balance
		updated, err := userRepo.FindByID(userID)
		if err != nil {
			t.Fatalf("FindByID() failed: %v", err)
		}

		expectedBalance := currentBalance - deductAmount
		if updated.Balance != expectedBalance {
			t.Errorf("Balance after deduct = %v, want exactly %v (current %v - deducted %v)",
				updated.Balance, expectedBalance, currentBalance, deductAmount)
		}
	})
}
