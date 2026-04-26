UPDATE users
SET account_type = 'MAX'
WHERE account_type = 'PAID';

CREATE TABLE IF NOT EXISTS user_devices (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  user_agent VARCHAR(512) NOT NULL,
  user_agent_hash VARCHAR(64) NOT NULL,
  first_seen_at DATETIME(3) NOT NULL,
  last_seen_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  CONSTRAINT fk_user_devices_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY uidx_user_device_hash (user_id, user_agent_hash),
  KEY idx_user_devices_user_id (user_id)
);
