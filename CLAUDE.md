# CLAUDE.me — Gity Point System 開発ガイドライン

このファイルはAIアシスタントへの永続的な指示です。コード生成・レビュー・修正時に必ず従ってください。

---

## 1. クリーンアーキテクチャ

### レイヤー構造と依存方向

```
Entities → UseCases → Gateways → Controllers → Frameworks
(内側)                                              (外側)
```

**内側のレイヤーは外側のレイヤーを絶対にimportしない。** 違反（DIP違反）を検出した場合は即座に修正すること。

### 各レイヤーの責務

| レイヤー | ディレクトリ | 責務 |
|---------|------------|------|
| **Entities** | `entities/` | ビジネスロジック、ドメインモデル、全レイヤー共通のinterface (Logger等) |
| **UseCases** | `usecases/inputport/` | ユースケースのinterface定義 |
| | `usecases/interactor/` | ユースケース実装（トランザクション管理含む） |
| | `usecases/repository/` | Repository interface定義 |
| **Gateways** | `gateways/repository/datasource/dsmysql/` | DataSource interface定義 |
| | `gateways/datasource/dsmysqlimpl/` | DataSource実装（SQL操作） |
| | `gateways/repository/<entity>/` | Repository実装 |
| | `gateways/infra/` | インフラ接続（DB, Cache等） |
| **Controllers** | `controllers/web/` | HTTP入出力変換 |
| | `controllers/web/presenter/` | レスポンス整形 |
| **Frameworks** | `frameworks/web/` | ルーティング、ミドルウェア |

### 命名規則

- **Repository**: `Create` / `Read` / `Update` / `Delete` + 用途 (例: `ReadByUserID`)
- **DataSource**: `Select` / `Insert` / `Update` / `Delete` + 条件 (例: `SelectByID`)
- **JOIN結果型**: `entities` パッケージに `〇〇WithUsers` として定義（`dsmysql` には置かない）

### 型の配置ルール

| 型の種類 | 配置先 | NG例 |
|---------|--------|------|
| ドメインモデル | `entities/` | — |
| JOIN結果（WithUsers等） | `entities/` | ~~`dsmysql/`~~ |
| Repository interface | `usecases/repository/` | ~~`gateways/`~~ |
| DataSource interface | `gateways/repository/datasource/dsmysql/` | — |
| DataSource固有の変換用struct | `dsmysqlimpl/`内にprivate | — |

#時間の実装について
`time_provider.go`を参照するように修正する。

---

## 2. データベース（PostgreSQL）

### パフォーマンス

- **N+1問題の禁止**: 一覧取得は必ずJOINクエリで実装する（ループ内SELECT禁止）
- **適切なインデックス**: WHERE/JOIN/ORDER BYに使うカラムに複合インデックスを設計する
- **ページネーション**: 一覧APIには必ず `OFFSET`/`LIMIT` を使用する
- **SELECT列の明示**: `SELECT *` を避け、必要なカラムのみ取得する

### トランザクション管理

- **`txManager.Do(ctx, func(ctx context.Context) error { ... })`** でトランザクションを管理する
- トランザクション内では引数の `ctx`（txCtx）を使う。外側の `ctx` を使うとデッドロックする
- ロック順序を統一してデッドロックを防止する

### ロック戦略

- **悲観的ロック**: 残高更新等の競合防止に `SELECT FOR UPDATE` を使用
- **楽観的ロック**: `Version` フィールドによる並行制御。更新時にバージョンをインクリメントする

### マイグレーション

- `entrypoint.sh` が**毎回全マイグレーションを再実行**するため、SQLは冪等に書く
- テーブル作成: `CREATE TABLE IF NOT EXISTS`
- カラム追加: `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`
- シードデータ: `INSERT ... WHERE NOT EXISTS (SELECT 1 FROM ...)` （`ON CONFLICT DO NOTHING` はUUID PKに効かない）

---

## 3. セキュリティ

### 認証・認可

- セッションベース認証を使用（`session_token` Cookie）
- 管理者APIには `role = "admin"` チェックを必ず入れる
- CSRFトークンによる保護

### 入力バリデーション

- Controllerレイヤーで `ShouldBindJSON` によるバリデーションを行う
- SQLインジェクション防止のため、GORMのプレースホルダー (`?`) を必ず使用する
- ユーザー入力を直接SQLに組み込まない

### データ整合性

- **冪等性キー**: ポイント付与・送金には `IdempotencyKey` で重複実行を防止する
- **残高の負数防止**: トランザクション内で残高チェック → ロック → 更新の順序を守る

---

## 4. テスト

- **Interactorテスト**: `tests/unit/interactor/` にモックを使ったユニットテストを書く
- モックは同じテストファイル内にprivate structとして定義する
- モックの型は `entities` または `usecases/repository` パッケージの型を参照する（`dsmysql` を参照しない）
- テスト実行: `go test ./tests/unit/... -v -count=1`

---

## 5. フロントエンド（React + Vite）

- `frontend/src/features/` 配下に機能ごとにディレクトリを分割
- APIクライアントは各feature内の `api/` に配置
- 管理者画面は `features/admin/` 配下

---

## 6. ビルド・検証コマンド

```bash
# バックエンドビルド
cd backend && go build ./cmd/clean_server/...

# ユニットテスト
cd backend && go test ./tests/unit/... -v -count=1

# DIP違反チェック（usecasesがdsmysqlをimportしていないか）
grep -rn 'dsmysql' backend/usecases/

# Docker開発環境
docker compose -f docker-compose.dev.yml up --build
```
