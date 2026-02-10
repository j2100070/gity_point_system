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
)

// ========================================
// Helper functions
// ========================================

func createFriendTestUser(t *testing.T, username, displayName string) (userID, sessionToken, csrfToken string) {
	registerReq := map[string]string{
		"username":     username,
		"email":        username + "@example.com",
		"password":     "TestPass123",
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
		"password": "TestPass123",
	}

	body, _ = json.Marshal(loginReq)
	resp, err = http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	var loginResp struct {
		SessionToken string `json:"session_token"`
		CSRFToken    string `json:"csrf_token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	return userID, loginResp.SessionToken, loginResp.CSRFToken
}

func sendFriendRequest(t *testing.T, sessionToken, csrfToken, addresseeID string) string {
	reqBody := map[string]string{"addressee_id": addresseeID}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", baseURL+"/friends/requests", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+sessionToken)
	req.Header.Set("X-CSRF-Token", csrfToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Send friend request failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Send friend request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Friendship struct {
			ID string `json:"id"`
		} `json:"friendship"`
	}
	json.Unmarshal(bodyBytes, &result)

	return result.Friendship.ID
}

func acceptFriendRequest(t *testing.T, sessionToken, csrfToken, friendshipID string) {
	req, _ := http.NewRequest("POST", baseURL+"/friends/requests/"+friendshipID+"/accept", nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)
	req.Header.Set("X-CSRF-Token", csrfToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Accept request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Accept request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
}

func rejectFriendRequest(t *testing.T, sessionToken, csrfToken, friendshipID string) {
	req, _ := http.NewRequest("POST", baseURL+"/friends/requests/"+friendshipID+"/reject", nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)
	req.Header.Set("X-CSRF-Token", csrfToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Reject request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Reject request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
}

func removeFriend(t *testing.T, sessionToken, csrfToken, friendshipID string) {
	req, _ := http.NewRequest("DELETE", baseURL+"/friends/"+friendshipID, nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)
	req.Header.Set("X-CSRF-Token", csrfToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Remove friend failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Remove friend failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
}

func getFriends(t *testing.T, sessionToken, csrfToken string) []map[string]interface{} {
	req, _ := http.NewRequest("GET", baseURL+"/friends", nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Get friends failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get friends failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Friends []map[string]interface{} `json:"friends"`
	}
	json.Unmarshal(bodyBytes, &result)

	return result.Friends
}

func getPendingRequests(t *testing.T, sessionToken, csrfToken string) []map[string]interface{} {
	req, _ := http.NewRequest("GET", baseURL+"/friends/requests", nil)
	req.Header.Set("Cookie", "session_token="+sessionToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Get pending requests failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get pending requests failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Requests []map[string]interface{} `json:"requests"`
	}
	json.Unmarshal(bodyBytes, &result)

	return result.Requests
}

// ========================================
// E2E Tests
// ========================================

func TestFriendRequest_E2E(t *testing.T) {
	time.Sleep(2 * time.Second)

	ts := time.Now().Unix()
	userAID, userASession, userACSRF := createFriendTestUser(t,
		fmt.Sprintf("friend_a_%d", ts), "Friend A")
	time.Sleep(100 * time.Millisecond)
	userBID, userBSession, userBCSRF := createFriendTestUser(t,
		fmt.Sprintf("friend_b_%d", ts), "Friend B")

	t.Run("フレンド申請→承認のフルフロー", func(t *testing.T) {
		// 1. AからBにフレンド申請
		friendshipID := sendFriendRequest(t, userASession, userACSRF, userBID)
		t.Logf("Friend request sent, friendshipID: %s", friendshipID)

		// 2. Bの保留中申請リストに表示される
		pending := getPendingRequests(t, userBSession, userBCSRF)
		if len(pending) == 0 {
			t.Fatal("Expected pending requests for user B")
		}
		t.Logf("User B has %d pending request(s)", len(pending))

		// 3. Bが承認
		acceptFriendRequest(t, userBSession, userBCSRF, friendshipID)
		t.Log("Friend request accepted")

		// 4. Aの友達リストにBが表示される
		friendsA := getFriends(t, userASession, userACSRF)
		if len(friendsA) == 0 {
			t.Fatal("Expected friends for user A after acceptance")
		}
		t.Logf("User A has %d friend(s)", len(friendsA))

		// 5. Bの友達リストにもAが表示される
		friendsB := getFriends(t, userBSession, userBCSRF)
		if len(friendsB) == 0 {
			t.Fatal("Expected friends for user B after acceptance")
		}
		t.Logf("User B has %d friend(s)", len(friendsB))
	})

	_ = userAID // linter
}

func TestFriendReject_E2E(t *testing.T) {
	time.Sleep(2 * time.Second)

	ts := time.Now().Unix()
	_, userASession, userACSRF := createFriendTestUser(t,
		fmt.Sprintf("reject_a_%d", ts), "Reject A")
	time.Sleep(100 * time.Millisecond)
	userBID, userBSession, userBCSRF := createFriendTestUser(t,
		fmt.Sprintf("reject_b_%d", ts), "Reject B")

	t.Run("フレンド申請→拒否のフロー", func(t *testing.T) {
		// 1. AからBにフレンド申請
		friendshipID := sendFriendRequest(t, userASession, userACSRF, userBID)
		t.Logf("Friend request sent, friendshipID: %s", friendshipID)

		// 2. Bが拒否
		rejectFriendRequest(t, userBSession, userBCSRF, friendshipID)
		t.Log("Friend request rejected")

		// 3. どちらの友達リストにも表示されない
		friendsA := getFriends(t, userASession, userACSRF)
		if len(friendsA) != 0 {
			t.Errorf("Expected 0 friends for user A, got %d", len(friendsA))
		}
	})
}

func TestFriendRemoveWithArchive_E2E(t *testing.T) {
	time.Sleep(2 * time.Second)

	ts := time.Now().Unix()
	_, userASession, userACSRF := createFriendTestUser(t,
		fmt.Sprintf("remove_a_%d", ts), "Remove A")
	time.Sleep(100 * time.Millisecond)
	userBID, userBSession, userBCSRF := createFriendTestUser(t,
		fmt.Sprintf("remove_b_%d", ts), "Remove B")

	t.Run("フレンド申請→承認→解散（アーカイブ）のフルフロー", func(t *testing.T) {
		// 1. フレンド申請→承認
		friendshipID := sendFriendRequest(t, userASession, userACSRF, userBID)
		acceptFriendRequest(t, userBSession, userBCSRF, friendshipID)

		// 友達リストに表示される
		friends := getFriends(t, userASession, userACSRF)
		if len(friends) == 0 {
			t.Fatal("Expected friends after acceptance")
		}
		t.Logf("Before removal: User A has %d friend(s)", len(friends))

		// 2. Aがフレンド解散
		removeFriend(t, userASession, userACSRF, friendshipID)
		t.Log("Friend removed (archived)")

		// 3. 友達リストから消える
		friendsAfter := getFriends(t, userASession, userACSRF)
		if len(friendsAfter) != 0 {
			t.Errorf("Expected 0 friends after removal, got %d", len(friendsAfter))
		}

		friendsBAfter := getFriends(t, userBSession, userBCSRF)
		if len(friendsBAfter) != 0 {
			t.Errorf("Expected 0 friends for user B after removal, got %d", len(friendsBAfter))
		}

		t.Log("Friend successfully removed from both users' lists")
	})
}

func TestFriendRejectAndReRequest_E2E(t *testing.T) {
	time.Sleep(2 * time.Second)

	ts := time.Now().Unix()
	_, userASession, userACSRF := createFriendTestUser(t,
		fmt.Sprintf("rereq_a_%d", ts), "ReReq A")
	time.Sleep(100 * time.Millisecond)
	userBID, userBSession, userBCSRF := createFriendTestUser(t,
		fmt.Sprintf("rereq_b_%d", ts), "ReReq B")

	t.Run("拒否後の再申請が成功する", func(t *testing.T) {
		// 1. フレンド申請→拒否
		friendshipID := sendFriendRequest(t, userASession, userACSRF, userBID)
		rejectFriendRequest(t, userBSession, userBCSRF, friendshipID)
		t.Log("First request rejected")

		// 2. 再申請（これが以前のバグパターン: unique constraint violation）
		newFriendshipID := sendFriendRequest(t, userASession, userACSRF, userBID)
		t.Logf("Re-request sent successfully, friendshipID: %s", newFriendshipID)

		// 3. Bの保留中リストに表示される
		pending := getPendingRequests(t, userBSession, userBCSRF)
		if len(pending) == 0 {
			t.Fatal("Expected pending request after re-request")
		}

		// 4. 今度は承認
		acceptFriendRequest(t, userBSession, userBCSRF, newFriendshipID)
		t.Log("Re-request accepted")

		// 5. 友達になっている
		friends := getFriends(t, userASession, userACSRF)
		if len(friends) == 0 {
			t.Fatal("Expected friends after re-request acceptance")
		}

		t.Log("Reject → Re-request → Accept flow passed successfully")
	})
}

func TestUnauthorizedFriendActions_E2E(t *testing.T) {
	time.Sleep(2 * time.Second)

	ts := time.Now().Unix()
	_, userASession, userACSRF := createFriendTestUser(t,
		fmt.Sprintf("unauth_a_%d", ts), "Unauth A")
	time.Sleep(100 * time.Millisecond)
	userBID, _, _ := createFriendTestUser(t,
		fmt.Sprintf("unauth_b_%d", ts), "Unauth B")
	time.Sleep(100 * time.Millisecond)
	_, userCSession, userCCSRF := createFriendTestUser(t,
		fmt.Sprintf("unauth_c_%d", ts), "Unauth C")

	t.Run("申請者自身が承認しようとするとエラー", func(t *testing.T) {
		friendshipID := sendFriendRequest(t, userASession, userACSRF, userBID)

		// Aが自分の申請を承認しようとする
		req, _ := http.NewRequest("POST", baseURL+"/friends/requests/"+friendshipID+"/accept", nil)
		req.Header.Set("Cookie", "session_token="+userASession)
		req.Header.Set("X-CSRF-Token", userACSRF)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("Requester should not be able to accept their own request")
		}
		t.Logf("Self-accept correctly rejected with status %d", resp.StatusCode)
	})

	t.Run("無関係のユーザーが解散しようとするとエラー", func(t *testing.T) {
		// 既存のフレンドシップ（前のテストで作成済み）を取得
		friends := getFriends(t, userASession, userACSRF)
		// この時点でまだpending状態なので友達リストは空
		// 新しいフレンドシップを作成して承認
		friendshipID := sendFriendRequest(t, userASession, userACSRF, userBID)
		// Bはログインしていないので、Cが解散を試みる

		req, _ := http.NewRequest("DELETE", baseURL+"/friends/"+friendshipID, nil)
		req.Header.Set("Cookie", "session_token="+userCSession)
		req.Header.Set("X-CSRF-Token", userCCSRF)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("Unrelated user should not be able to remove friendship")
		}
		t.Logf("Unauthorized removal correctly rejected with status %d", resp.StatusCode)

		_ = friends // linter
	})
}
