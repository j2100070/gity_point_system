// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	baseURL = "http://localhost:8080/api/v1"
)

// Helper functions
func createTestUser(t *testing.T, username, email, password, displayName string) (userID string, sessionToken, csrfToken string) {
	// Register
	registerReq := map[string]string{
		"username":     username,
		"email":        email,
		"password":     password,
		"display_name": displayName,
	}

	body, _ := json.Marshal(registerReq)
	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Register request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Register failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var registerResp struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	json.NewDecoder(resp.Body).Decode(&registerResp)
	userID = registerResp.User.ID

	// Login
	loginReq := map[string]string{
		"username": username,
		"password": password,
	}

	body, _ = json.Marshal(loginReq)
	resp, err = http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp struct {
		SessionToken string `json:"session_token"`
		CSRFToken    string `json:"csrf_token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	return userID, loginResp.SessionToken, loginResp.CSRFToken
}

func getBalance(t *testing.T, sessionToken, csrfToken, userID string) int64 {
	req, _ := http.NewRequest("GET", baseURL+"/users/"+userID+"/balance", nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Get balance request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Get balance failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var balanceResp struct {
		Balance int64 `json:"balance"`
	}
	json.NewDecoder(resp.Body).Decode(&balanceResp)

	return balanceResp.Balance
}

func transferPoints(t *testing.T, sessionToken, csrfToken, fromUserID, toUserID string, amount int64) (transactionID string) {
	transferReq := map[string]interface{}{
		"to_user_id":      toUserID,
		"amount":          amount,
		"idempotency_key": uuid.New().String(),
		"description":     "E2E test transfer",
	}

	body, _ := json.Marshal(transferReq)
	req, _ := http.NewRequest("POST", baseURL+"/points/transfer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+sessionToken)
	req.Header.Set("X-CSRF-Token", csrfToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Transfer request failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Transfer failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var transferResp struct {
		Transaction struct {
			ID string `json:"id"`
		} `json:"transaction"`
	}
	json.Unmarshal(bodyBytes, &transferResp)

	return transferResp.Transaction.ID
}

func TestPointTransfer_E2E(t *testing.T) {
	// Wait for server to be ready
	time.Sleep(2 * time.Second)

	// Create two test users
	sender1ID, sender1Session, sender1CSRF := createTestUser(t,
		fmt.Sprintf("sender_%d", time.Now().Unix()),
		fmt.Sprintf("sender_%d@example.com", time.Now().Unix()),
		"SenderPass123",
		"Sender User 1")

	time.Sleep(100 * time.Millisecond)

	receiver1ID, _, _ := createTestUser(t,
		fmt.Sprintf("receiver_%d", time.Now().Unix()),
		fmt.Sprintf("receiver_%d@example.com", time.Now().Unix()),
		"ReceiverPass123",
		"Receiver User 1")

	t.Run("Transfer points with actual balance verification", func(t *testing.T) {
		// Get initial balances
		senderInitialBalance := getBalance(t, sender1Session, sender1CSRF, sender1ID)
		receiverInitialBalance := getBalance(t, sender1Session, sender1CSRF, receiver1ID)

		t.Logf("Initial balances - Sender: %d, Receiver: %d", senderInitialBalance, receiverInitialBalance)

		// Transfer amount
		transferAmount := int64(1000)

		// Perform transfer
		transactionID := transferPoints(t, sender1Session, sender1CSRF, sender1ID, receiver1ID, transferAmount)
		t.Logf("Transfer completed, transaction ID: %s", transactionID)

		// Wait a bit for transaction to complete
		time.Sleep(500 * time.Millisecond)

		// Verify actual balances after transfer
		senderFinalBalance := getBalance(t, sender1Session, sender1CSRF, sender1ID)
		receiverFinalBalance := getBalance(t, sender1Session, sender1CSRF, receiver1ID)

		t.Logf("Final balances - Sender: %d, Receiver: %d", senderFinalBalance, receiverFinalBalance)

		// Verify exact amounts
		expectedSenderBalance := senderInitialBalance - transferAmount
		expectedReceiverBalance := receiverInitialBalance + transferAmount

		if senderFinalBalance != expectedSenderBalance {
			t.Errorf("Sender balance = %d, want exactly %d (initial %d - transferred %d)",
				senderFinalBalance, expectedSenderBalance, senderInitialBalance, transferAmount)
		}

		if receiverFinalBalance != expectedReceiverBalance {
			t.Errorf("Receiver balance = %d, want exactly %d (initial %d + received %d)",
				receiverFinalBalance, expectedReceiverBalance, receiverInitialBalance, transferAmount)
		}

		// Verify conservation of points
		totalBefore := senderInitialBalance + receiverInitialBalance
		totalAfter := senderFinalBalance + receiverFinalBalance

		if totalBefore != totalAfter {
			t.Errorf("Total points not conserved: before=%d, after=%d", totalBefore, totalAfter)
		}
	})

	t.Run("Multiple transfers with exact balance tracking", func(t *testing.T) {
		// Get current balance
		currentBalance := getBalance(t, sender1Session, sender1CSRF, sender1ID)
		t.Logf("Current sender balance: %d", currentBalance)

		// Perform 3 transfers of different amounts
		amounts := []int64{200, 350, 150}
		totalTransferred := int64(0)

		for i, amount := range amounts {
			transferPoints(t, sender1Session, sender1CSRF, sender1ID, receiver1ID, amount)
			totalTransferred += amount
			time.Sleep(300 * time.Millisecond)

			// Verify balance after each transfer
			balance := getBalance(t, sender1Session, sender1CSRF, sender1ID)
			expectedBalance := currentBalance - totalTransferred

			if balance != expectedBalance {
				t.Errorf("After transfer %d: balance = %d, want exactly %d (initial %d - transferred %d)",
					i+1, balance, expectedBalance, currentBalance, totalTransferred)
			}

			t.Logf("Transfer %d completed: amount=%d, new balance=%d", i+1, amount, balance)
		}

		// Final verification
		finalBalance := getBalance(t, sender1Session, sender1CSRF, sender1ID)
		expectedFinal := currentBalance - totalTransferred

		if finalBalance != expectedFinal {
			t.Errorf("Final balance = %d, want exactly %d (initial %d - total transferred %d)",
				finalBalance, expectedFinal, currentBalance, totalTransferred)
		}

		t.Logf("Total transferred: %d, Final balance: %d", totalTransferred, finalBalance)
	})
}

func TestInsufficientBalance_E2E(t *testing.T) {
	time.Sleep(2 * time.Second)

	sender2ID, sender2Session, sender2CSRF := createTestUser(t,
		fmt.Sprintf("pooruser_%d", time.Now().Unix()),
		fmt.Sprintf("pooruser_%d@example.com", time.Now().Unix()),
		"PoorPass123",
		"Poor User")

	time.Sleep(100 * time.Millisecond)

	receiver2ID, _, _ := createTestUser(t,
		fmt.Sprintf("richreceiver_%d", time.Now().Unix()),
		fmt.Sprintf("richreceiver_%d@example.com", time.Now().Unix()),
		"RichPass123",
		"Rich Receiver")

	// Get current balance
	currentBalance := getBalance(t, sender2Session, sender2CSRF, sender2ID)
	t.Logf("Current balance: %d", currentBalance)

	// Try to transfer more than balance
	excessiveAmount := currentBalance + 1000

	transferReq := map[string]interface{}{
		"to_user_id":      receiver2ID,
		"amount":          excessiveAmount,
		"idempotency_key": uuid.New().String(),
		"description":     "Excessive transfer test",
	}

	body, _ := json.Marshal(transferReq)
	req, _ := http.NewRequest("POST", baseURL+"/points/transfer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+sender2Session)
	req.Header.Set("X-CSRF-Token", sender2CSRF)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	// Should fail with 400 or 422
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Transfer with insufficient balance succeeded, should have failed")
	}

	// Verify balance unchanged
	balanceAfter := getBalance(t, sender2Session, sender2CSRF, sender2ID)
	if balanceAfter != currentBalance {
		t.Errorf("Balance changed after failed transfer: before=%d, after=%d", currentBalance, balanceAfter)
	}

	t.Logf("Insufficient balance test passed: balance remained %d", balanceAfter)
}

func TestIdempotency_E2E(t *testing.T) {
	time.Sleep(2 * time.Second)

	sender3ID, sender3Session, sender3CSRF := createTestUser(t,
		fmt.Sprintf("idemsender_%d", time.Now().Unix()),
		fmt.Sprintf("idemsender_%d@example.com", time.Now().Unix()),
		"IdemPass123",
		"Idempotency Sender")

	time.Sleep(100 * time.Millisecond)

	receiver3ID, _, _ := createTestUser(t,
		fmt.Sprintf("idemreceiver_%d", time.Now().Unix()),
		fmt.Sprintf("idemreceiver_%d@example.com", time.Now().Unix()),
		"IdemPass123",
		"Idempotency Receiver")

	initialBalance := getBalance(t, sender3Session, sender3CSRF, sender3ID)
	transferAmount := int64(500)
	idempotencyKey := uuid.New().String()

	// First transfer
	transferReq := map[string]interface{}{
		"to_user_id":      receiver3ID,
		"amount":          transferAmount,
		"idempotency_key": idempotencyKey,
		"description":     "Idempotency test",
	}

	body, _ := json.Marshal(transferReq)
	req, _ := http.NewRequest("POST", baseURL+"/points/transfer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+sender3Session)
	req.Header.Set("X-CSRF-Token", sender3CSRF)

	client := &http.Client{}
	resp, _ := client.Do(req)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("First transfer failed")
	}

	time.Sleep(500 * time.Millisecond)

	balanceAfterFirst := getBalance(t, sender3Session, sender3CSRF, sender3ID)

	// Retry with same idempotency key
	req, _ = http.NewRequest("POST", baseURL+"/points/transfer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+sender3Session)
	req.Header.Set("X-CSRF-Token", sender3CSRF)

	resp, _ = client.Do(req)
	resp.Body.Close()

	time.Sleep(500 * time.Millisecond)

	balanceAfterRetry := getBalance(t, sender3Session, sender3CSRF, sender3ID)

	// Balance should not change on retry
	if balanceAfterRetry != balanceAfterFirst {
		t.Errorf("Balance changed on idempotent retry: first=%d, retry=%d", balanceAfterFirst, balanceAfterRetry)
	}

	// Verify total deduction is only once
	expectedBalance := initialBalance - transferAmount
	if balanceAfterRetry != expectedBalance {
		t.Errorf("Balance = %d, want exactly %d (only one deduction)", balanceAfterRetry, expectedBalance)
	}

	t.Logf("Idempotency test passed: balance deducted only once (%d -> %d)", initialBalance, balanceAfterRetry)
}
