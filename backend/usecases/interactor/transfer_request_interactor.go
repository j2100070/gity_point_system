package interactor

import (
	"context"
	"errors"
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
)

// TransferRequestInteractor は送金リクエスト機能のユースケース実装
type TransferRequestInteractor struct {
	transferRequestRepo repository.TransferRequestRepository
	userRepo            repository.UserRepository
	pointTransferPort   inputport.PointTransferInputPort
	logger              entities.Logger
}

// NewTransferRequestInteractor は新しいTransferRequestInteractorを作成
func NewTransferRequestInteractor(
	transferRequestRepo repository.TransferRequestRepository,
	userRepo repository.UserRepository,
	pointTransferPort inputport.PointTransferInputPort,
	logger entities.Logger,
) inputport.TransferRequestInputPort {
	return &TransferRequestInteractor{
		transferRequestRepo: transferRequestRepo,
		userRepo:            userRepo,
		pointTransferPort:   pointTransferPort,
		logger:              logger,
	}
}

// CreateTransferRequest は送金リクエストを作成（QRスキャン時）
func (i *TransferRequestInteractor) CreateTransferRequest(ctx context.Context, req *inputport.CreateTransferRequestRequest) (*inputport.CreateTransferRequestResponse, error) {
	i.logger.Info("Creating transfer request",
		entities.NewField("from_user_id", req.FromUserID),
		entities.NewField("to_user_id", req.ToUserID),
		entities.NewField("amount", req.Amount))

	// 冪等性チェック
	existing, err := i.transferRequestRepo.ReadByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check idempotency key: %w", err)
	}
	if existing != nil {
		// 既存のリクエストが見つかった場合
		fromUser, _ := i.userRepo.Read(ctx, req.FromUserID)
		toUser, _ := i.userRepo.Read(ctx, req.ToUserID)
		return &inputport.CreateTransferRequestResponse{
			TransferRequest: existing,
			FromUser:        fromUser,
			ToUser:          toUser,
		}, nil
	}

	// 送信者と受取人の存在確認
	fromUser, err := i.userRepo.Read(ctx, req.FromUserID)
	if err != nil {
		return nil, errors.New("sender not found")
	}
	if !fromUser.IsActive {
		return nil, errors.New("sender is not active")
	}

	toUser, err := i.userRepo.Read(ctx, req.ToUserID)
	if err != nil {
		return nil, errors.New("receiver not found")
	}
	if !toUser.IsActive {
		return nil, errors.New("receiver is not active")
	}

	// 残高チェック
	if err := fromUser.CanTransfer(req.Amount); err != nil {
		return nil, fmt.Errorf("transfer validation failed: %w", err)
	}

	// 送金リクエストエンティティを作成
	transferRequest, err := entities.NewTransferRequest(
		req.FromUserID,
		req.ToUserID,
		req.Amount,
		req.Message,
		req.IdempotencyKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer request entity: %w", err)
	}

	// DB保存
	if err := i.transferRequestRepo.Create(ctx, transferRequest); err != nil {
		return nil, fmt.Errorf("failed to save transfer request: %w", err)
	}

	i.logger.Info("Transfer request created successfully",
		entities.NewField("request_id", transferRequest.ID))

	return &inputport.CreateTransferRequestResponse{
		TransferRequest: transferRequest,
		FromUser:        fromUser,
		ToUser:          toUser,
	}, nil
}

