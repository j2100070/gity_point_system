package infraakerun

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// AkerunConfig はAkerun APIの設定
type AkerunConfig struct {
	AccessToken    string
	OrganizationID string
	BaseURL        string // デフォルト: https://api.akerun.com
}

// AccessRecord はAkerun入退室履歴レコード
type AccessRecord struct {
	ID         json.Number `json:"id"`
	Action     string      `json:"action"`
	DeviceType string      `json:"device_type"`
	DeviceName string      `json:"device_name"`
	AccessedAt string      `json:"accessed_at"`
	Akerun     *AkerunInfo `json:"akerun"`
	User       *AkerunUser `json:"user"`
}

// AkerunInfo はAkerunデバイス情報
type AkerunInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

// AkerunUser はAkerunユーザー情報
type AkerunUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

// accessesResponse はAkerun APIレスポンス
type accessesResponse struct {
	Accesses []AccessRecord `json:"accesses"`
}

// AkerunClient はAkerun APIクライアント
type AkerunClient struct {
	config     *AkerunConfig
	httpClient *http.Client
}

// NewAkerunClient は新しいAkerunClientを作成
func NewAkerunClient(config *AkerunConfig) *AkerunClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.akerun.com"
	}
	return &AkerunClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAccesses は入退室履歴を取得
func (c *AkerunClient) GetAccesses(ctx context.Context, after, before time.Time) ([]AccessRecord, error) {
	endpoint := fmt.Sprintf("%s/v3/organizations/%s/accesses",
		c.config.BaseURL, c.config.OrganizationID)

	params := url.Values{}
	params.Set("limit", "300")
	params.Set("datetime_after", after.Format(time.RFC3339))
	params.Set("datetime_before", before.Format(time.RFC3339))

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Akerun API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Akerun API Error:\nURL: %s\nStatus: %d\nBody: %s\n", fullURL, resp.StatusCode, string(body))
		return nil, fmt.Errorf("Akerun API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result accessesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Akerun API response: %w", err)
	}

	return result.Accesses, nil
}

// IsConfigured はAkerun APIが設定されているかを返す
func (c *AkerunClient) IsConfigured() bool {
	return c.config.AccessToken != "" && c.config.OrganizationID != ""
}
