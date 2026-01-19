package usecase

import (
	"errors"
	"testing"

	"github.com/gity/point-system/internal/domain"
	"github.com/gity/point-system/internal/domain/mock"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestPointTransferUseCase_Transfer_Validation(t *testing.T) {
	// NOTE: Full transfer testing with database transactions requires integration tests
	// These tests focus on validation and idempotency logic that don't require DB transactions
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockTransactionRepo := mock.NewMockTransactionRepository(ctrl)
	mockIdempotencyRepo := mock.NewMockIdempotencyKeyRepository(ctrl)
	mockFriendshipRepo := mock.NewMockFriendshipRepository(ctrl)

	useCase := NewPointTransferUseCase(
		nil, // DB is nil for validation tests
		mockUserRepo,
		mockTransactionRepo,
		mockIdempotencyRepo,
		mockFriendshipRepo,
	)

	t.Run("idempotency: return existing completed transaction", func(t *testing.T) {
		fromUserID := uuid.New()
		toUserID := uuid.New()
		transferAmount := int64(1000)
		existingTransactionID := uuid.New()

		existingIdempotencyKey := &domain.IdempotencyKey{
			Key:           "existing-key",
			UserID:        fromUserID,
			TransactionID: &existingTransactionID,
			Status:        "completed",
		}

		existingTransaction := &domain.Transaction{
			ID:         existingTransactionID,
			FromUserID: &fromUserID,
			ToUserID:   &toUserID,
			Amount:     transferAmount,
			Status:     domain.TransactionStatusCompleted,
		}

		fromUser := &domain.User{
			ID:       fromUserID,
			Balance:  4000, // Balance after previous transfer
			IsActive: true,
		}

		toUser := &domain.User{
			ID:       toUserID,
			Balance:  4000, // Balance after previous transfer
			IsActive: true,
		}

		req := &TransferRequest{
			FromUserID:     fromUserID,
			ToUserID:       toUserID,
			Amount:         transferAmount,
			IdempotencyKey: "existing-key",
			Description:    "Duplicate transfer attempt",
		}

		mockIdempotencyRepo.EXPECT().
			FindByKey(req.IdempotencyKey).
			Return(existingIdempotencyKey, nil)

		mockTransactionRepo.EXPECT().
			FindByID(existingTransactionID).
			Return(existingTransaction, nil)

		mockUserRepo.EXPECT().
			FindByID(fromUserID).
			Return(fromUser, nil)

		mockUserRepo.EXPECT().
			FindByID(toUserID).
			Return(toUser, nil)

		resp, err := useCase.Transfer(req)

		if err != nil {
			t.Fatalf("Transfer() unexpected error = %v", err)
		}

		// Verify actual values: should return existing transaction
		if resp.Transaction.ID != existingTransactionID {
			t.Errorf("Transaction ID = %v, want %v (existing)", resp.Transaction.ID, existingTransactionID)
		}
		if resp.Transaction.Amount != transferAmount {
			t.Errorf("Transaction amount = %v, want %v", resp.Transaction.Amount, transferAmount)
		}
		// Verify balances reflect the already completed transaction
		if resp.FromUser.Balance != 4000 {
			t.Errorf("FromUser balance = %v, want 4000 (after previous transfer)", resp.FromUser.Balance)
		}
		if resp.ToUser.Balance != 4000 {
			t.Errorf("ToUser balance = %v, want 4000 (after previous transfer)", resp.ToUser.Balance)
		}
	})

	t.Run("idempotency: reject duplicate processing", func(t *testing.T) {
		fromUserID := uuid.New()

		existingIdempotencyKey := &domain.IdempotencyKey{
			Key:    "processing-key",
			UserID: fromUserID,
			Status: "processing",
		}

		req := &TransferRequest{
			FromUserID:     fromUserID,
			ToUserID:       uuid.New(),
			Amount:         1000,
			IdempotencyKey: "processing-key",
			Description:    "Duplicate request while processing",
		}

		mockIdempotencyRepo.EXPECT().
			FindByKey(req.IdempotencyKey).
			Return(existingIdempotencyKey, nil)

		_, err := useCase.Transfer(req)

		if err == nil {
			t.Errorf("Transfer() expected error for duplicate processing, got nil")
		}
		if err.Error() != "transfer is already in progress" {
			t.Errorf("Transfer() error = %v, want 'transfer is already in progress'", err.Error())
		}
	})

	t.Run("validation: same user transfer", func(t *testing.T) {
		userID := uuid.New()

		req := &TransferRequest{
			FromUserID:     userID,
			ToUserID:       userID, // Same user
			Amount:         1000,
			IdempotencyKey: "test-key",
			Description:    "Self transfer",
		}

		_, err := useCase.Transfer(req)

		if err == nil {
			t.Errorf("Transfer() expected error for same user transfer, got nil")
		}
		if err.Error() != "cannot transfer to the same user" {
			t.Errorf("Transfer() error = %v, want 'cannot transfer to the same user'", err.Error())
		}
	})

	t.Run("validation: zero amount", func(t *testing.T) {
		req := &TransferRequest{
			FromUserID:     uuid.New(),
			ToUserID:       uuid.New(),
			Amount:         0, // Zero amount
			IdempotencyKey: "test-key",
			Description:    "Zero transfer",
		}

		_, err := useCase.Transfer(req)

		if err == nil {
			t.Errorf("Transfer() expected error for zero amount, got nil")
		}
		if err.Error() != "amount must be positive" {
			t.Errorf("Transfer() error = %v, want 'amount must be positive'", err.Error())
		}
	})

	t.Run("validation: negative amount", func(t *testing.T) {
		req := &TransferRequest{
			FromUserID:     uuid.New(),
			ToUserID:       uuid.New(),
			Amount:         -1000, // Negative amount
			IdempotencyKey: "test-key",
			Description:    "Negative transfer",
		}

		_, err := useCase.Transfer(req)

		if err == nil {
			t.Errorf("Transfer() expected error for negative amount, got nil")
		}
		if err.Error() != "amount must be positive" {
			t.Errorf("Transfer() error = %v, want 'amount must be positive'", err.Error())
		}
	})

	t.Run("validation: empty idempotency key", func(t *testing.T) {
		req := &TransferRequest{
			FromUserID:     uuid.New(),
			ToUserID:       uuid.New(),
			Amount:         1000,
			IdempotencyKey: "", // Empty
			Description:    "Transfer without key",
		}

		_, err := useCase.Transfer(req)

		if err == nil {
			t.Errorf("Transfer() expected error for empty idempotency key, got nil")
		}
		if err.Error() != "idempotency key is required" {
			t.Errorf("Transfer() error = %v, want 'idempotency key is required'", err.Error())
		}
	})

	// NOTE: Tests for actual transfers with database transactions moved to integration tests
}

func TestPointTransferUseCase_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := &gorm.DB{}
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockTransactionRepo := mock.NewMockTransactionRepository(ctrl)
	mockIdempotencyRepo := mock.NewMockIdempotencyKeyRepository(ctrl)
	mockFriendshipRepo := mock.NewMockFriendshipRepository(ctrl)

	useCase := NewPointTransferUseCase(
		mockDB,
		mockUserRepo,
		mockTransactionRepo,
		mockIdempotencyRepo,
		mockFriendshipRepo,
	)

	t.Run("get balance with actual value verification", func(t *testing.T) {
		userID := uuid.New()
		expectedBalance := int64(123456)

		user := &domain.User{
			ID:       userID,
			Username: "testuser",
			Balance:  expectedBalance,
			IsActive: true,
		}

		req := &GetBalanceRequest{
			UserID: userID,
		}

		mockUserRepo.EXPECT().
			FindByID(userID).
			Return(user, nil)

		resp, err := useCase.GetBalance(req)

		if err != nil {
			t.Fatalf("GetBalance() unexpected error = %v", err)
		}

		// Verify actual balance value
		if resp.Balance != expectedBalance {
			t.Errorf("Balance = %v, want exactly %v", resp.Balance, expectedBalance)
		}
		if resp.User.Balance != expectedBalance {
			t.Errorf("User.Balance = %v, want exactly %v", resp.User.Balance, expectedBalance)
		}
	})

	t.Run("get balance for non-existent user", func(t *testing.T) {
		userID := uuid.New()

		req := &GetBalanceRequest{
			UserID: userID,
		}

		mockUserRepo.EXPECT().
			FindByID(userID).
			Return(nil, errors.New("user not found"))

		_, err := useCase.GetBalance(req)

		if err == nil {
			t.Errorf("GetBalance() expected error for non-existent user, got nil")
		}
	})
}

