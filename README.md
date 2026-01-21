# 現在製作中

# Gity Point System

PayPayのようなQRコードベースのポイント送受信システム

## 🎯 概要

React + Goで構築されたプロダクションレベルのポイント管理システムです。QRコードを使用したポイントのやり取り、友達機能、管理者機能を備えています。

## ✨ 主な機能

### ユーザー機能
- **ポイント転送**: QRコードまたは直接送金
- **QRコード**: 受取用・送信用QRコード生成
- **友達機能**: 友達申請、承認、一覧表示
- **取引履歴**: ポイント移動の完全な履歴

### 管理者機能
- **ポイント管理**: ユーザーへのポイント付与・減算
- **ユーザー管理**: 役割変更、アカウント無効化
- **監査**: 全トランザクションの閲覧


###  トランザクション保護
- **冪等性**: Idempotency Keyで重複送金を防止
- **楽観的ロック**: Version列で競合検知
- **悲観的ロック**: SELECT FOR UPDATEで残高整合性を保証
- **ACID特性**: PostgreSQLのトランザクションで原子性を保証

## 🏗️ アーキテクチャ

### バックエンド (Go)
```
backend/
├── cmd/server/           # エントリーポイント
├── internal/
│   ├── domain/          # ドメインモデル・ビジネスロジック
│   ├── usecase/         # ユースケース層
│   ├── interface/       # HTTPハンドラー・ミドルウェア
│   └── infrastructure/  # DB実装（GORM）
├── migrations/          # DBマイグレーション
└── config/             # 設定管理
```

**クリーンアーキテクチャ**を採用:
- 依存関係の方向: interface → usecase → domain
- テスタビリティ重視
- ドメイン層は外部に依存しない

### フロントエンド (React)
```
frontend/
├── src/
│   ├── features/        # Feature-based構造
│   │   ├── auth/       # 認証機能
│   │   ├── points/     # ポイント機能
│   │   ├── friends/    # 友達機能
│   │   └── admin/      # 管理者機能
│   ├── shared/         # 共通コンポーネント
│   └── core/           # 基盤（API、セキュリティ）
```

**Feature-based + Clean Architecture**:
- 機能ごとにディレクトリを分割
- 各機能内でレイヤー分離 (api/components/hooks/types)

## 📊 データベース設計

### 主要テーブル
```sql
-- ユーザー（楽観的ロック、ソフトデリート対応）
users (id, username, email, password_hash, balance, version, role, ...)

-- セッション（Session-based認証）
sessions (id, user_id, session_token, csrf_token, expires_at, ...)

-- トランザクション（ポイント移動履歴）
transactions (id, from_user_id, to_user_id, amount, idempotency_key, ...)

-- 冪等性キー（重複トランザクション防止）
idempotency_keys (key, user_id, transaction_id, status, expires_at)

-- 友達関係
friendships (id, requester_id, addressee_id, status, ...)

-- QRコード（一時的な受取・送信用）
qr_codes (id, user_id, code, amount, qr_type, expires_at, ...)
```

### 重要な設計ポイント
1. **残高制約**: `CHECK (balance >= 0)` で負の値を防止
2. **楽観的ロック**: `version`列で更新時の競合を検知
3. **冪等性**: `idempotency_keys`で同一キーの重複処理を防止
4. **監査ログ**: 管理者操作の完全な記録

## 🚀 セットアップ

### 前提条件
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+

### 起動方法

```bash
# 1. リポジトリのクローン
git clone <repository-url>
cd gity_point_system

# 2. 環境変数の設定
cp .env.example .env

# 3. Docker Composeで起動
docker-compose up -d

# 4. データベースマイグレーション実行
# PostgreSQLに接続してmigrationsを実行
psql -h localhost -U postgres -d point_system -f backend/migrations/001_initial_schema.sql
```

### 初期アカウント
```
管理者:
  Username: admin
  Password: Admin@123456
  Balance: 1,000,000 points

テストユーザー1:
  Username: user1
  Password: User@123456
  Balance: 10,000 points

テストユーザー2:
  Username: user2
  Password: User@123456
  Balance: 5,000 points
```

## 🔧 開発

### バックエンド開発
```bash
cd backend
go mod download
go run cmd/server/main.go
```

### フロントエンド開発
```bash
cd frontend
npm install
npm start
```

## 📝 API エンドポイント

### 認証
- `POST /api/auth/register` - ユーザー登録
- `POST /api/auth/login` - ログイン
- `POST /api/auth/logout` - ログアウト
- `GET /api/auth/me` - 現在のユーザー情報

### ポイント（要認証）
- `POST /api/points/transfer` - ポイント転送
- `GET /api/points/balance` - 残高取得
- `GET /api/points/history` - 取引履歴

### QRコード（要認証）
- `POST /api/qr/generate/receive` - 受取用QR生成
- `POST /api/qr/generate/send` - 送信用QR生成
- `POST /api/qr/scan` - QRスキャン
- `GET /api/qr/history` - QR履歴

### 友達（要認証）
- `POST /api/friends/request` - 友達申請
- `POST /api/friends/accept` - 申請承認
- `GET /api/friends` - 友達一覧
- `GET /api/friends/pending` - 保留中申請

### 管理者（要管理者権限）
- `POST /api/admin/points/grant` - ポイント付与
- `POST /api/admin/points/deduct` - ポイント減算
- `GET /api/admin/users` - 全ユーザー一覧
- `GET /api/admin/transactions` - 全トランザクション
- `POST /api/admin/users/role` - ユーザー役割変更
- `POST /api/admin/users/deactivate` - ユーザー無効化



### ポイント転送の安全性保証

```go
// 1. 冪等性チェック（重複送金防止）
existingKey := idempotencyRepo.FindByKey(req.IdempotencyKey)
if existingKey.Status == "completed" {
    return existingTransaction // 完了済みなら既存の結果を返す
}

// 2. トランザクション開始
db.Transaction(func(tx *gorm.DB) error {
    // 3. SELECT FOR UPDATE（悲観的ロック）
    userRepo.UpdateBalanceWithLock(tx, fromUserID, amount, true)
    userRepo.UpdateBalanceWithLock(tx, toUserID, amount, false)

    // 4. トランザクション記録作成
    transactionRepo.Create(tx, transaction)

    // 5. 冪等性キー更新
    idempotencyKey.Status = "completed"
    idempotencyRepo.Update(idempotencyKey)

    return nil // コミット
})
```



## 📈 今後の拡張案

- [ ] テスト拡張
- [ ] フレンド拡張
- [ ] 監査ログの実装
- [ ] Webhookサポート
- [ ] メール通知
- [ ] キャッシュ層（Redis）
- [ ] メトリクス・モニタリング（Prometheus）



### コードスタイル
- Go: `gofmt`でフォーマット
- React: ESLint + Prettier

### テスト
```bash
# バックエンド
cd backend
go test ./...

# フロントエンド
cd frontend
npm test
```

