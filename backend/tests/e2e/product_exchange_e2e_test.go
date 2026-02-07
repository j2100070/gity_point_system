// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testBaseURL = "http://localhost:8080/api"
)

// TestProductExchangeFlow は商品交換のE2Eテスト
func TestProductExchangeFlow(t *testing.T) {
	// Cookieを管理するクライアントを作成
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// 1. 管理者としてログイン
	adminCSRF := loginUser(t, client, "admin", "admin123")

	// 2. テスト商品を作成
	productID := createTestProduct(t, client, adminCSRF)

	// 3. テストユーザーとしてログイン
	userCSRF := loginUser(t, client, "testuser", "test123")

	// 4. テストユーザーにポイントを付与（管理者として）
	adminCSRF = loginUser(t, client, "admin", "admin123")
	grantPointsToUser(t, client, adminCSRF, "c26d81a1-5e4f-476e-b390-97824b9a69fd", 5000)

	// 5. テストユーザーとして商品一覧を取得
	userCSRF = loginUser(t, client, "testuser", "test123")
	products := getProductList(t, client, "")
	assert.GreaterOrEqual(t, len(products), 1, "商品が1件以上存在すること")

	// 6. 商品を交換
	exchange := exchangeProduct(t, client, userCSRF, productID, 2, "E2Eテスト交換")
	assert.Equal(t, "completed", exchange.Status, "交換ステータスがcompletedであること")
	assert.Equal(t, 2, exchange.Quantity, "交換数量が2であること")

	// 7. 交換履歴を確認
	history := getExchangeHistory(t, client, userCSRF)
	assert.GreaterOrEqual(t, len(history), 1, "交換履歴が1件以上存在すること")

	found := false
	for _, h := range history {
		if h.ID == exchange.ID {
			found = true
			assert.Equal(t, exchange.PointsUsed, h.PointsUsed, "使用ポイントが一致すること")
		}
	}
	assert.True(t, found, "作成した交換が履歴に存在すること")

	// 8. 管理者として配達完了をマーク
	adminCSRF = loginUser(t, client, "admin", "admin123")
	markDelivered(t, client, adminCSRF, exchange.ID)

	// 9. 管理者として全交換履歴を確認
	allExchanges := getAllExchanges(t, client, adminCSRF)
	assert.GreaterOrEqual(t, len(allExchanges), 1, "全交換履歴が1件以上存在すること")
}

// TestProductManagement は商品管理のE2Eテスト
func TestProductManagement(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// 管理者としてログイン
	csrfToken := loginUser(t, client, "admin", "admin123")

	// 商品を作成
	productID := createTestProduct(t, client, csrfToken)

	// 商品を更新
	updateReq := map[string]interface{}{
		"name":         "更新されたテスト商品",
		"description":  "更新された説明",
		"category":     "drink",
		"price":        150,
		"stock":        5,
		"is_available": true,
	}
	updateProduct(t, client, csrfToken, productID, updateReq)

	// 商品一覧で確認
	products := getProductList(t, client, "")
	found := false
	for _, p := range products {
		if p.ID == productID {
			found = true
			assert.Equal(t, "更新されたテスト商品", p.Name, "商品名が更新されていること")
			assert.Equal(t, int64(150), p.Price, "価格が更新されていること")
		}
	}
	assert.True(t, found, "更新した商品が存在すること")

	// 商品を削除
	deleteProduct(t, client, csrfToken, productID)
}

// Helper functions

