package repository

import "context"

// TransactionManager はトランザクション制御の抽象
// UseCase層はこのインターフェースに依存し、具体的な実装（GORM）には依存しない
type TransactionManager interface {
	// Do は関数fnをトランザクション内で実行します。
	// fn内でエラーが返ればRollback、nilならCommitされます。
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
