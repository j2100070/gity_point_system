package dsmysql

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// CategoryDataSource はカテゴリのデータソースインターフェース
type CategoryDataSource interface {
	// Insert は新しいカテゴリを挿入
	Insert(ctx context.Context, category *entities.Category) error

	// Select はIDでカテゴリを検索
	Select(ctx context.Context, id uuid.UUID) (*entities.Category, error)

	// SelectByCode はコードでカテゴリを検索
	SelectByCode(ctx context.Context, code string) (*entities.Category, error)

	// Update はカテゴリ情報を更新
	Update(ctx context.Context, category *entities.Category) error

	// Delete はカテゴリを論理削除
	Delete(ctx context.Context, id uuid.UUID) error

	// SelectList はカテゴリ一覧を取得
	SelectList(ctx context.Context, activeOnly bool) ([]*entities.Category, error)

	// Count はカテゴリ総数を取得
	Count(ctx context.Context) (int64, error)

	// ExistsCode はコードの存在確認
	ExistsCode(ctx context.Context, code string, excludeID *uuid.UUID) (bool, error)
}
