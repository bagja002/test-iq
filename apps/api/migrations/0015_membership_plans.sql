CREATE TABLE IF NOT EXISTS membership_plans (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  account_type VARCHAR(20) NOT NULL,
  product_code VARCHAR(64) NOT NULL,
  name VARCHAR(120) NOT NULL,
  description TEXT NOT NULL,
  amount INT NOT NULL,
  submit_limit_per_day INT NOT NULL DEFAULT 0,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uidx_membership_plans_account_type (account_type),
  UNIQUE KEY uidx_membership_plans_product_code (product_code),
  KEY idx_membership_plans_is_active (is_active)
);

INSERT INTO membership_plans (
  account_type,
  product_code,
  name,
  description,
  amount,
  submit_limit_per_day,
  is_active,
  created_at,
  updated_at
)
VALUES
  (
    'PRO',
    'PRO_UPGRADE',
    'Paket Pro',
    'Semua fitur terbuka dengan batas 10 submit IQ dan 10 submit SKB per hari.',
    40000,
    10,
    TRUE,
    NOW(3),
    NOW(3)
  ),
  (
    'MAX',
    'MAX_UPGRADE',
    'Paket Max',
    'Full akses tanpa batas submit harian untuk IQ dan SKB.',
    50000,
    0,
    TRUE,
    NOW(3),
    NOW(3)
  )
ON DUPLICATE KEY UPDATE
  product_code = VALUES(product_code),
  name = VALUES(name),
  description = VALUES(description),
  submit_limit_per_day = VALUES(submit_limit_per_day),
  is_active = VALUES(is_active),
  updated_at = NOW(3);
