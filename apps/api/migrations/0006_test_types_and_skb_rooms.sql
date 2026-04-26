ALTER TABLE test_configs
    ADD COLUMN test_type VARCHAR(20) NOT NULL DEFAULT 'IQ' AFTER title,
    ADD COLUMN room_code VARCHAR(64) NOT NULL DEFAULT '' AFTER test_type,
    ADD COLUMN room_label VARCHAR(191) NOT NULL DEFAULT '' AFTER room_code,
    ADD KEY idx_test_configs_type_room (test_type, room_code);

ALTER TABLE attempts
    ADD COLUMN test_type VARCHAR(20) NOT NULL DEFAULT 'IQ' AFTER test_config_id,
    ADD COLUMN room_code VARCHAR(64) NOT NULL DEFAULT '' AFTER test_type,
    ADD COLUMN room_label VARCHAR(191) NOT NULL DEFAULT '' AFTER room_code,
    ADD KEY idx_attempts_user_type_room_status (user_id, test_type, room_code, status);

UPDATE test_configs
SET
    test_type = 'IQ',
    room_code = '',
    room_label = ''
WHERE test_type = '' OR room_code = '' OR room_label = '';

UPDATE attempts
SET
    test_type = 'IQ',
    room_code = '',
    room_label = ''
WHERE test_type = '' OR room_code = '' OR room_label = '';

INSERT INTO test_configs (title, test_type, room_code, room_label, duration_minutes, question_count, is_active)
SELECT 'SKB Manajer Kopreasi (KDMP)', 'SKB', 'MANAJER_KOPERASI_KDMP', 'Manajer Kopreasi (KDMP)', 30, 10, TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM test_configs WHERE test_type = 'SKB' AND room_code = 'MANAJER_KOPERASI_KDMP' AND is_active = TRUE
);

INSERT INTO test_configs (title, test_type, room_code, room_label, duration_minutes, question_count, is_active)
SELECT 'SKB Manager Operesial (KNMP)', 'SKB', 'MANAGER_OPERASIONAL_KNMP', 'Manager Operesial (KNMP)', 30, 10, TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM test_configs WHERE test_type = 'SKB' AND room_code = 'MANAGER_OPERASIONAL_KNMP' AND is_active = TRUE
);

INSERT INTO test_configs (title, test_type, room_code, room_label, duration_minutes, question_count, is_active)
SELECT 'SKB Kepala Produksi (KNMP)', 'SKB', 'KEPALA_PRODUKSI', 'Kepala Produksi (KNMP)', 30, 10, TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM test_configs WHERE test_type = 'SKB' AND room_code = 'KEPALA_PRODUKSI' AND is_active = TRUE
);

INSERT INTO test_configs (title, test_type, room_code, room_label, duration_minutes, question_count, is_active)
SELECT 'SKB Pengelola Keuangan (KNMP)', 'SKB', 'PENGELOLA_KEUANGAN', 'Pengelola Keuangan (KNMP)', 30, 10, TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM test_configs WHERE test_type = 'SKB' AND room_code = 'PENGELOLA_KEUANGAN' AND is_active = TRUE
);

INSERT INTO test_configs (title, test_type, room_code, room_label, duration_minutes, question_count, is_active)
SELECT 'SKB Penjamin Mutu (KNMP)', 'SKB', 'PENJAMIN_MUTU', 'Penjamin Mutu (KNMP)', 30, 10, TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM test_configs WHERE test_type = 'SKB' AND room_code = 'PENJAMIN_MUTU' AND is_active = TRUE
);