func TestPointTransferUseCase_GetTransactionHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := &gorm.DB{}
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockTransactionRepo := mock.NewMockTransactionRepository(ctrl)
	mockIdempotencyRepo := mock.NewMockIdempotencyKeyRepository(ctrl)
	mockFriendshipRepo := mock.NewMockFriendshipRepository(ctrl)

	useCase := NewPointTransferUseCase(
		mockDB,
		mockUserRepo,
		mockTransactionRepo,
		mockIdempotencyRepo,
		mockFriendshipRepo,
	)

	t.Run("get transaction history with actual amounts", func(t *testing.T) {
		userID := uuid.New()

		transactions := []*domain.Transaction{
			{
				ID:     uuid.New(),
				Amount: 1000,
				Status: domain.TransactionStatusCompleted,
			},
			{
				ID:     uuid.New(),
				Amount: 2500,
				Status: domain.TransactionStatusCompleted,
			},
			{
				ID:     uuid.New(),
				Amount: 500,
				Status: domain.TransactionStatusCompleted,
			},
		}

		req := &GetTransactionHistoryRequest{
			UserID: userID,
			Offset: 0,
			Limit:  10,
		}

		mockTransactionRepo.EXPECT().
			ListByUserID(userID, 0, 10).
			Return(transactions, nil)

		mockTransactionRepo.EXPECT().
			CountByUserID(userID).
			Return(int64(3), nil)

		resp, err := useCase.GetTransactionHistory(req)

		if err != nil {
			t.Fatalf("GetTransactionHistory() unexpected error = %v", err)
		}

		// Verify actual values
		if len(resp.Transactions) != 3 {
			t.Errorf("Transaction count = %v, want 3", len(resp.Transactions))
		}
		if resp.Total != 3 {
			t.Errorf("Total = %v, want 3", resp.Total)
		}

		// Verify each transaction amount
		expectedAmounts := []int64{1000, 2500, 500}
		for i, tx := range resp.Transactions {
			if tx.Amount != expectedAmounts[i] {
				t.Errorf("Transaction[%d].Amount = %v, want %v", i, tx.Amount, expectedAmounts[i])
			}
		}
	})

	t.Run("get transaction history with pagination", func(t *testing.T) {
		userID := uuid.New()

		req := &GetTransactionHistoryRequest{
			UserID: userID,
			Offset: 10,
			Limit:  5,
		}

		mockTransactionRepo.EXPECT().
			ListByUserID(userID, 10, 5).
			Return([]*domain.Transaction{}, nil)

		mockTransactionRepo.EXPECT().
			CountByUserID(userID).
			Return(int64(15), nil)

		resp, err := useCase.GetTransactionHistory(req)

		if err != nil {
			t.Fatalf("GetTransactionHistory() unexpected error = %v", err)
		}

		// Verify total is correct even with empty page
		if resp.Total != 15 {
			t.Errorf("Total = %v, want 15", resp.Total)
		}
	})
}
