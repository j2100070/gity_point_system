package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewFriendship(t *testing.T) {
	requesterID := uuid.New()
	addresseeID := uuid.New()

	tests := []struct {
		name        string
		requesterID uuid.UUID
		addresseeID uuid.UUID
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid friendship",
			requesterID: requesterID,
			addresseeID: addresseeID,
			wantErr:     false,
		},
		{
			name:        "same user",
			requesterID: requesterID,
			addresseeID: requesterID,
			wantErr:     true,
			errMsg:      "cannot send friend request to yourself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			friendship, err := NewFriendship(tt.requesterID, tt.addresseeID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewFriendship() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewFriendship() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewFriendship() unexpected error = %v", err)
				return
			}

			// Verify actual values
			if friendship.RequesterID != tt.requesterID {
				t.Errorf("RequesterID = %v, want %v", friendship.RequesterID, tt.requesterID)
			}
			if friendship.AddresseeID != tt.addresseeID {
				t.Errorf("AddresseeID = %v, want %v", friendship.AddresseeID, tt.addresseeID)
			}
			if friendship.Status != FriendshipStatusPending {
				t.Errorf("Status = %v, want %v", friendship.Status, FriendshipStatusPending)
			}
		})
	}
}

func TestFriendship_Accept(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  FriendshipStatus
		wantErr        bool
		errMsg         string
		expectedStatus FriendshipStatus
	}{
		{
			name:           "accept pending friendship",
			initialStatus:  FriendshipStatusPending,
			wantErr:        false,
			expectedStatus: FriendshipStatusAccepted,
		},
		{
			name:          "accept already accepted friendship",
			initialStatus: FriendshipStatusAccepted,
			wantErr:       true,
			errMsg:        "friendship is not in pending status",
		},
		{
			name:          "accept rejected friendship",
			initialStatus: FriendshipStatusRejected,
			wantErr:       true,
			errMsg:        "friendship is not in pending status",
		},
		{
			name:          "accept blocked friendship",
			initialStatus: FriendshipStatusBlocked,
			wantErr:       true,
			errMsg:        "friendship is not in pending status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			friendship := &Friendship{
				ID:          uuid.New(),
				RequesterID: uuid.New(),
				AddresseeID: uuid.New(),
				Status:      tt.initialStatus,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			}
			beforeUpdate := friendship.UpdatedAt

			time.Sleep(10 * time.Millisecond)
			err := friendship.Accept()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Accept() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Accept() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Accept() unexpected error = %v", err)
				return
			}

			// Verify actual status changed
			if friendship.Status != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", friendship.Status, tt.expectedStatus)
			}

			// Verify UpdatedAt changed
			if !friendship.UpdatedAt.After(beforeUpdate) {
				t.Errorf("UpdatedAt not updated: before=%v, after=%v", beforeUpdate, friendship.UpdatedAt)
			}
		})
	}
}

func TestFriendship_Reject(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  FriendshipStatus
		wantErr        bool
		errMsg         string
		expectedStatus FriendshipStatus
	}{
		{
			name:           "reject pending friendship",
			initialStatus:  FriendshipStatusPending,
			wantErr:        false,
			expectedStatus: FriendshipStatusRejected,
		},
		{
			name:          "reject already accepted friendship",
			initialStatus: FriendshipStatusAccepted,
			wantErr:       true,
			errMsg:        "friendship is not in pending status",
		},
		{
			name:          "reject already rejected friendship",
			initialStatus: FriendshipStatusRejected,
			wantErr:       true,
			errMsg:        "friendship is not in pending status",
		},
		{
			name:          "reject blocked friendship",
			initialStatus: FriendshipStatusBlocked,
			wantErr:       true,
			errMsg:        "friendship is not in pending status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			friendship := &Friendship{
				ID:          uuid.New(),
				RequesterID: uuid.New(),
				AddresseeID: uuid.New(),
				Status:      tt.initialStatus,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			}
			beforeUpdate := friendship.UpdatedAt

			time.Sleep(10 * time.Millisecond)
			err := friendship.Reject()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Reject() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Reject() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Reject() unexpected error = %v", err)
				return
			}

			// Verify actual status changed
			if friendship.Status != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", friendship.Status, tt.expectedStatus)
			}

			// Verify UpdatedAt changed
			if !friendship.UpdatedAt.After(beforeUpdate) {
				t.Errorf("UpdatedAt not updated: before=%v, after=%v", beforeUpdate, friendship.UpdatedAt)
			}
		})
	}
}

