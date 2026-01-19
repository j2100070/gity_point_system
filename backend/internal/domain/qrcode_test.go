package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewReceiveQRCode(t *testing.T) {
	userID := uuid.New()
	amount := int64(1000)

	tests := []struct {
		name    string
		userID  uuid.UUID
		amount  *int64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid receive QR with fixed amount",
			userID:  userID,
			amount:  &amount,
			wantErr: false,
		},
		{
			name:    "valid receive QR with variable amount",
			userID:  userID,
			amount:  nil,
			wantErr: false,
		},
		{
			name:    "zero amount",
			userID:  userID,
			amount:  ptrInt64(0),
			wantErr: true,
			errMsg:  "amount must be positive",
		},
		{
			name:    "negative amount",
			userID:  userID,
			amount:  ptrInt64(-100),
			wantErr: true,
			errMsg:  "amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qrCode, err := NewReceiveQRCode(tt.userID, tt.amount)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewReceiveQRCode() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewReceiveQRCode() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewReceiveQRCode() unexpected error = %v", err)
				return
			}

			// Verify actual values
			if qrCode.UserID != tt.userID {
				t.Errorf("UserID = %v, want %v", qrCode.UserID, tt.userID)
			}
			if tt.amount != nil && (qrCode.Amount == nil || *qrCode.Amount != *tt.amount) {
				t.Errorf("Amount = %v, want %v", qrCode.Amount, tt.amount)
			}
			if tt.amount == nil && qrCode.Amount != nil {
				t.Errorf("Amount = %v, want nil", qrCode.Amount)
			}
			if qrCode.QRType != QRCodeTypeReceive {
				t.Errorf("QRType = %v, want %v", qrCode.QRType, QRCodeTypeReceive)
			}
			if qrCode.Code == "" {
				t.Errorf("Code is empty, want generated code")
			}
			if qrCode.UsedAt != nil {
				t.Errorf("UsedAt = %v, want nil", qrCode.UsedAt)
			}

			// Verify expiration is approximately 5 minutes from now
			expectedExpiry := time.Now().Add(5 * time.Minute)
			timeDiff := qrCode.ExpiresAt.Sub(expectedExpiry)
			if timeDiff > time.Second || timeDiff < -time.Second {
				t.Errorf("ExpiresAt = %v, want approximately %v", qrCode.ExpiresAt, expectedExpiry)
			}
		})
	}
}

func TestNewSendQRCode(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		userID  uuid.UUID
		amount  int64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid send QR",
			userID:  userID,
			amount:  1000,
			wantErr: false,
		},
		{
			name:    "zero amount",
			userID:  userID,
			amount:  0,
			wantErr: true,
			errMsg:  "amount must be positive",
		},
		{
			name:    "negative amount",
			userID:  userID,
			amount:  -100,
			wantErr: true,
			errMsg:  "amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qrCode, err := NewSendQRCode(tt.userID, tt.amount)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewSendQRCode() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewSendQRCode() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewSendQRCode() unexpected error = %v", err)
				return
			}

			// Verify actual values
			if qrCode.UserID != tt.userID {
				t.Errorf("UserID = %v, want %v", qrCode.UserID, tt.userID)
			}
			if qrCode.Amount == nil || *qrCode.Amount != tt.amount {
				t.Errorf("Amount = %v, want %v", qrCode.Amount, tt.amount)
			}
			if qrCode.QRType != QRCodeTypeSend {
				t.Errorf("QRType = %v, want %v", qrCode.QRType, QRCodeTypeSend)
			}
			if qrCode.Code == "" {
				t.Errorf("Code is empty, want generated code")
			}

			// Verify expiration is approximately 5 minutes from now
			expectedExpiry := time.Now().Add(5 * time.Minute)
			timeDiff := qrCode.ExpiresAt.Sub(expectedExpiry)
			if timeDiff > time.Second || timeDiff < -time.Second {
				t.Errorf("ExpiresAt = %v, want approximately %v", qrCode.ExpiresAt, expectedExpiry)
			}
		})
	}
}

