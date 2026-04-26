ALTER TABLE users
  ADD COLUMN account_type VARCHAR(20) NOT NULL DEFAULT 'FREE' AFTER status;

UPDATE users
SET account_type = 'FREE'
WHERE account_type IS NULL OR TRIM(account_type) = '';

CREATE INDEX idx_users_account_type ON users(account_type);