func TestFriendship_Block(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  FriendshipStatus
		expectedStatus FriendshipStatus
	}{
		{
			name:           "block pending friendship",
			initialStatus:  FriendshipStatusPending,
			expectedStatus: FriendshipStatusBlocked,
		},
		{
			name:           "block accepted friendship",
			initialStatus:  FriendshipStatusAccepted,
			expectedStatus: FriendshipStatusBlocked,
		},
		{
			name:           "block rejected friendship",
			initialStatus:  FriendshipStatusRejected,
			expectedStatus: FriendshipStatusBlocked,
		},
		{
			name:           "block already blocked friendship",
			initialStatus:  FriendshipStatusBlocked,
			expectedStatus: FriendshipStatusBlocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			friendship := &Friendship{
				ID:          uuid.New(),
				RequesterID: uuid.New(),
				AddresseeID: uuid.New(),
				Status:      tt.initialStatus,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			}
			beforeUpdate := friendship.UpdatedAt

			time.Sleep(10 * time.Millisecond)
			friendship.Block()

			// Verify actual status changed
			if friendship.Status != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", friendship.Status, tt.expectedStatus)
			}

			// Verify UpdatedAt changed
			if !friendship.UpdatedAt.After(beforeUpdate) {
				t.Errorf("UpdatedAt not updated: before=%v, after=%v", beforeUpdate, friendship.UpdatedAt)
			}
		})
	}
}

func TestFriendship_IsAccepted(t *testing.T) {
	tests := []struct {
		name   string
		status FriendshipStatus
		want   bool
	}{
		{
			name:   "accepted friendship",
			status: FriendshipStatusAccepted,
			want:   true,
		},
		{
			name:   "pending friendship",
			status: FriendshipStatusPending,
			want:   false,
		},
		{
			name:   "rejected friendship",
			status: FriendshipStatusRejected,
			want:   false,
		},
		{
			name:   "blocked friendship",
			status: FriendshipStatusBlocked,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			friendship := &Friendship{
				Status: tt.status,
			}

			if got := friendship.IsAccepted(); got != tt.want {
				t.Errorf("IsAccepted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFriendshipStatusTransitions(t *testing.T) {
	t.Run("full lifecycle: pending -> accepted", func(t *testing.T) {
		requesterID := uuid.New()
		addresseeID := uuid.New()

		// Create new friendship
		friendship, err := NewFriendship(requesterID, addresseeID)
		if err != nil {
			t.Fatalf("NewFriendship() unexpected error = %v", err)
		}

		// Verify initial state
		if friendship.Status != FriendshipStatusPending {
			t.Errorf("Initial status = %v, want %v", friendship.Status, FriendshipStatusPending)
		}
		if friendship.IsAccepted() {
			t.Errorf("IsAccepted() = true, want false for pending")
		}

		// Accept friendship
		err = friendship.Accept()
		if err != nil {
			t.Fatalf("Accept() unexpected error = %v", err)
		}

		// Verify accepted state
		if friendship.Status != FriendshipStatusAccepted {
			t.Errorf("Status after Accept = %v, want %v", friendship.Status, FriendshipStatusAccepted)
		}
		if !friendship.IsAccepted() {
			t.Errorf("IsAccepted() = false, want true for accepted")
		}
	})

	t.Run("full lifecycle: pending -> rejected", func(t *testing.T) {
		requesterID := uuid.New()
		addresseeID := uuid.New()

		// Create new friendship
		friendship, err := NewFriendship(requesterID, addresseeID)
		if err != nil {
			t.Fatalf("NewFriendship() unexpected error = %v", err)
		}

		// Reject friendship
		err = friendship.Reject()
		if err != nil {
			t.Fatalf("Reject() unexpected error = %v", err)
		}

		// Verify rejected state
		if friendship.Status != FriendshipStatusRejected {
			t.Errorf("Status after Reject = %v, want %v", friendship.Status, FriendshipStatusRejected)
		}
		if friendship.IsAccepted() {
			t.Errorf("IsAccepted() = true, want false for rejected")
		}
	})

	t.Run("block at any status", func(t *testing.T) {
		requesterID := uuid.New()
		addresseeID := uuid.New()

		// Create new friendship
		friendship, err := NewFriendship(requesterID, addresseeID)
		if err != nil {
			t.Fatalf("NewFriendship() unexpected error = %v", err)
		}

		// Accept first
		err = friendship.Accept()
		if err != nil {
			t.Fatalf("Accept() unexpected error = %v", err)
		}

		// Then block
		friendship.Block()

		// Verify blocked state
		if friendship.Status != FriendshipStatusBlocked {
			t.Errorf("Status after Block = %v, want %v", friendship.Status, FriendshipStatusBlocked)
		}
		if friendship.IsAccepted() {
			t.Errorf("IsAccepted() = true, want false for blocked")
		}
	})
}
