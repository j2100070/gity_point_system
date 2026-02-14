-- ================================================
-- 009: インデックス最適化
-- ================================================
-- 目的: 書き込みパフォーマンス改善のため、不要なインデックスを削除し、
--       実際のクエリパターンに基づいた複合インデックスに統合する。
--
-- 方針:
--   - UNIQUE制約と重複するインデックスを削除
--   - 低選択性の単一カラムインデックスを削除（status, boolean等）
--   - 現在クエリパターンのないインデックスを削除
--   - 頻出する WHERE + ORDER BY パターンに複合インデックスを作成
-- ================================================

-- ========================================
-- 1. users テーブル
-- ========================================
-- UNIQUE(username), UNIQUE(email), UNIQUE(personal_qr_code) が自動インデックスを持つ
-- ※ deleted_at は 005 で削除済みのため partial index は無効
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_personal_qr_code;

-- ========================================
-- 2. sessions テーブル
-- ========================================
-- UNIQUE(session_token) が自動インデックスを持つ
-- idx_sessions_user_id は DeleteByUserID で使用 → 保持
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_sessions_expires_at;

-- ========================================
-- 3. idempotency_keys テーブル
-- ========================================
-- PK(key) で全クエリをカバー
DROP INDEX IF EXISTS idx_idempotency_user_id;
DROP INDEX IF EXISTS idx_idempotency_expires_at;

-- ========================================
-- 4. transactions テーブル
-- ========================================
-- 単一カラムインデックスを複合インデックスに統合
DROP INDEX IF EXISTS idx_transactions_from_user;
DROP INDEX IF EXISTS idx_transactions_to_user;
DROP INDEX IF EXISTS idx_transactions_idempotency_key;
DROP INDEX IF EXISTS idx_transactions_type;
-- idx_transactions_created_at は管理画面の全件一覧で使用 → 保持

-- 複合インデックス: SelectListByUserID WHERE (from_user_id=? OR to_user_id=?) ORDER BY created_at DESC
CREATE INDEX idx_transactions_from_user_created ON transactions(from_user_id, created_at DESC);
CREATE INDEX idx_transactions_to_user_created ON transactions(to_user_id, created_at DESC);

-- ========================================
-- 5. friendships テーブル
-- ========================================
-- 単一カラムインデックスを複合インデックスに統合
DROP INDEX IF EXISTS idx_friendships_requester;
DROP INDEX IF EXISTS idx_friendships_addressee;
DROP INDEX IF EXISTS idx_friendships_status;

-- 複合インデックス: SelectListFriends WHERE (requester_id=? OR addressee_id=?) AND status=?
CREATE INDEX idx_friendships_requester_status ON friendships(requester_id, status);
-- SelectListPendingRequests WHERE addressee_id=? AND status=?
CREATE INDEX idx_friendships_addressee_status ON friendships(addressee_id, status);

-- ========================================
-- 6. qr_codes テーブル
-- ========================================
-- UNIQUE(code) が自動インデックスを持つ
DROP INDEX IF EXISTS idx_qr_codes_code;
DROP INDEX IF EXISTS idx_qr_codes_user_id;
DROP INDEX IF EXISTS idx_qr_codes_expires_at;

-- 複合インデックス: SelectListByUserID WHERE user_id=? ORDER BY created_at DESC
CREATE INDEX idx_qr_codes_user_created ON qr_codes(user_id, created_at DESC);

-- ========================================
-- 7. audit_logs テーブル
-- ========================================
-- idx_audit_logs_admin_user, idx_audit_logs_created_at は保持
DROP INDEX IF EXISTS idx_audit_logs_target_user;
DROP INDEX IF EXISTS idx_audit_logs_action;

-- ========================================
-- 8. products テーブル
-- ========================================
-- idx_products_category は ReadByCategory で使用 → 保持
DROP INDEX IF EXISTS idx_products_available;