// ApproveTransferRequest は送金リクエストを承認（受取人が承認）
func (i *TransferRequestInteractor) ApproveTransferRequest(ctx context.Context, req *inputport.ApproveTransferRequestRequest) (*inputport.ApproveTransferRequestResponse, error) {
	i.logger.Info("Approving transfer request",
		entities.NewField("request_id", req.RequestID),
		entities.NewField("user_id", req.UserID))

	// リクエストの取得
	transferRequest, err := i.transferRequestRepo.Read(ctx, req.RequestID)
	if err != nil {
		return nil, errors.New("transfer request not found")
	}
	if transferRequest == nil {
		return nil, errors.New("transfer request not found")
	}

	// 承認者が受取人であることを確認
	if transferRequest.ToUserID != req.UserID {
		return nil, errors.New("unauthorized to approve this request")
	}

	// 承認可能かチェック
	if err := transferRequest.CanApprove(); err != nil {
		return nil, fmt.Errorf("cannot approve request: %w", err)
	}

	// ポイント送金を実行（ポイント転送機能内でトランザクションが管理される）
	transferResp, err := i.pointTransferPort.Transfer(ctx, &inputport.TransferRequest{
		FromUserID:     transferRequest.FromUserID,
		ToUserID:       transferRequest.ToUserID,
		Amount:         transferRequest.Amount,
		IdempotencyKey: fmt.Sprintf("transfer-request-%s", transferRequest.ID.String()),
		Description:    fmt.Sprintf("送金リクエスト承認: %s", transferRequest.Message),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute transfer: %w", err)
	}

	transaction := transferResp.Transaction
	fromUser := transferResp.FromUser
	toUser := transferResp.ToUser

	// リクエストを承認済みにマーク
	if err := transferRequest.Approve(transaction.ID); err != nil {
		return nil, fmt.Errorf("failed to approve transfer request: %w", err)
	}

	// DB更新
	if err := i.transferRequestRepo.Update(ctx, transferRequest); err != nil {
		return nil, fmt.Errorf("failed to update transfer request: %w", err)
	}

	i.logger.Info("Transfer request approved successfully",
		entities.NewField("request_id", transferRequest.ID),
		entities.NewField("transaction_id", transaction.ID))

	return &inputport.ApproveTransferRequestResponse{
		TransferRequest: transferRequest,
		Transaction:     transaction,
		FromUser:        fromUser,
		ToUser:          toUser,
	}, nil
}

// RejectTransferRequest は送金リクエストを拒否（受取人が拒否）
func (i *TransferRequestInteractor) RejectTransferRequest(ctx context.Context, req *inputport.RejectTransferRequestRequest) (*inputport.RejectTransferRequestResponse, error) {
	i.logger.Info("Rejecting transfer request",
		entities.NewField("request_id", req.RequestID),
		entities.NewField("user_id", req.UserID))

	// リクエストの取得
	transferRequest, err := i.transferRequestRepo.Read(ctx, req.RequestID)
	if err != nil {
		return nil, errors.New("transfer request not found")
	}
	if transferRequest == nil {
		return nil, errors.New("transfer request not found")
	}

	// 拒否者が受取人であることを確認
	if transferRequest.ToUserID != req.UserID {
		return nil, errors.New("unauthorized to reject this request")
	}

	// 拒否可能かチェック
	if err := transferRequest.CanReject(); err != nil {
		return nil, fmt.Errorf("cannot reject request: %w", err)
	}

	// リクエストを拒否
	if err := transferRequest.Reject(); err != nil {
		return nil, fmt.Errorf("failed to reject transfer request: %w", err)
	}

	// DB更新
	if err := i.transferRequestRepo.Update(ctx, transferRequest); err != nil {
		return nil, fmt.Errorf("failed to update transfer request: %w", err)
	}

	i.logger.Info("Transfer request rejected successfully",
		entities.NewField("request_id", transferRequest.ID))

	return &inputport.RejectTransferRequestResponse{
		TransferRequest: transferRequest,
	}, nil
}

// CancelTransferRequest は送金リクエストをキャンセル（送信者がキャンセル）
func (i *TransferRequestInteractor) CancelTransferRequest(ctx context.Context, req *inputport.CancelTransferRequestRequest) (*inputport.CancelTransferRequestResponse, error) {
	i.logger.Info("Canceling transfer request",
		entities.NewField("request_id", req.RequestID),
		entities.NewField("user_id", req.UserID))

	// リクエストの取得
	transferRequest, err := i.transferRequestRepo.Read(ctx, req.RequestID)
	if err != nil {
		return nil, errors.New("transfer request not found")
	}
	if transferRequest == nil {
		return nil, errors.New("transfer request not found")
	}

	// キャンセル者が送信者であることを確認
	if transferRequest.FromUserID != req.UserID {
		return nil, errors.New("unauthorized to cancel this request")
	}

	// キャンセル可能かチェック
	if err := transferRequest.CanCancel(); err != nil {
		return nil, fmt.Errorf("cannot cancel request: %w", err)
	}

	// リクエストをキャンセル
	if err := transferRequest.Cancel(); err != nil {
		return nil, fmt.Errorf("failed to cancel transfer request: %w", err)
	}

	// DB更新
	if err := i.transferRequestRepo.Update(ctx, transferRequest); err != nil {
		return nil, fmt.Errorf("failed to update transfer request: %w", err)
	}

	i.logger.Info("Transfer request cancelled successfully",
		entities.NewField("request_id", transferRequest.ID))

	return &inputport.CancelTransferRequestResponse{
		TransferRequest: transferRequest,
	}, nil
}

// GetPendingRequests は受取人宛の承認待ちリクエスト一覧を取得
func (i *TransferRequestInteractor) GetPendingRequests(ctx context.Context, req *inputport.GetPendingTransferRequestsRequest) (*inputport.GetPendingTransferRequestsResponse, error) {
	requests, err := i.transferRequestRepo.ReadPendingByToUser(ctx, req.ToUserID, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}

	infos := make([]*inputport.TransferRequestInfo, 0, len(requests))
	for _, tr := range requests {
		// 期限切れチェック
		if tr.IsExpired() {
			tr.MarkAsExpired()
			i.transferRequestRepo.Update(ctx, tr)
			continue // 期限切れは除外
		}

		fromUser, err := i.userRepo.Read(ctx, tr.FromUserID)
		if err != nil {
			continue
		}

		toUser, err := i.userRepo.Read(ctx, tr.ToUserID)
		if err != nil {
			continue
		}

		infos = append(infos, &inputport.TransferRequestInfo{
			TransferRequest: tr,
			FromUser:        fromUser,
			ToUser:          toUser,
		})
	}

	return &inputport.GetPendingTransferRequestsResponse{
		Requests: infos,
	}, nil
}

// GetSentRequests は送信者が送った送金リクエスト一覧を取得
func (i *TransferRequestInteractor) GetSentRequests(ctx context.Context, req *inputport.GetSentTransferRequestsRequest) (*inputport.GetSentTransferRequestsResponse, error) {
	requests, err := i.transferRequestRepo.ReadSentByFromUser(ctx, req.FromUserID, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent requests: %w", err)
	}

	infos := make([]*inputport.TransferRequestInfo, 0, len(requests))
	for _, tr := range requests {
		// 期限切れチェック
		if tr.IsPending() && tr.IsExpired() {
			tr.MarkAsExpired()
			i.transferRequestRepo.Update(ctx, tr)
		}

		fromUser, err := i.userRepo.Read(ctx, tr.FromUserID)
		if err != nil {
			continue
		}

		toUser, err := i.userRepo.Read(ctx, tr.ToUserID)
		if err != nil {
			continue
		}

		infos = append(infos, &inputport.TransferRequestInfo{
			TransferRequest: tr,
			FromUser:        fromUser,
			ToUser:          toUser,
		})
	}

	return &inputport.GetSentTransferRequestsResponse{
		Requests: infos,
	}, nil
}

// GetRequestDetail は送金リクエスト詳細を取得
func (i *TransferRequestInteractor) GetRequestDetail(ctx context.Context, req *inputport.GetTransferRequestDetailRequest) (*inputport.GetTransferRequestDetailResponse, error) {
	transferRequest, err := i.transferRequestRepo.Read(ctx, req.RequestID)
	if err != nil {
		return nil, errors.New("transfer request not found")
	}
	if transferRequest == nil {
		return nil, errors.New("transfer request not found")
	}

	// アクセス権限チェック（送信者または受取人のみ閲覧可能）
	if transferRequest.FromUserID != req.UserID && transferRequest.ToUserID != req.UserID {
		return nil, errors.New("unauthorized to view this request")
	}

	fromUser, err := i.userRepo.Read(ctx, transferRequest.FromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender: %w", err)
	}

	toUser, err := i.userRepo.Read(ctx, transferRequest.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get receiver: %w", err)
	}

	return &inputport.GetTransferRequestDetailResponse{
		TransferRequest: transferRequest,
		FromUser:        fromUser,
		ToUser:          toUser,
	}, nil
}

// GetPendingRequestCount は受取人宛の承認待ちリクエスト数を取得
func (i *TransferRequestInteractor) GetPendingRequestCount(ctx context.Context, req *inputport.GetPendingRequestCountRequest) (*inputport.GetPendingRequestCountResponse, error) {
	count, err := i.transferRequestRepo.CountPendingByToUser(ctx, req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending requests: %w", err)
	}

	return &inputport.GetPendingRequestCountResponse{
		Count: count,
	}, nil
}
