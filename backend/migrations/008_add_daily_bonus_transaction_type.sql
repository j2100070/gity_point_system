-- 008_add_daily_bonus_transaction_type.sql
-- デイリーボーナス用のtransaction_typeを追加

-- 既存のCHECK制約を削除
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_transaction_type_check;

-- daily_bonusを含む新しいCHECK制約を追加
ALTER TABLE transactions ADD CONSTRAINT transactions_transaction_type_check
CHECK (transaction_type IN ('transfer', 'admin_grant', 'admin_deduct', 'system_grant', 'daily_bonus'));

COMMENT ON CONSTRAINT transactions_transaction_type_check ON transactions IS 'デイリーボーナス用のtransaction_typeを追加';
