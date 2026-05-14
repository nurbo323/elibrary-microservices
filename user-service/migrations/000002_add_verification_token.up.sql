ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_token TEXT;
CREATE INDEX IF NOT EXISTS idx_users_verification_token ON users(verification_token);