package transfer_request

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// RepositoryImpl はTransferRequestRepositoryの実装
type RepositoryImpl struct {
	transferRequestDS dsmysql.TransferRequestDataSource
	logger            entities.Logger
}

// NewTransferRequestRepository は新しいTransferRequestRepositoryを作成
func NewTransferRequestRepository(
	transferRequestDS dsmysql.TransferRequestDataSource,
	logger entities.Logger,
) repository.TransferRequestRepository {
	return &RepositoryImpl{
		transferRequestDS: transferRequestDS,
		logger:            logger,
	}
}

// Create は新しい送金リクエストを作成
func (r *RepositoryImpl) Create(ctx context.Context, transferRequest *entities.TransferRequest) error {
	r.logger.Debug("Creating transfer request",
		entities.NewField("from_user_id", transferRequest.FromUserID),
		entities.NewField("to_user_id", transferRequest.ToUserID),
		entities.NewField("amount", transferRequest.Amount))
	return r.transferRequestDS.Insert(ctx, transferRequest)
}

// Read はIDで送金リクエストを検索
func (r *RepositoryImpl) Read(ctx context.Context, id uuid.UUID) (*entities.TransferRequest, error) {
	return r.transferRequestDS.Select(ctx, id)
}

// ReadByIdempotencyKey は冪等性キーで送金リクエストを検索
func (r *RepositoryImpl) ReadByIdempotencyKey(ctx context.Context, key string) (*entities.TransferRequest, error) {
	return r.transferRequestDS.SelectByIdempotencyKey(ctx, key)
}

// Update は送金リクエストを更新
func (r *RepositoryImpl) Update(ctx context.Context, transferRequest *entities.TransferRequest) error {
	r.logger.Debug("Updating transfer request",
		entities.NewField("transfer_request_id", transferRequest.ID),
		entities.NewField("status", transferRequest.Status))
	return r.transferRequestDS.Update(ctx, transferRequest)
}

// ReadPendingByToUser は受取人宛の承認待ちリクエストを取得
func (r *RepositoryImpl) ReadPendingByToUser(ctx context.Context, toUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error) {
	return r.transferRequestDS.SelectPendingByToUser(ctx, toUserID, offset, limit)
}

// ReadSentByFromUser は送信者が送ったリクエストを取得
func (r *RepositoryImpl) ReadSentByFromUser(ctx context.Context, fromUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error) {
	return r.transferRequestDS.SelectSentByFromUser(ctx, fromUserID, offset, limit)
}

// CountPendingByToUser は受取人宛の承認待ちリクエスト数を取得
func (r *RepositoryImpl) CountPendingByToUser(ctx context.Context, toUserID uuid.UUID) (int64, error) {
	return r.transferRequestDS.CountPendingByToUser(ctx, toUserID)
}

// UpdateExpiredRequests は期限切れのリクエストを一括更新
func (r *RepositoryImpl) UpdateExpiredRequests(ctx context.Context) (int64, error) {
	r.logger.Debug("Updating expired transfer requests")
	return r.transferRequestDS.UpdateExpiredRequests(ctx)
}
