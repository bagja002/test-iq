CREATE TABLE IF NOT EXISTS attempt_sections (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    attempt_id BIGINT UNSIGNED NOT NULL,
    question_index VARCHAR(16) NOT NULL,
    order_no INT NOT NULL,
    status VARCHAR(20) NOT NULL,
    duration_minutes INT NOT NULL,
    question_count INT NOT NULL,
    started_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    submitted_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_attempt_sections_attempt FOREIGN KEY (attempt_id) REFERENCES attempts(id) ON DELETE CASCADE,
    UNIQUE KEY uidx_attempt_section_code (attempt_id, question_index),
    KEY idx_attempt_section_order (attempt_id, order_no),
    KEY idx_attempt_section_status (status)
);
