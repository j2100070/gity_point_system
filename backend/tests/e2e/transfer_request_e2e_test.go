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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const trBaseURL = "http://localhost:8080/api"

// ========================================
// Helper functions
// ========================================

func createTransferRequestTestUser(t *testing.T, username, displayName string) (userID, sessionToken, csrfToken, personalQR string) {
	registerReq := map[string]string{
		"username":     username,
		"email":        username + "@example.com",
		"password":     "TestPass123",
		"display_name": displayName,
	}

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", trBaseURL+"/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Register request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
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
		"password": "TestPass123",
	}

	body, _ = json.Marshal(loginReq)
	loginHttpReq, _ := http.NewRequest("POST", trBaseURL+"/auth/login", bytes.NewBuffer(body))
	loginHttpReq.Header.Set("Content-Type", "application/json")
	loginHttpReq.Header.Set("X-Forwarded-For", "127.0.0.1")

	resp, err = client.Do(loginHttpReq)
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	// SessionトークンをCookieから取得
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session_token" {
			sessionToken = cookie.Value
		}
	}

	var loginResp struct {
		SessionToken string `json:"session_token"`
		CSRFToken    string `json:"csrf_token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	if sessionToken == "" {
		sessionToken = loginResp.SessionToken
	}
	csrfToken = loginResp.CSRFToken

	// Get Personal QR Code
	qrReq, _ := http.NewRequest("GET", trBaseURL+"/transfer-requests/personal-qr", nil)
	qrReq.Header.Set("Cookie", "session_token="+sessionToken)

	resp, err = client.Do(qrReq)
	if err != nil {
		t.Fatalf("Get personal QR request failed: %v", err)
	}
	defer resp.Body.Close()

	var qrResp struct {
		QRCode string `json:"qr_code"`
	}
	json.NewDecoder(resp.Body).Decode(&qrResp)
	personalQR = qrResp.QRCode

	return userID, sessionToken, csrfToken, personalQR
}

func createTransferRequest(t *testing.T, sessionToken, csrfToken, toUserQR string, amount int64, message string) (requestID string, statusCode int) {
	// Extract UUID from QR code (format: "user:{uuid}")
	toUserID := toUserQR
	if len(toUserQR) > 5 && toUserQR[:5] == "user:" {
		toUserID = toUserQR[5:]
	}

	reqBody := map[string]interface{}{
		"to_user_id":      toUserID,
		"amount":          amount,
		"message":         message,
		"idempotency_key": fmt.Sprintf("e2e-test-%d", time.Now().UnixNano()),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", trBaseURL+"/transfer-requests", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+sessionToken)
	req.Header.Set("X-CSRF-Token", csrfToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Create transfer request failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	statusCode = resp.StatusCode

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var result struct {
			TransferRequest struct {
				ID string `json:"id"`
			} `json:"transfer_request"`
		}
		json.Unmarshal(bodyBytes, &result)
		requestID = result.TransferRequest.ID
	}

	return requestID, statusCode
}

func approveTransferRequest(t *testing.T, sessionToken, csrfToken, requestID string) (statusCode int) {
	req, _ := http.NewRequest("POST", trBaseURL+"/transfer-requests/"+requestID+"/approve", nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)
	req.Header.Set("X-CSRF-Token", csrfToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Approve request failed: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

func rejectTransferRequest(t *testing.T, sessionToken, csrfToken, requestID string) (statusCode int) {
	req, _ := http.NewRequest("POST", trBaseURL+"/transfer-requests/"+requestID+"/reject", nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)
	req.Header.Set("X-CSRF-Token", csrfToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Reject request failed: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

func cancelTransferRequest(t *testing.T, sessionToken, csrfToken, requestID string) (statusCode int) {
	req, _ := http.NewRequest("DELETE", trBaseURL+"/transfer-requests/"+requestID, nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)
	req.Header.Set("X-CSRF-Token", csrfToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Cancel request failed: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

func getPendingRequestCount(t *testing.T, sessionToken string) int64 {
	req, _ := http.NewRequest("GET", trBaseURL+"/transfer-requests/pending/count", nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Get pending count failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Count int64 `json:"count"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Count
}

