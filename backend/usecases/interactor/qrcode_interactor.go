package interactor

import (
	"errors"
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
)

// QRCodeInteractor はQRコード機能のユースケース実装
type QRCodeInteractor struct {
	qrCodeRepo      repository.QRCodeRepository
	pointTransferUC inputport.PointTransferInputPort
	logger          entities.Logger
}

// NewQRCodeInteractor は新しいQRCodeInteractorを作成
func NewQRCodeInteractor(
	qrCodeRepo repository.QRCodeRepository,
	pointTransferUC inputport.PointTransferInputPort,
	logger entities.Logger,
) inputport.QRCodeInputPort {
	return &QRCodeInteractor{
		qrCodeRepo:      qrCodeRepo,
		pointTransferUC: pointTransferUC,
		logger:          logger,
	}
}

// GenerateReceiveQR は受取用QRコードを生成
func (i *QRCodeInteractor) GenerateReceiveQR(req *inputport.GenerateReceiveQRRequest) (*inputport.GenerateReceiveQRResponse, error) {
	i.logger.Info("Generating receive QR code", entities.NewField("user_id", req.UserID))

	if req.Amount != nil && *req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	qrCode, err := entities.NewReceiveQRCode(req.UserID, req.Amount)
	if err != nil {
		return nil, err
	}

	if err := i.qrCodeRepo.Create(qrCode); err != nil {
		return nil, err
	}

	// QRコードデータの形式: qr_type:code:amount
	qrCodeData := fmt.Sprintf("receive:%s", qrCode.Code)
	if qrCode.Amount != nil {
		qrCodeData = fmt.Sprintf("%s:%d", qrCodeData, *qrCode.Amount)
	}

	return &inputport.GenerateReceiveQRResponse{
		QRCode:     qrCode,
		QRCodeData: qrCodeData,
	}, nil
}

// GenerateSendQR は送信用QRコードを生成
func (i *QRCodeInteractor) GenerateSendQR(req *inputport.GenerateSendQRRequest) (*inputport.GenerateSendQRResponse, error) {
	i.logger.Info("Generating send QR code", entities.NewField("user_id", req.UserID))

	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	qrCode, err := entities.NewSendQRCode(req.UserID, req.Amount)
	if err != nil {
		return nil, err
	}

	if err := i.qrCodeRepo.Create(qrCode); err != nil {
		return nil, err
	}

	qrCodeData := fmt.Sprintf("send:%s:%d", qrCode.Code, req.Amount)

	return &inputport.GenerateSendQRResponse{
		QRCode:     qrCode,
		QRCodeData: qrCodeData,
	}, nil
}

// ScanQR はQRコードをスキャンしてポイント転送
func (i *QRCodeInteractor) ScanQR(req *inputport.ScanQRRequest) (*inputport.ScanQRResponse, error) {
	i.logger.Info("Scanning QR code",
		entities.NewField("user_id", req.UserID),
		entities.NewField("code", req.Code))

	// QRコード取得
	qrCode, err := i.qrCodeRepo.ReadByCode(req.Code)
	if err != nil {
		return nil, errors.New("qr code not found")
	}

	// QRコード検証
	if err := qrCode.CanBeUsedBy(req.UserID); err != nil {
		return nil, err
	}

	// 転送金額の決定
	var amount int64
	if qrCode.Amount != nil {
		amount = *qrCode.Amount
	} else {
		if req.Amount == nil {
			return nil, errors.New("amount is required")
		}
		amount = *req.Amount
	}

	// ポイント転送の方向を決定
	var fromUserID, toUserID = req.UserID, qrCode.UserID
	if qrCode.QRType == entities.QRCodeTypeSend {
		// 送信用QRコードの場合は、QRコード作成者→スキャン者
		fromUserID, toUserID = qrCode.UserID, req.UserID
	}

	// ポイント転送実行
	transferResp, err := i.pointTransferUC.Transfer(&inputport.TransferRequest{
		FromUserID:     fromUserID,
		ToUserID:       toUserID,
		Amount:         amount,
		IdempotencyKey: req.IdempotencyKey,
		Description:    fmt.Sprintf("QR code transfer: %s", qrCode.Code),
	})

	if err != nil {
		return nil, err
	}

	// QRコードを使用済みにする
	if err := qrCode.MarkAsUsed(req.UserID); err != nil {
		i.logger.Warn("Failed to mark QR code as used", entities.NewField("error", err))
	} else {
		if err := i.qrCodeRepo.Update(qrCode); err != nil {
			i.logger.Warn("Failed to update QR code", entities.NewField("error", err))
		}
	}

	return &inputport.ScanQRResponse{
		Transaction: transferResp.Transaction,
		QRCode:      qrCode,
		FromUser:    transferResp.FromUser,
		ToUser:      transferResp.ToUser,
	}, nil
}

// GetQRCodeHistory はQRコード履歴を取得
func (i *QRCodeInteractor) GetQRCodeHistory(req *inputport.GetQRCodeHistoryRequest) (*inputport.GetQRCodeHistoryResponse, error) {
	qrCodes, err := i.qrCodeRepo.ReadListByUserID(req.UserID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	return &inputport.GetQRCodeHistoryResponse{
		QRCodes: qrCodes,
	}, nil
}
