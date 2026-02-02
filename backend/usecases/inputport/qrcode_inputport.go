package inputport

import (
	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// QRCodeInputPort はQRコード機能のユースケースインターフェース
type QRCodeInputPort interface {
	// GenerateReceiveQR は受取用QRコードを生成
	GenerateReceiveQR(req *GenerateReceiveQRRequest) (*GenerateReceiveQRResponse, error)

	// GenerateSendQR は送信用QRコードを生成
	GenerateSendQR(req *GenerateSendQRRequest) (*GenerateSendQRResponse, error)

	// ScanQR はQRコードをスキャンしてポイント転送
	ScanQR(req *ScanQRRequest) (*ScanQRResponse, error)

	// GetQRCodeHistory はQRコード履歴を取得
	GetQRCodeHistory(req *GetQRCodeHistoryRequest) (*GetQRCodeHistoryResponse, error)
}

// GenerateReceiveQRRequest は受取用QRコード生成リクエスト
type GenerateReceiveQRRequest struct {
	UserID uuid.UUID
	Amount *int64 // nil=送信者が金額指定、値あり=固定額
}

// GenerateReceiveQRResponse は受取用QRコード生成レスポンス
type GenerateReceiveQRResponse struct {
	QRCode     *entities.QRCode
	QRCodeData string // QRコードに含めるデータ
}

// GenerateSendQRRequest は送信用QRコード生成リクエスト
type GenerateSendQRRequest struct {
	UserID uuid.UUID
	Amount int64
}

// GenerateSendQRResponse は送信用QRコード生成レスポンス
type GenerateSendQRResponse struct {
	QRCode     *entities.QRCode
	QRCodeData string
}

// ScanQRRequest はQRコードスキャンリクエスト
type ScanQRRequest struct {
	UserID         uuid.UUID
	Code           string
	Amount         *int64  // QRコードに金額が含まれていない場合に指定
	IdempotencyKey string
}

// ScanQRResponse はQRコードスキャンレスポンス
type ScanQRResponse struct {
	Transaction *entities.Transaction
	QRCode      *entities.QRCode
	FromUser    *entities.User
	ToUser      *entities.User
}

// GetQRCodeHistoryRequest はQRコード履歴取得リクエスト
type GetQRCodeHistoryRequest struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}

// GetQRCodeHistoryResponse はQRコード履歴取得レスポンス
type GetQRCodeHistoryResponse struct {
	QRCodes []*entities.QRCode
}
