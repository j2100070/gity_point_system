# Gity Point System

社内向けのQRコードベース・ポイント管理プラットフォーム

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

Gity Point Systemは、React + Go + MySQLで構築されたポイント管理プラットフォームです。クリーンアーキテクチャを採用し、テスタビリティ、保守性、拡張性を重視して設計されています。

### 特徴

- **完全なACID保証**: トランザクションによるデータ整合性
- **デッドロック対策**: UUID順序ロックによる同時実行制御
- **冪等性保証**: Idempotency Keyで重複トランザクション防止
- **セキュアな認証**: Session + CSRF保護 + メール認証
- **QRコード送受信**: PayPayライクなユーザー体験
- **Akerunデイリーボーナス**: 入退室連携によるくじ引きボーナス
- **商品交換**: ポイントで商品と交換可能
- **ポイント有効期限**: FIFO消費による期限管理
- **管理者機能**: ポイント付与・減算、ユーザー管理、ダッシュボード

---

## 主な機能

### ユーザー機能

#### 認証・アカウント管理
- ユーザー登録 (メール・パスワード・氏名)
- メール認証 (登録時・変更時)
- ログイン / ログアウト
- セッション管理 (24時間有効)
- CSRF保護

#### プロフィール・設定
- プロフィール編集 (表示名、メール、氏名)
- アバター画像のアップロード・削除
- ユーザー名変更 (変更履歴記録)
- パスワード変更 (変更履歴記録)
- アカウント削除 (アーカイブ化)

#### ポイント転送
- **直接送金**: ユーザー間でポイント転送
- **PayPay風送金リクエスト**: 個人QRコードをスキャンして送金リクエスト作成、受取人が承認で完了
- **マイQRコード**: 永続的な個人QRコード（有効期限なし）
- **送金リクエスト管理**: 受信・送信リクエストの承認、拒否、キャンセル
- **取引履歴**: 全トランザクションの閲覧
- **残高確認**: リアルタイム残高表示

#### デイリーボーナス（くじ引き）
- **Akerun入退室連動**: Akerunアクセス記録から自動でボーナス付与
- **くじ引きアニメーション**: ボーナス受取時のくじ引き演出
- **抽選ティア**: 管理者設定の確率別ポイント付与（大当たり・当たり・ハズレ等）
- **本日のボーナス確認**: 今日のボーナス獲得状況
- **ボーナス履歴**: 過去の獲得ボーナス一覧

#### 友達機能
- 友達申請の送信
- 友達申請の承認・拒否
- 友達一覧の表示
- 保留中の申請表示

#### 商品交換
- 商品カタログ閲覧（カテゴリフィルタ付き）
- ポイントで商品交換
- 交換履歴の閲覧

### 管理者機能

#### ダッシュボード
- ユーザー統計（総ユーザー数、アクティブ数、管理者数）
- ポイント統計（総ポイント数、平均残高）
- 日別トランザクション推移グラフ

#### ポイント管理
- ユーザーへのポイント付与
- ユーザーからのポイント減算
- 理由・説明の記録

#### ユーザー管理
- 全ユーザー一覧表示（検索・ソート対応）
- ユーザー役割変更 (user ⇔ admin)
- アカウント無効化 / 復元

#### 商品・カテゴリ管理
- 商品の作成・編集・削除
- カテゴリの作成・編集・削除・並び替え
- 在庫管理

#### ボーナス設定
- デフォルトボーナスポイント設定
- 抽選ティアの作成・編集・確率設定

#### 監査
- 全トランザクション履歴の閲覧（種別・日付フィルタ対応）
- 管理者操作ログの記録

### バックグラウンドワーカー

#### Akerun Worker
- Akerun APIを定期ポーリング（5分間隔）
- アクセス記録からユーザー名マッチング
- 自動ボーナス付与（くじ引き方式）
- リカバリモード（長時間停止後の自動復旧）

#### ポイント有効期限Worker
- 期限切れポイントバッチの検出
- FIFO方式でのポイント消費管理

---

## アーキテクチャ

### バックエンド: クリーンアーキテクチャ (5層構造)

