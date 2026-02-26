-- ================================================
-- 残高整合スクリプト: transactionsテーブルから残高を再計算
-- ================================================
-- トランザクション履歴に基づいて全ユーザーの残高を再計算する。
--
-- 使い方:
--   docker exec -i gity_point_system-db-1 psql -U postgres -d gity_point < backend/scripts/reconcile_balances.sql
--
-- 残高計算ロジック:
--   受取: to_user_id = me (transfer受取, admin_grant, system_grant, daily_bonus)
--   支出: from_user_id = me AND type IN ('transfer', 'admin_deduct')
--        ※ admin_grant の from_user(admin) は自身のポイントを減らさない
-- ================================================

-- 整合前の状態を確認
SELECT '=== 整合前 ===' AS phase;
SELECT
    COUNT(*) AS total_users,
    SUM(balance) AS total_balance,
    AVG(balance)::INTEGER AS avg_balance,
    MIN(balance) AS min_balance,
    MAX(balance) AS max_balance
FROM users;

-- トランザクションから残高を再計算して更新
-- NOTE: balance >= 0 のCHECK制約があるため、マイナスは0にクランプ
WITH balance_calc AS (
    SELECT
        u.id,
        COALESCE(received.total, 0) - COALESCE(spent.total, 0) AS raw_balance,
        GREATEST(0, COALESCE(received.total, 0) - COALESCE(spent.total, 0)) AS calculated_balance
    FROM users u
    LEFT JOIN (
        -- 受取合計: to_user_id として受け取った全額
        SELECT to_user_id AS user_id, SUM(amount) AS total
        FROM transactions
        WHERE status = 'completed' AND to_user_id IS NOT NULL
        GROUP BY to_user_id
    ) received ON received.user_id = u.id
    LEFT JOIN (
        -- 支出合計: from_user_id として送った/減算された額
        -- admin_grant は管理者のポイントを減らさないので除外
        SELECT from_user_id AS user_id, SUM(amount) AS total
        FROM transactions
        WHERE status = 'completed'
          AND from_user_id IS NOT NULL
          AND transaction_type IN ('transfer', 'admin_deduct')
        GROUP BY from_user_id
    ) spent ON spent.user_id = u.id
)
UPDATE users
SET balance = bc.calculated_balance
FROM balance_calc bc
WHERE users.id = bc.id;

-- 整合後の状態を確認
SELECT '=== 整合後 ===' AS phase;
SELECT
    COUNT(*) AS total_users,
    SUM(balance) AS total_balance,
    AVG(balance)::INTEGER AS avg_balance,
    MIN(balance) AS min_balance,
    MAX(balance) AS max_balance
FROM users;

-- 残高がマイナスのユーザー数
SELECT COUNT(*) AS negative_balance_users
FROM users
WHERE balance < 0;

-- トランザクション種別ごとの総額
SELECT
    transaction_type,
    COUNT(*) AS count,
    SUM(amount) AS total_amount
FROM transactions
WHERE status = 'completed'
GROUP BY transaction_type
ORDER BY total_amount DESC;
