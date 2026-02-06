// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8080/api/v1"
)

// ========================================
// Helper Functions
// ========================================

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

func transferPoints(t *testing.T, sessionToken, csrfToken, fromUserID, toUserID string, amount int64, idempotencyKey string) (transactionID string, statusCode int) {
	transferReq := map[string]interface{}{
		"to_user_id":      toUserID,
		"amount":          amount,
		"idempotency_key": idempotencyKey,
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
	statusCode = resp.StatusCode

	if resp.StatusCode == http.StatusOK {
		var transferResp struct {
			Transaction struct {
				ID string `json:"id"`
			} `json:"transaction"`
		}
		json.Unmarshal(bodyBytes, &transferResp)
		transactionID = transferResp.Transaction.ID
	}

	return transactionID, statusCode
}

// ========================================
// E2E Transaction Tests
// ========================================

func TestPointTransferE2E_BasicScenario(t *testing.T) {
	time.Sleep(2 * time.Second)

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

	t.Run("基本的なポイント送金", func(t *testing.T) {
		senderInitialBalance := getBalance(t, sender1Session, sender1CSRF, sender1ID)
		receiverInitialBalance := getBalance(t, sender1Session, sender1CSRF, receiver1ID)

		t.Logf("Initial balances - Sender: %d, Receiver: %d", senderInitialBalance, receiverInitialBalance)

		transferAmount := int64(1000)
		txID, statusCode := transferPoints(t, sender1Session, sender1CSRF, sender1ID, receiver1ID, transferAmount, uuid.New().String())

		require.Equal(t, http.StatusOK, statusCode)
		require.NotEmpty(t, txID)
		t.Logf("Transfer completed, transaction ID: %s", txID)

		time.Sleep(500 * time.Millisecond)

		senderFinalBalance := getBalance(t, sender1Session, sender1CSRF, sender1ID)
		receiverFinalBalance := getBalance(t, sender1Session, sender1CSRF, receiver1ID)

		t.Logf("Final balances - Sender: %d, Receiver: %d", senderFinalBalance, receiverFinalBalance)

		assert.Equal(t, senderInitialBalance-transferAmount, senderFinalBalance)
		assert.Equal(t, receiverInitialBalance+transferAmount, receiverFinalBalance)

		// ポイント保存則の検証
		totalBefore := senderInitialBalance + receiverInitialBalance
		totalAfter := senderFinalBalance + receiverFinalBalance
		assert.Equal(t, totalBefore, totalAfter, "Total points must be conserved")
	})
}

func TestPointTransferE2E_ConcurrentTransfers(t *testing.T) {
	time.Sleep(2 * time.Second)

	// 3人のユーザーを作成
	users := make([]struct {
		id      string
		session string
		csrf    string
	}, 3)

	for i := 0; i < 3; i++ {
		id, session, csrf := createTestUser(t,
			fmt.Sprintf("concurrent_user_%d_%d", i, time.Now().UnixNano()),
			fmt.Sprintf("concurrent_user_%d_%d@example.com", i, time.Now().UnixNano()),
			"ConcurrentPass123",
			fmt.Sprintf("Concurrent User %d", i),
		)
		users[i] = struct {
			id      string
			session string
			csrf    string
		}{id, session, csrf}
		time.Sleep(100 * time.Millisecond)
	}

	t.Run("複数ユーザー間での並行送金", func(t *testing.T) {
		// 初期残高を取得
		initialBalances := make([]int64, 3)
		for i := 0; i < 3; i++ {
			initialBalances[i] = getBalance(t, users[i].session, users[i].csrf, users[i].id)
		}
		initialTotal := initialBalances[0] + initialBalances[1] + initialBalances[2]

		t.Logf("Initial balances: %v, Total: %d", initialBalances, initialTotal)

		var wg sync.WaitGroup

		// 並行送金シナリオ
		transfers := []struct {
			from   int
			to     int
			amount int64
		}{
			{0, 1, 500},
			{1, 2, 300},
			{2, 0, 400},
			{0, 2, 200},
			{1, 0, 150},
		}

		for _, transfer := range transfers {
			wg.Add(1)
			go func(tf struct {
				from   int
				to     int
				amount int64
			}) {
				defer wg.Done()
				_, statusCode := transferPoints(
					t,
					users[tf.from].session,
					users[tf.from].csrf,
					users[tf.from].id,
					users[tf.to].id,
					tf.amount,
					uuid.New().String(),
				)
				if statusCode != http.StatusOK {
					t.Logf("Transfer from user %d to user %d failed with status %d", tf.from, tf.to, statusCode)
				}
			}(transfer)
			time.Sleep(50 * time.Millisecond)
		}

		wg.Wait()
		time.Sleep(1 * time.Second)

		// 最終残高を取得
		finalBalances := make([]int64, 3)
		for i := 0; i < 3; i++ {
			finalBalances[i] = getBalance(t, users[i].session, users[i].csrf, users[i].id)
		}
		finalTotal := finalBalances[0] + finalBalances[1] + finalBalances[2]

		t.Logf("Final balances: %v, Total: %d", finalBalances, finalTotal)

		// ポイント保存則の検証
		assert.Equal(t, initialTotal, finalTotal, "Total points must be conserved across concurrent transfers")
	})
}

func TestPointTransferE2E_IdempotencyKey(t *testing.T) {
	time.Sleep(2 * time.Second)

	senderID, senderSession, senderCSRF := createTestUser(t,
		fmt.Sprintf("idem_sender_%d", time.Now().UnixNano()),
		fmt.Sprintf("idem_sender_%d@example.com", time.Now().UnixNano()),
		"IdemPass123",
		"Idempotency Sender",
	)

	time.Sleep(100 * time.Millisecond)

	receiverID, _, _ := createTestUser(t,
		fmt.Sprintf("idem_receiver_%d", time.Now().UnixNano()),
		fmt.Sprintf("idem_receiver_%d@example.com", time.Now().UnixNano()),
		"IdemPass123",
		"Idempotency Receiver",
	)

	t.Run("Idempotency Key による重複防止", func(t *testing.T) {
		initialSenderBalance := getBalance(t, senderSession, senderCSRF, senderID)
		initialReceiverBalance := getBalance(t, senderSession, senderCSRF, receiverID)

		idempotencyKey := uuid.New().String()
		amount := int64(1000)

		// 1回目のリクエスト
		txID1, status1 := transferPoints(t, senderSession, senderCSRF, senderID, receiverID, amount, idempotencyKey)
		require.Equal(t, http.StatusOK, status1)
		require.NotEmpty(t, txID1)

		time.Sleep(500 * time.Millisecond)

		// 2回目のリクエスト（同じIdempotency Key）
		txID2, status2 := transferPoints(t, senderSession, senderCSRF, senderID, receiverID, amount, idempotencyKey)

		// 2回目も成功するが、実際には処理されない（冪等性）
		if status2 == http.StatusOK {
			assert.Equal(t, txID1, txID2, "Same transaction ID should be returned for idempotent request")
		}

		time.Sleep(500 * time.Millisecond)

		// 最終残高確認
		finalSenderBalance := getBalance(t, senderSession, senderCSRF, senderID)
		finalReceiverBalance := getBalance(t, senderSession, senderCSRF, receiverID)

		// 1回のみ減算/加算されていることを確認
		assert.Equal(t, initialSenderBalance-amount, finalSenderBalance, "Amount should be deducted only once")
		assert.Equal(t, initialReceiverBalance+amount, finalReceiverBalance, "Amount should be credited only once")

		t.Logf("Idempotency test passed: Initial Sender=%d, Final Sender=%d", initialSenderBalance, finalSenderBalance)
	})
}

func TestPointTransferE2E_InsufficientBalance(t *testing.T) {
	time.Sleep(2 * time.Second)

	senderID, senderSession, senderCSRF := createTestUser(t,
		fmt.Sprintf("poor_sender_%d", time.Now().UnixNano()),
		fmt.Sprintf("poor_sender_%d@example.com", time.Now().UnixNano()),
		"PoorPass123",
		"Poor Sender",
	)

	time.Sleep(100 * time.Millisecond)

	receiverID, _, _ := createTestUser(t,
		fmt.Sprintf("rich_receiver_%d", time.Now().UnixNano()),
		fmt.Sprintf("rich_receiver_%d@example.com", time.Now().UnixNano()),
		"RichPass123",
		"Rich Receiver",
	)

	t.Run("残高不足での送金失敗", func(t *testing.T) {
		currentBalance := getBalance(t, senderSession, senderCSRF, senderID)
		t.Logf("Current balance: %d", currentBalance)

		// 残高を超える金額を送金
		excessiveAmount := currentBalance + 1000

		_, statusCode := transferPoints(t, senderSession, senderCSRF, senderID, receiverID, excessiveAmount, uuid.New().String())

		// 失敗することを確認
		assert.NotEqual(t, http.StatusOK, statusCode, "Transfer should fail with insufficient balance")

		time.Sleep(300 * time.Millisecond)

		// 残高が変わっていないことを確認
		finalBalance := getBalance(t, senderSession, senderCSRF, senderID)
		assert.Equal(t, currentBalance, finalBalance, "Balance should remain unchanged after failed transfer")

		t.Logf("Insufficient balance test passed: balance remained %d", finalBalance)
	})
}

func TestPointTransferE2E_MultipleSequentialTransfers(t *testing.T) {
	time.Sleep(2 * time.Second)

	senderID, senderSession, senderCSRF := createTestUser(t,
		fmt.Sprintf("seq_sender_%d", time.Now().UnixNano()),
		fmt.Sprintf("seq_sender_%d@example.com", time.Now().UnixNano()),
		"SeqPass123",
		"Sequential Sender",
	)

	time.Sleep(100 * time.Millisecond)

	receiverID, _, _ := createTestUser(t,
		fmt.Sprintf("seq_receiver_%d", time.Now().UnixNano()),
		fmt.Sprintf("seq_receiver_%d@example.com", time.Now().UnixNano()),
		"SeqPass123",
		"Sequential Receiver",
	)

	t.Run("連続した複数回の送金", func(t *testing.T) {
		initialBalance := getBalance(t, senderSession, senderCSRF, senderID)
		t.Logf("Initial balance: %d", initialBalance)

		amounts := []int64{200, 350, 150, 400, 250}
		totalTransferred := int64(0)

		for i, amount := range amounts {
			txID, statusCode := transferPoints(t, senderSession, senderCSRF, senderID, receiverID, amount, uuid.New().String())
			require.Equal(t, http.StatusOK, statusCode, "Transfer %d should succeed", i+1)
			require.NotEmpty(t, txID)

			totalTransferred += amount
			time.Sleep(300 * time.Millisecond)

			// 各送金後の残高確認
			currentBalance := getBalance(t, senderSession, senderCSRF, senderID)
			expectedBalance := initialBalance - totalTransferred
			assert.Equal(t, expectedBalance, currentBalance, "Balance after transfer %d should be correct", i+1)

			t.Logf("Transfer %d: amount=%d, new balance=%d", i+1, amount, currentBalance)
		}

		// 最終確認
		finalBalance := getBalance(t, senderSession, senderCSRF, senderID)
		expectedFinal := initialBalance - totalTransferred
		assert.Equal(t, expectedFinal, finalBalance, "Final balance should match expected value")

		t.Logf("Sequential transfers completed: Initial=%d, Transferred=%d, Final=%d",
			initialBalance, totalTransferred, finalBalance)
	})
}

func TestPointTransferE2E_RaceCondition(t *testing.T) {
	time.Sleep(2 * time.Second)

	senderID, senderSession, senderCSRF := createTestUser(t,
		fmt.Sprintf("race_sender_%d", time.Now().UnixNano()),
		fmt.Sprintf("race_sender_%d@example.com", time.Now().UnixNano()),
		"RacePass123",
		"Race Sender",
	)

	time.Sleep(100 * time.Millisecond)

	// 複数の受信者を作成
	receivers := make([]string, 5)
	for i := 0; i < 5; i++ {
		receiverID, _, _ := createTestUser(t,
			fmt.Sprintf("race_receiver_%d_%d", i, time.Now().UnixNano()),
			fmt.Sprintf("race_receiver_%d_%d@example.com", i, time.Now().UnixNano()),
			"RacePass123",
			fmt.Sprintf("Race Receiver %d", i),
		)
		receivers[i] = receiverID
		time.Sleep(50 * time.Millisecond)
	}

	t.Run("競合状態での同時送金", func(t *testing.T) {
		initialBalance := getBalance(t, senderSession, senderCSRF, senderID)
		t.Logf("Initial balance: %d", initialBalance)

		var wg sync.WaitGroup
		amount := int64(100)
		successCount := 0
		var mu sync.Mutex

		// 5人に同時送金
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(receiverID string) {
				defer wg.Done()
				_, statusCode := transferPoints(t, senderSession, senderCSRF, senderID, receiverID, amount, uuid.New().String())
				if statusCode == http.StatusOK {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}(receivers[i])
		}

		wg.Wait()
		time.Sleep(1 * time.Second)

		// 最終残高確認
		finalBalance := getBalance(t, senderSession, senderCSRF, senderID)
		expectedBalance := initialBalance - (int64(successCount) * amount)

		assert.Equal(t, expectedBalance, finalBalance, "Final balance should match expected after concurrent transfers")

		t.Logf("Race condition test: Initial=%d, Success=%d, Final=%d, Expected=%d",
			initialBalance, successCount, finalBalance, expectedBalance)
	})
}
