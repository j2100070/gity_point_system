package usecase

import (
	"errors"
	"fmt"

	"github.com/gity/point-system/internal/domain"
	"github.com/google/uuid"
)

// QRCodeUseCase はQRコードに関するユースケース
type QRCodeUseCase struct {
	qrCodeRepo          domain.QRCodeRepository
	pointTransferUC     *PointTransferUseCase
}

// NewQRCodeUseCase は新しいQRCodeUseCaseを作成
func NewQRCodeUseCase(
	qrCodeRepo domain.QRCodeRepository,
	pointTransferUC *PointTransferUseCase,
) *QRCodeUseCase {
	return &QRCodeUseCase{
		qrCodeRepo:      qrCodeRepo,
		pointTransferUC: pointTransferUC,
	}
}

// GenerateReceiveQRRequest は受取用QRコード生成リクエスト
type GenerateReceiveQRRequest struct {
	UserID uuid.UUID
	Amount *int64 // nil=送信者が金額指定、値あり=固定額
}

// GenerateReceiveQRResponse は受取用QRコード生成レスポンス
type GenerateReceiveQRResponse struct {
	QRCode     *domain.QRCode
	QRCodeData string // QRコードに含めるデータ
}

// GenerateReceiveQR は受取用QRコードを生成
// PayPayのような受取用QRコード
func (u *QRCodeUseCase) GenerateReceiveQR(req *GenerateReceiveQRRequest) (*GenerateReceiveQRResponse, error) {
	if req.Amount != nil && *req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	qrCode, err := domain.NewReceiveQRCode(req.UserID, req.Amount)
	if err != nil {
		return nil, err
	}

	if err := u.qrCodeRepo.Create(qrCode); err != nil {
		return nil, err
	}

	// QRコードデータの形式: qr_type:code:amount
	qrCodeData := fmt.Sprintf("receive:%s", qrCode.Code)
	if qrCode.Amount != nil {
		qrCodeData = fmt.Sprintf("%s:%d", qrCodeData, *qrCode.Amount)
	}

	return &GenerateReceiveQRResponse{
		QRCode:     qrCode,
		QRCodeData: qrCodeData,
	}, nil
}

// GenerateSendQRRequest は送信用QRコード生成リクエスト
type GenerateSendQRRequest struct {
	UserID uuid.UUID
	Amount int64
}

// GenerateSendQRResponse は送信用QRコード生成レスポンス
type GenerateSendQRResponse struct {
	QRCode     *domain.QRCode
	QRCodeData string
}

// GenerateSendQR は送信用QRコードを生成
func (u *QRCodeUseCase) GenerateSendQR(req *GenerateSendQRRequest) (*GenerateSendQRResponse, error) {
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	qrCode, err := domain.NewSendQRCode(req.UserID, req.Amount)
	if err != nil {
		return nil, err
	}

	if err := u.qrCodeRepo.Create(qrCode); err != nil {
		return nil, err
	}

	qrCodeData := fmt.Sprintf("send:%s:%d", qrCode.Code, req.Amount)

	return &GenerateSendQRResponse{
		QRCode:     qrCode,
		QRCodeData: qrCodeData,
	}, nil
}

// ScanQRRequest はQRコードスキャンリクエスト
type ScanQRRequest struct {
	ScannerUserID  uuid.UUID
	QRCode         string // QRコードから読み取ったデータ
	Amount         *int64 // 金額（受取用QRで金額未指定の場合）
	IdempotencyKey string // 冪等性キー
}

// ScanQRResponse はQRコードスキャンレスポンス
type ScanQRResponse struct {
	Transaction *domain.Transaction
	QRCode      *domain.QRCode
}

