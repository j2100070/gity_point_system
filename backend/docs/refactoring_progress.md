# TransactionManagerパターン リファクタリング進捗

## 完了した作業

### 1. TransactionManager パターンの実装

✅ **TransactionManager インターフェース** (`usecases/repository/transaction_manager.go`)
- UseCase層でトランザクション制御を抽象化
- GORMなどの具体的な実装に依存しない

✅ **GormTransactionManager 実装** (`gateways/infra/inframysql/transaction_manager.go`)
- Infrastructure層での具体的な実装
- contextを使ってトランザクションを伝播
- `GetDB(ctx, defaultDB)` ヘルパー関数でトランザクション取得

### 2. DailyBonusInteractorの完全リファクタリング

✅ **クリーンアーキテクチャ準拠の実装** (`usecases/interactor/daily_bonus_interactor.go`)

**変更前（問題のあるコード）:**
```go
import "gorm.io/gorm"  // ← GORMへの直接依存

func (i *DailyBonusInteractor) CheckLoginBonus(...) {
    tx := i.db.GetDB().Begin()  // ← 直接トランザクション開始
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // ビジネスロジック...

    if err := tx.Commit().Error; err != nil {
        return nil, err
    }
}
```

**変更後（クリーンアーキテクチャ準拠）:**
```go
import "context"  // ← GORMへの依存を削除！

func (i *DailyBonusInteractor) CheckLoginBonus(...) {
    var dailyBonus *entities.DailyBonus
    var bonusPoints int64
    var user *entities.User

    // TransactionManagerを使用（インターフェースに依存）
    err := i.txManager.Do(ctx, func(ctx context.Context) error {
        // contextを使ってトランザクション内で処理
        db, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, req.UserID, dateOnly)
        // ... ビジネスロジック
        return nil
    })

    return &inputport.CheckLoginBonusResponse{...}, nil
}
```

**改善点:**
1. ✅ `import "gorm.io/gorm"`を削除 - フレームワーク依存を除去
2. ✅ `txManager.Do()` - インターフェース経由でトランザクション制御
3. ✅ `context.Context` - トランザクションの伝播
4. ✅ ボイラープレート削除 - Begin/Commit/Rollbackのロジックが不要

### 3. リポジトリ層の更新

✅ **context対応リポジトリ** (DailyBonus, 一部のUserとTransaction)
- `ReadByUserAndDate(ctx context.Context, ...)` - contextを第一引数に
- `Create(ctx context.Context, ...)` - contextベースの操作
- `Update(ctx context.Context, ...)` - トランザクション対応

✅ **DataSource層の更新** (DailyBonus, 一部のUser)
- `inframysql.GetDB(ctx, defaultDB)` - contextからトランザクション取得
- トランザクションが存在しない場合はdefaultDBを使用

### 4. DI設定の更新

✅ **main.goの変更** (`cmd/clean_server/main.go`)
```go
// TransactionManagerを作成
txManager := inframysql.NewGormTransactionManager(db.GetDB())

dailyBonusUC := interactor.NewDailyBonusInteractor(
    dailyBonusRepo,
    userRepo,
    transactionRepo,
    txManager,  // ← インターフェースを注入
    logger,
)
```

## メリット

### 1. クリーンアーキテクチャの原則に準拠
- ✅ UseCase層はインターフェースに依存（`repository.TransactionManager`）
- ✅ Infrastructure層の実装（GORM）に直接依存しない
- ✅ 依存性逆転の原則（DIP）を満たす

### 2. テストが容易
```go
// モックTransactionManagerで簡単にテスト可能
type MockTransactionManager struct{}

func (m *MockTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
    return fn(ctx)
}
```

### 3. フレームワーク変更が容易
- GORMから別のORMに変更する場合、TransactionManagerの実装を変えるだけ
- UseCase層のコードは変更不要

### 4. コードの重複削減
- トランザクション開始・コミット・ロールバックのボイラープレートが不要
- 全てのInteractorで同じパターンを使用可能

## 未完了の作業（段階的移行）

以下のInteractorとRepositoryは、後で段階的に移行する予定です:

### 残りのInteractor
- ⏳ `AdminInteractor` - 管理者機能
- ⏳ `PointTransferInteractor` - ポイント送金
- ⏳ `ProductExchangeInteractor` - 商品交換
- ⏳ `FriendInteractor` - フレンド機能
- ⏳ その他のInteractor

### 残りのRepository/DataSource
- ⏳ UserRepository - 一部メソッドのみ対応済み（Read, Update）
- ⏳ TransactionRepository - Createのみ対応済み
- ⏳ その他のRepository

## 移行戦略

1. **DailyBonusInteractorを参考実装として使用**
   - 完全にクリーンアーキテクチャに準拠
   - 他のInteractorの移行時のテンプレートとして機能

2. **段階的な移行**
   - 各Interactorを個別に移行
   - テストを実行して動作確認
   - 一度に全てを変更せず、リスクを最小化

3. **後方互換性の維持**
   - 古いインターフェースを段階的に非推奨化
   - レガシーコードとの共存期間を設ける

## テスト方法

1. **DailyBonusの機能テスト**
   ```bash
   # ログインボーナスの取得
   curl -X POST http://localhost:8080/api/daily-bonus/claim-login \
     -H "Authorization: Bearer <token>"

   # 本日のボーナス状況確認
   curl http://localhost:8080/api/daily-bonus/today \
     -H "Authorization: Bearer <token>"
   ```

2. **トランザクションの動作確認**
   - ボーナスポイントが正しく付与されることを確認
   - エラー時にロールバックされることを確認
   - データベースの整合性を確認

## 結論

**DailyBonusInteractorは、GoのクリーンアーキテクチャにおけるTransactionManagerパターンの成功例です。**

- ✅ UseCase層でトランザクションの**境界**を定義（ビジネスロジックの単位）
- ✅ TransactionManagerパターンで**実装**を抽象化（フレームワークに依存しない）
- ✅ クリーンアーキテクチャの依存性逆転の原則に完全準拠

この実装は、他のInteractorの移行時の参考として機能し、段階的なリファクタリングを可能にします。