func loginUser(t *testing.T, client *http.Client, username, password string) string {
	loginReq := map[string]string{
		"username": username,
		"password": password,
	}

	body, _ := json.Marshal(loginReq)
	resp, err := client.Post(testBaseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err, "ログインリクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "ログインが成功すること")

	var loginResp struct {
		CSRFToken string `json:"csrf_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	require.NoError(t, err, "レスポンスのパースが成功すること")

	return loginResp.CSRFToken
}

func createTestProduct(t *testing.T, client *http.Client, csrfToken string) string {
	createReq := map[string]interface{}{
		"name":        fmt.Sprintf("E2Eテスト商品_%s", uuid.New().String()[:8]),
		"description": "E2Eテスト用の商品",
		"category":    "snack",
		"price":       100,
		"stock":       10,
		"image_url":   "https://example.com/test.jpg",
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", testBaseURL+"/admin/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfToken)

	resp, err := client.Do(req)
	require.NoError(t, err, "商品作成リクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "商品作成が成功すること")

	var createResp struct {
		Product struct {
			ID string `json:"ID"`
		} `json:"Product"`
	}
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	require.NoError(t, err, "レスポンスのパースが成功すること")

	return createResp.Product.ID
}

func getProductList(t *testing.T, client *http.Client, category string) []Product {
	url := testBaseURL + "/products"
	if category != "" {
		url += "?category=" + category
	}

	resp, err := client.Get(url)
	require.NoError(t, err, "商品一覧取得リクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "商品一覧取得が成功すること")

	var listResp struct {
		Products []Product `json:"Products"`
	}
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err, "レスポンスのパースが成功すること")

	return listResp.Products
}

func exchangeProduct(t *testing.T, client *http.Client, csrfToken, productID string, quantity int, notes string) Exchange {
	exchangeReq := map[string]interface{}{
		"product_id": productID,
		"quantity":   quantity,
		"notes":      notes,
	}

	body, _ := json.Marshal(exchangeReq)
	req, _ := http.NewRequest("POST", testBaseURL+"/products/exchange", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfToken)

	resp, err := client.Do(req)
	require.NoError(t, err, "商品交換リクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "商品交換が成功すること")

	var exchangeResp struct {
		Exchange Exchange `json:"Exchange"`
	}
	err = json.NewDecoder(resp.Body).Decode(&exchangeResp)
	require.NoError(t, err, "レスポンスのパースが成功すること")

	return exchangeResp.Exchange
}

func getExchangeHistory(t *testing.T, client *http.Client, csrfToken string) []Exchange {
	req, _ := http.NewRequest("GET", testBaseURL+"/products/exchanges/history", nil)
	req.Header.Set("X-CSRF-Token", csrfToken)

	resp, err := client.Do(req)
	require.NoError(t, err, "交換履歴取得リクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "交換履歴取得が成功すること")

	var historyResp struct {
		Exchanges []Exchange `json:"Exchanges"`
	}
	err = json.NewDecoder(resp.Body).Decode(&historyResp)
	require.NoError(t, err, "レスポンスのパースが成功すること")

	return historyResp.Exchanges
}

func markDelivered(t *testing.T, client *http.Client, csrfToken, exchangeID string) {
	req, _ := http.NewRequest("POST", testBaseURL+"/admin/exchanges/"+exchangeID+"/deliver", nil)
	req.Header.Set("X-CSRF-Token", csrfToken)

	resp, err := client.Do(req)
	require.NoError(t, err, "配達完了リクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "配達完了マークが成功すること")
}

func getAllExchanges(t *testing.T, client *http.Client, csrfToken string) []Exchange {
	req, _ := http.NewRequest("GET", testBaseURL+"/admin/exchanges", nil)
	req.Header.Set("X-CSRF-Token", csrfToken)

	resp, err := client.Do(req)
	require.NoError(t, err, "全交換履歴取得リクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "全交換履歴取得が成功すること")

	var historyResp struct {
		Exchanges []Exchange `json:"Exchanges"`
	}
	err = json.NewDecoder(resp.Body).Decode(&historyResp)
	require.NoError(t, err, "レスポンスのパースが成功すること")

	return historyResp.Exchanges
}

func grantPointsToUser(t *testing.T, client *http.Client, csrfToken, userID string, amount int64) {
	grantReq := map[string]interface{}{
		"user_id":         userID,
		"amount":          amount,
		"description":     "E2Eテスト用ポイント付与",
		"idempotency_key": fmt.Sprintf("e2e-grant-%s", uuid.New().String()),
	}

	body, _ := json.Marshal(grantReq)
	req, _ := http.NewRequest("POST", testBaseURL+"/admin/points/grant", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfToken)

	resp, err := client.Do(req)
	require.NoError(t, err, "ポイント付与リクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "ポイント付与が成功すること")
}

func updateProduct(t *testing.T, client *http.Client, csrfToken, productID string, updateData map[string]interface{}) {
	body, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", testBaseURL+"/admin/products/"+productID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfToken)

	resp, err := client.Do(req)
	require.NoError(t, err, "商品更新リクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "商品更新が成功すること")
}

func deleteProduct(t *testing.T, client *http.Client, csrfToken, productID string) {
	req, _ := http.NewRequest("DELETE", testBaseURL+"/admin/products/"+productID, nil)
	req.Header.Set("X-CSRF-Token", csrfToken)

	resp, err := client.Do(req)
	require.NoError(t, err, "商品削除リクエストが成功すること")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "商品削除が成功すること")
}

// Types

type Product struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Category    string `json:"Category"`
	Price       int64  `json:"Price"`
	Stock       int    `json:"Stock"`
	IsAvailable bool   `json:"IsAvailable"`
}

type Exchange struct {
	ID          string `json:"ID"`
	UserID      string `json:"UserID"`
	ProductID   string `json:"ProductID"`
	Quantity    int    `json:"Quantity"`
	PointsUsed  int64  `json:"PointsUsed"`
	Status      string `json:"Status"`
	Notes       string `json:"Notes"`
}
