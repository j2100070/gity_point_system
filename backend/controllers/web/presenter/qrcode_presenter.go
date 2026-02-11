package presenter

import (
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// QRCodePresenter はQRコード機能のプレゼンター
type QRCodePresenter struct{}

// NewQRCodePresenter は新しいQRCodePresenterを作成
func NewQRCodePresenter() *QRCodePresenter {
	return &QRCodePresenter{}
}

// QRCodeResponse はQRコードのレスポンス
type QRCodeResponse struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Code         string     `json:"code"`
	QRType       string     `json:"qr_type"`
	Amount       *int64     `json:"amount,omitempty"`
	IsUsed       bool       `json:"is_used"`
	UsedByUserID *uuid.UUID `json:"used_by_user_id,omitempty"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// PresentGenerateReceiveQR は受取用QRコード生成レスポンスを生成
func (p *QRCodePresenter) PresentGenerateReceiveQR(resp *inputport.GenerateReceiveQRResponse) map[string]interface{} {
	return map[string]interface{}{
		"qr_code":      p.toQRCodeResponse(resp.QRCode),
		"qr_code_data": resp.QRCodeData,
	}
}

// PresentGenerateSendQR は送信用QRコード生成レスポンスを生成
func (p *QRCodePresenter) PresentGenerateSendQR(resp *inputport.GenerateSendQRResponse) map[string]interface{} {
	return map[string]interface{}{
		"qr_code":      p.toQRCodeResponse(resp.QRCode),
		"qr_code_data": resp.QRCodeData,
	}
}

// PresentScanQR はQRコードスキャンレスポンスを生成
func (p *QRCodePresenter) PresentScanQR(resp *inputport.ScanQRResponse) map[string]interface{} {
	return map[string]interface{}{
		"transaction": TransactionResponse{
			ID:          resp.Transaction.ID,
			FromUserID:  resp.Transaction.FromUserID,
			ToUserID:    resp.Transaction.ToUserID,
			Amount:      resp.Transaction.Amount,
			Description: resp.Transaction.Description,
			CreatedAt:   resp.Transaction.CreatedAt,
		},
		"qr_code": p.toQRCodeResponse(resp.QRCode),
		"from_user": UserResponse{
			ID:          resp.FromUser.ID,
			Username:    resp.FromUser.Username,
			DisplayName: resp.FromUser.DisplayName,
			AvatarURL:   resp.FromUser.AvatarURL,
			Balance:     resp.FromUser.Balance,
			Role:        string(resp.FromUser.Role),
			IsActive:    resp.FromUser.IsActive,
			CreatedAt:   resp.FromUser.CreatedAt,
			UpdatedAt:   resp.FromUser.UpdatedAt,
		},
		"to_user": UserResponse{
			ID:          resp.ToUser.ID,
			Username:    resp.ToUser.Username,
			DisplayName: resp.ToUser.DisplayName,
			AvatarURL:   resp.ToUser.AvatarURL,
			Balance:     resp.ToUser.Balance,
			Role:        string(resp.ToUser.Role),
			IsActive:    resp.ToUser.IsActive,
			CreatedAt:   resp.ToUser.CreatedAt,
			UpdatedAt:   resp.ToUser.UpdatedAt,
		},
	}
}

// PresentGetQRCodeHistory はQRコード履歴レスポンスを生成
func (p *QRCodePresenter) PresentGetQRCodeHistory(resp *inputport.GetQRCodeHistoryResponse) map[string]interface{} {
	qrCodes := make([]QRCodeResponse, 0, len(resp.QRCodes))
	for _, qr := range resp.QRCodes {
		qrCodes = append(qrCodes, p.toQRCodeResponse(qr))
	}

	return map[string]interface{}{
		"qr_codes": qrCodes,
	}
}

// toQRCodeResponse はQRCodeエンティティをレスポンスに変換
func (p *QRCodePresenter) toQRCodeResponse(qrCode *entities.QRCode) QRCodeResponse {
	return QRCodeResponse{
		ID:           qrCode.ID,
		UserID:       qrCode.UserID,
		Code:         qrCode.Code,
		QRType:       string(qrCode.QRType),
		Amount:       qrCode.Amount,
		IsUsed:       qrCode.IsUsed(),
		UsedByUserID: qrCode.UsedByUserID,
		UsedAt:       qrCode.UsedAt,
		ExpiresAt:    qrCode.ExpiresAt,
		CreatedAt:    qrCode.CreatedAt,
	}
}
