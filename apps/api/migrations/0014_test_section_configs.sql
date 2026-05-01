CREATE TABLE IF NOT EXISTS test_section_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    test_config_id BIGINT UNSIGNED NOT NULL,
    question_index VARCHAR(16) NOT NULL,
    label VARCHAR(191) NOT NULL DEFAULT '',
    order_no INT NOT NULL,
    duration_minutes INT NOT NULL,
    question_count INT NOT NULL,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY uidx_test_section_config (test_config_id, question_index),
    KEY idx_test_section_configs_test_config_id (test_config_id),
    CONSTRAINT fk_test_section_configs_test_config
        FOREIGN KEY (test_config_id) REFERENCES test_configs(id)
        ON DELETE CASCADE
);

UPDATE test_configs
SET
    duration_minutes = 100,
    question_count = 130
WHERE test_type = 'IQ' AND is_active = TRUE;

UPDATE test_configs
SET
    duration_minutes = 60,
    question_count = 50
WHERE test_type = 'SKB' AND is_active = TRUE;

INSERT INTO test_section_configs (
    test_config_id,
    question_index,
    label,
    order_no,
    duration_minutes,
    question_count,
    created_at,
    updated_at
)
SELECT id, 'VCI', 'Verbal Reasoning Index', 1, 40, 50, NOW(3), NOW(3)
FROM test_configs
WHERE test_type = 'IQ' AND is_active = TRUE
ON DUPLICATE KEY UPDATE
    label = VALUES(label),
    order_no = VALUES(order_no),
    duration_minutes = VALUES(duration_minutes),
    question_count = VALUES(question_count),
    updated_at = NOW(3);

INSERT INTO test_section_configs (
    test_config_id,
    question_index,
    label,
    order_no,
    duration_minutes,
    question_count,
    created_at,
    updated_at
)
SELECT id, 'PRI', 'Perceptual Reasoning Index', 2, 15, 20, NOW(3), NOW(3)
FROM test_configs
WHERE test_type = 'IQ' AND is_active = TRUE
ON DUPLICATE KEY UPDATE
    label = VALUES(label),
    order_no = VALUES(order_no),
    duration_minutes = VALUES(duration_minutes),
    question_count = VALUES(question_count),
    updated_at = NOW(3);

INSERT INTO test_section_configs (
    test_config_id,
    question_index,
    label,
    order_no,
    duration_minutes,
    question_count,
    created_at,
    updated_at
)
SELECT id, 'WMI', 'Working Memory Index', 3, 30, 40, NOW(3), NOW(3)
FROM test_configs
WHERE test_type = 'IQ' AND is_active = TRUE
ON DUPLICATE KEY UPDATE
    label = VALUES(label),
    order_no = VALUES(order_no),
    duration_minutes = VALUES(duration_minutes),
    question_count = VALUES(question_count),
    updated_at = NOW(3);

INSERT INTO test_section_configs (
    test_config_id,
    question_index,
    label,
    order_no,
    duration_minutes,
    question_count,
    created_at,
    updated_at
)
SELECT id, 'PSI', 'Processing Speed Index', 4, 15, 20, NOW(3), NOW(3)
FROM test_configs
WHERE test_type = 'IQ' AND is_active = TRUE
ON DUPLICATE KEY UPDATE
    label = VALUES(label),
    order_no = VALUES(order_no),
    duration_minutes = VALUES(duration_minutes),
    question_count = VALUES(question_count),
    updated_at = NOW(3);
