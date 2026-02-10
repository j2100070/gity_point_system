# ユーザー設定機能のテスト手順

## 1. マイグレーション実行

### 前提条件
- PostgreSQLが起動していること
- データベース `point_system` が作成されていること

### マイグレーション実行
```bash
cd backend

# マイグレーションスクリプトを実行
./scripts/run_migration.sh 004_add_user_settings.sql
```

または直接psqlで実行：
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=point_system

PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migrations/004_add_user_settings.sql
```

### マイグレーション確認
```bash
# テーブルが作成されたか確認
PGPASSWORD=$DB_PASSWORD psql -h localhost -p 5432 -U postgres -d point_system -c "\dt"

# 新しいカラムが追加されたか確認
PGPASSWORD=$DB_PASSWORD psql -h localhost -p 5432 -U postgres -d point_system -c "\d users"

# アーカイブユーザーテーブルを確認
PGPASSWORD=$DB_PASSWORD psql -h localhost -p 5432 -U postgres -d point_system -c "\d archived_users"
```

## 2. ユニットテスト実行

### Entityのテスト
```bash
cd backend

# 全てのユニットテストを実行
go test ./tests/unit/entities/... -v

# 特定のテストのみ実行
go test ./tests/unit/entities/user_settings_test.go -v -run TestUser_UpdateProfile
```

### 期待される結果
```
=== RUN   TestUser_UpdateProfile
=== RUN   TestUser_UpdateProfile/表示名のみ更新
=== RUN   TestUser_UpdateProfile/メールアドレス変更時は認証リセット
=== RUN   TestUser_UpdateProfile/表示名とメールの両方を更新
--- PASS: TestUser_UpdateProfile (0.00s)
    --- PASS: TestUser_UpdateProfile/表示名のみ更新 (0.00s)
    --- PASS: TestUser_UpdateProfile/メールアドレス変更時は認証リセット (0.00s)
    --- PASS: TestUser_UpdateProfile/表示名とメールの両方を更新 (0.00s)
...
PASS
ok      github.com/gity/point-system/tests/unit/entities    0.XXXs
```

## 3. 統合テスト実行

### 前提条件
- マイグレーションが完了していること
- データベース接続情報が正しく設定されていること

### 統合テストの実行
```bash
cd backend

# 統合テストを実行（integrationタグ付き）
go test -tags=integration ./tests/integration/user_settings_datasource_test.go -v

# 特定のテストのみ実行
go test -tags=integration ./tests/integration/user_settings_datasource_test.go -v -run TestUserDataSource_InsertWithNewFields
```

### 期待される結果
```
=== RUN   TestUserDataSource_InsertWithNewFields
=== RUN   TestUserDataSource_InsertWithNewFields/新フィールド付きでユーザーを作成
--- PASS: TestUserDataSource_InsertWithNewFields (0.XXs)
    --- PASS: TestUserDataSource_InsertWithNewFields/新フィールド付きでユーザーを作成 (0.XXs)
...
PASS
ok      command-line-arguments  X.XXXs
```

## 4. 手動でのデータ確認

### ユーザーテーブルの確認
```sql
-- 新しいカラムが追加されているか確認
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'users'
  AND column_name IN ('avatar_url', 'avatar_type', 'email_verified', 'email_verified_at');

-- サンプルデータの確認
SELECT id, username, email, avatar_type, email_verified
FROM users
LIMIT 5;
```

### アーカイブユーザーテーブルの確認
```sql
-- テーブル構造の確認
\d archived_users

-- データがあれば表示
SELECT id, username, email, archived_at, deletion_reason
FROM archived_users
ORDER BY archived_at DESC
LIMIT 5;
```

### メール認証トークンテーブルの確認
```sql
-- テーブル構造の確認
\d email_verification_tokens

-- トークンの確認
SELECT id, email, token_type, expires_at, verified_at
FROM email_verification_tokens
ORDER BY created_at DESC
LIMIT 5;
```

## 5. トラブルシューティング

### マイグレーションエラー
```bash
# エラー: relation already exists
# 対処: テーブルが既に存在する場合は削除してから再実行
PGPASSWORD=$DB_PASSWORD psql -h localhost -p 5432 -U postgres -d point_system <<EOF
DROP TABLE IF EXISTS archived_users CASCADE;
DROP TABLE IF EXISTS email_verification_tokens CASCADE;
DROP TABLE IF EXISTS username_change_history CASCADE;
DROP TABLE IF EXISTS password_change_history CASCADE;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_type;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
EOF

# 再度マイグレーション実行
./scripts/run_migration.sh 004_add_user_settings.sql
```

### テスト失敗時
```bash
# データベースをクリーンアップ
PGPASSWORD=$DB_PASSWORD psql -h localhost -p 5432 -U postgres -d point_system <<EOF
TRUNCATE TABLE users CASCADE;
TRUNCATE TABLE archived_users CASCADE;
TRUNCATE TABLE email_verification_tokens CASCADE;
TRUNCATE TABLE username_change_history CASCADE;
TRUNCATE TABLE password_change_history CASCADE;
EOF

# テストを再実行
go test -tags=integration ./tests/integration/user_settings_datasource_test.go -v
```

### ログの確認
```bash
# Goのテストログを詳細表示
go test ./tests/unit/entities/... -v -count=1

# データベースクエリログを有効化（config.goで設定）
# GORM_LOG_LEVEL=4 go test ...
```

## 6. 次のステップ

マイグレーションとテストが成功したら：

1. ✅ データベーススキーマが正しく更新された
2. ✅ Entity層の新機能が動作する
3. ✅ DataSource層のCRUD操作が動作する

次は以下を実装：
- Repository層の実装
- UseCase層（Interactor）の実装
- Controller層とAPIエンドポイントの実装
- フロントエンド実装

## テスト完了チェックリスト

- [ ] マイグレーション004が正常に実行された
- [ ] usersテーブルに新カラムが追加された
- [ ] archived_usersテーブルが作成された
- [ ] email_verification_tokensテーブルが作成された
- [ ] username_change_historyテーブルが作成された
- [ ] password_change_historyテーブルが作成された
- [ ] ユニットテストが全てPASSした
- [ ] 統合テストが全てPASSした
