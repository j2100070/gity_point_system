# TransactionManager パターンによるクリーンアーキテクチャ改善

## 問題点

現在の`DailyBonusInteractor`は以下の問題を抱えています:

```go
import "gorm.io/gorm"  // ← UseCase層がインフラに直接依存

func (i *DailyBonusInteractor) CheckLoginBonus(...) {
    tx := i.db.GetDB().Begin()  // ← GORMのトランザクションを直接使用
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

**問題点:**
1. UseCase層がGORMに直接依存している（`import "gorm.io/gorm"`）
2. クリーンアーキテクチャの依存性逆転の原則に違反
3. 単体テストが困難（実際のデータベースが必要）
4. フレームワークの変更時に大規模な修正が必要

## 解決策: TransactionManager パターン

### 1. TransactionManager インターフェース (UseCase層)

```go
// usecases/repository/transaction_manager.go
package repository

import "context"

// TransactionManager はトランザクション制御の抽象
// UseCase層はこのインターフェースに依存し、具体的な実装（GORM）には依存しない
type TransactionManager interface {
    // Do は関数fnをトランザクション内で実行します。
    // fn内でエラーが返ればRollback、nilならCommitされます。
    Do(ctx context.Context, fn func(ctx context.Context) error) error
}
```

### 2. 具体的な実装 (Infrastructure層)

```go
// gateways/infra/inframysql/transaction_manager.go
package inframysql

import (
    "context"
    "fmt"
    "gorm.io/gorm"
)

type contextKey string
const txKey contextKey = "tx"

type GormTransactionManager struct {
    db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
    return &GormTransactionManager{db: db}
}

func (tm *GormTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
    tx := tm.db.Begin()
    if tx.Error != nil {
        return fmt.Errorf("failed to begin transaction: %w", tx.Error)
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    // トランザクションをcontextに保存
    ctxWithTx := context.WithValue(ctx, txKey, tx)

    if err := fn(ctxWithTx); err != nil {
        if rbErr := tx.Rollback().Error; rbErr != nil {
            return fmt.Errorf("failed to rollback (original error: %v): %w", err, rbErr)
        }
        return err
    }

    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

// GetDB はcontextからトランザクションを取得
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
    if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
        return tx
    }
    return defaultDB
}
```

### 3. リファクタリング後のInteractor

```go
// usecases/interactor/daily_bonus_interactor.go
package interactor

import (
    "context"
    "time"
    // "gorm.io/gorm" ← この依存を削除！
    "github.com/gity/point-system/entities"
    "github.com/gity/point-system/usecases/repository"
)

type DailyBonusInteractor struct {
    dailyBonusRepo  repository.DailyBonusRepository
    userRepo        repository.UserRepository
    transactionRepo repository.TransactionRepository
    txManager       repository.TransactionManager  // ← インターフェースに依存
    logger          entities.Logger
}

func NewDailyBonusInteractor(
    dailyBonusRepo repository.DailyBonusRepository,
    userRepo repository.UserRepository,
    transactionRepo repository.TransactionRepository,
    txManager repository.TransactionManager,  // ← TransactionManagerを受け取る
    logger entities.Logger,
) *DailyBonusInteractor {
    return &DailyBonusInteractor{
        dailyBonusRepo:  dailyBonusRepo,
        userRepo:        userRepo,
        transactionRepo: transactionRepo,
        txManager:       txManager,
        logger:          logger,
    }
}

