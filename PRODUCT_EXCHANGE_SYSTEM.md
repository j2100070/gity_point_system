# 商品交換システム実装ガイド

## 概要

ポイントを使ってお菓子・ジュースなどの商品と交換できるシステムを実装しました。
管理者が商品を登録・管理し、ユーザーがポイントで商品を購入できます。

## 📋 実装済みファイル

### 1. データベース
- **`migrations/002_add_product_exchange.sql`** - 商品テーブルと交換履歴テーブル
  - `products` - 商品マスタ
  - `product_exchanges` - 商品交換履歴

### 2. エンティティ
- **`entities/product.go`** - 商品エンティティと交換エンティティ
  - `Product` - 商品情報
  - `ProductExchange` - 交換記録

### 3. リポジトリ
- **`usecases/repository/product_repository.go`** - リポジトリインターフェース
  - `ProductRepository` - 商品操作
  - `ProductExchangeRepository` - 交換履歴操作

### 4. ユースケース
- **`usecases/inputport/product_input_port.go`** - 入力ポート定義
  - `ProductManagementInputPort` - 商品管理（管理者用）
  - `ProductExchangeInputPort` - 商品交換（ユーザー用）

- **`usecases/interactor/product_exchange_interactor.go`** - ビジネスロジック実装
  - 商品交換処理
  - 在庫管理
  - ポイント減算
  - トランザクション管理

## 🗄️ データベース構造

### products テーブル

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,              -- 商品名
    description TEXT,                         -- 説明
    category VARCHAR(100) NOT NULL,           -- カテゴリ (snack/drink/other)
    price BIGINT NOT NULL CHECK (price > 0), -- 必要ポイント数
    stock INTEGER NOT NULL,                   -- 在庫数 (-1 = 無制限)
    image_url TEXT,                          -- 商品画像URL
    is_available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### product_exchanges テーブル

```sql
CREATE TABLE product_exchanges (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity INTEGER NOT NULL,                -- 購入数量
    points_used BIGINT NOT NULL,              -- 使用ポイント
    status VARCHAR(50) NOT NULL,              -- pending/completed/cancelled/delivered
    transaction_id UUID,                      -- 関連トランザクション
    notes TEXT,                               -- 備考（受取場所など）
    created_at TIMESTAMP,
    completed_at TIMESTAMP,
    delivered_at TIMESTAMP
);
```

## 🎯 主な機能

### 管理者機能

1. **商品登録**
   ```go
   CreateProduct(req *CreateProductRequest) (*CreateProductResponse, error)
   ```
   - 商品名、カテゴリ、価格、在庫を登録

2. **商品更新**
   ```go
   UpdateProduct(req *UpdateProductRequest) (*UpdateProductResponse, error)
   ```
   - 商品情報の変更
   - 在庫数の調整
   - 交換可否の切り替え

3. **商品削除**
   ```go
   DeleteProduct(req *DeleteProductRequest) error
   ```
   - 論理削除（データは残る）

4. **配達完了**
   ```go
   MarkExchangeDelivered(req *MarkExchangeDeliveredRequest) error
   ```
   - 商品を配達済みにする

### ユーザー機能

1. **商品一覧表示**
   ```go
   GetProductList(req *GetProductListRequest) (*GetProductListResponse, error)
   ```
   - カテゴリ別フィルタリング
   - 交換可能商品のみ表示

2. **商品交換**
   ```go
   ExchangeProduct(req *ExchangeProductRequest) (*ExchangeProductResponse, error)
   ```
   - ポイント残高チェック
   - 在庫チェック
   - ポイント減算
   - 在庫減算
   - 交換記録作成

3. **交換履歴**
   ```go
   GetExchangeHistory(req *GetExchangeHistoryRequest) (*GetExchangeHistoryResponse, error)
   ```
   - 自分の交換履歴を表示

4. **交換キャンセル**
   ```go
   CancelExchange(req *CancelExchangeRequest) error
   ```
   - pending状態の交換をキャンセル
   - ポイント返却
   - 在庫復元

## 🔒 セキュリティ対策

### 1. トランザクション管理

商品交換は以下を原子的に実行：

```go
err := i.db.Transaction(func(tx *gorm.DB) error {
    // 1. 商品情報取得
    // 2. 在庫チェック
    // 3. 残高チェック
    // 4. 在庫減算
    // 5. ポイント減算（ロック付き）
    // 6. トランザクション記録作成
    // 7. 交換記録作成
    return nil
})
```

### 2. 悲観的ロック

ユーザー残高の更新時にロックを取得：

```go
updates := []repository.BalanceUpdate{
    {UserID: req.UserID, Amount: totalPoints, IsDeduct: true},
}
i.userRepo.UpdateBalancesWithLock(tx, updates)
```

### 3. バリデーション

- 残高不足チェック
- 在庫不足チェック
- 商品交換可否チェック
- アカウント有効性チェック

### 4. 権限チェック

- キャンセルは本人のみ可能
- 配達完了は管理者のみ可能

## 📊 交換フロー