func getPointsBalance(t *testing.T, sessionToken string) int64 {
	req, _ := http.NewRequest("GET", trBaseURL+"/points/balance", nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Get balance failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Balance int64 `json:"balance"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Balance
}

// ========================================
// E2E Transfer Request Tests
// ========================================

func TestTransferRequestE2E_BasicFlow(t *testing.T) {
	time.Sleep(2 * time.Second)

	// ユーザー作成
	senderID, senderSession, senderCSRF, _ := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_sender_%d", time.Now().UnixNano()),
		"TR Sender")

	time.Sleep(100 * time.Millisecond)

	receiverID, receiverSession, receiverCSRF, receiverQR := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_receiver_%d", time.Now().UnixNano()),
		"TR Receiver")

	t.Run("PayPay風送金リクエストフロー", func(t *testing.T) {
		// 初期残高確認
		senderInitialBalance := getPointsBalance(t, senderSession)
		receiverInitialBalance := getPointsBalance(t, receiverSession)

		t.Logf("Initial balances - Sender: %d, Receiver: %d", senderInitialBalance, receiverInitialBalance)

		// 1. 送信者がQRコードをスキャンしてリクエスト作成
		requestID, statusCode := createTransferRequest(t, senderSession, senderCSRF, receiverQR, 1000, "PayPay風テスト送金")
		require.Equal(t, http.StatusOK, statusCode)
		require.NotEmpty(t, requestID)

		t.Logf("Transfer request created: %s", requestID)

		// 2. 受信者の承認待ちカウント確認
		time.Sleep(500 * time.Millisecond)
		pendingCount := getPendingRequestCount(t, receiverSession)
		assert.GreaterOrEqual(t, pendingCount, int64(1))

		t.Logf("Receiver has %d pending requests", pendingCount)

		// 3. 受信者が承認
		approveStatus := approveTransferRequest(t, receiverSession, receiverCSRF, requestID)
		require.Equal(t, http.StatusOK, approveStatus)

		t.Logf("Transfer request approved")

		// 4. 最終残高確認
		time.Sleep(500 * time.Millisecond)

		senderFinalBalance := getPointsBalance(t, senderSession)
		receiverFinalBalance := getPointsBalance(t, receiverSession)

		t.Logf("Final balances - Sender: %d, Receiver: %d", senderFinalBalance, receiverFinalBalance)

		assert.Equal(t, senderInitialBalance-1000, senderFinalBalance)
		assert.Equal(t, receiverInitialBalance+1000, receiverFinalBalance)

		// 承認後はカウントが減る
		pendingCountAfter := getPendingRequestCount(t, receiverSession)
		assert.Less(t, pendingCountAfter, pendingCount)

		t.Logf("Test passed: Sender=%s, Receiver=%s, RequestID=%s", senderID, receiverID, requestID)
	})
}

func TestTransferRequestE2E_RejectFlow(t *testing.T) {
	time.Sleep(2 * time.Second)

	senderID, senderSession, senderCSRF, _ := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_reject_sender_%d", time.Now().UnixNano()),
		"Reject Sender")

	time.Sleep(100 * time.Millisecond)

	receiverID, receiverSession, receiverCSRF, receiverQR := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_reject_receiver_%d", time.Now().UnixNano()),
		"Reject Receiver")

	t.Run("送金リクエスト拒否", func(t *testing.T) {
		senderInitialBalance := getPointsBalance(t, senderSession)
		receiverInitialBalance := getPointsBalance(t, receiverSession)

		// リクエスト作成
		requestID, statusCode := createTransferRequest(t, senderSession, senderCSRF, receiverQR, 2000, "拒否テスト")
		require.Equal(t, http.StatusOK, statusCode)
		require.NotEmpty(t, requestID)

		time.Sleep(300 * time.Millisecond)

		// 受信者が拒否
		rejectStatus := rejectTransferRequest(t, receiverSession, receiverCSRF, requestID)
		require.Equal(t, http.StatusOK, rejectStatus)

		time.Sleep(300 * time.Millisecond)

		// 残高が変わっていないことを確認
		senderFinalBalance := getPointsBalance(t, senderSession)
		receiverFinalBalance := getPointsBalance(t, receiverSession)

		assert.Equal(t, senderInitialBalance, senderFinalBalance)
		assert.Equal(t, receiverInitialBalance, receiverFinalBalance)

		t.Logf("Reject test passed: Sender=%s, Receiver=%s", senderID, receiverID)
	})
}

func TestTransferRequestE2E_CancelFlow(t *testing.T) {
	time.Sleep(2 * time.Second)

	senderID, senderSession, senderCSRF, _ := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_cancel_sender_%d", time.Now().UnixNano()),
		"Cancel Sender")

	time.Sleep(100 * time.Millisecond)

	receiverID, receiverSession, _, receiverQR := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_cancel_receiver_%d", time.Now().UnixNano()),
		"Cancel Receiver")

	t.Run("送金リクエストキャンセル", func(t *testing.T) {
		senderInitialBalance := getPointsBalance(t, senderSession)
		receiverInitialBalance := getPointsBalance(t, receiverSession)

		// リクエスト作成
		requestID, statusCode := createTransferRequest(t, senderSession, senderCSRF, receiverQR, 1500, "キャンセルテスト")
		require.Equal(t, http.StatusOK, statusCode)
		require.NotEmpty(t, requestID)

		time.Sleep(300 * time.Millisecond)

		// 承認待ちカウント確認
		pendingCount := getPendingRequestCount(t, receiverSession)
		assert.GreaterOrEqual(t, pendingCount, int64(1))

		// 送信者がキャンセル
		cancelStatus := cancelTransferRequest(t, senderSession, senderCSRF, requestID)
		require.Equal(t, http.StatusOK, cancelStatus)

		time.Sleep(300 * time.Millisecond)

		// 残高が変わっていないことを確認
		senderFinalBalance := getPointsBalance(t, senderSession)
		receiverFinalBalance := getPointsBalance(t, receiverSession)

		assert.Equal(t, senderInitialBalance, senderFinalBalance)
		assert.Equal(t, receiverInitialBalance, receiverFinalBalance)

		// 承認待ちカウントが減る
		pendingCountAfter := getPendingRequestCount(t, receiverSession)
		assert.Less(t, pendingCountAfter, pendingCount)

		t.Logf("Cancel test passed: Sender=%s, Receiver=%s", senderID, receiverID)
	})
}

func TestTransferRequestE2E_MultipleRequests(t *testing.T) {
	time.Sleep(2 * time.Second)

	sender1ID, sender1Session, sender1CSRF, _ := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_multi_sender1_%d", time.Now().UnixNano()),
		"Multi Sender 1")

	time.Sleep(100 * time.Millisecond)

	sender2ID, sender2Session, sender2CSRF, _ := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_multi_sender2_%d", time.Now().UnixNano()),
		"Multi Sender 2")

	time.Sleep(100 * time.Millisecond)

	receiverID, receiverSession, receiverCSRF, receiverQR := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_multi_receiver_%d", time.Now().UnixNano()),
		"Multi Receiver")

	t.Run("複数の送金リクエスト処理", func(t *testing.T) {
		receiverInitialBalance := getPointsBalance(t, receiverSession)

		// 2人から同時にリクエスト
		req1ID, status1 := createTransferRequest(t, sender1Session, sender1CSRF, receiverQR, 500, "Sender 1")
		require.Equal(t, http.StatusOK, status1)

		req2ID, status2 := createTransferRequest(t, sender2Session, sender2CSRF, receiverQR, 800, "Sender 2")
		require.Equal(t, http.StatusOK, status2)

		time.Sleep(500 * time.Millisecond)

		// 承認待ちカウント確認
		pendingCount := getPendingRequestCount(t, receiverSession)
		assert.GreaterOrEqual(t, pendingCount, int64(2))

		t.Logf("Receiver has %d pending requests", pendingCount)

		// 1つ目を承認、2つ目を拒否
		approveStatus := approveTransferRequest(t, receiverSession, receiverCSRF, req1ID)
		require.Equal(t, http.StatusOK, approveStatus)

		rejectStatus := rejectTransferRequest(t, receiverSession, receiverCSRF, req2ID)
		require.Equal(t, http.StatusOK, rejectStatus)

		time.Sleep(500 * time.Millisecond)

		// 受信者の残高は1つ目のみ加算
		receiverFinalBalance := getPointsBalance(t, receiverSession)
		assert.Equal(t, receiverInitialBalance+500, receiverFinalBalance)

		t.Logf("Multiple requests test passed: Sender1=%s, Sender2=%s, Receiver=%s", sender1ID, sender2ID, receiverID)
	})
}

func TestTransferRequestE2E_ConcurrentApprovals(t *testing.T) {
	time.Sleep(2 * time.Second)

	// 複数の送信者を作成
	senders := make([]struct {
		id      string
		session string
		csrf    string
	}, 3)

	for i := 0; i < 3; i++ {
		id, session, csrf, _ := createTransferRequestTestUser(t,
			fmt.Sprintf("tr_concurrent_sender_%d_%d", i, time.Now().UnixNano()),
			fmt.Sprintf("Concurrent Sender %d", i))
		senders[i] = struct {
			id      string
			session string
			csrf    string
		}{id, session, csrf}
		time.Sleep(50 * time.Millisecond)
	}

	receiverID, receiverSession, receiverCSRF, receiverQR := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_concurrent_receiver_%d", time.Now().UnixNano()),
		"Concurrent Receiver")

	t.Run("並行承認処理", func(t *testing.T) {
		receiverInitialBalance := getPointsBalance(t, receiverSession)

		// 3つのリクエストを作成
		requestIDs := make([]string, 3)
		for i := 0; i < 3; i++ {
			reqID, status := createTransferRequest(t, senders[i].session, senders[i].csrf, receiverQR, int64(100*(i+1)), fmt.Sprintf("Request %d", i))
			require.Equal(t, http.StatusOK, status)
			requestIDs[i] = reqID
			time.Sleep(50 * time.Millisecond)
		}

		time.Sleep(500 * time.Millisecond)

		// 並行承認
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(reqID string) {
				defer wg.Done()
				approveTransferRequest(t, receiverSession, receiverCSRF, reqID)
			}(requestIDs[i])
		}

		wg.Wait()
		time.Sleep(1 * time.Second)

		// 最終残高確認（100 + 200 + 300 = 600）
		receiverFinalBalance := getPointsBalance(t, receiverSession)
		assert.Equal(t, receiverInitialBalance+600, receiverFinalBalance)

		t.Logf("Concurrent approvals test passed: Receiver=%s, Final Balance=%d", receiverID, receiverFinalBalance)
	})
}

func TestTransferRequestE2E_IdempotencyKey(t *testing.T) {
	time.Sleep(2 * time.Second)

	senderID, senderSession, senderCSRF, _ := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_idem_sender_%d", time.Now().UnixNano()),
		"Idempotency Sender")

	time.Sleep(100 * time.Millisecond)

	receiverID, receiverSession, _, receiverQR := createTransferRequestTestUser(t,
		fmt.Sprintf("tr_idem_receiver_%d", time.Now().UnixNano()),
		"Idempotency Receiver")

	t.Run("冪等性キーによる重複防止", func(t *testing.T) {
		idempotencyKey := fmt.Sprintf("fixed-key-%d", time.Now().UnixNano())

		// 同じ冪等性キーで2回リクエスト
		reqBody := map[string]interface{}{
			"to_user_id":      receiverQR,
			"amount":          1000,
			"message":         "Idempotency test",
			"idempotency_key": idempotencyKey,
		}
		body1, _ := json.Marshal(reqBody)

		req1, _ := http.NewRequest("POST", trBaseURL+"/transfer-requests", bytes.NewBuffer(body1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Cookie", "session_token="+senderSession)
		req1.Header.Set("X-CSRF-Token", senderCSRF)

		client := &http.Client{}
		resp1, _ := client.Do(req1)
		defer resp1.Body.Close()

		bodyBytes1, _ := io.ReadAll(resp1.Body)
		var result1 struct {
			TransferRequest struct {
				ID string `json:"id"`
			} `json:"transfer_request"`
		}
		json.Unmarshal(bodyBytes1, &result1)
		requestID1 := result1.TransferRequest.ID

		time.Sleep(200 * time.Millisecond)

		// 2回目のリクエスト（同じキー）
		body2, _ := json.Marshal(reqBody)
		req2, _ := http.NewRequest("POST", trBaseURL+"/transfer-requests", bytes.NewBuffer(body2))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Cookie", "session_token="+senderSession)
		req2.Header.Set("X-CSRF-Token", senderCSRF)

		resp2, _ := client.Do(req2)
		defer resp2.Body.Close()

		bodyBytes2, _ := io.ReadAll(resp2.Body)
		var result2 struct {
			TransferRequest struct {
				ID string `json:"id"`
			} `json:"transfer_request"`
		}
		json.Unmarshal(bodyBytes2, &result2)
		requestID2 := result2.TransferRequest.ID

		// 同じリクエストIDが返される
		assert.Equal(t, requestID1, requestID2)

		// 承認待ちカウントは1のまま
		pendingCount := getPendingRequestCount(t, receiverSession)
		assert.Equal(t, int64(1), pendingCount)

		t.Logf("Idempotency test passed: Sender=%s, Receiver=%s, RequestID=%s", senderID, receiverID, requestID1)
	})
}