func (i *DailyBonusInteractor) CheckLoginBonus(req *inputport.CheckLoginBonusRequest) (*inputport.CheckLoginBonusResponse, error) {
    i.logger.Info("Checking login bonus", entities.NewField("user_id", req.UserID))

    var dailyBonus *entities.DailyBonus
    var bonusPoints int64
    var user *entities.User

    // TransactionManagerを使用（GORMに依存しない）
    err := i.txManager.Do(ctx, func(ctx context.Context) error {
        // 本日のボーナスレコードを取得または作成
        dateOnly := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
        db, err := i.dailyBonusRepo.ReadByUserAndDate(ctx, req.UserID, dateOnly)
        if err != nil {
            return err
        }

        if db == nil {
            db = entities.NewDailyBonus(req.UserID, dateOnly)
        }
        dailyBonus = db

        // ログインボーナスを達成
        bonusPoints = dailyBonus.CompleteLogin()

        // ボーナスレコードを保存
        if dailyBonus.CreatedAt.IsZero() {
            if err := i.dailyBonusRepo.Create(ctx, dailyBonus); err != nil {
                return err
            }
        } else {
            if err := i.dailyBonusRepo.Update(ctx, dailyBonus); err != nil {
                return err
            }
        }

        // ポイントを付与
        if bonusPoints > 0 {
            u, err := i.grantBonusPoints(ctx, req.UserID, bonusPoints, "ログインボーナス")
            if err != nil {
                return err
            }
            user = u
        } else {
            u, err := i.userRepo.Read(ctx, req.UserID)
            if err != nil {
                return err
            }
            user = u
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    i.logger.Info("Login bonus checked",
        entities.NewField("user_id", req.UserID),
        entities.NewField("bonus_awarded", bonusPoints))

    return &inputport.CheckLoginBonusResponse{
        DailyBonus:   dailyBonus,
        BonusAwarded: bonusPoints,
        User:         user,
    }, nil
}

func (i *DailyBonusInteractor) grantBonusPoints(ctx context.Context, userID uuid.UUID, points int64, description string) (*entities.User, error) {
    // ユーザーを取得してロック
    user, err := i.userRepo.Read(ctx, userID)
    if err != nil {
        return nil, err
    }

    // 残高を増やす
    user.Balance += points
    user.UpdatedAt = time.Now()

    // ユーザーを更新（contextからトランザクションを取得して使用）
    if _, err := i.userRepo.Update(ctx, user); err != nil {
        return nil, err
    }

    // トランザクションレコードを作成
    transaction := &entities.Transaction{
        ID:              uuid.New(),
        FromUserID:      nil,
        ToUserID:        &userID,
        Amount:          points,
        TransactionType: "daily_bonus",
        Status:          "completed",
        Description:     description,
        CreatedAt:       time.Now(),
        CompletedAt:     timePtr(time.Now()),
    }

    if err := i.transactionRepo.Create(ctx, transaction); err != nil {
        return nil, errors.New("failed to create bonus transaction")
    }

    return user, nil
}
```

### 4. リポジトリの実装例

```go
// gateways/datasource/dsmysqlimpl/daily_bonus_datasource_impl.go
func (ds *DailyBonusDataSource) Insert(ctx context.Context, bonus *entities.DailyBonus) error {
    // contextからトランザクションを取得（存在しなければdefault DB）
    db := inframysql.GetDB(ctx, ds.db.GetDB())

    model := ds.toModel(bonus)
    return db.Create(model).Error
}

func (ds *DailyBonusDataSource) Update(ctx context.Context, bonus *entities.DailyBonus) error {
    db := inframysql.GetDB(ctx, ds.db.GetDB())

    model := ds.toModel(bonus)
    return db.Save(model).Error
}
```

## メリット

### 1. クリーンアーキテクチャの原則に準拠
- UseCase層はインターフェースに依存（`repository.TransactionManager`）
- Infrastructure層の実装（GORM）に直接依存しない
- 依存性逆転の原則を満たす

### 2. テストが容易
```go
// モックTransactionManagerで簡単にテスト可能
type MockTransactionManager struct{}

func (m *MockTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
    return fn(ctx)
}

// テストコード
mockTxManager := &MockTransactionManager{}
interactor := NewDailyBonusInteractor(mockRepo, mockUserRepo, mockTxRepo, mockTxManager, mockLogger)
```

### 3. フレームワーク変更が容易
- GORMから別のORMに変更する場合、TransactionManagerの実装を変えるだけ
- UseCase層のコードは変更不要

### 4. コードの重複削減
- トランザクション開始・コミット・ロールバックのボイラープレートが不要
- 全てのInteractorで同じパターンを使用可能

## 実装手順

1. ✅ TransactionManagerインターフェースを作成（UseCase層）
2. ✅ GormTransactionManagerを実装（Infrastructure層）
3. ✅ リポジトリインターフェースにcontext.Contextを追加
4. ⏳ リポジトリ実装をcontextベースに変更
5. ⏳ Interactorをリファクタリング
6. ⏳ main.goのDI設定を更新
7. ⏳ テストを実行

## 結論

**UseCase層でトランザクションを設定すべきか？**

**答え: はい、ただしインターフェース経由で**

- ✅ トランザクションの**境界**はUseCase層で定義すべき（ビジネスロジックの単位）
- ❌ トランザクションの**実装**はUseCase層に含めるべきではない
- ✅ TransactionManagerパターンを使えば、UseCase層がフレームワークに依存せずトランザクション制御できる

現在の実装は「トランザクション境界の定義」という点では正しいが、「GORMへの直接依存」という点でクリーンアーキテクチャに違反している。

TransactionManagerパターンを導入することで、この問題を解決できる。