```
backend/
├── entities/                    # 第1層: エンティティ (ドメインモデル)
│   ├── user.go                 # ユーザー + プロフィール管理
│   ├── transaction.go          # トランザクション (転送/付与/減算/交換)
│   ├── friendship.go           # 友達関係
│   ├── transfer_request.go     # 送金リクエスト
│   ├── qrcode.go              # QRコード
│   ├── session.go             # セッション
│   ├── daily_bonus.go         # デイリーボーナス + NormalizeName
│   ├── lottery_tier.go        # 抽選ティア + DrawLottery
│   ├── access_record.go       # Akerunアクセス記録DTO
│   ├── product.go             # 商品 + 商品交換
│   ├── category.go            # 商品カテゴリ
│   ├── point_batch.go         # ポイントバッチ (有効期限管理)
│   ├── archived_user.go       # アーカイブ済みユーザー
│   ├── email_verification.go  # メール認証トークン
│   ├── username_change_history.go  # ユーザー名変更履歴
│   ├── password_change_history.go  # パスワード変更履歴
│   ├── crypto.go              # パスワードサービスIF
│   └── logger.go              # ロガーインターフェース
│
├── usecases/                   # 第2層: ユースケース (ビジネスロジック)
│   ├── inputport/             # 入力ポート (ユースケースインターフェース)
│   │   ├── auth_inputport.go
│   │   ├── point_transfer_inputport.go
│   │   ├── transfer_request_inputport.go
│   │   ├── qrcode_inputport.go
│   │   ├── friendship_inputport.go
│   │   ├── admin_inputport.go
│   │   ├── daily_bonus_inputport.go
│   │   └── akerun_bonus_inputport.go
│   ├── interactor/            # インタラクター (ユースケース実装)
│   │   ├── auth_interactor.go
│   │   ├── point_transfer_interactor.go
│   │   ├── transfer_request_interactor.go
│   │   ├── qrcode_interactor.go
│   │   ├── friendship_interactor.go
│   │   ├── admin_interactor.go
│   │   ├── daily_bonus_interactor.go   # ボーナス参照 + Akerun付与ロジック
│   │   ├── product_management_interactor.go
│   │   ├── product_exchange_interactor.go
│   │   ├── category_management_interactor.go
│   │   ├── user_settings_interactor.go
│   │   └── user_query_interactor.go
│   ├── repository/            # リポジトリインターフェース (出力ポート)
│   └── service/               # サービスインターフェース
│       └── akerun_access_gateway.go
│
├── gateways/                   # 第3層: ゲートウェイ (リポジトリ実装)
│   ├── repository/            # リポジトリ実装層
│   ├── datasource/            # データソース実装
│   │   └── dsmysqlimpl/       # MySQL実装
│   └── infra/                 # インフラストラクチャ
│       ├── inframysql/        # DB接続
│       ├── infraakerun/       # Akerun API連携 (Worker + Client)
│       ├── infralogger/       # ロガー実装
│       ├── infrastorage/      # ファイルストレージ (アバター)
│       ├── infraemail/        # メール送信
│       └── infra/             # ポイント有効期限Worker
│
├── controllers/                # 第4層: コントローラー (入出力変換)
│   └── web/
│       ├── auth_controller.go
│       ├── point_controller.go
│       ├── transfer_request_controller.go
│       ├── qrcode_controller.go
│       ├── friend_controller.go
│       ├── admin_controller.go
│       ├── daily_bonus_controller.go
│       ├── product_controller.go
│       ├── category_controller.go
│       ├── user_settings_controller.go
│       └── presenter/         # プレゼンター (出力フォーマット)
│
├── frameworks/                 # 第5層: フレームワーク・外部ツール
│   └── web/
│       ├── router.go          # Ginルーター設定
│       ├── middleware/        # ミドルウェア (認証, CSRF, セキュリティ)
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
    ├── 001_initial_schema.sql
    ├── 002_add_product_exchange.sql
    ├── 003_add_categories.sql
    ├── 004_add_friendships_archive.sql
    ├── 005_add_user_settings.sql
    ├── 006_add_transfer_requests.sql
    ├── 007_add_daily_bonuses.sql
    ├── 008_add_daily_bonus_transaction_type.sql
    ├── 009_optimize_indexes.sql
    ├── 010_add_real_name.sql
    ├── 011_akerun_daily_bonus.sql
    ├── 012_lottery_bonus.sql
    └── 012_point_expiration.sql
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

### フロントエンド: Feature-based Architecture

```
frontend/
├── src/
│   ├── features/              # 機能ごとのモジュール
│   │   ├── auth/             # 認証 (ログイン, 登録)
│   │   ├── dashboard/        # ダッシュボード
│   │   ├── points/           # ポイント (転送, 履歴)
│   │   ├── transfer-requests/ # 送金リクエスト (個人QR, 一覧)
│   │   ├── qrcode/           # QRコードスキャン
│   │   ├── friends/          # 友達機能
│   │   ├── daily-bonus/      # デイリーボーナス (くじ引き)
│   │   ├── products/         # 商品カタログ・交換
│   │   ├── settings/         # ユーザー設定 (プロフィール, アバター, パスワード)
│   │   └── admin/            # 管理者 (ダッシュボード, ユーザー管理, 取引, ボーナス設定)
│   │
│   ├── core/                  # ドメイン層
│   │   ├── domain/           # ドメインモデル
│   │   └── repositories/     # リポジトリインターフェース
│   │
│   ├── infrastructure/        # インフラストラクチャ層
│   │   └── api/              # APIクライアント + リポジトリ実装
│   │
│   └── shared/               # 共通モジュール
│       ├── components/       # 共通コンポーネント
│       └── stores/           # 状態管理 (Zustand)
```

---

## 技術スタック

### バックエンド

| 技術 | バージョン | 用途 |
|-----|----------|------|
| Go | 1.23+ | メイン言語 |
| Gin | v1.9+ | HTTPフレームワーク |
| GORM | v1.25+ | ORM |
| MySQL | 8.0+ | メインデータベース |
| golang.org/x/crypto/bcrypt | - | パスワードハッシュ化 |
| google/uuid | v1.6+ | UUID生成 |

### フロントエンド

| 技術 | バージョン | 用途 |
|-----|----------|------|
| React | 18+ | UIフレームワーク |
| TypeScript | 5+ | 型安全性 |
| Vite | 5+ | ビルドツール |
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
  ├─────< (N) sessions                          │
  ├─────< (N) qr_codes                          │
  ├─────< (N) transfer_requests (N) ───────────┤
  ├─────< (N) friendships (N) ─────────────────┘
  ├─────< (N) daily_bonuses ───> lottery_tiers
  ├─────< (N) product_exchanges ───> products ───> categories
  ├─────< (N) point_batches
  ├─────< (N) email_verification_tokens
  ├─────< (N) username_change_histories
  └─────< (N) password_change_histories

archived_users (アカウント削除時に移動)
system_settings (ボーナスポイント等のシステム設定)
```