func TestQRCode_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired",
			expiresAt: time.Now().Add(1 * time.Minute),
			want:      false,
		},
		{
			name:      "expired",
			expiresAt: time.Now().Add(-1 * time.Minute),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qrCode := &QRCode{
				ExpiresAt: tt.expiresAt,
			}

			if got := qrCode.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQRCode_IsUsed(t *testing.T) {
	tests := []struct {
		name   string
		usedAt *time.Time
		want   bool
	}{
		{
			name:   "not used",
			usedAt: nil,
			want:   false,
		},
		{
			name:   "used",
			usedAt: ptrTime(time.Now()),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qrCode := &QRCode{
				UsedAt: tt.usedAt,
			}

			if got := qrCode.IsUsed(); got != tt.want {
				t.Errorf("IsUsed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQRCode_MarkAsUsed(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		usedAt      *time.Time
		expiresAt   time.Time
		markByUser  uuid.UUID
		wantErr     bool
		errMsg      string
	}{
		{
			name:       "mark fresh QR as used",
			usedAt:     nil,
			expiresAt:  time.Now().Add(5 * time.Minute),
			markByUser: userID,
			wantErr:    false,
		},
		{
			name:       "mark already used QR",
			usedAt:     ptrTime(time.Now().Add(-1 * time.Minute)),
			expiresAt:  time.Now().Add(5 * time.Minute),
			markByUser: userID,
			wantErr:    true,
			errMsg:     "qr code already used",
		},
		{
			name:       "mark expired QR",
			usedAt:     nil,
			expiresAt:  time.Now().Add(-1 * time.Minute),
			markByUser: userID,
			wantErr:    true,
			errMsg:     "qr code expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qrCode := &QRCode{
				ID:        uuid.New(),
				UsedAt:    tt.usedAt,
				ExpiresAt: tt.expiresAt,
			}

			err := qrCode.MarkAsUsed(tt.markByUser)

			if tt.wantErr {
				if err == nil {
					t.Errorf("MarkAsUsed() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("MarkAsUsed() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("MarkAsUsed() unexpected error = %v", err)
				return
			}

			// Verify actual values after marking as used
			if qrCode.UsedAt == nil {
				t.Errorf("UsedAt is nil, want timestamp")
			}
			if qrCode.UsedByUserID == nil || *qrCode.UsedByUserID != tt.markByUser {
				t.Errorf("UsedByUserID = %v, want %v", qrCode.UsedByUserID, tt.markByUser)
			}
		})
	}
}

func TestQRCode_CanBeUsedBy(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name      string
		ownerID   uuid.UUID
		usedAt    *time.Time
		expiresAt time.Time
		useByID   uuid.UUID
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "can be used by other user",
			ownerID:   ownerID,
			usedAt:    nil,
			expiresAt: time.Now().Add(5 * time.Minute),
			useByID:   otherUserID,
			wantErr:   false,
		},
		{
			name:      "cannot use own QR code",
			ownerID:   ownerID,
			usedAt:    nil,
			expiresAt: time.Now().Add(5 * time.Minute),
			useByID:   ownerID,
			wantErr:   true,
			errMsg:    "cannot use your own qr code",
		},
		{
			name:      "cannot use expired QR",
			ownerID:   ownerID,
			usedAt:    nil,
			expiresAt: time.Now().Add(-1 * time.Minute),
			useByID:   otherUserID,
			wantErr:   true,
			errMsg:    "qr code expired",
		},
		{
			name:      "cannot use already used QR",
			ownerID:   ownerID,
			usedAt:    ptrTime(time.Now().Add(-1 * time.Minute)),
			expiresAt: time.Now().Add(5 * time.Minute),
			useByID:   otherUserID,
			wantErr:   true,
			errMsg:    "qr code already used",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qrCode := &QRCode{
				ID:        uuid.New(),
				UserID:    tt.ownerID,
				UsedAt:    tt.usedAt,
				ExpiresAt: tt.expiresAt,
			}

			err := qrCode.CanBeUsedBy(tt.useByID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CanBeUsedBy() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("CanBeUsedBy() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("CanBeUsedBy() unexpected error = %v", err)
			}
		})
	}
}

func TestQRCodeAmounts(t *testing.T) {
	t.Run("verify exact fixed amount in receive QR", func(t *testing.T) {
		userID := uuid.New()
		expectedAmount := int64(77777)

		qrCode, err := NewReceiveQRCode(userID, &expectedAmount)

		if err != nil {
			t.Fatalf("NewReceiveQRCode() unexpected error = %v", err)
		}

		if qrCode.Amount == nil {
			t.Fatalf("Amount is nil, want %v", expectedAmount)
		}

		if *qrCode.Amount != expectedAmount {
			t.Errorf("Amount = %v, want exactly %v", *qrCode.Amount, expectedAmount)
		}
	})

	t.Run("verify exact amount in send QR", func(t *testing.T) {
		userID := uuid.New()
		expectedAmount := int64(88888)

		qrCode, err := NewSendQRCode(userID, expectedAmount)

		if err != nil {
			t.Fatalf("NewSendQRCode() unexpected error = %v", err)
		}

		if qrCode.Amount == nil {
			t.Fatalf("Amount is nil, want %v", expectedAmount)
		}

		if *qrCode.Amount != expectedAmount {
			t.Errorf("Amount = %v, want exactly %v", *qrCode.Amount, expectedAmount)
		}
	})

	t.Run("verify variable amount (nil) in receive QR", func(t *testing.T) {
		userID := uuid.New()

		qrCode, err := NewReceiveQRCode(userID, nil)

		if err != nil {
			t.Fatalf("NewReceiveQRCode() unexpected error = %v", err)
		}

		if qrCode.Amount != nil {
			t.Errorf("Amount = %v, want nil for variable amount", *qrCode.Amount)
		}
	})
}

func TestQRCodeGeneration(t *testing.T) {
	t.Run("QR codes have unique codes", func(t *testing.T) {
		userID := uuid.New()
		codes := make(map[string]bool)

		// Generate 100 QR codes and verify uniqueness
		for i := 0; i < 100; i++ {
			qrCode, err := NewReceiveQRCode(userID, nil)
			if err != nil {
				t.Fatalf("NewReceiveQRCode() unexpected error = %v", err)
			}

			if codes[qrCode.Code] {
				t.Errorf("Duplicate QR code generated: %v", qrCode.Code)
			}
			codes[qrCode.Code] = true
		}
	})
}

// Helper function
func ptrInt64(i int64) *int64 {
	return &i
}
