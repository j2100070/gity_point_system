package repository

import (
	"context"

	"github.com/gity/point-system/entities"
	"github.com/google/uuid"
)

// CategoryRepository はカテゴリのリポジトリインターフェース
type CategoryRepository interface {
	// Create は新しいカテゴリを作成
	Create(ctx context.Context, category *entities.Category) error

	// Read はIDでカテゴリを検索
	Read(ctx context.Context, id uuid.UUID) (*entities.Category, error)

	// ReadByCode はコードでカテゴリを検索
	ReadByCode(ctx context.Context, code string) (*entities.Category, error)

	// Update はカテゴリ情報を更新
	Update(ctx context.Context, category *entities.Category) error

	// Delete はカテゴリを論理削除
	Delete(ctx context.Context, id uuid.UUID) error

	// ReadList はカテゴリ一覧を取得（表示順序順）
	ReadList(ctx context.Context, activeOnly bool) ([]*entities.Category, error)

	// Count はカテゴリ総数を取得
	Count(ctx context.Context) (int64, error)

	// ExistsCode はコードが存在するか確認
	ExistsCode(ctx context.Context, code string, excludeID *uuid.UUID) (bool, error)
}
