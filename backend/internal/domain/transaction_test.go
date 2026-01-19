package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTransfer(t *testing.T) {
	fromUserID := uuid.New()
	toUserID := uuid.New()

	tests := []struct {
		name           string
		fromUserID     uuid.UUID
		toUserID       uuid.UUID
		amount         int64
		idempotencyKey string
		description    string
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "valid transfer",
			fromUserID:     fromUserID,
			toUserID:       toUserID,
			amount:         1000,
			idempotencyKey: "test-key-1",
			description:    "Test transfer",
			wantErr:        false,
		},
		{
			name:           "same user",
			fromUserID:     fromUserID,
			toUserID:       fromUserID,
			amount:         1000,
			idempotencyKey: "test-key-2",
			description:    "Test transfer",
			wantErr:        true,
			errMsg:         "cannot transfer to the same user",
		},
		{
			name:           "zero amount",
			fromUserID:     fromUserID,
			toUserID:       toUserID,
			amount:         0,
			idempotencyKey: "test-key-3",
			description:    "Test transfer",
			wantErr:        true,
			errMsg:         "amount must be positive",
		},
		{
			name:           "negative amount",
			fromUserID:     fromUserID,
			toUserID:       toUserID,
			amount:         -100,
			idempotencyKey: "test-key-4",
			description:    "Test transfer",
			wantErr:        true,
			errMsg:         "amount must be positive",
		},
		{
			name:           "empty idempotency key",
			fromUserID:     fromUserID,
			toUserID:       toUserID,
			amount:         1000,
			idempotencyKey: "",
			description:    "Test transfer",
			wantErr:        true,
			errMsg:         "idempotency key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := NewTransfer(tt.fromUserID, tt.toUserID, tt.amount, tt.idempotencyKey, tt.description)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTransfer() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewTransfer() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewTransfer() unexpected error = %v", err)
				return
			}

			// Verify actual values
			if *transaction.FromUserID != tt.fromUserID {
				t.Errorf("FromUserID = %v, want %v", *transaction.FromUserID, tt.fromUserID)
			}
			if *transaction.ToUserID != tt.toUserID {
				t.Errorf("ToUserID = %v, want %v", *transaction.ToUserID, tt.toUserID)
			}
			if transaction.Amount != tt.amount {
				t.Errorf("Amount = %v, want %v", transaction.Amount, tt.amount)
			}
			if *transaction.IdempotencyKey != tt.idempotencyKey {
				t.Errorf("IdempotencyKey = %v, want %v", *transaction.IdempotencyKey, tt.idempotencyKey)
			}
			if transaction.Description != tt.description {
				t.Errorf("Description = %v, want %v", transaction.Description, tt.description)
			}
			if transaction.TransactionType != TransactionTypeTransfer {
				t.Errorf("TransactionType = %v, want %v", transaction.TransactionType, TransactionTypeTransfer)
			}
			if transaction.Status != TransactionStatusPending {
				t.Errorf("Status = %v, want %v", transaction.Status, TransactionStatusPending)
			}
			if transaction.Metadata == nil {
				t.Errorf("Metadata is nil, want empty map")
			}
		})
	}
}