// ScanQR はQRコードをスキャンしてポイント転送
//
// 技術的説明:
// - 受取用QR: スキャンした人 → QRコード所有者にポイント送信
// - 送信用QR: QRコード所有者 → スキャンした人にポイント送信
// - QRコードは一度のみ使用可能（used_atで管理）
// - 期限切れチェックを実施
func (u *QRCodeUseCase) ScanQR(req *ScanQRRequest) (*ScanQRResponse, error) {
	if req.QRCode == "" {
		return nil, errors.New("qr code is required")
	}
	if req.IdempotencyKey == "" {
		return nil, errors.New("idempotency key is required")
	}

	// QRコードデータをパース（簡易実装）
	// 実際のコード部分を抽出
	// フォーマット: "receive:CODE" or "send:CODE:AMOUNT"
	var code string
	var qrType domain.QRCodeType

	// 簡易パース（実際はもっと厳密に）
	if len(req.QRCode) > 8 && req.QRCode[:8] == "receive:" {
		qrType = domain.QRCodeTypeReceive
		code = req.QRCode[8:]
		// ":"で分割して金額があればパース（省略）
		for i, c := range code {
			if c == ':' {
				code = code[:i]
				break
			}
		}
	} else if len(req.QRCode) > 5 && req.QRCode[:5] == "send:" {
		qrType = domain.QRCodeTypeSend
		code = req.QRCode[5:]
		for i, c := range code {
			if c == ':' {
				code = code[:i]
				break
			}
		}
	} else {
		return nil, errors.New("invalid qr code format")
	}

	// QRコード検索
	qrCode, err := u.qrCodeRepo.FindByCode(code)
	if err != nil {
		return nil, errors.New("qr code not found")
	}

	// 使用可能チェック
	if err := qrCode.CanBeUsedBy(req.ScannerUserID); err != nil {
		return nil, err
	}

	// 金額決定
	amount := int64(0)
	if qrCode.Amount != nil {
		amount = *qrCode.Amount
	} else {
		if req.Amount == nil || *req.Amount <= 0 {
			return nil, errors.New("amount is required for this qr code")
		}
		amount = *req.Amount
	}

	// ポイント転送
	var transferReq *TransferRequest
	if qrType == domain.QRCodeTypeReceive {
		// 受取用: スキャナー → QRコード所有者
		transferReq = &TransferRequest{
			FromUserID:     req.ScannerUserID,
			ToUserID:       qrCode.UserID,
			Amount:         amount,
			IdempotencyKey: req.IdempotencyKey,
			Description:    fmt.Sprintf("QR code transfer (receive): %s", qrCode.Code),
		}
	} else {
		// 送信用: QRコード所有者 → スキャナー
		transferReq = &TransferRequest{
			FromUserID:     qrCode.UserID,
			ToUserID:       req.ScannerUserID,
			Amount:         amount,
			IdempotencyKey: req.IdempotencyKey,
			Description:    fmt.Sprintf("QR code transfer (send): %s", qrCode.Code),
		}
	}

	transferResp, err := u.pointTransferUC.Transfer(transferReq)
	if err != nil {
		return nil, err
	}

	// QRコードを使用済みに
	if err := qrCode.MarkAsUsed(req.ScannerUserID); err != nil {
		return nil, err
	}

	if err := u.qrCodeRepo.Update(qrCode); err != nil {
		return nil, err
	}

	return &ScanQRResponse{
		Transaction: transferResp.Transaction,
		QRCode:      qrCode,
	}, nil
}

// GetQRCodeHistoryRequest はQRコード履歴取得リクエスト
type GetQRCodeHistoryRequest struct {
	UserID uuid.UUID
	Offset int
	Limit  int
}

// GetQRCodeHistoryResponse はQRコード履歴取得レスポンス
type GetQRCodeHistoryResponse struct {
	QRCodes []*domain.QRCode
}

// GetQRCodeHistory はQRコード履歴を取得
func (u *QRCodeUseCase) GetQRCodeHistory(req *GetQRCodeHistoryRequest) (*GetQRCodeHistoryResponse, error) {
	qrCodes, err := u.qrCodeRepo.ListByUserID(req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	return &GetQRCodeHistoryResponse{QRCodes: qrCodes}, nil
}
