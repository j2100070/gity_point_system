-- ================================================
-- ユーザー設定機能追加
-- ================================================
-- 機能:
--   1. アバター管理（自動生成 or アップロード）
--   2. メール認証（登録時・メール変更時）
--   3. アーカイブユーザー（論理削除の代替）
--   4. 変更履歴（監査用）
-- ================================================

-- ================================================
-- 1. アバター関連カラム追加
-- ================================================
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(500);
ALTER TABLE users ADD COLUMN avatar_type VARCHAR(50) DEFAULT 'generated'
    CHECK (avatar_type IN ('generated', 'uploaded'));

COMMENT ON COLUMN users.avatar_url IS 'アバター画像のURL（生成URLまたはアップロードされた画像のパス）';
COMMENT ON COLUMN users.avatar_type IS 'アバタータイプ: generated=自動生成, uploaded=ユーザーアップロード';

-- ================================================
-- 2. メール認証関連
-- ================================================

-- メール認証状態をusersテーブルに追加
ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMP WITH TIME ZONE;

COMMENT ON COLUMN users.email_verified IS 'メールアドレスが認証済みかどうか';
COMMENT ON COLUMN users.email_verified_at IS 'メールアドレス認証日時';

-- メール認証トークンテーブル
CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- 登録時はNULL
    email VARCHAR(255) NOT NULL, -- 認証待ちのメールアドレス
    token VARCHAR(255) UNIQUE NOT NULL,
    token_type VARCHAR(50) NOT NULL CHECK (token_type IN ('registration', 'email_change')),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_tokens_token ON email_verification_tokens(token);
CREATE INDEX idx_email_verification_tokens_expires_at ON email_verification_tokens(expires_at);

COMMENT ON TABLE email_verification_tokens IS 'メール認証トークン（登録時・メール変更時）';
COMMENT ON COLUMN email_verification_tokens.user_id IS 'ユーザーID（登録時の認証ではNULL）';
COMMENT ON COLUMN email_verification_tokens.token_type IS 'トークンタイプ: registration=新規登録, email_change=メール変更';

-- ================================================
-- 3. アーカイブユーザーテーブル
-- ================================================

CREATE TABLE archived_users (
    id UUID PRIMARY KEY, -- 元のuser_idを保持
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    balance BIGINT NOT NULL,
    role VARCHAR(50) NOT NULL,
    avatar_url VARCHAR(500),
    avatar_type VARCHAR(50),
    email_verified BOOLEAN NOT NULL,
    email_verified_at TIMESTAMP WITH TIME ZONE,

    -- アーカイブ情報
    archived_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_by UUID, -- 削除実行者（自分 or 管理者）
    deletion_reason TEXT, -- 削除理由

    -- 元のメタデータ
    original_created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    original_updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_archived_users_username ON archived_users(username);
CREATE INDEX idx_archived_users_email ON archived_users(email);
CREATE INDEX idx_archived_users_archived_at ON archived_users(archived_at DESC);
CREATE INDEX idx_archived_users_archived_by ON archived_users(archived_by);

COMMENT ON TABLE archived_users IS 'アーカイブされた（削除された）ユーザー情報。物理削除せずにここで管理。';
COMMENT ON COLUMN archived_users.archived_by IS '削除実行者のユーザーID（自己削除の場合は自分のID、管理者削除の場合は管理者ID）';
COMMENT ON COLUMN archived_users.deletion_reason IS 'アカウント削除理由（任意）';

-- ================================================
-- 4. ユーザー名変更履歴（監査用）
-- ================================================

CREATE TABLE username_change_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    old_username VARCHAR(255) NOT NULL,
    new_username VARCHAR(255) NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    changed_by UUID, -- 変更実行者（自分 or 管理者）
    ip_address INET
);

CREATE INDEX idx_username_change_history_user_id ON username_change_history(user_id);
CREATE INDEX idx_username_change_history_changed_at ON username_change_history(changed_at DESC);

COMMENT ON TABLE username_change_history IS 'ユーザー名変更履歴（監査・不正検知用）';
COMMENT ON COLUMN username_change_history.changed_by IS '変更実行者のユーザーID（通常は本人）';

-- ================================================
-- 5. パスワード変更履歴（セキュリティ監査用）
-- ================================================

CREATE TABLE password_change_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT
);

CREATE INDEX idx_password_change_history_user_id ON password_change_history(user_id);
CREATE INDEX idx_password_change_history_changed_at ON password_change_history(changed_at DESC);

COMMENT ON TABLE password_change_history IS 'パスワード変更履歴（セキュリティ監査用）。ハッシュは保存しない。';

-- ================================================
-- 6. 既存のdeleted_atカラムを削除
-- ================================================
-- アーカイブテーブル方式に移行するため、deleted_atは不要

-- 依存しているビューがある場合は先に削除
DROP VIEW IF EXISTS users_safe CASCADE;

ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;

-- deleted_atのインデックスも削除（存在する場合）
DROP INDEX IF EXISTS idx_users_deleted_at;

-- ================================================
-- 7. 既存データの調整
-- ================================================

-- 既存ユーザーのメール認証状態を全てtrueに設定（既存ユーザーは認証済みとみなす）
UPDATE users SET email_verified = true, email_verified_at = created_at WHERE email_verified = false;