### 主要テーブル

| テーブル | 説明 |
|---------|------|
| `users` | ユーザー情報（残高、役割、氏名、アバター） |
| `transactions` | ポイント取引（転送、付与、減算、交換、ボーナス） |
| `sessions` | セッション管理 |
| `qr_codes` | QRコード |
| `transfer_requests` | 送金リクエスト |
| `friendships` | 友達関係 |
| `daily_bonuses` | デイリーボーナス記録（Akerun連携） |
| `lottery_tiers` | 抽選ティア設定（くじ引き確率・ポイント） |
| `products` | 商品マスタ |
| `categories` | 商品カテゴリ |
| `product_exchanges` | 商品交換履歴 |
| `point_batches` | ポイントバッチ（FIFO有効期限管理） |
| `idempotency_keys` | 冪等性キー |
| `email_verification_tokens` | メール認証トークン |
| `username_change_histories` | ユーザー名変更履歴 |
| `password_change_histories` | パスワード変更履歴 |
| `archived_users` | アーカイブ済みユーザー |
| `system_settings` | システム設定（Key-Value） |

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
- MySQL: localhost:3306

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

`.env` ファイルまたは `docker-compose.yml` で設定:

**バックエンド:**
```yaml
DB_HOST: db
DB_PORT: 3306
DB_USER: root
DB_PASSWORD: password
DB_NAME: point_system
SERVER_PORT: 8080
ALLOWED_ORIGINS: http://localhost:3000,http://localhost:5173
AKERUN_ACCESS_TOKEN: (Akerun APIトークン)
AKERUN_ORGANIZATION_ID: (Akerun組織ID)
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

| メソッド | パス | 説明 | 認証 |
|---------|------|------|------|
| POST | `/api/auth/register` | ユーザー登録 | 不要 |
| POST | `/api/auth/login` | ログイン | 不要 |
| POST | `/api/auth/logout` | ログアウト | 要 |
| GET | `/api/auth/me` | 現在のユーザー情報 | 要 |

---

### ポイントAPI (要認証)

| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/api/points/transfer` | ポイント転送 |
| GET | `/api/points/balance` | 残高取得 |
| GET | `/api/points/history` | 取引履歴取得 |

---

### 送金リクエストAPI (要認証)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/transfer-requests/personal-qr` | 個人QRコード取得 |
| POST | `/api/transfer-requests` | 送金リクエスト作成 |
| GET | `/api/transfer-requests/pending` | 承認待ちリクエスト |
| GET | `/api/transfer-requests/sent` | 送信済みリクエスト |
| GET | `/api/transfer-requests/pending/count` | 承認待ち件数 |
| GET | `/api/transfer-requests/:id` | リクエスト詳細 |
| POST | `/api/transfer-requests/:id/approve` | 承認 |
| POST | `/api/transfer-requests/:id/reject` | 拒否 |
| DELETE | `/api/transfer-requests/:id` | キャンセル |

