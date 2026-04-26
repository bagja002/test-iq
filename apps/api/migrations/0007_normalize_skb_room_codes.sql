UPDATE test_configs
SET
    room_code = 'MANAGER_OPERASIONAL_KNMP',
    room_label = 'Manager Operesial (KNMP)',
    title = 'SKB Manager Operesial (KNMP)'
WHERE test_type = 'SKB' AND room_code = 'MANAGER_OPERASIONAL';

UPDATE test_configs
SET
    room_code = 'MANAJER_KOPERASI_KDMP',
    room_label = 'Manajer Kopreasi (KDMP)',
    title = 'SKB Manajer Kopreasi (KDMP)'
WHERE test_type = 'SKB' AND room_code = 'KEPALA_KOPERASI';

UPDATE test_configs
SET
    room_label = 'Kepala Produksi (KNMP)',
    title = 'SKB Kepala Produksi (KNMP)'
WHERE test_type = 'SKB' AND room_code = 'KEPALA_PRODUKSI';

UPDATE attempts
SET
    room_code = 'MANAGER_OPERASIONAL_KNMP',
    room_label = 'Manager Operesial (KNMP)'
WHERE test_type = 'SKB' AND room_code = 'MANAGER_OPERASIONAL';

UPDATE attempts
SET
    room_code = 'MANAJER_KOPERASI_KDMP',
    room_label = 'Manajer Kopreasi (KDMP)'
WHERE test_type = 'SKB' AND room_code = 'KEPALA_KOPERASI';

UPDATE attempts
SET
    room_label = 'Kepala Produksi (KNMP)'
WHERE test_type = 'SKB' AND room_code = 'KEPALA_PRODUKSI';

UPDATE questions
SET subtest_code = 'MANAGER_OPERASIONAL_KNMP'
WHERE question_index = 'SKB' AND subtest_code = 'MANAGER_OPERASIONAL';

UPDATE questions
SET subtest_code = 'MANAJER_KOPERASI_KDMP'
WHERE question_index = 'SKB' AND subtest_code = 'KEPALA_KOPERASI';

UPDATE attempt_questions
SET subtest_code = 'MANAGER_OPERASIONAL_KNMP'
WHERE question_index = 'SKB' AND subtest_code = 'MANAGER_OPERASIONAL';

UPDATE attempt_questions
SET subtest_code = 'MANAJER_KOPERASI_KDMP'
WHERE question_index = 'SKB' AND subtest_code = 'KEPALA_KOPERASI';

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
