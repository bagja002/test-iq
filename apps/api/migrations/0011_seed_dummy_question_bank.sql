UPDATE test_configs
SET question_count = 120
WHERE test_type = 'IQ' AND is_active = TRUE;

UPDATE test_configs
SET question_count = 30
WHERE test_type = 'SKB' AND is_active = TRUE;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
WITH RECURSIVE seq AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 26
)
SELECT
    CONCAT('DUMMY IQ VCI #', n, ' - Pilih jawaban paling tepat untuk analogi verbal dasar.'),
    'easy',
    'VCI',
    CASE MOD(n, 4)
        WHEN 1 THEN 'SIMILARITIES'
        WHEN 2 THEN 'VOCABULARY'
        WHEN 3 THEN 'INFORMATION'
        ELSE 'COMPREHENSION'
    END,
    'PUBLISHED'
FROM seq;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
WITH RECURSIVE seq AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 25
)
SELECT
    CONCAT('DUMMY IQ PRI #', n, ' - Tentukan pola visual atau bentuk yang paling sesuai.'),
    'medium',
    'PRI',
    CASE MOD(n, 5)
        WHEN 1 THEN 'BLOCK_DESIGN'
        WHEN 2 THEN 'MATRIX_REASONING'
        WHEN 3 THEN 'VISUAL_PUZZLES'
        WHEN 4 THEN 'PICTURE_COMPLETION'
        ELSE 'FIGURE_WEIGHTS'
    END,
    'PUBLISHED'
FROM seq;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
WITH RECURSIVE seq AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 27
)
SELECT
    CONCAT('DUMMY IQ WMI #', n, ' - Hitung atau urutkan informasi singkat sesuai instruksi.'),
    'medium',
    'WMI',
    CASE MOD(n, 3)
        WHEN 1 THEN 'DIGIT_SPAN'
        WHEN 2 THEN 'ARITHMETIC'
        ELSE 'LETTER_NUMBER_SEQUENCING'
    END,
    'PUBLISHED'
FROM seq;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
WITH RECURSIVE seq AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 27
)
SELECT
    CONCAT('DUMMY IQ PSI #', n, ' - Cari simbol, kode, atau target secepat dan setepat mungkin.'),
    'easy',
    'PSI',
    CASE MOD(n, 3)
        WHEN 1 THEN 'SYMBOL_SEARCH'
        WHEN 2 THEN 'CODING'
        ELSE 'CANCELLATION'
    END,
    'PUBLISHED'
FROM seq;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
WITH RECURSIVE seq AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 27
)
SELECT
    CONCAT('DUMMY SKB MANAJER_KOPERASI_KDMP #', n, ' - Pilih tindakan manajerial koperasi yang paling tepat.'),
    'medium',
    'SKB',
    'MANAJER_KOPERASI_KDMP',
    'PUBLISHED'
FROM seq;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
WITH RECURSIVE seq AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 27
)
SELECT
    CONCAT('DUMMY SKB MANAGER_OPERASIONAL_KNMP #', n, ' - Pilih keputusan operasional harian yang paling efektif.'),
    'medium',
    'SKB',
    'MANAGER_OPERASIONAL_KNMP',
    'PUBLISHED'
FROM seq;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
WITH RECURSIVE seq AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 27
)
SELECT
    CONCAT('DUMMY SKB KEPALA_PRODUKSI #', n, ' - Pilih langkah produksi dan pengendalian proses yang paling tepat.'),
    'medium',
    'SKB',
    'KEPALA_PRODUKSI',
    'PUBLISHED'
FROM seq;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
WITH RECURSIVE seq AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 27
)
SELECT
    CONCAT('DUMMY SKB PENGELOLA_KEUANGAN #', n, ' - Pilih tindakan pengelolaan keuangan dan verifikasi yang paling tepat.'),
    'medium',
    'SKB',
    'PENGELOLA_KEUANGAN',
    'PUBLISHED'
FROM seq;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
WITH RECURSIVE seq AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 27
)
SELECT
    CONCAT('DUMMY SKB PENJAMIN_MUTU #', n, ' - Pilih langkah mutu, audit, atau tindak korektif yang paling tepat.'),
    'medium',
    'SKB',
    'PENJAMIN_MUTU',
    'PUBLISHED'
FROM seq;

INSERT INTO question_options (question_id, `key`, content, is_correct)
SELECT
    q.id,
    opt.option_key,
    CASE opt.option_key
        WHEN 'A' THEN CONCAT('Jawaban dummy A untuk soal ', q.id)
        WHEN 'B' THEN CONCAT('Jawaban dummy B untuk soal ', q.id)
        WHEN 'C' THEN CONCAT('Jawaban dummy C untuk soal ', q.id)
        ELSE CONCAT('Jawaban dummy D untuk soal ', q.id)
    END,
    CASE WHEN opt.option_key = 'A' THEN TRUE ELSE FALSE END
FROM questions q
CROSS JOIN (
    SELECT 'A' AS option_key
    UNION ALL SELECT 'B'
    UNION ALL SELECT 'C'
    UNION ALL SELECT 'D'
) opt
LEFT JOIN question_options qo ON qo.question_id = q.id
WHERE q.prompt LIKE 'DUMMY %'
  AND qo.id IS NULL;
