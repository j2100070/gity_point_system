package dsmysql

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// TransferRequestDataSource はMySQLの送金リクエストデータソースインターフェース
type TransferRequestDataSource interface {
	// Insert は新しい送金リクエストを挿入
	Insert(ctx context.Context, transferRequest *entities.TransferRequest) error

	// Select はIDで送金リクエストを検索
	Select(ctx context.Context, id uuid.UUID) (*entities.TransferRequest, error)

	// SelectByIdempotencyKey は冪等性キーで送金リクエストを検索
	SelectByIdempotencyKey(ctx context.Context, key string) (*entities.TransferRequest, error)

	// Update は送金リクエストを更新
	Update(ctx context.Context, transferRequest *entities.TransferRequest) error

	// SelectPendingByToUser は受取人宛の承認待ちリクエストを取得
	SelectPendingByToUser(ctx context.Context, toUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error)

	// SelectSentByFromUser は送信者が送ったリクエストを取得
	SelectSentByFromUser(ctx context.Context, fromUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error)

	// CountPendingByToUser は受取人宛の承認待ちリクエスト数を取得
	CountPendingByToUser(ctx context.Context, toUserID uuid.UUID) (int64, error)

	// UpdateExpiredRequests は期限切れのリクエストを一括更新
	UpdateExpiredRequests(ctx context.Context) (int64, error)
}
