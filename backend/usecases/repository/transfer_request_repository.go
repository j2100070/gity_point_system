package repository

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// TransferRequestRepository は送金リクエストのリポジトリインターフェース
type TransferRequestRepository interface {
	// Create は新しい送金リクエストを作成
	Create(ctx context.Context, transferRequest *entities.TransferRequest) error

	// Read はIDで送金リクエストを検索
	Read(ctx context.Context, id uuid.UUID) (*entities.TransferRequest, error)

	// ReadByIdempotencyKey は冪等性キーで送金リクエストを検索
	ReadByIdempotencyKey(ctx context.Context, key string) (*entities.TransferRequest, error)

	// Update は送金リクエストを更新
	Update(ctx context.Context, transferRequest *entities.TransferRequest) error

	// ReadPendingByToUser は受取人宛の承認待ちリクエストを取得
	ReadPendingByToUser(ctx context.Context, toUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error)

	// ReadSentByFromUser は送信者が送ったリクエストを取得
	ReadSentByFromUser(ctx context.Context, fromUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error)

	// CountPendingByToUser は受取人宛の承認待ちリクエスト数を取得
	CountPendingByToUser(ctx context.Context, toUserID uuid.UUID) (int64, error)

	// UpdateExpiredRequests は期限切れのリクエストを一括更新
	UpdateExpiredRequests(ctx context.Context) (int64, error)

	// ReadPendingByToUserWithUsers は受取人宛の承認待ちリクエストをユーザー情報付きで取得（JOIN）
	ReadPendingByToUserWithUsers(ctx context.Context, toUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequestWithUsers, error)

	// ReadSentByFromUserWithUsers は送信者が送ったリクエストをユーザー情報付きで取得（JOIN）
	ReadSentByFromUserWithUsers(ctx context.Context, fromUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequestWithUsers, error)
}