-- ========================================
-- 9. product_exchanges テーブル
-- ========================================
-- 単一カラムインデックスを複合インデックスに統合
DROP INDEX IF EXISTS idx_product_exchanges_user;
DROP INDEX IF EXISTS idx_product_exchanges_product;
DROP INDEX IF EXISTS idx_product_exchanges_status;
DROP INDEX IF EXISTS idx_product_exchanges_created_at;

-- 複合インデックス: SelectListByUserID WHERE user_id=? ORDER BY created_at DESC
CREATE INDEX idx_product_exchanges_user_created ON product_exchanges(user_id, created_at DESC);

-- ========================================
-- 10. categories テーブル
-- ========================================
-- UNIQUE(code) が自動インデックスを持つ。データ量が極めて少ない（~4件）
DROP INDEX IF EXISTS idx_categories_code;
DROP INDEX IF EXISTS idx_categories_active;
DROP INDEX IF EXISTS idx_categories_order;

-- ========================================
-- 11. transfer_requests テーブル
-- ========================================
-- 既存の複合インデックス（idx_transfer_requests_to_user_status, idx_transfer_requests_from_user_created）は保持
-- UNIQUE(idempotency_key) が自動インデックスを持つ
DROP INDEX IF EXISTS idx_transfer_requests_from_user;
DROP INDEX IF EXISTS idx_transfer_requests_to_user;
DROP INDEX IF EXISTS idx_transfer_requests_status;
DROP INDEX IF EXISTS idx_transfer_requests_expires_at;
DROP INDEX IF EXISTS idx_transfer_requests_created_at;
DROP INDEX IF EXISTS idx_transfer_requests_idempotency_key;

-- ========================================
-- 12. daily_bonuses テーブル
-- ========================================
-- idx_daily_bonuses_user_date, idx_daily_bonuses_user_all_completed は保持
DROP INDEX IF EXISTS idx_daily_bonuses_date;

-- ========================================
-- 13. friendships_archive テーブル
-- ========================================
-- requester, addressee は保持
DROP INDEX IF EXISTS idx_friendships_archive_archived_at;

-- ========================================
-- 14. archived_users テーブル
-- ========================================
-- 低頻度テーブル。全インデックスを削除
DROP INDEX IF EXISTS idx_archived_users_username;
DROP INDEX IF EXISTS idx_archived_users_email;
DROP INDEX IF EXISTS idx_archived_users_archived_at;
DROP INDEX IF EXISTS idx_archived_users_archived_by;

-- ========================================
-- 15. 履歴テーブル（username_change_history, password_change_history）
-- ========================================
-- 単一カラムインデックスを複合インデックスに統合
DROP INDEX IF EXISTS idx_username_change_history_user_id;
DROP INDEX IF EXISTS idx_username_change_history_changed_at;
DROP INDEX IF EXISTS idx_password_change_history_user_id;
DROP INDEX IF EXISTS idx_password_change_history_changed_at;

-- 複合インデックス: user_id + changed_at DESC
CREATE INDEX idx_username_change_history_user_changed ON username_change_history(user_id, changed_at DESC);
CREATE INDEX idx_password_change_history_user_changed ON password_change_history(user_id, changed_at DESC);

-- ========================================
-- 16. email_verification_tokens テーブル
-- ========================================
-- UNIQUE(token) が自動インデックスを持つ。token 単体でのみ検索される
DROP INDEX IF EXISTS idx_email_verification_tokens_user_id;
DROP INDEX IF EXISTS idx_email_verification_tokens_token;
DROP INDEX IF EXISTS idx_email_verification_tokens_expires_at;

-- ========================================
-- 完了メッセージ
-- ========================================
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'インデックス最適化が完了しました';
    RAISE NOTICE '  - 不要なインデックスを削除';
    RAISE NOTICE '  - 複合インデックスに統合';
    RAISE NOTICE '  - 書き込みパフォーマンスが改善されます';
    RAISE NOTICE '========================================';
END $$;