func TestNewAdminGrant(t *testing.T) {
	toUserID := uuid.New()
	adminID := uuid.New()

	tests := []struct {
		name        string
		toUserID    uuid.UUID
		amount      int64
		description string
		adminID     uuid.UUID
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid admin grant",
			toUserID:    toUserID,
			amount:      5000,
			description: "Admin grant test",
			adminID:     adminID,
			wantErr:     false,
		},
		{
			name:        "zero amount",
			toUserID:    toUserID,
			amount:      0,
			description: "Admin grant test",
			adminID:     adminID,
			wantErr:     true,
			errMsg:      "amount must be positive",
		},
		{
			name:        "negative amount",
			toUserID:    toUserID,
			amount:      -100,
			description: "Admin grant test",
			adminID:     adminID,
			wantErr:     true,
			errMsg:      "amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := NewAdminGrant(tt.toUserID, tt.amount, tt.description, tt.adminID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAdminGrant() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewAdminGrant() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewAdminGrant() unexpected error = %v", err)
				return
			}

			// Verify actual values
			if transaction.FromUserID != nil {
				t.Errorf("FromUserID = %v, want nil (system grant)", transaction.FromUserID)
			}
			if *transaction.ToUserID != tt.toUserID {
				t.Errorf("ToUserID = %v, want %v", *transaction.ToUserID, tt.toUserID)
			}
			if transaction.Amount != tt.amount {
				t.Errorf("Amount = %v, want %v", transaction.Amount, tt.amount)
			}
			if transaction.Description != tt.description {
				t.Errorf("Description = %v, want %v", transaction.Description, tt.description)
			}
			if transaction.TransactionType != TransactionTypeAdminGrant {
				t.Errorf("TransactionType = %v, want %v", transaction.TransactionType, TransactionTypeAdminGrant)
			}
			if transaction.Status != TransactionStatusCompleted {
				t.Errorf("Status = %v, want %v", transaction.Status, TransactionStatusCompleted)
			}
			if transaction.CompletedAt == nil {
				t.Errorf("CompletedAt is nil, want timestamp")
			}
			// Verify admin ID in metadata
			if transaction.Metadata["admin_id"] != tt.adminID.String() {
				t.Errorf("Metadata admin_id = %v, want %v", transaction.Metadata["admin_id"], tt.adminID.String())
			}
		})
	}
}

func TestNewAdminDeduct(t *testing.T) {
	fromUserID := uuid.New()
	adminID := uuid.New()

	tests := []struct {
		name        string
		fromUserID  uuid.UUID
		amount      int64
		description string
		adminID     uuid.UUID
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid admin deduct",
			fromUserID:  fromUserID,
			amount:      3000,
			description: "Admin deduct test",
			adminID:     adminID,
			wantErr:     false,
		},
		{
			name:        "zero amount",
			fromUserID:  fromUserID,
			amount:      0,
			description: "Admin deduct test",
			adminID:     adminID,
			wantErr:     true,
			errMsg:      "amount must be positive",
		},
		{
			name:        "negative amount",
			fromUserID:  fromUserID,
			amount:      -100,
			description: "Admin deduct test",
			adminID:     adminID,
			wantErr:     true,
			errMsg:      "amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := NewAdminDeduct(tt.fromUserID, tt.amount, tt.description, tt.adminID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAdminDeduct() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewAdminDeduct() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewAdminDeduct() unexpected error = %v", err)
				return
			}

			// Verify actual values
			if *transaction.FromUserID != tt.fromUserID {
				t.Errorf("FromUserID = %v, want %v", *transaction.FromUserID, tt.fromUserID)
			}
			if transaction.ToUserID != nil {
				t.Errorf("ToUserID = %v, want nil (system deduct)", *transaction.ToUserID)
			}
			if transaction.Amount != tt.amount {
				t.Errorf("Amount = %v, want %v", transaction.Amount, tt.amount)
			}
			if transaction.Description != tt.description {
				t.Errorf("Description = %v, want %v", transaction.Description, tt.description)
			}
			if transaction.TransactionType != TransactionTypeAdminDeduct {
				t.Errorf("TransactionType = %v, want %v", transaction.TransactionType, TransactionTypeAdminDeduct)
			}
			if transaction.Status != TransactionStatusCompleted {
				t.Errorf("Status = %v, want %v", transaction.Status, TransactionStatusCompleted)
			}
			if transaction.CompletedAt == nil {
				t.Errorf("CompletedAt is nil, want timestamp")
			}
			// Verify admin ID in metadata
			if transaction.Metadata["admin_id"] != tt.adminID.String() {
				t.Errorf("Metadata admin_id = %v, want %v", transaction.Metadata["admin_id"], tt.adminID.String())
			}
		})
	}
}

