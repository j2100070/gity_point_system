-- ================================================
-- 本名（苗字・名前）カラム追加
-- ================================================

ALTER TABLE users ADD COLUMN first_name VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN last_name VARCHAR(100) NOT NULL DEFAULT '';

COMMENT ON COLUMN users.first_name IS 'ユーザーの名前（プロフィール表示用）';
COMMENT ON COLUMN users.last_name IS 'ユーザーの苗字（プロフィール表示用）';
