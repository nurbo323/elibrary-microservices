DROP INDEX IF EXISTS idx_borrows_date_to;
DROP INDEX IF EXISTS idx_borrows_status;
DROP INDEX IF EXISTS idx_borrows_user_id;
DROP INDEX IF EXISTS idx_borrows_exp_id;

ALTER TABLE borrows DROP COLUMN IF EXISTS returned_at;