func TestTransaction_Complete(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus TransactionStatus
		wantErr       bool
		errMsg        string
		expectedStatus TransactionStatus
	}{
		{
			name:           "complete pending transaction",
			initialStatus:  TransactionStatusPending,
			wantErr:        false,
			expectedStatus: TransactionStatusCompleted,
		},
		{
			name:          "complete already completed transaction",
			initialStatus: TransactionStatusCompleted,
			wantErr:       true,
			errMsg:        "transaction is not in pending status",
		},
		{
			name:          "complete failed transaction",
			initialStatus: TransactionStatusFailed,
			wantErr:       true,
			errMsg:        "transaction is not in pending status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction := &Transaction{
				ID:     uuid.New(),
				Status: tt.initialStatus,
			}

			err := transaction.Complete()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Complete() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Complete() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Complete() unexpected error = %v", err)
				return
			}

			// Verify actual status changed
			if transaction.Status != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", transaction.Status, tt.expectedStatus)
			}

			// Verify CompletedAt is set
			if transaction.CompletedAt == nil {
				t.Errorf("CompletedAt is nil, want timestamp")
			}
		})
	}
}

func TestTransaction_Fail(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus TransactionStatus
		wantErr       bool
		errMsg        string
		expectedStatus TransactionStatus
	}{
		{
			name:           "fail pending transaction",
			initialStatus:  TransactionStatusPending,
			wantErr:        false,
			expectedStatus: TransactionStatusFailed,
		},
		{
			name:          "fail already completed transaction",
			initialStatus: TransactionStatusCompleted,
			wantErr:       true,
			errMsg:        "transaction is not in pending status",
		},
		{
			name:          "fail already failed transaction",
			initialStatus: TransactionStatusFailed,
			wantErr:       true,
			errMsg:        "transaction is not in pending status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction := &Transaction{
				ID:     uuid.New(),
				Status: tt.initialStatus,
			}

			err := transaction.Fail()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Fail() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Fail() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Fail() unexpected error = %v", err)
				return
			}

			// Verify actual status changed
			if transaction.Status != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", transaction.Status, tt.expectedStatus)
			}
		})
	}
}

func TestNewIdempotencyKey(t *testing.T) {
	userID := uuid.New()
	key := "test-idempotency-key"

	idempotencyKey := NewIdempotencyKey(key, userID)

	// Verify actual values
	if idempotencyKey.Key != key {
		t.Errorf("Key = %v, want %v", idempotencyKey.Key, key)
	}
	if idempotencyKey.UserID != userID {
		t.Errorf("UserID = %v, want %v", idempotencyKey.UserID, userID)
	}
	if idempotencyKey.Status != "processing" {
		t.Errorf("Status = %v, want processing", idempotencyKey.Status)
	}
	if idempotencyKey.TransactionID != nil {
		t.Errorf("TransactionID = %v, want nil", idempotencyKey.TransactionID)
	}

	// Verify expiration is approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	timeDiff := idempotencyKey.ExpiresAt.Sub(expectedExpiry)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("ExpiresAt = %v, want approximately %v", idempotencyKey.ExpiresAt, expectedExpiry)
	}
}

func TestTransactionAmounts(t *testing.T) {
	t.Run("verify exact transfer amount", func(t *testing.T) {
		fromUserID := uuid.New()
		toUserID := uuid.New()
		expectedAmount := int64(12345)

		transaction, err := NewTransfer(fromUserID, toUserID, expectedAmount, "key", "test")

		if err != nil {
			t.Fatalf("NewTransfer() unexpected error = %v", err)
		}

		if transaction.Amount != expectedAmount {
			t.Errorf("Amount = %v, want exactly %v", transaction.Amount, expectedAmount)
		}
	})

	t.Run("verify exact admin grant amount", func(t *testing.T) {
		toUserID := uuid.New()
		adminID := uuid.New()
		expectedAmount := int64(999999)

		transaction, err := NewAdminGrant(toUserID, expectedAmount, "grant", adminID)

		if err != nil {
			t.Fatalf("NewAdminGrant() unexpected error = %v", err)
		}

		if transaction.Amount != expectedAmount {
			t.Errorf("Amount = %v, want exactly %v", transaction.Amount, expectedAmount)
		}
	})

	t.Run("verify exact admin deduct amount", func(t *testing.T) {
		fromUserID := uuid.New()
		adminID := uuid.New()
		expectedAmount := int64(54321)

		transaction, err := NewAdminDeduct(fromUserID, expectedAmount, "deduct", adminID)

		if err != nil {
			t.Fatalf("NewAdminDeduct() unexpected error = %v", err)
		}

		if transaction.Amount != expectedAmount {
			t.Errorf("Amount = %v, want exactly %v", transaction.Amount, expectedAmount)
		}
	})
}