```
┌─────────────┐
│ユーザー     │
│商品を選択   │
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│1. 商品情報取得      │
│   - 価格確認        │
│   - 在庫確認        │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│2. 残高チェック      │
│   - ポイント十分?   │
└──────┬──────────────┘
       │ YES
       ▼
┌─────────────────────┐
│3. 在庫減算          │
│   products.stock -= quantity
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│4. ポイント減算      │
│   users.balance -= price * quantity
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│5. トランザクション  │
│   記録作成          │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│6. 交換記録作成      │
│   status = completed │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│完了！               │
│受取通知をユーザーへ │
└─────────────────────┘
```

## 🚀 次のステップ（実装が必要）

### 1. リポジトリ実装
```
backend/gateways/repository/product/
  ├── product_repository_impl.go
  └── product_exchange_repository_impl.go
```

### 2. DataSource実装
```
backend/gateways/datasource/dsmysqlimpl/
  ├── product_datasource_impl.go
  └── product_exchange_datasource_impl.go
```

### 3. コントローラー実装
```
backend/controllers/
  ├── product_controller.go      (管理者用)
  └── product_exchange_controller.go (ユーザー用)
```

### 4. ルーティング追加
```go
// 管理者用
admin.POST("/products", productController.CreateProduct)
admin.PUT("/products/:id", productController.UpdateProduct)
admin.DELETE("/products/:id", productController.DeleteProduct)
admin.POST("/exchanges/:id/deliver", productController.MarkDelivered)

// ユーザー用
api.GET("/products", productController.GetProductList)
api.POST("/exchanges", exchangeController.ExchangeProduct)
api.GET("/exchanges/history", exchangeController.GetHistory)
api.POST("/exchanges/:id/cancel", exchangeController.CancelExchange)
```

### 5. フロントエンド実装
```
frontend/src/features/products/
  ├── ProductList.tsx          (商品一覧)
  ├── ProductDetail.tsx        (商品詳細)
  ├── ExchangeHistory.tsx      (交換履歴)
  └── AdminProductManage.tsx   (管理者用)
```

## 💡 使用例

### 商品交換

```go
// ユーザーがコカ・コーラ2本を交換
req := &inputport.ExchangeProductRequest{
    UserID:    userID,
    ProductID: cokeID,
    Quantity:  2,
    Notes:     "受取場所: 事務室、時間: 14:00",
}

resp, err := interactor.ExchangeProduct(req)
// => ポイント200消費、コカ・コーラ2本の交換完了
```

### 交換キャンセル

```go
// まだ配達されていない交換をキャンセル
req := &inputport.CancelExchangeRequest{
    UserID:     userID,
    ExchangeID: exchangeID,
}

err := interactor.CancelExchange(req)
// => ポイント返却、在庫復元
```

## 📝 サンプルデータ

マイグレーション実行時に以下のサンプル商品が登録されます：

| 商品名 | カテゴリ | 価格（ポイント） | 在庫 |
|--------|---------|-----------------|------|
| コカ・コーラ 500ml | drink | 100 | 50 |
| ポカリスエット 500ml | drink | 120 | 30 |
| お〜いお茶 500ml | drink | 90 | 40 |
| ポテトチップス | snack | 150 | 100 |
| じゃがりこ | snack | 140 | 80 |
| キットカット | snack | 120 | 60 |
| ブラックサンダー | snack | 80 | 120 |
| うまい棒 | snack | 20 | 200 |
| ハーゲンダッツ | snack | 300 | 20 |

## 🎨 UI設計案

### 商品一覧画面
```
┌────────────────────────────────────┐
│ 🛍️ 商品一覧                       │
├────────────────────────────────────┤
│ カテゴリ: [全て▼] [飲み物] [お菓子] │
├────────────────────────────────────┤
│ ┌──────┬──────────────────┬─────┐ │
│ │ 🥤   │ コカ・コーラ      │100pt│ │
│ │      │ 定番の炭酸飲料    │     │ │
│ │      │ 在庫: 50個        │[交換]│ │
│ └──────┴──────────────────┴─────┘ │
│ ┌──────┬──────────────────┬─────┐ │
│ │ 🍫   │ キットカット      │120pt│ │
│ │      │ チョコレート菓子  │     │ │
│ │      │ 在庫: 60個        │[交換]│ │
│ └──────┴──────────────────┴─────┘ │
└────────────────────────────────────┘
```

### 交換履歴画面
```
┌────────────────────────────────────┐
│ 📋 交換履歴                        │
├────────────────────────────────────┤
│ 2026/02/07 14:30                   │
│ コカ・コーラ x2    -200pt          │
│ ステータス: 配達済み ✅            │
├────────────────────────────────────┤
│ 2026/02/06 10:15                   │
│ キットカット x1    -120pt          │
│ ステータス: 完了 📦  [キャンセル]  │
└────────────────────────────────────┘
```

## 🔧 マイグレーション実行

```bash
# Dockerコンテナ再起動（マイグレーション自動実行）
docker compose down
docker compose up -d

# または手動実行
docker exec -i point_system_db psql -U postgres -d point_system < backend/migrations/002_add_product_exchange.sql
```

---

これでポイント交換システムの基盤が完成しました！
次はリポジトリとコントローラーの実装を進めてください。
