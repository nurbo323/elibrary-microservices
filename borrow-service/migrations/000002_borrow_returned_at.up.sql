ALTER TABLE borrows ADD COLUMN IF NOT EXISTS returned_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_borrows_exp_id ON borrows(exp_id);
CREATE INDEX IF NOT EXISTS idx_borrows_user_id ON borrows(user_id);
CREATE INDEX IF NOT EXISTS idx_borrows_status ON borrows(status);
CREATE INDEX IF NOT EXISTS idx_borrows_date_to ON borrows(date_to);