---

### QRコードAPI (要認証)

| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/api/qrcode/generate` | QRコード生成 |
| POST | `/api/qrcode/scan` | QRコードスキャン |

---

### 友達API (要認証)

| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/api/friends/request` | 友達申請送信 |
| POST | `/api/friends/accept` | 友達申請承認 |
| POST | `/api/friends/reject` | 友達申請拒否 |
| GET | `/api/friends` | 友達一覧 |
| GET | `/api/friends/pending` | 保留中の申請 |

---

### デイリーボーナスAPI (要認証)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/daily-bonus/today` | 本日のボーナス状況 |
| GET | `/api/daily-bonus/recent` | 最近のボーナス履歴 |
| GET | `/api/daily-bonus/settings` | ボーナス設定取得 |
| POST | `/api/daily-bonus/:id/viewed` | ボーナス閲覧済みマーク |

---

### 商品API (要認証)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/products` | 商品一覧 |
| GET | `/api/products/:id` | 商品詳細 |
| POST | `/api/products/:id/exchange` | 商品交換 |
| GET | `/api/products/exchanges` | 交換履歴 |

---

### カテゴリAPI (要認証)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/api/categories` | カテゴリ一覧 |

---

### ユーザー設定API (要認証)

| メソッド | パス | 説明 |
|---------|------|------|
| PUT | `/api/settings/profile` | プロフィール更新 |
| POST | `/api/settings/avatar` | アバターアップロード |
| DELETE | `/api/settings/avatar` | アバター削除 |
| PUT | `/api/settings/username` | ユーザー名変更 |
| PUT | `/api/settings/password` | パスワード変更 |
| POST | `/api/settings/verify-email` | メール認証送信 |
| POST | `/api/settings/confirm-email` | メール認証確認 |
| DELETE | `/api/settings/account` | アカウント削除 |

---

### 管理者API (要管理者権限)

| メソッド | パス | 説明 |
|---------|------|------|
| POST | `/api/admin/points/grant` | ポイント付与 |
| POST | `/api/admin/points/deduct` | ポイント減算 |
| GET | `/api/admin/users` | ユーザー一覧（検索・ソート対応） |
| GET | `/api/admin/transactions` | トランザクション一覧（フィルタ対応） |
| POST | `/api/admin/users/role` | ユーザー役割変更 |
| POST | `/api/admin/users/deactivate` | ユーザー無効化 |
| GET | `/api/admin/dashboard` | ダッシュボード統計 |
| GET | `/api/admin/bonus/settings` | ボーナス設定 |
| PUT | `/api/admin/bonus/lottery-tiers` | 抽選ティア更新 |
| POST | `/api/admin/products` | 商品作成 |
| PUT | `/api/admin/products/:id` | 商品更新 |
| DELETE | `/api/admin/products/:id` | 商品削除 |
| POST | `/api/admin/categories` | カテゴリ作成 |
| PUT | `/api/admin/categories/:id` | カテゴリ更新 |
| DELETE | `/api/admin/categories/:id` | カテゴリ削除 |

---

## セキュリティ

### 認証・認可

#### セッションベース認証
- **Cookie**: `session_token` (HttpOnly, SameSite=Lax)
- **有効期限**: 24時間
- **セッション管理**: MySQLに永続化

#### CSRF保護
- CSRFトークンをセッションと紐付け
- ミドルウェアで検証

#### パスワードセキュリティ
- **ハッシュアルゴリズム**: bcrypt (cost=10)
- **最小長**: 8文字
- パスワードは平文保存なし

### トランザクション保護

#### 冪等性保証
```go
// Idempotency Keyで重複送金を防止
existingKey := idempotencyRepo.FindByKey(req.IdempotencyKey)
if existingKey.Status == "completed" {
    return existingTransaction // 同じ結果を返す
}
```

#### 悲観的ロック (SELECT FOR UPDATE)
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

#### トランザクション分離
- **分離レベル**: REPEATABLE READ (金融システム要件)
- **ロック戦略**: 行レベル悲観的ロック + UUID順序ロック

### データ整合性

#### 残高保護
```sql
-- DB制約で負の値を防止
balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
```

#### 自己送金防止
```sql
CHECK (from_user_id != to_user_id)
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

# バックエンド統合テスト (MySQL必要)
go test -tags=integration ./tests/integration/... -v

# フロントエンド
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
- **新しいフロントエンド機能**: `features/` に機能単位でディレクトリ追加

---

## ライセンス

MIT License
