# Gity Point System

PayPayのようなQRコードベースのポイント送受信システム

## 目次

- [概要](#概要)
- [主な機能](#主な機能)
- [アーキテクチャ](#アーキテクチャ)
- [技術スタック](#技術スタック)
- [データベース設計](#データベース設計)
- [セットアップ](#セットアップ)
- [API仕様](#api仕様)
- [セキュリティ](#セキュリティ)
- [開発](#開発)

---

## 概要

Gity Point Systemは、React + Go + PostgreSQLで構築されたポイント管理プラットフォームです。クリーンアーキテクチャを採用し、テスタビリティ、保守性、拡張性を重視して設計されています。

### 特徴

- **完全なACID保証**: PostgreSQLトランザクションによるデータ整合性
- **デッドロック対策**: UUID順序ロックによる同時実行制御
- **冪等性保証**: Idempotency Keyで重複トランザクション防止
- **セキュアな認証**: Session + CSRF保護
- **QRコード送受信**: PayPayライクなユーザー体験
- **管理者機能**: ポイント付与・減算、ユーザー管理

---

## 主な機能

### ユーザー機能

#### 認証・アカウント管理
- ユーザー登録 (メール・パスワード)
- ログイン / ログアウト
- セッション管理 (24時間有効)
- CSRF保護

#### ポイント転送
- **直接送金**: ユーザー間でポイント転送
- **PayPay風送金リクエスト**: 個人QRコードをスキャンして送金リクエスト作成、受取人が承認で完了
- **マイQRコード**: 永続的な個人QRコード（有効期限なし）
- **送金リクエスト管理**: 受信・送信リクエストの承認、拒否、キャンセル
- **取引履歴**: 全トランザクションの閲覧
- **残高確認**: リアルタイム残高表示

#### 友達機能
- 友達申請の送信
- 友達申請の承認・拒否
- 友達一覧の表示
- 保留中の申請表示


### 管理者機能

#### ポイント管理
- ユーザーへのポイント付与
- ユーザーからのポイント減算
- 理由・説明の記録

#### ユーザー管理
- 全ユーザー一覧表示
- ユーザー役割変更 (user ⇔ admin)
- アカウント無効化

#### 監査
- 全トランザクション履歴の閲覧
- 管理者操作ログの記録

---

## アーキテクチャ

### バックエンド: クリーンアーキテクチャ (5層構造)

```
backend/
├── entities/                    # 第1層: エンティティ (ドメインモデル)
│   ├── user.go                 # ユーザーエンティティ + ビジネスルール
│   ├── transaction.go          # トランザクションエンティティ
│   ├── friendship.go           # 友達関係エンティティ
│   ├── transfer_request.go     # 送金リクエストエンティティ
│   ├── qrcode.go              # QRコードエンティティ
│   ├── session.go             # セッションエンティティ
│   └── logger.go              # ロガーインターフェース
│
├── usecases/                   # 第2層: ユースケース (ビジネスロジック)
│   ├── inputport/             # 入力ポート (ユースケースインターフェース)
│   │   ├── auth_inputport.go
│   │   ├── point_transfer_inputport.go
│   │   ├── transfer_request_inputport.go
│   │   ├── qrcode_inputport.go
│   │   ├── friendship_inputport.go
│   │   └── admin_inputport.go
│   ├── interactor/            # インタラクター (ユースケース実装)
│   │   ├── auth_interactor.go
│   │   ├── point_transfer_interactor.go
│   │   ├── transfer_request_interactor.go
│   │   ├── qrcode_interactor.go
│   │   ├── friendship_interactor.go
│   │   └── admin_interactor.go
│   └── repository/            # リポジトリインターフェース (出力ポート)
│       ├── user_repository.go
│       ├── transaction_repository.go
│       ├── session_repository.go
│       ├── transfer_request_repository.go
│       ├── qrcode_repository.go
│       └── friendship_repository.go
│
├── gateways/                   # 第3層: ゲートウェイ (リポジトリ実装)
│   ├── repository/            # リポジトリ実装層
│   │   ├── user/             # ユーザーリポジトリ実装
│   │   ├── transaction/      # トランザクションリポジトリ実装
│   │   ├── session/          # セッションリポジトリ実装
│   │   ├── transfer_request/ # 送金リクエストリポジトリ実装
│   │   ├── qrcode/           # QRコードリポジトリ実装
│   │   └── friendship/       # 友達リポジトリ実装
│   ├── datasource/           # データソース実装
│   │   └── dsmysqlimpl/      # PostgreSQL実装
│   └── infra/                # インフラストラクチャ
│       ├── inframysql/       # DB接続
│       └── infralogger/      # ロガー実装
│
├── controllers/                # 第4層: コントローラー (入出力変換)
│   └── web/
│       ├── auth_controller.go
│       ├── point_controller.go
│       ├── transfer_request_controller.go
│       ├── qrcode_controller.go
│       ├── friend_controller.go
│       ├── admin_controller.go
│       └── presenter/         # プレゼンター (出力フォーマット)
│           ├── auth_presenter.go
│           ├── point_presenter.go
│           ├── transfer_request_presenter.go
│           ├── qrcode_presenter.go
│           ├── friend_presenter.go
│           └── admin_presenter.go
│
├── frameworks/                 # 第5層: フレームワーク・外部ツール
│   └── web/
│       ├── router.go          # Ginルーター設定
│       ├── middleware/        # ミドルウェア
│       │   ├── auth.go       # 認証ミドルウェア
│       │   ├── csrf.go       # CSRF保護
│       │   └── security.go   # セキュリティヘッダー
│       └── time_provider.go   # 時刻プロバイダー
│
├── cmd/
│   └── clean_server/          # アプリケーションエントリーポイント
│       └── main.go
│
├── config/                     # 設定管理
│   └── config.go
│
└── migrations/                 # DBマイグレーション
    ├── 001_init_schema.sql
    ├── 002_update_passwords.sql
    └── 006_add_transfer_requests.sql
```

#### 依存関係の方向

```
Frameworks (Web/DB)
    ↓ depends on
Controllers (HTTP Handlers)
    ↓ depends on
Gateways (Repository Impl)
    ↓ depends on
UseCases (Business Logic)
    ↓ depends on
Entities (Domain Models)
```

**重要な原則:**
- 内側の層は外側の層を知らない (依存性逆転の原則)
- インターフェースによる疎結合
- ドメイン層 (Entities) は完全に独立

### フロントエンド: Feature-based + Clean Architecture

```
frontend/
├── src/
│   ├── features/              # 機能ごとのモジュール
│   │   ├── auth/             # 認証機能
│   │   │   └── pages/
│   │   │       ├── LoginPage.tsx
│   │   │       └── RegisterPage.tsx
│   │   ├── dashboard/        # ダッシュボード
│   │   │   └── pages/
│   │   │       └── DashboardPage.tsx
│   │   ├── points/           # ポイント機能
│   │   │   └── pages/
│   │   │       ├── TransferPage.tsx
│   │   │       └── HistoryPage.tsx
│   │   ├── transfer-requests/ # 送金リクエスト機能
│   │   │   └── pages/
│   │   │       ├── PersonalQRPage.tsx
│   │   │       └── TransferRequestsPage.tsx
│   │   ├── qrcode/           # QRコード機能
│   │   │   └── pages/
│   │   │       └── ScanQRPage.tsx
│   │   ├── friends/          # 友達機能
│   │   │   └── pages/
│   │   │       └── FriendsPage.tsx
│   │   └── admin/            # 管理者機能
│   │       └── pages/
│   │           ├── AdminDashboardPage.tsx
│   │           ├── AdminUsersPage.tsx
│   │           └── AdminTransactionsPage.tsx
│   │
│   ├── core/                  # ドメイン層
│   │   ├── domain/           # ドメインモデル
│   │   │   ├── User.ts
│   │   │   ├── Transaction.ts
│   │   │   ├── TransferRequest.ts
│   │   │   ├── QRCode.ts
│   │   │   └── Friendship.ts
│   │   └── repositories/     # リポジトリインターフェース
│   │       └── interfaces.ts
│   │
│   ├── infrastructure/        # インフラストラクチャ層
│   │   └── api/
│   │       ├── client.ts     # API クライアント
│   │       └── repositories/ # リポジトリ実装
│   │           ├── AuthRepository.ts
│   │           ├── PointRepository.ts
│   │           ├── TransferRequestRepository.ts
│   │           ├── QRCodeRepository.ts
│   │           ├── FriendshipRepository.ts
│   │           └── AdminRepository.ts
│   │
│   └── shared/               # 共通モジュール
│       ├── components/       # 共通コンポーネント
│       │   ├── Layout.tsx
│       │   └── ProtectedRoute.tsx
│       └── stores/           # 状態管理
│           └── authStore.ts  # Zustand store
```

---

## 技術スタック

### バックエンド

| 技術 | バージョン | 用途 |
|-----|----------|------|
| Go | 1.23+ | メイン言語 |
| Gin | v1.9+ | HTTPフレームワーク |
| GORM | v1.25+ | ORM |
| PostgreSQL | 15+ | メインデータベース |
| golang.org/x/crypto/bcrypt | - | パスワードハッシュ化 |
| google/uuid | v1.6+ | UUID生成 |

### フロントエンド

| 技術 | バージョン | 用途 |
|-----|----------|------|
| React | 18+ | UIフレームワーク |
| TypeScript | 5+ | 型安全性 |
| React Router | v6+ | ルーティング |
| Zustand | - | 状態管理 |
| Axios | - | HTTP クライアント |
| QRCode.react | - | QRコード生成 |
| Html5-qrcode | - | QRコードスキャン |

### インフラ

| 技術 | 用途 |
|-----|------|
| Docker | コンテナ化 |
| Docker Compose | オーケストレーション |

---

## データベース設計

### ER図 (概念)

```
users (1) ─────< (N) transactions (N) ─────> (1) users
  │                                              │
  │                                              │
  ├─────< (N) sessions                          │
  │                                              │
  ├─────< (N) qr_codes                          │
  │                                              │
  ├─────< (N) transfer_requests (N) ───────────┤
  │                                              │
  └─────< (N) friendships (N) ─────────────────┘
```

### テーブル詳細

#### users (ユーザー)

| カラム | 型 | 制約 | 説明 |
|-------|------|------|------|
| id | UUID | PK | ユーザーID |
| username | VARCHAR(50) | UNIQUE, NOT NULL | ユーザー名 |
| email | VARCHAR(255) | UNIQUE, NOT NULL | メールアドレス |
| password_hash | VARCHAR(255) | NOT NULL | bcryptハッシュ |
| display_name | VARCHAR(100) | - | 表示名 |
| balance | BIGINT | NOT NULL, CHECK >= 0 | 残高 (負の値禁止) |
| role | VARCHAR(20) | NOT NULL, DEFAULT 'user' | 役割 (user/admin) |
| version | INTEGER | NOT NULL, DEFAULT 1 | 楽観的ロック用 |
| is_active | BOOLEAN | NOT NULL, DEFAULT true | アカウント有効フラグ |
| personal_qr_code | VARCHAR(255) | NOT NULL | 個人固定QRコード (user:{uuid}形式) |
| created_at | TIMESTAMPTZ | NOT NULL | 作成日時 |
| updated_at | TIMESTAMPTZ | NOT NULL | 更新日時 |

**インデックス:**
- `idx_users_username` on username
- `idx_users_email` on email
- `idx_users_is_active` on is_active

#### transactions (トランザクション)

| カラム | 型 | 制約 | 説明 |
|-------|------|------|------|
| id | UUID | PK | トランザクションID |
| from_user_id | UUID | FK users(id) | 送信者ID (NULL=システム付与) |
| to_user_id | UUID | FK users(id) | 受信者ID (NULL=システム減算) |
| amount | BIGINT | NOT NULL, CHECK > 0 | 金額 |
| transaction_type | VARCHAR(50) | NOT NULL | 種別 (transfer/qr_receive/admin_grant等) |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | 状態 (pending/completed/failed) |
| idempotency_key | VARCHAR(255) | UNIQUE | 冪等性キー |
| description | TEXT | NOT NULL | 説明 |
| metadata | JSONB | DEFAULT '{}' | メタデータ |
| created_at | TIMESTAMPTZ | NOT NULL | 作成日時 |
| completed_at | TIMESTAMPTZ | - | 完了日時 |

**制約:**
- `from_user_id IS NOT NULL OR to_user_id IS NOT NULL`
- `from_user_id != to_user_id` (自分自身への送金禁止)

**インデックス:**
- `idx_transactions_from_user` on from_user_id
- `idx_transactions_to_user` on to_user_id
- `idx_transactions_idempotency` on idempotency_key
- `idx_transactions_status` on status
- `idx_transactions_created_at` on created_at DESC

#### sessions (セッション)

| カラム | 型 | 制約 | 説明 |
|-------|------|------|------|
| id | UUID | PK | セッションID |
| user_id | UUID | FK users(id), NOT NULL | ユーザーID |
| token | VARCHAR(255) | UNIQUE, NOT NULL | セッショントークン |
| csrf_token | VARCHAR(255) | NOT NULL | CSRFトークン |
| expires_at | TIMESTAMPTZ | NOT NULL | 有効期限 |
| created_at | TIMESTAMPTZ | NOT NULL | 作成日時 |
| last_accessed_at | TIMESTAMPTZ | NOT NULL | 最終アクセス日時 |

**インデックス:**
- `idx_sessions_user` on user_id
- `idx_sessions_token` on token
- `idx_sessions_expires` on expires_at

#### idempotency_keys (冪等性キー)

| カラム | 型 | 制約 | 説明 |
|-------|------|------|------|
| key | VARCHAR(255) | PK | 冪等性キー |
| user_id | UUID | FK users(id), NOT NULL | ユーザーID |
| transaction_id | UUID | FK transactions(id) | 関連トランザクションID |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'processing' | 状態 (processing/completed/failed) |
| created_at | TIMESTAMPTZ | NOT NULL | 作成日時 |
| expires_at | TIMESTAMPTZ | NOT NULL | 有効期限 |

**インデックス:**
- `idx_idempotency_user` on user_id
- `idx_idempotency_expires` on expires_at

#### friendships (友達関係)

| カラム | 型 | 制約 | 説明 |
|-------|------|------|------|
| id | UUID | PK | 友達関係ID |
| requester_id | UUID | FK users(id), NOT NULL | 申請者ID |
| addressee_id | UUID | FK users(id), NOT NULL | 相手ID |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | 状態 (pending/accepted/rejected/blocked) |
| created_at | TIMESTAMPTZ | NOT NULL | 作成日時 |
| updated_at | TIMESTAMPTZ | NOT NULL | 更新日時 |

**制約:**
- `requester_id != addressee_id`
- UNIQUE(requester_id, addressee_id)

**インデックス:**
- `idx_friendships_requester` on requester_id
- `idx_friendships_addressee` on addressee_id
- `idx_friendships_status` on status

#### qr_codes (QRコード)

| カラム | 型 | 制約 | 説明 |
|-------|------|------|------|
| id | UUID | PK | QRコードID |
| user_id | UUID | FK users(id), NOT NULL | ユーザーID |
| code | VARCHAR(255) | UNIQUE, NOT NULL | QRコード文字列 |
| qr_type | VARCHAR(20) | NOT NULL | 種別 (receive/send) |
| amount | BIGINT | CHECK > 0 | 金額 |
| expires_at | TIMESTAMPTZ | NOT NULL | 有効期限 |
| used_at | TIMESTAMPTZ | - | 使用日時 |
| used_by_user_id | UUID | FK users(id) | 使用者ID |
| created_at | TIMESTAMPTZ | NOT NULL | 作成日時 |

**インデックス:**
- `idx_qrcodes_user` on user_id
- `idx_qrcodes_code` on code
- `idx_qrcodes_expires` on expires_at

#### transfer_requests (送金リクエスト)

| カラム | 型 | 制約 | 説明 |
|-------|------|------|------|
| id | UUID | PK | リクエストID |
| from_user_id | UUID | FK users(id), NOT NULL | 送信者ID |
| to_user_id | UUID | FK users(id), NOT NULL | 受取人ID |
| amount | BIGINT | NOT NULL, CHECK > 0 | 金額 |
| message | TEXT | DEFAULT '' | メッセージ |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | 状態 (pending/approved/rejected/cancelled/expired) |
| idempotency_key | VARCHAR(255) | UNIQUE, NOT NULL | 冪等性キー |
| expires_at | TIMESTAMPTZ | NOT NULL | 有効期限 (24時間) |
| approved_at | TIMESTAMPTZ | - | 承認日時 |
| rejected_at | TIMESTAMPTZ | - | 拒否日時 |
| cancelled_at | TIMESTAMPTZ | - | キャンセル日時 |
| transaction_id | UUID | FK transactions(id) | 関連トランザクションID |
| created_at | TIMESTAMPTZ | NOT NULL | 作成日時 |
| updated_at | TIMESTAMPTZ | NOT NULL | 更新日時 |

**制約:**
- `from_user_id != to_user_id` (自分自身への送金禁止)

**インデックス:**
- `idx_transfer_requests_from_user` on from_user_id
- `idx_transfer_requests_to_user` on to_user_id
- `idx_transfer_requests_status` on status
- `idx_transfer_requests_idempotency` on idempotency_key
- `idx_transfer_requests_expires_at` on expires_at

---

## セットアップ

### 前提条件

- Docker & Docker Compose
- Git

### クイックスタート

```bash
# 1. リポジトリのクローン
git clone https://github.com/yourusername/gity_point_system.git
cd gity_point_system

# 2. Docker Composeで全サービス起動
docker-compose up -d

# 3. 起動確認
docker-compose ps
```

サービスが起動したら:
- フロントエンド: http://localhost:5173
- バックエンドAPI: http://localhost:8080
- PostgreSQL: localhost:5432

### 初期アカウント

データベースマイグレーションで自動作成されます:

**管理者アカウント:**
- Username: `admin`
- Password: `admin123`
- Balance: 1,000,000 points
- Role: admin

**テストユーザー:**
- Username: `testuser`
- Password: `test123`
- Balance: 10,000 points
- Role: user

### 環境変数

`docker-compose.yml` で設定済み:

**バックエンド:**
```yaml
DB_HOST: db
DB_PORT: 5432
DB_USER: postgres
DB_PASSWORD: postgres
DB_NAME: point_system
SERVER_PORT: 8080
ALLOWED_ORIGINS: http://localhost:3000,http://localhost:5173
```

**フロントエンド:**
```yaml
VITE_API_URL: http://localhost:8080
```

---

## API仕様

### ベースURL

```
http://localhost:8080/api
```

### 認証

セッションベース認証。Cookie `session_token` を使用。

### 共通レスポンスフォーマット

**成功:**
```json
{
  "data": { ... },
  "message": "success"
}
```

**エラー:**
```json
{
  "error": "error message"
}
```

---

### 認証API

#### POST /api/auth/register
ユーザー登録

**リクエスト:**
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "password123",
  "display_name": "New User"
}
```

**バリデーション:**
- username: 3-50文字
- email: 有効なメールアドレス
- password: 8文字以上
- display_name: 1-100文字

**レスポンス (201):**
```json
{
  "user": {
    "id": "uuid",
    "username": "newuser",
    "email": "newuser@example.com",
    "display_name": "New User",
    "balance": 0,
    "role": "user"
  },
  "session": {
    "session_token": "token",
    "expires_at": "2024-01-02T00:00:00Z"
  }
}
```

#### POST /api/auth/login
ログイン

**リクエスト:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**レスポンス (200):**
```json
{
  "user": {
    "id": "uuid",
    "username": "admin",
    "balance": 1000000,
    "role": "admin"
  },
  "session": {
    "session_token": "token",
    "expires_at": "2024-01-02T00:00:00Z"
  }
}
```

#### POST /api/auth/logout
ログアウト (要認証)

**レスポンス (200):**
```json
{
  "message": "logout successful"
}
```

#### GET /api/auth/me
現在のユーザー情報取得 (要認証)

**レスポンス (200):**
```json
{
  "user": {
    "id": "uuid",
    "username": "admin",
    "email": "admin@example.com",
    "display_name": "System Administrator",
    "balance": 1000000,
    "role": "admin"
  }
}
```

---

### ポイント転送API (要認証)

#### POST /api/points/transfer
ポイント転送

**リクエスト:**
```json
{
  "to_user_id": "uuid",
  "amount": 1000,
  "description": "ランチ代",
  "idempotency_key": "unique-key-123"
}
```

**バリデーション:**
- amount: 1以上
- idempotency_key: 必須 (重複送金防止)

**レスポンス (200):**
```json
{
  "transaction": {
    "id": "uuid",
    "from_user_id": "uuid",
    "to_user_id": "uuid",
    "amount": 1000,
    "transaction_type": "transfer",
    "status": "completed",
    "description": "ランチ代",
    "created_at": "2024-01-01T12:00:00Z"
  },
  "from_user": {
    "id": "uuid",
    "balance": 9000
  },
  "to_user": {
    "id": "uuid",
    "balance": 11000
  }
}
```

**エラー (400):**
- `insufficient balance`: 残高不足
- `duplicate idempotency key`: 重複キー
- `transfer is already in progress`: 処理中

#### GET /api/points/balance
残高取得

**レスポンス (200):**
```json
{
  "balance": 10000,
  "user": {
    "id": "uuid",
    "username": "testuser",
    "balance": 10000
  }
}
```

#### GET /api/points/history
取引履歴取得

**クエリパラメータ:**
- `offset`: オフセット (デフォルト: 0)
- `limit`: 件数 (デフォルト: 20)

**レスポンス (200):**
```json
{
  "transactions": [
    {
      "id": "uuid",
      "from_user_id": "uuid",
      "to_user_id": "uuid",
      "amount": 1000,
      "transaction_type": "transfer",
      "status": "completed",
      "description": "ランチ代",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 50
}
```

---

### 送金リクエストAPI (要認証)

#### GET /api/transfer-requests/personal-qr
個人QRコード取得

**レスポンス (200):**
```json
{
  "personal_qr_code": "user:uuid",
  "user": {
    "id": "uuid",
    "username": "testuser"
  }
}
```

#### POST /api/transfer-requests
送金リクエスト作成

**リクエスト:**
```json
{
  "to_user_id": "uuid",
  "amount": 1000,
  "message": "ランチ代ありがとう",
  "idempotency_key": "unique-key-123"
}
```

**バリデーション:**
- amount: 1以上
- idempotency_key: 必須 (重複送金防止)

**レスポンス (200):**
```json
{
  "transfer_request": {
    "id": "uuid",
    "from_user_id": "uuid",
    "to_user_id": "uuid",
    "amount": 1000,
    "message": "ランチ代ありがとう",
    "status": "pending",
    "expires_at": "2024-01-02T12:00:00Z",
    "created_at": "2024-01-01T12:00:00Z"
  },
  "from_user": {
    "id": "uuid",
    "username": "sender"
  },
  "to_user": {
    "id": "uuid",
    "username": "receiver"
  }
}
```

#### GET /api/transfer-requests/pending
承認待ちリクエスト取得

**クエリパラメータ:**
- `offset`: オフセット (デフォルト: 0)
- `limit`: 件数 (デフォルト: 20)

**レスポンス (200):**
```json
{
  "requests": [
    {
      "transfer_request": {
        "id": "uuid",
        "from_user_id": "uuid",
        "to_user_id": "uuid",
        "amount": 1000,
        "message": "ランチ代ありがとう",
        "status": "pending",
        "expires_at": "2024-01-02T12:00:00Z",
        "created_at": "2024-01-01T12:00:00Z"
      },
      "from_user": {
        "id": "uuid",
        "username": "sender"
      },
      "to_user": {
        "id": "uuid",
        "username": "receiver"
      }
    }
  ]
}
```

#### GET /api/transfer-requests/sent
送信したリクエスト取得

**クエリパラメータ:**
- `offset`, `limit`

**レスポンス (200):**
同上

#### GET /api/transfer-requests/pending/count
承認待ちリクエスト数取得

**レスポンス (200):**
```json
{
  "count": 3
}
```

#### GET /api/transfer-requests/:id
リクエスト詳細取得

**レスポンス (200):**
```json
{
  "transfer_request": { ... },
  "from_user": { ... },
  "to_user": { ... }
}
```

#### POST /api/transfer-requests/:id/approve
リクエスト承認

**レスポンス (200):**
```json
{
  "transfer_request": {
    "id": "uuid",
    "status": "approved",
    "approved_at": "2024-01-01T12:05:00Z",
    "transaction_id": "uuid"
  },
  "transaction": {
    "id": "uuid",
    "from_user_id": "uuid",
    "to_user_id": "uuid",
    "amount": 1000,
    "transaction_type": "transfer",
    "status": "completed"
  },
  "from_user": {
    "id": "uuid",
    "balance": 9000
  },
  "to_user": {
    "id": "uuid",
    "balance": 11000
  }
}
```

**エラー:**
- `insufficient balance`: 残高不足
- `request has expired`: 期限切れ
- `unauthorized`: 承認権限なし

#### POST /api/transfer-requests/:id/reject
リクエスト拒否

**レスポンス (200):**
```json
{
  "transfer_request": {
    "id": "uuid",
    "status": "rejected",
    "rejected_at": "2024-01-01T12:05:00Z"
  }
}
```

#### DELETE /api/transfer-requests/:id
リクエストキャンセル (送信者のみ)

**レスポンス (200):**
```json
{
  "transfer_request": {
    "id": "uuid",
    "status": "cancelled",
    "cancelled_at": "2024-01-01T12:05:00Z"
  }
}
```

**エラー:**
- `unauthorized`: キャンセル権限なし (送信者のみ可能)

---

### 友達API (要認証)

#### POST /api/friends/request
友達申請送信

**リクエスト:**
```json
{
  "addressee_id": "uuid"
}
```

**レスポンス (200):**
```json
{
  "friendship": {
    "id": "uuid",
    "requester_id": "uuid",
    "addressee_id": "uuid",
    "status": "pending",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

#### POST /api/friends/accept
友達申請承認

**リクエスト:**
```json
{
  "friendship_id": "uuid"
}
```

**レスポンス (200):**
```json
{
  "friendship": {
    "id": "uuid",
    "status": "accepted",
    "updated_at": "2024-01-01T12:05:00Z"
  }
}
```

#### POST /api/friends/reject
友達申請拒否

**リクエスト:**
```json
{
  "friendship_id": "uuid"
}
```

#### GET /api/friends
友達一覧取得

**レスポンス (200):**
```json
{
  "friends": [
    {
      "id": "uuid",
      "user": {
        "id": "uuid",
        "username": "friend1",
        "display_name": "Friend One"
      },
      "status": "accepted",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### GET /api/friends/pending
保留中の友達申請取得

**レスポンス (200):**
```json
{
  "pending_requests": [
    {
      "id": "uuid",
      "requester": {
        "id": "uuid",
        "username": "user1",
        "display_name": "User One"
      },
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### 管理者API (要管理者権限)

#### POST /api/admin/points/grant
ポイント付与

**リクエスト:**
```json
{
  "user_id": "uuid",
  "amount": 10000,
  "description": "キャンペーン報酬",
  "idempotency_key": "admin-grant-123"
}
```

**レスポンス (200):**
```json
{
  "transaction": {
    "id": "uuid",
    "to_user_id": "uuid",
    "amount": 10000,
    "transaction_type": "admin_grant",
    "status": "completed"
  },
  "user": {
    "id": "uuid",
    "balance": 20000
  }
}
```

#### POST /api/admin/points/deduct
ポイント減算

**リクエスト:**
```json
{
  "user_id": "uuid",
  "amount": 5000,
  "description": "規約違反ペナルティ",
  "idempotency_key": "admin-deduct-456"
}
```

#### GET /api/admin/users
全ユーザー一覧

**クエリパラメータ:**
- `offset`, `limit`

**レスポンス (200):**
```json
{
  "users": [
    {
      "id": "uuid",
      "username": "user1",
      "email": "user1@example.com",
      "balance": 10000,
      "role": "user",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 100
}
```

#### GET /api/admin/transactions
全トランザクション一覧

**クエリパラメータ:**
- `offset`, `limit`

**レスポンス (200):**
```json
{
  "transactions": [ ... ],
  "total": 500
}
```

#### POST /api/admin/users/role
ユーザー役割変更

**リクエスト:**
```json
{
  "user_id": "uuid",
  "role": "admin"
}
```

**バリデーション:**
- role: "user" または "admin"

**レスポンス (200):**
```json
{
  "user": {
    "id": "uuid",
    "role": "admin"
  }
}
```

#### POST /api/admin/users/deactivate
ユーザー無効化

**リクエスト:**
```json
{
  "user_id": "uuid"
}
```

**レスポンス (200):**
```json
{
  "user": {
    "id": "uuid",
    "is_active": false
  }
}
```

**エラー:**
- `cannot deactivate yourself`: 自分自身の無効化は禁止

---

## セキュリティ

### 認証・認可

#### セッションベース認証
- **Cookie**: `session_token` (HttpOnly, SameSite=Lax)
- **有効期限**: 24時間
- **セッション管理**: PostgreSQLに永続化

#### CSRF保護
- CSRFトークンをセッションと紐付け
- ミドルウェアで検証

#### パスワードセキュリティ
- **ハッシュアルゴリズム**: bcrypt (cost=10)
- **最小長**: 8文字
- パスワードは平文保存なし

### トランザクション保護

#### 0. トランザクション分離レベル (金融システム要件)

**REPEATABLE READ を採用**

PostgreSQLでは、REPEATABLE READ分離レベルを設定することで:
- **Non-Repeatable Read防止**: 同じトランザクション内で同じデータを読むと常に同じ値
- **ファントムリード防止**: PostgreSQL特有の実装により、他DBMSと異なりファントムリードも発生しない
- **スナップショット分離**: トランザクション開始時点のデータベーススナップショットを使用

```go
// inframysql/postgres.go:67-72
// セッション全体でREPEATABLE READを設定
db.Exec("SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ")
```

**なぜREPEATABLE READを選択したか:**
1. **金融要件**: トランザクション内でのデータ一貫性が保証される
2. **PostgreSQL最適化**: ファントムリードも防止（他DBMSのREPEATABLE READより強力）
3. **パフォーマンス**: SERIALIZABLEより高速で、実用的
4. **デッドロック対策**: UUID順序ロックと組み合わせて安全性を確保

#### 1. 冪等性保証
```go
// Idempotency Keyで重複送金を防止
existingKey := idempotencyRepo.FindByKey(req.IdempotencyKey)
if existingKey.Status == "completed" {
    return existingTransaction // 同じ結果を返す
}
```

#### 2. 悲観的ロック (SELECT FOR UPDATE)
```go
// デッドロック回避: UUID順でロック
if toUserID < fromUserID {
    lock(toUserID)
    lock(fromUserID)
} else {
    lock(fromUserID)
    lock(toUserID)
}
```

#### 3. トランザクション分離
- **分離レベル**: REPEATABLE READ (金融システム要件)
- **ロック戦略**: 行レベル悲観的ロック + UUID順序ロック
- **PostgreSQLの特性**: REPEATABLE READでファントムリードも防止

#### 4. デッドロック対策 (NEW)
- **UUID順序ロック**: 小さいUUIDから順にロック取得
- 双方向送金でもデッドロック発生なし

```go
// point_transfer_interactor.go:130-173
// UUID比較: 小さい方を先にロック
if req.ToUserID.String() < req.FromUserID.String() {
    firstUserID = req.ToUserID
    secondUserID = req.FromUserID
}
```

### データ整合性

#### 残高保護
```sql
-- DB制約で負の値を防止
balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
```

#### 楽観的ロック (Version列)
```go
// 更新時にversionをチェック
WHERE id = ? AND version = ?
```

#### 自己送金防止
```sql
-- DB制約
CHECK (from_user_id != to_user_id)
```

### CORSセキュリティ

```go
// config.go
AllowedOrigins: []string{
    "http://localhost:3000",
    "http://localhost:5173",
}
```

---

## 開発

### ローカル開発環境

#### バックエンド

```bash
cd backend

# 依存関係インストール
go mod download

# ビルド
go build -o bin/clean_server ./cmd/clean_server

# 実行
./bin/clean_server
```

#### フロントエンド

```bash
cd frontend

# 依存関係インストール
npm install

# 開発サーバー起動
npm run dev
```

### テスト

```bash
# バックエンド単体テスト
cd backend
go test ./tests/unit/... -v

# バックエンド統合テスト (PostgreSQL必要)
go test -tags=integration ./tests/integration/... -v

# バックエンドE2Eテスト (サーバー起動必要)
go test -tags=e2e ./tests/e2e/... -v

# フロントエンド (今後実装予定)
cd frontend
npm test
```

### コードフォーマット

```bash
# Go
gofmt -w .

# TypeScript/React
npm run format
```

### ディレクトリ構造の追加ルール

- **新しいエンティティ**: `entities/` に追加
- **新しいユースケース**: `usecases/interactor/` + `usecases/inputport/` に追加
- **新しいリポジトリ**: `usecases/repository/` (interface) + `gateways/repository/` (実装)
- **新しいコントローラー**: `controllers/web/` + `controllers/web/presenter/`

---

## ライセンス

MIT License

---


